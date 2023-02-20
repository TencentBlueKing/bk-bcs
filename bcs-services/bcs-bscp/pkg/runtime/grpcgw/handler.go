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

package grpcgw

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"bscp.io/pkg/criteria/constant"
)

// ErrStatus 返回的错误结果体
type errStatus struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      []byte `json:"data"`
}

func errorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	s := status.Convert(err)
	status := &errStatus{Code: 400, Message: s.Message()}
	body, _ := json.Marshal(status)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write(body)
}

type jsonResponse struct {
	runtime.JSONPb
}

func (j *jsonResponse) Marshal(v interface{}) ([]byte, error) {
	buf, err := j.JSONPb.Marshal(v)
	if err != nil {
		return nil, err
	}

	b := fmt.Sprintf(`{"code": 0, "message": "OK", "request_id": "", "data": %s}`, buf)

	return []byte(b), nil
}

// convert http header to grpc metadata
func metadataHandler(ctx context.Context, req *http.Request) metadata.MD {
	return metadata.Pairs(
		constant.RidKey, req.Header.Get(constant.RidKey),
		constant.UserKey, req.Header.Get(constant.UserKey),
		constant.AppCodeKey, req.Header.Get(constant.AppCodeKey),
	)
}
