/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package marshal

import (
	"encoding/json"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// JSONBcs is a Marshaler which marshals/unmarshals into/from JSON
// the default JSON marshaler is JSONPb, when proto message is in the list,
// use JSONBuiltin.
type JSONBcs struct {
	JSONPb *runtime.JSONPb
}

// ContentType always Returns "application/json".
func (*JSONBcs) ContentType() string {
	return "application/json"
}

// Marshal marshals "v" into JSON
func (j *JSONBcs) Marshal(v interface{}) ([]byte, error) {
	if j.isBuiltin(v) {
		return json.Marshal(v)
	}
	return j.JSONPb.Marshal(v)
}

// Unmarshal unmarshals JSON data into "v".
func (j *JSONBcs) Unmarshal(data []byte, v interface{}) error {
	if j.isBuiltin(v) {
		return json.Unmarshal(data, v)
	}
	return j.JSONPb.Unmarshal(data, v)
}

// NewDecoder returns a Decoder which reads JSON stream from "r".
func (j *JSONBcs) NewDecoder(r io.Reader) runtime.Decoder {
	return j.JSONPb.NewDecoder(r)
}

// NewEncoder returns an Encoder which writes JSON stream into "w".
func (j *JSONBcs) NewEncoder(w io.Writer) runtime.Encoder {
	return j.JSONPb.NewEncoder(w)
}

// Delimiter for newline encoded JSON streams.
func (j *JSONBcs) Delimiter() []byte {
	return []byte("\n")
}

func (j *JSONBcs) isBuiltin(v interface{}) bool {
	p, ok := v.(proto.Message)
	if !ok {
		return false
	}

	name := proto.MessageReflect(p).Descriptor().FullName().Name()
	for _, v := range builtinMessage {
		if v == string(name) {
			return true
		}
	}
	return false
}

var builtinMessage []string = []string{
	"GetResourceSchemaResponse",
	"ListResourceSchemaResponse",
}
