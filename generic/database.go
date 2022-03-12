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
	Write(fields []string, values [][]string, opt Options)
}

type DBConfigs struct {
	Addr     string
	Username string
	Passwd   string
}

type Options struct {
	UniqIdx int
	Persist bool
	Expiry  time.Duration
}
