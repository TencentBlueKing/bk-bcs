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

package selector

import (
	"encoding/json"

	pbstruct "github.com/golang/protobuf/ptypes/struct"
)

// UnmarshalStrategyFromPbStruct parsed a strategy from pb struct message.
func UnmarshalStrategyFromPbStruct(st *pbstruct.Struct) (*Selector, error) {
	bytes, err := st.MarshalJSON()
	if err != nil {
		return nil, err
	}
	s := new(Selector)
	if err = json.Unmarshal(bytes, &s); err != nil {
		return nil, err
	}

	// validate the strategy.
	if err := s.Validate(); err != nil {
		return nil, err
	}

	return s, nil
}

// MarshalPbStructStrategyToJSONRaw marshal a pb struct message to a standard strategy json raw string.
func MarshalPbStructStrategyToJSONRaw(st *pbstruct.Struct) (string, error) {
	bytes, err := st.MarshalJSON()
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
