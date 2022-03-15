package main

import (
	"flag"
	"fmt"
	"github.com/YasiruR/db-writer/generic"
	log2 "github.com/YasiruR/db-writer/log"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
)

func parseArg() (dbCfg generic.DBConfigs, dataCfg generic.DataConfigs, file string) {
	db := flag.String(`db`, ``, `database type [OPTIONS: redis, neo4j, couchbase]`)
	hostAddr := flag.String(`host`, ``, `database host address`)
	pwdEnabled := flag.Bool(`pw`, false, `true if password is enabled`)
	data := flag.String(`csv`, ``, `csv file path`)
	key := flag.String(`unique`, ``, `unique key identifier`)
	limit := flag.Int(`limit`, -1, `number of data items [maximum if not defined]`)
	caCert := flag.String(`ca`, ``, `CA certificate for elasticsearch`)

	flag.Parse()

	if *db == `` || *hostAddr == `` || *data == `` {
		log.Fatalln(`null command arguments found`)
	}

	if *db == generic.Redis && *key == `` {
		log.Fatalln(`unique key field should be provided for redis`)
	}

	if *db != generic.Redis && *db != generic.Neo4j && *db != generic.ElasticSearch {
		log.Fatalln(`invalid database type`)
	}

	var pw string
	if *pwdEnabled {
		pw = getPw()
	}

	return generic.DBConfigs{Typ: *db, Addr: *hostAddr, Passwd: pw, CACert: *caCert}, generic.DataConfigs{UniqKey: *key, Limit: *limit}, *data
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
