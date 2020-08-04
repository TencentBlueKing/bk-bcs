package controllers

import (
	meshv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/api/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/config"

	corev1 "k8s.io/client-go/listers/core/v1"
	appsv1 "k8s.io/client-go/listers/apps/v1"
	restclient "k8s.io/client-go/rest"
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
	//component list
	installedComponents []string
}

func (m *istioClusterManager) uninstallIstio(){

}

func (m *istioClusterManager) installIstio(){

}