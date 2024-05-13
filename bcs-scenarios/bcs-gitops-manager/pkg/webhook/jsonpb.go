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
 */

package webhook

import (
	"fmt"
	"io"
	"reflect"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

var (
	typeOfBytes = reflect.TypeOf([]byte(nil))
	rawJSONMIME = "application/raw-json" // made-up MIME type for our webhook
)

type rawJSONPb struct {
	*runtime.JSONPb
}

// ContentType return the raw json content type
func (*rawJSONPb) ContentType(v interface{}) string {
	return rawJSONMIME
}

// NewDecoder the decoder will read the request raw-json and then reflect
// the json to []byte data
func (*rawJSONPb) NewDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(func(v interface{}) error {
		raw, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		rv := reflect.ValueOf(v)

		if rv.Kind() != reflect.Ptr {
			return fmt.Errorf("%T is not a pointer", v)
		}

		rv = rv.Elem()
		if rv.Type() != typeOfBytes {
			return fmt.Errorf("type must be []byte but got %T", v)
		}

		rv.Set(reflect.ValueOf(raw))
		return nil
	})
}
