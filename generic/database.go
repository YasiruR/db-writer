package generic

// database types
const (
	Redis         = `redis`
	Neo4j         = `neo4j`
	ElasticSearch = `elasticsearch`
	ArangoDB      = `arangodb`
)

type Database interface {
	Init(cfg DBConfigs) Database
	Write(values [][]string, cfg DataConfigs)
	BenchmarkRead(values [][]string, dataCfg DataConfigs)
	BenchmarkWrite(values [][]string, dataCfg DataConfigs)
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
	Typ  string
	Load int
}
