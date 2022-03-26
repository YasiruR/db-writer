package arangodb

import (
	"bytes"
	"fmt"
	"github.com/YasiruR/db-writer/domain"
	"strings"
)

type data struct {
	body []string
}

func (d data) MarshalBinary() ([]byte, error) {
	return []byte(fmt.Sprintf("%v", d)), nil
}

func (d data) JSON(dataCfg domain.DataConfigs) (body string) {
	var b bytes.Buffer
	b.WriteString(`{`)
	for i, f := range dataCfg.Fields {
		// eliminating escaping and invalid characters of string body to parse as a json
		if strings.Contains(d.body[i], "\n") {
			d.body[i] = strings.ReplaceAll(d.body[i], "\n", " ")
		}

		if strings.Contains(d.body[i], "'") {
			d.body[i] = strings.ReplaceAll(d.body[i], "'", "")
		}

		if strings.Contains(d.body[i], "\"") {
			d.body[i] = strings.ReplaceAll(d.body[i], "\"", "'")
		}

		if strings.Contains(d.body[i], "\\") {
			d.body[i] = strings.ReplaceAll(d.body[i], "\\", "")
		}

		if f == dataCfg.Unique.Key {
			b.WriteString(`"_key" : "`)
		} else {
			b.WriteString(`"` + f + `" : "`)
		}
		b.WriteString(d.body[i] + `"`)

		if i != len(dataCfg.Fields)-1 {
			b.WriteString(`,`)
			b.WriteString("\n")
		}
	}
	b.WriteString("}")

	return b.String()
}

func (d data) Str() string {
	return fmt.Sprintf("%v", d)
}

func (d data) document(dataCfg domain.DataConfigs) map[string]interface{} {
	m := make(map[string]interface{})

	for i, f := range dataCfg.Fields {
		if dataCfg.Unique.Key != `` && f == dataCfg.Unique.Key {
			m["_key"] = d.body[i]
		}
		m[f] = d.body[i]
	}

	return m
}
