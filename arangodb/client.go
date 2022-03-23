package arangodb

import (
	"context"
	"fmt"
	"github.com/YasiruR/db-writer/generic"
	"github.com/YasiruR/db-writer/log"
	driver "github.com/arangodb/go-driver"
	dbHttp "github.com/arangodb/go-driver/http"
	"github.com/google/uuid"
	traceableContext "github.com/tryfix/traceable-context"
	"sync"
	"sync/atomic"
)

type arangodb struct {
	db driver.Database
}

func Client() generic.Database {
	return &arangodb{}
}

func (a *arangodb) Init(cfg generic.DBConfigs) generic.Database {
	ctx := traceableContext.WithUUID(uuid.New())
	conn, err := dbHttp.NewConnection(dbHttp.ConnectionConfig{Endpoints: cfg.Hosts})
	if err != nil {
		fmt.Println(`HOSTS: `, cfg.Hosts)
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

func (a *arangodb) Write(values [][]string, dataCfg generic.DataConfigs) {
	ctx := traceableContext.WithUUID(uuid.New())
	wg := &sync.WaitGroup{}
	var success uint64

	collExists, err := a.db.CollectionExists(ctx, dataCfg.TableName)
	if err != nil {
		log.Fatal(err)
	}

	var coll driver.Collection
	switch collExists {
	case true:
		coll, err = a.db.Collection(ctx, dataCfg.TableName)
		if err != nil {
			log.Fatal(err)
		}
	case false:
		coll, err = a.db.CreateCollection(ctx, dataCfg.TableName, &driver.CreateCollectionOptions{})
		if err != nil {
			log.Fatal(err)
		}
	}

	for i, val := range values {
		if dataCfg.Limit >= 0 && dataCfg.Limit == i {
			break
		}

		fmt.Printf("\rSending data: %d/%d", i+1, len(values))

		d := data{Body: val}

		wg.Add(1)
		go func(val []string) {
			defer wg.Done()
			_, err = coll.CreateDocument(ctx, d.Map(dataCfg))
			if err != nil {
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

func (a *arangodb) BenchmarkRead(values [][]string, dataCfg generic.DataConfigs)  {}
func (a *arangodb) BenchmarkWrite(values [][]string, dataCfg generic.DataConfigs) {}
