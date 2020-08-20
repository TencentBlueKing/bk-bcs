package bkdataapi

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/esb/bkdata"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/apis/bkbcs.tencent.com/v1"
	internalclientset "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/generated/clientset/versioned"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/generated/informers/externalversions"
	listers "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/generated/listers/bkbcs.tencent.com/v1"
)

type BKDataController struct {
	stopCh                   chan struct{}
	apiextensionClientset    *apiextensionsclient.Clientset
	bkDataApiConfigClientset *internalclientset.Clientset
	bkDataApiConfigInformer  cache.SharedIndexInformer
	bkDataApiConfigLister    listers.BKDataApiConfigLister
	kubeConfig               string
}

func NewBKDataController(kubeConfig string) *BKDataController {
	return &BKDataController{
		kubeConfig: kubeConfig,
	}
}

func (c *BKDataController) Start() error {
	err := c.initKubeConfig()
	if err != nil {
		blog.Errorf("Initialization of LogController failed: %s", err.Error())
		return err
	}
	return nil
}

func (c *BKDataController) initKubeConfig() error {
	var restConf *rest.Config
	var err error
	if c.kubeConfig != "" {
		restConf, err = clientcmd.BuildConfigFromFlags("", c.kubeConfig)
	} else {
		restConf, err = rest.InClusterConfig()
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
	internalFactory := externalversions.NewSharedInformerFactory(c.bkDataApiConfigClientset, time.Hour)
	c.bkDataApiConfigInformer = internalFactory.Bkbcs().V1().BKDataApiConfigs().Informer()
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
	blog.Infof("build BKDataApiConfigClientset for config %s success", c.kubeConfig)
	return nil
}

func (c *BKDataController) createBKDataApiConfig() error {
	bkDataApiConfigPlural := "bkdataapiconfigs"
	bkDataApiConfigFullName := "bkdataapiconfigs" + "." + bcsv1.SchemeGroupVersion.Group
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: bkDataApiConfigFullName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   bcsv1.SchemeGroupVersion.Group,   // BKDataApiConfigsGroup,
			Version: bcsv1.SchemeGroupVersion.Version, // BKDataApiConfigsVersion,
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

func (c *BKDataController) handleAddBKDataApiConfig(obj interface{}) {
	// get BKDataClientConfig from crd
	conf, ok := obj.(*bcsv1.BKDataApiConfig)
	if !ok {
		blog.Errorf("Conver new object to BKDataApiConfig struct failed")
		return
	}
	bkDataApiConfig := conf.DeepCopy()
	// get api method
	switch bkDataApiConfig.Spec.ApiName {
	case "v3_access_deploy_plan_post":
		client, err := bkdata.NewBKDataClient(bkdata.BKDataClientConfig{
			BkAppCode:   bkDataApiConfig.Spec.AccessDeployPlanConfig.BkAppCode,
			BkAppSecret: bkDataApiConfig.Spec.AccessDeployPlanConfig.BkAppSecret,
			BkUsername:  bkDataApiConfig.Spec.AccessDeployPlanConfig.BkUsername,
		})
		if err != nil {
			blog.Errorf("Initializing bkdata api client error: %s", err.Error())
			c.respondFailed(bkDataApiConfig, err)
			break
		}
		dataid, err := client.ObtainDataId(bkDataApiConfig.Spec.AccessDeployPlanConfig)
		if err != nil {
			blog.Errorf("Application for dataid failed: %s", err)
			c.respondFailed(bkDataApiConfig, err)
			break
		}
		jsonstr, err := json.Marshal(map[string]interface{}{
			"dataid": dataid,
		})
		if err != nil {
			blog.Errorf("Convert dataid struct to jsonstr error: %s", err.Error())
			c.respondFailed(bkDataApiConfig, err)
			break
		}
		c.respondOK(bkDataApiConfig, string(jsonstr))
	case "v3_databus_cleans_post":
		client, err := bkdata.NewBKDataClient(bkdata.BKDataClientConfig{
			BkAppCode:   bkDataApiConfig.Spec.DataCleanStrategyConfig.BkAppCode,
			BkAppSecret: bkDataApiConfig.Spec.DataCleanStrategyConfig.BkAppSecret,
			BkUsername:  bkDataApiConfig.Spec.DataCleanStrategyConfig.BkUsername,
		})
		if err != nil {
			blog.Errorf("Initializing bkdata api client error: %s", err.Error())
			c.respondFailed(bkDataApiConfig, err)
			break
		}
		err = client.SetCleanStrategy(bkDataApiConfig.Spec.DataCleanStrategyConfig)
		if err != nil {
			blog.Errorf("Application for dataid failed: %s", err)
			c.respondFailed(bkDataApiConfig, err)
			break
		}
		if err != nil {
			blog.Errorf("Convert dataid struct to jsonstr error: %s", err.Error())
			c.respondFailed(bkDataApiConfig, err)
			break
		}
		c.respondOK(bkDataApiConfig, "")
	default:
		c.respondFailed(bkDataApiConfig, fmt.Errorf("Invalid API name"))
	}
}

func (c *BKDataController) handleUpdatedBKDataApiConfig(oldobj, newobj interface{}) {
	//TODO
}

func (c *BKDataController) respondFailed(conf *bcsv1.BKDataApiConfig, err error) {
	errs, _ := json.Marshal([]string{err.Error()})
	resp := bcsv1.BKDataApiResponse{
		Result:  false,
		Errors:  string(errs),
		Message: err.Error(),
	}
	conf.Spec.Response = resp
	_, err = c.bkDataApiConfigClientset.BkbcsV1().BKDataApiConfigs(conf.GetNamespace()).Update(conf)
	if err != nil {
		blog.Errorf("Update BKDataApiConfig failed: %s, crd info: %+v", err.Error(), *conf)
	}
}

func (c *BKDataController) respondOK(conf *bcsv1.BKDataApiConfig, retstr string) {
	resp := bcsv1.BKDataApiResponse{
		Result:  true,
		Message: "success",
		Data:    retstr,
	}
	conf.Spec.Response = resp
	_, err := c.bkDataApiConfigClientset.BkbcsV1().BKDataApiConfigs(conf.GetNamespace()).Update(conf)
	if err != nil {
		blog.Errorf("Update BKDataApiConfig failed: %s, crd info: %+v", err.Error(), *conf)
	}
}
