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
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/app"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/types"
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
	cfg.IsExternal = op.IsExternal

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

	cfg.KubeConfig = op.Kubeconfig
	cfg.StoreDriver = op.StoreDriver

	cfg.NetServiceZK = op.NetServiceZK
	if len(cfg.NetServiceZK) == 0 {
		cfg.NetServiceZK = cfg.RegDiscvSvr
	}
	// etcd registry feature
	cfg.Etcd = op.Etcd
	storageAddr := op.StorageAddress
	storageAddr = strings.Replace(storageAddr, ";", ",", -1)
	cfg.StorageAddresses = strings.Split(storageAddr, ",")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	op := &MesosWatchOptions{
		Etcd: registry.CMDOptions{
			Feature: false,
		},
	}
	conf.Parse(op)

	blog.InitLogs(op.LogConfig)
	defer blog.CloseLogs()

	if err := common.SavePid(op.ProcessConfig); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
	}

	cfg = app.DefaultConfig()
	setCfg(op)

	if err := app.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Startting bcs-mesos-watch Err: %s", err.Error())
		os.Exit(1)
	}
}

//MesosWatchOptions options for mesos watch
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
	IsExternal             bool   `json:"is_external" value:"false" usage:"the cluster whether external deployment"`
	Kubeconfig             string `json:"kubeconfig" value:"" usage:"kubeconfig, when store_driver is etcd"`
	StoreDriver            string `json:"store_driver" value:"zookeeper" usage:"the store driver, enum: zookeeper, etcd"`

	// NetServiceZK is zookeeper address config for netservice discovery,
	// reuse RegDiscvSvr by default.
	NetServiceZK string `json:"netservice_zookeeper" value:"" usage:"netservice discovery zookeeper address"`

	// Etcd etcd options for service discovery
	Etcd registry.CMDOptions `json:"etcdRegistry"`

	// StorageAddress storage address
	StorageAddress string `json:"storage_address" value:"" usage:"storage address"`
}
