package domain

type Data interface {
	MarshalBinary() ([]byte, error)
	JSON(dataCfg DataConfigs) (body string)
	Str() string // todo name and update in other dbs
}

type DataConfigs struct {
	TableName string // todo
	Fields    []string
	Unique    struct {
		Key   string
		Index int
	}
	Limit int
}
