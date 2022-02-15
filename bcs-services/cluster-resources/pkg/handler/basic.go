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

// Package handler basic.go 模块基础类接口实现，含 Ping，Healthz 等
package handler

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cache/redis"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ClusterResourcesHandler ...
type ClusterResourcesHandler struct{}

// NewClusterResourcesHandler 创建服务处理逻辑集
func NewClusterResourcesHandler() *ClusterResourcesHandler {
	return &ClusterResourcesHandler{}
}

// Echo 回显测试
func (crh *ClusterResourcesHandler) Echo(
	_ context.Context,
	req *clusterRes.EchoReq,
	resp *clusterRes.EchoResp,
) error {
	resp.Ret = "Echo: " + req.Str
	return nil
}

// Ping 服务可达检测
func (crh *ClusterResourcesHandler) Ping(
	_ context.Context,
	_ *clusterRes.PingReq,
	resp *clusterRes.PingResp,
) error {
	resp.Ret = "pong"
	return nil
}

// Version 服务版本信息
func (crh *ClusterResourcesHandler) Version(
	_ context.Context,
	_ *clusterRes.VersionReq,
	resp *clusterRes.VersionResp,
) error {
	resp.Version = version.Version
	resp.GitCommit = version.GitCommit
	resp.BuildTime = version.BuildTime
	resp.GoVersion = version.GoVersion
	resp.RunMode = runtime.RunMode
	resp.CallTime = util.GetCurTime()
	return nil
}

// Healthz 服务健康信息
func (crh *ClusterResourcesHandler) Healthz(
	_ context.Context,
	_ *clusterRes.HealthzReq,
	resp *clusterRes.HealthzResp,
) error {
	// 服务是否健康标志
	allOK := true

	// 检查 redis 状态
	ret, err := redis.GetDefaultClient().Ping(context.TODO()).Result() // nolint:contextcheck
	if ret != "PONG" || err != nil {
		resp.Redis = genHealthzStatus(false, "Ping Failed")
		allOK = false
	} else {
		resp.Redis = genHealthzStatus(true, "")
	}

	// 转换为可读状态
	resp.Status = genHealthzStatus(allOK, "")
	resp.CallTime = util.GetCurTime()
	return nil
}

// 生成可读状态信息
func genHealthzStatus(isOK bool, moreInfo string) string {
	if isOK {
		return "OK"
	}
	if moreInfo != "" {
		return "UnHealthy: " + moreInfo
	}
	return "UnHealthy"
}
