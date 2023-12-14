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

// Package webannotation webannotation
package webannotation

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/auth"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

var (
	webAnnotationFuncHub = map[string]AnnotationFunc{}
)

// Perm 类型
type Perm map[string]bool

// Annotation 注解类型
type Annotation struct {
	Perms map[string]Perm `json:"perms"`
}

// AnnotationFunc 函数类型
type AnnotationFunc func(context.Context, *kit.Kit, auth.Authorizer, proto.Message) (*Annotation, error)

// AnnotationInterface 接口
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

// GetAnnotationFunc 获取 msg 已注册的函数
func GetAnnotationFunc(msg proto.Message) (AnnotationFunc, bool) {
	f, ok := webAnnotationFuncHub[name(msg)]
	return f, ok
}
