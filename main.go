package main

import (
	"github.com/YasiruR/db-writer/arangodb"
	"github.com/YasiruR/db-writer/domain"
	"github.com/YasiruR/db-writer/elasticsearch"
	"github.com/YasiruR/db-writer/neo4j"
	"github.com/YasiruR/db-writer/redis"
)

func main() {
	dbCfg, dataCfg, testCfg, file := parseArg()
	values := readData(file, &dataCfg)

	var db domain.Database
	switch dbCfg.Typ {
	case domain.Redis:
		db = redis.Client()
	case domain.Neo4j:
		neo4j.Client()
	case domain.ElasticSearch:
		elasticsearch.Client()
	case domain.ArangoDB:
		arangodb.Client()
	}

	db = db.Init(dbCfg)
	if testCfg.Typ == `` {
		db.Write(values, dataCfg)
		return
	}

	switch testCfg.Typ {
	case domain.BenchmarkRead:
		db.BenchmarkRead(values[:testCfg.Load], dataCfg, testCfg)
	case domain.BenchmarkWrite:
		db.BenchmarkWrite(values[:testCfg.Load], dataCfg, testCfg)
	}
}
