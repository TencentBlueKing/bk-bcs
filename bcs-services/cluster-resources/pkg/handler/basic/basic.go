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

// Package basic 基础类接口实现，含 Ping，Healthz 等
package basic

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cache/redis"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/timex"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// Handler ...
type Handler struct{}

// New 创建服务处理逻辑集
func New() *Handler {
	return &Handler{}
}

// Echo 回显测试
func (h *Handler) Echo(
	ctx context.Context,
	req *clusterRes.EchoReq,
	resp *clusterRes.EchoResp,
) error {
	resp.Ret = fmt.Sprintf("Caller: %s, Echo: %s", ctx.Value(ctxkey.UsernameKey), req.Str)
	return nil
}

// Ping 服务可达检测
func (h *Handler) Ping(
	_ context.Context,
	_ *clusterRes.PingReq,
	resp *clusterRes.PingResp,
) error {
	resp.Ret = "pong"
	return nil
}

// Version 服务版本信息
func (h *Handler) Version(
	_ context.Context,
	_ *clusterRes.VersionReq,
	resp *clusterRes.VersionResp,
) error {
	resp.Version = version.Version
	resp.GitCommit = version.GitCommit
	resp.BuildTime = version.BuildTime
	resp.GoVersion = version.GoVersion
	resp.RunMode = runtime.RunMode
	resp.CallTime = timex.Current()
	return nil
}

// Healthz 服务健康信息
func (h *Handler) Healthz(
	_ context.Context,
	req *clusterRes.HealthzReq,
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
	resp.CallTime = timex.Current()

	// 一般用于健康探针等，可直接根据 http statusCode 判断是否成功
	// 若需要查看总的服务状态，则无需指定 raiseErr 的值
	if req.RaiseErr {
		return err
	}
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
