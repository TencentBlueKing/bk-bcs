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

package view

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/tidwall/sjson"
	"google.golang.org/protobuf/proto"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/auth"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest/view/modifier"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest/view/webannotation"
)

const (
	// BK_APIv1_CODE_KEY 蓝鲸规范返回的 code key
	BK_APIv1_CODE_KEY = "code"

	// BK_APIv1_CODE_OK_VALUE 蓝鲸规范返回正常请求的 code value
	BK_APIv1_CODE_OK_VALUE = 0
)

// DataStructInterface 判断是否已经包含 data 结构体实现, 处理 structpb.Struct 问题
type DataStructInterface interface {
	IsDataStruct() bool
}

// GenericResponseWriter 自定义Write，自动补充 data 和 web_annotations 数据
type GenericResponseWriter struct {
	http.ResponseWriter

	isDataStruct bool
	ctx          context.Context
	msg          proto.Message
	authorizer   auth.Authorizer
	annotation   *webannotation.Annotation
	err          error // low-level runtime error
}

// Write http write 接口实现
func (w *GenericResponseWriter) Write(data []byte) (int, error) {
	// 错误不需要特殊处理
	if w.err != nil {
		return w.ResponseWriter.Write(data)
	}

	// data struct 类型不需要处理
	if w.isDataStruct {
		// data 需要是合法的 json 格式, 蓝鲸老的规范需要添加 code
		if ndata, err := sjson.SetBytes(data, BK_APIv1_CODE_KEY, BK_APIv1_CODE_OK_VALUE); err != nil {
			return w.ResponseWriter.Write(ndata)
		}
		return w.ResponseWriter.Write(data)
	}

	if err := w.beforeWriteHook(w.ctx, w.msg); err != nil {
		return 0, err
	}

	if w.msg != nil {
		w, ok := w.msg.(modifier.RespModifier)
		if ok {
			var err error
			data, err = w.ModifyResp(data)
			if err != nil {
				return 0, err
			}
		}
	}

	buf := bytes.NewBufferString(`{"data":`)
	buf.Write(data)

	if w.annotation != nil {
		abody, err := json.Marshal(w.annotation)
		if err != nil {
			return 0, err
		}

		buf.WriteString(`,"web_annotations":`)
		buf.Write(abody)
	}
	buf.WriteString("}")

	return w.ResponseWriter.Write(buf.Bytes())
}

// beforeWriteHook is a hook before write response
func (w *GenericResponseWriter) beforeWriteHook(ctx context.Context, msg proto.Message) error {
	return w.BuildWebAnnotation(ctx, msg)
}

// BuildWebAnnotation 动态执行 webannotions 函数
func (w *GenericResponseWriter) BuildWebAnnotation(ctx context.Context, msg proto.Message) error {
	// when not using grpc-gateway
	if ctx == nil {
		return nil
	}

	kt := kit.MustGetKit(ctx)

	var (
		annotation *webannotation.Annotation
		err        error
	)

	// 优先 interface 模式
	iface, ok := msg.(webannotation.AnnotationInterface)
	if ok {
		annotation, err = iface.Annotation(ctx, kt, w.authorizer)
	} else {
		// 注册模式
		f, ok := webannotation.GetAnnotationFunc(msg)
		if ok {
			annotation, err = f(ctx, kt, w.authorizer, msg)
		}
	}

	if err != nil {
		return err
	}

	if annotation != nil {
		w.annotation = annotation
	}

	return nil
}

// SetWriterAttrs set attributes of the writer
func (w *GenericResponseWriter) SetWriterAttrs(ctx context.Context, msg proto.Message) error {
	w.ctx = ctx
	w.msg = msg
	return nil
}

// SetError 设置错误请求
func (w *GenericResponseWriter) SetError(err error) {
	w.err = err
}

// SetDataStructFlag 设置是否是 DataStruct 类型
func (w *GenericResponseWriter) SetDataStructFlag(ok bool) {
	w.isDataStruct = ok
}

// NewGenericResponseWriter GenericResponseWriter初始化
func NewGenericResponseWriter(w http.ResponseWriter, authorizer auth.Authorizer) *GenericResponseWriter {
	return &GenericResponseWriter{authorizer: authorizer, ResponseWriter: w}
}

// Generic http 中间件, 返回蓝鲸规范的数据结构
func Generic(authorizer auth.Authorizer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := NewGenericResponseWriter(w, authorizer)
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
