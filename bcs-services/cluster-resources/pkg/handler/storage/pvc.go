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

package storage

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/perm"
	resAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resource"
	respUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resp"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/web"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListPVC ...
func (h *Handler) ListPVC(
	ctx context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	if err = perm.CheckNSAccess(ctx, req.ClusterID, req.Namespace); err != nil {
		return err
	}
	respData, err := respUtil.BuildListAPIRespData(
		ctx, respUtil.ListParams{
			req.ClusterID, resCsts.PVC, req.ApiVersion, req.Namespace, req.Format, req.Scene,
		}, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(respData); err != nil {
		return err
	}
	resp.WebAnnotations, err = web.GenListPVCWebAnnos(ctx, req.ClusterID, req.Namespace, respData)
	return err
}

// GetPVC ...
func (h *Handler) GetPVC(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	if err = perm.CheckNSAccess(ctx, req.ClusterID, req.Namespace); err != nil {
		return err
	}
	respData, err := respUtil.BuildRetrieveAPIRespData(
		ctx, respUtil.GetParams{
			req.ClusterID, resCsts.PVC, req.ApiVersion, req.Namespace, req.Name, req.Format,
		}, metav1.GetOptions{},
	)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(respData); err != nil {
		return err
	}
	resp.WebAnnotations, err = web.GenRetrievePVCWebAnnos(ctx, req.ClusterID, req.Namespace, resp.Data.AsMap())
	return err
}

// GetPVCMountInfo 获取 PVC 被挂载的 Pod 的名称信息
func (h *Handler) GetPVCMountInfo(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	if err = perm.CheckNSAccess(ctx, req.ClusterID, req.Namespace); err != nil {
		return err
	}

	pvcMountInfo := cli.NewPodCliByClusterID(ctx, req.ClusterID).GetPVCMountInfo(
		ctx, req.Namespace, metav1.ListOptions{},
	)
	podNames, ok := pvcMountInfo[req.Name]
	if !ok {
		podNames = []string{}
	}
	resp.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"podNames": podNames})
	return err
}

// CreatePVC ...
func (h *Handler) CreatePVC(
	ctx context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.PVC).Create(
		ctx, req.RawData, req.Format, true, metav1.CreateOptions{},
	)
	return err
}

// UpdatePVC ...
func (h *Handler) UpdatePVC(
	ctx context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.PVC).Update(
		ctx, req.Namespace, req.Name, req.RawData, req.Format, metav1.UpdateOptions{},
	)
	return err
}

// DeletePVC ...
func (h *Handler) DeletePVC(
	ctx context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return resAction.NewResMgr(req.ClusterID, "", resCsts.PVC).Delete(
		ctx, req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}
