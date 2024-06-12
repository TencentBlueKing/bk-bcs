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

// Package option xxx
package option

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// ControllerOption defines the option of workflow controller
type ControllerOption struct {
	conf.FileConfig
	conf.LogConfig

	Address    string `json:"address" value:"0.0.0.0" usage:"local address"`
	Port       int    `json:"port" value:"8080" usage:"http port"`
	MetricPort int    `json:"metricPort" value:"8081" usage:"metric port"`
	GRPCPort   int    `json:"grpcPort" value:"8082" usage:"grpc port"`
	HealthPort int    `json:"healthPort" value:"8083" usage:"health port"`
	HTTPPort   int    `json:"httpPort" value:"8088" usage:"http port"`

	EnableLeaderElection bool `json:"enableLeaderElection" value:"true" usage:"whether enable leader election"`
	KubernetesQPS        int  `json:"kubernetesQPS" value:"50" usage:"kubernetes qps"`
	KubernetesBurst      int  `json:"kubernetesBurst" value:"100" usage:"kubernetes burst"`
	MaxWorkers           int  `json:"maxWorkers" value:"10" usage:"workers of controller"`

	BKDevOpsUrl       string `json:"bkDevOpsUrl,omitempty"`
	BKDevOpsAppCode   string `json:"bkDevOpsAppCode,omitempty"`
	BKDevOpsAppSecret string `json:"bkDevOpsAppSecret,omitempty"`
}

var (
	globalOption *ControllerOption
)

// Parse the ControllerOperation from config file
func Parse() {
	globalOption = &ControllerOption{}
	conf.Parse(globalOption)
}

// GlobalOption return the global option
func GlobalOption() *ControllerOption {
	return globalOption
}
