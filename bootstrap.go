package main

import (
	"flag"
	"fmt"
	"github.com/YasiruR/db-writer/domain"
	log2 "github.com/YasiruR/db-writer/log"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"strings"
)

func parseArg() (dbCfg domain.DBConfigs, dataCfg domain.DataConfigs, testCfg domain.TestConfigs, file string) {
	db := flag.String(`db`, ``, `database type [OPTIONS: redis, neo4j, couchbase]`)
	hostAddr := flag.String(`host`, ``, `database host address`)
	uname := flag.String(`uname`, ``, `database username if login is required`)
	pw := flag.String(`pw`, ``, `database password (use pwhide for sensitive cases)`)
	data := flag.String(`csv`, ``, `csv file path`)
	key := flag.String(`unique`, ``, `unique key identifier`)
	limit := flag.Int(`limit`, -1, `number of data items [maximum if not defined]`)
	caCert := flag.String(`ca`, ``, `CA certificate file path for elasticsearch`)
	pwHide := flag.Bool(`pwhide`, false, `[OPTIONAL] to enter password in hidden format (true/false)`)
	table := flag.String(`table`, ``, `collection name for arangodb ['my_collection' will be used if omitted]`)
	dbName := flag.String(`dbname`, ``, `database name for arangodb [_system will be used if omitted]`)
	testType := flag.String(`benchmark`, ``, `functionality to be tested with a load (read/write)`)
	loadSize := flag.Int(`load`, 0, `batch size of the benchmark test`)

	flag.Parse()
	fmt.Println()

	// todo add validate func

	if *db == `` || *hostAddr == `` || *data == `` {
		log.Fatalln(`null command arguments found`)
	}

	if *db == domain.Redis && *key == `` {
		log.Fatalln(`unique key field should be provided for redis`)
	}

	if *db != domain.Redis && *db != domain.Neo4j && *db != domain.ElasticSearch && *db != domain.ArangoDB {
		log.Fatalln(`invalid database type`)
	}

	if *db == domain.ElasticSearch {
		if *key == `` {
			fmt.Println(`Documents will be indexed iteratively since no unique key is provided`)
		}

		if *caCert == `` {
			fmt.Println(`Provide CA certificate for the access (compulsory from v8 upwards)`)
		}
	}

	if *db == domain.ArangoDB {
		if *table == `` {
			*table = `my_collection`
		}

		if *dbName == `` {
			*dbName = `_system`
		}
	}

	if *testType != `` {
		if *loadSize == 0 {
			log.Fatalln(`load size should be specified for benchmark test`)
		}

		if *testType != domain.BenchmarkRead && *testType != domain.BenchmarkWrite {
			log.Fatalln(`test type should either be read or write`)
		}
	}

	h := hosts(*hostAddr)
	if *pwHide {
		if *pw != `` {
			log.Fatalln(`password has already been provided`)
		}
		*pw = getPw()
	}

	dataCfg = domain.DataConfigs{TableName: *table, Unique: struct {
		Key   string
		Index int
	}{Key: *key, Index: -1}, Limit: *limit}

	dbCfg = domain.DBConfigs{Typ: *db, Hosts: h, Username: *uname, Passwd: *pw, CACert: *caCert, Name: *dbName}
	testCfg = domain.TestConfigs{Database: *db, Typ: *testType, Load: *loadSize}

	return dbCfg, dataCfg, testCfg, *data
}

func getPw() (pw string) {
	fmt.Println(`Database password: `)

	raw, err := terminal.MakeRaw(0)
	if err != nil {
		log2.Fatal(err)
	}
	defer terminal.Restore(0, raw)

	var prompt string
	term := terminal.NewTerminal(os.Stdin, prompt)

	pw, err = term.ReadPassword(prompt)
	if err != nil {
		log2.Fatal(err)
	}

	return pw
}

func hosts(arg string) []string {
	list := strings.Split(arg, `,`)
	if len(list) == 0 {
		log.Fatalln(`host list is empty`)
	}

	return list
}
