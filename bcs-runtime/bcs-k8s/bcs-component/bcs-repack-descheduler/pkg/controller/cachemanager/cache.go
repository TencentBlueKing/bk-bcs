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

package cachemanager

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	appsv1Informers "k8s.io/client-go/informers/apps/v1"
	corev1Informers "k8s.io/client-go/informers/core/v1"
	policyv1Informers "k8s.io/client-go/informers/policy/v1"
	policyv1beta1Informers "k8s.io/client-go/informers/policy/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/options"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	gamev1alpha1 "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/apis/tkex/v1alpha1"
	gameversioned "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/clientset/versioned"
	gameexternal "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/informers/externalversions"
	gameinformers "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/informers/externalversions/tkex/v1alpha1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/internal/utils"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/calculator"
)

var (
	once sync.Once

	cacheManager *CacheManager
)

// RegisterPodEventsFunc defines the func type that register to pod events
type RegisterPodEventsFunc func(pod *corev1.Pod)

// CacheInterface defines the interface of cache manager
type CacheInterface interface {
	Init() error
	InitWithKubeConfig(kubeConfig string) error
	Start(ctx context.Context) error
	GetKubernetesClient() *kubernetes.Clientset

	CordonNode(ctx context.Context, name string) error
	GetNode(ctx context.Context, name string) (*corev1.Node, error)
	ListNodes(ctx context.Context, names []string) ([]*corev1.Node, error)

	ListNamespaces(ctx context.Context) ([]*corev1.Namespace, error)

	GetPodOwnerName(ctx context.Context, namespace, podName string) (*corev1.Pod, string, error)
	GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error)
	PodBind(ctx context.Context, bind *corev1.Binding) error
	EvictionPod(ctx context.Context, podNamespace, podName string) error
	ListPods(ctx context.Context, namespace string, selector labels.Selector) ([]*corev1.Pod, error)

	RegisterPodDeleteEvents(id string, f RegisterPodEventsFunc)
	UnRegisterPodDeleteEvents(id string)

	GetReplicaset(ctx context.Context, namespace, name string) (*appsv1.ReplicaSet, error)
	GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error)
	GetStatefulSet(ctx context.Context, namespace, name string) (*appsv1.StatefulSet, error)
	GetGameDeployment(ctx context.Context, namespace, name string) (*gamev1alpha1.GameDeployment, error)
	GetGameStatefulSet(ctx context.Context, namespace, name string) (*gamev1alpha1.GameStatefulSet, error)

	ListPDBs(ctx context.Context, namespace string) ([]*policyv1.PodDisruptionBudget,
		[]*policyv1beta1.PodDisruptionBudget, error)
	ListPDBPods(ctx context.Context, namespace string) (podsMap map[string]*corev1.Pod, err error)

	BuildCalculatorRequest(ctx context.Context) (*calculator.CalculateConvergeRequest, error)
}

// CacheManager manages all the cache informers and provide resource query.
type CacheManager struct {
	op         *options.DeSchedulerOption
	kubeConfig string
	client     *kubernetes.Clientset
	gameClient *gameversioned.Clientset

	informerFactory     informers.SharedInformerFactory
	podInformer         corev1Informers.PodInformer
	nodeInformer        corev1Informers.NodeInformer
	replicasetInformer  appsv1Informers.ReplicaSetInformer
	deploymentInformer  appsv1Informers.DeploymentInformer
	statefulsetInformer appsv1Informers.StatefulSetInformer

	pdbV1Beta1Informer policyv1beta1Informers.PodDisruptionBudgetInformer
	pdbV1Informer      policyv1Informers.PodDisruptionBudgetInformer

	gameInformerFactory gameexternal.SharedInformerFactory
	gameDeployInformer  gameinformers.GameDeploymentInformer
	gameStateInformer   gameinformers.GameStatefulSetInformer

	policyGroupVersion string
	pdbGroupVersion    string

	registers *sync.Map
	stopChan  chan struct{}
}

// NewCacheManager create hte instance of CacheManager
func NewCacheManager() CacheInterface {
	once.Do(func() {
		cacheManager = &CacheManager{
			stopChan:  make(chan struct{}),
			registers: &sync.Map{},
			op:        options.GlobalConfigHandler().GetOptions(),
		}
	})
	return cacheManager
}

// GetKubernetesClient return kubernetes client
func (m *CacheManager) GetKubernetesClient() *kubernetes.Clientset {
	return m.client
}

// Init the cache manager with kubernetes in-cluster client
func (m *CacheManager) Init() error {
	client, err := utils.GetK8sInClusterClient()
	if err != nil {
		return errors.Wrapf(err, "CacheManager get in-cluster client failed")
	}
	m.client = client
	return m.init()
}

// InitWithKubeConfig the cache manager with kubernetes out-of-cluster client
func (m *CacheManager) InitWithKubeConfig(kubeConfig string) error {
	client, err := utils.GetK8sOutOfClusterClient(kubeConfig)
	if err != nil {
		return errors.Wrapf(err, "create k8s client failed")
	}
	m.client = client
	m.kubeConfig = kubeConfig
	return m.init()
}

// init will create some informers and start them
func (m *CacheManager) init() error {
	m.informerFactory = informers.NewSharedInformerFactory(m.client, apis.InformerReSyncPeriod)

	m.initGeneralInformer()
	if err := m.initPDBInformer(); err != nil {
		return errors.Wrapf(err, "init pdb informer failed")
	}
	if err := m.initGameWorkloadInformer(); err != nil {
		blog.Warnf("Init game workload informer failed: %v", err.Error())
	}

	policyGroupVersion, err := utils.SupportEviction(m.client)
	if err != nil {
		return errors.Wrapf(err, "CacheManager get policy group version failed")
	}
	m.policyGroupVersion = policyGroupVersion

	// Starts all the shared informers that have been created by the factory so far.
	m.informerFactory.Start(m.stopChan)
	// start game informer
	if m.gameInformerFactory != nil {
		m.gameInformerFactory.Start(m.stopChan)
	}

	blog.Infof("CacheManager waiting for informer sync...")
	waitCh := make(chan error)
	go func() {
		err := m.waitForSync()
		waitCh <- err
	}()
	waitTimeout := time.After(apis.WaitInformerSyncTimeout)
	select {
	case err = <-waitCh:
		if err != nil {
			return errors.Wrapf(err, "wait for informer sync failed")
		}
		return nil
	case <-waitTimeout:
		close(m.stopChan)
		return errors.Errorf("CacheManager sync informers cache timeout")
	}
}

func (m *CacheManager) initGeneralInformer() {
	m.podInformer = m.informerFactory.Core().V1().Pods()
	m.podInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) {},
			UpdateFunc: func(oldObj, newObj interface{}) {},
			DeleteFunc: func(obj interface{}) {
				pod := obj.(*corev1.Pod)
				m.registers.Range(func(key, value interface{}) bool {
					f := value.(RegisterPodEventsFunc)
					f(pod)
					return true
				})
			},
		})

	m.nodeInformer = m.informerFactory.Core().V1().Nodes()
	m.nodeInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) {},
			UpdateFunc: func(oldObj, newObj interface{}) {},
			DeleteFunc: func(obj interface{}) {},
		})

	m.replicasetInformer = m.informerFactory.Apps().V1().ReplicaSets()
	m.replicasetInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) {},
			UpdateFunc: func(oldObj, newObj interface{}) {},
			DeleteFunc: func(obj interface{}) {},
		})

	m.deploymentInformer = m.informerFactory.Apps().V1().Deployments()
	m.deploymentInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) {},
			UpdateFunc: func(oldObj, newObj interface{}) {},
			DeleteFunc: func(obj interface{}) {},
		})

	m.statefulsetInformer = m.informerFactory.Apps().V1().StatefulSets()
	m.statefulsetInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) {},
			UpdateFunc: func(oldObj, newObj interface{}) {},
			DeleteFunc: func(obj interface{}) {},
		})
}

func (m *CacheManager) initPDBInformer() error {
	pdbGroupVersion, err := utils.SupportPDB(m.client)
	if err != nil {
		return errors.Wrapf(err, "CacheManager get pdb group version failed")
	}
	m.pdbGroupVersion = pdbGroupVersion
	switch m.pdbGroupVersion {
	case apis.PDBGroupBetaVersion:
		m.pdbV1Beta1Informer = m.informerFactory.Policy().V1beta1().PodDisruptionBudgets()
		m.pdbV1Beta1Informer.Informer().AddEventHandler(
			cache.ResourceEventHandlerFuncs{
				AddFunc:    func(obj interface{}) {},
				UpdateFunc: func(oldObj, newObj interface{}) {},
				DeleteFunc: func(obj interface{}) {},
			})
	case apis.PDBGroupV1Version:
		m.pdbV1Informer = m.informerFactory.Policy().V1().PodDisruptionBudgets()
		m.pdbV1Informer.Informer().AddEventHandler(
			cache.ResourceEventHandlerFuncs{
				AddFunc:    func(obj interface{}) {},
				UpdateFunc: func(oldObj, newObj interface{}) {},
				DeleteFunc: func(obj interface{}) {},
			})
	default:
		return errors.Errorf("not found pdb group version '%s'", m.pdbGroupVersion)
	}
	return nil
}

// initGameWorkloadInformer init GameDeployment/GameStatefulSet informers
func (m *CacheManager) initGameWorkloadInformer() (err error) {
	supportGameDeploy, err := utils.SupportGameWorkload(m.client, apis.GameDeploymentName, apis.GameDeploymentKind)
	if err != nil {
		return errors.Wrapf(err, "check support gamedeployment failed")
	}
	if supportGameDeploy {
		var gameClient *gameversioned.Clientset
		if m.kubeConfig == "" {
			gameClient, err = utils.GetGameDeploymentClient()
		} else {
			gameClient, err = utils.GetGameDeployClientWithKubeCfg(m.kubeConfig)
		}
		if err != nil {
			return errors.Wrapf(err, "CacheManager get gamedeployment client failed")
		}
		m.gameClient = gameClient
		m.gameInformerFactory = gameexternal.
			NewSharedInformerFactory(gameClient, apis.InformerReSyncPeriod)
		m.gameDeployInformer = m.gameInformerFactory.Tkex().V1alpha1().GameDeployments()
		m.gameDeployInformer.Informer().AddEventHandler(
			cache.ResourceEventHandlerFuncs{
				AddFunc:    func(obj interface{}) {},
				UpdateFunc: func(oldObj, newObj interface{}) {},
				DeleteFunc: func(obj interface{}) {},
			})
		m.gameStateInformer = m.gameInformerFactory.Tkex().V1alpha1().GameStatefulSets()
		m.gameStateInformer.Informer().AddEventHandler(
			cache.ResourceEventHandlerFuncs{
				AddFunc:    func(obj interface{}) {},
				UpdateFunc: func(oldObj, newObj interface{}) {},
				DeleteFunc: func(obj interface{}) {},
			})
	} else {
		blog.Warnf("not support game workload type this cluster")
	}
	return nil
}

// waitForSync wait informers synced
func (m *CacheManager) waitForSync() error {
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(m.stopChan, m.nodeInformer.Informer().HasSynced) {
		return errors.Errorf("failed to sync node informer")
	}
	blog.Infof("CacheManager sync node informer success.")

	if !cache.WaitForCacheSync(m.stopChan, m.replicasetInformer.Informer().HasSynced) {
		return errors.Errorf("failed to sync replicaset informer")
	}
	blog.Infof("CacheManager sync replicaset informer success.")

	if !cache.WaitForCacheSync(m.stopChan, m.deploymentInformer.Informer().HasSynced) {
		return errors.Errorf("failed to sync deployment informer")
	}
	blog.Infof("CacheManager sync deployment informer success.")

	if !cache.WaitForCacheSync(m.stopChan, m.statefulsetInformer.Informer().HasSynced) {
		return errors.Errorf("failed to sync statefulset informer")
	}
	blog.Infof("CacheManager sync statefulset informer success.")

	if !cache.WaitForCacheSync(m.stopChan, m.podInformer.Informer().HasSynced) {
		return errors.Errorf("failed to sync pod informer")
	}
	blog.Infof("CacheManager sync pod informer success.")

	if m.pdbV1Informer != nil {
		if !cache.WaitForCacheSync(m.stopChan, m.pdbV1Informer.Informer().HasSynced) {
			return errors.Errorf("failed to sync pdb v1 informer")
		}
		blog.Infof("CacheManager sync pdb v1 informer success.")
	}
	if m.pdbV1Beta1Informer != nil {
		if !cache.WaitForCacheSync(m.stopChan, m.pdbV1Beta1Informer.Informer().HasSynced) {
			return errors.Errorf("failed to sync pdb v1beta1 informer")
		}
		blog.Infof("CacheManager sync pdb v1beta1 informer success.")
	}

	if m.gameDeployInformer != nil {
		if !cache.WaitForCacheSync(m.stopChan, m.gameDeployInformer.Informer().HasSynced) {
			return errors.Errorf("failed to sync gamedeployment informer")
		}
		blog.Infof("CacheManager sync gamedeployment informer success.")
	}
	if m.gameStateInformer != nil {
		if !cache.WaitForCacheSync(m.stopChan, m.gameStateInformer.Informer().HasSynced) {
			return errors.Errorf("failed to sync gamedeployment informer")
		}
		blog.Infof("CacheManager sync gamedeployment informer success.")
	}
	return nil
}

// Start will start the CacheManager with context
func (m *CacheManager) Start(ctx context.Context) error {
	for range ctx.Done() {
		close(m.stopChan)
		blog.Infof("cache manager is stopped.")
	}
	return nil
}

// CordonNode cordon the node
func (m *CacheManager) CordonNode(ctx context.Context, name string) error {
	payload := []map[string]interface{}{{
		"op":    "replace",
		"path":  "/spec/unschedulable",
		"value": true,
	}}
	payloadBytes, _ := json.Marshal(payload)
	_, err := m.client.CoreV1().Nodes().Patch(ctx, name, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		return errors.Wrapf(err, "cordon node '%s' failed", name)
	}
	return nil
}

// GetNode will query the node with name
func (m *CacheManager) GetNode(ctx context.Context, name string) (*corev1.Node, error) {
	node, err := m.nodeInformer.Lister().Get(name)
	if err != nil {
		return nil, errors.Wrapf(err, "get node '%s' from informer failed", name)
	}
	return node, nil
}

// ListNodes will list all the nodes
func (m *CacheManager) ListNodes(ctx context.Context, names []string) ([]*corev1.Node, error) {
	nodes, err := m.nodeInformer.Lister().List(labels.NewSelector())
	if err != nil {
		return nil, errors.Wrapf(err, "list nodes from informer failed")
	}
	if len(names) == 0 {
		return nodes, nil
	}

	nodeMap := make(map[string]string)
	for i := range names {
		nodeMap[names[i]] = names[i]
	}
	result := make([]*corev1.Node, 0, len(names))
	for _, node := range nodes {
		if _, ok := nodeMap[node.Name]; ok {
			result = append(result, node)
		}
	}
	return result, nil
}

// ListNamespaces will list all the namespaces
func (m *CacheManager) ListNamespaces(ctx context.Context) ([]*corev1.Namespace, error) {
	namespaceList, err := m.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "list namespaces failed")
	}
	namespaces := make([]*corev1.Namespace, 0, len(namespaceList.Items))
	for i := range namespaceList.Items {
		namespaces = append(namespaces, &namespaceList.Items[i])
	}
	return namespaces, nil
}

// GetPodOwnerName will return the owner name of pod.
func (m *CacheManager) GetPodOwnerName(ctx context.Context, namespace, podName string) (*corev1.Pod, string, error) {
	pod, err := m.GetPod(ctx, namespace, podName)
	if err != nil {
		return nil, "", errors.Wrapf(err, "get pod '%s/%s' failed", namespace, podName)
	}

	if len(pod.OwnerReferences) == 0 {
		return nil, "", errors.Wrapf(err, "pod '%s/%s' no ownerReferences", namespace, podName)
	}
	var podOwnKind, podOwnName string
	for _, reference := range pod.OwnerReferences {
		if _, ok := apis.SupportKind[strings.ToLower(reference.Kind)]; ok {
			podOwnKind = reference.Kind
			podOwnName = reference.Name
			break
		}

		// If pod owner kind is ReplicaSet, should return deployment information
		// as pod's owner.
		if reference.Kind == apis.ReplicaSetKind {
			refer, err := m.getReplicaSetOwnerReference(ctx, namespace, reference.Name)
			if err != nil {
				return nil, "", errors.Wrapf(err, "replicaset get owner failed")
			}
			if refer.Kind != apis.DeploymentKind {
				return nil, "", errors.Errorf("replicaset '%s/%s' owner not deployment",
					namespace, reference.Name)
			}
			podOwnKind = refer.Kind
			podOwnName = refer.Name
			break
		}
	}
	return pod, apis.NamespacedWorkloadKind(namespace, podOwnName, podOwnKind), nil
}

func (m *CacheManager) getReplicaSetOwnerReference(ctx context.Context, namespace,
	replicaSetName string) (refer metav1.OwnerReference, err error) {
	rs, err := m.GetReplicaset(ctx, namespace, replicaSetName)
	if err != nil {
		return refer, errors.Wrapf(err, "get replicaset '%s/%s' failed", namespace, replicaSetName)
	}
	if len(rs.OwnerReferences) == 0 {
		return refer, errors.Wrapf(err, "replicaset '%s/%s' no owner reference", namespace, replicaSetName)
	}
	for _, reference := range rs.OwnerReferences {
		if _, ok := apis.SupportKind[strings.ToLower(reference.Kind)]; ok {
			return reference, nil
		}
	}
	return refer, errors.Wrapf(err, "replicaset '%s/%s' not get support owner", namespace, replicaSetName)
}

// GetPod will query pod with namespace and name
func (m *CacheManager) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	pod, err := m.podInformer.Lister().Pods(namespace).Get(name)
	if err != nil {
		return nil, errors.Wrapf(err, "get node '%s' from informer failed", name)
	}
	return pod, nil
}

// PodBind create the bind resource with pod
func (m *CacheManager) PodBind(ctx context.Context, bind *corev1.Binding) error {
	return m.client.CoreV1().Pods(bind.Namespace).Bind(ctx, bind, metav1.CreateOptions{})
}

// EvictionPod will evict pod
func (m *CacheManager) EvictionPod(ctx context.Context, podNamespace, podName string) error {
	if strings.Contains(m.policyGroupVersion, "v1beta1") {
		if err := m.client.PolicyV1beta1().Evictions(podNamespace).Evict(ctx, &policyv1beta1.Eviction{
			TypeMeta: metav1.TypeMeta{
				APIVersion: m.policyGroupVersion,
				Kind:       apis.EvictionKind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: podNamespace,
			},
			DeleteOptions: &metav1.DeleteOptions{},
		}); err != nil {
			return errors.Wrapf(err, "evict '%s/%s' failed", podNamespace, podName)
		}
		return nil
	}

	eviction := &policyv1.Eviction{
		TypeMeta: metav1.TypeMeta{
			APIVersion: m.policyGroupVersion,
			Kind:       apis.EvictionKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: podNamespace,
		},
		DeleteOptions: &metav1.DeleteOptions{},
	}
	if err := m.client.PolicyV1().Evictions(eviction.Namespace).Evict(ctx, eviction); err != nil {
		return errors.Wrapf(err, "evict '%s/%s' failed", podNamespace, podName)
	}
	return nil
}

// ListPods will list pods all namespaces
func (m *CacheManager) ListPods(ctx context.Context, namespace string,
	selector labels.Selector) ([]*corev1.Pod, error) {
	var pods []*corev1.Pod
	var err error
	if namespace != "" {
		pods, err = m.podInformer.Lister().Pods(namespace).List(selector)
	} else {
		pods, err = m.podInformer.Lister().List(selector)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "list pods from informer failed")
	}
	return pods, nil
}

// GetReplicaset get replicaset with namespace and name
func (m *CacheManager) GetReplicaset(ctx context.Context, namespace, name string) (*appsv1.ReplicaSet, error) {
	replicaset, err := m.replicasetInformer.Lister().ReplicaSets(namespace).Get(name)
	if err != nil {
		return nil, errors.Wrapf(err, "get replicaset '%s' from informer failed", name)
	}
	return replicaset, nil
}

// GetDeployment get deployment with namespace and name
func (m *CacheManager) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	deployment, err := m.deploymentInformer.Lister().Deployments(namespace).Get(name)
	if err != nil {
		return nil, errors.Wrapf(err, "get deployment '%s' from informer failed", name)
	}
	return deployment, nil
}

// GetStatefulSet get statefulset with namespace and name
func (m *CacheManager) GetStatefulSet(ctx context.Context, namespace, name string) (*appsv1.StatefulSet, error) {
	statefulset, err := m.statefulsetInformer.Lister().StatefulSets(namespace).Get(name)
	if err != nil {
		return nil, errors.Wrapf(err, "get statefulset '%s' from informer failed", name)
	}
	return statefulset, nil
}

// GetGameDeployment get game deployment with namespace and name
func (m *CacheManager) GetGameDeployment(ctx context.Context, namespace,
	name string) (*gamev1alpha1.GameDeployment, error) {
	if m.gameDeployInformer == nil {
		return nil, errors.Errorf("not support GameDeployment type")
	}

	gameDeployment, err := m.gameDeployInformer.Lister().GameDeployments(namespace).Get(name)
	if err != nil {
		return nil, errors.Wrapf(err, "get gamedeployment '%s' from informer failed", name)
	}
	return gameDeployment, nil
}

// GetGameStatefulSet get game statefulset with namespace and name
func (m *CacheManager) GetGameStatefulSet(ctx context.Context, namespace,
	name string) (*gamev1alpha1.GameStatefulSet, error) {
	if m.gameStateInformer == nil {
		return nil, errors.Errorf("not support GameStatefuSet type")
	}

	gameStatefulset, err := m.gameStateInformer.Lister().GameStatefulSets(namespace).Get(name)
	if err != nil {
		return nil, errors.Wrapf(err, "get gamestatefulset '%s' from informer failed", name)
	}
	return gameStatefulset, nil
}

// ListPDBs list all PDBs with namespace, if namespace is empty, it will return all namespaces
func (m *CacheManager) ListPDBs(ctx context.Context, namespace string) ([]*policyv1.PodDisruptionBudget,
	[]*policyv1beta1.PodDisruptionBudget, error) {
	var err error
	if m.pdbV1Informer != nil {
		var v1PDBs []*policyv1.PodDisruptionBudget
		if namespace != "" {
			v1PDBs, err = m.pdbV1Informer.Lister().PodDisruptionBudgets(namespace).List(labels.NewSelector())
		} else {
			v1PDBs, err = m.pdbV1Informer.Lister().List(labels.NewSelector())
		}
		if err != nil {
			return nil, nil, errors.Wrapf(err, "list v1 pdbs '%s' from informer failed", namespace)
		}
		return v1PDBs, nil, nil
	}
	if m.pdbV1Beta1Informer != nil {
		var v1beta1PDBs []*policyv1beta1.PodDisruptionBudget
		if namespace != "" {
			v1beta1PDBs, err = m.pdbV1Beta1Informer.Lister().PodDisruptionBudgets(namespace).List(labels.NewSelector())
		} else {
			v1beta1PDBs, err = m.pdbV1Beta1Informer.Lister().List(labels.NewSelector())
		}
		if err != nil {
			return nil, nil, errors.Wrapf(err, "list v1beta1 pdbs '%s' from informer failed", namespace)
		}
		return nil, v1beta1PDBs, nil
	}
	return nil, nil, errors.Errorf("no such pdb group version can be found")
}

// ListPDBPods will return the podsMap that pdb selector selected
func (m *CacheManager) ListPDBPods(ctx context.Context, namespace string) (podsMap map[string]*corev1.Pod, err error) {
	v1PDBs, v1beta1PDBs, err := m.ListPDBs(ctx, namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "list pdbs failed from '%s'", namespace)
	}
	podsMap = make(map[string]*corev1.Pod)
	for _, pdb := range v1PDBs {
		if err = m.listPodsWithPDBSelector(ctx, pdb.Namespace, pdb.Spec.Selector, podsMap); err != nil {
			blog.Errorf("ListPDBPods v1 list pdb '%s/%s' pods failed: %s",
				pdb.Namespace, pdb.Name, err.Error())
		}
	}
	for _, pdb := range v1beta1PDBs {
		if err = m.listPodsWithPDBSelector(ctx, pdb.Namespace, pdb.Spec.Selector, podsMap); err != nil {
			blog.Errorf("ListPDBPods v1beta1 list pdb '%s/%s' pods failed: %s",
				pdb.Namespace, pdb.Name, err.Error())
		}
	}
	return podsMap, nil
}

func (m *CacheManager) listPodsWithPDBSelector(ctx context.Context, namespace string,
	pdbSelector *metav1.LabelSelector, podsMap map[string]*corev1.Pod) error {
	selector, err := metav1.LabelSelectorAsSelector(pdbSelector)
	if err != nil {
		return errors.Wrapf(err, "selector convert failed")
	}
	pods, err := m.ListPods(ctx, namespace, selector)
	if err != nil {
		return errors.Wrapf(err, "list pods failed")
	}
	for _, pod := range pods {
		podsMap[apis.PodName(pod.Namespace, pod.Name)] = pod
	}
	return nil
}

// RegisterPodDeleteEvents register the events of pod deletion
func (m *CacheManager) RegisterPodDeleteEvents(id string, f RegisterPodEventsFunc) {
	m.registers.Store(id, f)
}

// UnRegisterPodDeleteEvents unregister the events of pod deletion.
func (m *CacheManager) UnRegisterPodDeleteEvents(id string) {
	m.registers.Delete(id)
}
