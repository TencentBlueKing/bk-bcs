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

package mesos

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcsv2 "github.com/Tencent/bk-bcs/bcs-mesos/mesosv2/apis/bkbcs/v2"
	bcsfactory "github.com/Tencent/bk-bcs/bcs-mesos/mesosv2/generated/informers/externalversions"
	bcsclientset "github.com/Tencent/bk-bcs/bcs-mesos/mesosv2/generated/clientset/versioned"
	bcslister "github.com/Tencent/bk-bcs/bcs-mesos/mesosv2/generated/listers/bkbcs/v2"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-networkpolicy/options"

	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listercorev1 "k8s.io/client-go/listers/core/v1"
	listernetv1 "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
)

//containerInfo hold info from BcsContainer
type containerInfo struct {
	IPAddress   string `json:"IPAddress"`
	NodeAddress string `json:"NodeAddress"`
}

// MesosInformer data informer for kubernetes
type MesosInformer struct {
	opt                   *options.NetworkPolicyOption
	kubeClient            kubernetes.Interface
	informerFactory       informers.SharedInformerFactory
	bcsInformerFactory    bcsfactory.SharedInformerFactory
	taskgroupInformer     cache.SharedIndexInformer
	taskgroupLister       bcslister.TaskGroupLister
	nsInformer            cache.SharedIndexInformer
	nsLister              listercorev1.NamespaceLister
	netPolicyInformer     cache.SharedIndexInformer
	netPolicyLister       listernetv1.NetworkPolicyLister
	podEventHandlers      []cache.ResourceEventHandler
	taskgroupEventHandler cache.ResourceEventHandler

	stopCh chan struct{}
}

// New create MesosInformer
func New(opt *options.NetworkPolicyOption) *MesosInformer {
	return &MesosInformer{
		opt:    opt,
		stopCh: make(chan struct{}),
	}
}

// Init init informer for mesos
func (mi *MesosInformer) Init(client kubernetes.Interface, bcsClient bcsclientset.Interface) {
	mi.kubeClient = client
	informerFactory := informers.NewSharedInformerFactory(client, time.Duration(mi.opt.KubeResyncPeriod)*time.Second)
	bcsInformerFacotry := bcsfactory.NewSharedInformerFactory(bcsClient, time.Duration(mi.opt.KubeResyncPeriod)*time.Second)
	tgInformer := bcsInformerFacotry.Bkbcs().V2().TaskGroups()
	nsInformer := informerFactory.Core().V1().Namespaces()
	netPolicyInformer := informerFactory.Networking().V1().NetworkPolicies()

	mi.informerFactory = informerFactory
	mi.bcsInformerFactory = bcsInformerFacotry

	mi.taskgroupInformer = tgInformer.Informer()
	mi.nsInformer = nsInformer.Informer()
	mi.netPolicyInformer = netPolicyInformer.Informer()

	mi.taskgroupLister = tgInformer.Lister()
	mi.nsLister = nsInformer.Lister()
	mi.netPolicyLister = netPolicyInformer.Lister()

	mi.taskgroupEventHandler = mi.newTaskgroupEventHandler()
	mi.taskgroupInformer.AddEventHandler(mi.taskgroupEventHandler)
}

func (mi *MesosInformer) newTaskgroupEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			mi.onTaskgroupAdd(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			mi.onTaskgroupUpdate(oldObj, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			mi.onTaskgroupDelete(obj)
		},
	}
}

func (mi *MesosInformer) onTaskgroupAdd(obj interface{}) {
	tg := obj.(*bcsv2.TaskGroup)
	pod, err := taskgroupToPod(tg)
	if err != nil {
		blog.Warnf("convert new taskgroup to pod failed, err %s", err.Error())
		return
	}

	for _, h := range mi.podEventHandlers {
		h.OnAdd(pod)
	}
}

func (mi *MesosInformer) onTaskgroupUpdate(oldObj, newObj interface{}) {
	tgOld := oldObj.(*bcsv2.TaskGroup)
	tgNew := newObj.(*bcsv2.TaskGroup)
	podOld, err := taskgroupToPod(tgOld)
	if err != nil {
		blog.Warnf("convert old taskgroup to pod failed, err %s", err.Error())
		return
	}
	podNew, err := taskgroupToPod(tgNew)
	if err != nil {
		blog.Warnf("convert old taskgroup to pod failed, err %s", err.Error())
		return
	}

	for _, h := range mi.podEventHandlers {
		h.OnUpdate(podOld, podNew)
	}
}

func (mi *MesosInformer) onTaskgroupDelete(obj interface{}) {
	tg := obj.(*bcsv2.TaskGroup)
	pod, err := taskgroupToPod(tg)
	if err != nil {
		blog.Warnf("convert deleted taskgroup to pod failed, err %s", err.Error())
		return
	}

	for _, h := range mi.podEventHandlers {
		h.OnDelete(pod)
	}
}

func taskgroupToPod(tg *bcsv2.TaskGroup) (*corev1.Pod, error) {
	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              tg.GetName(),
			Namespace:         tg.GetNamespace(),
			ResourceVersion:   tg.GetResourceVersion(),
			Labels:            tg.GetLabels(),
			Annotations:       tg.GetAnnotations(),
			CreationTimestamp: tg.GetCreationTimestamp(),
			DeletionTimestamp: tg.GetDeletionTimestamp(),
		},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodPhase(tg.Spec.Status),
		},
	}
	for _, task := range tg.Spec.Taskgroup {
		info := new(containerInfo)
		if err := json.Unmarshal([]byte(task.StatusData), info); err != nil {
			blog.Errorf("taskgroup %s/%s decode task %s status data failed, %s", tg.GetNamespace(), tg.GetName(), task.ID, err.Error())
			continue
		}
		if len(info.IPAddress) != 0 {
			newPod.Status.HostIP = info.NodeAddress
			newPod.Status.PodIP = info.IPAddress
		}
		if len(info.IPAddress) == 0 && len(info.NodeAddress) != 0 {
			newPod.Status.HostIP = info.NodeAddress
			newPod.Status.PodIP = info.NodeAddress
		}
	}
	if len(newPod.Status.HostIP) == 0 {
		return nil, fmt.Errorf("converted pod from taskgroup %s/%s lost host ip", tg.GetNamespace(), tg.GetName())
	}
	return newPod, nil
}

func taskgroupListToPodList(taskgroups []*bcsv2.TaskGroup) []*corev1.Pod {
	var podList []*corev1.Pod
	for _, tg := range taskgroups {
		pod, err := taskgroupToPod(tg)
		if err != nil {
			blog.Warnf("convert taskgroup to pod failed, err %s", err.Error())
			continue
		}
		podList = append(podList, pod)
	}
	return podList
}

// AddPodEventHandler implements DataInformer interface
func (mi *MesosInformer) AddPodEventHandler(handler cache.ResourceEventHandler) {
	mi.podEventHandlers = append(mi.podEventHandlers, handler)
}

// AddNamespaceEventHandler implements DataInformer interface
func (mi *MesosInformer) AddNamespaceEventHandler(handler cache.ResourceEventHandler) {
	mi.nsInformer.AddEventHandler(handler)
}

// AddNetworkpolicyEventHandler implements DataInformer interface
func (mi *MesosInformer) AddNetworkpolicyEventHandler(handler cache.ResourceEventHandler) {
	mi.netPolicyInformer.AddEventHandler(handler)
}

// Run implements DataInformer interface
func (mi *MesosInformer) Run() error {
	// start informer factory and wait for cache sync, when timeout, return error
	syncFlag := make(chan bool)
	isFactorySync := false
	isBcsFactorySync := false
	mi.informerFactory.Start(mi.stopCh)
	mi.bcsInformerFactory.Start(mi.stopCh)
	go func() {
		blog.Infof("wait for informer factory cache sync")
		mi.informerFactory.WaitForCacheSync(mi.stopCh)
		isFactorySync = true
		syncFlag <- true
	}()
	go func() {
		blog.Infof("wait for bcs informer factory cache sync")
		mi.bcsInformerFactory.WaitForCacheSync(mi.stopCh)
		isBcsFactorySync = true
		syncFlag <- true
	}()
	select {
	case <-time.After(time.Duration(mi.opt.KubeCacheSyncTimeout) * time.Second):
		return fmt.Errorf("wait for cache sync timeout after %d seconds", mi.opt.KubeCacheSyncTimeout)
	case <-syncFlag:
		if isFactorySync && isBcsFactorySync {
			return nil
		}
	}
	return nil
}

// Stop implements DataInformer interface
func (mi *MesosInformer) Stop() {
	blog.Infof("stop mesos informer")
	close(mi.stopCh)
}

// ListAllPods implements DataInformer interface
func (mi *MesosInformer) ListAllPods() ([]*corev1.Pod, error) {
	taskgroups, err := mi.taskgroupLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("list all taskgroup failed, err %s", err.Error())
	}
	return taskgroupListToPodList(taskgroups), nil
}

// ListPodsByNamespace implements DataInformer interface
func (mi *MesosInformer) ListPodsByNamespace(ns string, labelsToMatch labels.Set) ([]*corev1.Pod, error) {
	taskgroups, err := mi.taskgroupLister.TaskGroups(ns).List(labels.SelectorFromSet(labelsToMatch))
	if err != nil {
		return nil, fmt.Errorf("list taskgroup in ns %s with label %v failed, err %s", ns, labelsToMatch.String(), err.Error())
	}
	return taskgroupListToPodList(taskgroups), nil
}

// ListNamespaces implements DataInformer interface
func (mi *MesosInformer) ListNamespaces(labelsToMatch labels.Set) ([]*corev1.Namespace, error) {
	return mi.nsLister.List(labels.SelectorFromSet(labelsToMatch))
}

// ListAllNetworkPolicy implements DataInformer interface
func (mi *MesosInformer) ListAllNetworkPolicy() ([]*networking.NetworkPolicy, error) {
	return mi.netPolicyLister.List(labels.Everything())
}
