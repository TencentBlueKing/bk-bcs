package handler

import (
	"context"
	"fmt"
	"reflect"
	"encoding/json"

	meshv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/config"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanager"

	"k8s.io/klog"
	kubeclient "github.com/kubernetes-client/go/kubernetes/client"
)

type MeshHandler struct{
	//config
	conf config.Config
	//kubernetes api client
	kubeApiClient *kubeclient.APIClient
}

func NewMeshHandler (conf config.Config)*MeshHandler{
	m := &MeshHandler{
		conf: conf,
	}
	//kubernetes api client for create IstioOperator Object
	cfg := kubeclient.NewConfiguration()
	cfg.BasePath = m.conf.ServerAddress
	cfg.DefaultHeader["authorization"] = fmt.Sprintf("Bearer %s", m.conf.UserToken)
	by,_ := json.Marshal(cfg)
	m.kubeApiClient = kubeclient.NewAPIClient(cfg)
	klog.Infof("build MeshHandler kubeapiclient for config %s success", string(by))
	return m
}

// CreateMeshCluster is a single request handler called via client.Call or the generated client code
func (e *MeshHandler) CreateMeshCluster(ctx context.Context, req *meshmanager.CreateMeshClusterReq) (*meshmanager.CreateMeshClusterResp, error) {
	klog.Infof("Received MeshManager.CreateMeshCluster request(%s)", req.String())
	resp := &meshmanager.CreateMeshClusterResp{
		ErrCode: meshmanager.ErrCode_ERROR_OK,
	}
	//check MeshCluster whether exist
	mCluster,err := e.getMeshClusterByClusterId(req.Clusterid)
	if err!=nil {
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = err.Error()
		return resp, nil
	}
	if mCluster!=nil {
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = fmt.Sprintf("Cluster(%s) MeshCluster already exist", req.Clusterid)
		return resp, nil
	}

	meshCluster := meshv1.MeshCluster{
		TypeMeta: metav1.TypeMeta{
			Kind: reflect.TypeOf(meshv1.MeshCluster{}).Name(),
			APIVersion: meshv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Name,
			Namespace: "bcs-system",
		},
		Spec: meshv1.MeshClusterSpec{
			Version: req.Version,
			ClusterId: req.Clusterid,
			MeshType: meshv1.MeshType(req.Meshtype),
		},
	}

	_,_,err = e.kubeApiClient.CustomObjectsApi.CreateNamespacedCustomObject(context.Background(), meshv1.GroupVersion.Group,
		meshv1.GroupVersion.Version, meshCluster.Namespace, "meshclusters", meshCluster, nil)
	if err!=nil {
		klog.Errorf("Create MeshCluster(%s) error %s", req.String(), err.Error())
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = err.Error()
	}
	klog.Infof("Create Cluster(%s) MeshCluster success", req.Clusterid)
	return resp, nil
}

// DeleteMeshCluster is a server side stream handler called via client.DeleteMeshCluster or the generated client code
func (e *MeshHandler) DeleteMeshCluster(ctx context.Context, req *meshmanager.DeleteMeshClusterReq) (*meshmanager.DeleteMeshClusterResp, error) {
	klog.Infof("Received meshmanager.DeleteMeshCluster request(%s)", req.String())

	resp := &meshmanager.DeleteMeshClusterResp{
		ErrCode: meshmanager.ErrCode_ERROR_OK,
	}
	mCluster,err := e.getMeshClusterByClusterId(req.Clusterid)
	if err!=nil {
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = err.Error()
		return resp, nil
	}
	if mCluster==nil {
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = fmt.Sprintf("Cluster(%s) MeshCluster NotFound", req.Clusterid)
		return resp, nil
	}

	_,_,err = e.kubeApiClient.CustomObjectsApi.DeleteNamespacedCustomObject(context.Background(), meshv1.GroupVersion.Group,
		meshv1.GroupVersion.Version, "bcs-system", "meshclusters", mCluster.Name, kubeclient.V1DeleteOptions{},nil)
	if err!=nil {
		klog.Errorf("Create MeshCluster(%s) error %s", req.String(), err.Error())
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = err.Error()
	}
	klog.Infof("Delete Cluster(%s) MeshCluster success", req.Clusterid)
	return resp, nil
}

// ListMeshCluster is a bidirectional stream handler called via client.ListMeshCluster or the generated client code
func (e *MeshHandler) ListMeshCluster(ctx context.Context, req *meshmanager.ListMeshClusterReq) (*meshmanager.ListMeshClusterResp, error) {
	klog.Infof("Received meshmanager.ListMeshCluster request with count: %d", req.String())
	resp := &meshmanager.ListMeshClusterResp{
		ErrCode: meshmanager.ErrCode_ERROR_OK,
		MeshClusters: make([]*meshmanager.MeshCluster, 0),
	}
	mClusterList,err := e.listMeshClusters()
	if err!=nil {
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = err.Error()
		return resp, nil
	}
	if len(mClusterList.Items)==0 {
		return resp, nil
	}
	for _,in :=range mClusterList.Items {
		mCluster := &meshmanager.MeshCluster{
			Version: in.Spec.Version,
			Clusterid: in.Spec.ClusterId,
			Components: make(map[string]*meshmanager.InstallStatus),
		}
		for k,v :=range in.Status.ComponentStatus {
			mCluster.Components[k] = &meshmanager.InstallStatus{
				Name: v.Name,
				Namespace: v.Namespace,
				Status: string(v.Status),
				Message: v.Message,
			}
		}
		resp.MeshClusters = append(resp.MeshClusters, mCluster)
	}

	return resp, nil
}

func (e *MeshHandler) getMeshClusterByClusterId(clusterId string)(*meshv1.MeshCluster,error){
	mClusterList,err := e.listMeshClusters()
	if err!=nil {
		return nil, err
	}

	for _,in :=range mClusterList.Items {
		if in.Spec.ClusterId==clusterId {
			return &in, nil
		}
	}
	return nil, nil
}

//list all meshclusters
func (e *MeshHandler) listMeshClusters()(meshv1.MeshClusterList, error){
	mClusterList := meshv1.MeshClusterList{}
	object,_,err := e.kubeApiClient.CustomObjectsApi.ListNamespacedCustomObject(context.Background(), meshv1.GroupVersion.Group,
		meshv1.GroupVersion.Version, "bcs-system", "meshclusters",nil)
	if err!=nil {
		klog.Errorf("ListNamespacedCustomObject failed: %s", err.Error())
		return mClusterList, err
	}
	by,_ := json.Marshal(object)
	klog.Infof(string(by))

	mClusterList,ok := object.(meshv1.MeshClusterList)
	if !ok {
		err = fmt.Errorf("interface to meshv1.MeshClusterList failed")
		return mClusterList, err
	}

	return mClusterList, nil
}