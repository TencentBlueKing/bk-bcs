package meshmanager

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanager"
)

type MeshManager interface {
	//create meshcluster crd and install istio service
	CreateMeshCluster(req *meshmanager.CreateMeshClusterReq)(*meshmanager.CreateMeshClusterResp,error)
	//delete meshcluster crd and uninstall istio service
	DeleteMeshCluster(req *meshmanager.DeleteMeshClusterReq)(*meshmanager.DeleteMeshClusterResp,error)
	//list meshcluster crds, contains istio components service status
	ListMeshCluster(req *meshmanager.ListMeshClusterReq)(*meshmanager.ListMeshClusterResp,error)
}
