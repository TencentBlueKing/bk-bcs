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

// Package logrecorder xxx
package logrecorder

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"

	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog"
)

// Plugin xxx
type Plugin struct {
	stopChan  chan int
	opt       *Options
	checkLock sync.Mutex
}

var (
	etcdTookTooLongMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "etcd_took_too_long",
		Help: "etcd_took_too_long",
	}, []string{"target", "target_biz", "request"})
	stopFlag = false
)

func init() {
	metric_manager.Register(etcdTookTooLongMetric)
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
			return fmt.Errorf("decode logrecorder config file %s failed, err %s", configFilePath, err.Error())
		}
	}

	if err = p.opt.Validate(); err != nil {
		return err
	}

	p.stopChan = make(chan int)

	cluster := plugin_manager.Pm.GetConfig().InClusterConfig
	if cluster.Config == nil {
		klog.Fatalf("eventrecorder get incluster config failed")
	}

	getLogRecord(cluster)

	return nil
}

// Stop xxx
func (p *Plugin) Stop() error {
	p.checkLock.Lock()
	// p.stopChan <- 1
	stopFlag = true
	klog.Infof("plugin %s stopped", p.Name())
	p.checkLock.Unlock()
	return nil
}

// Name xxx
func (p *Plugin) Name() string {
	return "logrecorder"
}

var (
	// pod container log
	logCacheLock sync.RWMutex
	logCache     map[string]map[string][]PodLog
)

// PodLog xxx
type PodLog struct {
	LogTime time.Time
	Log     string
}

func getLogRecord(cluster plugin_manager.ClusterConfig) {
	logCache = make(map[string]map[string][]PodLog)
	clientSet, err := k8s.GetClientsetByConfig(cluster.Config)
	if err != nil {
		klog.Fatalf("logrecorder getLogMetric failed: %s", err.Error())
	}

	podList, err := clientSet.CoreV1().Pods("kube-system").List(context.Background(), v1.ListOptions{
		ResourceVersion: "0",
		LabelSelector:   "component=etcd",
	})

	if err != nil {
		klog.Fatalf("logrecorder getLogMetric failed: %s", err.Error())
	}

	// 创建日志选项
	var sinceSec int64 = 300
	logOptions := &corev1.PodLogOptions{
		Container:    "etcd",
		Follow:       true,
		Timestamps:   true,
		SinceSeconds: &sinceSec,
	}

	// 获取容器的标准输出日志
	for _, pod := range podList.Items {
		// 初始化对应pod的日志缓存
		podName := pod.Name
		containerName := "etcd"
		_, ok := logCache[pod.Name]
		if !ok {
			logCache[pod.Name] = make(map[string][]PodLog)
			logCache[pod.Name]["etcd"] = make([]PodLog, 0, 0)
		}

		logsStream, err := clientSet.CoreV1().Pods("kube-system").GetLogs(pod.Name, logOptions).Stream(context.Background())
		if err != nil {
			klog.Fatal(err.Error())
		}
		scanner := bufio.NewScanner(logsStream)
		go func(podName, containerName string) {
			defer func() {
				logsStream.Close()
			}()

			for {
				if scanner.Scan() {
					line := scanner.Text()
					CacheLog(podName, containerName, line)
				} else {
					if err = scanner.Err(); err != nil {
						klog.Infof(err.Error())
					}
					// 重建链接
					logsStream.Close()
					logsStream, err = clientSet.CoreV1().Pods("kube-system").GetLogs(pod.Name, logOptions).Stream(context.Background())
					if err != nil {
						klog.Fatalf(err.Error())
					}
					scanner = bufio.NewScanner(logsStream)
				}

				if stopFlag {
					stopFlag = false
					return
				}

			}
		}(podName, containerName)
	}

	go func() {
		for {
			select {
			case <-time.After(1 * time.Minute):
				LogAnalysis(plugin_manager.Pm.GetConfig().InClusterConfig)
			}
		}
	}()
}

func cleanPodLogCache() {
	for podName, podLogs := range logCache {
		for containerName, logList := range podLogs {
			newLogList := make([]PodLog, 0, 0)
			for _, logItem := range logList {
				if !logItem.LogTime.Before(time.Now().Add(-5 * time.Minute)) {
					newLogList = append(newLogList, logItem)
				}
			}
			logCacheLock.Lock()
			logCache[podName][containerName] = newLogList
			logCacheLock.Unlock()
		}
	}
}

// CacheLog xxx
func CacheLog(podName, containerName, log string) {
	if t, err := time.Parse("2006-01-02T15:04:05.999999999Z", strings.Split(log, " ")[0]); err == nil {
		logCacheLock.Lock()
		logCache[podName][containerName] = append(logCache[podName][containerName],
			PodLog{
				LogTime: t,
				Log:     strings.SplitN(log, " ", 2)[1],
			})
		logCacheLock.Unlock()
	} else {
		klog.Infof(err.Error())
	}
}

// LogAnalysis xxx
func LogAnalysis(cluster plugin_manager.ClusterConfig) {
	// 清理日志缓存
	cleanPodLogCache()

	logCacheLock.RLock()
	defer func() {
		logCacheLock.RUnlock()
	}()

	etcdTookTooLongGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)

	tookToLongList := make(map[string]int)

	for podName, podLogs := range logCache {
		for containerName, logList := range podLogs {
			if strings.Contains(podName, "etcd") && containerName == "etcd" {
				for _, logItem := range logList {
					logMap := make(map[string]interface{})
					err := json.Unmarshal([]byte(strings.Replace(logItem.Log, "-", "", -1)), &logMap)
					if err != nil {
						klog.Errorf("unmarshal failed: %s, %s", err.Error(), logItem.Log)
						break
					}
					if strings.Contains(logMap["msg"].(string), "took too long") {
						var requestPath string
						for _, str := range strings.Split(logMap["request"].(string), " ") {
							if strings.Contains(str, "key:") {
								requestPath = strings.Replace(str, "key:", "", -1)
								requestPath = strings.Replace(requestPath, "\"", "", -1)
								break
							}

						}

						var responseSize string
						if strings.Contains(logMap["response"].(string), " ") {
							responseSize = strings.Replace(
								strings.Split(logMap["response"].(string), " ")[1], "size:", "", -1)
						} else {
							responseSize = strings.Replace(logMap["response"].(string), "size:", "", -1)
						}

						size, err := strconv.Atoi(responseSize)
						if err != nil {
							klog.Errorf(err.Error())
							continue
						}

						if _, ok := tookToLongList[requestPath]; !ok {
							tookToLongList[requestPath] += size
						}
					}

				}

			}
		}
	}

	// top 5
	for request, size := range tookToLongList {
		if len(etcdTookTooLongGaugeVecSetList) == 0 {
			etcdTookTooLongGaugeVecSetList = append(etcdTookTooLongGaugeVecSetList,
				&metric_manager.GaugeVecSet{Labels: []string{cluster.ClusterID, cluster.BusinessID, request}, Value: float64(size)})
		} else {
			for index := len(etcdTookTooLongGaugeVecSetList) - 1; index >= 0; index-- {
				if float64(size) > etcdTookTooLongGaugeVecSetList[index].Value {
					if index+1 < len(etcdTookTooLongGaugeVecSetList) {
						etcdTookTooLongGaugeVecSetList = append(etcdTookTooLongGaugeVecSetList[:index+1],
							append(
								[]*metric_manager.GaugeVecSet{{Labels: []string{cluster.ClusterID,
									cluster.BusinessID, request}, Value: float64(size)}},
								etcdTookTooLongGaugeVecSetList[index+1:]...)...)
					} else {
						etcdTookTooLongGaugeVecSetList = append(etcdTookTooLongGaugeVecSetList,
							&metric_manager.GaugeVecSet{Labels: []string{cluster.ClusterID, cluster.BusinessID, request},
								Value: float64(size)})
					}
					break
				}
			}
		}
	}
	if len(etcdTookTooLongGaugeVecSetList) < 5 {
		metric_manager.SetMetric(etcdTookTooLongMetric, etcdTookTooLongGaugeVecSetList)
	} else {
		metric_manager.SetMetric(etcdTookTooLongMetric,
			etcdTookTooLongGaugeVecSetList[len(etcdTookTooLongGaugeVecSetList)-5:])
	}
}
