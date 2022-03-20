package generic

// database types
const (
	Redis         = `redis`
	Neo4j         = `neo4j`
	ElasticSearch = `elasticsearch`
	ArangoDB      = `arangodb`
	Couchbase     = `couchbase`
)

type Database interface {
	Init(cfg DBConfigs) Database
	Write(values [][]string, cfg DataConfigs)
}

type DBConfigs struct {
	Typ      string
	Hosts    []string
	Username string
	Passwd   string
	CACert   string
	Name     string
}
