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
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"
	"k8s.io/klog/v2"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/criteria/uuid"
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
	lowRidKey         = strings.ToLower(constant.RidKey)
	lowUserKey        = strings.ToLower(constant.UserKey)
	lowACKey          = strings.ToLower(constant.AppCodeKey)
	lowSpaceIDKey     = strings.ToLower(constant.SpaceIDKey)
	lowSpaceTypeIDKey = strings.ToLower(constant.SpaceTypeIDKey)
	lowBizIDKey       = strings.ToLower(constant.BizIDKey)
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

	spaceID := md[lowSpaceIDKey]
	if len(spaceID) != 0 {
		kit.SpaceID = spaceID[0]
	}

	spaceTypeID := md[lowSpaceTypeIDKey]
	if len(spaceTypeID) != 0 {
		kit.SpaceTypeID = spaceTypeID[0]
	}

	bizIDs := md[lowBizIDKey]
	if len(bizIDs) != 0 {
		bizID, err := strconv.ParseUint(bizIDs[0], 10, 64)
		if err != nil {
			klog.ErrorS(err, "parse lowBizID %s", bizIDs[0])
		} else {
			kit.BizID = uint32(bizID)
		}
	}

	kit.Ctx = context.WithValue(kit.Ctx, constant.RidKey, rid)

	// Note: need to add supplier id and authorization field.
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
	AppCode     string
	AppID       uint32 // 对应的应用ID
	BizID       uint32 // 对应的业务ID
	SpaceID     string // 应用对应的SpaceID
	SpaceTypeID string // 应用对应的SpaceTypeID
}

// ContextWithRid NOTES
func (c *Kit) ContextWithRid() context.Context {
	return context.WithValue(c.Ctx, constant.RidKey, c.Rid)
}

// RPCMetaData rpc 头部元数据
func (c *Kit) RPCMetaData() metadata.MD {
	m := map[string]string{
		constant.RidKey:         c.Rid,
		constant.UserKey:        c.User,
		constant.AppCodeKey:     c.AppCode,
		constant.SpaceIDKey:     c.SpaceID,
		constant.SpaceTypeIDKey: c.SpaceTypeID,
		constant.BizIDKey:       strconv.FormatUint(uint64(c.BizID), 10),
	}

	md := metadata.New(m)
	return md
}

// RpcCtx create a new rpc request context, context's metadata is copied current context's metadata info.
func (c *Kit) RpcCtx() context.Context {
	return metadata.NewOutgoingContext(c.Ctx, c.RPCMetaData())
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

// WithKit 封装 kit 到当前的 context
func WithKit(ctx context.Context, kit *Kit) context.Context {
	return context.WithValue(ctx, constant.KitKey, kit)
}

// MustGetKit 从 context 获取 kit, 注意: 如果没有, 会panic, 一般在中间件中使用
func MustGetKit(ctx context.Context) *Kit {
	k, ok := ctx.Value(constant.KitKey).(*Kit)
	if !ok {
		panic(fmt.Errorf("ctx not found kit value"))
	}
	return k
}
