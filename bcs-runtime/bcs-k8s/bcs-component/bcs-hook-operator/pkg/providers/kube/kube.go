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

package kube

import (
	metricutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/util/metric"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	cached "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

const (
	// ProviderType indicates the provider is kubernetes.
	ProviderType = "Kubernetes"
	// FunctionTypeGet is to get field values of the resource object.
	FunctionTypeGet = "get"
	// FunctionTypePatch is to patch the field of the resource object.
	FunctionTypePatch = "patch"
)

// Provider contains all the required components to run k8s operations
type Provider struct {
	dynamicClient dynamic.Interface
	cachedClient  discovery.CachedDiscoveryInterface
}

// Type incidates provider is a kubernetes provider
func (p *Provider) Type() string {
	return ProviderType
}

// Run kubernetes operations for the metric
func (p *Provider) Run(run *v1alpha1.HookRun, metric v1alpha1.Metric) v1alpha1.Measurement {
	startTime := metav1.Now()
	newMeasurement := v1alpha1.Measurement{
		StartedAt: &startTime,
	}

	gvr, err := p.getGroupVersionResource(run.Spec.Args)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}
	dr, name, err := p.getDynamicResource(gvr, run.Spec.Args)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	err = p.handleFunction(dr, name, metric.Provider.Kubernetes)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	newMeasurement.Phase = v1alpha1.HookPhaseSuccessful
	finishedTime := metav1.Now()
	newMeasurement.FinishedAt = &finishedTime

	return newMeasurement

}

// Resume should not be used the kubernetes provider since all the work should occur in the Run method
func (p *Provider) Resume(
	run *v1alpha1.HookRun,
	metric v1alpha1.Metric,
	measurement v1alpha1.Measurement) v1alpha1.Measurement {
	klog.Warningf("HookRun: %s/%s, metric: %s. Kubernetes provider should not execute the Resume method",
		run.Namespace, run.Name, metric.Name)
	return measurement
}

// Terminate should not be used the kubernetes provider since all the work should occur in the Run method
func (p *Provider) Terminate(
	run *v1alpha1.HookRun,
	metric v1alpha1.Metric,
	measurement v1alpha1.Measurement) v1alpha1.Measurement {
	klog.Warningf("HookRun: %s/%s, metric: %s. Kubernetes provider should not execute the Resume method",
		run.Namespace, run.Name, metric.Name)
	return measurement
}

// GarbageCollect is a no-op for the kubernetes provider
func (p *Provider) GarbageCollect(run *v1alpha1.HookRun, metric v1alpha1.Metric, limit int) error {
	return nil
}

// NewKubeProvider creates a new Kube client
func NewKubeProvider(dynamicClient dynamic.Interface, cachedClient discovery.CachedDiscoveryInterface) *Provider {
	return &Provider{
		dynamicClient: dynamicClient,
		cachedClient:  cachedClient,
	}
}

// NewKubeClient generates a dynamic client and a discovery client
func NewKubeClient(metric v1alpha1.Metric) (dynamic.Interface, discovery.CachedDiscoveryInterface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		return nil, nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	cachedClient := cached.NewMemCacheClient(discoveryClient)
	if !cachedClient.Fresh() {
		cachedClient.Invalidate()
	}
	return dynamicClient, cachedClient, nil
}
