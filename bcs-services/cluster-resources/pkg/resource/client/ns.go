/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/project"
	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam"
	clusterAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/cluster"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// NSClient xxx
type NSClient struct {
	ResClient
}

// NewNSClient xxx
func NewNSClient(ctx context.Context, conf *res.ClusterConf) *NSClient {
	NSRes, _ := res.GetGroupVersionResource(ctx, conf, resCsts.NS, "")
	return &NSClient{ResClient{NewDynamicClient(conf), conf, NSRes}}
}

// NewNSCliByClusterID xxx
func NewNSCliByClusterID(ctx context.Context, clusterID string) *NSClient {
	return NewNSClient(ctx, res.NewClusterConf(clusterID))
}

// List xxx
func (c *NSClient) List(ctx context.Context, opts metav1.ListOptions) (map[string]interface{}, error) {
	ret, err := c.ResClient.List(ctx, "", opts)
	if err != nil {
		return nil, err
	}
	manifest := ret.UnstructuredContent()
	// 共享集群命名空间，需要过滤出属于指定项目的
	clusterInfo, err := cluster.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	projInfo, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	if clusterInfo.Type == cluster.ClusterTypeShared && clusterInfo.ProjID != projInfo.ID {
		return filterProjNSList(ctx, manifest)
	}
	return manifest, nil
}

// ListByClusterViewPerm 获取集群中的全量命名空间，权限控制为：集群查看，建议只搭配 selectItems 使用
func (c *NSClient) ListByClusterViewPerm(
	ctx context.Context, projectID, clusterID string, opts metav1.ListOptions,
) (map[string]interface{}, error) {
	// 权限控制为集群查看
	permCtx := clusterAuth.NewPermCtx(
		ctx.Value(ctxkey.UsernameKey).(string), projectID, clusterID,
	)
	if allow, err := iam.NewClusterPerm(projectID).CanView(permCtx); err != nil {
		return nil, err
	} else if !allow {
		return nil, errorx.New(errcode.NoIAMPerm, i18n.GetMsg(ctx, "无集群查看权限"))
	}

	// 获取集群中所有的命名空间数据
	ret, err := c.cli.Resource(c.res).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	manifest := ret.UnstructuredContent()
	// 根据集群类型，决定是否按项目过滤命名空间
	clusterInfo, err := cluster.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	projInfo, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if clusterInfo.Type == cluster.ClusterTypeShared && clusterInfo.ProjID != projInfo.ID {
		return filterProjNSList(ctx, manifest)
	}
	return manifest, nil
}

// Get xxx
func (c *NSClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (map[string]interface{}, error) {
	if err := c.permValidate(ctx, action.View, name); err != nil {
		return nil, err
	}
	ret, err := c.cli.Resource(c.res).Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}
	return ret.UnstructuredContent(), nil
}

// Watch xxx
func (c *NSClient) Watch(
	ctx context.Context, projectCode string, clusterType string, opts metav1.ListOptions,
) (watch.Interface, error) {
	rawWatch, err := c.ResClient.Watch(ctx, "", opts)
	return &NSWatcher{rawWatch, projectCode, clusterType}, err
}

// CheckIsProjNSinSharedCluster 判断某命名空间，是否属于指定项目（仅共享集群有效）
func CheckIsProjNSinSharedCluster(ctx context.Context, clusterID, namespace string) error {
	if namespace == "" {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "命名空间为空"))
	}
	manifest, err := NewNSCliByClusterID(ctx, clusterID).Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return err
	}
	projInfo, err := project.FromContext(ctx)
	if err != nil {
		return err
	}
	clusterInfo, err := cluster.FromContext(ctx)
	if err != nil {
		return err
	}
	// 共享集群非管理端项目id和管理端不一致的情况需要判断某命名空间，是否属于指定项目
	if clusterInfo.ProjID != projInfo.ID && !isProjNSinSharedCluster(manifest, projInfo.Code) {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "命名空间 %s 在该共享集群中不属于指定项目"), namespace)
	}
	return nil
}

// filterProjNSList 从命名空间列表中，过滤出属于指定项目的（注：项目信息通过 context 获取）
func filterProjNSList(ctx context.Context, manifest map[string]interface{}) (map[string]interface{}, error) {
	projInfo, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	projNSList := []interface{}{}
	for _, ns := range mapx.GetList(manifest, "items") {
		if isProjNSinSharedCluster(ns.(map[string]interface{}), projInfo.Code) {
			projNSList = append(projNSList, ns)
		}
	}
	manifest["items"] = projNSList
	return manifest, nil
}

// isProjNSinSharedCluster 判断某命名空间，是否属于指定项目（仅共享集群有效）
func isProjNSinSharedCluster(manifest map[string]interface{}, projectCode string) bool {
	// 规则：属于项目的命名空间满足以下两点，但这里只需要检查 annotations 即可
	//   1. 命名(name) 以 ieg-{project_code}- 开头
	//   2. annotations 中包含 {annotation_key_project_code}: {project_code}
	//   3. {annotation_key_project_code} 默认值为 io.tencent.bcs.projectcode
	return mapx.GetStr(manifest, []string{
		"metadata", "annotations", conf.G.SharedCluster.AnnotationKeyProjectCode,
	}) == projectCode
}

// NSWatcher xxx
type NSWatcher struct {
	watch.Interface

	projectCode string
	clusterType string
}

// ResultChan xxx
func (w *NSWatcher) ResultChan() <-chan watch.Event {
	if w.clusterType == cluster.ClusterTypeSingle {
		return w.Interface.ResultChan()
	}
	// 共享集群，只能保留项目拥有的命名空间的事件
	resultChan := make(chan watch.Event)
	go func() {
		for event := range w.Interface.ResultChan() {
			if obj, ok := event.Object.(*unstructured.Unstructured); ok {
				if !isProjNSinSharedCluster(obj.UnstructuredContent(), w.projectCode) {
					continue
				}
			}
			resultChan <- event
		}
	}()
	return resultChan
}
