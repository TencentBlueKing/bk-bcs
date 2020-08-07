package controllers

import (
	"fmt"
	"io/ioutil"
	"context"
	"reflect"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/helmclient"
	meshv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/api/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/config"

	"k8s.io/klog"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/informers"
	corev1 "k8s.io/client-go/listers/core/v1"
	appsv1 "k8s.io/client-go/listers/apps/v1"
	restclient "k8s.io/client-go/rest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kubeclient "github.com/kubernetes-client/go/kubernetes/client"
	"k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/api/errors"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

var (
	MeshComponents = []string{"istio-operator","istiod","istio-egressgateway","istio-ingressgateway","istio-tracing","kiali"}
)

type MeshClusterManager struct {
	sync.RWMutex
	meshCluster *meshv1.MeshCluster
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
	//mesh components list
	componentStatus map[string]*meshv1.InstallStatus_VersionStatus
}

func NewMeshClusterManager(conf config.Config, meshCluster *meshv1.MeshCluster)(*MeshClusterManager,error){
	m := &MeshClusterManager{
		meshCluster: meshCluster,
		componentStatus: make(map[string]*meshv1.InstallStatus_VersionStatus),
	}
	//init mesh components status
	for _,component :=range MeshComponents {
		status := &meshv1.InstallStatus_VersionStatus{
			Name: component,
			Namespace: "istio-system",
			Status: meshv1.InstallStatus_NONE,
		}
		if component=="istio-operator" {
			status.Namespace = "istio-operator"
		}
		m.componentStatus[component] = status
	}
	m.kubeconfig = &restclient.Config{
		Host: conf.ServerAddress,
		BearerToken: conf.UserToken,
		QPS: 1e3,
		Burst: 2e3,
	}

	stopCh := make(chan struct{})
	//kubernetes clientset
	kubeClient, err := kubernetes.NewForConfig(m.kubeconfig)
	if err != nil {
		klog.Errorf("build kubeclient by kubeconfig %s error %s", m.kubeconfig, err.Error())
		return nil, err
	}
	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	m.podLister = factory.Core().V1().Pods().Lister()
	m.deploymentLister = factory.Apps().V1().Deployments().Lister()
	factory.Start(stopCh)
	// Wait for all caches to sync.
	factory.WaitForCacheSync(stopCh)
	klog.Infof("build kubeclient for config %s success", m.kubeconfig)

	//apiextensions clientset for creating IstioOperator、MeshCluster Crd
	m.extensionClientset, err = apiextensionsclient.NewForConfig(m.kubeconfig)
	if err != nil {
		klog.Errorf("build apiextension client by kubeconfig % error %s", m.kubeconfig, err.Error())
		return nil, err
	}
	//create MeshCluster Crd in kube-apiserver
	err = m.createMeshClusterCrd()
	if err!=nil {
		return nil, err
	}

	//kubernetes api client for create IstioOperator Object
	cfg := kubeclient.NewConfiguration()
	cfg.Host = m.conf.ServerAddress
	cfg.DefaultHeader["authorization"] = fmt.Sprintf("Bearer %s", m.conf.UserToken)
	m.kubeApiClient = kubeclient.NewAPIClient(cfg)
	klog.Infof("build kubeapiclient for config %s success", m.kubeconfig)
	klog.Infof("New MeshClusterManager(%s) success", meshCluster.GetUuid())
	return m, nil
}

func (m *MeshClusterManager) uninstallIstio(){
	//TODO
}

func (m *MeshClusterManager) installIstio()error{
	//helm chart install IstioOperator
	setParam := map[string]string{
		"hub": m.conf.DockerHub,
		"tag": m.meshCluster.Spec.Version,
	}
	err := m.helm.InstallChart(setParam)
	if err!=nil {
		klog.Errorf("Install cluster(%s) istio-operator failed: %s", m.meshCluster.Spec.ClusterId, err.Error())
		return err
	}
	klog.Infof("Install cluster(%s) istio-operator done", m.meshCluster.Spec.ClusterId)

	//create IstioOperator Crds
	return m.createIstioOperatorCrds()
}

func (m *MeshClusterManager) meshComponentStatus()map[string]*meshv1.InstallStatus_VersionStatus{
	for _,cStatus :=range m.componentStatus {
		m.updateComponentStatus(cStatus)
	}

	return m.componentStatus
}

func (m *MeshClusterManager) updateComponentStatus(status *meshv1.InstallStatus_VersionStatus){
	klog.Infof("MeshClusterManager start component(%s) status", status.Name)
	deployment,err := m.deploymentLister.Deployments(status.Namespace).Get(status.Name)
	if err!=nil {
		if errors.IsNotFound(err){
			klog.Infof("Mesh Component(%s:%s) is NotFound", status.Namespace, status.Name)
			status.Status = meshv1.InstallStatus_NONE
			return
		}
		klog.Errorf("Mesh Component(%s:%s) Get Deployment failed: %s", status.Namespace, status.Name, err.Error())
		return
	}
	status.Message = deployment.Status.String()
	//deployment is deploying pods now
	if deployment.Status.Replicas<*deployment.Spec.Replicas {
		klog.Infof("Mesh Component(%s:%s) Spec.Replicas(%d) Status.Replicas(%d)", status.Namespace, status.Name,
			*deployment.Spec.Replicas, deployment.Status.Replicas)
		status.Status = meshv1.InstallStatus_DEPLOY
		return
	}
	//deployment is updating pods now
	if deployment.Status.Replicas>deployment.Status.UpdatedReplicas {
		klog.Infof("Mesh Component(%s:%s) Status.Replicas(%d) Status.UpdatedReplicas(%d)", status.Namespace, status.Name,
			deployment.Status.Replicas, deployment.Status.UpdatedReplicas)
		status.Status = meshv1.InstallStatus_UPDATE
		return
	}
	//deployment is starting pods now
	if deployment.Status.Replicas>deployment.Status.AvailableReplicas {
		klog.Infof("Mesh Component(%s:%s) Status.Replicas(%d) Status.AvailableReplicas(%d)", status.Namespace, status.Name,
			deployment.Status.Replicas, deployment.Status.AvailableReplicas)
		status.Status = meshv1.InstallStatus_STARTING
		return
	}

	//deployment is ready now
	if deployment.Status.Replicas==deployment.Status.AvailableReplicas {
		klog.Infof("Mesh Component(%s:%s) Status.Replicas(%d) Status.AvailableReplicas(%d)", status.Namespace, status.Name,
			deployment.Status.Replicas, deployment.Status.AvailableReplicas)
		status.Status = meshv1.InstallStatus_RUNNING
		return
	}
	//deployment have failed pods now
	if deployment.Status.UnavailableReplicas>0 {
		klog.Infof("Mesh Component(%s:%s) Status.Replicas(%d) Status.UnavailableReplicas(%d)", status.Namespace, status.Name,
			deployment.Status.Replicas, deployment.Status.AvailableReplicas)
		status.Status = meshv1.InstallStatus_FAILED
		return
	}
}

// create crd of istiooperator
func (m *MeshClusterManager) createIstioOperatorCrds() error {
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

// create crd of MeshCluster
func (m *MeshClusterManager) createMeshClusterCrd() error {
	meshclusterPlural := "meshclusters"
	meshclusterFullName := "meshclusters" + "." + meshv1.GroupVersion.Group
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: meshclusterFullName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   meshv1.GroupVersion.Group,   // BcsLogConfigsGroup,
			Version: meshv1.GroupVersion.Version, // BcsLogConfigsVersion,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   meshclusterPlural,
				Kind:     reflect.TypeOf(meshv1.MeshCluster{}).Name(),
				ListKind: reflect.TypeOf(meshv1.MeshClusterList{}).Name(),
			},
		},
	}

	_, err := m.extensionClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			klog.Infof("MeshCluster Crd is already exists")
			return nil
		}
		klog.Errorf("create MeshCluster Crd error %s", err.Error())
		return err
	}
	klog.Infof("create MeshCluster Crd success")
	return nil
}