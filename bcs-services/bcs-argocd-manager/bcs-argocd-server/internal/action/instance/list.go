/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package instance

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-server/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/apis/tkex/v1alpha1"
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/sdk/instance"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewListArgocdInstancesAction return a new ListArgocdInstancesAction instance
func NewListArgocdInstancesAction(tkexIf tkexv1alpha1.TkexV1alpha1Interface) *ListArgocdInstancesAction {
	return &ListArgocdInstancesAction{tkexIf: tkexIf}
}

// ListArgocdInstancesAction provides the action to list argocd instance
type ListArgocdInstancesAction struct {
	ctx context.Context

	tkexIf tkexv1alpha1.TkexV1alpha1Interface

	req  *instance.ListArgocdInstancesRequest
	resp *instance.ListArgocdInstancesResponse
}

// Handle the list process
func (action *ListArgocdInstancesAction) Handle(ctx context.Context,
	req *instance.ListArgocdInstancesRequest, resp *instance.ListArgocdInstancesResponse) error {
	if req == nil || resp == nil {
		blog.Errorf("action/instance/list: list instance failed, req or resp is empty")
		return common.ErrArgocdServerReqOrRespEmpty.GenError()
	}
	action.ctx = ctx
	action.req = req
	action.resp = resp

	listOptions := metav1.ListOptions{}
	if req.Project != nil && req.GetProject() != "" {
		// TODO: check permission?
		labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{common.ArgocdProjectLabel: req.GetProject()}}
		listOptions.LabelSelector = metav1.FormatLabelSelector(&labelSelector)
	}
	blog.Info("tkexIf: %v", action.tkexIf)
	list, err := action.tkexIf.ArgocdInstances(common.ArgocdManagerNamespace).List(ctx, listOptions)
	if err != nil {
		blog.Errorf("list instances failed, err: %s", err.Error())
		action.setResp(common.ErrActionFailed, "list instances failed", nil)
		return nil
	}
	blog.Info("list instances success")
	action.setResp(common.ErrArgocdServerSuccess, "", list)
	return nil
}

func (action *ListArgocdInstancesAction) setResp(err common.ArgocdServerError, message string, instances *v1alpha1.ArgocdInstanceList) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	action.resp.Code = &code
	action.resp.Message = &msg
	action.resp.Instances = instances
}
