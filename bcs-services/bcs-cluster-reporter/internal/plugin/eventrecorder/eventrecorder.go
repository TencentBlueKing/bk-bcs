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

// Package eventrecorder xxx
package eventrecorder

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"

	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

// Plugin xxx
type Plugin struct {
	stopChan  chan int
	opt       *Options
	checkLock sync.Mutex
	// eventChecktGaugeVecSetMap map[string]map[string]map[string]*metric_manager.GaugeVecSet
	eventChecktGaugeVecSetMap map[string]map[string]*metric_manager.GaugeVecSet
}

var (
	eventRecord = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_record",
		Help: "event_record, count of event",
	}, []string{"target", "target_biz", "resource_kind", "event_reason"})
)

func init() {
	metric_manager.Register(eventRecord)
}

// Setup xxx
func (p *Plugin) Setup(configFilePath string) error {
	configFileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("read dnscheck config file %s failed, err %s", configFilePath, err.Error())
	}
	p.opt = &Options{}
	if err = yaml.Unmarshal(configFileBytes, p.opt); err != nil {
		if err = yaml.Unmarshal(configFileBytes, p.opt); err != nil {
			return fmt.Errorf("decode eventrecorder config file %s failed, err %s", configFilePath, err.Error())
		}
	}

	if err = p.opt.Validate(); err != nil {
		return err
	}

	p.stopChan = make(chan int)

	// 开始获取数据
	// p.eventChecktGaugeVecSetMap = make(map[string]map[string]map[string]*metric_manager.GaugeVecSet)
	p.eventChecktGaugeVecSetMap = make(map[string]map[string]*metric_manager.GaugeVecSet)
	cluster := plugin_manager.Pm.GetConfig().InClusterConfig
	if cluster.Config == nil {
		klog.Fatalf("eventrecorder get incluster config failed")
	}

	go func() {
		recordEvent(p.eventChecktGaugeVecSetMap, cluster, p.stopChan)
	}()

	interval := p.opt.Interval
	if interval == 0 {
		interval = 60
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
				klog.V(3).Infof("the former eventrecorder didn't over, skip in this loop")
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
	p.stopChan <- 1
	klog.Infof("plugin %s stopped", p.Name())
	p.checkLock.Unlock()
	return nil
}

// Name xxx
func (p *Plugin) Name() string {
	return "eventrecorder"
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
		metric_manager.SetCommonDurationMetric([]string{"eventrecorder", "", "", ""}, start)
	}()
	metric_manager.SetMetric(eventRecord, flatten(p.eventChecktGaugeVecSetMap))
}

func flatten(data interface{}) []*metric_manager.GaugeVecSet {
	result := make([]*metric_manager.GaugeVecSet, 0, 0)

	if m, ok := data.(map[string]map[string]*metric_manager.GaugeVecSet); ok {
		for _, value := range m {
			result = append(result, flatten(value)...)
		}
	} else {
		for _, value := range data.(map[string]*metric_manager.GaugeVecSet) {
			result = append(result, value)
		}
	}

	return result
}

func recordEvent(eventChecktGaugeVecSetMap map[string]map[string]*metric_manager.GaugeVecSet,
	cluster plugin_manager.ClusterConfig, stopChan <-chan int) {
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("%s eventrecorder failed: %s, stack: %v\n", cluster.ClusterID, r, string(debug.Stack()))
			clientSet1, _ := k8s.GetClientsetByConfig(cluster.Config)
			var responseContentType string
			body, _ := clientSet1.RESTClient().Get().
				AbsPath("/apis").
				SetHeader("Accept", "application/json").
				Do(context.TODO()).
				ContentType(&responseContentType).
				Raw()
			klog.V(3).Infof("Try get apis for %s: %s", cluster.ClusterID, string(body))
		}
	}()

	clientSet, err := k8s.GetClientsetByConfig(cluster.Config)
	if err != nil {
		klog.Fatalf("eventrecorder GetClientsetByConfig failed: %s", err.Error())
	}

	clientSet.CoreV1().Events(metav1.NamespaceAll).Watch(context.Background(), metav1.ListOptions{})
	factory := informers.NewSharedInformerFactoryWithOptions(clientSet, time.Second*60,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.ResourceVersion = "0"
		}))

	eventInformer := factory.Core().V1().Events().Informer()
	_, err = eventInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if event, ok := obj.(*corev1.Event); ok {
				// 只记录非Normal的event
				if event.Type == "Normal" {
					return
				}
				if eventChecktGaugeVecSetMap[event.Kind] == nil {
					eventChecktGaugeVecSetMap[event.Kind] = make(map[string]*metric_manager.GaugeVecSet)
				}
				if eventChecktGaugeVecSetMap[event.Kind][event.Reason] == nil {
					eventChecktGaugeVecSetMap[event.Kind][event.Reason] = &metric_manager.GaugeVecSet{Labels: []string{
						cluster.ClusterID, cluster.BusinessID,
						event.InvolvedObject.Kind, event.Reason}, Value: float64(1)}
				} else {
					eventChecktGaugeVecSetMap[event.Kind][event.Reason].Value++
				}

			} else {
				klog.Infof("unknown obj: %s", obj)
			}

		},
	})
	if err != nil {
		klog.Fatalf("eventrecorder AddEventHandler failed: %s", err.Error())
	}

	informerStopChan := make(<-chan struct{})
	factory.Start(informerStopChan)
	if !cache.WaitForCacheSync(informerStopChan, eventInformer.HasSynced) {
		klog.Fatalf("Timed out waiting for event caches to sync")
	}

	<-stopChan
}
