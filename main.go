package main

import (
	"github.com/YasiruR/db-writer/arangodb"
	"github.com/YasiruR/db-writer/elasticsearch"
	"github.com/YasiruR/db-writer/generic"
	"github.com/YasiruR/db-writer/neo4j"
	"github.com/YasiruR/db-writer/redis"
)

func main() {
	dbCfg, dataCfg, testCfg, file := parseArg()
	values := readData(file, &dataCfg)

	var db generic.Database
	switch dbCfg.Typ {
	case generic.Redis:
		db = redis.Client()
	case generic.Neo4j:
		neo4j.Client()
	case generic.ElasticSearch:
		elasticsearch.Client()
	case generic.ArangoDB:
		arangodb.Client()
	}

	db = db.Init(dbCfg)
	if testCfg.Typ == `` {
		db.Write(values, dataCfg)
		return
	}

	db.BenchmarkRead(values[:testCfg.Load], dataCfg) // todo param
}
