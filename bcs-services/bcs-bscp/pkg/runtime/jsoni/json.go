/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package jsoni

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Marshal json with json iter
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalIndent marshal the object to json with indent.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

// Unmarshal json with json iter
func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// UnmarshalFromString unmarshal json to object with string.
func UnmarshalFromString(str string, v interface{}) error {
	return json.UnmarshalFromString(str, v)
}

// Decoder return a json decoder to decode to json.
func Decoder(reader io.Reader) *jsoniter.Decoder {
	return json.NewDecoder(reader)
}

// Encoder return a json encoder to encode a struct to json.
func Encoder(writer io.Writer) *jsoniter.Encoder {
	return json.NewEncoder(writer)
}
