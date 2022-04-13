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

package plugin

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-server/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/apis/tkex/v1alpha1"
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/sdk/plugin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewListArgocdPluginsAction return a new ListArgocdPluginsAction instance
func NewListArgocdPluginsAction(tkexIf tkexv1alpha1.TkexV1alpha1Interface) *ListArgocdPluginsAction {
	return &ListArgocdPluginsAction{tkexIf: tkexIf}
}

// ListArgocdPluginsAction provides the action to list argocd plugin
type ListArgocdPluginsAction struct {
	ctx context.Context

	tkexIf tkexv1alpha1.TkexV1alpha1Interface

	req  *plugin.ListArgocdPluginsRequest
	resp *plugin.ListArgocdPluginsResponse
}

// Handle the list process
func (action *ListArgocdPluginsAction) Handle(ctx context.Context,
	req *plugin.ListArgocdPluginsRequest, resp *plugin.ListArgocdPluginsResponse) error {

	if req == nil || resp == nil {
		blog.Errorf("action/plugin/list: list plugin failed, req or resp is empty")
		return common.ErrArgocdServerReqOrRespEmpty.GenError()
	}
	action.ctx = ctx
	action.req = req
	action.resp = resp

	// TODO: check project permission?
	listOptions := metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: action.getMatchLabels()}),
	}
	list, err := action.tkexIf.ArgocdPlugins(common.ArgocdManagerNamespace).List(ctx, listOptions)
	if err != nil {
		blog.Errorf("list plugins failed, err: %s", err.Error())
		action.setResp(common.ErrActionFailed, err.Error(), nil)
		return nil
	}

	blog.Infof("success to list plugins")
	action.setResp(common.ErrArgocdServerSuccess, "", list)
	return nil
}

func (action *ListArgocdPluginsAction) getMatchLabels() map[string]string {
	r := make(map[string]string)
	if action.req.Project != nil {
		r[common.ArgocdProjectLabel] = action.req.GetProject()
	}
	if action.req.NickName != nil {
		r[common.ArgocdNickNameLabel] = action.req.GetNickName()
	}

	return r
}

func (action *ListArgocdPluginsAction) setResp(err common.ArgocdServerError, message string, ps *v1alpha1.ArgocdPluginList) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	action.resp.Code = &code
	action.resp.Message = &msg
	action.resp.Plugins = ps
}
