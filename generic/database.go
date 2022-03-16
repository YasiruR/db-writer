package generic

// database types
const (
	Redis         = `redis`
	Neo4j         = `neo4j`
	ElasticSearch = `elasticsearch`
	Couchbase     = `couchbase`
)

type Database interface {
	Init(cfg DBConfigs) Database
	Write(values [][]string, cfg DataConfigs)
}

type DBConfigs struct {
	Typ      string
	Addr     string //todo list of hosts
	Username string
	Passwd   string
	CACert   string
}
