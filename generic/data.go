package generic

type Data interface {
	MarshalBinary() ([]byte, error)
	JSON(dataCfg DataConfigs) (body string)
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
