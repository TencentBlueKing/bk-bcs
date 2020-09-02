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
	"reflect"
	"strings"
	"sync"
	"time"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/util"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/pkg/apis/bk-bcs/v1"
	internalclientset "github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/pkg/client/clientset/versioned"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/pkg/client/informers/externalversions"
	bkbcsv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/pkg/client/listers/bk-bcs/v1"
)

const (
	// DefaultLogConfigNamespace is default namespace for bcslogconfigs CRD
	DefaultLogConfigNamespace = "default"
)

// LogConfigAPIVersion is api version of bcslogconfigs
var LogConfigAPIVersion string

// LogConfigKind is crd name of bcslogconfigs
var LogConfigKind string

// ClusterLogController is controller for single cluster of bcslogconfigs
type ClusterLogController struct {
	// AddCollectionTask is used to inform this instance to create bcslogconfigs
	AddCollectionTask chan config.CollectionConfig

	// DeleteCollectionTask is used to inform this instance to delete bcslogconfigs
	DeleteCollectionTask chan *config.CollectionFilterConfig

	// DeleteCollectionTask is used to inform this instance to update bcslogconfigs
	UpdateCollectionTask chan config.CollectionConfig

	clusterInfo          *bcsapi.ClusterCredential
	caFile               string
	collectionTasks      map[string]*config.CollectionConfig
	taskLock             sync.Mutex
	tick                 int64
	extensionClientset   *apiextensionsclient.Clientset
	bcsLogConfigLister   bkbcsv1.BcsLogConfigLister
	bcsLogConfigInformer cache.SharedIndexInformer
	bcsClientset         *internalclientset.Clientset
	stopCh               chan struct{}
}

func init() {
	LogConfigAPIVersion = fmt.Sprintf("%s/%s", bcsv1.SchemeGroupVersion.Group, bcsv1.SchemeGroupVersion.Version)
	LogConfigKind = reflect.TypeOf(bcsv1.BcsLogConfig{}).Name()
}

// NewClusterLogController create ClusterLogController instance
func NewClusterLogController(conf *config.ControllerConfig) (*ClusterLogController, error) {
	ctlr := &ClusterLogController{
		clusterInfo:          conf.Credential,
		tick:                 0,
		caFile:               conf.CAFile,
		collectionTasks:      make(map[string]*config.CollectionConfig),
		AddCollectionTask:    make(chan config.CollectionConfig),
		DeleteCollectionTask: make(chan *config.CollectionFilterConfig),
		UpdateCollectionTask: make(chan config.CollectionConfig),
		stopCh:               make(chan struct{}),
	}
	err := ctlr.initKubeConf()
	if err != nil {
		blog.Errorf("Initialization of LogController of Cluster %s failed: %s", ctlr.clusterInfo.ClusterID, err.Error())
		return nil, fmt.Errorf("Initialization of LogController of Cluster %s failed: %s", ctlr.clusterInfo.ClusterID, err.Error())
	}
	return ctlr, nil
}

// BuildBcsLogConfigKey build namespace/name key string
func BuildBcsLogConfigKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

// Start start the controller
func (c *ClusterLogController) Start() {
	go c.run()
}

// Stop stop the controller
func (c *ClusterLogController) Stop() {
	close(c.stopCh)
}

// SetTick is used to judge whether this cluster has been destroyed since last syncing cluster info
func (c *ClusterLogController) SetTick(tick int64) {
	c.tick = tick
}

// GetTick GetTick
func (c *ClusterLogController) GetTick() int64 {
	return c.tick
}

func (c *ClusterLogController) initKubeConf() error {
	var err error
	urls := strings.Split(c.clusterInfo.ServerAddresses, ",")
	initFlag := false
	restConf := &rest.Config{}
	for _, url := range urls {
		restConf.Host = url
		restConf.BearerToken = c.clusterInfo.UserToken
		// restConf.CAFile = c.caFile
		// TODO tsl secure
		restConf.TLSClientConfig.Insecure = true
		// create CRD clientset
		c.extensionClientset, err = apiextensionsclient.NewForConfig(restConf)
		if err != nil {
			blog.Errorf("APIExtensionClientset initialization failed: server %s, cluster %s, %s", url, c.clusterInfo.ClusterID, err.Error())
			continue
		}
		// create bcslogconfigs CRD
		err = c.createBcsLogConfig()
		if err != nil {
			continue
		}

		// create bcslogconfigs clientset
		c.bcsClientset, err = internalclientset.NewForConfig(restConf)
		if err != nil {
			blog.Errorf("Clientset initialization failed: server %s, cluster %s, %s", url, c.clusterInfo.ClusterID, err.Error())
			continue
		}
		internalFactory := externalversions.NewSharedInformerFactory(c.bcsClientset, time.Hour)
		c.bcsLogConfigLister = internalFactory.Bkbcs().V1().BcsLogConfigs().Lister()
		c.bcsLogConfigInformer = internalFactory.Bkbcs().V1().BcsLogConfigs().Informer()
		c.bcsLogConfigInformer.AddEventHandler(
			cache.ResourceEventHandlerFuncs{
				AddFunc:    c.handleAddTask,
				UpdateFunc: c.handleUpdateTask,
				DeleteFunc: c.handleDeleteTask,
			},
		)
		internalFactory.Start(c.stopCh)
		internalFactory.WaitForCacheSync(c.stopCh)
		initFlag = true
	}
	if !initFlag {
		return fmt.Errorf("Cannot access kubeapi server of cluster %s, please check config file first", c.clusterInfo.ClusterID)
	}
	blog.Info("Build clientset of cluster %s success", c.clusterInfo.ClusterID)
	return nil
}

// create crd of BcsLogConf
func (c *ClusterLogController) createBcsLogConfig() error {
	bcsLogConfigPlural := "bcslogconfigs"
	bcsLogConfigFullName := "bcslogconfigs" + "." + bcsv1.SchemeGroupVersion.Group
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: bcsLogConfigFullName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   bcsv1.SchemeGroupVersion.Group,   // BcsLogConfigsGroup,
			Version: bcsv1.SchemeGroupVersion.Version, // BcsLogConfigsVersion,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   bcsLogConfigPlural,
				Kind:     reflect.TypeOf(bcsv1.BcsLogConfig{}).Name(),
				ListKind: reflect.TypeOf(bcsv1.BcsLogConfigList{}).Name(),
			},
		},
	}

	_, err := c.extensionClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			blog.Infof("BcsLogConfig Crd is already exists")
			return nil
		}
		blog.Errorf("create BcsLogConfig Crd error %s", err.Error())
		return err
	}
	blog.Infof("create BcsLogConfig Crd success")
	return nil
}

func (c *ClusterLogController) run() {
	blog.Infof("Controller of cluster %s start working", c.clusterInfo.ClusterID)
	for {
		select {
		case _, ok := <-c.stopCh:
			if !ok {
				blog.Errorf("Stop channel closed, cluster controller of %s stop working", c.clusterInfo.ClusterID)
				return
			}
		// create new bcslogconfigs CRD
		case task, ok := <-c.AddCollectionTask:
			if !ok {
				blog.Errorf("AddCollectionTask chan of cluster %s has been closed", c.clusterInfo.ClusterID)
				return
			}
			blog.Infof("Receive new collectiontask: %+v", task)
			logconf := &bcsv1.BcsLogConfig{}
			logconf.TypeMeta.Kind = LogConfigKind
			logconf.TypeMeta.APIVersion = LogConfigAPIVersion
			if task.ConfigName == "" {
				task.ConfigName = fmt.Sprintf("%s-%s-%d", LogConfigKind, c.clusterInfo.ClusterID, util.GenerateID())
			}
			logconf.ObjectMeta.Name = task.ConfigName
			logconf.SetName(task.ConfigName)
			if task.ConfigNamespace == "" {
				task.ConfigNamespace = DefaultLogConfigNamespace
			}
			logconf.ObjectMeta.Namespace = task.ConfigNamespace
			task.ConfigSpec.ClusterId = c.clusterInfo.ClusterID
			logconf.Spec = task.ConfigSpec
			_, err := c.bcsClientset.Bkbcs().BcsLogConfigs(task.ConfigNamespace).Create(logconf)
			if err != nil {
				blog.Warnf("Create BcsLogConfig of Cluster %s failed: %s (config info: %+v)", c.clusterInfo.ClusterID, err.Error(), logconf)
				break
			}
			c.taskLock.Lock()
			c.collectionTasks[BuildBcsLogConfigKey(task.ConfigNamespace, task.ConfigName)] = &task
			c.taskLock.Unlock()
			blog.Infof("Create BcsLogConfig of Cluster %s success. (config info: %+v)", c.clusterInfo.ClusterID, logconf)
		case conf, ok := <-c.DeleteCollectionTask:
			if !ok {
				blog.Errorf("DeleteCollectionTask chan of cluster %s has been closed", c.clusterInfo.ClusterID)
				return
			}
			blog.Infof("Receive delete collection task: %+v", conf)
			tasks := make(map[string]*config.CollectionConfig)
			// extract matched configs
			c.taskLock.Lock()
			for key, task := range c.collectionTasks {
				if task.ConfigName == conf.ConfigName || conf.ConfigName == "" {
					if task.ConfigNamespace == conf.ConfigNamespace || conf.ConfigNamespace == "" {
						tasks[key] = task
					}
				}
			}
			c.taskLock.Unlock()
			blog.Infof("deleted tasks: %+v", tasks)
			// delete configs
			for key, task := range tasks {
				err := c.bcsClientset.Bkbcs().BcsLogConfigs(task.ConfigNamespace).Delete(task.ConfigName, nil)
				if err != nil {
					blog.Errorf("Delete BcsLogConfig (%s) of Cluster %s failed: %s", key, c.clusterInfo.ClusterID, err.Error())
					delete(tasks, key)
					continue
				}
				blog.Infof("Delete BcsLogConfig (%s) of Cluster %s success.", key, c.clusterInfo.ClusterID)
			}
			// delete configs from config map
			c.taskLock.Lock()
			for key, _ := range tasks {
				delete(c.collectionTasks, key)
			}
			c.taskLock.Unlock()
		}
	}
}

func (c *ClusterLogController) getLogCollectionTaskByFilter(filter *config.CollectionFilterConfig) []config.CollectionConfig {
	c.taskLock.Lock()
	ret := make([]config.CollectionConfig, 0, len(c.collectionTasks))
	for _, task := range c.collectionTasks {
		if task.ConfigName == filter.ConfigName || filter.ConfigName == "" {
			if task.ConfigNamespace == filter.ConfigNamespace || filter.ConfigNamespace == "" {
				ret = append(ret, *task)
			}
		}
	}
	c.taskLock.Unlock()
	return ret
}

func (c *ClusterLogController) handleAddTask(obj interface{}) {
	conf, ok := obj.(*bcsv1.BcsLogConfig)
	if !ok {
		blog.Errorf("Parse obj to *BcsLogConfig failed: obj(%+v)", obj)
		return
	}
	blog.Debug("Handle bcslogconfigs crd created: %+v", conf)
	task := &config.CollectionConfig{
		ConfigName:      conf.GetName(),
		ConfigNamespace: conf.GetNamespace(),
		ClusterIDs:      conf.Spec.ClusterId,
		ConfigSpec:      *conf.Spec.DeepCopy(),
	}
	key := BuildBcsLogConfigKey(task.ConfigNamespace, task.ConfigName)
	c.taskLock.Lock()
	if _, ok := c.collectionTasks[key]; !ok {
		c.collectionTasks[key] = task
		blog.Debug("Added bcslogconfigs to cluster controller's collectionTasks map, config value (%+v)", task)
	}
	blog.Debug("config already exist (%+v)", conf)
	c.taskLock.Unlock()
}

func (c *ClusterLogController) handleUpdateTask(oldObj interface{}, newObj interface{}) {
	oldConf, ok := oldObj.(*bcsv1.BcsLogConfig)
	if !ok {
		blog.Errorf("Parse obj to *BcsLogConfig failed: obj(%+v)", oldObj)
		return
	}
	newConf, ok := newObj.(*bcsv1.BcsLogConfig)
	if !ok {
		blog.Errorf("Parse obj to *BcsLogConfig failed: obj(%+v)", newObj)
		return
	}
	blog.Debug("Handle bcslogconfigs crd created: old(%+v), new(%+v)", oldConf, newConf)
	task := &config.CollectionConfig{
		ConfigName:      newConf.GetName(),
		ConfigNamespace: newConf.GetNamespace(),
		ClusterIDs:      newConf.Spec.ClusterId,
		ConfigSpec:      *newConf.Spec.DeepCopy(),
	}
	oldKey := BuildBcsLogConfigKey(oldConf.GetNamespace(), oldConf.GetName())
	newKey := BuildBcsLogConfigKey(newConf.GetNamespace(), newConf.GetName())
	c.taskLock.Lock()
	if _, ok := c.collectionTasks[oldKey]; ok {
		delete(c.collectionTasks, oldKey)
		blog.Debug("deleted old bcslogconfigs from cluster controller's collectionTasks map, config value (%+v)", oldConf)
	}
	c.collectionTasks[newKey] = task
	blog.Debug("add new bcslogconfigs to cluster controller's collectionTasks map, config value (%+v)", newConf)
	c.taskLock.Unlock()
}

func (c *ClusterLogController) handleDeleteTask(obj interface{}) {
	conf, ok := obj.(*bcsv1.BcsLogConfig)
	if !ok {
		blog.Errorf("Parse obj to *BcsLogConfig failed: obj(%+v)", obj)
		return
	}
	blog.Debug("Handle bcslogconfigs crd deleted: %+v", conf)
	key := BuildBcsLogConfigKey(conf.GetNamespace(), conf.GetName())
	c.taskLock.Lock()
	if _, ok := c.collectionTasks[key]; ok {
		delete(c.collectionTasks, key)
		blog.Debug("deleted bcslogconfigs to cluster controller's collectionTasks map, config value (%+v)", conf)
	}
	blog.Debug("config already deleted (%+v)", conf)
	c.taskLock.Unlock()
}
