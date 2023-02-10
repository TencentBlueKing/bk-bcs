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

// Package kit NOTES
package kit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/criteria/uuid"

	"google.golang.org/grpc/metadata"
)

// New initial a kit with rid and context.
func New() *Kit {
	rid := uuid.UUID()
	return &Kit{
		Rid: rid,
		Ctx: context.WithValue(context.TODO(), constant.RidKey, rid),
	}
}

var (
	lowRidKey  = strings.ToLower(constant.RidKey)
	lowUserKey = strings.ToLower(constant.UserKey)
	lowACKey   = strings.ToLower(constant.AppCodeKey)
)

// FromGrpcContext used only to obtain Kit through grpc context.
func FromGrpcContext(ctx context.Context) *Kit {
	kit := &Kit{
		Ctx: ctx,
	}

	md, _ := metadata.FromIncomingContext(ctx)
	rid := md[lowRidKey]
	if len(rid) != 0 {
		kit.Rid = rid[0]
	} else {
		kit.Rid = "bscp-" + uuid.UUID()
	}

	user := md[lowUserKey]
	if len(user) != 0 {
		kit.User = user[0]
	}

	appCode := md[lowACKey]
	if len(appCode) != 0 {
		kit.AppCode = appCode[0]
	}

	kit.Ctx = context.WithValue(kit.Ctx, constant.RidKey, rid)

	// TODO: need to add supplier id and authorization field.
	return kit
}

// User 用户信息
type User struct {
	Username  string `json:"username"`
	AvatarUrl string `json:"avatar_url"`
}

// Kit defines the basic metadata info within a task.
type Kit struct {
	// Ctx is request context.
	Ctx context.Context

	// User's name.
	User string

	// Rid is request id.
	Rid string

	// AppCode is app code.
	AppCode string
}

// ContextWithRid NOTES
func (c *Kit) ContextWithRid() context.Context {
	return context.WithValue(c.Ctx, constant.RidKey, c.Rid)
}

// RpcCtx create a new rpc request context, context's metadata is copied current context's metadata info.
func (c *Kit) RpcCtx() context.Context {
	md := metadata.Pairs(
		constant.RidKey, c.Rid,
		constant.UserKey, c.User,
		constant.AppCodeKey, c.AppCode,
	)
	return metadata.NewOutgoingContext(c.Ctx, md)
}

// CtxWithTimeoutMS create a new context with basic info and timout configuration.
func (c *Kit) CtxWithTimeoutMS(timeoutMS int) context.CancelFunc {
	ctx := context.WithValue(context.TODO(), constant.RidKey, c.Rid)
	var cancel context.CancelFunc
	c.Ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMS)*time.Millisecond)
	return cancel
}

// Validate context kit.
func (c *Kit) Validate() error {
	if c.Ctx == nil {
		return errors.New("context is required")
	}

	if len(c.User) == 0 {
		return errors.New("user is required")
	}

	ridLen := len(c.Rid)
	if ridLen == 0 {
		return errors.New("rid is required")
	}

	if ridLen < 16 || ridLen > 48 {
		return errors.New("rid length not right, length should in 16~48")
	}

	if len(c.AppCode) == 0 {
		return errors.New("app code is required")
	}

	return nil
}

// ValidateBase validate basic kit info
func (c *Kit) ValidateBase() error {
	if c.Ctx == nil {
		return errors.New("context is required")
	}

	if len(c.User) == 0 {
		return errors.New("user is required")
	}

	ridLen := len(c.Rid)
	if ridLen == 0 {
		return errors.New("rid is required")
	}

	if ridLen < 16 || ridLen > 48 {
		return errors.New("rid length not right, length should in 16~48")
	}

	return nil
}

// Vas convert kit to vas
func (c *Kit) Vas() *Vas {
	return &Vas{
		Rid: c.Rid,
		Ctx: c.Ctx,
	}
}

func WithKit(ctx context.Context, kit *Kit) context.Context {
	return context.WithValue(ctx, constant.KitKey, kit)
}

func MustGetKit(ctx context.Context) *Kit {
	k, ok := ctx.Value(constant.KitKey).(*Kit)
	if !ok {
		panic(fmt.Errorf("ctx not found kit value"))
	}
	return k
}
