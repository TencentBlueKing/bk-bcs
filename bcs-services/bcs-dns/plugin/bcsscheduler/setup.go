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

package bcsscheduler

import (
	"errors"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"

	bsMetrics "github.com/Tencent/bk-bcs/bcs-services/bcs-dns/plugin/bcsscheduler/metrics"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/coredns/coredns/plugin/pkg/parse"
	"github.com/coredns/coredns/plugin/proxy"
	"github.com/mholt/caddy"
)

//setup is for parse config item for bcs-scheduler and register
//plugin bcs-scheduler with caddy

//init register bcs-scheduler plugin
func init() {
	caddy.RegisterPlugin("bcsscheduler", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

//setup setup
func setup(c *caddy.Controller) error {
	//pid
	if err := common.SavePid(conf.ProcessConfig{PidDir: "./pid"}); err != nil {
		return plugin.Error("bcsscheduler", err)
	}

	config, err := schedulerParse(c)
	if err != nil {
		return plugin.Error("bcsscheduler", err)
	}

	scheduler := NewScheduler(config)
	if err := scheduler.InitSchedulerCache(); err != nil {
		return plugin.Error("bcsscheduler", err)
	}
	//start & stop register
	c.OnStartup(func() error {
		metrics.MustRegister(c, bsMetrics.RequestCount, bsMetrics.RequestLatency, bsMetrics.RequestOutProxyCount, bsMetrics.DnsTotal, bsMetrics.StorageOperatorTotal, bsMetrics.StorageOperatorLatency, bsMetrics.ZkNotifyTotal)
		return scheduler.Start()
	})
	c.OnShutdown(func() error {
		return scheduler.Stop()
	})
	//pluging register
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		scheduler.Next = next
		return scheduler
	})
	return nil
}

//ConfigItem item from bcs-scheduler config file
type ConfigItem struct {
	Zones        []string    //zone list
	Cluster      string      //cluster id for mesos
	Register     []string    //registery for server node
	ResyncPeriod int         //resync all data from scheduler zookeeper
	Endpoints    []string    //scheduler event storage endpoints
	EndpointPath string      //path for original datas
	KubeConfig   string      //path of kubeconfig for kube-apiserver
	EndpointCA   string      //ca for endpoints
	EndpointKey  string      //key for endpoint
	EndpointCert string      //cert for endpoint
	Storage      []string    //link for storage
	StoragePath  string      //path for DNS data storage
	StorageCA    string      //ca for storage
	StorageKey   string      //key for storage
	StorageCert  string      //cert for storage
	UpStream     []string    //dns upstream
	Fallthrough  bool        //pass to next plugin when no data
	Proxy        proxy.Proxy //proxy for upstream
	MetricPort   uint        //port for prometheus metric
}

func defaultConfigItem() *ConfigItem {
	config := new(ConfigItem)
	config.EndpointPath = "/blueking"
	config.ResyncPeriod = 60
	config.Fallthrough = false
	return config
}

// bcs-scheduler parameter parse
// #lizard forgives
func schedulerParse(c *caddy.Controller) (*ConfigItem, error) {
	config := defaultConfigItem()
	//parse config from configuration
	for c.Next() {
		if c.Val() == "bcsscheduler" {
			//done(developer): zone feature support
			config.Zones = c.RemainingArgs()

			if len(config.Zones) == 0 {
				config.Zones = make([]string, len(c.ServerBlockKeys))
				copy(config.Zones, c.ServerBlockKeys)
			}
			plugin.Zones(config.Zones).Normalize()
			if config.Zones == nil || len(config.Zones) < 1 {
				return nil, errors.New("zone name must be provided for bcs-scheduler")
			}
			//parameter parse
			for c.NextBlock() {
				switch c.Val() {
				case "cluster":
					args := c.RemainingArgs()
					if len(args) == 1 {
						config.Cluster = args[0]
						continue
					}
					return nil, c.ArgErr()
				case "registery":
					args := c.RemainingArgs()
					if len(args) > 0 {
						config.Register = args
						continue
					}
					return nil, c.ArgErr()
				case "resyncperiod":
					args := c.RemainingArgs()
					if len(args) == 1 {
						period, err := strconv.Atoi(args[0])
						if err != nil {
							return nil, c.ArgErr()
						}
						config.ResyncPeriod = period
						continue
					}
					return nil, c.ArgErr()
				case "endpoints":
					args := c.RemainingArgs()
					if len(args) > 0 {
						config.Endpoints = append(config.Endpoints, args...)
						continue
					}
					return nil, c.ArgErr()
				case "endpoints-tls":
					// cert key cacertfile
					args := c.RemainingArgs()
					if len(args) == 3 {
						config.EndpointCert, config.EndpointKey, config.EndpointCA = args[0], args[1], args[2]
						continue
					}
					return nil, c.ArgErr()
				case "endpoints-path":
					args := c.RemainingArgs()
					if len(args) == 1 {
						config.EndpointPath = args[0]
						continue
					}
					return nil, c.ArgErr()
				case "storage":
					args := c.RemainingArgs()
					if len(args) > 0 {
						config.Storage = append(config.Storage, args...)
						continue
					}
					return nil, c.ArgErr()
				case "storage-path":
					args := c.RemainingArgs()
					if len(args) == 1 {
						config.StoragePath = args[0]
						continue
					}
					return nil, c.ArgErr()
				case "storage-tls": // cert key cacertfile
					args := c.RemainingArgs()
					if len(args) == 3 {
						config.StorageCert, config.StorageKey, config.StorageCA = args[0], args[1], args[2]
						continue
					}
					return nil, c.ArgErr()
				case "upstream":
					args := c.RemainingArgs()
					if len(args) == 0 {
						return nil, c.ArgErr()
					}
					ups, err := parse.HostPortOrFile(args...)
					if err != nil {
						return nil, err
					}
					config.Proxy = proxy.NewLookup(ups)
				case "fallthrough":
					config.Fallthrough = true
				case "metric-port":
					args := c.RemainingArgs()
					if len(args) == 1 {
						metricPortStr := args[0]
						metricPort, err := strconv.Atoi(metricPortStr)
						if err != nil {
							return nil, c.ArgErr()
						}
						config.MetricPort = uint(metricPort)
						continue
					}
					return nil, c.ArgErr()
				case "kubeconfig":
					args := c.RemainingArgs()
					if len(args) == 1 {
						config.KubeConfig = args[0]
						continue
					}
					return nil, c.ArgErr()
				}
			}
			return config, nil
		}
	}
	return nil, errors.New("bcsscheduler plugin called without keyword 'bcs-scheduler' in Corefile")
}
