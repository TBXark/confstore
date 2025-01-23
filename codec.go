package confstore

import "encoding/json"

type Codec interface {
	Marshal(value any) ([]byte, error)
	Unmarshal(data []byte, value any) error
}

type JsonCodec struct{}

func (JsonCodec) Marshal(value any) ([]byte, error) {
	return json.Marshal(value)
}

func (JsonCodec) Unmarshal(data []byte, value any) error {
	return json.Unmarshal(data, value)
}
