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
	"testing"

	"github.com/stretchr/testify/assert"

	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestBasicHandler(t *testing.T) {
	crh := NewClusterResourcesHandler()

	// Echo API
	echoReq, echoResp := clusterRes.EchoReq{Str: "testString"}, clusterRes.EchoResp{}
	err := crh.Echo(context.TODO(), &echoReq, &echoResp)
	assert.Equal(t, "Echo: testString", echoResp.Ret)
	assert.Nil(t, err)

	// Ping API
	pingReq, pingResp := clusterRes.PingReq{}, clusterRes.PingResp{}
	err = crh.Ping(context.TODO(), &pingReq, &pingResp)
	assert.Equal(t, "pong", pingResp.Ret)
	assert.Nil(t, err)

	// Healthz API
	healthzReq, healthzResp := clusterRes.HealthzReq{}, clusterRes.HealthzResp{}
	err = crh.Healthz(context.TODO(), &healthzReq, &healthzResp)
	assert.Equal(t, "OK", healthzResp.Status)
	assert.Nil(t, err)

	// Version API
	versionReq, versionResp := clusterRes.VersionReq{}, clusterRes.VersionResp{}
	err = crh.Version(context.TODO(), &versionReq, &versionResp)
	assert.Equal(t, "go1.14.15", versionResp.GoVersion)
	assert.Nil(t, err)
}
