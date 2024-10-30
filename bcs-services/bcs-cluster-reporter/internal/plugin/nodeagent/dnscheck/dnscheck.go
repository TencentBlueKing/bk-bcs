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
 */

// Package dnscheck xxx
package dnscheck

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

// Plugin xxx
type Plugin struct {
	opt     *Options
	dnsLock sync.Mutex
	ready   bool
	Detail  Detail
	plugin_manager.NodePlugin
}

type DnsCheckResult struct {
	Type   string `yaml:"type"`
	Node   string `yaml:"node"`
	Status string `yaml:"status"`
}

type Detail struct {
}

var (
	dnsAvailabilityLabels = []string{"type", "node", "status"}
	dnsAvailability       = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dns_availability",
		Help: "dns_availability, 1 means OK",
	}, dnsAvailabilityLabels)

	dnsLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "dns_latency",
		Help:    "dns_latency",
		Buckets: []float64{0.001, 0.01, 0.1, 0.2, 0.4, 0.8, 1.6, 3.2},
	}, []string{})
)

func init() {
	metric_manager.Register(dnsAvailability)
	metric_manager.Register(dnsLatency)
}

// Setup xxx
func (p *Plugin) Setup(configFilePath string, runMode string) error {
	p.opt = &Options{}
	err := util.ReadorInitConf(configFilePath, p.opt, initContent)
	if err != nil {
		return err
	}

	if err = p.opt.Validate(); err != nil {
		return err
	}

	p.StopChan = make(chan int)
	interval := p.opt.Interval
	if interval == 0 {
		interval = 60
	}

	// run as daemon
	if runMode == plugin_manager.RunModeDaemon {
		go func() {
			for {
				if p.CheckLock.TryLock() {
					p.CheckLock.Unlock()
					go p.Check()
				} else {
					klog.Infof("the former %s didn't over, skip in this loop", p.Name())
				}
				select {
				case result := <-p.StopChan:
					klog.Infof("stop plugin %s by signal %d", p.Name(), result)
					return
				case <-time.After(time.Duration(interval) * time.Second):
					continue
				}
			}
		}()
	} else if runMode == plugin_manager.RunModeOnce {
		p.Check()
	}

	return nil
}

// Stop xxx
func (p *Plugin) Stop() error {
	p.CheckLock.Lock()
	p.StopChan <- 1
	klog.Infof("plugin %s stopped", p.Name())
	p.CheckLock.Unlock()
	return nil
}

// Name xxx
func (p *Plugin) Name() string {
	return pluginName
}

// Check xxx
func (p *Plugin) Check() {
	result := make([]plugin_manager.CheckItem, 0, 0)
	p.CheckLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		p.CheckLock.Unlock()
	}()

	node := plugin_manager.Pm.GetConfig().NodeConfig
	nodeName := node.NodeName
	p.ready = false

	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("dnscheck failed: %s, stack: %v\n", r, string(debug.Stack()))
		}
	}()

	dnsStatusGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)

	ctx := util.GetCtx(time.Second * 10)

	status, err := p.checkDNS(ctx, append(p.opt.CheckDomain, "kubernetes.default.svc.cluster.local"), "", node.ClientSet)
	dnsStatusGaugeVecSetList = append(dnsStatusGaugeVecSetList, &metric_manager.GaugeVecSet{
		Labels: []string{"pod", nodeName, status},
		Value:  float64(1),
	})
	if status != NormalStatus {
		result = append(result, plugin_manager.CheckItem{
			// 写入configmap默认使用英文
			ItemName:   pluginName,
			ItemTarget: nodeName,
			Level:      plugin_manager.RISKLevel,
			Normal:     false,
			Detail:     fmt.Sprintf("pod cluster dns resolv failed: %s", err.Error()),
			Status:     status,
		})
	} else {
		result = append(result, plugin_manager.CheckItem{
			// 写入configmap默认使用英文
			ItemName:   pluginName,
			ItemTarget: nodeName,
			Status:     status,
			Normal:     true,
			Level:      plugin_manager.RISKLevel,
			Detail:     fmt.Sprintf("pod cluster dns resolv %v normally", append(p.opt.CheckDomain, "kubernetes.default.svc.cluster.local")),
		})
		klog.Infof("cluster dns check ok")
	}

	ctx = util.GetCtx(time.Second * 10)
	status, err = p.checkDNS(ctx, p.opt.CheckDomain, fmt.Sprintf("%s/etc/resolv.conf", node.HostPath), node.ClientSet)
	dnsStatusGaugeVecSetList = append(dnsStatusGaugeVecSetList, &metric_manager.GaugeVecSet{
		Labels: []string{"host", nodeName, status},
		Value:  float64(1),
	})

	if status != NormalStatus {
		if err != nil {
			result = append(result, plugin_manager.CheckItem{
				ItemName:   pluginName,
				ItemTarget: nodeName,
				Level:      plugin_manager.RISKLevel,
				Normal:     false,
				Detail:     fmt.Sprintf("pod cluster dns failed: %s", err.Error()),
				Status:     status,
			})
			klog.Errorf("host dns check failed: %s %s", status, err.Error())
		}
	} else {
		result = append(result, plugin_manager.CheckItem{
			ItemName:   pluginName,
			ItemTarget: nodeName,
			Level:      plugin_manager.RISKLevel,
			Normal:     true,
			Status:     plugin_manager.NormalStatus,
			Detail:     fmt.Sprintf("pod cluster dns resolv %v normally", p.opt.CheckDomain),
		})
		klog.Infof("host dns check ok")
	}

	p.Result = plugin_manager.CheckResult{
		Items: result,
	}
	metric_manager.RefreshMetric(dnsAvailability, dnsStatusGaugeVecSetList)

	if !p.ready {
		p.ready = true
	}
}

func (p *Plugin) checkDNS(ctx context.Context, domainList []string, path string, clientSet *kubernetes.Clientset) (string, error) {
	status := NormalStatus
	select {
	case <-ctx.Done():
		status = "timeout"
		break

	default:
		ipList := make([]string, 0, 0)

		if path != "" {
			config, _ := dns.ClientConfigFromFile(path)
			ipList = config.Servers
		} else {
			ep, err := clientSet.CoreV1().Endpoints("kube-system").Get(context.Background(), "kube-dns", v1.GetOptions{ResourceVersion: "0"})
			if err != nil {
				klog.Errorf("get dns endpoint failed: %s", err.Error())
				return status, err
			}

			for _, subset := range ep.Subsets {
				for _, address := range subset.Addresses {
					ipList = append(ipList, address.IP)
				}
			}
		}

		if len(ipList) > 0 {
			for ip := 0; ip < len(ipList); ip++ {
				r, err := createResolver(ipList[ip])
				if err != nil {
					klog.Errorf("create resolver failed: %s", err.Error())
					status = "setresolverfailed"
					return status, err
				}

				for _, domain := range domainList {
					if path != "" && strings.Contains(domain, "svc.cluster.local") {
						continue
					}

					latency, err := dnsLookup(r, domain)
					if err != nil {
						klog.Errorf("%s resolve %s failed: %s", ipList[ip], domain, err.Error())
						status = ResolvFailStauts
						return status, err
					} else {
						klog.Errorf("%s resolve %s success", ipList[ip], domain)
					}

					dnsLatency.WithLabelValues().Observe(float64(latency) / float64(time.Second))
				}
			}
		} else {
			status = "noserver"
			err := fmt.Errorf("No available dns server")
			klog.Errorf(err.Error())
			return status, err
		}
	}

	return status, nil
}

func createResolver(ip string) (*net.Resolver, error) {
	r := &net.Resolver{}
	// if we're supplied a null string, return an error
	if len(ip) < 1 {
		return r, fmt.Errorf("Need a valid ip to create Resolver")
	}
	// attempt to create the resolver based on the string
	r = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address2 string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, "udp", ip+":53")
		},
	}
	return r, nil
}

func dnsLookup(r *net.Resolver, host string) (time.Duration, error) {
	start := time.Now()
	addrs, err := r.LookupHost(util.GetCtx(10*time.Second), host)
	if err != nil {
		return 0, fmt.Errorf("DNS Status check determined that %s is DOWN: %s", host, err.Error())
	}

	if len(addrs) == 0 {
		return 0, fmt.Errorf("No host was found")
	}

	return time.Since(start), nil
}

// GetResult return check result by cluster ID
func (p *Plugin) GetResult(string) plugin_manager.CheckResult {
	return p.Result
}

func (p *Plugin) GetDetail() interface{} {
	return p.Detail
}

// Ready return true if cluster check is over
func (p *Plugin) Ready(string) bool {
	return p.ready
}
