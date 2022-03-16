package main

import (
	"fmt"
	"github.com/YasiruR/db-writer/elasticsearch"
	"github.com/YasiruR/db-writer/generic"
	"github.com/YasiruR/db-writer/neo4j"
	"github.com/YasiruR/db-writer/redis"
)

func main() {
	fmt.Println()
	dbCfg, dataCfg, file := parseArg()
	values := readData(file, &dataCfg)

	switch dbCfg.Typ {
	case generic.Redis:
		redis.Client().Init(dbCfg).Write(values, dataCfg)
	case generic.Neo4j:
		neo4j.Client().Init(dbCfg).Write(values, dataCfg)
	case generic.ElasticSearch:
		elasticsearch.Client().Init(dbCfg).Write(values, dataCfg)
	}
}
