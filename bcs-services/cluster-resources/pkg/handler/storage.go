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

	respUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/util/resp"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListPV ...
func (crh *ClusterResourcesHandler) ListPV(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = respUtil.BuildListAPIResp(
		req.ClusterID, res.PV, "", "", metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetPV ...
func (crh *ClusterResourcesHandler) GetPV(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = respUtil.BuildRetrieveAPIResp(
		req.ClusterID, res.PV, "", "", req.Name, metav1.GetOptions{},
	)
	return err
}

// CreatePV ...
func (crh *ClusterResourcesHandler) CreatePV(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = respUtil.BuildCreateAPIResp(
		req.ClusterID, res.PV, "", req.Manifest, false, metav1.CreateOptions{},
	)
	return err
}

// UpdatePV ...
func (crh *ClusterResourcesHandler) UpdatePV(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = respUtil.BuildUpdateAPIResp(
		req.ClusterID, res.PV, "", "", req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeletePV ...
func (crh *ClusterResourcesHandler) DeletePV(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return respUtil.BuildDeleteAPIResp(
		req.ClusterID, res.PV, "", "", req.Name, metav1.DeleteOptions{},
	)
}

// ListPVC ...
func (crh *ClusterResourcesHandler) ListPVC(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = respUtil.BuildListAPIResp(
		req.ClusterID, res.PVC, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetPVC ...
func (crh *ClusterResourcesHandler) GetPVC(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = respUtil.BuildRetrieveAPIResp(
		req.ClusterID, res.PVC, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreatePVC ...
func (crh *ClusterResourcesHandler) CreatePVC(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = respUtil.BuildCreateAPIResp(
		req.ClusterID, res.PVC, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdatePVC ...
func (crh *ClusterResourcesHandler) UpdatePVC(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = respUtil.BuildUpdateAPIResp(
		req.ClusterID, res.PVC, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeletePVC ...
func (crh *ClusterResourcesHandler) DeletePVC(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return respUtil.BuildDeleteAPIResp(
		req.ClusterID, res.PVC, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListSC ...
func (crh *ClusterResourcesHandler) ListSC(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = respUtil.BuildListAPIResp(
		req.ClusterID, res.SC, "", "", metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetSC ...
func (crh *ClusterResourcesHandler) GetSC(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = respUtil.BuildRetrieveAPIResp(
		req.ClusterID, res.SC, "", "", req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateSC ...
func (crh *ClusterResourcesHandler) CreateSC(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = respUtil.BuildCreateAPIResp(
		req.ClusterID, res.SC, "", req.Manifest, false, metav1.CreateOptions{},
	)
	return err
}

// UpdateSC ...
func (crh *ClusterResourcesHandler) UpdateSC(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = respUtil.BuildUpdateAPIResp(
		req.ClusterID, res.SC, "", "", req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteSC ...
func (crh *ClusterResourcesHandler) DeleteSC(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return respUtil.BuildDeleteAPIResp(
		req.ClusterID, res.SC, "", "", req.Name, metav1.DeleteOptions{},
	)
}
