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

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	quotautils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/quota"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// GetNamespace implement for GetNamespace interface
func (a *IndependentNamespaceAction) GetNamespace(ctx context.Context,
	req *proto.GetNamespaceRequest, resp *proto.GetNamespaceResponse) error {
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	ns, err := client.CoreV1().Namespaces().Get(ctx, req.GetName(), metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return errorx.NewClusterErr(err)
	}
	if errors.IsNotFound(err) {
		return errorx.NewReadableErr(errorx.ParamErr, "命名空间不存在")
	}
	retData := &proto.NamespaceData{
		Name:        ns.GetName(),
		Uid:         string(ns.GetUID()),
		Status:      string(ns.Status.Phase),
		CreateTime:  ns.GetCreationTimestamp().Format(config.TimeLayout),
		Labels:      []*proto.Label{},
		Annotations: []*proto.Annotation{},
	}
	for k, v := range ns.Labels {
		retData.Labels = append(retData.Labels, &proto.Label{Key: k, Value: v})
	}
	for k, v := range ns.Annotations {
		retData.Annotations = append(retData.Annotations, &proto.Annotation{Key: k, Value: v})
	}
	// get quota
	quota, err := getNamespaceQuota(ctx, req.GetProjectCode(), req.GetClusterID(), ns.GetName(), client)
	if err != nil {
		return err
	}
	if quota != nil {
		retData.Quota, retData.Used, retData.CpuUseRate, retData.MemoryUseRate = quotautils.TransferToProto(quota)
	}

	// get variables
	variables, err := listNamespaceVariables(ctx, req.GetProjectCode(), req.GetClusterID(), ns.GetName())
	if err != nil {
		logging.Error("get namespace %s/%s variables failed, err: %s", req.GetClusterID(), ns.GetName(), err.Error())
		return errorx.NewDBErr(err.Error())
	}
	retData.Variables = variables
	resp.Data = retData
	return nil
}
