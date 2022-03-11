package redis

import (
	"github.com/YasiruR/db-writer/generic"
	goRedis "github.com/go-redis/redis/v8"
)

type redis struct {
	db *goRedis.Client
}

func NewClient() generic.Database {
	return &redis{}
}

func (r *redis) Init(addr, pw string) {
	db := goRedis.NewClient(&goRedis.Options{
		Addr:     addr,
		Password: pw,
		DB:       0,
	})

	r.db = db
}

func (r *redis) Write() {

}
