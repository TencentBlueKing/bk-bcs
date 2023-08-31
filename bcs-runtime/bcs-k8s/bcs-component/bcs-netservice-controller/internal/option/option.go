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

// ControllerOption controller option
type ControllerOption struct {
	// Address address for server
	Address string

	// Port port for server
	Port int

	// MetricPort port for metric server
	MetricPort int

	// ProbePort port for probe
	ProbePort int

	// EnableLeaderElect enable leader elect
	EnableLeaderElect bool

	conf.LogConfig

	// HttpServerPort port for http api
	HttpServerPort uint

	Conf Conf

	ServCert ServCert
}

// Conf 服务配置
type Conf struct {
	ServCert        ServCert
	InsecureAddress string
	InsecurePort    uint
	VerifyClientTLS bool
}

// ServCert 服务证书配置
type ServCert struct {
	CAFile     string
	CertFile   string
	KeyFile    string
	CertPasswd string
	IsSSL      bool
}
