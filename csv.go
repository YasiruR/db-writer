package main

import (
	"encoding/csv"
	"github.com/YasiruR/db-writer/generic"
	"log"
	"os"
)

func readData(file string, dataCfg *generic.DataConfigs) (fields []string, values [][]string) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalln(err, file)
	}

	defer f.Close()
	uniqIdx := -1
	r := csv.NewReader(f)

	fields, err = r.Read()
	if err != nil {
		log.Fatalln(err)
	}

	values, err = r.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}

	if dataCfg.UniqKey == `` {
		return
	}

	for i, field := range fields {
		if field == dataCfg.UniqKey {
			uniqIdx = i
			break
		}
	}

	dataCfg.UniqIdx = uniqIdx
	return
}
