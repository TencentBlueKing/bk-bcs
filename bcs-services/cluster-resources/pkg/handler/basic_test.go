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

	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestBasicHandler(t *testing.T) {
	crh := NewClusterResourcesHandler()

	// EchoAPI
	echoReq, echoResp := clusterRes.EchoReq{Str: "testString"}, clusterRes.EchoResp{}
	if err := crh.Echo(context.TODO(), &echoReq, &echoResp); echoResp.Ret != "Echo: testString" || err != nil {
		t.Errorf("Test CR.Echo failed, resp.Ret excepted: 'Echo: testString', result: %s", echoResp.Ret)
	}

	// PingAPI
	pingReq, pingResp := clusterRes.PingReq{}, clusterRes.PingResp{}
	if err := crh.Ping(context.TODO(), &pingReq, &pingResp); pingResp.Ret != "pong" || err != nil {
		t.Errorf("Test CR.Ping failed, resp.Ret excepted: 'ping', result: %s", pingResp.Ret)
	}

	// HealthzAPI
	healthzReq, healthzResp := clusterRes.HealthzReq{}, clusterRes.HealthzResp{}
	if err := crh.Healthz(context.TODO(), &healthzReq, &healthzResp); healthzResp.Status != "OK" || err != nil {
		t.Errorf("Test CR.Ping failed, resp.Status excepted: 'OK', result: %s", healthzResp.Status)
	}
}
