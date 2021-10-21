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

package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	meshv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/api/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanager"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MeshHandler struct {
	//config
	conf config.Config
	//kubeclient
	client client.Client
}

func NewMeshHandler(conf config.Config, client client.Client) *MeshHandler {
	m := &MeshHandler{
		conf:   conf,
		client: client,
	}
	return m
}

// CreateMeshCluster is a single request handler called via client.Call or the generated client code
func (e *MeshHandler) CreateMeshCluster(ctx context.Context, req *meshmanager.CreateMeshClusterReq, resp *meshmanager.CreateMeshClusterResp) error {
	klog.Infof("Received MeshManager.CreateMeshCluster request(%s)", req.String())
	resp.ErrCode = meshmanager.ErrCode_ERROR_OK
	//check MeshCluster whether exist
	mCluster, err := e.getMeshClusterByClusterID(req.Clusterid)
	if err != nil {
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = err.Error()
		return nil
	}
	if mCluster != nil {
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_EXIST
		resp.ErrMsg = fmt.Sprintf("Cluster(%s) MeshCluster already exist", req.Clusterid)
		return nil
	}

	meshCluster := &meshv1.MeshCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       reflect.TypeOf(meshv1.MeshCluster{}).Name(),
			APIVersion: meshv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.ToLower(fmt.Sprintf("%s-%d", req.Clusterid, time.Now().Unix())),
			Namespace: "default",
		},
		Spec: meshv1.MeshClusterSpec{
			Version:       req.Version,
			ClusterID:     req.Clusterid,
			MeshType:      meshv1.MeshType(req.Meshtype),
			Configuration: req.Configurations,
		},
	}

	err = e.client.Create(context.TODO(), meshCluster)
	if err != nil {
		klog.Errorf("Create MeshCluster(%s) error %s", req.String(), err.Error())
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = err.Error()
	}
	klog.Infof("Create Cluster(%s) MeshCluster success", req.Clusterid)
	return nil
}

// DeleteMeshCluster is a server side stream handler called via client.DeleteMeshCluster or the generated client code
func (e *MeshHandler) DeleteMeshCluster(ctx context.Context, req *meshmanager.DeleteMeshClusterReq, resp *meshmanager.DeleteMeshClusterResp) error {
	klog.Infof("Received meshmanager.DeleteMeshCluster request(%s)", req.String())
	resp.ErrCode = meshmanager.ErrCode_ERROR_OK
	mCluster, err := e.getMeshClusterByClusterID(req.Clusterid)
	if err != nil {
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = err.Error()
		return nil
	}
	if mCluster == nil {
		resp.ErrCode = meshmanager.ErrCode_ERROR_NOT_FOUND
		resp.ErrMsg = fmt.Sprintf("Cluster(%s) MeshCluster NotFound", req.Clusterid)
		return nil
	}

	err = e.client.Delete(context.TODO(), mCluster)
	if err != nil {
		klog.Errorf("Create MeshCluster(%s) error %s", req.String(), err.Error())
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = err.Error()
	}
	klog.Infof("Delete Cluster(%s) MeshCluster success", req.Clusterid)
	return nil
}

// ListMeshCluster is a bidirectional stream handler called via client.ListMeshCluster or the generated client code
func (e *MeshHandler) ListMeshCluster(ctx context.Context, req *meshmanager.ListMeshClusterReq, resp *meshmanager.ListMeshClusterResp) error {
	klog.Infof("Received meshmanager.ListMeshCluster request with(%s)", req.String())
	resp.ErrCode = meshmanager.ErrCode_ERROR_OK
	resp.MeshClusters = make([]*meshmanager.MeshCluster, 0)
	mClusterList, err := e.listMeshClusters()
	if err != nil {
		resp.ErrCode = meshmanager.ErrCode_ERROR_MESH_CLUSTER_FAILED
		resp.ErrMsg = err.Error()
		return nil
	}
	if len(mClusterList.Items) == 0 {
		klog.Warning("List meshcluster empty")
		return nil
	}
	for _, in := range mClusterList.Items {
		if req.Clusterid != "" && in.Spec.ClusterID != req.Clusterid {
			continue
		}

		mCluster := &meshmanager.MeshCluster{
			Version:    in.Spec.Version,
			Clusterid:  in.Spec.ClusterID,
			Deletion:   !in.ObjectMeta.DeletionTimestamp.IsZero(),
			Components: make(map[string]*meshmanager.InstallStatus),
		}
		for k, v := range in.Status.ComponentStatus {
			mCluster.Components[k] = &meshmanager.InstallStatus{
				Name:      v.Name,
				Namespace: v.Namespace,
				Status:    string(v.Status),
				Message:   v.Message,
			}
		}
		resp.MeshClusters = append(resp.MeshClusters, mCluster)
	}
	by, _ := json.Marshal(resp)
	klog.Infof("response body(%s)", string(by))
	return nil
}

func (e *MeshHandler) getMeshClusterByClusterID(clusterId string) (*meshv1.MeshCluster, error) {
	mClusterList, err := e.listMeshClusters()
	if err != nil {
		return nil, err
	}

	for _, in := range mClusterList.Items {
		if in.Spec.ClusterID == clusterId {
			return &in, nil
		}
	}
	return nil, nil
}

//list all meshclusters
func (e *MeshHandler) listMeshClusters() (*meshv1.MeshClusterList, error) {
	mClusterList := &meshv1.MeshClusterList{}
	err := e.client.List(context.TODO(), mClusterList)
	if err != nil {
		klog.Errorf("list meshcluster failed: %s", err.Error())
	}
	return mClusterList, err
}
