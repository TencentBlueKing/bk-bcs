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

package main

import (
	"net"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-daemon/process-daemon/api"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-daemon/process-daemon/config"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-daemon/process-daemon/manager"
)

//Option daemon process Option
type Option struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.CertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig

	DataDir      string `json:"data_dir" value:"" usage:"the process daemon data dir"`
	UnixSocket   string `json:"unix_socket" value:"" usage:"the unix socket path"`
	WorkspaceDir string `json:"workspace_dir" value:"" usage:"the process packages dir"`
}

// Init process init
func Init() {
	op := &Option{}
	conf.Parse(op)
	op.LogConfig.ToStdErr = true
	// op.DataDir = "./data"
	// op.UnixSocket = "/var/run/process.sock"
	// op.WorkspaceDir = "/data/bcs/workspace"

	blog.InitLogs(op.LogConfig)
	defer blog.CloseLogs()

	_, err := os.Stat(op.DataDir)
	if err != nil {
		blog.Errorf("datadir %s don't exist", err.Error())
		return
	}

	config := &config.Config{
		DataDir:      op.DataDir,
		WorkspaceDir: op.WorkspaceDir,
	}
	manager := manager.NewManager(config)
	err = manager.Init()
	if err != nil {
		blog.Errorf("manager init error %s", err.Error())
		return
	}
	manager.Start()
	route := api.NewRouter(manager)

	os.Remove(op.UnixSocket)

	unixaddr, err := net.ResolveUnixAddr("unix", op.UnixSocket)
	if err != nil {
		blog.Errorf("resolve unixaddr %s error %s", op.UnixSocket, err.Error())
		return
	}

	l, err := net.ListenUnix("unix", unixaddr)
	if err != nil {
		blog.Errorf("listen unix addr %s error %s", op.UnixSocket, err.Error())
		return
	}

	httpServ := httpserver.NewHttpServer(uint(0), "", "")
	httpServ.RegisterWebServer("/bcsapi/v1/processdaemon", nil, route.GetActions())

	httpServ.Serve(l)
}

func main() {
	Init()
}
