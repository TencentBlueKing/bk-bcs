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

package customresource

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/util/perm"
	respUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/util/resp"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/util/trans"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/util/web"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListCObj xxx
func (h *Handler) ListCObj(
	ctx context.Context, req *clusterRes.CObjListReq, resp *clusterRes.CommonResp,
) error {
	crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	if err = validateNSParam(ctx, crdInfo, req.Namespace); err != nil {
		return err
	}
	if err = perm.CheckCObjAccess(ctx, req.ClusterID, req.CRDName, req.Namespace); err != nil {
		return err
	}
	kind, apiVersion := crdInfo["kind"].(string), crdInfo["apiVersion"].(string)
	respData, err := respUtil.BuildListAPIRespData(
		ctx, req.ClusterID, kind, apiVersion, req.Namespace, req.Format, metav1.ListOptions{},
	)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(respData); err != nil {
		return err
	}

	resp.WebAnnotations, err = web.GenCObjListWebAnnos(ctx, respData, crdInfo, req.Format)
	return err
}

// GetCObj xxx
func (h *Handler) GetCObj(
	ctx context.Context, req *clusterRes.CObjGetReq, resp *clusterRes.CommonResp,
) error {
	crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	if err = validateNSParam(ctx, crdInfo, req.Namespace); err != nil {
		return err
	}
	if err = perm.CheckCObjAccess(ctx, req.ClusterID, req.CRDName, req.Namespace); err != nil {
		return err
	}
	kind, apiVersion := crdInfo["kind"].(string), crdInfo["apiVersion"].(string)
	respData, err := respUtil.BuildRetrieveAPIRespData(
		ctx, req.ClusterID, kind, apiVersion, req.Namespace, req.CobjName, req.Format, metav1.GetOptions{},
	)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(respData); err != nil {
		return err
	}

	resp.WebAnnotations, err = web.GenRetrieveCObjWebAnnos(ctx, respData, crdInfo, req.Format)
	return err
}

// CreateCObj xxx
func (h *Handler) CreateCObj(
	ctx context.Context, req *clusterRes.CObjCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	kind, apiVersion := crdInfo["kind"].(string), crdInfo["apiVersion"].(string)

	transformer, err := trans.New(ctx, req.RawData.AsMap(), req.ClusterID, kind, req.Format)
	if err != nil {
		return err
	}
	manifest, err := transformer.ToManifest()
	if err != nil {
		return err
	}
	namespace := mapx.GetStr(manifest, "metadata.namespace")

	if err = validateNSParam(ctx, crdInfo, namespace); err != nil {
		return err
	}
	if err = perm.CheckCObjAccess(ctx, req.ClusterID, req.CRDName, namespace); err != nil {
		return err
	}
	// 经过命名空间检查后，若不需要指定命名空间，则认为是集群维度的
	resp.Data, err = respUtil.BuildCreateAPIResp(
		ctx, req.ClusterID, kind, apiVersion, manifest, namespace != "", metav1.CreateOptions{},
	)
	return err
}

// UpdateCObj xxx
func (h *Handler) UpdateCObj(
	ctx context.Context, req *clusterRes.CObjUpdateReq, resp *clusterRes.CommonResp,
) error {
	crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	kind, apiVersion := crdInfo["kind"].(string), crdInfo["apiVersion"].(string)

	transformer, err := trans.New(ctx, req.RawData.AsMap(), req.ClusterID, kind, req.Format)
	if err != nil {
		return err
	}
	manifest, err := transformer.ToManifest()
	if err != nil {
		return err
	}
	if err = validateNSParam(ctx, crdInfo, req.Namespace); err != nil {
		return err
	}
	if err = perm.CheckCObjAccess(ctx, req.ClusterID, req.CRDName, req.Namespace); err != nil {
		return err
	}
	resp.Data, err = respUtil.BuildUpdateCObjAPIResp(
		ctx, req.ClusterID, kind, apiVersion, req.Namespace, req.CobjName, manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteCObj xxx
func (h *Handler) DeleteCObj(
	ctx context.Context, req *clusterRes.CObjDeleteReq, resp *clusterRes.CommonResp,
) error {
	crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	if err = validateNSParam(ctx, crdInfo, req.Namespace); err != nil {
		return err
	}
	if err = perm.CheckCObjAccess(ctx, req.ClusterID, req.CRDName, req.Namespace); err != nil {
		return err
	}
	kind, apiVersion := crdInfo["kind"].(string), crdInfo["apiVersion"].(string)
	return respUtil.BuildDeleteAPIResp(
		ctx, req.ClusterID, kind, apiVersion, req.Namespace, req.CobjName, metav1.DeleteOptions{},
	)
}

// validateNSParam 校验 CObj 相关请求中命名空间参数，若 CRD 中定义为集群维度，则不需要，否则需要指定命名空间
func validateNSParam(ctx context.Context, crdInfo map[string]interface{}, namespace string) error {
	if namespace == "" && crdInfo["scope"].(string) == res.NamespacedScope {
		return errorx.New(errcode.ValidateErr, i18n.GetMsg(ctx, "查看/操作自定义资源 %s 需要指定命名空间"), crdInfo["name"])
	}
	return nil
}
