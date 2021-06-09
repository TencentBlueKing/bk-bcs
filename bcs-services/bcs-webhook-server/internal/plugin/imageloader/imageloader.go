/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package imageloader can load container images before in-place update really happens,
// which shortens the time of container unavailable.
package imageloader

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/internal/pluginmanager"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corev1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	//	"k8s.io/client-go/tools/clientcmd"
)

const (
	// pluginName is the unique of the hook-server plugin,
	// in this plugin, its also the namespace to manage jobs.
	pluginName          = "imageloader"
	pluginAnnotationKey = pluginName + ".webhook.bkbcs.tencent.com"
)

var (
	jsonPatchType = v1beta1.PatchTypeJSONPatch
)

func init() {
	// TODO register plugin to hook-server
	p := &imageLoader{}
	pluginmanager.Register(pluginName, p)
}

type imageLoader struct {
	stopCh chan struct{}

	kubeConfig *rest.Config
	k8sClient  *kubernetes.Clientset
	workloads  map[string]Workload

	nodeLister corev1lister.NodeLister
}

func (i *imageLoader) registWorkloads() error {
	wls, err := InitWorkloads(i)
	if err != nil {
		return err
	}
	i.workloads = wls
	return nil
}

// AnnotationKey returns key of the imageloader plugin for hook server to identify.
func (i *imageLoader) AnnotationKey() string {
	return pluginAnnotationKey
}

// Init inits job, node and workload's informer.
func (i *imageLoader) Init(configFilePath string) error {
	blog.V(3).Infof("init imageloader with configFilePath %s", configFilePath)
	// TODO read config from configFilePath
	// TODO set burst and qps
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	i.kubeConfig = config

	// create job and node client
	i.k8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		blog.Errorf("%v", err)
		return err
	} else {
		blog.Info("connect to k8s with default client success")
	}

	// TODO select workload by config
	// regist workloads
	err = i.registWorkloads()
	if err != nil {
		blog.Errorf("%v", err)
		return err
	}
	if !workloadsWaitForCacheSync(i.stopCh) {
		return fmt.Errorf("workloads cache synced failed")
	}

	// TODO create imageloader namespace if not exist

	// listen bcs-gamedeployment to compare update, listen imageload job to execute the update
	// node is a cluster-scoped resource, it's ok to specify the namespace
	corev1InformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(i.k8sClient, 0,
		kubeinformers.WithNamespace(pluginName))

	// TODO start workload informer

	// add handler for imageload job
	// TODO use workqueue
	corev1InformerFactory.Batch().V1().Jobs().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    i.addJob,
			UpdateFunc: i.updateJob,
			// do nothing
			DeleteFunc: func(interface{}) {},
		})

	// set node lister to get images on node
	// add node informer
	nodeInformer := corev1InformerFactory.Core().V1().Nodes().Informer()
	i.nodeLister = corev1InformerFactory.Core().V1().Nodes().Lister()
	i.stopCh = make(chan struct{})
	corev1InformerFactory.Start(i.stopCh)
	if !cache.WaitForCacheSync(i.stopCh, nodeInformer.HasSynced) {
		return fmt.Errorf("node cache synced failed")
	}
	return nil
}

// Handle handles webhook request of imageloader.
func (i *imageLoader) Handle(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	// check if it is an update operation
	if ar.Request.Operation != v1beta1.Update {
		return toAdmissionResponse(nil)
	}

	started := time.Now()
	// call different workload handle by metav1.GroupVersionKind(like v1.Pod)
	// find workload
	reqName := ar.Request.Kind.String()
	workload, ok := i.workloads[reqName]
	if !ok {
		metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusFailure, started)
		err := fmt.Errorf("workload %s is not supported in imageloader", reqName)
		blog.Errorf("%v", err)
		return toAdmissionResponse(err)
	}

	status := metrics.StatusSuccess
	admissionResponse := workload.LoadImageBeforeUpdate(ar)
	if !admissionResponse.Allowed {
		status = metrics.StatusFailure
	}
	metrics.ReportBcsWebhookServerPluginLantency(pluginName, status, started)
	return admissionResponse
}

// TODO clean resources like connections, files
func (i *imageLoader) Close() error {
	return nil
}

// toAdmissionResponse is a helper function to create an AdmissionResponse
// with an embedded error
// or if err is nil, return patch with allowance
// only first patch used
func toAdmissionResponse(err error, patchs ...string) *v1beta1.AdmissionResponse {
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}
	patch := ""
	if len(patchs) != 0 {
		patch = patchs[0]
	}
	return &v1beta1.AdmissionResponse{
		Allowed:   true,
		Patch:     []byte(patch),
		PatchType: &jsonPatchType,
	}
}
