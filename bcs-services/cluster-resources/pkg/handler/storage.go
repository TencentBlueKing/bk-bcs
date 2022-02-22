/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package handler storage.go 存储类接口实现
package handler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/service"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListPV ...
func (h *ClusterResourcesHandler) ListPV(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.PV).List(
		"", metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetPV ...
func (h *ClusterResourcesHandler) GetPV(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.PV).Get(
		"", req.Name, metav1.GetOptions{},
	)
	return err
}

// CreatePV ...
func (h *ClusterResourcesHandler) CreatePV(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.PV).Create(
		req.Manifest, false, metav1.CreateOptions{},
	)
	return err
}

// UpdatePV ...
func (h *ClusterResourcesHandler) UpdatePV(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.PV).Update(
		"", req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeletePV ...
func (h *ClusterResourcesHandler) DeletePV(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.PV).Delete(
		"", req.Name, metav1.DeleteOptions{},
	)
}

// ListPVC ...
func (h *ClusterResourcesHandler) ListPVC(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.PVC).List(
		req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetPVC ...
func (h *ClusterResourcesHandler) GetPVC(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.PVC).Get(
		req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreatePVC ...
func (h *ClusterResourcesHandler) CreatePVC(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.PVC).Create(
		req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdatePVC ...
func (h *ClusterResourcesHandler) UpdatePVC(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.PVC).Update(
		req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeletePVC ...
func (h *ClusterResourcesHandler) DeletePVC(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.PVC).Delete(
		req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListSC ...
func (h *ClusterResourcesHandler) ListSC(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.SC).List(
		"", metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetSC ...
func (h *ClusterResourcesHandler) GetSC(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.SC).Get(
		"", req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateSC ...
func (h *ClusterResourcesHandler) CreateSC(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.SC).Create(
		req.Manifest, false, metav1.CreateOptions{},
	)
	return err
}

// UpdateSC ...
func (h *ClusterResourcesHandler) UpdateSC(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.SC).Update(
		"", req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteSC ...
func (h *ClusterResourcesHandler) DeleteSC(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.SC).Delete(
		"", req.Name, metav1.DeleteOptions{},
	)
}
