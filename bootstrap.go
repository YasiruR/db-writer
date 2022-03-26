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

func banner() {
	fmt.Println(`    ____  ____      _       __     _ __           
   / __ \/ __ )    | |     / /____(_) /____  _____
  / / / / __  |____| | /| / / ___/ / __/ _ \/ ___/
 / /_/ / /_/ /_____/ |/ |/ / /  / / /_/  __/ /    
/_____/_____/      |__/|__/_/  /_/\__/\___/_/     
                                                  `)
}

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

	validate(&dbCfg, &dataCfg, &testCfg, *data)

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

func validate(dbCfg *domain.DBConfigs, dataCfg *domain.DataConfigs, testCfg *domain.TestConfigs, csvPath string) {
	if dbCfg.Typ == `` {
		log.Fatalln(`database type can not be null`)
	}

	if len(dbCfg.Hosts) == 0 {
		log.Fatalln(`host addresses can not be null`)
	}

	if csvPath == `` {
		log.Fatalln(`csv file path should be provided`)
	}

	if dbCfg.Typ == domain.Redis && dataCfg.Unique.Key == `` {
		log.Fatalln(`unique key field should be provided for redis`)
	}

	if dbCfg.Typ != domain.Redis && dbCfg.Typ != domain.Neo4j &&
		dbCfg.Typ != domain.ElasticSearch && dbCfg.Typ != domain.ArangoDB {
		log.Fatalln(`invalid database`)
	}

	if dbCfg.Typ == domain.ElasticSearch {
		if dbCfg.Typ == `` {
			fmt.Println(`Documents will be indexed iteratively since no unique key is provided`)
		}

		if dataCfg.TableName == `` {
			dataCfg.TableName = `my_index`
			fmt.Println(`Index name set as my_table by default since not provided explicitly`)
		}

		if dbCfg.CACert == `` {
			fmt.Println(`Provide CA certificate for the access (compulsory from v8 upwards)`)
		}
	}

	if dbCfg.Typ == domain.ArangoDB {
		if dataCfg.TableName == `` {
			dataCfg.TableName = `my_collection`
			fmt.Println(`Table name set as my_collection by default since not provided explicitly`)
		}

		if dbCfg.Name == `` {
			dbCfg.Name = `_system`
			fmt.Println(`Database name set as _system by default since not provided explicitly`)
		}
	}

	if testCfg.Typ != `` {
		if testCfg.Load == 0 {
			log.Fatalln(`load size should be specified for benchmark test`)
		}

		if testCfg.Typ != domain.BenchmarkRead && testCfg.Typ != domain.BenchmarkWrite && testCfg.Typ != domain.BenchmarkUpdate {
			log.Fatalln(`test type should either be read or write or update (for arangodb only)`)
		}
	}
}
