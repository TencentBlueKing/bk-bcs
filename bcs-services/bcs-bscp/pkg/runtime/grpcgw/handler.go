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
	"bytes"
	"context"
	"net/http"

	"github.com/go-chi/render"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"bscp.io/pkg/kit"
	"bscp.io/pkg/rest"
	"bscp.io/pkg/rest/view"
)

var (
	// grpcCodeMap 蓝鲸 Code 映射
	grpcCodeMap = map[codes.Code]string{
		codes.Canceled:           "CANCELLED",
		codes.Unknown:            "UNKNOWN",
		codes.InvalidArgument:    "INVALID_ARGUMENT",
		codes.DeadlineExceeded:   "DEADLINE_EXCEEDED",
		codes.NotFound:           "NOT_FOUND",
		codes.AlreadyExists:      "ALREADY_EXISTS",
		codes.PermissionDenied:   "PERMISSION_DENIED",
		codes.ResourceExhausted:  "RESOURCE_EXHAUSTED",
		codes.FailedPrecondition: "FAILED_PRECONDITION",
		codes.Aborted:            "ABORTED",
		codes.OutOfRange:         "OUT_OF_RANGE",
		codes.Unimplemented:      "UNIMPLEMENTED",
		codes.Internal:           "INTERNAL",
		codes.Unavailable:        "UNAVAILABLE",
		codes.DataLoss:           "DATA_LOSS",
		codes.Unauthenticated:    "UNAUTHENTICATED",
	}

	// grpcCodeMap 蓝鲸 status 映射
	grpcHttpStatusMap = map[codes.Code]int{
		codes.Canceled:           http.StatusBadRequest,
		codes.Unknown:            http.StatusBadRequest,
		codes.InvalidArgument:    http.StatusBadRequest,
		codes.DeadlineExceeded:   http.StatusBadRequest,
		codes.NotFound:           http.StatusNotFound,
		codes.AlreadyExists:      http.StatusBadRequest,
		codes.PermissionDenied:   http.StatusForbidden,
		codes.ResourceExhausted:  http.StatusBadRequest,
		codes.FailedPrecondition: http.StatusBadRequest,
		codes.Aborted:            http.StatusBadRequest,
		codes.OutOfRange:         http.StatusBadRequest,
		codes.Unimplemented:      http.StatusBadRequest,
		codes.Internal:           http.StatusBadRequest,
		codes.Unavailable:        http.StatusBadRequest,
		codes.DataLoss:           http.StatusBadRequest,
		codes.Unauthenticated:    http.StatusUnauthorized,
	}
)

// grpcGatewayErr GRPC-Gateway 错误
func grpcGatewayErr(s *status.Status) render.Renderer {
	code := grpcCodeMap[s.Code()]
	if code == "" {
		code = "INVALID_REQUEST"
	}

	status := grpcHttpStatusMap[s.Code()]
	if status == 0 {
		status = http.StatusBadRequest
	}

	payload := &rest.ErrorPayload{Code: code, Message: s.Message(), Details: s.Details()}
	return &rest.ErrorResponse{Error: payload, HTTPStatusCode: status}
}

// bkErrorHandler 蓝鲸规范化的错误返回
func bkErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	s := status.Convert(err)
	ww, ok := w.(*view.GenericResponseWriter)
	if ok {
		ww.SetError(err)
	}

	render.Render(w, r, grpcGatewayErr(s))
}

// bkJSONResponse 蓝鲸规范返回
type bkJSONResponse struct {
	runtime.JSONPb
}

// Marshal 蓝鲸规范序列化, 外层统一添加 {"data": %s} 结构
func (j *bkJSONResponse) Marshal(v interface{}) ([]byte, error) {
	body, err := j.JSONPb.Marshal(v)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBufferString(`{"data":`)
	buf.Write(body)
	buf.WriteString(`}`)

	return buf.Bytes(), nil
}

// kitMetadataHandler convert http header to grpc metadata
func kitMetadataHandler(ctx context.Context, r *http.Request) metadata.MD {
	kt := kit.MustGetKit(ctx)
	return kt.RPCMetaData()
}

// bscpResponse 可动态处理 webannotation
func bscpResponse(ctx context.Context, w http.ResponseWriter, msg proto.Message) error {
	ww, ok := w.(*view.GenericResponseWriter)
	if !ok {
		return nil
	}

	if d, ok := msg.(view.DataStructInterface); ok {
		ww.SetDataStructFlag(d.IsDataStruct())
	}

	return ww.SetWriterAttrs(ctx, msg)
}
