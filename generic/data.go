package generic

import "fmt"

type Data struct {
	Body []string
}

func (d Data) MarshalBinary() ([]byte, error) {
	return []byte(fmt.Sprintf("%v", d)), nil
}

func (d Data) String() string {
	//return fmt.Sprintf("{%v}", d.Body)

	return fmt.Sprintf(`{"data": "%s"}`, d.Body[0])
}
