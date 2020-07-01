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
 *
 */

package options

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-consoleproxy/console-proxy/config"
)

//ConsoleProxyOption is option in flags
type ConsoleProxyOption struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.CertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig

	Privilege      bool     `json:"privilege" value:"" usage:"container exec privilege"`
	Cmd            []string `json:"cmd" value:"" usage:"cosntainer exec cmd"`
	Tty            bool     `json:"tty" value:"" usage:"tty"`
	DockerEndpoint string   `json:"docker-endpoint" value:"" usage:"docker endpoint"`
	Ips            []string `json:"ips" value:"" usage:"IP white list"`
	IsAuth         bool     `json:"is-auth" value:"" usage:"is auth"`
	IsOneSession   bool     `json:"is-one-session" value:"" usage:"support just one session for an container"`

	Conf config.ConsoleProxyConfig
}

//NewConsoleProxyOption create ConsoleProxyOption object
func NewConsoleProxyOption() *ConsoleProxyOption {
	return &ConsoleProxyOption{
		Conf: config.NewConsoleProxyConfig(),
	}
}
