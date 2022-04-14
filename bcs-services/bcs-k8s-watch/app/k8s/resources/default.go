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

package resources

import (
	"fmt"
	"net/url"

	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	tkexGDClientSet "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/client/clientset/versioned"
	tkexGSClientSet "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator/pkg/clientset/internalclientset"
	webhookClientSet "github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs/client/clientset/versioned"
	"github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/internalclientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/options"
	kubefedClientSet "github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/pkg/kubefed/client/clientset/versioned"
	extensionsClientSet "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	CoreV1GroupVersion             = "v1"
	AppsV1GroupVersion             = "apps/v1"
	AppsV1Beta1GroupVersion        = "apps/v1beta1"
	AppsV1Beta2GroupVersion        = "apps/v1beta2"
	ExtensionsV1Beta1GroupVersion  = "extensions/v1beta1"
	AutoScalingV1GroupVersion      = "autoscaling/v1"
	AutoScalingV2Beta1GroupVersion = "autoscaling/v2beta1"
	AutoScalingV2Beta2GroupVersion = "autoscaling/v2beta2"
	StorageV1GroupVersion          = "storage.k8s.io/v1"

	BatchV1GroupVersion                      = "batch/v1"
	BatchV1Beta1GroupVersion                 = "batch/v1beta1"
	RbacV1GroupVersion                       = "rbac.authorization.k8s.io/v1"
	RbacV1Beta1GroupVersion                  = "rbac.authorization.k8s.io/v1beta1"
	AdmissionRegistrationV1Beta1GroupVersion = "admissionregistration.k8s.io/v1beta1"
	ApiExtensionsV1Beta1GroupVersion         = "apiextensions.k8s.io/v1beta1"

	BkbcsGroupName        = "bkbcs.tencent.com"
	MesosV2GroupVersion   = "bkbcs.tencent.com/v2"
	WebhookV1GroupVersion = "bkbcs.tencent.com/v1"

	TkexV1alpha1GroupName    = "tkex.tencent.com"
	TkexV1alpha1GroupVersion = "tkex.tencent.com/v1alpha1"
	TkexGameDeploymentName   = "gamedeployments.tkex.tencent.com"
	TkexGameStatefulSetName  = "gamestatefulsets.tkex.tencent.com"

	KubefedTypesV1Beta1GroupVersion            = "types.kubefed.io/v1beta1"
	KubefedCoreV1Alpha1GroupVersion            = "core.kubefed.io/v1alpha1"
	KubefedCoreV1Beta1GroupVersion             = "core.kubefed.io/v1beta1"
	KubefedMultiClusterDnsV1Alpha1GroupVersion = "multiclusterdns.kubefed.io/v1alpha1"
	KubefedSchedulingV1Alpha1GroupVersion      = "scheduling.kubefed.io/v1alpha1"
)

// resource list to watch
var WatcherConfigList, BkbcsWatcherConfigList map[string]ResourceObjType
var K8sClientList, CrdClientList map[string]rest.Interface

// ResourceObjType used for build target watchers.
type ResourceObjType struct {
	ResourceName string
	ObjType      runtime.Object
	Client       *rest.Interface
	Namespaced   bool
}

// InitResourceList init resource list to watch
func InitResourceList(k8sConfig *options.K8sConfig, filterConfig *options.FilterConfig, watchResource *options.WatchResource) error {
	restConfig, err := GetRestConfig(k8sConfig)
	if err != nil {
		return fmt.Errorf("error creating rest config: %s", err.Error())
	}

	filter := make(map[string]map[string]struct{})
	if filterConfig != nil {
		for _, gv := range filterConfig.APIResourceException {
			filter[gv.GroupVersion] = make(map[string]struct{})
			for _, resource := range gv.ResourceKinds {
				filter[gv.GroupVersion][resource] = struct{}{}
			}
		}
	}

	// 初始化待watch的k8s资源
	WatcherConfigList, err = initK8sWatcherConfigList(restConfig, filter, watchResource.Namespace != "")
	if err != nil {
		return err
	}

	// 初始化待watch的crd资源
	CrdClientList = make(map[string]rest.Interface)

	// 初始化联邦集群crd各个apiVersion的RestClient
	kubefedClientList, err := initKubefedClient(restConfig)
	if err != nil {
		return err
	}
	for groupVersion, client := range kubefedClientList {
		CrdClientList[groupVersion] = client
	}

	// 初始化mesos crd各个apiVersion的RestClient
	mesosClientList, err := initMesosClient(restConfig)
	if err != nil {
		return err
	}
	for groupVersion, client := range mesosClientList {
		CrdClientList[groupVersion] = client
	}

	// 初始化bcs-webhook-server crd各个apiVersion的RestClient
	webhookClientList, err := initWebhookClient(restConfig)
	if err != nil {
		return err
	}
	for groupVersion, client := range webhookClientList {
		CrdClientList[groupVersion] = client
	}

	// 初始化tkex crd各个apiVersion的RestClient
	tkeClientList, err := initTkexClient(restConfig)
	if err != nil {
		return err
	}
	for groupVersion, client := range tkeClientList {
		CrdClientList[groupVersion] = client
	}

	return nil
}

// initKubefedClient init kubefed resources restclient
func initKubefedClient(restConfig *rest.Config) (map[string]rest.Interface, error) {
	// create kubefed clientset
	clientset, err := kubefedClientSet.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	kubefedCoreV1Alpha1Client := clientset.CoreV1alpha1().RESTClient()
	kubefedCoreV1Beta1Client := clientset.CoreV1beta1().RESTClient()
	kubefedMultiDnsV1Alpha1Client := clientset.MulticlusterdnsV1alpha1().RESTClient()
	kubefedSchedulingV1Alpha1Client := clientset.SchedulingV1alpha1().RESTClient()
	kubefedTypesV1Beta1Client := clientset.TypesV1beta1().RESTClient()

	kubefedClientList := map[string]rest.Interface{
		KubefedCoreV1Alpha1GroupVersion:            kubefedCoreV1Alpha1Client,
		KubefedCoreV1Beta1GroupVersion:             kubefedCoreV1Beta1Client,
		KubefedMultiClusterDnsV1Alpha1GroupVersion: kubefedMultiDnsV1Alpha1Client,
		KubefedSchedulingV1Alpha1GroupVersion:      kubefedSchedulingV1Alpha1Client,
		KubefedTypesV1Beta1GroupVersion:            kubefedTypesV1Beta1Client,
	}

	return kubefedClientList, nil
}

// initMesosClient init mesos resources restclient
func initMesosClient(restConfig *rest.Config) (map[string]rest.Interface, error) {
	mesosClientset, err := internalclientset.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	mesosClient := mesosClientset.BkbcsV2().RESTClient()

	mesosClientList := map[string]rest.Interface{
		MesosV2GroupVersion: mesosClient,
	}

	return mesosClientList, nil
}

// initWebhookClient init bcs-webhook-server resources restclient
func initWebhookClient(restConfig *rest.Config) (map[string]rest.Interface, error) {
	// create webhook clientset
	clientset, err := webhookClientSet.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	webhookClientList := map[string]rest.Interface{
		WebhookV1GroupVersion: clientset.BkbcsV1().RESTClient(),
	}

	return webhookClientList, nil
}

func initTkexClient(restConfig *rest.Config) (map[string]rest.Interface, error) {
	tkexGDClient, err := tkexGDClientSet.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	tkexGSClient, err := tkexGSClientSet.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	tkexClientList := map[string]rest.Interface{
		TkexGameDeploymentName:  tkexGDClient.TkexV1alpha1().RESTClient(),
		TkexGameStatefulSetName: tkexGSClient.TkexV1alpha1().RESTClient(),
	}
	return tkexClientList, nil
}

// initK8sWatcherConfigList init k8s resource
func initK8sWatcherConfigList(restConfig *rest.Config, filter map[string]map[string]struct{}, onlyWatchNamespacedResource bool) (map[string]ResourceObjType, error) {
	// create k8s clientset.
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating kube clientset: %s", err.Error())
	}

	crdClientSet, err := extensionsClientSet.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating crd clientset: %s", err.Error())
	}

	K8sClientList = map[string]rest.Interface{
		CoreV1GroupVersion:                       clientSet.CoreV1().RESTClient(),
		AppsV1GroupVersion:                       clientSet.AppsV1().RESTClient(),
		AppsV1Beta1GroupVersion:                  clientSet.AppsV1beta1().RESTClient(),
		AppsV1Beta2GroupVersion:                  clientSet.AppsV1beta2().RESTClient(),
		ExtensionsV1Beta1GroupVersion:            clientSet.ExtensionsV1beta1().RESTClient(),
		AutoScalingV1GroupVersion:                clientSet.AutoscalingV1().RESTClient(),
		AutoScalingV2Beta1GroupVersion:           clientSet.AutoscalingV2beta1().RESTClient(),
		AutoScalingV2Beta2GroupVersion:           clientSet.AutoscalingV2beta2().RESTClient(),
		StorageV1GroupVersion:                    clientSet.StorageV1().RESTClient(),
		BatchV1GroupVersion:                      clientSet.BatchV1().RESTClient(),
		BatchV1Beta1GroupVersion:                 clientSet.BatchV1beta1().RESTClient(),
		RbacV1GroupVersion:                       clientSet.RbacV1().RESTClient(),
		RbacV1Beta1GroupVersion:                  clientSet.RbacV1beta1().RESTClient(),
		AdmissionRegistrationV1Beta1GroupVersion: clientSet.AdmissionregistrationV1beta1().RESTClient(),
		ApiExtensionsV1Beta1GroupVersion:         crdClientSet.ApiextensionsV1beta1().RESTClient(),
	}

	k8sWatcherConfigList := make(map[string]ResourceObjType)

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating discovery client: %s", err.Error())
	}
	apiResourceLists, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		glog.Warnf("error getting apiResourceLists: %s", err.Error())
		return nil, err
	}

	for _, apiResourceList := range apiResourceLists {
		if kubeClient, ok := K8sClientList[apiResourceList.GroupVersion]; ok {
			resourceFiltered, resourceFilterOK := filter[apiResourceList.GroupVersion]
			for _, apiResource := range apiResourceList.APIResources {
				if apiResource.Kind != "Namespace" {
					if resourceFilterOK && len(resourceFiltered) == 0 {
						glog.Warnf("filter has banned all resource in groupversion %s", apiResourceList.GroupVersion)
						continue
					}
					if _, filtered := resourceFiltered[apiResource.Kind]; filtered && resourceFilterOK {
						glog.Warnf("filter has banned resource kind %s in groupversion %s", apiResource.Kind, apiResourceList.GroupVersion)
						continue
					}
				}
				var obj runtime.Object
				_, ok := k8sWatcherConfigList[apiResource.Kind]
				if ok && apiResourceList.GroupVersion == ExtensionsV1Beta1GroupVersion {
					// 如果 deployment, daemonset 在apps和extensions下面都有，则只watch apps下面的资源
					continue
				}

				if apiResource.Kind == "ComponentStatus" || apiResource.Kind == "Binding" || apiResource.Kind == "ReplicationControllerDummy" {
					// 这几种类型的资源无法watch，跳过
					continue
				}

				if apiResourceList.GroupVersion == StorageV1GroupVersion && apiResource.Kind != "StorageClass" {
					// 1.12版本的 VolumeAttachment在v1beta1下，但1.14版本放到了v1下，为了避免list报错，暂时只同步StorageClass
					continue
				}
				//如果指定了namespace则不监听非namespace的资源
				if onlyWatchNamespacedResource && !apiResource.Namespaced {
					continue
				}
				k8sWatcherConfigList[apiResource.Kind] = ResourceObjType{
					ResourceName: apiResource.Name,
					ObjType:      obj,
					Client:       &kubeClient,
					Namespaced:   apiResource.Namespaced,
				}
			}
		}
	}

	return k8sWatcherConfigList, nil
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

		u, err := url.Parse(k8sConfig.Master)
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
