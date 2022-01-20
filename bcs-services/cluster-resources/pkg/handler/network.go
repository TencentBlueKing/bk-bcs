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

// ListIng ...
func (crh *clusterResourcesHandler) ListIng(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.Ing, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetIng ...
func (crh *clusterResourcesHandler) GetIng(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.Ing, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateIng ...
func (crh *clusterResourcesHandler) CreateIng(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.Ing, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateIng ...
func (crh *clusterResourcesHandler) UpdateIng(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.Ing, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteIng ...
func (crh *clusterResourcesHandler) DeleteIng(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.Ing, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListSVC ...
func (crh *clusterResourcesHandler) ListSVC(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.SVC, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetSVC ...
func (crh *clusterResourcesHandler) GetSVC(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.SVC, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateSVC ...
func (crh *clusterResourcesHandler) CreateSVC(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.SVC, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateSVC ...
func (crh *clusterResourcesHandler) UpdateSVC(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.SVC, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteSVC ...
func (crh *clusterResourcesHandler) DeleteSVC(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.SVC, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListEP ...
func (crh *clusterResourcesHandler) ListEP(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.EP, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetEP ...
func (crh *clusterResourcesHandler) GetEP(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.EP, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateEP ...
func (crh *clusterResourcesHandler) CreateEP(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.EP, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateEP ...
func (crh *clusterResourcesHandler) UpdateEP(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.EP, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteEP ...
func (crh *clusterResourcesHandler) DeleteEP(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.EP, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}
