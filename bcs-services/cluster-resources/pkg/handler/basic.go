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

/*
 * basic.go 模块基础类接口，含 Ping，Healthz 等
 */

package handler

import (
	"context"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

type clusterResourcesHandler struct{}

// NewClusterResourcesHandler 创建服务处理逻辑集
func NewClusterResourcesHandler() *clusterResourcesHandler {
	return &clusterResourcesHandler{}
}

// Echo 回显测试
func (crh *clusterResourcesHandler) Echo(
	ctx context.Context,
	req *clusterRes.EchoReq,
	resp *clusterRes.EchoResp,
) error {
	if err := req.Validate(); err != nil {
		blog.Errorf("echo string validate failed: %s", err.Error())
		return err
	}
	resp.Ret = "Echo: " + req.Str
	return nil
}

// Ping 服务可达检测
func (crh *clusterResourcesHandler) Ping(
	ctx context.Context,
	req *clusterRes.PingReq,
	resp *clusterRes.PingResp,
) error {
	resp.Ret = "pong"
	return nil
}

// Healthz 服务健康信息
func (crh *clusterResourcesHandler) Healthz(
	ctx context.Context,
	req *clusterRes.HealthzReq,
	resp *clusterRes.HealthzResp,
) error {
	resp.Status = "OK"
	return nil
}
