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

package codec

import (
	"io"
	"reflect"

	"github.com/ugorji/go/codec"
)

var defaultJsonHandle = codec.JsonHandle{MapKeyAsString: true}

// DecJson json decode encapsulation
func DecJson(s []byte, v interface{}) error {
	dec := codec.NewDecoderBytes(s, &defaultJsonHandle)
	return dec.Decode(v)
}

// DecJsonReader json reader encapsulation
func DecJsonReader(s io.Reader, v interface{}) error {
	dec := codec.NewDecoder(s, &defaultJsonHandle)
	return dec.Decode(v)
}

// EncJson json encoder
func EncJson(v interface{}, s *[]byte) error {
	enc := codec.NewEncoderBytes(s, &defaultJsonHandle)
	return enc.Encode(v)
}

// EncJsonWriter json writer
func EncJsonWriter(v interface{}, s io.Writer) error {
	enc := codec.NewEncoder(s, &defaultJsonHandle)
	return enc.Encode(v)
}

func init() {
	defaultJsonHandle.MapType = reflect.TypeOf(map[string]interface{}(nil))
}
