/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at * http://opensource.org/licenses/MIT
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

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	glog "bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/bcs"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/options"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output"
)

// WatcherManager is resource watcher manager.
type WatcherManager struct {
	// normal k8s resource watchers.
	watchers map[string]WatcherInterface

	// synchronizer syncs normal metadata got by watchers to storage.
	synchronizer *Synchronizer

	// special resource watchers.
	netserviceWatcher *NetServiceWatcher
}

// NewWatcherManager creates a new WatcherManager instance.
func NewWatcherManager(writer *output.Writer, k8sConfig *options.K8sConfig,
	clusterID string, storageService, netservice *bcs.InnerService) (*WatcherManager, error) {

	mgr := &WatcherManager{
		watchers: make(map[string]WatcherInterface),
	}
	mgr.initWatchers(clusterID, writer, k8sConfig, storageService, netservice)

	mgr.synchronizer = NewSynchronizer(clusterID, mgr.watchers, storageService)
	return mgr, nil
}

func (mgr *WatcherManager) newClientSet(k8sConfig *options.K8sConfig) (*kubernetes.Clientset, error) {
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

	return kubernetes.NewForConfig(config)
}

func (mgr *WatcherManager) initWatchers(clusterID string, writer *output.Writer,
	k8sconfig *options.K8sConfig, storageService, netservice *bcs.InnerService) {

	// create k8s clientset.
	clientSet, err := mgr.newClientSet(k8sconfig)
	if err != nil {
		panic(err)
	}

	coreClient := clientSet.CoreV1().RESTClient()
	extensionsV1Beta1Client := clientSet.ExtensionsV1beta1().RESTClient()
	batchV1Client := clientSet.BatchV1().RESTClient()
	appsV1Beta1Client := clientSet.AppsV1beta1().RESTClient()

	// build watcher configs.
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
		watcher := NewWatcher(resourceObjType.client, name, resourceObjType.resourceName, resourceObjType.objType, writer, mgr.watchers)
		mgr.watchers[name] = watcher
	}

	/* NOTE: only for k8s1.8
	appsV1beta2WatcherList := map[string]ResourceObjType{
	   "Deployment": {
		   "deployments",
	 	   &v1beta2.Deployment{},
	    },
	    "ReplicaSet": {
	 	   "replicasets",
	        &v1beta2.ReplicaSet{},
	    },
	}
	for name, resourceObjType := range appsV1beta2WatcherList{
	    watcher := NewWatcher(&appsV1beta2Client, name, resourceObjType.resourceName, resourceObjType.objType, writer)
	    cluster.watchers[name] = watcher
	}
	*/

	// init netservice watcher.
	mgr.netserviceWatcher = NewNetServiceWatcher(clusterID, "IPPoolStatic", storageService, netservice)
}

// Run starts the watcher manager, and runs all watchers.
func (mgr *WatcherManager) Run(stopCh <-chan struct{}) {
	// run normal resource watchers.
	for resourceName, watcher := range mgr.watchers {
		glog.Infof("watcher manager, start list-watcher[%+v]", resourceName)
		go watcher.Run(stopCh)
	}

	// run netservice watcher.
	go mgr.netserviceWatcher.Run(stopCh)

	// run synchronizer.
	go mgr.synchronizer.Run(stopCh)
}
