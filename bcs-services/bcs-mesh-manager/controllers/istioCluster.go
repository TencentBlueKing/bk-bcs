package controllers

import (
	"io/ioutil"
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/helmclient"
	meshv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/api/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/config"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/klog"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/client-go/listers/core/v1"
	appsv1 "k8s.io/client-go/listers/apps/v1"
	restclient "k8s.io/client-go/rest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclient "github.com/kubernetes-client/go/kubernetes/client"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

type istioClusterManager struct {
	istioCluster *meshv1.IstioCluster
	kubeconfig *restclient.Config
	//config
	conf *config.Config
	//pod Lister
	podLister corev1.PodLister
	//deployment Lister
	deploymentLister appsv1.DeploymentLister
	//apiextensions clientset
	extensionClientset *apiextensionsclient.Clientset
	//kubernetes api client
	kubeApiClient *kubeclient.APIClient
	//helm client
	helm helmclient.HelmClient
	//component list
	installedComponents []string
}

func (m *istioClusterManager) uninstallIstio(){
	//TODO
}

func (m *istioClusterManager) installIstio()error{
	//helm chart install IstioOperator
	setParam := map[string]string{
		"hub": m.conf.DockerHub,
		"tag": m.istioCluster.Spec.Version,
	}
	err := m.helm.InstallChart(setParam)
	if err!=nil {
		klog.Errorf("Install cluster(%s) istio-operator failed: %s", m.istioCluster.Spec.ClusterId, err.Error())
		return err
	}
	klog.Infof("Install cluster(%s) istio-operator done", m.istioCluster.Spec.ClusterId)

	//create IstioOperator Crds
	return m.createIstioOperatorCrds()
}

// create crd of istiooperator
func (m *istioClusterManager) createIstioOperatorCrds() error {
	//unmarshal IstioOperator Object from yaml file
	var crd *apiextensionsv1beta1.CustomResourceDefinition
	body, err := ioutil.ReadFile(m.conf.IstioOperatorCrdFile)
	if err!=nil {
		klog.Errorf("Read IstioOperatorCrd file(%s) failed: %s", m.conf.IstioOperatorCrdFile, err.Error())
		return err
	}
	klog.Infof("Read IstioOperatorCrd file(%s) body(%s)", m.conf.IstioOperatorCrdFile, string(body))
	err = yaml.Unmarshal(body, &crd)
	if err!=nil {
		klog.Errorf("Read IstioOperatorCrd file(%s) failed: %s", m.conf.IstioOperatorCrdFile, err.Error())
		return err
	}
	//create IstioOperator Crd
	_, err = m.extensionClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			klog.Infof("IstioOperator Crd is already exists")
			return nil
		}
		klog.Errorf("create IstioOperator Crd error %s", err.Error())
		return err
	}
	klog.Infof("create IstioOperator Crd success")

	//get IstioOperator Body from file
	body, err = ioutil.ReadFile(m.conf.IstioOperatorCrFile)
	if err!=nil {
		klog.Errorf("Read IstioOperatorCr file(%s) failed: %s", m.conf.IstioOperatorCrFile, err.Error())
		return err
	}
	klog.Infof("Read IstioOperatorCr file(%s) body(%s)", m.conf.IstioOperatorCrFile, string(body))
	//unmarshal ObjectMeta, to get field Namespace
	var metaData metav1.ObjectMeta
	err = yaml.Unmarshal(body, &metaData)
	if err!=nil {
		klog.Errorf("Read metav1.ObjectMeta file(%s) failed: %s", m.conf.IstioOperatorCrFile, err.Error())
		return err
	}
	//create IstioOperator Cr Object
	_,_,err = m.kubeApiClient.CustomObjectsApi.CreateNamespacedCustomObject(context.Background(), crd.Spec.Group,
		crd.Spec.Versions[0].Name, metaData.Namespace, crd.Spec.Names.Plural, body, nil)
	if err!=nil {
		klog.Errorf("create IstioOperator Cr Object error %s", err.Error())
		return err
	}
	klog.Infof("create IstioOperator Cr Object success")
	return nil
}