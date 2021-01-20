/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package json

import (
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

// MarshalPB is used for marshal protobuf object to string.
func MarshalPB(pb proto.Message) (string, error) {
	marshaler := jsonpb.Marshaler{
		EnumsAsInts:  true,
		EmitDefaults: true,
		OrigName:     true,
	}
	return marshaler.MarshalToString(pb)
}

// UnmarshalPB is used for unmarshal protobuf object from string.
func UnmarshalPB(data string, pb proto.Message) error {
	unmarshaler := jsonpb.Unmarshaler{AllowUnknownFields: true}
	return unmarshaler.Unmarshal(strings.NewReader(data), pb)
}
