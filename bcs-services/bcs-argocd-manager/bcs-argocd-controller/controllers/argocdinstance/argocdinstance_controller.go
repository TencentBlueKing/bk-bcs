/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package argocdinstance

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-controller/common"
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/apis/tkex/v1alpha1"
	clientset "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned"
	tkexscheme "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned/scheme"
	informers "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/informers/externalversions/tkex/v1alpha1"
	listers "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/listers/tkex/v1alpha1"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const controllerAgentName = "bcs-argocd-controller"

const (
	// SuccessSynced 用来表示事件被成功同步
	SuccessSynced = "Synced"
	// MessageResourceSynced 表示事件被触发时的消息信息
	MessageResourceSynced = "argocd instance synced successfully"
)

type InstanceController struct {
	kubeclientset   kubernetes.Interface
	tkexclientset   clientset.Interface
	instanceLister  listers.ArgocdInstanceLister
	namespaceLister corev1listers.NamespaceLister
	serviceLister   corev1listers.ServiceLister
	instanceSynced  cache.InformerSynced
	namespaceSynced cache.InformerSynced
	serviceSynced   cache.InformerSynced
	workqueue       workqueue.RateLimitingInterface
	recorder        record.EventRecorder
	kubeconfig      *restclient.Config
}

// NewController 初始化Controller
func NewController(kubeconfig *restclient.Config, kubeclientset kubernetes.Interface, clientset clientset.Interface,
	instanceInformer informers.ArgocdInstanceInformer, namespaceInformer corev1informers.NamespaceInformer, serviceInformer corev1informers.ServiceInformer) *InstanceController {

	utilruntime.Must(tkexscheme.AddToScheme(scheme.Scheme))
	blog.Info("Create event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(blog.Infof)
	// report events to APIServer
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	// 初始化Controller
	controller := &InstanceController{
		kubeclientset:   kubeclientset,
		tkexclientset:   clientset,
		instanceLister:  instanceInformer.Lister(),
		instanceSynced:  instanceInformer.Informer().HasSynced,
		namespaceLister: namespaceInformer.Lister(),
		namespaceSynced: namespaceInformer.Informer().HasSynced,
		serviceLister:   serviceInformer.Lister(),
		serviceSynced:   serviceInformer.Informer().HasSynced,
		workqueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ArgocdController"),
		recorder:        recorder,
		kubeconfig:      kubeconfig,
	}
	blog.Info("Start up event handlers")

	// register event handlers
	instanceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueArgocdInstance,
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldInstance := oldObj.(*tkexv1alpha1.ArgocdInstance)
			newInstance := newObj.(*tkexv1alpha1.ArgocdInstance)
			if oldInstance.ResourceVersion == newInstance.ResourceVersion {
				return
			}
			controller.enqueueArgocdInstance(newObj)
		},
		DeleteFunc: controller.enqueueArgocdInstanceForDelete,
	})
	return controller
}

// Run start the controller
func (c *InstanceController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShuttingDown()

	blog.Info("start controller, cache sync")
	// sync cache
	if ok := cache.WaitForCacheSync(stopCh, c.instanceSynced, c.namespaceSynced, c.serviceSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	blog.Info("begin start worker thread")
	// start worker thread
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	blog.Info("worker thread started")
	<-stopCh
	blog.Info("worker thread stopped")
	return nil
}

func (c *InstanceController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem process the next item in the queue
func (c *InstanceController) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// handle business logic
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		c.workqueue.Forget(obj)
		blog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler syncs the ArgocdInstance with the given key
func (c *InstanceController) syncHandler(key string) error {
	namespace, instanceName, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	// init helm action client
	flags := genericclioptions.NewConfigFlags(false)
	flags.Namespace = &instanceName
	flags.BearerToken = &c.kubeconfig.BearerToken
	flags.CAFile = &c.kubeconfig.CAFile
	flags.KeyFile = &c.kubeconfig.KeyFile
	flags.APIServer = &c.kubeconfig.Host
	flags.Username = &c.kubeconfig.Username
	flags.Password = &c.kubeconfig.Password
	flags.TLSServerName = &c.kubeconfig.ServerName

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(flags, instanceName, "", blog.Info); err != nil {
		blog.Errorf("init helm action config failed: %v", err)
		return err
	}

	// get ArgocdInstance from cache
	instance, err := c.instanceLister.ArgocdInstances(namespace).Get(instanceName)
	if err != nil {
		if errors.IsNotFound(err) {
			blog.Infof("ArgocdInstance %s/%s deleted", namespace, instanceName)
			// check helm release exists
			actionStatus := action.NewStatus(actionConfig)
			_, err := actionStatus.Run(instanceName)
			if err == nil {
				// exists, uninstall it
				actionDelete := action.NewUninstall(actionConfig)
				_, err = actionDelete.Run(instanceName)
				if err != nil {
					utilruntime.HandleError(err)
					return err
				}
			} else if err == driver.ErrReleaseNotFound {
				// not exists, do nothing
			} else {
				utilruntime.HandleError(err)
				return err
			}
			// check and delete ns
			ns, err := c.kubeclientset.CoreV1().Namespaces().Get(context.TODO(), instanceName, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					return nil
				} else {
					utilruntime.HandleError(err)
					return err
				}
			}
			blog.Infof("deleting ns [%s]", ns.GetName())
			if err := c.kubeclientset.CoreV1().Namespaces().
				Delete(context.TODO(), ns.GetName(), metav1.DeleteOptions{}); err != nil {
				utilruntime.HandleError(err)
				return err
			} else {
				blog.Infof("Namespace %s deleted", instanceName)
				return nil
			}
		} else {
			utilruntime.HandleError(err)
			return err
		}
	}

	// sync argocd instance to desired state
	// 1. sync namespace
	ns, err := c.namespaceLister.Get(instance.GetName())
	if errors.IsNotFound(err) {
		// 如果没有找到，就创建
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: instanceName,
				Labels: map[string]string{
					common.ArgoCDKeyPartOf:   common.ArgocdManagerAppName,
					common.ArgocdKeyProject:  instance.Spec.Project,
					common.ArgocdKeyInstance: instance.GetName(),
				},
			},
		}
		if _, err = c.kubeclientset.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{}); err != nil {
			utilruntime.HandleError(err)
			return err
		}
	}
	// if both get and create failed, return err
	if err != nil {
		utilruntime.HandleError(err)
		return err
	}
	// 2. check helm release exists
	actionStatus := action.NewStatus(actionConfig)
	_, err = actionStatus.Run(instance.GetName())
	if err != nil {
		if err == driver.ErrReleaseNotFound {
			actionInstall := action.NewInstall(actionConfig)
			actionInstall.ReleaseName = instance.GetName()
			actionInstall.Namespace = instance.GetName()
			argocdChart, err := loader.Load("charts/bcs-argocd")
			if err != nil {
				blog.Errorf("load argocd chart failed: %v", err)
				return err
			}
			_, err = actionInstall.Run(argocdChart, make(map[string]interface{}))
		} else {
			// if both get and install failed, return err
			utilruntime.HandleError(err)
			return err
		}
	}
	// 3. check service exists
	service, err := c.serviceLister.Services(instanceName).Get("argocd-server")
	if err != nil {
		utilruntime.HandleError(err)
		return err
	}
	// 4. set service host to ArgocdInstance.Status.Service
	if err := c.updateArgocdInstanceStatus(instance, service); err != nil {
		utilruntime.HandleError(err)
		return err
	}
	c.recorder.Event(instance, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

// updateDatabaseManagerStatus update ArgocdInstance status
func (c *InstanceController) updateArgocdInstanceStatus(instance *tkexv1alpha1.ArgocdInstance, service *corev1.Service) error {
	instanceCopy := instance.DeepCopy()
	blog.Info("service.Spec.ClusterIP: %s", service.Spec.ClusterIP)
	instanceCopy.Status.ServerHost = service.Spec.ClusterIP
	updated, err := c.tkexclientset.TkexV1alpha1().ArgocdInstances(common.ArgocdManagerNamespace).UpdateStatus(context.TODO(), instanceCopy, metav1.UpdateOptions{})
	if err != nil {
		utilruntime.HandleError(err)
		return err
	}
	blog.Info("updated.Status.ServerHost: %s", updated.Status.ServerHost)
	return nil
}

// cache object and enqueue key
func (c *InstanceController) enqueueArgocdInstance(obj interface{}) {
	var key string
	var err error
	// cache object
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	// enqueue key
	c.workqueue.AddRateLimited(key)
}

// delete object cache and enqueue key
func (c *InstanceController) enqueueArgocdInstanceForDelete(obj interface{}) {
	var key string
	var err error
	// delete object from cache
	key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	// enqueue key
	c.workqueue.AddRateLimited(key)
}
