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

package addons

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewListAddonsAction return a new ListAddonsAction instance
func NewListAddonsAction(model store.HelmManagerModel, addons release.AddonsSlice) *ListAddonsAction {
	return &ListAddonsAction{
		model:  model,
		addons: addons,
	}
}

// ListAddonsAction provides the action to do list addons
type ListAddonsAction struct {
	model  store.HelmManagerModel
	addons release.AddonsSlice

	req  *helmmanager.ListAddonsReq
	resp *helmmanager.ListAddonsResp
}

// Handle the addons listing process
func (l *ListAddonsAction) Handle(ctx context.Context,
	req *helmmanager.ListAddonsReq, resp *helmmanager.ListAddonsResp) error {
	l.req = req
	l.resp = resp
	l.setResp(common.ErrHelmManagerSuccess, "ok", nil)
	return nil
}

func (l *ListAddonsAction) setResp(err common.HelmManagerError, message string, r []*helmmanager.Addons) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	l.resp.Code = &code
	l.resp.Message = &msg
	l.resp.Result = err.OK()
	l.resp.Data = r
}
