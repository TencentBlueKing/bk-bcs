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

// ListDeploy 获取 Deployment 列表
func (crh *clusterResourcesHandler) ListDeploy(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResListReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.Deploy, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetDeploy 获取单个 Deployment
func (crh *clusterResourcesHandler) GetDeploy(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResGetReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.Deploy, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateDeploy 创建 Deployment
func (crh *clusterResourcesHandler) CreateDeploy(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResCreateReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.Deploy, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateDeploy 更新 Deployment
func (crh *clusterResourcesHandler) UpdateDeploy(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResUpdateReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.Deploy, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteDeploy 删除 Deployment
func (crh *clusterResourcesHandler) DeleteDeploy(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResDeleteReq,
	resp *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.Deploy, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListDS 获取 DaemonSet 列表
func (crh *clusterResourcesHandler) ListDS(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResListReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.DS, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetDS 获取单个 DaemonSet
func (crh *clusterResourcesHandler) GetDS(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResGetReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.DS, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateDS 创建 DaemonSet
func (crh *clusterResourcesHandler) CreateDS(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResCreateReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.DS, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateDS 更新 DaemonSet
func (crh *clusterResourcesHandler) UpdateDS(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResUpdateReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.DS, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteDS 删除 DaemonSet
func (crh *clusterResourcesHandler) DeleteDS(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResDeleteReq,
	resp *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.DS, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListSTS 获取 StatefulSet 列表
func (crh *clusterResourcesHandler) ListSTS(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResListReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.STS, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetSTS 获取单个 StatefulSet
func (crh *clusterResourcesHandler) GetSTS(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResGetReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.STS, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateSTS 创建 StatefulSet
func (crh *clusterResourcesHandler) CreateSTS(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResCreateReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.STS, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateSTS 更新 StatefulSet
func (crh *clusterResourcesHandler) UpdateSTS(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResUpdateReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.STS, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteSTS 删除 StatefulSet
func (crh *clusterResourcesHandler) DeleteSTS(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResDeleteReq,
	resp *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.STS, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListCJ 获取 CronJob 列表
func (crh *clusterResourcesHandler) ListCJ(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResListReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.CJ, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetCJ 获取单个 CronJob
func (crh *clusterResourcesHandler) GetCJ(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResGetReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.CJ, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateCJ 创建 CronJob
func (crh *clusterResourcesHandler) CreateCJ(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResCreateReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.CJ, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateCJ 更新 CronJob
func (crh *clusterResourcesHandler) UpdateCJ(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResUpdateReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.CJ, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteCJ 删除 CronJob
func (crh *clusterResourcesHandler) DeleteCJ(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResDeleteReq,
	resp *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.CJ, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListJob 获取 Job 列表
func (crh *clusterResourcesHandler) ListJob(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResListReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.Job, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetJob 获取单个 Job
func (crh *clusterResourcesHandler) GetJob(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResGetReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.Job, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateJob 创建 Job
func (crh *clusterResourcesHandler) CreateJob(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResCreateReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.Job, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateJob 更新 Job
func (crh *clusterResourcesHandler) UpdateJob(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResUpdateReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.Job, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteJob 删除 Job
func (crh *clusterResourcesHandler) DeleteJob(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResDeleteReq,
	resp *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.Job, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListPo 获取 Pod 列表
func (crh *clusterResourcesHandler) ListPo(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResListReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.Po, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetPo 获取单个 Pod
func (crh *clusterResourcesHandler) GetPo(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResGetReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.Po, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreatePo 创建 Pod
func (crh *clusterResourcesHandler) CreatePo(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResCreateReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.Po, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdatePo 更新 Pod
func (crh *clusterResourcesHandler) UpdatePo(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResUpdateReq,
	resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.Po, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeletePo 删除 Pod
func (crh *clusterResourcesHandler) DeletePo(
	ctx context.Context,
	req *clusterRes.NamespaceScopedResDeleteReq,
	resp *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.Po, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}
