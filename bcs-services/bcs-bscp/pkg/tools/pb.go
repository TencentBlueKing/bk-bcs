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

package tools

import (
	"encoding/json"

	pbstruct "github.com/golang/protobuf/ptypes/struct"
)

// UnmarshalFromPbStructToMap parsed a map from pb struct message.
func UnmarshalFromPbStructToMap(st *pbstruct.Struct) (map[string]interface{}, error) {
	bytes, err := st.MarshalJSON()
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	if err = json.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}

	return m, nil
}

// MarshalMapToPbStruct marshal map to proto struct field type.
func MarshalMapToPbStruct(m map[string]interface{}) (*pbstruct.Struct, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	st := new(pbstruct.Struct)
	if err = st.UnmarshalJSON(bytes); err != nil {
		return nil, err
	}

	return st, nil
}
