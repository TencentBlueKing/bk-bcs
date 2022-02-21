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

// Package handler config.go 配置类接口实现
package handler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/service"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListCM ...
func (crh *ClusterResourcesHandler) ListCM(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.CM).List(
		req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetCM ...
func (crh *ClusterResourcesHandler) GetCM(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.CM).Get(
		req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateCM ...
func (crh *ClusterResourcesHandler) CreateCM(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.CM).Create(
		req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateCM ...
func (crh *ClusterResourcesHandler) UpdateCM(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.CM).Update(
		req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteCM ...
func (crh *ClusterResourcesHandler) DeleteCM(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.CM).Delete(
		req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListSecret ...
func (crh *ClusterResourcesHandler) ListSecret(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.Secret).List(
		req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetSecret ...
func (crh *ClusterResourcesHandler) GetSecret(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.Secret).Get(
		req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateSecret ...
func (crh *ClusterResourcesHandler) CreateSecret(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.Secret).Create(
		req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateSecret ...
func (crh *ClusterResourcesHandler) UpdateSecret(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.Secret).Update(
		req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteSecret ...
func (crh *ClusterResourcesHandler) DeleteSecret(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return service.NewK8SResMgr(req.ProjectID, req.ClusterID, "", res.Secret).Delete(
		req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}
