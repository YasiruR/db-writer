package main

import (
	"encoding/csv"
	"log"
	"os"
)

func readData(file, uniqKey string) (uniqIdx int, fields []string, values [][]string) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalln(err, file)
	}

	defer f.Close()
	uniqIdx = -1
	r := csv.NewReader(f)

	fields, err = r.Read()
	if err != nil {
		log.Fatalln(err)
	}

	values, err = r.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}

	if uniqKey == `` {
		return
	}

	for i, field := range fields {
		if field == uniqKey {
			uniqIdx = i
			break
		}
	}

	return
}
