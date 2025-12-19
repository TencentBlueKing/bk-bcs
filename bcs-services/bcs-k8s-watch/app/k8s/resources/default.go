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

// Package resources xxx
package resources

import (
	"fmt"
	"net/url"

	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/options"
)

const (
	// CoreV1GroupVersion groupversion
	CoreV1GroupVersion = "v1"
	// AppsV1GroupVersion groupversion
	AppsV1GroupVersion = "apps/v1"
	// AppsV1Beta1GroupVersion groupversion
	AppsV1Beta1GroupVersion = "apps/v1beta1"
	// AppsV1Beta2GroupVersion groupversion
	AppsV1Beta2GroupVersion = "apps/v1beta2"
	// ExtensionsV1Beta1GroupVersion groupversion
	ExtensionsV1Beta1GroupVersion = "extensions/v1beta1"
	// AutoScalingV1GroupVersion groupversion
	AutoScalingV1GroupVersion = "autoscaling/v1"
	// AutoScalingV2Beta1GroupVersion groupversion
	AutoScalingV2Beta1GroupVersion = "autoscaling/v2beta1"
	// AutoScalingV2Beta2GroupVersion groupversion
	AutoScalingV2Beta2GroupVersion = "autoscaling/v2beta2"
	// StorageV1GroupVersion groupversion
	StorageV1GroupVersion = "storage.k8s.io/v1"

	// BatchV1GroupVersion groupversion
	BatchV1GroupVersion = "batch/v1"
	// BatchV1Beta1GroupVersion groupversion
	BatchV1Beta1GroupVersion = "batch/v1beta1"
	// RbacV1GroupVersion groupversion
	RbacV1GroupVersion = "rbac.authorization.k8s.io/v1"
	// RbacV1Beta1GroupVersion groupversion
	RbacV1Beta1GroupVersion = "rbac.authorization.k8s.io/v1beta1"
	// AdmissionRegistrationV1Beta1GroupVersion groupversion
	AdmissionRegistrationV1Beta1GroupVersion = "admissionregistration.k8s.io/v1beta1"
	// ApiExtensionsV1Beta1GroupVersion groupversion
	ApiExtensionsV1Beta1GroupVersion = "apiextensions.k8s.io/v1beta1"

	// BkbcsGroupName groupversion
	BkbcsGroupName = "bkbcs.tencent.com"
	// MesosV2GroupVersion groupversion
	MesosV2GroupVersion = "bkbcs.tencent.com/v2"
	// WebhookV1GroupVersion groupversion
	WebhookV1GroupVersion = "bkbcs.tencent.com/v1"

	// TkexV1alpha1GroupName groupversion
	TkexV1alpha1GroupName = "tkex.tencent.com"
	// TkexV1alpha1GroupVersion groupversion
	TkexV1alpha1GroupVersion = "tkex.tencent.com/v1alpha1"
	// TkexGameDeploymentName groupversion
	TkexGameDeploymentName = "gamedeployments.tkex.tencent.com"
	// TkexGameStatefulSetName groupversion
	TkexGameStatefulSetName = "gamestatefulsets.tkex.tencent.com"
	// TkexGPAName groupversion
	TkexGPAName = "generalpodautoscalers.autoscaling.tkex.tencent.com"

	// KubefedTypesV1Beta1GroupVersion groupversion
	KubefedTypesV1Beta1GroupVersion = "types.kubefed.io/v1beta1"
	// KubefedCoreV1Alpha1GroupVersion groupversion
	KubefedCoreV1Alpha1GroupVersion = "core.kubefed.io/v1alpha1"
	// KubefedCoreV1Beta1GroupVersion groupversion
	KubefedCoreV1Beta1GroupVersion = "core.kubefed.io/v1beta1"
	// KubefedMultiClusterDnsV1Alpha1GroupVersion groupversion
	KubefedMultiClusterDnsV1Alpha1GroupVersion = "multiclusterdns.kubefed.io/v1alpha1"
	// KubefedSchedulingV1Alpha1GroupVersion groupversion
	KubefedSchedulingV1Alpha1GroupVersion = "scheduling.kubefed.io/v1alpha1"
)

// K8sWatcherConfigList resource list to watch
// map[Kind]ResourceObjType
var K8sWatcherConfigList map[string]ResourceObjType

// K8sClientList map[string]*dynamic.Interface,
// CrdClientList map[string]*dynamic.Interface
var K8sClientList, CrdClientList map[string]*dynamic.DynamicClient // nolint

// ResourceObjType used for build target watchers.
type ResourceObjType struct {
	ResourceName string
	ObjType      runtime.Object
	Client       *dynamic.DynamicClient
	Namespaced   bool
	GroupVersion string
}

// InitResourceList init resource list to watch
func InitResourceList(k8sConfig *options.K8sConfig, filterConfig *options.FilterConfig,
	watchResource *options.WatchResource) error {
	restConfig, err := GetRestConfig(k8sConfig)
	if err != nil {
		return fmt.Errorf("error creating rest config: %s", err.Error())
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("error creating dynamic client for k8s resource: %s", err.Error())
	}

	// init k8s client list
	K8sClientList, err = initK8sClientList(dynamicClient, filterConfig)
	if err != nil {
		return err
	}

	// init crd client list
	CrdClientList, err = initCrdClientList(dynamicClient, filterConfig)
	if err != nil {
		return err
	}

	// init k8s watcher config list
	resFilter := NewResourceFilter(filterConfig)
	apiResourceLists, err := getApiResourceLists(restConfig)
	if err != nil {
		glog.Warnf("error getting apiResourceLists: %s, get from config file", err.Error())
		apiResourceLists = filterConfig.APIResourceLists
	}
	K8sWatcherConfigList, err = initK8sWatcherConfigList(apiResourceLists, resFilter, watchResource.Namespace != "")
	if err != nil {
		return err
	}

	return nil
}

func initK8sClientList(
	dynamicClient *dynamic.DynamicClient, filterConfig *options.FilterConfig) (map[string]*dynamic.DynamicClient, error) { //nolint
	k8sClientList := map[string]*dynamic.DynamicClient{}
	for _, gv := range filterConfig.K8sGroupVersionWhiteList {
		glog.Infof("add client into k8sClientList for %s", gv)
		k8sClientList[gv] = dynamicClient
	}
	return k8sClientList, nil
}

func initCrdClientList(
	dynamicClient *dynamic.DynamicClient, filterConfig *options.FilterConfig) (map[string]*dynamic.DynamicClient, error) { //nolint
	crdClientList := map[string]*dynamic.DynamicClient{}
	for _, gv := range filterConfig.CrdGroupVersionWhiteList {
		glog.Infof("add client into crdClientList for %s", gv)
		crdClientList[gv] = dynamicClient
	}
	return crdClientList, nil
}

// initK8sWatcherConfigList init k8s resource
func initK8sWatcherConfigList(apiResourceLists []options.ApiResourceList, filter *ResourceFilter,
	onlyWatchNamespacedResource bool) (map[string]ResourceObjType, error) { // nolint

	k8sWatcherConfigList := make(map[string]ResourceObjType)
	for _, apiResourceList := range apiResourceLists {
		if kubeClient, ok := K8sClientList[apiResourceList.GroupVersion]; ok {
			// resourceFiltered, resourceFilterOK := filter[apiResourceList.GroupVersion]
			for _, apiResource := range apiResourceList.APIResources {
				if filter.IsBanned(apiResourceList.GroupVersion, apiResource) {
					continue
				}
				var obj runtime.Unstructured
				_, ok := k8sWatcherConfigList[apiResource.Kind]
				if ok && apiResourceList.GroupVersion == ExtensionsV1Beta1GroupVersion {
					// 如果 deployment, daemonset 在apps和extensions下面都有，则只watch apps下面的资源
					continue
				}
				// 如果指定了namespace则不监听非namespace的资源
				if onlyWatchNamespacedResource && !apiResource.Namespaced {
					continue
				}
				k8sWatcherConfigList[apiResource.Kind] = ResourceObjType{
					ResourceName: apiResource.Name,
					ObjType:      obj,
					Client:       kubeClient,
					Namespaced:   apiResource.Namespaced,
					GroupVersion: apiResourceList.GroupVersion,
				}
			}
		}
	}

	return k8sWatcherConfigList, nil
}

func getApiResourceLists(restConfig *rest.Config) ([]options.ApiResourceList, error) {
	// discover apiResourceLists
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating discovery client: %s", err.Error())
	}
	k8sApiResourceLists, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, err
	}

	// convert k8s apiResourceList to local apiResourceList
	retApiResourceLists := []options.ApiResourceList{}
	for _, serverApiResourceList := range k8sApiResourceLists {
		retApiResourceList := options.ApiResourceList{
			GroupVersion: serverApiResourceList.GroupVersion,
			APIResources: []options.APIResource{},
		}
		for _, serverApiResource := range serverApiResourceList.APIResources {
			retApiResourceList.APIResources = append(retApiResourceList.APIResources, options.APIResource{
				Name:       serverApiResource.Name,
				Namespaced: serverApiResource.Namespaced,
				Kind:       serverApiResource.Kind,
			})
		}
		retApiResourceLists = append(retApiResourceLists, retApiResourceList)
	}
	return retApiResourceLists, nil
}

// GetRestConfig generate rest config
func GetRestConfig(k8sConfig *options.K8sConfig) (*rest.Config, error) {
	var config *rest.Config
	var err error

	// build k8s client config.
	if k8sConfig.Kubeconfig != "" {
		glog.Info("k8sConfig.Kubeconfig is set: %s", k8sConfig.Kubeconfig)
		// use the current context in kubeconfig
		return clientcmd.BuildConfigFromFlags("", k8sConfig.Kubeconfig)
	}
	if k8sConfig.Master != "" {
		glog.Info("k8sConfig.Master is set: %s", k8sConfig.Master)
		// NOCC:vetshadow/shadow(设计如此:这里err可以被覆盖)
		u, err := url.Parse(k8sConfig.Master) // nolint
		if err != nil {
			return nil, err
		}

		var tlsConfig rest.TLSClientConfig
		if u.Scheme == "https" {
			if k8sConfig.TLS.CAFile == "" || k8sConfig.TLS.CertFile == "" || k8sConfig.TLS.KeyFile == "" {
				return nil, fmt.Errorf("use https, kube-ca-file, kube-cert-file, kube-key-file required")
			}

			tlsConfig = rest.TLSClientConfig{
				CAFile:   k8sConfig.TLS.CAFile,
				CertFile: k8sConfig.TLS.CertFile,
				KeyFile:  k8sConfig.TLS.KeyFile,
			}
		}
		return &rest.Config{
			Host:            k8sConfig.Master,
			QPS:             1e6,
			Burst:           1e6,
			TLSClientConfig: tlsConfig,
		}, nil
	}

	glog.Info("k8sConfig.Master and k8sConfig.kubeconfig is not be set, use in cluster mode")

	config, err = rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}
