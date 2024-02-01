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

// Package netcheck xxx
package netcheck

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
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

// Plugin  xxx
type Plugin struct {
	stopChan   chan int
	opt        *Options
	checkLock  sync.Mutex
	clusterId  string
	clientSet  *kubernetes.Clientset
	businessID string
	conn       *icmp.PacketConn
	dnsConn    *icmp.PacketConn
	msg        []byte
	svcNetLock sync.Mutex
	cancel     context.CancelFunc
}

var (
	// NewGaugeVec creates a new GaugeVec based on the provided GaugeOpts and
	// partitioned by the given label names.
	podNetAvailability = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pod_net_availability",
		Help: "pod_net_availability, 1 means OK",
	}, []string{"target", "target_biz", "status"})

	// NewHistogramVec creates a new HistogramVec based on the provided HistogramOpts and
	// partitioned by the given label names.
	podNetLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "pod_net_latency",
		Help:    "pod_net_latency",
		Buckets: []float64{0.001, 0.01, 0.1, 0.2, 0.4, 0.8, 1.6, 3.2},
	}, []string{"target", "target_biz"})

	// NewGaugeVec creates a new GaugeVec based on the provided GaugeOpts and
	// partitioned by the given label names.
	svcNetAvailability = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "svc_net_availability",
		Help: "svc_net_availability, 1 means OK",
	}, []string{"target", "target_biz", "status"})

	// NewHistogramVec creates a new HistogramVec based on the provided HistogramOpts and
	// partitioned by the given label names.
	svcNetLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "svc_net_latency",
		Help:    "pod_net_latency",
		Buckets: []float64{0.001, 0.01, 0.1, 0.2, 0.4, 0.8, 1.6, 3.2},
	}, []string{"target", "target_biz"})
)

func init() {
	metric_manager.Register(podNetAvailability)
	metric_manager.Register(podNetLatency)
	metric_manager.Register(svcNetAvailability)
	metric_manager.Register(svcNetLatency)
}

// Setup  xxx
func (p *Plugin) Setup(configFilePath string) error {
	configFileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("read netcheck config file %s failed, err %s", configFilePath, err.Error())
	}
	p.opt = &Options{}
	if err = json.Unmarshal(configFileBytes, p.opt); err != nil {
		if err = yaml.Unmarshal(configFileBytes, p.opt); err != nil {
			return fmt.Errorf("decode netcheck config file %s failed, err %s", configFilePath, err.Error())
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

	p.conn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		klog.Fatalf(err.Error())
	}
	p.dnsConn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		klog.Fatalf(err.Error())
	}

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("Hello, world!"),
		},
	}

	p.msg, err = msg.Marshal(nil)
	if err != nil {
		klog.Fatalf(err.Error())
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
				klog.V(3).Infof("the former netcheck didn't over, skip in this loop")
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
	return "netcheck"
}

// Check xxx
func (p *Plugin) Check() {
	start := time.Now()
	p.checkLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		if p.opt.Synchronization {
			plugin_manager.Pm.UnLock()
		}
		p.checkLock.Unlock()
		metric_manager.SetCommonDurationMetric([]string{"netcheck", "", "", ""}, start)
	}()

	netChecktGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)

	status := "error"
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("%s netcheck failed: %s, stack: %v\n", p.clusterId, r, string(debug.Stack()))
			status = "panic"
		}
	}()

	p.checkPodNet(&status)
	klog.Infof("%s netcheck result %s", p.clusterId, status)
	netChecktGaugeVecSetList = append(netChecktGaugeVecSetList,
		&metric_manager.GaugeVecSet{Labels: []string{p.clusterId, p.businessID, status}, Value: float64(1)})
	metric_manager.SetMetric(podNetAvailability, netChecktGaugeVecSetList)

	if p.svcNetLock.TryLock() {
		p.svcNetLock.Unlock()
		ctx, cancel := context.WithCancel(context.Background())
		p.cancel = cancel
		go p.checkSVCNet(ctx)
	}
}

func (p *Plugin) checkPodNet(status *string) {
	failedToPing := false
	defer func() {
		if !failedToPing {
			*status = "ok"
		}
	}()

	pods, err := p.clientSet.CoreV1().Pods(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		ResourceVersion: "0",
	})
	if err != nil {
		*status = "getpodfailed"
		klog.Errorf("%s failed to list all pods, %s", p.clusterId, err.Error())
	}

	wg := sync.WaitGroup{}

	pingChan := make(chan struct{}, 5)
	for _, pod := range pods.Items {
		if pod.Status.Phase != v1.PodRunning || pod.Spec.HostNetwork {
			continue
		}

		wg.Add(1)
		plugin_manager.Pm.Add()
		go func(pod v1.Pod) {
			start := time.Now()
			pingChan <- struct{}{}
			defer func() {
				wg.Done()
				plugin_manager.Pm.Done()
			}()

			conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
			if err != nil {
				klog.Fatalf(err.Error())
			}
			err = conn.SetDeadline(start.Add(60 * time.Second))
			if err != nil {
				klog.Errorf(err.Error())
			}

			podIP := pod.Status.PodIP

			_, err = conn.WriteTo(p.msg, &net.IPAddr{IP: net.ParseIP(podIP)})
			if err != nil {
				failedToPing = true
				klog.Errorf(err.Error())
				*status = "sendfailed"
				return
			}

			reply := make([]byte, 1500)
			_, _, err = conn.ReadFrom(reply)
			duration := time.Since(start)
			if err != nil {
				failedToPing = true
				klog.Errorf("read reply from %s:%s:%s failed: %s", pod.Namespace, pod.Name, podIP, err.Error())
				*status = "readfailed"
				return
			}

			//  统一指标配置方法
			podNetLatency.WithLabelValues(p.clusterId, p.businessID).Observe(float64(duration) / float64(time.Second))
			<-pingChan
		}(pod)
	}

	wg.Wait()
}

func (p *Plugin) checkSVCNet(ctx context.Context) {
	p.svcNetLock.Lock()
	defer func() {
		p.svcNetLock.Unlock()
	}()

	svc, err := p.clientSet.CoreV1().Services(metav1.NamespaceSystem).Get(context.TODO(), "kube-dns", metav1.GetOptions{})
	if err != nil {
		klog.Fatalf("%s failed to list all pods, %s", p.clusterId, err.Error())
	}

	failedToPing := false

	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			klog.Infof("Stop checkSVCNet")
			return
		default:
			err = p.dnsConn.SetDeadline(start.Add(60 * time.Second))
			if err != nil {
				klog.Errorf(err.Error())
			}

			_, err = p.dnsConn.WriteTo(p.msg, &net.IPAddr{IP: net.ParseIP(svc.Spec.ClusterIP)})
			if err != nil {
				failedToPing = true
				klog.Errorf(err.Error())
			}

			reply := make([]byte, 1500)
			_, _, err = p.dnsConn.ReadFrom(reply)
			if err != nil {
				failedToPing = true
				klog.Errorf("read reply from %s:%s:%s failed: %s", svc.Namespace, svc.Name, svc.Spec.ClusterIP, err.Error())
			}
			duration := time.Since(start)
			svcNetLatency.WithLabelValues(p.clusterId, p.businessID).Observe(float64(duration) / float64(time.Second))

			svcNetAvailability.Reset()
			if failedToPing {
				svcNetAvailability.WithLabelValues(p.clusterId, p.businessID, "notok").Set(1)
			} else {
				svcNetAvailability.WithLabelValues(p.clusterId, p.businessID, "ok").Set(1)
			}
		}

		// 最快1s执行一次
		if time.Since(start) < time.Second {
			<-time.After(time.Second - time.Since(start))
		}
		start = time.Now()
	}
}
