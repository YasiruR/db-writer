package main

import (
	"github.com/YasiruR/db-writer/generic"
	"github.com/YasiruR/db-writer/neo4j"
	"github.com/YasiruR/db-writer/redis"
)

func main() {
	dbCfg, dataCfg, file := parseArg()
	fields, values := readData(file, &dataCfg)

	switch dbCfg.Typ {
	case generic.Redis:
		redis.Client().Init(dbCfg).Write(fields, values, dataCfg)
	case generic.Neo4j:
		neo4j.Client().Init(dbCfg).Write(fields, values, dataCfg)
	}
}
