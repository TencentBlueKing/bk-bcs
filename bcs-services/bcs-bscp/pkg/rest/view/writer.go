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
package view

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"google.golang.org/protobuf/proto"

	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/rest/view/webannotation"
)

// GenericResponseWriter
type GenericResponseWriter struct {
	http.ResponseWriter
	authorizer auth.Authorizer
	annotation *webannotation.Annotation
	err        error // low-level runtime error
}

// Write http write 接口实现
func (w *GenericResponseWriter) Write(data []byte) (int, error) {
	// 错误不需要特殊处理
	if w.err != nil {
		return w.ResponseWriter.Write(data)
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

// Build 动态执行 webannotions 函数
func (w *GenericResponseWriter) BuildWebAnnotation(ctx context.Context, msg proto.Message) error {
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

// SetError 设置错误请求
func (w *GenericResponseWriter) SetError(err error) {
	w.err = err
}

// NewGenericResponseWriter
func NewGenericResponseWriter(w http.ResponseWriter, authorizer auth.Authorizer) *GenericResponseWriter {
	return &GenericResponseWriter{authorizer: authorizer, ResponseWriter: w}
}
