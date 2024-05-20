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

// Package metric xxx
package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// RpcCallFailed rpc call failed
	RpcCallFailed *prometheus.CounterVec
)

// Init init metric
func Init() {
	RpcCallFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "remotecommand_rpc_call_failed",
		Help: "call rpc failed counter",
	}, []string{"service", "endpoint"})
	prometheus.MustRegister(RpcCallFailed)
}
