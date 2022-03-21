/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package handler

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

type mockSubscribeStream struct{}

func (x *mockSubscribeStream) Context() context.Context {
	panic("implement me")
}

func (x *mockSubscribeStream) SendMsg(i interface{}) error {
	panic("implement me")
}

func (x *mockSubscribeStream) RecvMsg(i interface{}) error {
	panic("implement me")
}

func (x *mockSubscribeStream) Close() error {
	panic("implement me")
}

// 目前单测中仅使用该方法，可按需实现其他方法的 Mock
func (x *mockSubscribeStream) Send(m *clusterRes.SubscribeResp) error {
	return errorx.New(errcode.General, "force break websocket loop")
}
