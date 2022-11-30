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

package independent

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bcscc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// DeleteNamespace implement for DeleteNamespace interface
func (c *IndependentNamespaceAction) DeleteNamespace(ctx context.Context,
	req *proto.DeleteNamespaceRequest, resp *proto.DeleteNamespaceResponse) error {
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	if err := client.CoreV1().Namespaces().Delete(ctx, req.GetNamespace(), metav1.DeleteOptions{}); err != nil {
		logging.Error("delete namespace %s/%s failed, errr: %s", req.GetClusterID(), req.GetNamespace(), err.Error())
		return err
	}
	// delete variables
	if _, err := c.model.DeleteVariableValuesByNamespace(ctx, req.GetClusterID(), req.GetNamespace()); err != nil {
		logging.Error("delete variables in %s/%s failed, err: %s", req.GetClusterID(), req.GetNamespace(), err.Error())
		return errorx.NewDBErr(err.Error())
	}
	go func() {
		if err := bcscc.DeleteNamespace(req.GetProjectCode(), req.GetClusterID(), req.GetNamespace()); err != nil {
			logging.Error("[ALARM-CC-NAMESPACE] delete namespace %s/%s/%s in paas-cc failed, err: %s",
				req.GetProjectCode(), req.GetClusterID(), req.GetNamespace(), err.Error())
		}
	}()
	return nil
}
