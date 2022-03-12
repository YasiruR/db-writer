package generic

import "time"

// database types
const (
	Redis     = `redis`
	Neo4j     = `neo4j`
	Couchbase = `couchbase`
)

type Database interface {
	Init(cfg DBConfigs) Database
	Write(fields []string, values [][]string, cfg DataConfigs)
}

type DBConfigs struct {
	Typ      string
	Addr     string
	Username string
	Passwd   string
}

type DataConfigs struct {
	UniqKey string
	UniqIdx int
	Limit   int
}

type Options struct {
	UniqIdx int
	Persist bool
	Expiry  time.Duration
}
