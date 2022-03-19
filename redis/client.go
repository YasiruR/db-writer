package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/YasiruR/db-writer/generic"
	"github.com/YasiruR/db-writer/log"
	goRedis "github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	traceableContext "github.com/tryfix/traceable-context"
	"sync"
	"sync/atomic"
)

type redis struct {
	client goRedis.Cmdable
}

func Client() generic.Database {
	return &redis{}
}

func (r *redis) Init(cfg generic.DBConfigs) generic.Database {
	switch len(cfg.Hosts) {
	case 1:
		db := goRedis.NewClient(&goRedis.Options{
			Addr:     cfg.Hosts[0],
			Password: cfg.Passwd,
			DB:       0,
		})
		r.client = db
	default:
		c := goRedis.NewClusterClient(&goRedis.ClusterOptions{Addrs: cfg.Hosts})
		if err := c.Ping(context.Background()).Err(); err != nil {
			log.Fatal(errors.New("Unable to connect to redis " + err.Error()))
		}
		r.client = c
	}

	fmt.Println(`Database connection established with redis`)
	return r
}

func (r *redis) Write(values [][]string, dataCfg generic.DataConfigs) {
	if dataCfg.Unique.Index < 0 {
		log.Fatal(errors.New(`no unique id to store as key`))
	}

	ctx := traceableContext.WithUUID(uuid.New())
	wg := &sync.WaitGroup{}
	var success uint64

	for i, val := range values {
		if dataCfg.Limit >= 0 && dataCfg.Limit == i {
			break
		}

		fmt.Printf("\rSending data: %d/%d", i+1, len(values))

		rv := data{body: val}
		wg.Add(1)

		go func(val []string, wg *sync.WaitGroup) {
			defer wg.Done()
			cmd := r.client.Set(ctx, val[dataCfg.Unique.Index], rv, 0) // check expiry
			if cmd.Err() != nil {
				log.Error(errors.New("ERROR: " + cmd.Err().Error()))
				return
			}

			_, err := cmd.Result()
			if err != nil {
				log.Error(err)
				return
			}

			atomic.AddUint64(&success, 1)
		}(val, wg)
	}

	fmt.Println("\nWaiting for the database to complete operations...")
	wg.Wait()
	fmt.Println(`Total successful writes: `, int(success))
}

func (r *redis) read(values [][]string, dataCfg generic.DataConfigs) {
	ctx := traceableContext.WithUUID(uuid.New())
	for i, val := range values {
		if dataCfg.Limit >= 0 && dataCfg.Limit == i {
			break
		}

		cmd := r.client.Get(ctx, val[dataCfg.Unique.Index])
		if cmd.Err() != nil {
			log.Error(cmd.Err())
			continue
		}

		fmt.Printf("key: %s val: %s\n", val[dataCfg.Unique.Index], cmd.Val())
	}
}
