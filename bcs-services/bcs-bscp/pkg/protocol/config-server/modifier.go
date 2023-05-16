/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package pbcs

import (
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// ModifyResp implements interface modifier，it converts byte_size type from string to int64
// see grpc-gateway issue: https://github.com/grpc-ecosystem/grpc-gateway/issues/296
func (r *ListContentsResp) ModifyResp(resp []byte) ([]byte, error) {
	js := string(resp)
	result := gjson.Get(js, "details.#.spec.byte_size")
	if !result.Exists() {
		return nil, fmt.Errorf("can't find json path details.#.spec.byte_size in response")
	}

	destJs := js
	rs := result.Array()
	var err error
	for i, r := range rs {
		// convert byte_size type from string to int64
		destJs, err = sjson.Set(destJs, fmt.Sprintf("details.%d.spec.byte_size", i), r.Int())
		if err != nil {
			return nil, err
		}
	}
	return []byte(destJs), nil
}

// ModifyResp implements interface modifier，it converts byte_size type from string to int64
// see grpc-gateway issue: https://github.com/grpc-ecosystem/grpc-gateway/issues/296
func (r *ListCommitsResp) ModifyResp(resp []byte) ([]byte, error) {
	js := string(resp)
	result := gjson.Get(js, "details.#.spec.content.byte_size")
	if !result.Exists() {
		return nil, fmt.Errorf("can't find json path details.#.spec.content.byte_size in response")
	}

	destJs := js
	rs := result.Array()
	var err error
	for i, r := range rs {
		// convert byte_size type from string to int64
		destJs, err = sjson.Set(destJs, fmt.Sprintf("details.%d.spec.content.byte_size", i), r.Int())
		if err != nil {
			return nil, err
		}
	}
	return []byte(destJs), nil
}
