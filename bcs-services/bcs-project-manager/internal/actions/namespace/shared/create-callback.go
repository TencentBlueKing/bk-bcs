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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bcscc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	vvm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variablevalue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	quotautils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/quota"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// CreateNamespaceCallback implement for CreateNamespaceCallback interface
func (a *SharedNamespaceAction) CreateNamespaceCallback(ctx context.Context,
	req *proto.NamespaceCallbackRequest, resp *proto.NamespaceCallbackResponse) error {
	if !req.GetApproveResult() {
		return a.model.DeleteNamespace(ctx, req.GetProjectCode(), req.GetClusterID(), req.GetNamespace())
	}
	ns, err := a.model.GetNamespaceByItsmTicketType(ctx, req.GetProjectCode(), req.GetClusterID(),
		req.GetNamespace(), nsm.ItsmTicketTypeCreate)
	if err != nil {
		logging.Error("get namespace %s/%s from db failed, err: %s", req.GetClusterID(), req.GetNamespace(), err.Error())
		return errorx.NewDBErr(err.Error())
	}
	// create variables
	for _, variable := range ns.Variables {
		entity := &vvm.VariableValue{}
		entity.Value = variable.VariableID
		entity.ClusterID = variable.ClusterID
		entity.Namespace = variable.Namespace
		entity.Scope = vdm.VariableScopeNamespace
		if uErr := a.model.UpsertVariableValue(ctx, entity); uErr != nil {
			logging.Error("create variable in %s/%s failed, err: %s", req.GetClusterID(), req.GetNamespace(), uErr.Error())
			return errorx.NewDBErr(uErr.Error())
		}
	}

	if req.GetApplyInCluster() {
		client, err := clientset.GetClientGroup().Client(req.GetClusterID())
		if err != nil {
			logging.Error("get client for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
			return err
		}
		// create namespace in cluster
		namespace := &corev1.Namespace{}
		namespace.SetName(ns.Name)
		namespace.SetAnnotations(map[string]string{
			config.GlobalConf.SharedClusterConfig.AnnoKeyProjCode: req.GetProjectCode(),
			constant.AnnotationKeyCreator:                         ns.Creator,
		})
		_, err = client.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
		if err != nil {
			logging.Error("create namespace in cluster %s failed, err: %s", req.GetClusterID(), err.Error())
			return errorx.NewClusterErr(err.Error())
		}
		// 授权创建者命名空间编辑和查看权限
		if e := iam.GrantNamespaceCreatorActions(ctx, ns.Creator, req.GetClusterID(), req.GetNamespace()); e != nil {
			logging.Error("grant namespace %s/%s for creator %s permission failed, err: %s",
				req.GetClusterID(), req.GetNamespace(), ns.Creator, e.Error())
		}
		// create quota in cluster
		if ns.ResourceQuota != nil {
			quota := &corev1.ResourceQuota{
				Spec: corev1.ResourceQuotaSpec{
					Hard: corev1.ResourceList{},
				},
			}
			quota.SetName(req.GetNamespace())
			quota.SetNamespace(req.GetNamespace())

			if lErr := quotautils.LoadFromModel(quota, ns.ResourceQuota); lErr != nil {
				return err
			}

			_, err = client.CoreV1().ResourceQuotas(req.GetNamespace()).Create(ctx, quota, metav1.CreateOptions{})
			if err != nil {
				logging.Error("create quota in cluster %s failed, err: %s", req.GetClusterID(), err.Error())
				return errorx.NewClusterErr(err.Error())
			}
		}
	}

	// delete namespace in db
	if err := a.model.DeleteNamespace(ctx, req.GetProjectCode(), req.GetClusterID(), req.GetNamespace()); err != nil {
		logging.Error("delete namespace %s/%s from db failed, err: %s", req.GetClusterID(), req.GetNamespace(), err.Error())
		return errorx.NewDBErr(err.Error())
	}
	go func() {
		if err := bcscc.CreateNamespace(ns.ProjectCode, ns.ClusterID, ns.Name, ns.Creator); err != nil {
			logging.Error("[ALARM-CC-NAMESPACE] create namespace %s/%s/%s in paas-cc failed, err: %s",
				ns.ProjectCode, ns.ClusterID, ns.Name, err.Error())
		}
	}()
	return nil
}
