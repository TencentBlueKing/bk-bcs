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

package option

import (
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/conf"
	"path/filepath"

	"github.com/kardianos/osext"
)

const (
	ProxyHaproxy               = "haproxy"
	ProxyHaproxyDefaultBinPath = "/usr/sbin/haproxy"
	ProxyHaproxyDefaultCfgPath = "/etc/haproxy/haproxy.cfg"
	ProxyNginx                 = "nginx"
	ProxyNginxDefaultBinPath   = "/usr/local/nginx/sbin/nginx"
	ProxyNginxDefaultCfgPath   = "/usr/local/nginx/conf/nginx.conf"
)

//LBConfig hold load balance all config
type LBConfig struct {
	conf.CertConfig
	Zookeeper      string //zk links
	WatchPath      string //zk watch path
	Group          string //group to serve
	Proxy          string //proxy implenmentation, nginx or haproxy
	BcsZkAddr      string //bcs zookeeper address
	ClusterID      string //cluster id to register path
	WorkDir        string //bcs_loadbalance workspace, this is one parameter
	LogDir         string //bcs_loadbalance log directory
	CfgBackupDir   string //haproxy cfg backup directory
	GeneratingDir  string //haproxy cfg generation directory
	TemplateDir    string //template file used to generate haproxy.cfg
	BinPath        string //haproxy bin path
	CfgPath        string //haproxy configuration file path
	SyncPeriod     int    //time period for syncing data
	CfgCheckPeriod int    //period check cache update
	MetricPort     uint   //metric check port
	CAFile         string //tls ca file
	ClientCertFile string //tls cert file
	ClientKeyFile  string //tls key file
}

//NewDefaultConfig create default config item
func NewDefaultConfig() *LBConfig {
	config := new(LBConfig)
	basepath, err := osext.Executable()
	if err != nil {
		blog.Warnf("osext.Executable err %s", err.Error())
	}
	config.WorkDir = filepath.Dir(basepath)
	config.LogDir = filepath.Join(config.WorkDir, "logs")
	config.CfgBackupDir = filepath.Join(config.WorkDir, "backup")
	config.GeneratingDir = filepath.Join(config.WorkDir, "generate")
	config.TemplateDir = filepath.Join(config.WorkDir, "template")
	config.CAFile = ""
	config.ClientCertFile = ""
	config.ClientKeyFile = ""
	config.Proxy = ProxyHaproxy
	config.Group = "external"
	config.BinPath = ProxyHaproxyDefaultBinPath
	config.CfgPath = ProxyHaproxyDefaultCfgPath
	config.SyncPeriod = 30
	config.CfgCheckPeriod = 4
	config.MetricPort = 59090
	return config
}
