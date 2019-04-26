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
	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/conf"
	"bk-bcs/bcs-common/common/license"
	"bk-bcs/bcs-mesos/bcs-mesos-watch/app"
	"bk-bcs/bcs-mesos/bcs-mesos-watch/types"
	"fmt"
	"os"
	"runtime"
)

var (
	cfg *types.CmdConfig
)

func setCfg(op *MesosWatchOptions) {
	cfg.ClusterInfo = op.ClusterInfo
	cfg.ApplicationThreadNum = int(op.ApplicationThreadNum)
	cfg.TaskgroupThreadNum = int(op.TaskgroupThreadNum)
	cfg.ExportserviceThreadNum = int(op.ExportserviceThreadNum)
	cfg.CAFile = op.CAFile
	cfg.CertFile = op.ClientCertFile
	cfg.KeyFile = op.ClientKeyFile
	cfg.RegDiscvSvr = op.BCSZk
	cfg.Address = op.Address

	cfg.MetricPort = op.MetricPort

	cfg.ServerCAFile = op.CAFile
	cfg.ServerCertFile = op.ServerCertFile
	cfg.ServerKeyFile = op.ServerKeyFile

	cfg.ClusterID = op.Cluster

	if cfg.ServerCertFile != "" && cfg.ServerKeyFile != "" {
		cfg.ServerSchem = "https"
	} else {
		cfg.ServerSchem = "http"
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	op := &MesosWatchOptions{}
	conf.Parse(op)

	blog.InitLogs(op.LogConfig)
	defer blog.CloseLogs()

	if err := common.SavePid(op.ProcessConfig); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
	}

	cfg = app.DefaultConfig()
	setCfg(op)

	license.CheckLicense(op.LicenseServerConfig)

	if err := app.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Startting bcs-mesos-watch Err: %s", err.Error())
		os.Exit(1)
	}
}

type MesosWatchOptions struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.CertConfig
	conf.LicenseServerConfig

	conf.LogConfig
	conf.ProcessConfig

	ClusterInfo            string `json:"clusterinfo" value:"127.0.0.1:2181/blueking" usage:"cluster data storage information"`
	ApplicationThreadNum   uint   `json:"app_threads" value:"20" usage:"application thread num"`
	TaskgroupThreadNum     uint   `json:"taskgroup_threads" value:"100" usage:"taskgroup thread num"`
	ExportserviceThreadNum uint   `json:"exportservice_threads" value:"100" usage:"exportservice thread num"`
	Cluster                string `json:"cluster" value:"" usage:"the cluster ID under bcs"`
}
