package bkdataapi

import (
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/bkbcs.tencent.com/v1"
	internalclientset "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/clientset/versioned"
	"github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/informers/externalversions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type BKDataController struct {
	stopCh chan struct{}
	apiextensionClientset
	clientset
	informer
	lister
	kubeConfig
}

func NewBKDataController(kubeConfig string) *BKDataController {
	return &BKDataController{
		kubeConfig: kubeConfig,
	}
}

func (c *BKDataController) Start() {
	err := c.initKubeConfig()
	if err != nil {
		blog.Errorf("Initialization of LogController of Cluster %s failed: %s", c.clusterInfo.ClusterID, err.Error())
		return
	}
	go c.run()
}

func (c *BKDataController) initKubeConf() error {
	if c.KubeConfig != "" {
		restConf, err := clientcmd.BuildConfigFromFlags("", c.kubeConfig)
	} else {
		restConf, err := rest.InClusterConfig()
	}
	if err != nil {
		blog.Errorf("build kubeconfig %s error :%s", c.kubeConfig, err.Error())
		return err
	}
	c.apiextensionClientset, err = apiextensionsclient.NewForConfig(restConf)
	if err != nil {
		blog.Errorf("build apiextension client by kubeconfig % error %s", c.kubeConfig, err.Error())
		return err
	}
	err = c.createBKDataApiConfig()
	if err != nil {
		return err
	}

	//internal clientset for informer BKDataApiConfig Crd
	c.bkDataApiConfigClientset, err = internalclientset.NewForConfig(restConf)
	if err != nil {
		blog.Errorf("build BKDataApiConfig clientset by kubeconfig %s error %s", c.kubeConfig, err.Error())
		return err
	}
	internalFactory := externalversions.NewSharedInformerFactory(bkDataApiConfigClientset, time.Hour)
	c.bkDataApiConfigInformer = internalFactory.Bkbcs().V1().BcsLogConfigs().Informer()
	internalFactory.Start(c.stopCh)
	// Wait for all caches to sync.
	internalFactory.WaitForCacheSync(c.stopCh)
	//add k8s resources event handler functions
	c.bkDataApiConfigInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAddBKDataApiConfig,
			UpdateFunc: c.handleUpdatedBKDataApiConfig,
		},
	)
	blog.Infof("build BKDataApiConfigClientset for config %s success", s.conf.kubeConfig)
	return nil
}

func (c *BKDataController) createBKDataApiConfig() error {
	bkDataApiConfigPlural := "bkdataapiconfigs"
	bkDataApiConfigFullName := "bkdataapiconfigs" + "." + bcsv1.SchemeGroupVersion.Group
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: bcsLogConfigFullName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   bcsv1.SchemeGroupVersion.Group,   // BcsLogConfigsGroup,
			Version: bcsv1.SchemeGroupVersion.Version, // BcsLogConfigsVersion,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   bkDataApiConfigPlural,
				Kind:     reflect.TypeOf(bcsv1.BKDataApiConfig{}).Name(),
				ListKind: reflect.TypeOf(bcsv1.BKDataApiConfigList{}).Name(),
			},
		},
	}

	_, err := c.apiextensionClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			blog.Infof("BKDataApiConfig Crd is already exists")
			return nil
		}
		blog.Errorf("create BKDataApiConfig Crd error %s", err.Error())
		return err
	}
	blog.Infof("create BKDataApiConfig Crd success")
	return nil
}

func (c *BKDataController) handleAddBKDataApiConfig() {
	// get BKDataClientConfig from crd

	// get api method

	// verify data

	// requestdata

	// set response
}

func (c *BKDataController) handleUpdatedBKDataApiConfig() {
	// get BKDataApiConfig status

	// if resolved, then delete it

	// else keep status
}
