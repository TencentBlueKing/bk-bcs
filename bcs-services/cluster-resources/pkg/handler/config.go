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

// ListCM ...
func (crh *clusterResourcesHandler) ListCM(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.CM, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetCM ...
func (crh *clusterResourcesHandler) GetCM(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.CM, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateCM ...
func (crh *clusterResourcesHandler) CreateCM(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.CM, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateCM ...
func (crh *clusterResourcesHandler) UpdateCM(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.CM, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteCM ...
func (crh *clusterResourcesHandler) DeleteCM(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.CM, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListSecret ...
func (crh *clusterResourcesHandler) ListSecret(
	_ context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.Secret, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetSecret ...
func (crh *clusterResourcesHandler) GetSecret(
	_ context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.Secret, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateSecret ...
func (crh *clusterResourcesHandler) CreateSecret(
	_ context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.Secret, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateSecret ...
func (crh *clusterResourcesHandler) UpdateSecret(
	_ context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.Secret, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteSecret ...
func (crh *clusterResourcesHandler) DeleteSecret(
	_ context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.Secret, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}
