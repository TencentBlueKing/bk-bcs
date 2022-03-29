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
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/sdk/instance"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewDeleteArgocdInstanceAction return a new DeleteArgocdInstanceAction instance
func NewDeleteArgocdInstanceAction(tkexIf tkexv1alpha1.TkexV1alpha1Interface) *DeleteArgocdInstanceAction {
	return &DeleteArgocdInstanceAction{tkexIf: tkexIf}
}

// DeleteArgocdInstanceAction provides the action to delete argocd instance
type DeleteArgocdInstanceAction struct {
	ctx context.Context

	tkexIf tkexv1alpha1.TkexV1alpha1Interface

	req  *instance.DeleteArgocdInstanceRequest
	resp *instance.DeleteArgocdInstanceResponse
}

// Handle the delete process
func (action *DeleteArgocdInstanceAction) Handle(ctx context.Context,
	req *instance.DeleteArgocdInstanceRequest, resp *instance.DeleteArgocdInstanceResponse) error {
	if req == nil || resp == nil {
		blog.Errorf("action/instance/delete: delete instance failed, req or resp is empty")
		return common.ErrArgocdServerReqOrRespEmpty.GenError()
	}
	action.ctx = ctx
	action.req = req
	action.resp = resp

	name := req.GetName()
	i, err := action.tkexIf.ArgocdInstances(common.ArgocdManagerNamespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("get instance %s failed, err: %s", name, err.Error())
		action.setResp(common.ErrActionFailed, "get argocd instance failed")
		return nil
	}
	project, err := action.tkexIf.ArgocdProjects(common.ArgocdManagerNamespace).Get(ctx, i.Spec.Project, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("get argocd project %s failed, err: %s", i.Spec.Project, err.Error())
		action.setResp(common.ErrActionFailed, "get argocd project failed")
		return nil
	}
	if project == nil {
		blog.Errorf("delete argocd instance failed, project %s not found", i.Spec.Project)
		action.setResp(common.ErrProjectNotExist, "")
		return nil
	}
	// TODO: check if the operator has permission in project
	err = action.tkexIf.ArgocdInstances(common.ArgocdManagerNamespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		blog.Errorf("delete argocd instance failed, err: %s", err.Error())
		action.setResp(common.ErrActionFailed, "delete argocd instance failed")
		return nil
	}
	blog.Infof("delete argocd instance %s success", name)
	action.setResp(common.ErrArgocdServerSuccess, "")
	return nil
}

func (action *DeleteArgocdInstanceAction) setResp(err common.ArgocdServerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	action.resp.Code = &code
	action.resp.Message = &msg
}
