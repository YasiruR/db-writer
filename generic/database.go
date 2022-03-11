package generic

import "time"

// database types
const (
	Redis     = `redis`
	Neo4j     = `neo4j`
	Couchbase = `couchbase`
)

type Database interface {
	Init(addr, pw string) Database
	Write(values [][]string, opt Options)
}

type Options struct {
	UniqIdx int
	Persist bool
	Expiry  time.Duration
}
