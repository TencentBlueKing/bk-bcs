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
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	handlerUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/util"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListDeploy 获取 Deployment 列表
func (crh *clusterResourcesHandler) ListDeploy(
	ctx context.Context, req *clusterRes.NamespaceScopedResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.Deploy, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetDeploy 获取单个 Deployment
func (crh *clusterResourcesHandler) GetDeploy(
	ctx context.Context, req *clusterRes.NamespaceScopedResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.Deploy, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateDeploy 创建 Deployment
func (crh *clusterResourcesHandler) CreateDeploy(
	ctx context.Context, req *clusterRes.NamespaceScopedResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.Deploy, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateDeploy 更新 Deployment
func (crh *clusterResourcesHandler) UpdateDeploy(
	ctx context.Context, req *clusterRes.NamespaceScopedResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.Deploy, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteDeploy 删除 Deployment
func (crh *clusterResourcesHandler) DeleteDeploy(
	ctx context.Context, req *clusterRes.NamespaceScopedResDeleteReq, resp *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.Deploy, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListDS 获取 DaemonSet 列表
func (crh *clusterResourcesHandler) ListDS(
	ctx context.Context, req *clusterRes.NamespaceScopedResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.DS, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetDS 获取单个 DaemonSet
func (crh *clusterResourcesHandler) GetDS(
	ctx context.Context, req *clusterRes.NamespaceScopedResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.DS, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateDS 创建 DaemonSet
func (crh *clusterResourcesHandler) CreateDS(
	ctx context.Context, req *clusterRes.NamespaceScopedResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.DS, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateDS 更新 DaemonSet
func (crh *clusterResourcesHandler) UpdateDS(
	ctx context.Context, req *clusterRes.NamespaceScopedResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.DS, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteDS 删除 DaemonSet
func (crh *clusterResourcesHandler) DeleteDS(
	ctx context.Context, req *clusterRes.NamespaceScopedResDeleteReq, resp *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.DS, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListSTS 获取 StatefulSet 列表
func (crh *clusterResourcesHandler) ListSTS(
	ctx context.Context, req *clusterRes.NamespaceScopedResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.STS, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetSTS 获取单个 StatefulSet
func (crh *clusterResourcesHandler) GetSTS(
	ctx context.Context, req *clusterRes.NamespaceScopedResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.STS, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateSTS 创建 StatefulSet
func (crh *clusterResourcesHandler) CreateSTS(
	ctx context.Context, req *clusterRes.NamespaceScopedResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.STS, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateSTS 更新 StatefulSet
func (crh *clusterResourcesHandler) UpdateSTS(
	ctx context.Context, req *clusterRes.NamespaceScopedResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.STS, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteSTS 删除 StatefulSet
func (crh *clusterResourcesHandler) DeleteSTS(
	ctx context.Context, req *clusterRes.NamespaceScopedResDeleteReq, resp *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.STS, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListCJ 获取 CronJob 列表
func (crh *clusterResourcesHandler) ListCJ(
	ctx context.Context, req *clusterRes.NamespaceScopedResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.CJ, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetCJ 获取单个 CronJob
func (crh *clusterResourcesHandler) GetCJ(
	ctx context.Context, req *clusterRes.NamespaceScopedResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.CJ, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateCJ 创建 CronJob
func (crh *clusterResourcesHandler) CreateCJ(
	ctx context.Context, req *clusterRes.NamespaceScopedResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.CJ, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateCJ 更新 CronJob
func (crh *clusterResourcesHandler) UpdateCJ(
	ctx context.Context, req *clusterRes.NamespaceScopedResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.CJ, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteCJ 删除 CronJob
func (crh *clusterResourcesHandler) DeleteCJ(
	ctx context.Context, req *clusterRes.NamespaceScopedResDeleteReq, resp *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.CJ, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListJob 获取 Job 列表
func (crh *clusterResourcesHandler) ListJob(
	ctx context.Context, req *clusterRes.NamespaceScopedResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListApiResp(
		req.ClusterID, res.Job, "", req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetJob 获取单个 Job
func (crh *clusterResourcesHandler) GetJob(
	ctx context.Context, req *clusterRes.NamespaceScopedResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.Job, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreateJob 创建 Job
func (crh *clusterResourcesHandler) CreateJob(
	ctx context.Context, req *clusterRes.NamespaceScopedResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.Job, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateJob 更新 Job
func (crh *clusterResourcesHandler) UpdateJob(
	ctx context.Context, req *clusterRes.NamespaceScopedResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.Job, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeleteJob 删除 Job
func (crh *clusterResourcesHandler) DeleteJob(
	ctx context.Context, req *clusterRes.NamespaceScopedResDeleteReq, resp *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.Job, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListPo 获取 Pod 列表
func (crh *clusterResourcesHandler) ListPo(
	ctx context.Context, req *clusterRes.PodResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildPodListApiResp(
		req.ClusterID, req.Namespace, req.OwnerKind, req.OwnerName, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetPo 获取单个 Pod
func (crh *clusterResourcesHandler) GetPo(
	ctx context.Context, req *clusterRes.NamespaceScopedResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildRetrieveApiResp(
		req.ClusterID, res.Po, "", req.Namespace, req.Name, metav1.GetOptions{},
	)
	return err
}

// CreatePo 创建 Pod
func (crh *clusterResourcesHandler) CreatePo(
	ctx context.Context, req *clusterRes.NamespaceScopedResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildCreateApiResp(
		req.ClusterID, res.Po, "", req.Manifest, true, metav1.CreateOptions{},
	)
	return err
}

// UpdatePo 更新 Pod
func (crh *clusterResourcesHandler) UpdatePo(
	ctx context.Context, req *clusterRes.NamespaceScopedResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildUpdateApiResp(
		req.ClusterID, res.Po, "", req.Namespace, req.Name, req.Manifest, metav1.UpdateOptions{},
	)
	return err
}

// DeletePo 删除 Pod
func (crh *clusterResourcesHandler) DeletePo(
	ctx context.Context, req *clusterRes.NamespaceScopedResDeleteReq, resp *clusterRes.CommonResp,
) error {
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.Po, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListPoPVC 获取 Pod PVC 列表
func (crh *clusterResourcesHandler) ListPoPVC(
	ctx context.Context, req *clusterRes.NamespaceScopedResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListPodRelatedResResp(req.ClusterID, req.Namespace, req.Name, res.PVC)
	return err
}

// ListPoCM 获取 Pod ConfigMap 列表
func (crh *clusterResourcesHandler) ListPoCM(
	ctx context.Context, req *clusterRes.NamespaceScopedResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListPodRelatedResResp(req.ClusterID, req.Namespace, req.Name, res.CM)
	return err
}

// ListPoSecret 获取 Pod Secret 列表
func (crh *clusterResourcesHandler) ListPoSecret(
	ctx context.Context, req *clusterRes.NamespaceScopedResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListPodRelatedResResp(req.ClusterID, req.Namespace, req.Name, res.Secret)
	return err
}

func (crh *clusterResourcesHandler) ReschedulePo(
	ctx context.Context, req *clusterRes.NamespaceScopedResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	podManifest, err := cli.NewPodResCliByClusterID(req.ClusterID).FetchManifest(req.Namespace, req.Name)
	if err != nil {
		return err
	}

	// 检查 Pod 配置，必须有父级资源且不为 Job 才可以重新调度
	ownerReferences, err := util.GetItems(podManifest, "metadata.ownerReferences")
	if err != nil {
		return fmt.Errorf("Pod %s/%s 不存在父级资源，不允许重新调度", req.Namespace, req.Name)
	}
	// 检查确保父级资源不为 Job
	for _, ref := range ownerReferences.([]interface{}) {
		if ref.(map[string]interface{})["kind"].(string) == res.Job {
			return fmt.Errorf("Pod %s/%s 父级资源存在 Job，不允许重新调度", req.Namespace, req.Name)
		}
	}

	// 重新调度的原理是直接删除 Pod，利用父级资源重新拉起服务
	return handlerUtil.BuildDeleteApiResp(
		req.ClusterID, res.Po, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListContainer 获取指定 Pod 容器列表
func (crh *clusterResourcesHandler) ListContainer(
	ctx context.Context, req *clusterRes.ContainerListReq, resp *clusterRes.CommonListResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildListContainerApiResp(req.ClusterID, req.Namespace, req.PodName)
	return err
}

// GetContainer 获取指定容器详情
func (crh *clusterResourcesHandler) GetContainer(
	ctx context.Context, req *clusterRes.ContainerGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = handlerUtil.BuildGetContainerApiResp(req.ClusterID, req.Namespace, req.PodName, req.ContainerName)
	return err
}

// GetContainerEnvInfo 获取指定容器环境变量信息
func (crh *clusterResourcesHandler) GetContainerEnvInfo(
	ctx context.Context, req *clusterRes.ContainerGetReq, resp *clusterRes.CommonListResp,
) error {
	envResp, _, err := cli.NewPodResCliByClusterID(req.ClusterID).ExecCommand(
		req.Namespace, req.PodName, req.ContainerName, []string{"/bin/sh", "-c", "env"},
	)
	if err != nil {
		return err
	}

	// 逐行解析 stdout，生成容器 env 信息
	envs := []map[string]interface{}{}
	for _, info := range strings.Split(envResp, "\n") {
		if len(info) == 0 {
			continue
		}
		key, val := util.Partition(info, "=")
		envs = append(envs, map[string]interface{}{
			"name": key, "value": val,
		})
	}
	resp.Data, err = util.MapSlice2ListValue(envs)
	return err
}
