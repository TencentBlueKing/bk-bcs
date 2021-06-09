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
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-loadbalance/types"
	"os"
	"strings"
)

const (
	// ProxyHaproxy proxy haproxy
	ProxyHaproxy = "haproxy"
	// ProxyHaproxyDefaultBinPath haproxy proxy default bin path
	ProxyHaproxyDefaultBinPath = "/usr/sbin/haproxy"
	// ProxyHaproxyDefaultCfgPath haproxy proxy default config path
	ProxyHaproxyDefaultCfgPath = "/etc/haproxy/haproxy.cfg"
	// ProxyNginx proxy nginx
	ProxyNginx = "nginx"
	// ProxyNginxDefaultBinPath proxy nginx default bin path
	ProxyNginxDefaultBinPath = "/usr/local/nginx/sbin/nginx"
	// ProxyNginxDefaultCfgPath proxy nginx default config path
	ProxyNginxDefaultCfgPath = "/usr/local/nginx/conf/nginx.conf"
)

//LBConfig hold load balance all config
type LBConfig struct {
	conf.ClientOnlyCertConfig
	conf.LogConfig
	conf.FileConfig
	conf.ServiceConfig
	MetricPort int    `json:"metric_port" value:"59090" usage:"port for query metric info, version info and status info" mapstructure:"metric_port"`
	Zookeeper  string `json:"zk" value:"127.0.0.1:2381" usage:"zookeeper links for data source" mapstructure:"zk"`   //zk links
	WatchPath  string `json:"zkpath" value:"" usage:"service info path for watch, [required]" mapstructure:"zkpath"` //zk watch path
	// Name will be used in metric
	Name              string `json:"name" value:"" usage:"loadbalance instance name" mapstructure:"name"`
	Group             string `json:"group" value:"external" usage:"bcs loadbalance label for service join in" mapstructure:"group"` //group to serve
	Proxy             string `json:"proxy" value:"haproxy" usage:"proxy model, nginx or haproxy" mapstructure:"proxy"`              //proxy implenmentation, nginx or haproxy
	BcsZkAddr         string `json:"bcszkaddr" value:"127.0.0.1:2181" usage:"bcs zookeeper address" mapstructure:"bcszkaddr"`       //bcs zookeeper address
	ClusterZk         string `json:"clusterzk" value:"" usage:"cluster zookeeper address" mapstructure:"clusterzk"`
	ClusterID         string `json:"clusterid" value:"" usage:"loadbalance server mesos cluster id" mapstructure:"clusterid"`                     //cluster id to register path
	CfgBackupDir      string `json:"cfg_backup_dir" value:"" usage:"backup dir for loadbalance config file" mapstructure:"cfg_backup_dir"`        //haproxy cfg backup directory
	GeneratingDir     string `json:"generate_dir" value:"" usage:"dir for generated loadbalance config file" mapstructure:"generate_dir"`         //haproxy cfg generation directory
	TemplateDir       string `json:"template_dir" value:"" usage:"dir for template of loadbalance config file" mapstructure:"template_dir"`       //template file used to generate haproxy.cfg
	BinPath           string `json:"bin_path" value:"" usage:"bin path for proxy binary" mapstructure:"bin_path"`                                 //haproxy bin path
	CfgPath           string `json:"config_path" value:"" usage:"config path for proxy config" mapstructure:"config_path"`                        //haproxy configuration file path
	SyncPeriod        int    `json:"sync_period" value:"20" usage:"time period for syncing data" mapstructure:"sync_period"`                      //time period for syncing data
	CfgCheckPeriod    int    `json:"config_check_period" value:"5" usage:"time period for check cache update" mapstructure:"config_check_period"` //period check cache update
	StatusFetchPeriod int    `json:"stats_fetch_period" value:"15" usage:"time period for fetch proxy stats" mapstructure:"stats_fetch_period"`
}

// NewConfig new a config
func NewConfig() *LBConfig {
	return &LBConfig{}
}

// Parse parse config
func (c *LBConfig) Parse() error {
	conf.Parse(c)
	if len(c.WatchPath) == 0 {
		return fmt.Errorf("watch path cannot be empty")
	}
	if len(c.CfgBackupDir) == 0 {
		c.CfgBackupDir = "./backup"
	}
	if len(c.GeneratingDir) == 0 {
		c.GeneratingDir = "./generate"
	}
	if len(c.TemplateDir) == 0 {
		c.TemplateDir = "./template"
	}
	if c.Proxy == ProxyHaproxy {
		if len(c.BinPath) == 0 {
			c.BinPath = ProxyHaproxyDefaultBinPath
		}
		if len(c.CfgPath) == 0 {
			c.CfgPath = ProxyHaproxyDefaultCfgPath
		}
	} else if c.Proxy == ProxyNginx {
		if len(c.BinPath) == 0 {
			c.BinPath = ProxyNginxDefaultBinPath
		}
		if len(c.CfgPath) == 0 {
			c.CfgPath = ProxyNginxDefaultCfgPath
		}
	} else {
		return fmt.Errorf("invalid proxy %s", c.Proxy)
	}
	if c.SyncPeriod < 5 || c.SyncPeriod > 300 {
		return fmt.Errorf("sync_period must be [5, 300]")
	}
	if c.CfgCheckPeriod < 0 {
		return fmt.Errorf("config_check_period must be > 0")
	}
	c.Zookeeper = strings.Replace(c.Zookeeper, ";", ",", -1)
	c.BcsZkAddr = strings.Replace(c.BcsZkAddr, ";", ",", -1)
	c.ClusterZk = strings.Replace(c.ClusterZk, ";", ",", -1)

	// if name length is zero, read name from env "BCS_POD_ID"
	if len(c.Name) == 0 {
		c.Name = os.Getenv("BCS_POD_ID")
		if len(c.Name) == 0 {
			return fmt.Errorf("either option \"name\" or env BCS_POD_ID is needed")
		}
		pos := strings.LastIndex(c.Name, ".")
		if pos < 1 {
			return fmt.Errorf("invalid env BCS_POD_ID %s", c.Name)
		}
		c.Name = c.Name[:pos]
	}
	err := os.Setenv(types.EnvBcsLoadbalanceName, c.Name)
	if err != nil {
		return err
	}
	return nil
}
