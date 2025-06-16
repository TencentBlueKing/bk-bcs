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

// Package deployment 提供 Istio 的部署分析器
package deployment

import (
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/analysis"
	"istio.io/istio/pkg/config/analysis/analyzers/util"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/config/schema/gvk"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/msg"
)

// ApplicationUIDAnalyzer 应用UID分析器
type ApplicationUIDAnalyzer struct{}

var _ analysis.Analyzer = &ApplicationUIDAnalyzer{}

const (
	UserID = int64(1337)
)

// Metadata 返回分析器的元数据
func (appUID *ApplicationUIDAnalyzer) Metadata() analysis.Metadata {
	return analysis.Metadata{
		Name:        "applicationUID.Analyzer",
		Description: "Checks invalid application UID",
		Inputs: []config.GroupVersionKind{
			gvk.Pod,
			gvk.Deployment,
		},
	}
}

// Analyze 分析器的主要逻辑
func (appUID *ApplicationUIDAnalyzer) Analyze(context analysis.Context) {
	context.ForEach(gvk.Pod, func(resource *resource.Instance) bool {
		appUID.analyzeAppUIDForPod(resource, context)
		return true
	})
	context.ForEach(gvk.Deployment, func(resource *resource.Instance) bool {
		appUID.analyzeAppUIDForDeployment(resource, context)
		return true
	})
}

func (appUID *ApplicationUIDAnalyzer) analyzeAppUIDForPod(resource *resource.Instance, context analysis.Context) {
	p := resource.Message.(*v1.PodSpec)
	// Skip analyzing control plane for IST0144
	if util.IsIstioControlPlane(resource) {
		return
	}
	message := msg.NewInvalidApplicationUID(resource)

	if p.SecurityContext != nil && p.SecurityContext.RunAsUser != nil {
		if *p.SecurityContext.RunAsUser == UserID {
			context.Report(gvk.Pod, message)
		}
	}
	for _, container := range p.Containers {
		if container.Name != util.IstioProxyName && container.Name != util.IstioOperator {
			if container.SecurityContext != nil && container.SecurityContext.RunAsUser != nil {
				if *container.SecurityContext.RunAsUser == UserID {
					context.Report(gvk.Pod, message)
				}
			}
		}
	}
}

func (appUID *ApplicationUIDAnalyzer) analyzeAppUIDForDeployment(resource *resource.Instance, context analysis.Context) {
	d := resource.Message.(*appsv1.DeploymentSpec)
	// Skip analyzing control plane for IST0144
	if util.IsIstioControlPlane(resource) {
		return
	}
	message := msg.NewInvalidApplicationUID(resource)
	spec := d.Template.Spec

	if spec.SecurityContext != nil && spec.SecurityContext.RunAsUser != nil {
		if *spec.SecurityContext.RunAsUser == UserID {
			context.Report(gvk.Deployment, message)
		}
	}
	for _, container := range spec.Containers {
		if container.Name != util.IstioProxyName && container.Name != util.IstioOperator {
			if container.SecurityContext != nil && container.SecurityContext.RunAsUser != nil {
				if *container.SecurityContext.RunAsUser == UserID {
					context.Report(gvk.Deployment, message)
				}
			}
		}
	}
}
