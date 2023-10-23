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

package addons

import (
	"context"
	"errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewListAddonsAction return a new ListAddonsAction instance
func NewListAddonsAction(model store.HelmManagerModel, addons release.AddonsSlice,
	platform repo.Platform) *ListAddonsAction {
	return &ListAddonsAction{
		model:    model,
		addons:   addons,
		platform: platform,
	}
}

// ListAddonsAction provides the action to do list addons
type ListAddonsAction struct {
	model    store.HelmManagerModel
	addons   release.AddonsSlice
	platform repo.Platform

	req  *helmmanager.ListAddonsReq
	resp *helmmanager.ListAddonsResp
}

// Handle the addons listing process
func (l *ListAddonsAction) Handle(ctx context.Context,
	req *helmmanager.ListAddonsReq, resp *helmmanager.ListAddonsResp) error {
	l.req = req
	l.resp = resp

	if err := req.Validate(); err != nil {
		blog.Errorf("get addons detail failed, invalid request, %s, param: %v", err.Error(), l.req)
		l.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	// 创建一个Addons数组，容量为配置文件数组的长度
	addons := make([]*helmmanager.Addons, 0, len(l.addons.Addons))

	for _, value := range l.addons.Addons {
		addon := value.ToAddonsProto()

		// get current status
		rl, err := l.model.GetRelease(ctx, l.req.GetClusterID(), *addon.Namespace, *addon.Name)
		if err != nil {
			if errors.Is(err, drivers.ErrTableRecordNotFound) {
				// 没有记录情况下不处理，继续
				addons = append(addons, addon)
				continue
			} else {
				blog.Errorf("get addons status failed, %s", err.Error())
				l.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
				return nil
			}
		}
		addon.CurrentVersion = &rl.ChartVersion
		addon.Status = &rl.Status
		addon.Message = &rl.Message
		if len(rl.Values) > 0 {
			addon.CurrentValues = &rl.Values[len(rl.Values)-1]
		}

		addons = append(addons, addon)
	}

	l.setResp(common.ErrHelmManagerSuccess, "ok", addons)
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
