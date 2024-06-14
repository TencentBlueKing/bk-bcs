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

// Package apis xxx
package apis

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	traceconst "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/constants"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go-micro.dev/v4/client"
)

const (
	// StrategyManagerName the name of StrategyManager service
	StrategyManagerName = "strategymanager.bkbcs.tencent.com"
	// ClusterManagerServiceName the service name of cluster-manager
	ClusterManagerServiceName = "clustermanager.bkbcs.tencent.com"
	// ResourceManagerServiceName defines the service name of resource-manager
	ResourceManagerServiceName = "resourcemanager.bkbcs.tencent.com"
	// NodeGroupManagerServiceName defines the service name of nodegroup manager
	NodeGroupManagerServiceName = "nodegroupmanager.bkbcs.tencent.com"
	// ProjectManagerServiceName defines the service name of project-manager
	ProjectManagerServiceName = "project.bkbcs.tencent.com"
	// RetryTimes default 30
	RetryTimes = 30
	// RetryDuration default 10
	RetryDuration = 10
	// RpcDialTimeout rpc dial timeout
	RpcDialTimeout = 20
	// RpcRequestTimeout rpc request timeout
	RpcRequestTimeout = 20
	// RegisterTTL register ttl
	RegisterTTL = 20
	// RegisterInterval register interval
	RegisterInterval = 10
)

// MicroClientWrapper wrapper micro client with retry
func MicroClientWrapper(microClient client.Client, counter *prometheus.CounterVec) error {
	err := microClient.Init(
		client.Retries(RetryTimes),
		client.DialTimeout(RpcDialTimeout*time.Second),
		client.RequestTimeout(RpcRequestTimeout*time.Second),
		client.Retry(func(ctx context.Context, req client.Request, retryCount int, err error) (bool, error) {
			counter.WithLabelValues(req.Service(), req.Endpoint()).Inc()
			time.Sleep(RetryDuration * time.Second)
			blog.Warnf("Client retry %d, method: %s, body: %v", retryCount, req.Method(), req.Body())
			return true, nil
		}),
	)
	if err != nil {
		return errors.Wrapf(err, "wrapper micro client failed")
	}
	return nil
}

// BkAuthOpts bkAuth option
type BkAuthOpts struct {
	BkAppCode   string `json:"bk_app_code"`
	BkAppSecret string `json:"bk_app_secret"`
	AccessToken string `json:"access_token"`
}

// ClientOptions client option
type ClientOptions struct {
	Endpoint    string
	UserName    string
	AccessToken string
	AppCode     string
	AppSecret   string
}

// GetAuthUserFromCtx 从上下文获取用户信息
func GetAuthUserFromCtx(ctx context.Context) middleware.AuthUser {
	user := ctx.Value(middleware.AuthUserKey).(middleware.AuthUser)
	if user.Username == "" {
		user.Username = user.ClientName
	}
	return user
}

// GetRequestIDFromCtx 从上下文获取 RequestID
func GetRequestIDFromCtx(ctx context.Context) string {
	return ctx.Value(traceconst.RequestIDHeaderKey).(string)
}
