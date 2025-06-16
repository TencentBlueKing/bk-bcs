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

package injection

import (
	"strings"

	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/analysis"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/config/schema/gvk"
	admitv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/msg"
)

// ImageAutoAnalyzer reports an error if Pods and Deployments with `image: auto` are not going to be injected.
type ImageAutoAnalyzer struct{}

var _ analysis.Analyzer = &ImageAutoAnalyzer{}

const (
	istioProxyContainerName = "istio-proxy"
	manualInjectionImage    = "auto"
)

// Metadata implements Analyzer.
func (a *ImageAutoAnalyzer) Metadata() analysis.Metadata {
	return analysis.Metadata{
		Name:        "injection.ImageAutoAnalyzer",
		Description: "Makes sure that Pods and Deployments with `image: auto` are going to be injected",
		Inputs: []config.GroupVersionKind{
			gvk.Namespace,
			gvk.Pod,
			gvk.Deployment,
			gvk.MutatingWebhookConfiguration,
		},
	}
}

// Analyze implements Analyzer.
func (a *ImageAutoAnalyzer) Analyze(c analysis.Context) {
	var istioWebhooks []admitv1.MutatingWebhook
	c.ForEach(gvk.MutatingWebhookConfiguration, func(resource *resource.Instance) bool {
		mwhc := resource.Message.(*admitv1.MutatingWebhookConfiguration)
		for _, wh := range mwhc.Webhooks {
			if strings.HasSuffix(wh.Name, "istio.io") {
				istioWebhooks = append(istioWebhooks, wh)
			}
		}
		return true
	})
	c.ForEach(gvk.Pod, func(resource *resource.Instance) bool {
		p := resource.Message.(*v1.PodSpec)
		// If a pod has `image: auto` it is broken whether the webhooks match or not
		if !hasAutoImage(p) {
			return true
		}
		m := msg.NewImageAutoWithoutInjectionError(resource, "Pod", resource.Metadata.FullName.Name.String())
		c.Report(gvk.Pod, m)
		return true
	})
	c.ForEach(gvk.Deployment, func(resource *resource.Instance) bool {
		d := resource.Message.(*appsv1.DeploymentSpec)
		if !hasAutoImage(&d.Template.Spec) {
			return true
		}
		nsLabels := getNamespaceLabels(c, resource.Metadata.FullName.Namespace.String())
		if !matchesWebhooks(nsLabels, d.Template.Labels, istioWebhooks) {
			m := msg.NewImageAutoWithoutInjectionWarning(resource, "Deployment", resource.Metadata.FullName.Name.String())
			c.Report(gvk.Deployment, m)
		}
		return true
	})
}

func hasAutoImage(spec *v1.PodSpec) bool {
	for _, c := range spec.Containers {
		if c.Name == istioProxyContainerName && c.Image == manualInjectionImage {
			return true
		}
	}
	return false
}

func getNamespaceLabels(c analysis.Context, nsName string) map[string]string {
	if nsName == "" {
		nsName = "default"
	}
	ns := c.Find(gvk.Namespace, resource.NewFullName("", resource.LocalName(nsName)))
	if ns == nil {
		return nil
	}
	return ns.Metadata.Labels
}

func matchesWebhooks(nsLabels, podLabels map[string]string, istioWebhooks []admitv1.MutatingWebhook) bool {
	for _, w := range istioWebhooks {
		if selectorMatches(w.NamespaceSelector, nsLabels) && selectorMatches(w.ObjectSelector, podLabels) {
			return true
		}
	}
	return false
}

func selectorMatches(selector *metav1.LabelSelector, labels klabels.Set) bool {
	// From webhook spec: "Default to the empty LabelSelector, which matchesWebhooks everything."
	if selector == nil {
		return true
	}
	s, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return false
	}
	return s.Matches(labels)
}
