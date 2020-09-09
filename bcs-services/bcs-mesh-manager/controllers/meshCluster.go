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

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/kubehelm"
	meshv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/api/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/types"

	"k8s.io/klog"
	"sigs.k8s.io/yaml"
	istiov1alpha1 "istio.io/istio/operator/pkg/apis/istio/v1alpha1"
	kubeclient "github.com/kubernetes-client/go/kubernetes/client"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	apitypes "k8s.io/apimachinery/pkg/types"
)

var (
	MeshComponents = []string{"istio-operator","istiod","istio-egressgateway","istio-ingressgateway","istio-tracing","kiali"}
)

type MeshClusterManager struct {
	sync.RWMutex
	stopped bool
	stopCh chan struct{}
	meshCluster *meshv1.MeshCluster
	namespacedName apitypes.NamespacedName
	kubeconfig *restclient.Config
	kubeclientset *kubernetes.Clientset
	//config
	conf config.Config
	//MeshCluster Status Client
	meshClusterClient client.Client
	//kube-apiserver address
	kubeAddr string
	//kube Bearer token
	kubeToken string

	//apiextensions clientset
	extensionClientset *apiextensionsclient.Clientset
	//kubernetes api client
	kubeApiClient *kubeclient.APIClient
	//helm client
	helm kubehelm.KubeHelm
}

func NewMeshClusterManager(conf config.Config, meshCluster *meshv1.MeshCluster, client client.Client)(*MeshClusterManager,error){
	m := &MeshClusterManager{
		meshCluster: meshCluster,
		conf: conf,
		meshClusterClient: client,
		helm: kubehelm.NewCmdHelm(),
		namespacedName: apitypes.NamespacedName{
			Name: meshCluster.Name,
			Namespace: meshCluster.Namespace,
		},
		stopCh: make(chan struct{}),
		stopped: true,
		kubeAddr: fmt.Sprintf("%s/%s", conf.ServerAddress, meshCluster.Spec.ClusterId),
		kubeToken: conf.UserToken,
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
	m.kubeconfig = &restclient.Config{
		Host: m.kubeAddr,
		BearerToken: m.kubeToken,
		QPS: 1e3,
		Burst: 2e3,
	}
	klog.Infof("build kubeconfig Host(%s) BearerToken(%s)", m.kubeAddr, m.kubeToken)
	//kubernetes clientset
	var err error
	m.kubeclientset, err = kubernetes.NewForConfig(m.kubeconfig)
	if err != nil {
		klog.Errorf("build kubeclient by kubeconfig %s error %s", m.kubeconfig, err.Error())
		return nil, err
	}
	klog.Infof("build kubeclient for config %s success", m.kubeconfig)

	//apiextensions clientset for creating IstioOperator、MeshCluster Crd
	m.extensionClientset, err = apiextensionsclient.NewForConfig(m.kubeconfig)
	if err != nil {
		klog.Errorf("build apiextension client by kubeconfig % error %s", m.kubeconfig, err.Error())
		return nil, err
	}
	//update mesh cluster components status
	m.updateComponentStatus()
	//create MeshCluster Crd in kube-apiserver
	/*err = m.createMeshClusterCrd()
	if err!=nil {
		return nil, err
	}*/

	//kubernetes api client for create IstioOperator Object
	cfg := kubeclient.NewConfiguration()
	cfg.BasePath = m.kubeAddr
	cfg.DefaultHeader["authorization"] = fmt.Sprintf("Bearer %s", m.kubeToken)
	by,_ := json.Marshal(cfg)
	m.kubeApiClient = kubeclient.NewAPIClient(cfg)
	klog.Infof("build kubeapiclient for config %s success", string(by))
	klog.Infof("New MeshClusterManager(%s) success", meshCluster.GetUuid())
	//
	return m, nil
}

func (m *MeshClusterManager) stop(){
	close(m.stopCh)
	m.stopped = true
}

//if uninstall istio done, then return true
//else return false
func (m *MeshClusterManager) uninstallIstio()bool{
	m.Lock()
	m.Unlock()
	if !m.stopped {
		m.stop()
	}
	//delete IstioOperator Crd
	_,_,err := m.kubeApiClient.CustomObjectsApi.DeleteNamespacedCustomObject(context.Background(), types.IstioOperatorGroup,
		types.IstioOperatorVersion, types.IstioOperatorNamespace, types.IstioOperatorPlural, types.IstioOperatorName, kubeclient.V1DeleteOptions{}, nil)
	if err != nil && !strings.Contains(err.Error(), "404 Not Found") {
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
	//clear namespace istio-operator、istio-system resources
	return m.clearIstioOperatorResources()
}

func (m *MeshClusterManager) clearIstioOperatorResources()bool{
	//delete all resources in namespace istio-operator
	_,_,err := m.kubeApiClient.CoreV1Api.DeleteNamespace(context.Background(), "istio-operator",
		kubeclient.V1DeleteOptions{GracePeriodSeconds: 0}, nil)
	if err!=nil && !strings.Contains(err.Error(), "404 Not Found") {
		klog.Errorf("Delete Cluster(%s) Namespace(istio-operator) error %s", m.meshCluster.Spec.ClusterId, err.Error())
		return false
	}
	klog.Infof("Delete Cluster(%s) Namespace(istio-operator) success", m.meshCluster.Spec.ClusterId)

	//delete all resources in namespace istio-system
	_,_,err = m.kubeApiClient.CoreV1Api.DeleteNamespace(context.Background(), "istio-system",
		kubeclient.V1DeleteOptions{GracePeriodSeconds: 0}, nil)
	if err!=nil && !strings.Contains(err.Error(), "404 Not Found") {
		klog.Errorf("Delete Cluster(%s) Namespace(istio-system) error %s", m.meshCluster.Spec.ClusterId, err.Error())
		return false
	}
	klog.Infof("Delete Cluster(%s) Namespace(istio-system) success", m.meshCluster.Spec.ClusterId)

	//delete ClusterRole istio-operator
	_,_,err = m.kubeApiClient.RbacAuthorizationV1Api.DeleteClusterRole(context.Background(), "istio-operator",
		kubeclient.V1DeleteOptions{}, nil)
	if err!=nil && !strings.Contains(err.Error(), "404 Not Found") {
		klog.Errorf("Delete Cluster(%s) ClusterRole(istio-operator) error %s", m.meshCluster.Spec.ClusterId, err.Error())
		return false
	}
	klog.Infof("Delete Cluster(%s) ClusterRole(istio-operator) success", m.meshCluster.Spec.ClusterId)

	//delete ClusterRoleBinding istio-operator
	_,_,err = m.kubeApiClient.RbacAuthorizationV1Api.DeleteClusterRoleBinding(context.Background(), "istio-operator",
		kubeclient.V1DeleteOptions{}, nil)
	if err!=nil && !strings.Contains(err.Error(), "404 Not Found") {
		klog.Errorf("Delete Cluster(%s) ClusterRoleBinding(istio-operator) error %s", m.meshCluster.Spec.ClusterId, err.Error())
		return false
	}
	klog.Infof("Delete Cluster(%s) ClusterRoleBinding(istio-operator) success", m.meshCluster.Spec.ClusterId)
	return true
}

func (m *MeshClusterManager) installIstio()bool{
	m.Lock()
	m.Unlock()
	//create IstioOperator Crds
	/*err := m.createIstioOperatorCrds()
	if err!=nil {
		return false
	}*/

	//check deployment istio-operator whether installed
	if m.istioOperatorInstalled() {
		klog.Infof("Cluster(%s) Deployment IstioOperator have installed", m.meshCluster.Spec.ClusterId)
		return true
	}
	//helm chart install IstioOperator
	inf := kubehelm.InstallFlags{
		Chart: m.conf.IstioOperatorCharts,
		Name: fmt.Sprintf("istio-%d", time.Now().Unix()),
	}
	glf := kubehelm.GlobalFlags{
		Kubeconfig: "kubeconfig",
	}
	//clear istio resources
	if !m.clearIstioOperatorResources() {
		return false
	}
	//create namespace istio-system in kube-apiserver
	istiosystem := kubeclient.V1Namespace{
		ApiVersion: "v1",
		Kind: "Namespace",
		Metadata: &kubeclient.V1ObjectMeta{
			Name: "istio-system",
		},
	}
	_,_,err := m.kubeApiClient.CoreV1Api.CreateNamespace(context.Background(), istiosystem, nil)
	if err!=nil && !strings.Contains(err.Error(), "404 Not Found") {
		klog.Errorf("Create Cluster(%s) Namespace(istio-system) error %s", m.meshCluster.Spec.ClusterId, err.Error())
		return false
	}
	klog.Infof("Create Cluster(%s) Namespace(istio-system) success", m.meshCluster.Spec.ClusterId)
	//install istio-operator in cluster
	err = m.helm.InstallChart(inf, glf)
	if err!=nil {
		klog.Errorf("Install cluster(%s) istio-operator failed: %s", m.meshCluster.Spec.ClusterId, err.Error())
		return false
	}
	klog.Infof("Install cluster(%s) istio-operator done", m.meshCluster.Spec.ClusterId)
	//update MeshCluster.Status in kube-apiserver
	m.updateComponentStatus()
	go m.loopUpdateComponentStatus()
	m.stopped = false
	return true
}

func (m *MeshClusterManager) loopUpdateComponentStatus(){
	klog.Infof("Cluster(%s) start ticker update Istio Components status", m.meshCluster.Spec.ClusterId)
	ticker := time.NewTicker(time.Second*5)
	for{
		select {
		case <-ticker.C:
			m.updateComponentStatus()

		case <-m.stopCh:
			klog.Infof("Cluster(%s) stop ticker update Istio Components status", m.meshCluster.Spec.ClusterId)
			return
		}
	}
}

func (m *MeshClusterManager) updateComponentStatus(){
	var update bool
	for _,cStatus :=range m.meshCluster.Status.ComponentStatus {
		//if istio component status changed
		if m.getComponentStatus(cStatus) {
			update = true
		}
		//if last updatetime more than a minute, then update
		if (time.Now().Unix()-cStatus.UpdateTime) > 360 {
			update = true
			cStatus.UpdateTime = time.Now().Unix()
		}
	}

	if update {
		//update MeshCluster.Status in kube-apiserver
		err := m.meshClusterClient.Update(context.Background(), m.meshCluster)
		if err!=nil {
			klog.Errorf("Update ClusterId(%s) MeshCluster(%s) Status failed: %s", m.meshCluster.Spec.ClusterId,
				m.meshCluster.GetUuid(), err.Error())
			return
		}
		klog.Infof("Save ClusterId(%s) MeshCluster(%s) Status success", m.meshCluster.Spec.ClusterId, m.meshCluster.GetUuid())
	}
}

//if istio-operator installed, then return true
//else return false
func (m *MeshClusterManager) meshInstalled()bool{
	return m.istioOperatorInstalled()
}

//check deployment istio-operator whether installed
func (m *MeshClusterManager) istioOperatorInstalled()bool{
	istioOperator := m.meshCluster.Status.ComponentStatus["istio-operator"]
	//if component istio-operator status==nil, show  istio-operator uninstalled
	if istioOperator==nil || istioOperator.Status==meshv1.InstallStatus_NONE {
		return false
	}
	return true
}

//check deployment istio-operator whether installed
func (m *MeshClusterManager) istioOperatorCrdInstalled()bool{
	//read IstioOperator CR definition
	by,err := ioutil.ReadFile(m.conf.IstioOperatorCr)
	if err!=nil {
		klog.Errorf("read IstioOperator CR definition(%s) error %s", m.conf.IstioOperatorCr, err.Error())
		return false
	}
	var istioOperator *istiov1alpha1.IstioOperator
	err = yaml.Unmarshal(by, &istioOperator)
	if err!=nil {
		klog.Errorf("Unmarshal IstioOperator CR definition(%s) to types.IstioOperator error %s",
			string(by), err.Error())
		return false
	}
	by,_ = json.Marshal(istioOperator)
	klog.Infof("istioOperator %s", string(by))
	group := strings.Split(istioOperator.APIVersion, "/")[0]
	apiVersion := strings.Split(istioOperator.APIVersion, "/")[1]
	_,_,err = m.kubeApiClient.CustomObjectsApi.GetNamespacedCustomObject(context.Background(), group,
		apiVersion, istioOperator.Namespace, types.IstioOperatorPlural, istioOperator.Name)
	if err!=nil {
		klog.Errorf("Get IstioOperator(%s:%s) error %s",
			istioOperator.Namespace, istioOperator.Name, err.Error())
		return false
	}
	return true
}

func (m *MeshClusterManager) getComponentStatus(status *meshv1.InstallStatus_VersionStatus)(changed bool){
	oldStatus := status.Status
	defer func(){
		changed = false
		if oldStatus!=status.Status {
			klog.Infof("Cluster(%s) istio component(%s) status changed, from(%s)->to(%s)", m.meshCluster.Spec.ClusterId,
				status.Name, oldStatus, status.Status)
			changed = true
		}
	}()

	klog.Infof("MeshClusterManager start component(%s:%s) status", status.Namespace, status.Name)
	deployment,err := m.kubeclientset.AppsV1().Deployments(status.Namespace).Get(context.Background(), status.Name, metav1.GetOptions{})
	if err!=nil {
		if errors.IsNotFound(err){
			klog.Infof("Mesh Component(%s:%s) is NotFound", status.Namespace, status.Name)
			status.Status = meshv1.InstallStatus_NONE
			return
		}
		klog.Errorf("Mesh Component(%s:%s) Get Deployment failed: %s", status.Namespace, status.Name, err.Error())
		return
	}
	klog.Infof("Cluster(%s) Istio Component(%s:%s) status(%s)", m.meshCluster.Spec.ClusterId,
		status.Namespace, status.Name, deployment.Status.String())
	//status.Message = deployment.Status.
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

	return
}

// create crd of istiooperator
func (m *MeshClusterManager) createIstioOperatorCrds() error {
	istiooperatorPlural := types.IstioOperatorPlural
	istiooperatorFullName := istiooperatorPlural + "." + types.IstioOperatorGroup
	crd := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: istiooperatorFullName,
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group:   types.IstioOperatorGroup,   // BcsLogConfigsGroup,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				apiextensionsv1.CustomResourceDefinitionVersion{
					Name: types.IstioOperatorVersion,
				},
			}, // BcsLogConfigsVersion,
			Scope:   apiextensionsv1.NamespaceScoped,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   istiooperatorPlural,
				Kind:     types.IstioOperatorKind,
				ListKind: types.IstioOperatorListKind,
			},
		},
	}
	//create IstioOperator Crd
	_, err := m.extensionClientset.ApiextensionsV1().CustomResourceDefinitions().Create(context.Background(),
		crd, metav1.CreateOptions{})
	if err!=nil && !apierrors.IsAlreadyExists(err) {
		klog.Errorf("create IstioOperator Crd error %s", err.Error())
		return err
	}
	klog.Infof("create IstioOperator Crd success")

	//read IstioOperator CR definition
	/*by,err := ioutil.ReadFile(m.conf.IstioOperatorCr)
	if err!=nil {
		klog.Errorf("read IstioOperator CR definition(%s) error %s", m.conf.IstioOperatorCr, err.Error())
		return err
	}
	klog.Infof("byte(%s)", string(by))
	var istioOperator istiov1alpha1.IstioOperator
	err = yaml.Unmarshal(by, &istioOperator)
	if err!=nil {
		klog.Errorf("Unmarshal IstioOperator CR definition(%s) to types.IstioOperator error %s",
			string(by), err.Error())
		return err
	}
	group := strings.Split(istioOperator.APIVersion, "/")[0]
	apiVersion := strings.Split(istioOperator.APIVersion, "/")[1]
	//create IstioOperator Cr Object
	_,_,err = m.kubeApiClient.CustomObjectsApi.CreateNamespacedCustomObject(context.Background(), group,
		apiVersion, istioOperator.Namespace, types.IstioOperatorPlural, istioOperator, nil)
	if err!=nil {
		if apierrors.IsAlreadyExists(err) {
			klog.Infof("IstioOperator Cr Object is already exists")
			return nil
		}
		klog.Errorf("create IstioOperator Cr Object error %s", err.Error())
		return err
	}
	klog.Infof("create IstioOperator Cr Object success")*/
	return nil
}

// create crd of MeshCluster
func (m *MeshClusterManager) createMeshClusterCrd() error {
	meshclusterPlural := "meshclusters"
	meshclusterFullName := "meshclusters" + "." + meshv1.GroupVersion.Group
	crd := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: meshclusterFullName,
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group:   meshv1.GroupVersion.Group,   // BcsLogConfigsGroup,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				apiextensionsv1.CustomResourceDefinitionVersion{
					Name: meshv1.GroupVersion.Version,
				},
			}, // BcsLogConfigsVersion,
			Scope:   apiextensionsv1.NamespaceScoped,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   meshclusterPlural,
				Kind:     reflect.TypeOf(meshv1.MeshCluster{}).Name(),
				ListKind: reflect.TypeOf(meshv1.MeshClusterList{}).Name(),
			},
		},
	}

	_, err := m.extensionClientset.ApiextensionsV1().CustomResourceDefinitions().Create(context.Background(),
		crd, metav1.CreateOptions{})
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