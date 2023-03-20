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

package webannotation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/protobuf/proto"

	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/kit"
)

var (
	webAnnotationFuncHub = map[string]AnnotationFunc{}
)

// Perm
type Perm map[string]bool

// Annotation 注解类型
type Annotation struct {
	Perms map[string]Perm `json:"perms"`
}

// AnnotationFunc
type AnnotationFunc func(context.Context, *kit.Kit, auth.Authorizer, proto.Message) (*Annotation, error)

// AnnotationInterface
type AnnotationInterface interface {
	Annotation(context.Context, *kit.Kit, auth.Authorizer) (*Annotation, error)
}

// name 类型唯一名称
func name(msg proto.Message) string {
	name := proto.MessageName(msg)
	return string(name)
}

// Register 注册，部分为防止循环引用使用这种方式
func Register(msg proto.Message, f AnnotationFunc) {
	_, ok := webAnnotationFuncHub[name(msg)]
	if ok {
		panic(fmt.Errorf("%s duplicate registration", name(msg)))
	}

	webAnnotationFuncHub[name(msg)] = f
}

// AnnotationResponseWriter
type AnnotationResponseWriter struct {
	http.ResponseWriter
	authorizer auth.Authorizer
	annotation *Annotation
	err        error // low-level runtime error
}

// NewWrapResponseWriter
func NewWrapResponseWriter(w http.ResponseWriter, authorizer auth.Authorizer) *AnnotationResponseWriter {
	return &AnnotationResponseWriter{authorizer: authorizer, ResponseWriter: w}
}

// Write http write 接口实现
func (w *AnnotationResponseWriter) Write(data []byte) (int, error) {
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
func (w *AnnotationResponseWriter) Build(ctx context.Context, msg proto.Message) error {
	kt := kit.MustGetKit(ctx)

	var (
		annotation *Annotation
		err        error
	)

	// 优先 interface 模式
	iface, ok := msg.(AnnotationInterface)
	if ok {
		annotation, err = iface.Annotation(ctx, kt, w.authorizer)
	} else {
		// 注册模式
		f, ok := webAnnotationFuncHub[name(msg)]
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
func (w *AnnotationResponseWriter) SetError(err error) {
	w.err = err
}
