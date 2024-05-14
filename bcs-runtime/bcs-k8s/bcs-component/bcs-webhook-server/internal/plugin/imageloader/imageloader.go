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

// Package imageloader can load container images before in-place update really happens,
// which shortens the time of container unavailable.
package imageloader

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/api/admission/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	batchlister "k8s.io/client-go/listers/batch/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginmanager"
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
	p := &imageLoader{}
	pluginmanager.Register(pluginName, p)
}

// ImageLoaderConfig config of image loader
// nolint
type ImageLoaderConfig struct {
	// Workload supported workload
	Workload string `json:"workload"`
	// JobTimeoutSeconds timeout seconds of job
	JobTimeoutSeconds int64 `json:"jobTimeoutSeconds"`
}

type imageLoader struct {
	stopCh chan struct{}
	config ImageLoaderConfig

	kubeConfig *rest.Config
	k8sClient  kubernetes.Interface
	workloads  map[string]Workload

	nodeLister   corev1lister.NodeLister
	jobLister    batchlister.JobLister
	secretLister corev1lister.SecretLister

	queue workqueue.RateLimitingInterface
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
	fileBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		blog.Errorf("load config file %s failed, err %s", configFilePath, err.Error())
		return fmt.Errorf("load config file %s failed, err %s", configFilePath, err.Error())
	}
	newConfig := &ImageLoaderConfig{}
	err = json.Unmarshal(fileBytes, &newConfig)
	if err != nil {
		blog.Errorf("decode config %s failed, err %s", string(fileBytes), err.Error())
		return fmt.Errorf("decode config %s failed, err %s", string(fileBytes), err.Error())
	}
	i.config = *newConfig

	// DOTO set burst and qps
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
	}
	blog.Info("connect to k8s with default client success")

	// regist workloads
	err = i.registWorkloads()
	if err != nil {
		blog.Errorf("%v", err)
		return err
	}
	if !workloadsWaitForCacheSync(i.stopCh) {
		return fmt.Errorf("workloads cache synced failed")
	}

	// create imageloader namespace if not exist
	_, err = i.k8sClient.CoreV1().Namespaces().Get(context.Background(), pluginName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		namespace := apiv1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: pluginName},
		}
		_, err = i.k8sClient.CoreV1().Namespaces().Create(context.Background(), &namespace, metav1.CreateOptions{})
		if err != nil {
			blog.Errorf("failed to create namespace: %v", err)
			return err
		}
	}

	// listen bcs-gamedeployment to compare update, listen imageload job to execute the update
	corev1InformerFactory := kubeinformers.NewSharedInformerFactory(i.k8sClient, 0)

	// add handler for imageload job
	i.queue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),
		pluginName)
	jobInformer := corev1InformerFactory.Batch().V1().Jobs().Informer()
	jobInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    i.addJob,
			UpdateFunc: i.updateJob,
			// do nothing
			DeleteFunc: func(interface{}) {},
		})
	i.jobLister = corev1InformerFactory.Batch().V1().Jobs().Lister()

	// set node lister to get images on node
	nodeInformer := corev1InformerFactory.Core().V1().Nodes().Informer()
	i.nodeLister = corev1InformerFactory.Core().V1().Nodes().Lister()

	secretInformer := corev1InformerFactory.Core().V1().Secrets().Informer()
	i.secretLister = corev1InformerFactory.Core().V1().Secrets().Lister()

	i.stopCh = make(chan struct{})
	corev1InformerFactory.Start(i.stopCh)
	if !cache.WaitForCacheSync(i.stopCh, nodeInformer.HasSynced, jobInformer.HasSynced, secretInformer.HasSynced) {
		return fmt.Errorf("Wait for cache failed")
	}

	workers := 1
	go i.run(workers) // nolint
	return nil
}

// processNextWorkItem dequeues items, processes them, and marks them done. It enforces that the syncHandler is never
// invoked concurrently with the same key.
func (i *imageLoader) processNextWorkItem() bool {
	key, quit := i.queue.Get()
	if quit {
		return false
	}
	defer i.queue.Done(key)
	blog.Infof("processNextWorkItem get item: %#v", key)
	if err := i.sync(key.(string)); err != nil {
		utilruntime.HandleError(fmt.Errorf("error syncing GameDeployment %v, requeuing: %v", key.(string), err))
		i.queue.AddRateLimited(key)
	} else {
		i.queue.Forget(key)
	}
	return true
}

// worker runs a worker goroutine that invokes processNextWorkItem until the controller's queue is closed
func (i *imageLoader) worker() {
	for i.processNextWorkItem() {
	}
}

// nolint always nil
func (i *imageLoader) run(workers int) error {
	defer utilruntime.HandleCrash()
	defer i.queue.ShutDown()

	for j := 0; j < workers; j++ {
		go wait.Until(i.worker, time.Second, i.stopCh)
	}

	blog.Info("Started workers")
	<-i.stopCh
	blog.Info("Shutting down workers")

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

// Close xxx
// DOTO clean resources like connections, files
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
