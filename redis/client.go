package redis

import (
	"github.com/YasiruR/db-writer/generic"
	goRedis "github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	traceableContext "github.com/tryfix/traceable-context"
	"log"
)

type redis struct {
	db *goRedis.Client
}

func Client() generic.Database {
	return &redis{}
}

func (r *redis) Init(addr, pw string) generic.Database {
	db := goRedis.NewClient(&goRedis.Options{
		Addr:     addr,
		Password: pw,
		DB:       0,
	})

	r.db = db
	return r
}

func (r *redis) Write(values [][]string, opt generic.Options) {
	if opt.UniqIdx < 0 {
		log.Fatalln(`no unique id to store as key`)
	}

	ctx := traceableContext.WithUUID(uuid.New())
	for _, val := range values {
		go func(val []string) {
			r.db.Set(ctx, val[opt.UniqIdx], val, opt.Expiry) // check expiry
		}(val)
	}
}
