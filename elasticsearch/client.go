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
	"sync"
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
		Addresses: []string{cfg.Addr},
		Username:  `elastic`,
		Password:  `*on0YcAz49+NpmGjgilV`,
		CACert:    e.readCert(),
	})

	//es, err := goEs.NewDefaultClient()
	if err != nil {
		log.Fatal(err, cfg.CACert)
	}

	//res, err := es.Info()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer res.Body.Close()
	//fmt.Println(`connected to elasticsearch: `, res)

	e.db = es
	return e
}

func (e *elasticsearch) readCert() []byte {
	cert, err := ioutil.ReadFile(`/home/yasi/Documents/http_ca.crt`)
	if err != nil {
		log.Fatal(err)
	}

	return cert
}

func (e *elasticsearch) Write(values [][]string, dataCfg generic.DataConfigs) {
	var success uint64
	wg := &sync.WaitGroup{}
	ctx := traceableContext.WithUUID(uuid.New())

	for i, val := range values {
		if dataCfg.Limit >= 0 && dataCfg.Limit == i {
			break
		}

		wg.Add(1)
		go func(i int, val []string) {
			defer wg.Done()
			jsonVal := data{Body: val}.JSON(dataCfg)

			var docID string
			if dataCfg.UniqIdx < 0 {
				docID = strconv.Itoa(i + 1)
			} else {
				docID = val[dataCfg.UniqIdx]
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
				return
			}

			if res.IsError() {
				log.Error(errors.New(res.String()), val[7]) // todo remove index
				fmt.Println()
			} else {
				atomic.AddUint64(&success, 1)
			}
			defer res.Body.Close()
		}(i, val)
	}

	wg.Wait()
	fmt.Println(`total writes (elasticsearch): `, int(success))
}
