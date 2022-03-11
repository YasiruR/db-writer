package main

import (
	"github.com/YasiruR/db-writer/generic"
	"github.com/YasiruR/db-writer/redis"
)

func main() {
	db, addr, pw, file, uniqueKey := parseArg()
	uniqIdx, _, values := readData(file, uniqueKey)

	switch db {
	case generic.Redis:
		redis.Client().Init(addr, pw).Write(values, generic.Options{UniqIdx: uniqIdx})
	}
}
