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
	"crypto/tls"
	"net/http"

	microRgt "github.com/micro/go-micro/v2/registry"
	microSvc "github.com/micro/go-micro/v2/service"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/options"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

type ClusterResources struct {
	opts *options.ClusterResourcesOptions

	microSvc microSvc.Service
	microRtr microRgt.Registry

	httpServer *http.Server

	tlsConfig       *tls.Config
	clientTLSConfig *tls.Config

	stopCh chan struct{}
}

func NewClusterResources(opts *options.ClusterResourcesOptions) *ClusterResources {
	return &ClusterResources{opts: opts}
}

func (cr *ClusterResources) Echo(ctx context.Context, req *clusterRes.EchoReq, resp *clusterRes.EchoResp) error {
	if err := req.Validate(); err != nil {
		blog.Errorf("echo string validate failed: %s", err.Error())
		return err
	}
	resp.Ret = "Echo: " + req.Str
	return nil
}

func (cr *ClusterResources) Ping(ctx context.Context, req *clusterRes.PingReq, resp *clusterRes.PingResp) error {
	resp.Ret = "pong"
	return nil
}

func (cr *ClusterResources) Healthz(ctx context.Context, req *clusterRes.HealthzReq, resp *clusterRes.HealthzResp) error {
	resp.Status = "OK"
	return nil
}
