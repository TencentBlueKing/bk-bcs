package meta

import (
	"encoding/json"
	"fmt"
)

//Encoder define data object encode
type Encoder interface {
	Encode(obj Object) ([]byte, error)
}

//Decoder decode bytes to object
type Decoder interface {
	Decode(data []byte, obj Object) error
}

//Codec combine Encoder & Decoder
type Codec interface {
	Encoder
	Decoder
}

//JsonCodec decode & encode with json
type JsonCodec struct{}

//Encode implements Encoder
func (jc *JsonCodec) Encode(obj Object) ([]byte, error) {
	if obj == nil {
		return nil, fmt.Errorf("nil object")
	}
	return json.Marshal(obj)
}

//Decode implements Decoder
func (jc *JsonCodec) Decode(data []byte, obj Object) error {
	if obj == nil {
		return fmt.Errorf("nil object")
	}
	if len(data) == 0 {
		return fmt.Errorf("empty decode data")
	}
	return json.Unmarshal(data, obj)
}
