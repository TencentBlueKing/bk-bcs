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

	glog "bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/options"
	mesosv2 "bk-bcs/bcs-mesos/pkg/apis/bkbcs/v2"
	"bk-bcs/bcs-mesos/pkg/client/internalclientset"
	wbbcsv2 "bk-bcs/bcs-services/bcs-webhook-server/pkg/apis/bk-bcs/v2"
	webhookClientSet "bk-bcs/bcs-services/bcs-webhook-server/pkg/client/clientset/versioned"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	BkbcsGroupName = "bkbcs.tencent.com"
)

// resource list to watch
var WatcherConfigList, BkbcsWatcherConfigLister map[string]ResourceObjType

// ResourceObjType used for build target watchers.
type ResourceObjType struct {
	ResourceName string
	ObjType      runtime.Object
	Client       *rest.Interface
	Namespaced   bool
}

// InitResourceList init resource list to watch
func InitResourceList(k8sConfig *options.K8sConfig) error {
	restConfig, err := GetRestConfig(k8sConfig)
	if err != nil {
		return err
	}

	WatcherConfigList, err = initNormalWatcherConfigList(restConfig)
	if err != nil {
		return err
	}

	BkbcsWatcherConfigLister = make(map[string]ResourceObjType)

	webhookWatcherConfigList, err := initWebhookWatcherConfigList(restConfig)
	if err != nil {
		return err
	}
	for resourceKind, resourceObjType := range webhookWatcherConfigList {
		BkbcsWatcherConfigLister[resourceKind] = resourceObjType
	}

	mesosWatcherConfigList, err := initMesosWatcherConfigList(restConfig)
	if err != nil {
		return err
	}
	for resourceKind, resourceObjType := range mesosWatcherConfigList {
		BkbcsWatcherConfigLister[resourceKind] = resourceObjType
	}

	return nil
}

// initNormalWatcherConfigList init k8s resource
func initNormalWatcherConfigList(restConfig *rest.Config) (map[string]ResourceObjType, error) {
	// create k8s clientset.
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	coreClient := clientSet.CoreV1().RESTClient()
	extensionsV1Beta1Client := clientSet.ExtensionsV1beta1().RESTClient()
	batchV1Client := clientSet.BatchV1().RESTClient()
	appsV1Client := clientSet.AppsV1().RESTClient()

	// build k8s watcher configs.
	normalWatcherConfigList := map[string]ResourceObjType{
		"Node": {
			"nodes",
			&v1.Node{},
			&coreClient,
			false,
		},
		"Pod": {
			"pods",
			&v1.Pod{},
			&coreClient,
			true,
		},
		"ReplicationController": {
			"replicationcontrollers",
			&v1.ReplicationController{},
			&coreClient,
			true,
		},
		"Service": {
			"services",
			&v1.Service{},
			&coreClient,
			true,
		},
		"EndPoints": {
			"endpoints",
			&v1.Endpoints{},
			&coreClient,
			true,
		},
		"ConfigMap": {
			"configmaps",
			&v1.ConfigMap{},
			&coreClient,
			true,
		},
		"Secret": {
			"secrets",
			&v1.Secret{},
			&coreClient,
			true,
		},
		"Namespace": {
			"namespaces",
			&v1.Namespace{},
			&coreClient,
			false,
		},
		"Event": {
			"events",
			&v1.Event{},
			&coreClient,
			true,
		},
		"Deployment": {
			"deployments",
			&appsv1.Deployment{},
			&appsV1Client,
			true,
		},
		"Ingress": {
			"ingresses",
			&v1beta1.Ingress{},
			&extensionsV1Beta1Client,
			true,
		},
		"ReplicaSet": {
			"replicasets",
			&appsv1.ReplicaSet{},
			&appsV1Client,
			true,
		},
		"DaemonSet": {
			"daemonsets",
			&appsv1.DaemonSet{},
			&appsV1Client,
			true,
		},
		"Job": {
			"jobs",
			&batchv1.Job{},
			&batchV1Client,
			true,
		},
		"StatefulSet": {
			"statefulsets",
			&appsv1.StatefulSet{},
			&appsV1Client,
			true,
		},
	}

	return normalWatcherConfigList, nil
}

// initWebhookWatcherConfigList init bcs-webhook-server crd resource
func initWebhookWatcherConfigList(restConfig *rest.Config) (map[string]ResourceObjType, error) {
	// create webhook crd clientset
	whClientSet, err := webhookClientSet.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	webhookClient := whClientSet.BkbcsV2().RESTClient()

	webhookWatcherConfigLister := map[string]ResourceObjType{
		"BcsLogConfig": {
			"bcslogconfigs",
			&wbbcsv2.BcsLogConfig{},
			&webhookClient,
			true,
		},
		"BcsDbPrivConfig": {
			"bcsdbprivconfigs",
			&wbbcsv2.BcsDbPrivConfig{},
			&webhookClient,
			true,
		},
	}

	return webhookWatcherConfigLister, nil
}

// initMesosWatcherConfigList init mesos crd resource
func initMesosWatcherConfigList(restConfig *rest.Config) (map[string]ResourceObjType, error) {
	mesosClientset, err := internalclientset.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	mesosClient := mesosClientset.BkbcsV2().RESTClient()

	mesosWatcherConfigLister := map[string]ResourceObjType{
		"AdmissionWebhookConfiguration": {
			"admissionwebhookconfigurations",
			&mesosv2.AdmissionWebhookConfiguration{},
			&mesosClient,
			true,
		},
		"Agent": {
			"agents",
			&mesosv2.Agent{},
			&mesosClient,
			true,
		},
		"AgentSchedInfo": {
			"agentschedinfos",
			&mesosv2.AgentSchedInfo{},
			&mesosClient,
			true,
		},
		"Application": {
			"applications",
			&mesosv2.Application{},
			&mesosClient,
			true,
		},
		"BcsClusterAgentSetting": {
			"bcsclusteragentsettings",
			&mesosv2.BcsClusterAgentSetting{},
			&mesosClient,
			true,
		},
		"BcsCommandInfo": {
			"bcscommandinfos",
			&mesosv2.BcsCommandInfo{},
			&mesosClient,
			true,
		},
		"BcsConfigMap": {
			"bcsconfigmaps",
			&mesosv2.BcsConfigMap{},
			&mesosClient,
			true,
		},
		"BcsEndpoint": {
			"bcsendpoints",
			&mesosv2.BcsEndpoint{},
			&mesosClient,
			true,
		},
		"BcsSecret": {
			"bcssecrets",
			&mesosv2.BcsSecret{},
			&mesosClient,
			true,
		},
		"BcsService": {
			"bcsservices",
			&mesosv2.BcsService{},
			&mesosClient,
			true,
		},
		"Crd": {
			"crds",
			&mesosv2.Crd{},
			&mesosClient,
			true,
		},
		"Crr": {
			"crrs",
			&mesosv2.Crr{},
			&mesosClient,
			true,
		},
		"Deployment": {
			"deployments",
			&mesosv2.Deployment{},
			&mesosClient,
			true,
		},
		"Framework": {
			"frameworks",
			&mesosv2.Framework{},
			&mesosClient,
			true,
		},
		"Task": {
			"tasks",
			&mesosv2.Task{},
			&mesosClient,
			true,
		},
		"TaskGroup": {
			"taskgroups",
			&mesosv2.TaskGroup{},
			&mesosClient,
			true,
		},
		"Version": {
			"versions",
			&mesosv2.Version{},
			&mesosClient,
			true,
		},
	}

	return mesosWatcherConfigLister, nil
}

// GetRestConfig generate rest config
func GetRestConfig(k8sConfig *options.K8sConfig) (*rest.Config, error) {
	var config *rest.Config
	var err error

	// build k8s client config.
	if k8sConfig.Master == "" {
		glog.Info("k8sConfig.Master is not be set, use in cluster mode")

		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
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
		config = &rest.Config{
			Host:            k8sConfig.Master,
			QPS:             1e6,
			Burst:           1e6,
			TLSClientConfig: tlsConfig,
		}
	}

	return config, nil
}
