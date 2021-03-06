package arangodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/YasiruR/db-writer/domain"
	"github.com/YasiruR/db-writer/log"
	driver "github.com/arangodb/go-driver"
	dbHttp "github.com/arangodb/go-driver/http"
	"github.com/google/uuid"
	traceableContext "github.com/tryfix/traceable-context"
	"sync"
	"sync/atomic"
	"time"
)

type arangodb struct {
	db driver.Database
}

func Client() domain.Database {
	return &arangodb{}
}

func (a *arangodb) Init(cfg domain.DBConfigs) domain.Database {
	ctx := traceableContext.WithUUID(uuid.New())
	conn, err := dbHttp.NewConnection(dbHttp.ConnectionConfig{Endpoints: cfg.Hosts})
	if err != nil {
		log.Fatal(err)
	}

	c, err := driver.NewClient(driver.ClientConfig{Connection: conn})
	if err != nil {
		log.Fatal(err)
	}

	exists, err := c.DatabaseExists(ctx, cfg.Name)
	if err != nil {
		log.Fatal(err)
	}

	var db driver.Database
	switch exists {
	case true:
		db, err = c.Database(ctx, cfg.Name)
		if err != nil {
			log.Fatal(err)
		}
	case false:
		db, err = c.Database(context.Background(), cfg.Name)
		if err != nil {
			log.Fatal(err)
		}
	}

	a.db = db
	return a
}

func (a *arangodb) Write(values [][]string, dataCfg domain.DataConfigs) {
	ctx := traceableContext.WithUUID(uuid.New())
	wg := &sync.WaitGroup{}
	var success uint64
	coll := a.collection(ctx, dataCfg)

	for i, val := range values {
		if dataCfg.Limit >= 0 && dataCfg.Limit == i {
			break
		}

		fmt.Printf("\rSending data: %d/%d", i+1, len(values))

		d := data{body: val}

		wg.Add(1)
		go func(val []string) {
			defer wg.Done()
			doc := d.document(dataCfg)
			_, err := coll.CreateDocument(ctx, doc)
			if err != nil {
				arangoErr, ok := driver.AsArangoError(err)
				if !ok {
					log.Error(errors.New("(non-arangodb error) " + err.Error()))
					return
				}

				// update the document if key already exists
				if driver.IsArangoErrorWithCode(arangoErr, 409) {
					_, err = coll.UpdateDocument(ctx, val[dataCfg.Unique.Index], doc)
					if err != nil {
						log.Error(err)
						return
					}
					atomic.AddUint64(&success, 1)
					return
				}
				log.Error(err)
				return
			}
			atomic.AddUint64(&success, 1)
		}(val)
	}

	fmt.Println("\nWaiting for the database to complete operations...")
	wg.Wait()
	fmt.Println(`Total successful writes: `, int(success))
}

func (a *arangodb) BenchmarkRead(values [][]string, dataCfg domain.DataConfigs, testCfg domain.TestConfigs) {
	values = values[:testCfg.Load]
	var aggrLatencyMicSec, success uint64
	wg := &sync.WaitGroup{}
	ctx := traceableContext.WithUUID(uuid.New())

	// setting up ids
	var ids []string
	for _, val := range values {
		ids = append(ids, val[dataCfg.Unique.Index])
	}

	coll, err := a.db.Collection(ctx, dataCfg.Table)
	if err != nil {
		log.Fatal(err)
	}

	testStartedTime := time.Now()
	for _, id := range ids {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			doc := make(map[string]interface{})
			startedTime := time.Now()
			_, err = coll.ReadDocument(ctx, id, &doc)
			elapsedTime := time.Since(startedTime).Microseconds()
			if err != nil {
				log.Error(err)
				return
			}

			atomic.AddUint64(&aggrLatencyMicSec, uint64(elapsedTime))
			atomic.AddUint64(&success, 1)
		}(id)
	}

	wg.Wait()
	totalDurMicSec := time.Since(testStartedTime).Microseconds()
	log.Output(testCfg, success, uint64(totalDurMicSec), aggrLatencyMicSec, true)
}

func (a *arangodb) BenchmarkWrite(values [][]string, dataCfg domain.DataConfigs, testCfg domain.TestConfigs) {
	if len(testCfg.TxSizes) == 0 {
		values = values[:testCfg.Load]
	}

	var aggrLatencyMicSec, success uint64
	wg := &sync.WaitGroup{}
	ctx := traceableContext.WithUUID(uuid.New())
	ids, docs := a.getData(values, dataCfg, testCfg)
	coll := a.collection(ctx, dataCfg)

	testStartedTime := time.Now()
	for i, val := range docs {
		wg.Add(1)
		go func(i int, val map[string]interface{}) {
			defer wg.Done()
			var startedTime time.Time
			var elapsedTime int64

			switch testCfg.Typ {
			case domain.BenchmarkWrite:
				startedTime = time.Now()
				_, err := coll.CreateDocument(ctx, val)
				elapsedTime = time.Since(startedTime).Microseconds()
				if err != nil {
					log.Error(err)
					return
				}
			case domain.BenchmarkUpdate:
				startedTime = time.Now()
				_, err := coll.UpdateDocument(ctx, ids[i], val)
				elapsedTime = time.Since(startedTime).Microseconds()
				if err != nil {
					log.Error(err)
					return
				}
			}

			atomic.AddUint64(&aggrLatencyMicSec, uint64(elapsedTime))
			atomic.AddUint64(&success, 1)
		}(i, val)
	}

	wg.Wait()
	totalDurMicSec := time.Since(testStartedTime).Microseconds()
	log.Output(testCfg, success, uint64(totalDurMicSec), aggrLatencyMicSec, true)
}

func (a *arangodb) getData(values [][]string, dataCfg domain.DataConfigs, testCfg domain.TestConfigs) (ids []string, docs []map[string]interface{}) {
	var d data

	// if tx sizes are provided, filter the inputs
	if len(testCfg.TxSizes) != 0 {
		for _, val := range values {
			if len(ids) == testCfg.Load {
				return
			}

			d = data{body: val}
			dataSize := len(d.Str())

			for _, txSize := range testCfg.TxSizes {
				var upper, lower int
				upper = txSize + testCfg.TxBuffer
				lower = txSize - testCfg.TxBuffer
				if lower < dataSize && dataSize < upper {
					ids = append(ids, val[dataCfg.Unique.Index])
					docs = append(docs, d.document(dataCfg))
					break
				}
			}
		}

		return
	}

	// if no tx size filtering add all given data
	for _, val := range values {
		ids = append(ids, val[dataCfg.Unique.Index])
		docs = append(docs, data{body: val}.document(dataCfg))
	}

	return
}

func (a *arangodb) collection(ctx context.Context, dataCfg domain.DataConfigs) driver.Collection {
	collExists, err := a.db.CollectionExists(ctx, dataCfg.Table)
	if err != nil {
		log.Fatal(err)
	}

	var coll driver.Collection
	switch collExists {
	case true:
		coll, err = a.db.Collection(ctx, dataCfg.Table)
		if err != nil {
			log.Fatal(err)
		}
	case false:
		coll, err = a.db.CreateCollection(ctx, dataCfg.Table, &driver.CreateCollectionOptions{})
		if err != nil {
			log.Fatal(err)
		}
	}

	return coll
}