package main

import (
	"flag"
	"fmt"
	"github.com/YasiruR/db-writer/generic"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
)

func parseArg() (database, addr, passwd, file, uniqueKey string) {
	db := flag.String(`db`, ``, `database type [OPTIONS: redis, neo4j, couchbase]`)
	hostAddr := flag.String(`host`, ``, `database host address`)
	pwdEnabled := flag.Bool(`pw`, false, `true if password is enabled`)
	data := flag.String(`csv`, ``, `csv file path`)
	key := flag.String(`unique`, ``, `unique key identifier`)

	flag.Parse()

	if *db != generic.Redis && *db != generic.Neo4j && *db != generic.Couchbase {
		log.Fatalln(`invalid database type`)
	}

	var pw string
	if *pwdEnabled {
		pw = getPw()
	}

	return *db, *hostAddr, pw, *data, *key
}

func getPw() (pw string) {
	fmt.Println(`Database password: `)

	raw, err := terminal.MakeRaw(0)
	if err != nil {
		log.Fatalln(err)
	}
	defer terminal.Restore(0, raw)

	var prompt string
	term := terminal.NewTerminal(os.Stdin, prompt)

	pw, err = term.ReadPassword(prompt)
	if err != nil {
		log.Fatalln(err)
	}

	return pw
}
