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

package auth

import (
	"errors"
	"fmt"

	bkiam "github.com/TencentBlueKing/iam-go-sdk"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sys"
)

// AdaptAuthOptions convert bscp auth resource to iam action id and resources
func AdaptAuthOptions(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	if a == nil {
		return "", nil, errf.New(errf.InvalidParameter, "resource attribute is not set")
	}

	// skip actions do not need to relate to resources
	if a.Basic.Action == meta.SkipAction {
		return genSkipResource(a)
	}

	switch a.Basic.Type {
	case meta.Biz:
		return genBizResource(a)
	case meta.App:
		return genAppResource(a)
	case meta.Credential:
		return genCredResource(a)
	case meta.Commit, meta.ConfigItem, meta.Content, meta.CRInstance, meta.Release, meta.ReleasedCI, meta.Strategy,
		meta.StrategySet, meta.PSH, meta.Repo, meta.Sidecar:
		return genSkipResource(a)
	// case meta.Commit:
	//	return genCommitResource(a)
	// case meta.ConfigItem:
	//	return genConfigItemResource(a)
	// case meta.Content:
	//	return genContentResource(a)
	// case meta.CRInstance:
	//	return genCRInstanceResource(a)
	// case meta.Release:
	//	return genReleaseRes(a)
	// case meta.ReleasedCI:
	//	return genReleasedCIRes(a)
	// case meta.Strategy:
	//	return genStrategyRes(a)
	// case meta.StrategySet:
	//	return genStrategySetRes(a)
	// case meta.PSH:
	//	return genPSHRes(a)
	// case meta.Repo:
	//	return genRepoRes(a)
	// case meta.Sidecar:
	//	return genSidecarRes(a)
	// case meta.Credential:
	//	return genCredentialRes(a)

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
	case meta.App:
		return genAppIAMResource(a)
	case meta.Credential:
		return genCredIAMResource(a)
	default:
		return nil, fmt.Errorf("unsupported bscp auth type: %s", a.Basic.Type)
	}
}

// AdaptIAMApplicationOptions 申请链接, applyURL 接口使用
func AdaptIAMApplicationOptions(as []*meta.ResourceAttribute) (*bkiam.Application, error) {
	if len(as) == 0 {
		return nil, errors.New("resource attribute is not set")
	}
	actions := make([]bkiam.ApplicationAction, 0)
	for _, a := range as {
		switch a.Basic.Type {
		case meta.Biz:
			action, err := genBizIAMApplication(a)
			if err != nil {
				return nil, err
			}
			actions = append(actions, action)
		case meta.App:
			action, err := genAppIAMApplication(a)
			if err != nil {
				return nil, err
			}
			actions = append(actions, action)
		case meta.Credential:
			action, err := genCredIAMApplication(a)
			if err != nil {
				return nil, err
			}
			actions = append(actions, action)
		default:
			return nil, fmt.Errorf("unsupported bscp auth type: %s", a.Basic.Type)
		}
	}
	application := bkiam.NewApplication(sys.SystemIDBSCP, actions)
	return &application, nil
}
