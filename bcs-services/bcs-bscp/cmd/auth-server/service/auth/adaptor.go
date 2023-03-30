/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package auth

import (
	"errors"
	"fmt"

	bkiam "github.com/TencentBlueKing/iam-go-sdk"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/client"
	"bscp.io/pkg/iam/meta"
)

// AdaptAuthOptions convert bscp auth resource to iam action id and resources
func AdaptAuthOptions(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	if a == nil {
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("resource attribute is not set"))
	}

	// skip actions do not need to relate to resources
	if a.Basic.Action == meta.SkipAction {
		return genSkipResource(a)
	}

	switch a.Basic.Type {
	case meta.Biz:
		return genBizResource(a)
	case meta.App, meta.Commit, meta.ConfigItem, meta.Content, meta.CRInstance, meta.Release, meta.ReleasedCI, meta.Strategy, meta.StrategySet, meta.PSH, meta.Repo, meta.Sidecar:
		return genSkipResource(a)
	// case meta.App:
	// 	return genAppResource(a)
	// case meta.Commit:
	// 	return genCommitResource(a)
	// case meta.ConfigItem:
	// 	return genConfigItemResource(a)
	// case meta.Content:
	// 	return genContentResource(a)
	// case meta.CRInstance:
	// 	return genCRInstanceResource(a)
	// case meta.Release:
	// 	return genReleaseRes(a)
	// case meta.ReleasedCI:
	// 	return genReleasedCIRes(a)
	// case meta.Strategy:
	// 	return genStrategyRes(a)
	// case meta.StrategySet:
	// 	return genStrategySetRes(a)
	// case meta.PSH:
	// 	return genPSHRes(a)
	// case meta.Repo:
	// 	return genRepoRes(a)
	// case meta.Sidecar:
	// 	return genSidecarRes(a)
	default:
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported bscp auth type: %s", a.Basic.Type))
	}
}

// AdaptIAMResourceOptions 鉴权, 查看 isAllow 接口使用
func AdaptIAMResourceOptions(a *meta.ResourceAttribute) (*bkiam.Request, error) {
	if a == nil {
		return nil, errors.New("resource attribute is not set")
	}

	switch a.Basic.Type {
	case meta.Biz:
		return genBizIAMResource(a)
	default:
		return nil, fmt.Errorf("unsupported bscp auth type: %s", a.Basic.Type)
	}
}

// AdaptIAMApplicationOptions 申请链接, applyURL 接口使用
func AdaptIAMApplicationOptions(a *meta.ResourceAttribute) (*bkiam.Application, error) {
	if a == nil {
		return nil, errors.New("resource attribute is not set")
	}
	switch a.Basic.Type {
	case meta.Biz:
		return genBizIAMApplication(a)
	default:
		return nil, fmt.Errorf("unsupported bscp auth type: %s", a.Basic.Type)
	}
}
