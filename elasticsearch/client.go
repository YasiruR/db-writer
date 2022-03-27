package elasticsearch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/YasiruR/db-writer/domain"
	"github.com/YasiruR/db-writer/log"
	goEs "github.com/elastic/go-elasticsearch/v8"
	goEsApi "github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
	traceableContext "github.com/tryfix/traceable-context"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const index = `elastic-db`

type elasticsearch struct {
	db *goEs.Client
}

func Client() domain.Database {
	return &elasticsearch{}
}

func (e *elasticsearch) Init(cfg domain.DBConfigs) domain.Database {
	es, err := goEs.NewClient(goEs.Config{
		Addresses: cfg.Hosts,
		Username:  cfg.Username,
		Password:  cfg.Passwd,
		CACert:    e.readCert(cfg.CACert),
	})

	if err != nil {
		log.Fatal(err)
	}

	e.db = es
	fmt.Println(`Database connection established with elasticsearch`)

	return e
}

func (e *elasticsearch) readCert(file string) []byte {
	cert, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	return cert
}

func (e *elasticsearch) Write(values [][]string, dataCfg domain.DataConfigs) {
	var success uint64
	ctx := traceableContext.WithUUID(uuid.New())

	for i, val := range values {
		if dataCfg.Limit >= 0 && dataCfg.Limit == i {
			break
		}

		fmt.Printf("\rSending data synchronously: %d/%d", i+1, len(values))

		jsonVal := data{Body: val}.JSON(dataCfg)
		var docID string
		if dataCfg.Unique.Index < 0 {
			docID = strconv.Itoa(i + 1)
		} else {
			docID = val[dataCfg.Unique.Index]
		}

		req := goEsApi.IndexRequest{
			Index:      dataCfg.Table, // todo change to database
			DocumentID: docID,
			Body:       strings.NewReader(jsonVal),
			Refresh:    "true",
		}

		res, err := req.Do(ctx, e.db)
		if err != nil {
			log.Error(err)
		}

		if res.IsError() {
			log.Error(errors.New(res.String()))
		} else {
			atomic.AddUint64(&success, 1)
		}
		res.Body.Close()
	}

	fmt.Println("\nWaiting for the database to complete operations...")
	fmt.Println("Total successful writes: ", int(success))
}

func (e *elasticsearch) BenchmarkRead(values [][]string, dataCfg domain.DataConfigs, testCfg domain.TestConfigs) {
	values = values[:testCfg.Load]
	var aggrLatencyMicSec, success uint64
	wg := &sync.WaitGroup{}
	ctx := traceableContext.WithUUID(uuid.New())

	var queries []bytes.Buffer
	var ids []string
	for _, val := range values {
		var buf bytes.Buffer
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					dataCfg.Unique.Key: val[dataCfg.Unique.Index],
				},
			},
		}

		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			log.Error(err)
			continue
		}

		queries = append(queries, buf)
		ids = append(ids, val[dataCfg.Unique.Index])
	}

	testStartedTime := time.Now()
	for i, q := range queries {
		wg.Add(1)
		go func(i int, q bytes.Buffer) {
			defer wg.Done()
			startedTime := time.Now()
			res, err := e.db.Search(
				e.db.Search.WithContext(ctx),
				e.db.Search.WithIndex(dataCfg.Table),
				e.db.Search.WithBody(&q),
				e.db.Search.WithTrackTotalHits(true),
				e.db.Search.WithPretty())
			elapsedTime := time.Since(startedTime).Microseconds()

			if err != nil {
				log.Error(err)
				return
			}
			defer res.Body.Close()

			if res.IsError() {
				var errMap map[string]interface{}
				if err = json.NewDecoder(res.Body).Decode(&errMap); err != nil {
					log.Error(err, "Error parsing the response body")
					return
				}

				// Print the response status and error information.
				log.Error(errors.New(
					fmt.Sprintf("[%s] %s: %s",
						res.Status(),
						errMap["error"].(map[string]interface{})["type"],
						errMap["error"].(map[string]interface{})["reason"])),
				)
				return
			}

			var resData map[string]interface{}
			if err = json.NewDecoder(res.Body).Decode(&resData); err != nil {
				log.Error(err, "Error parsing the response body")
				return
			}

			atomic.AddUint64(&aggrLatencyMicSec, uint64(elapsedTime))
			atomic.AddUint64(&success, 1)
		}(i, q)
	}

	wg.Wait()
	totalDurMicSec := time.Since(testStartedTime).Microseconds()
	log.Output(testCfg, success, uint64(totalDurMicSec), aggrLatencyMicSec, true)
}

func (e *elasticsearch) BenchmarkWrite(values [][]string, dataCfg domain.DataConfigs, testCfg domain.TestConfigs) {
	if len(testCfg.TxSizes) == 0 {
		values = values[:testCfg.Load]
	}

	var aggrLatencyMicSec, success uint64
	wg := &sync.WaitGroup{}
	ctx := traceableContext.WithUUID(uuid.New())
	reqs := e.getData(values, dataCfg, testCfg)

	testStartedTime := time.Now()
	for _, req := range reqs {
		wg.Add(1)
		go func(req goEsApi.IndexRequest) {
			defer wg.Done()
			startedTime := time.Now()
			res, err := req.Do(ctx, e.db)
			elapsedTime := time.Since(startedTime).Microseconds()
			if err != nil {
				log.Error(err)
				return
			}

			defer res.Body.Close()
			if res.IsError() {
				log.Error(errors.New(res.String()))
				return
			}

			atomic.AddUint64(&aggrLatencyMicSec, uint64(elapsedTime))
			atomic.AddUint64(&success, 1)
		}(req)
	}

	wg.Wait()
	totalDurMicSec := time.Since(testStartedTime).Microseconds()
	log.Output(testCfg, success, uint64(totalDurMicSec), aggrLatencyMicSec, true)
}

func (e *elasticsearch) getData(values [][]string, dataCfg domain.DataConfigs, testCfg domain.TestConfigs) (reqs []goEsApi.IndexRequest) {
	var d data
	// if tx sizes are provided, filter the inputs
	if len(testCfg.TxSizes) != 0 {
		for i, val := range values {
			d = data{Body: val}
			dataSize := len(d.Str())

			for _, txSize := range testCfg.TxSizes {
				var upper, lower int
				upper = txSize + testCfg.TxBuffer
				lower = txSize - testCfg.TxBuffer
				if lower < dataSize && dataSize < upper {
					var docID string
					if dataCfg.Unique.Index < 0 {
						docID = strconv.Itoa(i + 1)
					} else {
						docID = val[dataCfg.Unique.Index]
					}

					req := goEsApi.IndexRequest{
						Index:      dataCfg.Table, // todo change to database
						DocumentID: docID,
						Body:       strings.NewReader(d.JSON(dataCfg)),
						Refresh:    "true",
					}

					reqs = append(reqs, req)
					break
				}
			}
		}

		return
	}

	// if no tx size filtering add all given indexes
	for i, val := range values {
		jsonVal := data{Body: val}.JSON(dataCfg)
		var docID string
		if dataCfg.Unique.Index < 0 {
			docID = strconv.Itoa(i + 1)
		} else {
			docID = val[dataCfg.Unique.Index]
		}

		req := goEsApi.IndexRequest{
			Index:      dataCfg.Table, // todo change to database
			DocumentID: docID,
			Body:       strings.NewReader(jsonVal),
			Refresh:    "true",
		}

		reqs = append(reqs, req)
	}

	return
}
