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

package brpc

import (
	"errors"

	"google.golang.org/grpc"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
)

// PoolOption defines the gRPC client pool related options.
type PoolOption struct {
	PoolSize          int
	ReadBufferSizeMB  int
	WriteBufferSizeMB int
	ServiceName       cc.Name
	SvrDiscover       serviced.Discover
	TLS               cc.TLSConfig
	NewClient         func(conn *grpc.ClientConn) interface{}
}

// Validate the pool option is validate or not.
func (o *PoolOption) Validate() error {
	if o.PoolSize < 1 {
		return errors.New("invalid pool size, should >= 1")
	}

	if o.SvrDiscover == nil {
		return errors.New("service discover is nil")
	}

	if len(o.ServiceName) == 0 {
		return errors.New("service name not set")
	}

	if o.NewClient == nil {
		return errors.New("new client function is nil")
	}

	return nil
}
