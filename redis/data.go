package redis

import (
	"fmt"
	"github.com/YasiruR/db-writer/generic"
)

type data struct {
	body []string
}

func (d data) MarshalBinary() ([]byte, error) {
	return []byte(fmt.Sprintf("%v", d)), nil
}

func (d data) JSON(_ generic.DataConfigs) (body string) {
	return ""
}