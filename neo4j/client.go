package neo4j

import (
	"fmt"
	"github.com/YasiruR/db-writer/generic"
	goNeo4j "github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"log"
	"sync/atomic"
)

const bufferSize = 100

type neo4j struct {
	db        goNeo4j.Driver
	tx        string
	paramChan chan map[string]interface{}
}

func Client() generic.Database {
	return &neo4j{paramChan: make(chan map[string]interface{}, bufferSize)}
}

func (n *neo4j) Init(cfg generic.DBConfigs) generic.Database {
	db, err := goNeo4j.NewDriver(cfg.Addr, goNeo4j.BasicAuth(cfg.Username, cfg.Passwd, ``))
	if err != nil {
		log.Fatalln(err)
	}

	n.db = db
	return n
}

func (n *neo4j) Write(fields []string, values [][]string, _ generic.Options) {
	var failed uint64
	n.setTx(fields)

	for _, val := range values {
		go func(val []string) {
			session := n.db.NewSession(goNeo4j.SessionConfig{})
			defer session.Close()
			go n.sendParams(fields, val)

			_, err := session.WriteTransaction(n.insertFunc)
			if err != nil {
				atomic.AddUint64(&failed, 1)
			}
		}(val)
	}
}

func (n *neo4j) setTx(fields []string) {
	tx := fmt.Sprintf("CREATE (n:Item {")
	for i, f := range fields {
		if i == len(fields)-1 {
			tx += fmt.Sprintf(` %s: $%s })`, f, f)
			continue
		}
		tx += fmt.Sprintf(` %s: $%s,`, f, f)
	}

	n.tx = tx
}

func (n *neo4j) sendParams(fields []string, val []string) {
	paramMap := make(map[string]interface{})
	for i, f := range fields {
		paramMap[f] = val[i]
	}

	n.paramChan <- paramMap
}

func (n *neo4j) insertFunc(tx goNeo4j.Transaction) (interface{}, error) {
	paramMap := <-n.paramChan
	_, err := tx.Run(n.tx, paramMap)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
