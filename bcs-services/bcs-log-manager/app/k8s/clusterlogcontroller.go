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
	bcsv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/pkg/apis/bk-bcs/v1"
	internalclientset "github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/pkg/client/clientset/versioned"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/pkg/client/informers/externalversions"
	bkbcsv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/pkg/client/listers/bk-bcs/v1"
)

type ClusterLogController struct {
	clusterInfo *bcsapi.ClusterCredential
	caFile      string
	// TODO: task info
	tick                 int64
	extensionClientset   *apiextensionsclient.Clientset
	bcsLogConfigLister   bkbcsv1.BcsLogConfigLister
	bcsLogConfigInformer cache.SharedIndexInformer
	bcsClientset         *internalclientset.Clientset
	stopCh               chan struct{}
}

func NewClusterLogController(conf *config.ControllerConfig) (*ClusterLogController, error) {
	ctlr := &ClusterLogController{
		clusterInfo: conf.Credential,
		tick:        0,
		caFile:      conf.CAFile,
	}
	err := ctlr.initKubeConf()
	if err != nil {
		blog.Errorf("Initialization of LogController of Cluster %s failed: %s", ctlr.clusterInfo.ClusterID, err.Error())
		return nil, fmt.Errorf("Initialization of LogController of Cluster %s failed: %s", ctlr.clusterInfo.ClusterID, err.Error())
	}
	return ctlr, nil
}

func (c *ClusterLogController) Start() {
	err := c.initKubeConf()
	if err != nil {
		blog.Errorf("Initialization of LogController of Cluster %s failed: %s", c.clusterInfo.ClusterID, err.Error())
		return
	}
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
		// TODO informer inform functions
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
