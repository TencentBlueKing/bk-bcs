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

// Package handler cobj.go CRD，自定义资源接口实现
package handler

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/util/perm"
	respUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/util/resp"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListCRD ...
func (h *ClusterResourcesHandler) ListCRD(
	ctx context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) error {
	ret, err := cli.NewCRDCliByClusterID(req.ClusterID).List(metav1.ListOptions{LabelSelector: req.LabelSelector})
	if err != nil {
		return err
	}
	resp.Data, err = respUtil.GenListResRespData(ret, res.CRD)
	return err
}

// GetCRD ...
func (h *ClusterResourcesHandler) GetCRD(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) error {
	clusterInfo, err := cluster.GetClusterInfo(req.ClusterID)
	if err != nil {
		return err
	}
	if clusterInfo.Type == cluster.ClusterTypeShared && !cli.IsSharedClusterEnabledCRD(req.Name) {
		return fmt.Errorf("共享集群中不支持查看 CRD %s 信息", req.Name)
	}
	resp.Data, err = respUtil.BuildRetrieveAPIResp(
		req.ClusterID, res.CRD, "", "", req.Name, metav1.GetOptions{},
	)
	return err
}

// ListCObj ...
func (h *ClusterResourcesHandler) ListCObj(
	ctx context.Context, req *clusterRes.CObjListReq, resp *clusterRes.CommonResp,
) error {
	crdInfo, err := cli.GetCRDInfo(req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	if err = validateNSParam(crdInfo, req.Namespace); err != nil {
		return err
	}
	if err = perm.AccessCObjCheck(req.ProjectID, req.ClusterID, req.CRDName, req.Namespace); err != nil {
		return err
	}
	resp.Data, err = respUtil.BuildListAPIResp(
		req.ClusterID, crdInfo["kind"].(string), crdInfo["apiVersion"].(string), req.Namespace, metav1.ListOptions{},
	)
	return err
}

// GetCObj ...
func (h *ClusterResourcesHandler) GetCObj(
	ctx context.Context, req *clusterRes.CObjGetReq, resp *clusterRes.CommonResp,
) error {
	crdInfo, err := cli.GetCRDInfo(req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	if err = validateNSParam(crdInfo, req.Namespace); err != nil {
		return err
	}
	if err = perm.AccessCObjCheck(req.ProjectID, req.ClusterID, req.CRDName, req.Namespace); err != nil {
		return err
	}
	resp.Data, err = respUtil.BuildRetrieveAPIResp(
		req.ClusterID, crdInfo["kind"].(string), crdInfo["apiVersion"].(string), req.Namespace, req.CobjName, metav1.GetOptions{},
	)
	return err
}

// CreateCObj ...
func (h *ClusterResourcesHandler) CreateCObj(
	ctx context.Context, req *clusterRes.CObjCreateReq, resp *clusterRes.CommonResp,
) error {
	manifest := req.Manifest.AsMap()
	namespace := mapx.Get(manifest, "metadata.namespace", "").(string)

	crdInfo, err := cli.GetCRDInfo(req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	if err = validateNSParam(crdInfo, namespace); err != nil {
		return err
	}
	if err = perm.AccessCObjCheck(req.ProjectID, req.ClusterID, req.CRDName, namespace); err != nil {
		return err
	}
	// 经过命名空间检查后，若不需要指定命名空间，则认为是集群维度的
	resp.Data, err = respUtil.BuildCreateAPIResp(
		req.ClusterID, crdInfo["kind"].(string), crdInfo["apiVersion"].(string), req.Manifest, namespace != "", metav1.CreateOptions{},
	)
	return err
}

// UpdateCObj ...
func (h *ClusterResourcesHandler) UpdateCObj(
	ctx context.Context, req *clusterRes.CObjUpdateReq, resp *clusterRes.CommonResp,
) error {
	crdInfo, err := cli.GetCRDInfo(req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	if err = validateNSParam(crdInfo, req.Namespace); err != nil {
		return err
	}
	if err = perm.AccessCObjCheck(req.ProjectID, req.ClusterID, req.CRDName, req.Namespace); err != nil {
		return err
	}
	resp.Data, err = respUtil.BuildUpdateCObjAPIResp(
		req.ClusterID, crdInfo["kind"].(string), crdInfo["apiVersion"].(string), req.Namespace, req.CobjName, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteCObj ...
func (h *ClusterResourcesHandler) DeleteCObj(
	ctx context.Context, req *clusterRes.CObjDeleteReq, resp *clusterRes.CommonResp,
) error {
	crdInfo, err := cli.GetCRDInfo(req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	if err = validateNSParam(crdInfo, req.Namespace); err != nil {
		return err
	}
	if err = perm.AccessCObjCheck(req.ProjectID, req.ClusterID, req.CRDName, req.Namespace); err != nil {
		return err
	}
	return respUtil.BuildDeleteAPIResp(
		req.ClusterID, crdInfo["kind"].(string), crdInfo["apiVersion"].(string), req.Namespace, req.CobjName, metav1.DeleteOptions{},
	)
}

// 校验 CObj 相关请求中命名空间参数，若 CRD 中定义为集群维度，则不需要，否则需要指定命名空间
func validateNSParam(crdInfo map[string]interface{}, namespace string) error {
	if namespace == "" && crdInfo["scope"].(string) == res.NamespacedScope {
		return fmt.Errorf("查看/操作自定义资源 %s 需要指定命名空间", crdInfo["name"])
	}
	return nil
}
