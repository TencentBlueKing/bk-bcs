package k8s

import (
	"fmt"
	"reflect"
	"strings"
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
	DefaultLogConfigNamespace = "default"
)

var LogConfigAPIVersion string
var LogConfigKind string

type ClusterLogController struct {
	AddCollectionTask    chan *config.CollectionConfig
	DeleteCollectionTask chan string
	UpdateCollectionTask chan *config.CollectionConfig

	clusterInfo          *bcsapi.ClusterCredential
	caFile               string
	collectionTasks      map[string]*config.CollectionConfig
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

func NewClusterLogController(conf *config.ControllerConfig) (*ClusterLogController, error) {
	ctlr := &ClusterLogController{
		clusterInfo:          conf.Credential,
		tick:                 0,
		caFile:               conf.CAFile,
		collectionTasks:      make(map[string]*config.CollectionConfig),
		AddCollectionTask:    make(chan *config.CollectionConfig),
		DeleteCollectionTask: make(chan string),
		UpdateCollectionTask: make(chan *config.CollectionConfig),
	}
	err := ctlr.initKubeConf()
	if err != nil {
		blog.Errorf("Initialization of LogController of Cluster %s failed: %s", ctlr.clusterInfo.ClusterID, err.Error())
		return nil, fmt.Errorf("Initialization of LogController of Cluster %s failed: %s", ctlr.clusterInfo.ClusterID, err.Error())
	}
	return ctlr, nil
}

func BuildBcsLogConfigKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func BuildDefaultBcsLogConfigName(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func (c *ClusterLogController) Start() {
	go c.run()
}

func (c *ClusterLogController) Stop() {
	close(c.stopCh)
}

func (c *ClusterLogController) SetTick(tick int64) {
	c.tick = tick
}

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
		restConf.CAFile = c.caFile

		c.extensionClientset, err = apiextensionsclient.NewForConfig(restConf)
		if err != nil {
			blog.Errorf("APIExtensionClientset initialization failed: server %s, cluster %s, %s", url, c.clusterInfo.ClusterID, err.Error())
			continue
		}
		err = c.createBcsLogConfig()
		if err != nil {
			continue
		}

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
				blog.Errorf("Stop channel closed, stop working")
				return
			}
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
			logconf.Spec = task.ConfigSpec
			logconf.Spec.ClusterId = c.clusterInfo.ClusterID
			_, err := c.bcsClientset.Bkbcs().BcsLogConfigs(task.ConfigNamespace).Create(logconf)
			if err != nil {
				blog.Errorf("Create BcsLogConfig of Cluster %s failed: %s (config info: %+v)", c.clusterInfo.ClusterID, err.Error(), logconf)
				break
			}
			c.collectionTasks[BuildBcsLogConfigKey(task.ConfigNamespace, task.ConfigName)] = task
			blog.Infof("Create BcsLogConfig of Cluster %s success. (config info: %+v)", c.clusterInfo.ClusterID, logconf)
		case key, ok := <-c.DeleteCollectionTask:
			if !ok {
				blog.Errorf("DeleteCollectionTask chan of cluster %s has been closed", c.clusterInfo.ClusterID)
				return
			}
			var configName, configNamespace string
			names := strings.Split(key, "/")
			if len(names) == 1 {
				configNamespace = DefaultLogConfigNamespace
				configName = names[0]
			} else if len(names) == 2 {
				configNamespace = names[0]
				configName = names[1]
			} else {
				blog.Errorf("DeleteCollectionTask chan receive an error bcslogconfig key %s, want {namespace}/{name} or {name} in default namespace", key)
			}
			err := c.bcsClientset.Bkbcs().BcsLogConfigs(configNamespace).Delete(configName, nil)
			if err != nil {
				blog.Errorf("Delete BcsLogConfig (%s) of Cluster %s failed: %s", key, c.clusterInfo.ClusterID, err.Error())
				break
			}
			delete(c.collectionTasks, key)
			blog.Infof("Delete BcsLogConfig (%s) of Cluster %s success.", key, c.clusterInfo.ClusterID)
		}
	}
}

func (c *ClusterLogController) handleAddTask(obj interface{}) {
	// TODO
}

func (c *ClusterLogController) handleUpdateTask(oldObj interface{}, newObj interface{}) {
}

func (c *ClusterLogController) handleDeleteTask(obj interface{}) {
}
