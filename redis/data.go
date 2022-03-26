package redis

import (
	"fmt"
	"github.com/YasiruR/db-writer/domain"
)

type data struct {
	body []string
}

func (d data) MarshalBinary() ([]byte, error) {
	return []byte(fmt.Sprintf("%v", d)), nil
}

func (d data) JSON(_ domain.DataConfigs) (body string) {
	return ""
}

func (d data) Str() string {
	return fmt.Sprintf("%v", d)
}
