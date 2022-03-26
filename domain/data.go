package domain

type Data interface {
	MarshalBinary() ([]byte, error)
	JSON(dataCfg DataConfigs) (body string)
	Str() string
}

type DataConfigs struct {
	Table  string
	Fields []string
	Unique struct {
		Key   string
		Index int
	}
	Limit int
}
