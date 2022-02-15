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

// Package handler rbac.go 权限类接口实现
package handler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	respUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/util/resp"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListSA ...
func (crh *ClusterResourcesHandler) ListSA(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = respUtil.BuildListAPIResp(
		req.ClusterID, res.SA, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetSA ...
func (crh *ClusterResourcesHandler) GetSA(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = respUtil.BuildRetrieveAPIResp(
		req.ClusterID, res.SA, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateSA ...
func (crh *ClusterResourcesHandler) CreateSA(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = respUtil.BuildCreateAPIResp(
		req.ClusterID, res.SA, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateSA ...
func (crh *ClusterResourcesHandler) UpdateSA(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = respUtil.BuildUpdateAPIResp(
		req.ClusterID, res.SA, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteSA ...
func (crh *ClusterResourcesHandler) DeleteSA(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return respUtil.BuildDeleteAPIResp(
		req.ClusterID, res.SA, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}
