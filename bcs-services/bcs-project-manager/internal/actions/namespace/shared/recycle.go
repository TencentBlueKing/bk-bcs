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

package shared

import (
	"context"
	"fmt"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// RecycleNamespace implement for RecycleNamespace interface
// 回收后命名空间不再归属任何项目，原项目将看不到、也无法操作其中的资源；
// 命名空间和其中的工作负载不会被删除；可选同时删除命名空间中的 ResourceQuota（默认删除）
func (a *SharedNamespaceAction) RecycleNamespace(ctx context.Context,
	req *proto.RecycleNamespaceRequest, resp *proto.RecycleNamespaceResponse) error {
	cluster, err := clustermanager.GetCluster(ctx, req.GetClusterID(), true)
	if err != nil {
		logging.Error("get cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	// 仅集群所属项目可回收（与分配/变更归属的权限语义一致）
	operatorProject, err := a.model.GetProject(ctx, req.GetProjectCode())
	if err != nil {
		logging.Error("get operator project %s failed, err: %s", req.GetProjectCode(), err.Error())
		return errorx.NewReadableErr(errorx.ProjectNotExistsErr, "操作方项目不存在")
	}
	if operatorProject.ProjectID != cluster.GetProjectID() {
		return errorx.NewReadableErr(errorx.PermDeniedErr, "仅集群所属项目可回收命名空间")
	}

	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	ns, err := client.CoreV1().Namespaces().Get(ctx, req.GetNamespace(), metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return errorx.NewReadableErr(errorx.ParamErr, fmt.Sprintf("命名空间 [%s] 不存在", req.GetNamespace()))
	}
	if err != nil {
		logging.Error("get namespace %s in cluster %s failed, err: %s",
			req.GetNamespace(), req.GetClusterID(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}

	annoKey := config.GlobalConf.SharedClusterConfig.AnnoKeyProjCode
	originProjectCode := ns.GetAnnotations()[annoKey]
	if originProjectCode == "" {
		return errorx.NewReadableErr(errorx.ParamErr,
			fmt.Sprintf("命名空间 [%s] 未分配给任何项目，无需回收", req.GetNamespace()))
	}
	// 审批中的命名空间不可回收
	if err := a.checkPendingTicket(ctx, originProjectCode, req.GetClusterID(), req.GetNamespace()); err != nil {
		return err
	}

	// 1. 删除归属 annotation，解除归属（命名空间与其中的工作负载保留）
	if ns.GetAnnotations() != nil {
		delete(ns.Annotations, annoKey)
		if _, err := client.CoreV1().Namespaces().Update(ctx, ns, metav1.UpdateOptions{}); err != nil {
			logging.Error("remove projectcode annotation from namespace %s/%s failed, err: %s",
				req.GetClusterID(), req.GetNamespace(), err.Error())
			return errorx.NewClusterErr(err.Error())
		}
	}

	// 2. 可选删除 ResourceQuota（「同时删除命名空间中的 ResourceQuota」勾选项）
	if req.GetDeleteQuota() {
		if err := client.CoreV1().ResourceQuotas(req.GetNamespace()).Delete(ctx,
			req.GetNamespace(), metav1.DeleteOptions{}); err != nil && !k8serrors.IsNotFound(err) {
			logging.Error("delete resourcequota %s/%s failed, err: %s",
				req.GetClusterID(), req.GetNamespace(), err.Error())
			return errorx.NewClusterErr(err.Error())
		}
	}

	// 3. 软删原归属项目的 DB 记录（无记录时跳过，保持幂等）
	if _, err := a.model.GetNamespace(ctx, originProjectCode, req.GetClusterID(), req.GetNamespace()); err == nil {
		if err := a.model.DeleteNamespace(ctx, originProjectCode, req.GetClusterID(), req.GetNamespace()); err != nil {
			logging.Error("delete namespace %s/%s in db failed, err: %s",
				req.GetClusterID(), req.GetNamespace(), err.Error())
			return errorx.NewDBErr(err.Error())
		}
	}

	resp.Data = &proto.NamespaceData{
		Name:       req.GetNamespace(),
		CreateTime: ns.GetCreationTimestamp().UTC().Format(time.RFC3339),
	}
	return nil
}
