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

// Package handler workload.go 工作负载类接口实现
package handler

import (
	"context"

	"google.golang.org/protobuf/types/known/structpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListDeploy 获取 Deployment 列表
func (crh *clusterResourcesHandler) ListDeploy(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResListReq,
	resp *clusterRes.CommonResp,
) error {
	clusterConf := res.NewClusterConfig(req.ClusterID)
	deployRes, err := res.GetGroupVersionResource(clusterConf, req.ClusterID, res.Deploy, "")
	if err != nil {
		return err
	}
	opts := metav1.ListOptions{LabelSelector: req.LabelSelector}
	ret, err := res.ListNamespaceScopedRes(clusterConf, req.Namespace, deployRes, opts)
	if err != nil {
		return err
	}
	manifest, _ := structpb.NewValue(ret.UnstructuredContent())
	resp.Data = &structpb.Struct{Fields: map[string]*structpb.Value{"manifest": manifest}}
	return nil
}

// GetDeploy 获取单个 Deployment
func (crh *clusterResourcesHandler) GetDeploy(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResGetReq,
	resp *clusterRes.CommonResp,
) error {
	clusterConf := res.NewClusterConfig(req.ClusterID)
	deployRes, err := res.GetGroupVersionResource(clusterConf, req.ClusterID, res.Deploy, "")
	if err != nil {
		return err
	}
	ret, err := res.GetNamespaceScopedRes(clusterConf, req.Namespace, req.Name, deployRes, metav1.GetOptions{})
	if err != nil {
		return err
	}
	manifest, _ := structpb.NewValue(ret.UnstructuredContent())
	resp.Data = &structpb.Struct{Fields: map[string]*structpb.Value{"manifest": manifest}}
	return nil
}

// CreateDeploy 创建 Deployment
func (crh *clusterResourcesHandler) CreateDeploy(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResCreateReq,
	resp *clusterRes.CommonResp,
) error {
	clusterConf := res.NewClusterConfig(req.ClusterID)
	deployRes, err := res.GetGroupVersionResource(clusterConf, req.ClusterID, res.Deploy, "")
	if err != nil {
		return err
	}
	ret, err := res.CreateNamespaceScopedRes(clusterConf, req.Manifest.AsMap(), deployRes, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	resp.Data = util.Unstructured2pbStruct(ret)
	return nil
}

// UpdateDeploy 更新 Deployment
func (crh *clusterResourcesHandler) UpdateDeploy(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResUpdateReq,
	resp *clusterRes.CommonResp,
) error {
	clusterConf := res.NewClusterConfig(req.ClusterID)
	deployRes, err := res.GetGroupVersionResource(clusterConf, req.ClusterID, res.Deploy, "")
	if err != nil {
		return err
	}
	ret, err := res.UpdateNamespaceScopedRes(
		clusterConf, req.Namespace, req.Name, req.Manifest.AsMap(), deployRes, metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}
	resp.Data = util.Unstructured2pbStruct(ret)
	return nil
}

// DeleteDeploy 删除 Deployment
func (crh *clusterResourcesHandler) DeleteDeploy(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResDeleteReq,
	resp *clusterRes.CommonResp,
) error {
	clusterConf := res.NewClusterConfig(req.ClusterID)
	deployRes, err := res.GetGroupVersionResource(clusterConf, req.ClusterID, res.Deploy, "")
	if err != nil {
		return err
	}
	err = res.DeleteNamespaceScopedRes(
		clusterConf, req.Namespace, req.Name, deployRes, metav1.DeleteOptions{},
	)
	return err
}
