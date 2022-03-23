package arangodb

import (
	"bytes"
	"fmt"
	"github.com/YasiruR/db-writer/domain"
	"strings"
)

type data struct {
	Body []string
}

func (d data) MarshalBinary() ([]byte, error) {
	return []byte(fmt.Sprintf("%v", d)), nil
}

func (d data) JSON(dataCfg domain.DataConfigs) (body string) {
	var b bytes.Buffer
	b.WriteString(`{`)
	for i, f := range dataCfg.Fields {
		// eliminating escaping and invalid characters of string body to parse as a json
		if strings.Contains(d.Body[i], "\n") {
			d.Body[i] = strings.ReplaceAll(d.Body[i], "\n", " ")
		}

		if strings.Contains(d.Body[i], "'") {
			d.Body[i] = strings.ReplaceAll(d.Body[i], "'", "")
		}

		if strings.Contains(d.Body[i], "\"") {
			d.Body[i] = strings.ReplaceAll(d.Body[i], "\"", "'")
		}

		if strings.Contains(d.Body[i], "\\") {
			d.Body[i] = strings.ReplaceAll(d.Body[i], "\\", "")
		}

		if f == dataCfg.Unique.Key {
			b.WriteString(`"_key" : "`)
		} else {
			b.WriteString(`"` + f + `" : "`)
		}
		b.WriteString(d.Body[i] + `"`)

		if i != len(dataCfg.Fields)-1 {
			b.WriteString(`,`)
			b.WriteString("\n")
		}
	}
	b.WriteString("}")

	return b.String()
}

func (d data) Map(dataCfg domain.DataConfigs) map[string]interface{} {
	m := make(map[string]interface{})

	for i, f := range dataCfg.Fields {
		m[f] = d.Body[i]
	}

	return m
}
