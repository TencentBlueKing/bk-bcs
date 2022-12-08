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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bcscc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	nsutils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/namespace"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
)

// SyncNamespace implement for SyncNamespace interface
func (a *SharedNamespaceAction) SyncNamespace(ctx context.Context,
	req *proto.SyncNamespaceRequest, resp *proto.SyncNamespaceResponse) error {
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	var creator string
	authUser, err := middleware.GetUserFromContext(ctx)
	if err == nil && authUser.Username != "" {
		// 授权创建者命名空间编辑和查看权限
		creator = authUser.Username
	}
	nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	nsItems := nsutils.FilterNamespaces(nsList, true, req.GetProjectCode())
	// namespaces =
	if err != nil {
		return errorx.NewClusterErr(err.Error())
	}
	ccNsList, err := bcscc.ListNamespaces(req.GetProjectCode(), req.GetClusterID())
	if err != nil {
		return errorx.NewRequestBCSCCErr(err.Error())
	}
	// insert new namespace to bcscc
	ccnsMap := map[string]bcscc.NamespaceData{}
	for _, ccns := range ccNsList.Results {
		ccnsMap[ccns.Name] = ccns
	}
	for _, item := range nsItems {
		if _, ok := ccnsMap[item.GetName()]; !ok {
			if err := bcscc.CreateNamespace(req.GetProjectCode(), req.GetClusterID(), item.GetName(), creator); err != nil {
				return errorx.NewRequestBCSCCErr(err.Error())
			}
		}
	}
	// delete old namespace in bcscc
	bcsnsMap := map[string]corev1.Namespace{}
	for _, item := range nsItems {
		bcsnsMap[item.GetName()] = item
	}
	for _, ns := range ccNsList.Results {
		if _, ok := bcsnsMap[ns.Name]; !ok {
			if err := bcscc.DeleteNamespace(req.GetProjectCode(), req.GetClusterID(), ns.Name); err != nil {
				return errorx.NewRequestBCSCCErr(err.Error())
			}
		}
	}
	return nil
}
