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

package app

import (
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	comconf "github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/manager"
)

// Console is an console struct
type Console struct {
	backend manager.Manager
	route   *api.Router
	conf    *config.ConsoleConfig
}

// NewConsole create an ConsoleProxy object
func NewConsole(op *options.ConsoleOption) *Console {
	setConfig(op)

	c := &Console{
		conf:    &op.Conf,
		backend: manager.NewManager(&op.Conf),
	}

	err := c.backend.Start()
	if err != nil {
		blog.Errorf("start manager error %s", err.Error())
		os.Exit(1)
	}

	//http server
	c.route = api.NewRouter(c.backend, c.conf)
	return c
}

// Run create a pid
func (c *Console) Run() {
	//pid
	if err := common.SavePid(comconf.ProcessConfig{}); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
	}
}

func setConfig(op *options.ConsoleOption) {
	op.Conf.Address = op.Address
	op.Conf.Port = int(op.Port)
	op.Conf.Tty = op.Tty
	op.Conf.Privilege = op.Privilege
	op.Conf.Cmd = op.Cmd
	op.Conf.Ips = op.Ips
	op.Conf.IsAuth = op.IsAuth
	op.Conf.WebConsoleImage = "ccr.ccs.tencentyun.com/bk-cmdb-lf/bcs-webconsole:v0.1" // TODO
	op.Conf.IndexPageTemplatesFile = op.IndexPageTemplatesFile
	op.Conf.MgrPageTemplatesFile = op.MgrPageTemplatesFile

	//server cert directoty
	if op.CertConfig.ServerCertFile != "" && op.CertConfig.CAFile != "" &&
		op.CertConfig.ServerKeyFile != "" {

		op.Conf.ServCert.CertFile = op.CertConfig.ServerCertFile
		op.Conf.ServCert.KeyFile = op.CertConfig.ServerKeyFile
		op.Conf.ServCert.CAFile = op.CertConfig.CAFile
		op.Conf.ServCert.IsSSL = true
	}
}
