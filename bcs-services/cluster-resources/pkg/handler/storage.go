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

package handler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	handlerUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/util"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListPV ...
func (crh *clusterResourcesHandler) ListPV(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.PV, "", "", metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetPV ...
func (crh *clusterResourcesHandler) GetPV(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.PV, "", "", req.Name, metav1.GetOptions{},
	)
	return err
}

// CreatePV ...
func (crh *clusterResourcesHandler) CreatePV(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.PV, "", req.Manifest, false, metav1.CreateOptions{},
	)
	return err
}

// UpdatePV ...
func (crh *clusterResourcesHandler) UpdatePV(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.PV, "", "", req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeletePV ...
func (crh *clusterResourcesHandler) DeletePV(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.PV, "", "", req.Name, metav1.DeleteOptions{},
	)
}

// ListPVC ...
func (crh *clusterResourcesHandler) ListPVC(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.PVC, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetPVC ...
func (crh *clusterResourcesHandler) GetPVC(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.PVC, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreatePVC ...
func (crh *clusterResourcesHandler) CreatePVC(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.PVC, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdatePVC ...
func (crh *clusterResourcesHandler) UpdatePVC(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.PVC, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeletePVC ...
func (crh *clusterResourcesHandler) DeletePVC(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.PVC, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListSC ...
func (crh *clusterResourcesHandler) ListSC(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.SC, "", "", metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetSC ...
func (crh *clusterResourcesHandler) GetSC(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.SC, "", "", req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateSC ...
func (crh *clusterResourcesHandler) CreateSC(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.SC, "", req.Manifest, false, metav1.CreateOptions{},
	)
	return err
}

// UpdateSC ...
func (crh *clusterResourcesHandler) UpdateSC(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.SC, "", "", req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteSC ...
func (crh *clusterResourcesHandler) DeleteSC(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.SC, "", "", req.Name, metav1.DeleteOptions{},
	)
}
