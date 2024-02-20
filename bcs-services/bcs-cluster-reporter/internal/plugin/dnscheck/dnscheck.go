/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
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
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"

	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

// Plugin xxx
type Plugin struct {
	stopChan   chan int
	opt        *Options
	checkLock  sync.Mutex
	clusterId  string
	clientSet  *kubernetes.Clientset
	businessID string
	dnsLock    sync.Mutex
	cancel     context.CancelFunc
}

var (
	dnsAvailability = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dns_availability",
		Help: "dns_availability, 1 means OK",
	}, []string{"target", "target_biz", "status"})

	dnsLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "dns_latency",
		Help:    "dns_latency",
		Buckets: []float64{0.001, 0.01, 0.1, 0.2, 0.4, 0.8, 1.6, 3.2},
	}, []string{"target", "target_biz"})
)

func init() {
	metric_manager.Register(dnsAvailability)
	metric_manager.Register(dnsLatency)
}

// Setup xxx
func (p *Plugin) Setup(configFilePath string) error {
	configFileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("read dnscheck config file %s failed, err %s", configFilePath, err.Error())
	}
	p.opt = &Options{}
	if err = json.Unmarshal(configFileBytes, p.opt); err != nil {
		if err = yaml.Unmarshal(configFileBytes, p.opt); err != nil {
			return fmt.Errorf("decode dnscheck config file %s failed, err %s", configFilePath, err.Error())
		}
	}

	if err = p.opt.Validate(); err != nil {
		return err
	}

	p.stopChan = make(chan int)
	interval := p.opt.Interval
	if interval == 0 {
		interval = 60
	}

	clusterConfig := plugin_manager.Pm.GetConfig().InClusterConfig
	if clusterConfig.Config == nil {
		klog.Fatalf("netcheck get incluster config failed, only can run as incluster mode")
	}
	p.clusterId = clusterConfig.ClusterID
	p.businessID = clusterConfig.BusinessID

	p.clientSet, err = k8s.GetClientsetByConfig(clusterConfig.Config)
	if err != nil {
		klog.Fatalf("netcheck get incluster config failed, only can run as incluster mode")
	}

	go func() {
		for {
			if p.checkLock.TryLock() {
				p.checkLock.Unlock()
				if p.opt.Synchronization {
					plugin_manager.Pm.Lock()
				}
				go p.Check()
			} else {
				klog.V(3).Infof("the former dnscheck didn't over, skip in this loop")
			}
			select {
			case result := <-p.stopChan:
				klog.V(3).Infof("stop plugin %s by signal %d", p.Name(), result)
				return
			case <-time.After(time.Duration(interval) * time.Second):
				continue
			}
		}
	}()

	return nil
}

// Stop xxx
func (p *Plugin) Stop() error {
	p.checkLock.Lock()
	p.stopChan <- 1
	klog.Infof("plugin %s stopped", p.Name())
	p.checkLock.Unlock()
	p.cancel()
	return nil
}

// Name xxx
func (p *Plugin) Name() string {
	return "dnscheck"
}

// Check xxx
func (p *Plugin) Check() {
	p.checkLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		if p.opt.Synchronization {
			plugin_manager.Pm.UnLock()
		}
		p.checkLock.Unlock()
	}()

	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("%s dnscheck failed: %s, stack: %v\n", p.clusterId, r, string(debug.Stack()))
		}
	}()

	if p.dnsLock.TryLock() {
		p.dnsLock.Unlock()
		ctx, cancel := context.WithCancel(context.Background())
		p.cancel = cancel
		go p.checkDNSEndpoints(ctx, p.opt.CheckDomain)
	}
}

func (p *Plugin) checkDNSEndpoints(ctx context.Context, domainList []string) {
	p.dnsLock.Lock()
	defer func() {
		p.dnsLock.Unlock()
	}()

	for {
		start := time.Now()
		select {
		case <-ctx.Done():
			klog.Infof("Stop checkDNSEndpoints")
			break

		default:
			status := "ok"
			ep, err := p.clientSet.CoreV1().Endpoints("kube-system").Get(context.Background(), "kube-dns", v1.GetOptions{})
			if err != nil {
				klog.Errorf(err.Error())
				status = "getepfailed"
			}

			ipList := make([]string, 0, 0)
			for _, subset := range ep.Subsets {
				for _, address := range subset.Addresses {
					ipList = append(ipList, address.IP)
				}
			}

			if len(ipList) > 0 {
				for ip := 0; ip < len(ipList); ip++ {
					r, err := createResolver(ipList[ip])
					if err != nil {
						klog.Errorf(err.Error())
						status = "setresolverfailed"
					}

					for _, domain := range domainList {
						latency, err := dnsLookup(r, domain)
						if err != nil {
							klog.Errorf(err.Error())
							status = "resolvefailed"
						}

						dnsLatency.WithLabelValues(p.clusterId, p.businessID).Observe(float64(latency) / float64(time.Second))
					}
				}
			} else {
				status = "noepfound"
				klog.Errorf("No endpoints available for service")
			}

			dnsAvailability.WithLabelValues(p.clusterId, p.businessID, status).Set(1)
		}

		// 最快1s执行一次
		if time.Since(start) < time.Second {
			<-time.After(time.Second - time.Since(start))
		}
	}

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
	addrs, err := r.LookupHost(context.Background(), host)
	if err != nil {
		errorMessage := "DNS Status check determined that " + host + " is DOWN: " + err.Error()
		return 0, fmt.Errorf(errorMessage)
	}

	if len(addrs) == 0 {
		return 0, fmt.Errorf("No host was found")
	}

	return time.Since(start), nil
}
