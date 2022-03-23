package tester

import (
	"encoding/csv"
	"fmt"
	"github.com/YasiruR/db-writer/domain"
	"github.com/YasiruR/db-writer/log"
	"io"
	"os"
	"strconv"
)

func Output(cfg domain.TestConfigs, successCount uint64, totalElapsedTime, aggrLatency uint64, persist bool) {

	fmt.Println("========= Load Test Results (Read) ==========")
	fmt.Println("successful reads: ", successCount)
	fmt.Println("total time taken (micro seconds): ", totalElapsedTime)
	fmt.Println("throughput (req/s) : ", successCount*1e6/totalElapsedTime) // todo check if success or total
	fmt.Println("average latency (micro seconds): ", float64(aggrLatency)/float64(successCount*1e3))

	if persist == false {
		return
	}

	writeToCsv(cfg, successCount, totalElapsedTime, aggrLatency)
}

func writeToCsv(cfg domain.TestConfigs, successCount uint64, totalElapsedTime, aggrLatency uint64) {
	dir := `./data/`
	fileName := dir + `results_` + cfg.Database + `.csv`

	// create directory is not exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	var data [][]string
	var f *os.File
	defer f.Close()

	if fileExists(fileName) {
		f, err := os.Open(fileName)
		if err != nil {
			log.Error(err)
		}

		r := csv.NewReader(f)
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatal(err)
			}

			data = append(data, record)
		}
	} else {
		data = append(data, []string{`database`, `operation`, `load`, `success count`, `total test duration (micro seconds)`,
			`aggregated latency (micro seconds)`, `throughput (req/s)`, ` average latency (ms)`})
	}

	f, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	w := csv.NewWriter(f)

	// appending new data
	data = append(data, []string{cfg.Database, cfg.Typ, strconv.Itoa(cfg.Load), strconv.Itoa(int(successCount)),
		strconv.Itoa(int(totalElapsedTime)), strconv.Itoa(int(aggrLatency)), strconv.Itoa(int(successCount * 1e6 / totalElapsedTime)),
		strconv.FormatFloat(float64(aggrLatency)/float64(successCount*1e3), 'f', -1, 64)})

	err = w.WriteAll(data)
	if err != nil {
		log.Fatal(err)
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
