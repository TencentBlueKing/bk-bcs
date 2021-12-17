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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
)

//ConsoleOption is option in flags
type ConsoleOption struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.CertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig

	Privilege              bool     `json:"privilege" value:"" usage:"container exec privilege"`
	Cmd                    []string `json:"cmd" value:"" usage:"cosntainer exec cmd"`
	Tty                    bool     `json:"tty" value:"true" usage:"tty"`
	WebConsoleImage        string   `json:"web-console-image" value:"" usage:"web-console images url"`
	Ips                    []string `json:"ips" value:"" usage:"IP white list"`
	IsAuth                 bool     `json:"is-auth" value:"" usage:"is auth"`
	IsOneSession           bool     `json:"is-one-session" value:"" usage:"support just one session for an container"`
	IndexPageTemplatesFile string   `json:"index-page-templates-file" value:"web/templates/index.html" usage:"index page templates file path"`
	MgrPageTemplatesFile   string   `json:"mgr-page-templates-file" value:"web/templates/mgr.html" usage:"mgr page templates file path"`

	Conf config.ConsoleConfig
}

//NewConsoleOption create ConsoleOption object
func NewConsoleOption() *ConsoleOption {
	return &ConsoleOption{
		Conf: config.NewConsoleConfig(),
	}
}
