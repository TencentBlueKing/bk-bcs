package controllers

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/helmclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/types"
	meshv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/api/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/config"

	kubeclient "github.com/kubernetes-client/go/kubernetes/client"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listersappsv1 "k8s.io/client-go/listers/apps/v1"
	listerscorev1 "k8s.io/client-go/listers/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog"
	"k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	podLister listerscorev1.PodLister
	//deployment Lister
	deploymentLister listersappsv1.DeploymentLister
	//MeshCluster Status Client
	meshClusterClient client.StatusClient

	//apiextensions clientset
	extensionClientset *apiextensionsclient.Clientset
	//kubernetes api client
	kubeApiClient *kubeclient.APIClient
	//helm client
	helm helmclient.HelmClient
}

func NewMeshClusterManager(conf config.Config, meshCluster *meshv1.MeshCluster)(*MeshClusterManager,error){
	m := &MeshClusterManager{
		meshCluster: meshCluster,
	}
	if m.meshCluster.Status.ComponentStatus==nil {
		m.meshCluster.Status.ComponentStatus = make(map[string]*meshv1.InstallStatus_VersionStatus)
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
			m.meshCluster.Status.ComponentStatus[component] = status
		}
	}
	//update mesh cluster components status
	m.updateComponentStatus()
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
	//
	return m, nil
}

//if uninstall istio done, then return true
//else return false
func (m *MeshClusterManager) uninstallIstio()bool{
	m.Lock()
	m.Unlock()

	//delete IstioOperator Crd
	_,_,err := m.kubeApiClient.CustomObjectsApi.DeleteNamespacedCustomObject(context.Background(), types.IstioOperatorGroup,
		types.IstioOperatorVersion, types.IstioOperatorNamespace, types.IstioOperatorPlural, types.IstioOperatorName, kubeclient.V1DeleteOptions{}, nil)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("Delete Cluster(%s) IstioOperator Crd error %s", m.meshCluster.Spec.ClusterId, err.Error())
		return false
	}

	//update Istio Components Status
	m.updateComponentStatus()
	//check Istio Components whether deleted
	for _,component :=range m.meshCluster.Status.ComponentStatus {
		if component.Name=="istio-operator" {
			continue
		}

		//if istio component not deleted, waiting
		if component.Status!=meshv1.InstallStatus_NONE {
			klog.Infof("Delete Cluster(%s) IstioMesh, and waiting component(%s:%s) deleted",
				m.meshCluster.Spec.ClusterId, component.Name, component.Status)
			return false
		}
	}
	//delete all resources in namespace istio-operator
	_,_,err = m.kubeApiClient.CoreV1Api.DeleteNamespace(context.Background(), "istio-operator",
	 	kubeclient.V1DeleteOptions{GracePeriodSeconds: 0}, nil)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("Delete Cluster(%s) Namespace(istio-operator) error %s", m.meshCluster.Spec.ClusterId, err.Error())
		return false
	}
	klog.Infof("Delete Cluster(%s) Namespace(istio-operator) success", m.meshCluster.Spec.ClusterId)

	//delete all resources in namespace istio-system
	_,_,err = m.kubeApiClient.CoreV1Api.DeleteNamespace(context.Background(), "istio-system",
		kubeclient.V1DeleteOptions{GracePeriodSeconds: 0}, nil)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("Delete Cluster(%s) Namespace(istio-system) error %s", m.meshCluster.Spec.ClusterId, err.Error())
		return false
	}
	klog.Infof("Delete Cluster(%s) Namespace(istio-system) success", m.meshCluster.Spec.ClusterId)
	return true
}

func (m *MeshClusterManager) installIstio(){
	m.Lock()
	m.Unlock()
	//helm chart install IstioOperator
	setParam := map[string]string{
		"hub": m.conf.DockerHub,
		"tag": m.meshCluster.Spec.Version,
	}
	//create IstioOperator Crds
	m.createIstioOperatorCrds()
	//install istio-operator in cluster
	err := m.helm.InstallChart(setParam, m.conf.IstioOperatorCharts)
	if err!=nil {
		klog.Errorf("Install cluster(%s) istio-operator failed: %s", m.meshCluster.Spec.ClusterId, err.Error())
		return
	}
	klog.Infof("Install cluster(%s) istio-operator done", m.meshCluster.Spec.ClusterId)
	//update MeshCluster.Status in kube-apiserver
	m.updateComponentStatus()
}

func (m *MeshClusterManager) loopUpdateComponentStatus(){
	ticker := time.NewTicker(time.Minute)
	select {
	case <-ticker.C:
		m.updateComponentStatus()
	}
}

func (m *MeshClusterManager) updateComponentStatus(){
	for _,cStatus :=range m.meshCluster.Status.ComponentStatus {
		m.getComponentStatus(cStatus)
	}
	//update MeshCluster.Status in kube-apiserver
	err := m.meshClusterClient.Status().Update(context.Background(), m.meshCluster, nil)
	if err!=nil {
		klog.Errorf("Update ClusterId(%s) MeshCluster(%s) Status failed: %s", m.meshCluster.Spec.ClusterId,
			m.meshCluster.GetUuid(), err.Error())
	}
	klog.Infof("Save ClusterId(%s) MeshCluster(%s) Status success", m.meshCluster.Spec.ClusterId, m.meshCluster.GetUuid())
}

//if istio-operator installed, then return true
//else return false
func (m *MeshClusterManager) meshInstalled()bool{
	istioOperator := m.meshCluster.Status.ComponentStatus["istio-operator"]
	//if component istio-operator status==nil, show  istio-operator uninstalled
	if istioOperator==nil || istioOperator.Status==meshv1.InstallStatus_NONE {
		return false
	}

	return true
}

func (m *MeshClusterManager) getComponentStatus(status *meshv1.InstallStatus_VersionStatus){
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
	istiooperatorPlural := types.IstioOperatorPlural
	istiooperatorFullName := istiooperatorPlural + "." + types.IstioOperatorGroup
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: istiooperatorFullName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   types.IstioOperatorGroup,   // BcsLogConfigsGroup,
			Version: types.IstioOperatorVersion, // BcsLogConfigsVersion,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   istiooperatorPlural,
				Kind:     types.IstioOperatorKind,
				ListKind: types.IstioOperatorListKind,
			},
		},
	}
	//create IstioOperator Crd
	_, err := m.extensionClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			klog.Infof("IstioOperator Crd is already exists")
			return nil
		}
		klog.Errorf("create IstioOperator Crd error %s", err.Error())
		return err
	}
	klog.Infof("create IstioOperator Crd success")

	istioOperator := types.IstioOperator{
		TypeMeta: metav1.TypeMeta{
			Kind: types.IstioOperatorKind,
			APIVersion: fmt.Sprintf("%s/%s", types.IstioOperatorGroup, types.IstioOperatorVersion),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: types.IstioOperatorName,
			Namespace: types.IstioOperatorNamespace,
		},
		Spec: types.IstioOperatorSpec{
			Profile: types.ProfileTypeDefault,
		},
	}
	//create IstioOperator Cr Object
	_,_,err = m.kubeApiClient.CustomObjectsApi.CreateNamespacedCustomObject(context.Background(), types.IstioOperatorGroup,
		types.IstioOperatorVersion, istioOperator.Namespace, types.IstioOperatorPlural, istioOperator, nil)
	if err!=nil {
		if apierrors.IsAlreadyExists(err) {
			klog.Infof("IstioOperator Cr Object is already exists")
			return nil
		}
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