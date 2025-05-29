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

// Package istio is the plugin for istio.
package istio

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	"istio.io/istio/pkg/config/analysis"
	"istio.io/istio/pkg/config/analysis/local"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/kube"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	messageConfig "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/msg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"
)

// Plugin implements pluginmanager.Plugin for Istio rule checks
// 继承ClusterPlugin，便于多集群处理
type Plugin struct {
	pluginmanager.ClusterPlugin
	opt *Options
}

// Name 插件名称
func (p *Plugin) Name() string {
	return "istiocheck"
}

// Setup 初始化插件
func (p *Plugin) Setup(configFilePath string, runMode string) error {
	configFileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("read clustercheck config file %s failed, err %s", configFilePath, err.Error())
	}
	p.opt = &Options{}
	if err = json.Unmarshal(configFileBytes, p.opt); err != nil {
		if err = yaml.Unmarshal(configFileBytes, p.opt); err != nil {
			return fmt.Errorf("decode clustercheck config file %s failed, err %s", configFilePath, err.Error())
		}
	}

	if err = p.opt.Validate(); err != nil {
		klog.Errorf("istiocheck plugin: validate options failed, err %s", err.Error())
		return err
	}

	p.Result = make(map[string]pluginmanager.CheckResult)
	p.ReadyMap = make(map[string]bool)

	return nil
}

// Stop 停止插件
func (p *Plugin) Stop() error {
	p.StopChan <- 1
	return nil
}

// Ready 检查插件是否准备就绪
func (p *Plugin) Ready(clusterID string) bool {
	return true
}

// GetResult 获取插件结果
func (p *Plugin) GetResult(clusterID string) pluginmanager.CheckResult {
	klog.Infof("istio plugin: get result for %s", clusterID)

	messages, err := p.analyze(clusterID)
	if err != nil {
		klog.Errorf("istio plugin: analyze failed for %s: %v", clusterID, err)
		return pluginmanager.CheckResult{}
	}

	result := pluginmanager.CheckResult{}
	checkItems := make([]pluginmanager.CheckItem, 0)
	// 转为CheckItem
	for _, msg := range messages.Messages {
		detail := fmt.Sprintf("%s %s", msg.Origin(), fmt.Sprintf(msg.Type.Template(), msg.Parameters...))
		friendlyName := messageConfig.CodeToFriendlyName[msg.Type.Code()]
		itemName := fmt.Sprintf("[%s]%s", msg.Type.Code(), friendlyName)
		checkItems = append(checkItems, pluginmanager.CheckItem{
			ItemName:   itemName,
			ItemTarget: msg.Resource.Metadata.FullName.String(),
			Detail:     "",
			Level:      msg.Type.Level().String(),
			Normal:     msg.Type.Level().String() == "Info",
			Status:     detail,
			// Status:     msg.Type.Code(),
		})
	}
	result.Items = checkItems
	p.Result[clusterID] = result
	p.ReadyMap[clusterID] = true

	return result
}

// GetDetail 获取插件详情
func (p *Plugin) GetDetail() interface{} {
	return nil
}

// analyze 执行插件检查
func (p *Plugin) analyze(clusterID string) (local.AnalysisResult, error) {
	clusterConfig := pluginmanager.Pm.GetConfig().ClusterConfigs[clusterID]
	if clusterConfig == nil {
		return local.AnalysisResult{}, fmt.Errorf("istio plugin: cluster %s not found", clusterID)
	}
	// 检查istio-system是否存在
	_, err := clusterConfig.ClientSet.CoreV1().Namespaces().Get(context.Background(), p.opt.IstioNamespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Errorf("istio plugin: istio-system namespace not found for %s", clusterID)
			return local.AnalysisResult{}, err
		}
		klog.Errorf("istio plugin: get istio-system namespace for %s failed: %v", clusterID, err)
		return local.AnalysisResult{}, err
	}

	cancel := make(chan struct{})

	// 2. 构造 analyzer
	analyzers := analysis.Combine("Combine", p.opt.enabledAnalyzersObject...)
	sa := local.NewIstiodAnalyzer(analyzers, "", resource.Namespace(p.opt.IstioNamespace), nil)

	k, err := kube.NewClient(kube.NewClientConfigForRestConfig(clusterConfig.Config), "")
	if err != nil {
		klog.Errorf("istio plugin: create dynamic client for %s failed: %v", clusterID, err)
		return local.AnalysisResult{}, err
	}
	// 获取集群的rev
	revision, err := p.getIstioRevision(clusterID)
	if err != nil {
		klog.Errorf("istio plugin: get istio revision for %s failed: %v", clusterID, err)
		return local.AnalysisResult{}, err
	}
	sa.AddRunningKubeSourceWithRevision(k, revision)

	// 3. 执行分析
	messages, err := sa.Analyze(cancel)
	if err != nil {
		klog.Errorf("istio plugin: analyze failed for %s: %v", clusterID, err)
		return local.AnalysisResult{}, err
	}
	klog.Infof("istio plugin: analyze result for %s: %v", clusterID, messages)
	return messages, nil
}

func (p *Plugin) getIstioRevision(clusterID string) (string, error) {
	clusterConfig := pluginmanager.Pm.GetConfig().ClusterConfigs[clusterID]
	if clusterConfig == nil {
		return "", fmt.Errorf("istio plugin: cluster %s not found", clusterID)
	}

	// 查询所有 MutatingWebhookConfiguration
	webhookList, err := clusterConfig.ClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("istio plugin: list MutatingWebhookConfigurations failed: %v", err)
	}
	for _, wh := range webhookList.Items {
		if rev, ok := wh.Labels["istio.io/rev"]; ok {
			return rev, nil
		}
	}
	// 没有找到 revision，返回空字符串
	return "", nil
}

// Check 执行插件检查
func (p *Plugin) Check(pluginmanager.CheckOption) {

}
