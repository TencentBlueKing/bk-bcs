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

package k8s

import (
	"fmt"
	"net/url"

	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/pkg/api/v1"
	//"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	//appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	//"k8s.io/api/apps/v1beta2"

	"k8s.io/client-go/rest"

	glog "bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/bcs"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/options"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output"
)

// =================== interface & struct ===================

type Cluster struct {
	watchers                map[string]WatcherInterface
	synchronizer            *Synchronizer
	exportServiceController *ExportServiceController
}

type ResourceObjType struct {
	resourceName string
	objType      runtime.Object
	client       *rest.Interface
}

// =================== New & Run ===================

func NewCluster(writer *output.Writer, k8sConfig *options.K8sConfig, clusterID string, storageService *bcs.StorageService) (Cluster, error) {
	//"k8s.io/client-go/tools/clientcmd"
	//clientcmd.BuildConfigFromFlags()

	// 1. new cluster
	cluster := Cluster{
		watchers: make(map[string]WatcherInterface),
	}

	// 3. register watchers
	cluster.InitWatchers(writer, k8sConfig, clusterID)

	cluster.synchronizer = &Synchronizer{
		watchers:       cluster.watchers,
		ClusterID:      clusterID,
		StorageService: storageService,
	}

	return cluster, nil
}

func newClientSet(k8sConfig *options.K8sConfig) *kubernetes.Clientset {

	var config *rest.Config
	var err error

	// 2.1 build client.Config
	if k8sConfig.Master == "" {
		glog.Info("k8sConfig.Master is not be set, use in cluster mode")
		// 2.1 creates the in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	} else {
		glog.Info("k8sConfig.Master is set: %s", k8sConfig.Master)
		// TODO: modify here, need to use master url and cert file to make clientset

		u, err := url.Parse(k8sConfig.Master)
		if err != nil {
			panic(fmt.Errorf("invalid master url:%v", err))
		}

		var tlsConfig rest.TLSClientConfig
		if u.Scheme == "https" {
			if k8sConfig.TLS.CAFile == "" || k8sConfig.TLS.CertFile == "" || k8sConfig.TLS.KeyFile == "" {
				panic(fmt.Errorf("use https, kube-ca-file, kube-cert-file, kube-key-file required"))
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

	// 2.2 creates the clientSet
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientSet
}

func (cluster *Cluster) InitWatchers(writer *output.Writer, k8sconfig *options.K8sConfig, clusterID string) {
	// 2. new client
	clientSet := newClientSet(k8sconfig)

	// 2.3 got client
	coreClient := clientSet.CoreV1().RESTClient()
	extensionsV1Beta1Client := clientSet.ExtensionsV1beta1().RESTClient()
	batchV1Client := clientSet.BatchV1().RESTClient()
	appsV1Beta1Client := clientSet.AppsV1beta1().RESTClient()

	//appsV1beta2Client := clientSet.AppsV1beta2().RESTClient()

	// 2.4 create export service controller
	esController := NewExportServiceController()
	var watcherList = map[string]*Watcher{}

	watcherConfigList := map[string]ResourceObjType{
		"Node": ResourceObjType{
			"nodes",
			&v1.Node{},
			&coreClient,
		},
		"Pod": {
			"pods",
			&v1.Pod{},
			&coreClient,
		},
		"ReplicationController": {
			"replicationcontrollers",
			&v1.ReplicationController{},
			&coreClient,
		},
		"Service": {
			"services",
			&v1.Service{},
			&coreClient,
		},
		"EndPoints": {
			"endpoints",
			&v1.Endpoints{},
			&coreClient,
		},
		"ConfigMap": {
			"configmaps",
			&v1.ConfigMap{},
			&coreClient,
		},
		"Secret": {
			"secrets",
			&v1.Secret{},
			&coreClient,
		},
		"Namespace": {
			"namespaces",
			&v1.Namespace{},
			&coreClient,
		},
		"Event": {
			"events",
			&v1.Event{},
			&coreClient,
		},
		"Deployment": {
			"deployments",
			&v1beta1.Deployment{},
			&extensionsV1Beta1Client,
		},
		"Ingress": {
			"ingresses",
			&v1beta1.Ingress{},
			&extensionsV1Beta1Client,
		},
		"ReplicaSet": {
			"replicasets",
			&v1beta1.ReplicaSet{},
			&extensionsV1Beta1Client,
		},
		"DaemonSet": {
			"daemonsets",
			&v1beta1.DaemonSet{},
			//&appsv1beta2.DaemonSet{},
			&extensionsV1Beta1Client,
		},
		"Job": {
			"jobs",
			&batchv1.Job{},
			&batchV1Client,
		},
		"StatefulSet": {
			"statefulsets",
			&appsv1beta1.StatefulSet{},
			&appsV1Beta1Client,
		},
	}

	for name, resourceObjType := range watcherConfigList {
		watcher := NewWatcher(resourceObjType.client, name, resourceObjType.resourceName, resourceObjType.objType, writer, cluster.watchers, esController)
		cluster.watchers[name] = watcher
		watcherList[name] = watcher
	}

	// TODO: create writer, with stop signal

	// // NOTE: only for k8s1.8
	//appsV1beta2WatcherList := map[string]ResourceObjType{
	//	"Deployment": {
	//		"deployments",
	//		&v1beta2.Deployment{},
	//	},
	//	"ReplicaSet": {
	//		"replicasets",
	//		&v1beta2.ReplicaSet{},
	//	},
	//}
	//for name, resourceObjType := range appsV1beta2WatcherList{
	//	watcher := NewWatcher(&appsV1beta2Client, name, resourceObjType.resourceName, resourceObjType.objType, writer, esController)
	//	cluster.watchers[name] = watcher
	//	watcherList[name] = watcher
	//}

	esController.writer = writer
	esController.clusterID = clusterID
	for resourceName, watcher := range watcherList {
		switch resourceName {
		case "Ingress":
			esController.lister.Ingress.Store = watcher.store
		case "Service":
			esController.lister.Service.Store = watcher.store
		case "EndPoints":
			esController.lister.Endpoint.Store = watcher.store
		case "ConfigMap":
			esController.lister.ConfigMap.Store = watcher.store
		case "Secret":
			esController.lister.Secret.Store = watcher.store
		default:
			continue
		}
	}
}

func (cluster *Cluster) Run(stop <-chan struct{}) {
	glog.Info("Cluster is ready to Run......")

	for resourceName, watcher := range cluster.watchers {
		glog.Infof("Cluster start list-watcher for: %s", resourceName)
		go watcher.Run(stop)
	}

	// TODO: go sync run here
	go cluster.synchronizer.Run(stop)
}
