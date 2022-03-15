package generic

import (
	"encoding/json"
	"fmt"
	"github.com/YasiruR/db-writer/log"
)

type Data struct {
	Body []string `json:"body"`
}

func (d Data) MarshalBinary() ([]byte, error) {
	return []byte(fmt.Sprintf("%v", d)), nil
}

func (d Data) String() string {
	//return fmt.Sprintf("{%v}", d.Body)

	//return fmt.Sprintf(`{"data": "%s"}`, d.Body[0])

	data, err := json.Marshal(d)
	if err != nil {
		log.Fatal(err)
	}

	return string(data)
}
