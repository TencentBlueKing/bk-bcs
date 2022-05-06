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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-server/internal/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/apis/tkex/v1alpha1"
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/sdk/instance"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewCreateArgocdInstanceAction return a new CreateArgocdInstanceAction instance
func NewCreateArgocdInstanceAction(tkexIf tkexv1alpha1.TkexV1alpha1Interface) *CreateArgocdInstanceAction {
	return &CreateArgocdInstanceAction{tkexIf: tkexIf}
}

// CreateArgocdInstanceAction provides the action to create argocd instance
type CreateArgocdInstanceAction struct {
	ctx context.Context

	tkexIf tkexv1alpha1.TkexV1alpha1Interface

	req  *instance.CreateArgocdInstanceRequest
	resp *instance.CreateArgocdInstanceResponse
}

// Handle the create process
func (action *CreateArgocdInstanceAction) Handle(ctx context.Context,
	req *instance.CreateArgocdInstanceRequest, resp *instance.CreateArgocdInstanceResponse) error {
	if req == nil || resp == nil {
		blog.Errorf("action/instance/create: create instance failed, req or resp is empty")
		return common.ErrArgocdServerReqOrRespEmpty.GenError()
	}
	action.ctx = ctx
	action.req = req
	action.resp = resp

	i := req.GetInstance()
	project, err := action.tkexIf.ArgocdProjects(common.ArgocdManagerNamespace).Get(ctx, i.Spec.Project, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("get argocd project %s failed, err: %s", i.Spec.Project, err.Error())
		action.setResp(common.ErrActionFailed, "get argocd project failed", i)
		return nil
	}
	if project == nil {
		blog.Errorf("create argocd instance failed, project %s not found", i.Spec.Project)
		action.setResp(common.ErrProjectNotExist, "", req.Instance)
		return nil
	}
	// TODO: check if the operator has permission in project
	i.Name = utils.RandomString(common.InstanceNamePrefix, 5)
	// set label
	if i.Labels == nil {
		i.Labels = make(map[string]string)
	}
	i.Labels[common.ArgocdProjectLabel] = i.Spec.Project
	created := &v1alpha1.ArgocdInstance{}
	for {
		created, err = action.tkexIf.ArgocdInstances(common.ArgocdManagerNamespace).Create(ctx, i, metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
			blog.Errorf("create argocd instance failed, instance %s already exists, retrying...", i.Name)
			continue
		}
		if err != nil {
			blog.Errorf("create argocd instance failed, err: %s", err.Error())
			action.setResp(common.ErrActionFailed, "create argocd instance failed", i)
			return nil
		}
		break
	}
	blog.Infof("create argocd instance %s success", created.Name)
	action.setResp(common.ErrArgocdServerSuccess, "", created)
	return nil
}

func (action *CreateArgocdInstanceAction) setResp(err common.ArgocdServerError, message string, instance *v1alpha1.ArgocdInstance) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	action.resp.Code = &code
	action.resp.Message = &msg
	action.resp.Instance = instance
}
