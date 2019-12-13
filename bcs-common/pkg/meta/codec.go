/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
