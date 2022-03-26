package elasticsearch

import (
	"bytes"
	"fmt"
	"github.com/YasiruR/db-writer/domain"
	"strings"
)

type data struct {
	Body []string `json:"body"`
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

		b.WriteString(`"` + f + `" : "`)
		b.WriteString(d.Body[i] + `"`)

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
