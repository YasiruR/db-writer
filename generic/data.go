package generic

type Data interface {
	MarshalBinary() ([]byte, error)
	JSON(dataCfg DataConfigs) (body string)
}

type DataConfigs struct {
	TableName string
	Fields    []string
	UniqKey   string // todo combine
	UniqIdx   int
	Limit     int
}
