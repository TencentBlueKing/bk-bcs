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

// Package handler network.go 网络类接口实现
package handler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	resAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resource"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListIng ...
func (h *ClusterResourcesHandler) ListIng(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.Ing).List(
		req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetIng ...
func (h *ClusterResourcesHandler) GetIng(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.Ing).Get(
		req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateIng ...
func (h *ClusterResourcesHandler) CreateIng(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.Ing).Create(
		req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateIng ...
func (h *ClusterResourcesHandler) UpdateIng(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.Ing).Update(
		req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteIng ...
func (h *ClusterResourcesHandler) DeleteIng(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.Ing).Delete(
		req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListSVC ...
func (h *ClusterResourcesHandler) ListSVC(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.SVC).List(
		req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetSVC ...
func (h *ClusterResourcesHandler) GetSVC(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.SVC).Get(
		req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateSVC ...
func (h *ClusterResourcesHandler) CreateSVC(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.SVC).Create(
		req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateSVC ...
func (h *ClusterResourcesHandler) UpdateSVC(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.SVC).Update(
		req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteSVC ...
func (h *ClusterResourcesHandler) DeleteSVC(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.SVC).Delete(
		req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListEP ...
func (h *ClusterResourcesHandler) ListEP(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.EP).List(
		req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetEP ...
func (h *ClusterResourcesHandler) GetEP(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.EP).Get(
		req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateEP ...
func (h *ClusterResourcesHandler) CreateEP(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.EP).Create(
		req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateEP ...
func (h *ClusterResourcesHandler) UpdateEP(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.EP).Update(
		req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteEP ...
func (h *ClusterResourcesHandler) DeleteEP(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.EP).Delete(
		req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}
