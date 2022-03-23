package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/YasiruR/db-writer/domain"
	"github.com/YasiruR/db-writer/log"
	"github.com/YasiruR/db-writer/tester"
	goRedis "github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	traceableContext "github.com/tryfix/traceable-context"
	"sync"
	"sync/atomic"
	"time"
)

type redis struct {
	client goRedis.Cmdable
}

func Client() domain.Database {
	return &redis{}
}

func (r *redis) Init(cfg domain.DBConfigs) domain.Database {
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

func (r *redis) Write(values [][]string, dataCfg domain.DataConfigs) {
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
				log.Error(errors.New(cmd.Err().Error()))
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

func (r *redis) read(values [][]string, dataCfg domain.DataConfigs) {
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

func (r *redis) BenchmarkRead(values [][]string, dataCfg domain.DataConfigs, testCfg domain.TestConfigs) {
	var aggrLatencyMicSec, success uint64
	wg := &sync.WaitGroup{}
	ctx := traceableContext.WithUUID(uuid.New())

	// setting up ids
	var ids []string
	for _, val := range values {
		ids = append(ids, val[dataCfg.Unique.Index])
	}

	testStartedTime := time.Now()
	for i, id := range ids {
		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()
			startedTime := time.Now()
			cmd := r.client.Get(ctx, id)
			elapsedTime := time.Since(startedTime).Microseconds()

			if cmd.Err() != nil {
				log.Error(cmd.Err())
				return
			}

			rv := data{body: values[i]}

			if cmd.Val() != rv.Str() {
				log.Error(errors.New(`read values are not equal`), cmd.Val(), rv.Str())
				return
			}

			atomic.AddUint64(&aggrLatencyMicSec, uint64(elapsedTime))
			atomic.AddUint64(&success, 1)
		}(i, id)
	}

	wg.Wait()
	totalDurMicSec := time.Since(testStartedTime).Microseconds()
	tester.Output(testCfg, success, uint64(totalDurMicSec), aggrLatencyMicSec, true)
}

func (r *redis) BenchmarkWrite(values [][]string, dataCfg domain.DataConfigs, testCfg domain.TestConfigs) {
	var aggrLatencyMicSec, success uint64
	wg := &sync.WaitGroup{}
	ctx := traceableContext.WithUUID(uuid.New())

	// setting up ids
	var ids []string
	var rValues []data
	for _, val := range values {
		ids = append(ids, val[dataCfg.Unique.Index])
		rValues = append(rValues, data{body: val})
	}

	testStartedTime := time.Now()
	for i, val := range rValues {
		wg.Add(1)
		go func(i int, val data) {
			defer wg.Done()
			startedTime := time.Now()
			cmd := r.client.Set(ctx, ids[i], val, 0)
			elapsedTime := time.Since(startedTime).Microseconds()

			if cmd.Err() != nil {
				log.Error(cmd.Err())
				return
			}

			atomic.AddUint64(&aggrLatencyMicSec, uint64(elapsedTime))
			atomic.AddUint64(&success, 1)
		}(i, val)
	}

	wg.Wait()
	totalDurMicSec := time.Since(testStartedTime).Microseconds()
	tester.Output(testCfg, success, uint64(totalDurMicSec), aggrLatencyMicSec, true)
}
