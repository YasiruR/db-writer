package main

import (
	"flag"
	"fmt"
	"github.com/YasiruR/db-writer/generic"
	log2 "github.com/YasiruR/db-writer/log"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"strings"
)

func parseArg() (dbCfg generic.DBConfigs, dataCfg generic.DataConfigs, file string) {
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

	flag.Parse()

	if *db == `` || *hostAddr == `` || *data == `` {
		log.Fatalln(`null command arguments found`)
	}

	if *db == generic.Redis && *key == `` {
		log.Fatalln(`unique key field should be provided for redis`)
	}

	if *db != generic.Redis && *db != generic.Neo4j && *db != generic.ElasticSearch && *db != generic.ArangoDB {
		log.Fatalln(`invalid database type`)
	}

	if *db == generic.ElasticSearch {
		if *key == `` {
			fmt.Println(`Documents will be indexed iteratively since no unique key is provided`)
		}

		if *caCert == `` {
			fmt.Println(`Provide CA certificate for the access (compulsory from v8 upwards)`)
		}
	}

	if *db == generic.ArangoDB {
		if *table == `` {
			*table = `my_collection`
		}

		if *dbName == `` {
			*dbName = `_system`
		}
	}

	h := hosts(*hostAddr)
	if *pwHide {
		if *pw != `` {
			log.Fatalln(`password has already been provided`)
		}
		*pw = getPw()
	}

	dataCfg = generic.DataConfigs{TableName: *table, Unique: struct {
		Key   string
		Index int
	}{Key: *key, Index: -1}, Limit: *limit}

	return generic.DBConfigs{Typ: *db, Hosts: h, Username: *uname, Passwd: *pw, CACert: *caCert, Name: *dbName}, dataCfg, *data
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
