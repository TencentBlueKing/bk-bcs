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

package independent

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ListNativeNamespaces implement for ListNativeNamespaces interface
func (c *IndependentNamespaceAction) ListNativeNamespaces(ctx context.Context,
	req *proto.ListNativeNamespacesRequest, resp *proto.ListNativeNamespacesResponse) error {
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return errorx.NewClusterErr(err.Error())
	}
	retDatas := []*proto.NativeNamespaceData{}
	cluster, err := clustermanager.GetCluster(req.GetClusterID())
	if err != nil {
		logging.Error("get cluster %s from cluster-manager failed, err: %s", cluster, err.Error())
		return err
	}
	project, err := c.model.GetProject(ctx, cluster.GetProjectID())
	if err != nil {
		logging.Error("get project %s from db failed, err: %s", cluster, err.Error())
		return errorx.NewDBErr(err.Error())
	}
	for _, namespace := range nsList.Items {
		retData := &proto.NativeNamespaceData{
			Uid:         string(namespace.GetUID()),
			Name:        namespace.GetName(),
			Status:      string(namespace.Status.Phase),
			CreateTime:  namespace.GetCreationTimestamp().Format(constant.TimeLayout),
			ProjectID:   project.ProjectID,
			ProjectCode: project.ProjectCode,
		}
		retDatas = append(retDatas, retData)
	}
	resp.Data = retDatas
	return nil
}
