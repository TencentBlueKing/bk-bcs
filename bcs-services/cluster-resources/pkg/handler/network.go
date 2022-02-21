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

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/service"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListIng ...
func (crh *ClusterResourcesHandler) ListIng(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.Ing).List(
		req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetIng ...
func (crh *ClusterResourcesHandler) GetIng(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.Ing).Get(
		req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateIng ...
func (crh *ClusterResourcesHandler) CreateIng(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.Ing).Create(
		req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateIng ...
func (crh *ClusterResourcesHandler) UpdateIng(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.Ing).Update(
		req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteIng ...
func (crh *ClusterResourcesHandler) DeleteIng(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.Ing).Delete(
		req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListSVC ...
func (crh *ClusterResourcesHandler) ListSVC(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.SVC).List(
		req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetSVC ...
func (crh *ClusterResourcesHandler) GetSVC(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.SVC).Get(
		req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateSVC ...
func (crh *ClusterResourcesHandler) CreateSVC(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.SVC).Create(
		req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateSVC ...
func (crh *ClusterResourcesHandler) UpdateSVC(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.SVC).Update(
		req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteSVC ...
func (crh *ClusterResourcesHandler) DeleteSVC(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.SVC).Delete(
		req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListEP ...
func (crh *ClusterResourcesHandler) ListEP(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.EP).List(
		req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetEP ...
func (crh *ClusterResourcesHandler) GetEP(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.EP).Get(
		req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateEP ...
func (crh *ClusterResourcesHandler) CreateEP(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.EP).Create(
		req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateEP ...
func (crh *ClusterResourcesHandler) UpdateEP(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.EP).Update(
		req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteEP ...
func (crh *ClusterResourcesHandler) DeleteEP(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.EP).Delete(
		req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}
