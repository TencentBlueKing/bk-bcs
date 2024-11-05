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

// Package uploader xxx
package uploader

import (
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

// Plugin uploader
type Plugin struct {
	opt *Options
	pluginmanager.NodePlugin
	ready bool
}

// GetDetail xxx
func (p *Plugin) GetDetail() interface{} {
	return nil
}

// Setup plugin
func (p *Plugin) Setup(configFilePath string, runMode string) error {
	p.opt = &Options{}
	err := util.ReadorInitConf(configFilePath, p.opt, initContent)
	if err != nil {
		return err
	}

	if err = p.opt.Validate(); err != nil {
		return err
	}

	interval := p.opt.Interval
	if interval == 0 {
		interval = 60
	}

	// run as daemon
	if runMode == pluginmanager.RunModeDaemon {
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
	} else if runMode == pluginmanager.RunModeOnce {
		p.Check()
	}

	return nil
}

// Check xxx
func (p *Plugin) Check() {
	p.CheckLock.Lock()
	klog.Infof("start %s", p.Name())
	p.ready = false
	defer func() {
		klog.Infof("end %s", p.Name())
		p.CheckLock.Unlock()
		p.ready = true
	}()

	pluginstr := strings.Replace(pluginmanager.Pm.GetPluginstr(), p.Name(), "", 1)
	pluginstr = strings.Replace(pluginstr, ",,", ",", 1)
	pluginmanager.Pm.Ready(pluginstr, "")
	checkResult := pluginmanager.Pm.GetNodeResult(pluginmanager.Pm.GetPluginstr())
	checkDetail := pluginmanager.Pm.GetNodeDetail(pluginmanager.Pm.GetPluginstr())
	uploadResult := make(map[string]pluginmanager.PluginInfo)

	for _, name := range strings.Split(pluginmanager.Pm.GetPluginstr(), ",") {
		uploadResult[name] = pluginmanager.PluginInfo{
			Result: checkResult[name],
			Detail: checkDetail[name],
		}
	}

	nodeinfoData, _ := yaml.Marshal(uploadResult)

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		klog.Errorf("env NODE_NAME must be set, skip upload")
		return
	}

	restConfig, err := k8s.GetRestConfig()
	if err != nil {
		klog.Errorf(err.Error())
		return
	}

	cs, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		klog.Errorf(err.Error())
		return
	}

	if p.opt.Type == "k8s" {
		ctx := util.GetCtx(time.Second * 10)

		cmList, err := cs.CoreV1().ConfigMaps(p.opt.Namespace).List(ctx, v1.ListOptions{
			ResourceVersion: "0", LabelSelector: "nodeagent=" + nodeName})
		if err != nil {
			klog.Errorf(err.Error())
			return
		}

		sort.Slice(cmList.Items, func(i, j int) bool {
			return cmList.Items[i].Name > cmList.Items[j].Name
		})

		// 获取当前所有的configmap
		versionCMList := make(map[int]corev1.ConfigMap)
		for _, cm := range cmList.Items {
			if !strings.Contains(cm.Name, nodeName+"-v") {
				continue
			}

			if len(strings.Split(cm.Name, nodeName+"-v")) != 2 {
				continue
			}

			version, err := strconv.Atoi(strings.Split(cm.Name, nodeName+"-v")[1])
			if err != nil {
				klog.Errorf(err.Error())
				return
			}

			if version > p.opt.CopyNum {
				ctx = util.GetCtx(time.Second * 10)
				err = cs.CoreV1().ConfigMaps(p.opt.Namespace).Delete(ctx, cm.Name, v1.DeleteOptions{})
				if err != nil {
					klog.Errorf(err.Error())
					return
				}

			}
			versionCMList[version] = cm
		}

		for version, cm := range versionCMList {
			if version == p.opt.CopyNum {
				//版本最大的configmap无需处理
				continue
			}
			//滚动当前configmap的内容
			if targetCM, ok := versionCMList[version+1]; ok {
				//apiVersion := "v1"
				//kind := "ConfigMap"

				targetCM.Data = cm.Data

				ctx = util.GetCtx(time.Second * 10)
				_, err = cs.CoreV1().ConfigMaps(p.opt.Namespace).Update(ctx, &targetCM, v1.UpdateOptions{})
				if err != nil {
					klog.Errorf(err.Error())
					return
				}

				if err != nil {
					klog.Errorf(err.Error())
					return
				}

			} else {
				newCm := corev1.ConfigMap{
					ObjectMeta: v1.ObjectMeta{
						Name:      nodeName + "-v" + strconv.Itoa(version+1),
						Namespace: p.opt.Namespace,
						Labels: map[string]string{
							"nodeagent": nodeName,
						},
					},
					Data: cm.Data,
				}

				ctx = util.GetCtx(time.Second * 10)
				_, err = cs.CoreV1().ConfigMaps(p.opt.Namespace).Create(ctx, &newCm, v1.CreateOptions{})
				if err != nil {
					klog.Errorf(err.Error())
					return
				}
			}

		}

		_, ok := versionCMList[1]

		// 如果**-v1不存在则创建
		if len(versionCMList) == 0 || !ok {
			// 新建configmap
			newCm := corev1.ConfigMap{
				ObjectMeta: v1.ObjectMeta{
					Name:      nodeName + "-v1",
					Namespace: p.opt.Namespace,
					Labels: map[string]string{
						"nodeagent": nodeName,
					},
				},
			}

			newCm.Data = map[string]string{
				"nodeinfo":   string(nodeinfoData),
				"updateTime": time.Now().Format("2006-01-02 15:04:05.999999999 -0700 MST"),
				"nodename":   nodeName,
			}

			ctx = util.GetCtx(time.Second * 10)
			_, err = cs.CoreV1().ConfigMaps(p.opt.Namespace).Create(ctx, &newCm, v1.CreateOptions{})
			if err != nil {
				klog.Errorf(err.Error())
				return
			}
		} else {
			// 存在则直接patch
			//apiVersion := "v1"
			//kind := "ConfigMap"
			targetCM := versionCMList[1]

			targetCM.Data = map[string]string{
				"nodeinfo":   string(nodeinfoData),
				"updateTime": time.Now().Format("2006-01-02 15:04:05.999999999 -0700 MST"),
				"nodename":   nodeName,
			}

			ctx = util.GetCtx(time.Second * 10)
			_, err = cs.CoreV1().ConfigMaps(p.opt.Namespace).Update(ctx, &targetCM, v1.UpdateOptions{})
			if err != nil {
				klog.Errorf(err.Error())
				return
			}

			if err != nil {
				klog.Errorf(err.Error())
				return
			}
		}
	}
}

// Name xxx
func (p *Plugin) Name() string {
	return pluginName
}

// Ready xxx
func (p *Plugin) Ready(string) bool {
	return p.ready
}

// GetResult xxx
func (p *Plugin) GetResult(string) pluginmanager.CheckResult {
	return pluginmanager.CheckResult{}
}

// Stop xxx
func (p *Plugin) Stop() error {
	p.StopChan <- 0
	return nil
}

// GetString xxx
func (p *Plugin) GetString(key string) string {
	return StringMap[key]
}
