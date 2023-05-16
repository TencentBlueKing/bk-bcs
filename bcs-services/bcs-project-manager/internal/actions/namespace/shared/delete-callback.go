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

package shared

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bcscc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// DeleteNamespaceCallback implement for DeleteNamespaceCallback interface
func (a *SharedNamespaceAction) DeleteNamespaceCallback(ctx context.Context,
	req *proto.NamespaceCallbackRequest, resp *proto.NamespaceCallbackResponse) error {
	if !req.GetApproveResult() {
		return a.model.DeleteNamespace(ctx, req.GetProjectCode(), req.GetClusterID(), req.GetNamespace())
	}
	namespace, err := a.model.GetNamespaceByItsmTicketType(ctx, req.GetProjectCode(), req.GetClusterID(),
		req.GetNamespace(), nsm.ItsmTicketTypeDelete)
	if err != nil {
		logging.Error("get namespace %s/%s from db failed, err: %s", req.GetClusterID(), req.GetNamespace(), err.Error())
		return errorx.NewDBErr(err.Error())
	}
	if req.GetApplyInCluster() {
		// delete namespace in cluster
		client, err := clientset.GetClientGroup().Client(req.GetClusterID())
		if err != nil {
			logging.Error("get client for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
			return err
		}
		if err := client.CoreV1().Namespaces().Delete(ctx, namespace.Name, metav1.DeleteOptions{}); err != nil {
			logging.Error("delete namespace %s in cluster %s failed, err: %s",
				namespace.Name, req.GetClusterID(), err.Error())
			return errorx.NewClusterErr(err.Error())
		}
	}

	// delete variables
	if _, err := a.model.DeleteVariableValuesByNamespace(ctx, req.GetClusterID(), req.GetNamespace()); err != nil {
		logging.Error("delete variables in %s/%s failed, err: %s", req.GetClusterID(), req.GetNamespace(), err.Error())
		return errorx.NewDBErr(err.Error())
	}

	// delete namespace in db
	if err := a.model.DeleteNamespace(ctx, req.GetProjectCode(), req.GetClusterID(), req.GetNamespace()); err != nil {
		logging.Error("delete namespace %s/%s from db failed, err: %s", req.GetClusterID(), req.GetNamespace(), err.Error())
		return errorx.NewDBErr(err.Error())
	}
	go func() {
		if err := bcscc.DeleteNamespace(namespace.ProjectCode, namespace.ClusterID, namespace.Name); err != nil {
			logging.Error("[ALARM-CC-NAMESPACE] delete namespace %s/%s/%s in paas-cc failed, err: %s",
				namespace.ProjectCode, namespace.ClusterID, namespace.Name, err.Error())
		}
	}()
	return nil
}
