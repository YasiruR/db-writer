package main

import (
	"encoding/csv"
	"log"
	"os"
)

func readData(fileName string) (fields []string, values [][]string) {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalln(err)
	}

	defer f.Close()
	r := csv.NewReader(f)

	fields, err = r.Read()
	if err != nil {
		log.Fatalln(err)
	}

	values, err = r.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}

	return
}
