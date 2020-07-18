/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package option

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// NewConfig create new config object
func NewConfig() *Config {
	cfg := new(Config)
	return cfg
}

// Config config for bcs cloud netservice
type Config struct {
	// Kubeconfig kubeconfig for kubernetes store
	Kubeconfig string `json:"kubeconfig" value:"" usage:"kubeconfig for kubernetes apiserver"`

	// InCluster is in kubernetes
	InCluster bool `json:"inCluster" value:"false" usage:"if is cloud netservice is in kubernetes cluster"`

	// Debug debug flag
	Debug bool `json:"debug" value:"false" usage:"debug flag, open pprof"`

	// SwaggerDir
	SwaggerDir string `json:"swaggerDir" value:"" usage:"swagger dir"`

	// CloudMode cloud mod
	CloudMode string `json:"cloudMode" value:"" usage:"cloud mode, option [tencentcloud, aws]"`

	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.LogConfig
}
