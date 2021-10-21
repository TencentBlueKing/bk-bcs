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

package util

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"

	//"github.com/urfave/cli"
	"os"
	"strconv"
)

// SchedulerOptions for bcs-scheduler
type SchedulerOptions struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.CertConfig
	conf.LicenseServerConfig

	conf.LogConfig
	conf.ProcessConfig

	AlertServer       string `json:"alertServer" value:"" usage:"bcs-alert-manager server address"`
	ClientAuth        bool   `json:"clientAuth" value:"false" usage:"bcs-alert-manager server client auth"`
	Token             string `json:"token" value:"" usage:"api-gateway token for call alertmanager api"`
	AlertDebug        bool   `json:"alertDebug" value:"false" usage:"alertDebug debug bcs-alert-manager http api"`
	MesosMasterZK     string `json:"mesos_regdiscv" value:"" usage:"the address to discove mesos master"`
	RegDiscvSvr       string `json:"regdiscv" value:"" usage:"the address to register and discove scheduler"`
	UseCache          bool   `json:"use_cache" value:"false" usage:"whether use cache or not"`
	DoRecover         bool   `json:"do_recover" value:"false" usage:"whether recover taskgroup LOST to RUNNING in master role"`
	Plugins           string `json:"plugins" value:"" usage:"whether use plugins"`
	ZkHost            string `json:"zkhost" value:"" usage:"zk address"`
	Cluster           string `json:"cluster" value:"" usage:"the cluster ID under bcs"`
	PluginDir         string `json:"plugin_dir" value:"" usage:"the plugin dir"`
	ContainerExecutor string `json:"container_executor" value:"" usage:"the container executor path"`
	ProcessExecutor   string `json:"process_executor" value:"" usage:"the process executor path"`
	CniDir            string `json:"cni_dir" value:"" usage:"the cni directory"`
	NetImage          string `json:"net_image" value:"" usage:"the network image"`
	Kubeconfig        string `json:"kubeconfig" value:"" usage:"kubeconfig, when store_driver is etcd"`
	StoreDriver       string `json:"store_driver" value:"zookeeper" usage:"the store driver, enum: zookeeper, etcd"`
	DebugMode         bool   `json:"debug_mode" value:"false" usage:"Debug mode, use pprof."`
}

// AlertManagerConfig for bcs-alert-manager config
type AlertManagerConfig struct {
	Server     string
	Token      string
	ClientAuth bool
	Debug      bool
}

// SchedConfig for parse conf
type SchedConfig struct {
	Scheduler    Scheduler
	HttpListener HttpListener
	AlertManager AlertManagerConfig
	ZkHost       string
}

// Scheduler for scheduler config
type Scheduler struct {
	Hostname      string
	MesosMasterZK string
	BcsZK         string
	RegDiscvSvr   string
	Address       string
	Scheme        string
	ZK            string
	UseCache      bool
	DoRecover     bool
	Plugins       string
	PluginDir     string

	ClientCAFile   string
	ClientCertFile string
	ClientKeyFile  string

	ServerCAFile   string
	ServerCertFile string
	ServerKeyFile  string
	MetricPort     uint

	Cluster           string
	ContainerExecutor string
	ProcessExecutor   string
	CniDir            string
	NetImage          string

	Kubeconfig  string
	StoreDriver string

	//whether use pprof
	DebugMode bool
}

// HttpListener for http server conf
type HttpListener struct {
	TCPAddr  string
	UnixAddr string
	IsSSL    bool
	//CertDir    string
	CAFile     string
	CertFile   string
	KeyFile    string
	CertPasswd string
}

// NewSchedulerCfg init scheduler config
func NewSchedulerCfg() *SchedConfig {
	config := SchedConfig{
		ZkHost: "",
		AlertManager: AlertManagerConfig{
			Server:     "",
			Token:      "",
			ClientAuth: false,
			Debug:      false,
		},
		HttpListener: HttpListener{
			TCPAddr:  "",
			UnixAddr: "",
			IsSSL:    false,
			//CertDir:    "",
			CAFile:     "",
			CertFile:   "",
			KeyFile:    "",
			CertPasswd: static.ServerCertPwd,
		},

		Scheduler: Scheduler{
			MesosMasterZK: "",
			BcsZK:         "",
			//ClientCertDir:			"",
			ClientCAFile:   "",
			ClientCertFile: "",
			ClientKeyFile:  "",
			RegDiscvSvr:    "",
			Hostname:       hostname(),
			Scheme:         "http",
			UseCache:       false,
			DoRecover:      false,
			Cluster:        "",
		},
	}

	return &config
}

func SetSchedulerCfg(config *SchedConfig, op *SchedulerOptions) {

	config.ZkHost = op.ZkHost

	config.Scheduler.MesosMasterZK = op.MesosMasterZK
	config.Scheduler.BcsZK = op.BCSZk
	//config.Scheduler.ClientCertDir = op.ClientCertDir
	config.Scheduler.ClientCAFile = op.CAFile
	config.Scheduler.ClientCertFile = op.ClientCertFile
	config.Scheduler.ClientKeyFile = op.ClientKeyFile

	config.Scheduler.ServerCAFile = op.CAFile
	config.Scheduler.ServerCertFile = op.ServerCertFile
	config.Scheduler.ServerKeyFile = op.ServerKeyFile
	config.Scheduler.MetricPort = op.MetricPort

	config.Scheduler.RegDiscvSvr = op.RegDiscvSvr
	config.Scheduler.Address = op.ServiceConfig.Address + ":" + strconv.Itoa(int(op.Port))
	config.Scheduler.UseCache = op.UseCache
	config.Scheduler.DoRecover = op.DoRecover
	config.Scheduler.Plugins = op.Plugins
	config.Scheduler.Cluster = op.Cluster
	config.Scheduler.PluginDir = op.PluginDir
	config.Scheduler.ContainerExecutor = op.ContainerExecutor
	config.Scheduler.ProcessExecutor = op.ProcessExecutor
	config.Scheduler.CniDir = op.CniDir
	config.Scheduler.NetImage = op.NetImage

	config.HttpListener.TCPAddr = op.ServiceConfig.Address + ":" + strconv.Itoa(int(op.Port))
	//config.HttpListener.CertDir = op.ServerCertDir
	config.HttpListener.CAFile = op.CAFile
	config.HttpListener.CertFile = op.ServerCertFile
	config.HttpListener.KeyFile = op.ServerKeyFile
	if config.HttpListener.CertFile != "" && config.HttpListener.KeyFile != "" {
		config.HttpListener.IsSSL = true
		config.Scheduler.Scheme = "https"
	}

	config.Scheduler.Kubeconfig = op.Kubeconfig
	config.Scheduler.StoreDriver = op.StoreDriver
	config.Scheduler.DebugMode = op.DebugMode

	config.AlertManager.Server = op.AlertServer
	config.AlertManager.ClientAuth = op.ClientAuth
	config.AlertManager.Debug = op.AlertDebug
	if op.ClientAuth {
		config.AlertManager.Token = op.Token
	}
}

func hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}

	return hostname
}
