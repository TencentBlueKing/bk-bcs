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

// Package util provides some util functions
package util

import (
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	client "k8s.io/client-go/kubernetes"
	v1appslister "k8s.io/client-go/listers/apps/v1"
	v1batchlister "k8s.io/client-go/listers/batch/v1"
	v1lister "k8s.io/client-go/listers/core/v1"
	v1storagelister "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/tools/cache"
)

// ListerRegistryExtend is a registry providing various listers to list pods or nodes matching conditions
type ListerRegistryExtend interface {
	kube_util.ListerRegistry
	PVLister() v1lister.PersistentVolumeLister
	PVCLister() v1lister.PersistentVolumeClaimLister
	SCLister() v1storagelister.StorageClassLister
}

type listerRegistryImpl struct {
	allNodeLister               kube_util.NodeLister
	readyNodeLister             kube_util.NodeLister
	scheduledPodLister          kube_util.PodLister
	unschedulablePodLister      kube_util.PodLister
	podDisruptionBudgetLister   kube_util.PodDisruptionBudgetLister
	daemonSetLister             v1appslister.DaemonSetLister
	replicationControllerLister v1lister.ReplicationControllerLister
	jobLister                   v1batchlister.JobLister
	replicaSetLister            v1appslister.ReplicaSetLister
	statefulSetLister           v1appslister.StatefulSetLister
	pvLister                    v1lister.PersistentVolumeLister
	pvcLister                   v1lister.PersistentVolumeClaimLister
	scLister                    v1storagelister.StorageClassLister
}

// NewListerRegistry returns a registry providing various listers to list pods or nodes matching conditions
func NewListerRegistry(allNode kube_util.NodeLister, readyNode kube_util.NodeLister, scheduledPod kube_util.PodLister,
	unschedulablePod kube_util.PodLister, podDisruptionBudgetLister kube_util.PodDisruptionBudgetLister,
	daemonSetLister v1appslister.DaemonSetLister, replicationControllerLister v1lister.ReplicationControllerLister,
	jobLister v1batchlister.JobLister, replicaSetLister v1appslister.ReplicaSetLister,
	statefulSetLister v1appslister.StatefulSetLister, pvLister v1lister.PersistentVolumeLister,
	pvcLister v1lister.PersistentVolumeClaimLister, scLister v1storagelister.StorageClassLister) ListerRegistryExtend {
	return listerRegistryImpl{
		allNodeLister:               allNode,
		readyNodeLister:             readyNode,
		scheduledPodLister:          scheduledPod,
		unschedulablePodLister:      unschedulablePod,
		podDisruptionBudgetLister:   podDisruptionBudgetLister,
		daemonSetLister:             daemonSetLister,
		replicationControllerLister: replicationControllerLister,
		jobLister:                   jobLister,
		replicaSetLister:            replicaSetLister,
		statefulSetLister:           statefulSetLister,
		pvLister:                    pvLister,
		pvcLister:                   pvcLister,
		scLister:                    scLister,
	}
}

// NewListerRegistryWithDefaultListers returns a registry filled with listers of the default implementations
func NewListerRegistryWithDefaultListers(
	kubeClient client.Interface, stopChannel <-chan struct{}) ListerRegistryExtend {
	unschedulablePodLister := kube_util.NewUnschedulablePodLister(kubeClient, stopChannel)
	scheduledPodLister := kube_util.NewScheduledPodLister(kubeClient, stopChannel)
	readyNodeLister := kube_util.NewReadyNodeLister(kubeClient, stopChannel)
	allNodeLister := kube_util.NewAllNodeLister(kubeClient, stopChannel)
	podDisruptionBudgetLister := kube_util.NewPodDisruptionBudgetLister(kubeClient, stopChannel)
	daemonSetLister := kube_util.NewDaemonSetLister(kubeClient, stopChannel)
	replicationControllerLister := kube_util.NewReplicationControllerLister(kubeClient, stopChannel)
	jobLister := kube_util.NewJobLister(kubeClient, stopChannel)
	replicaSetLister := kube_util.NewReplicaSetLister(kubeClient, stopChannel)
	statefulSetLister := kube_util.NewStatefulSetLister(kubeClient, stopChannel)
	pvLister := NewPVLister(kubeClient, stopChannel)
	pvcLister := NewPVCLister(kubeClient, stopChannel)
	scLister := NewSCLister(kubeClient, stopChannel)
	return NewListerRegistry(allNodeLister, readyNodeLister, scheduledPodLister,
		unschedulablePodLister, podDisruptionBudgetLister, daemonSetLister,
		replicationControllerLister, jobLister, replicaSetLister, statefulSetLister,
		pvLister, pvcLister, scLister)
}

// AllNodeLister returns the AllNodeLister registered to this registry
func (r listerRegistryImpl) AllNodeLister() kube_util.NodeLister {
	return r.allNodeLister
}

// ReadyNodeLister returns the ReadyNodeLister registered to this registry
func (r listerRegistryImpl) ReadyNodeLister() kube_util.NodeLister {
	return r.readyNodeLister
}

// ScheduledPodLister returns the ScheduledPodLister registered to this registry
func (r listerRegistryImpl) ScheduledPodLister() kube_util.PodLister {
	return r.scheduledPodLister
}

// UnschedulablePodLister returns the UnschedulablePodLister registered to this registry
func (r listerRegistryImpl) UnschedulablePodLister() kube_util.PodLister {
	return r.unschedulablePodLister
}

// PodDisruptionBudgetLister returns the podDisruptionBudgetLister registered to this registry
func (r listerRegistryImpl) PodDisruptionBudgetLister() kube_util.PodDisruptionBudgetLister {
	return r.podDisruptionBudgetLister
}

// DaemonSetLister returns the daemonSetLister registered to this registry
func (r listerRegistryImpl) DaemonSetLister() v1appslister.DaemonSetLister {
	return r.daemonSetLister
}

// ReplicationControllerLister returns the replicationControllerLister registered to this registry
func (r listerRegistryImpl) ReplicationControllerLister() v1lister.ReplicationControllerLister {
	return r.replicationControllerLister
}

// JobLister returns the jobLister registered to this registry
func (r listerRegistryImpl) JobLister() v1batchlister.JobLister {
	return r.jobLister
}

// ReplicaSetLister returns the replicaSetLister registered to this registry
func (r listerRegistryImpl) ReplicaSetLister() v1appslister.ReplicaSetLister {
	return r.replicaSetLister
}

// StatefulSetLister returns the statefulSetLister registered to this registry
func (r listerRegistryImpl) StatefulSetLister() v1appslister.StatefulSetLister {
	return r.statefulSetLister
}

// PVLister returns the pvLister registered to this registry
func (r listerRegistryImpl) PVLister() v1lister.PersistentVolumeLister {
	return r.pvLister
}

// PVCLister returns the pvcLister registered to this registry
func (r listerRegistryImpl) PVCLister() v1lister.PersistentVolumeClaimLister {
	return r.pvcLister
}

// SCLister returns the scLister registered to this registry
func (r listerRegistryImpl) SCLister() v1storagelister.StorageClassLister {
	return r.scLister
}

// NewPVLister builds a pv lister.
func NewPVLister(kubeClient client.Interface, stopchannel <-chan struct{}) v1lister.PersistentVolumeLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(),
		"persistentvolumes", apiv1.NamespaceAll, fields.Everything())
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(listWatcher, &appsv1.StatefulSet{}, time.Hour)
	lister := v1lister.NewPersistentVolumeLister(store)
	go reflector.Run(stopchannel)
	return lister
}

// NewPVCLister builds a pvc lister.
func NewPVCLister(kubeClient client.Interface, stopchannel <-chan struct{}) v1lister.PersistentVolumeClaimLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(),
		"persistentvolumeclaims", apiv1.NamespaceAll, fields.Everything())
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(listWatcher, &appsv1.StatefulSet{}, time.Hour)
	lister := v1lister.NewPersistentVolumeClaimLister(store)
	go reflector.Run(stopchannel)
	return lister
}

// NewSCLister builds a storageclasses lister.
func NewSCLister(kubeClient client.Interface, stopchannel <-chan struct{}) v1storagelister.StorageClassLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.StorageV1().RESTClient(),
		"storageclasses", apiv1.NamespaceAll, fields.Everything())
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(listWatcher, &appsv1.StatefulSet{}, time.Hour)
	lister := v1storagelister.NewStorageClassLister(store)
	go reflector.Run(stopchannel)
	return lister
}
