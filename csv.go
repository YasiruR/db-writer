package main

import (
	"encoding/csv"
	"github.com/YasiruR/db-writer/domain"
	"log"
	"os"
)

func readData(file string, dataCfg *domain.DataConfigs) (values [][]string) {
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

	if dataCfg.Unique.Key == `` {
		dataCfg.Unique.Index = -1
		return
	}

	for i, field := range dataCfg.Fields {
		if field == dataCfg.Unique.Key {
			uniqIdx = i
			break
		}
	}

	dataCfg.Unique.Index = uniqIdx
	return
}
