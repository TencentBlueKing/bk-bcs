package handler

import (
	"context"
	"fmt"
	"reflect"
	"encoding/json"

	meshv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/config"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmicro"

	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kubeclient "github.com/kubernetes-client/go/kubernetes/client"
)

type MeshHandler struct{
	//config
	conf config.Config
	//kubernetes api client
	kubeApiClient *kubeclient.APIClient
	//kubeclient
	client client.Client
}

func NewMeshHandler (conf config.Config, client client.Client)*MeshHandler{
	m := &MeshHandler{
		conf: conf,
		client: client,
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
func (e *MeshHandler) CreateMeshCluster(ctx context.Context, req *meshmanager.CreateMeshClusterReq,resp *meshmanager.CreateMeshClusterResp)error{
	klog.Infof("Received MeshManager.CreateMeshCluster request(%s)", req.String())
	resp.ErrCode = meshmanager.ErrCode_ERROR_OK
	//check MeshCluster whether exist
	mCluster,err := e.getMeshClusterByClusterId(req.Clusterid)
	if err!=nil {
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = err.Error()
		return err
	}
	if mCluster!=nil {
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_EXIST
		resp.ErrMsg = fmt.Sprintf("Cluster(%s) MeshCluster already exist", req.Clusterid)
		return nil
	}

	meshCluster := &meshv1.MeshCluster{
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

	err = e.client.Create(context.TODO(), meshCluster)
	if err!=nil {
		klog.Errorf("Create MeshCluster(%s) error %s", req.String(), err.Error())
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = err.Error()
	}
	klog.Infof("Create Cluster(%s) MeshCluster success", req.Clusterid)
	return err
}

// DeleteMeshCluster is a server side stream handler called via client.DeleteMeshCluster or the generated client code
func (e *MeshHandler) DeleteMeshCluster(ctx context.Context, req *meshmanager.DeleteMeshClusterReq, resp *meshmanager.DeleteMeshClusterResp)error{
	klog.Infof("Received meshmanager.DeleteMeshCluster request(%s)", req.String())
	resp.ErrCode = meshmanager.ErrCode_ERROR_OK
	mCluster,err := e.getMeshClusterByClusterId(req.Clusterid)
	if err!=nil {
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = err.Error()
		return err
	}
	if mCluster==nil {
		resp.ErrCode = meshmanager.ErrCode_ERROR_NOT_FOUND
		resp.ErrMsg = fmt.Sprintf("Cluster(%s) MeshCluster NotFound", req.Clusterid)
		return nil
	}

	err = e.client.Delete(context.TODO(), mCluster)
	if err!=nil {
		klog.Errorf("Create MeshCluster(%s) error %s", req.String(), err.Error())
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = err.Error()
	}
	klog.Infof("Delete Cluster(%s) MeshCluster success", req.Clusterid)
	return err
}

// ListMeshCluster is a bidirectional stream handler called via client.ListMeshCluster or the generated client code
func (e *MeshHandler) ListMeshCluster(ctx context.Context, req *meshmanager.ListMeshClusterReq, resp *meshmanager.ListMeshClusterResp)error{
	klog.Infof("Received meshmanager.ListMeshCluster request with count: %d", req.String())
	resp.ErrCode = meshmanager.ErrCode_ERROR_OK
	resp.MeshClusters = make([]*meshmanager.MeshCluster, 0)
	mClusterList,err := e.listMeshClusters()
	if err!=nil {
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = err.Error()
		return err
	}
	if len(mClusterList.Items)==0 {
		return nil
	}
	for _,in :=range mClusterList.Items {
		if req.Clusterid!="" && in.Spec.ClusterId!=req.Clusterid {
			continue
		}

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

	return nil
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
func (e *MeshHandler) listMeshClusters()(*meshv1.MeshClusterList, error){
	mClusterList := &meshv1.MeshClusterList{}
	err := e.client.List(context.TODO(), mClusterList)
	return mClusterList, err
}