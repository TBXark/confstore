package confstore

import (
	"encoding/json"
	"fmt"
)

type Codec interface {
	Marshal(value any) ([]byte, error)
	Unmarshal(data []byte, value any) error
}

type JsonCodec struct{}

func (JsonCodec) Marshal(value any) ([]byte, error) {
	return json.MarshalIndent(value, "", "  ")
}

func (JsonCodec) Unmarshal(data []byte, value any) error {
	return json.Unmarshal(data, value)
}

type FallbackCodecGroup struct {
	codecs []Codec
}

func NewCodecGroup(codecs ...Codec) *FallbackCodecGroup {
	return &FallbackCodecGroup{codecs: codecs}
}

func (m *FallbackCodecGroup) Marshal(value any) ([]byte, error) {
	for _, codec := range m.codecs {
		data, err := codec.Marshal(value)
		if err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("failed to marshal value")
}

func (m *FallbackCodecGroup) Unmarshal(data []byte, value any) error {
	for _, codec := range m.codecs {
		err := codec.Unmarshal(data, value)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("failed to unmarshal data")
}
