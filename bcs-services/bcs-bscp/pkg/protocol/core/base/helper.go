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

package pbbase

import (
	"errors"
	"fmt"

	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

// Validate the sidecar's version is validate or not.
func (x *Versioning) Validate() error {
	if x.Major <= 0 {
		return errors.New("invalid version major, should >=1")
	}

	return nil
}

// Format the version to literal string, such as 1.0.1
func (x *Versioning) Format() string {
	if x == nil {
		return ""
	}

	return fmt.Sprintf("%d.%d.%d", x.Major, x.Minor, x.Patch)
}

// InvalidArgumentsErr 错误参数返回
func InvalidArgumentsErr(e *InvalidArgument, others ...*InvalidArgument) error {
	errPb, _ := anypb.New(e)
	s := &spb.Status{
		Code:    int32(codes.InvalidArgument),
		Details: []*anypb.Any{errPb},
	}

	othersPb := make([]*anypb.Any, 0, len(others))
	for _, v := range others {
		o, _ := anypb.New(v)
		othersPb = append(othersPb, o)
	}
	s.Details = append(s.Details, othersPb...)

	return status.ErrorProto(s)
}
