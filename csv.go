package main

import (
	"encoding/csv"
	"github.com/YasiruR/db-writer/generic"
	"log"
	"os"
)

func readData(file string, dataCfg *generic.DataConfigs) (values [][]string) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalln(err, file)
	}

	defer f.Close()
	uniqIdx := -1
	r := csv.NewReader(f)

	dataCfg.Fields, err = r.Read()
	if err != nil {
		log.Fatalln(err)
	}

	values, err = r.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}

	if dataCfg.UniqKey == `` {
		dataCfg.UniqIdx = -1
		return
	}

	for i, field := range dataCfg.Fields {
		if field == dataCfg.UniqKey {
			uniqIdx = i
			break
		}
	}

	dataCfg.UniqIdx = uniqIdx
	return
}
