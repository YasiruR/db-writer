package redis

import (
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
	db *goRedis.Client
}

func Client() generic.Database {
	return &redis{}
}

func (r *redis) Init(cfg generic.DBConfigs) generic.Database {
	db := goRedis.NewClient(&goRedis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Passwd,
		DB:       0,
	})

	r.db = db
	return r
}

type redisVal struct {
	body []string
}

func (v redisVal) MarshalBinary() ([]byte, error) {
	return []byte(fmt.Sprintf("%v", v)), nil
}

func (r *redis) Write(values [][]string, dataCfg generic.DataConfigs) {
	if dataCfg.UniqIdx < 0 {
		log.Fatal(errors.New(`no unique id to store as key`))
	}

	ctx := traceableContext.WithUUID(uuid.New())
	wg := &sync.WaitGroup{}
	var success uint64

	for i, val := range values {
		if dataCfg.Limit >= 0 && dataCfg.Limit == i {
			break
		}

		rv := redisVal{body: val}
		wg.Add(1)

		go func(val []string, wg *sync.WaitGroup) {
			defer wg.Done()
			cmd := r.db.Set(ctx, val[dataCfg.UniqIdx], rv, 0) // check expiry
			if cmd.Err() != nil {
				log.Error(cmd.Err())
				return
			}
			atomic.AddUint64(&success, 1)
		}(val, wg)
	}

	wg.Wait()
	fmt.Println(`total writes (redis): `, int(success))
}
