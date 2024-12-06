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

package project

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	nsutils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/namespace"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// GetAction action for get project
type GetAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.GetProjectRequest
}

// NewGetAction new get project action
func NewGetAction(model store.ProjectModel) *GetAction {
	return &GetAction{
		model: model,
	}
}

// Do get project info
func (ga *GetAction) Do(ctx context.Context, req *proto.GetProjectRequest) (*pm.Project, error) {
	ga.ctx = ctx
	ga.req = req

	p, err := ga.model.GetProject(ctx, req.ProjectIDOrCode)
	if err != nil {
		return nil, errorx.NewDBErr(err.Error())
	}

	return p, nil
}

// Active get project active
func (ga *GetAction) Active(ctx context.Context, req *proto.GetProjectActiveRequest) (bool, error) {
	ga.ctx = ctx

	p, err := ga.model.GetProject(ctx, req.ProjectIDOrCode)
	if err != nil {
		return false, errorx.NewDBErr(err.Error())
	}

	// 未开启容器服务，直接返回不活跃
	if p.Kind == "" {
		return false, nil
	}

	clusters, err := clustermanager.ListClusters(p.ProjectID)
	if err != nil {
		return false, err
	}

	for _, v := range clusters {
		// 存在独立集群则返回活跃
		if !v.IsShared {
			return true, nil
		}
		isActive, err := listSharedNamespace(ctx, v.ClusterID, p.ProjectCode)
		if err != nil {
			return false, err
		}
		// 存在共享集群命名空间则返回活跃
		if isActive {
			return true, nil
		}
	}

	return false, nil
}

// listSharedNamespace list shares namespace
func listSharedNamespace(ctx context.Context, clusterID, projectCode string) (bool, error) {
	client, err := clientset.GetClientGroup().Client(clusterID)
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", clusterID, err.Error())
		return false, err
	}
	nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, errorx.NewClusterErr(err.Error())
	}

	namespaces := nsutils.FilterNamespaces(nsList, true, projectCode)

	if len(namespaces) != 0 {
		return true, nil
	}

	return false, nil
}
