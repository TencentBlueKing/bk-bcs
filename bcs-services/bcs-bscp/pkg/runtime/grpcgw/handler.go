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

// Package grpcgw is a grpc-gateway wrapper
package grpcgw

import (
	"bytes"
	"context"
	"net/http"

	"github.com/go-chi/render"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest/view"
)

// bkErrorHandler 蓝鲸规范化的错误返回
func bkErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler,
	w http.ResponseWriter, r *http.Request, err error) {
	ww, ok := w.(*view.GenericResponseWriter)
	if ok {
		ww.SetError(err)
	}

	render.Render(w, r, rest.GRPCErr(err))
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
