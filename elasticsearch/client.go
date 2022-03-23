package elasticsearch

import (
	"errors"
	"fmt"
	"github.com/YasiruR/db-writer/generic"
	"github.com/YasiruR/db-writer/log"
	goEs "github.com/elastic/go-elasticsearch/v8"
	goEsApi "github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
	traceableContext "github.com/tryfix/traceable-context"
	"io/ioutil"
	"strconv"
	"strings"
	"sync/atomic"
)

const index = `elastic-db`

type elasticsearch struct {
	db *goEs.Client
}

func Client() generic.Database {
	return &elasticsearch{}
}

func (e *elasticsearch) Init(cfg generic.DBConfigs) generic.Database {
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

func (e *elasticsearch) Write(values [][]string, dataCfg generic.DataConfigs) {
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
			Index:      index,
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
			fmt.Println()
		} else {
			atomic.AddUint64(&success, 1)
		}
		res.Body.Close()
	}

	fmt.Println("\nWaiting for the database to complete operations...")
	fmt.Println("Total successful writes: ", int(success))
}

func (e *elasticsearch) BenchmarkRead(values [][]string, dataCfg generic.DataConfigs)  {}
func (e *elasticsearch) BenchmarkWrite(values [][]string, dataCfg generic.DataConfigs) {}
