package domain

// database types
const (
	Redis         = `redis`
	Neo4j         = `neo4j`
	ElasticSearch = `elasticsearch`
	ArangoDB      = `arangodb`
)

// benchmark types
const (
	BenchmarkRead   = `read`
	BenchmarkWrite  = `write`
	BenchmarkUpdate = `update`
)

type Database interface {
	Init(cfg DBConfigs) Database
	Write(values [][]string, cfg DataConfigs)
	BenchmarkRead(values [][]string, dataCfg DataConfigs, testCfg TestConfigs)
	BenchmarkWrite(values [][]string, dataCfg DataConfigs, testCfg TestConfigs)
}

type DBConfigs struct {
	Typ      string
	Hosts    []string
	Username string
	Passwd   string
	CACert   string
	Name     string
}

type TestConfigs struct {
	Database string
	Typ      string
	Load     int
	TxSizes  []int
	TxBuffer int // todo
}
