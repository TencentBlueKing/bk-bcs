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

package perm

import (
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
)

// IAMPerm ...
type IAMPerm struct {
	Cli           *IAMClient
	ResType       string
	PermCtx       Ctx
	ResReq        ResRequest
	ParentResPerm *IAMPerm
}

// CanAction 校验用户的 action_id 权限
func (p *IAMPerm) CanAction(ctx Ctx, actionID string, useCache bool) (bool, error) {
	if err := ctx.Validate([]string{actionID}); err != nil {
		return false, err
	}
	if ctx.ForceRaise() {
		return false, p.genIAMPermError(ctx, actionID)
	}
	allow, err := p.canAction(ctx, actionID, useCache)
	if allow && err == nil {
		return true, nil
	}
	return false, p.genIAMPermError(ctx, actionID)
}

// CanMultiActions 校验同类型单个资源的多个 action 权限
func (p *IAMPerm) CanMultiActions(ctx Ctx, actionIDs []string) (allow bool, err error) {
	if err = ctx.Validate(actionIDs); err != nil {
		return false, err
	}
	perms := map[string]bool{}
	if ctx.ForceRaise() {
		for _, actionID := range actionIDs {
			perms[actionID] = false
		}
	} else {
		resReq := p.makeResReq(ctx)
		perms, _ = p.Cli.ResInstMultiActionsAllowed(
			ctx.GetUsername(), actionIDs, resReq.MakeResources([]string{ctx.GetResID()}),
		)
	}
	return p.canMultiActions(ctx, perms)
}

// BatchResMultiActionAllowed 判断用户对某些资源是否具有多个指定操作的权限. 当前 sdk 仅支持同类型的资源
func (p *IAMPerm) BatchResMultiActionAllowed(
	username string, actionIDs []string, resIDs []string, resRequest ResRequest,
) (map[string]map[string]bool, error) {
	return p.Cli.BatchResMultiActionsAllowed(username, actionIDs, resRequest.MakeResources(resIDs))
}

func (p *IAMPerm) hasParentRes() bool {
	return p.ParentResPerm != nil
}

func (p *IAMPerm) makeResReq(ctx Ctx) ResRequest {
	return p.ResReq.FormMap(ctx.ToMap())
}

func (p *IAMPerm) canAction(ctx Ctx, actionID string, useCache bool) (bool, error) {
	if resID := ctx.GetResID(); resID != "" {
		reqReq := p.makeResReq(ctx)
		resources := reqReq.MakeResources([]string{resID})
		return p.Cli.ResInstAllowed(ctx.GetUsername(), actionID, resources, useCache)
	}

	if !p.hasParentRes() {
		return p.Cli.ResTypeAllowed(ctx.GetUsername(), actionID, useCache)
	}

	pPermCtx := p.ParentResPerm.PermCtx.FromMap(ctx.ToMap())
	resReq := p.ParentResPerm.makeResReq(pPermCtx)
	resources := resReq.MakeResources([]string{pPermCtx.GetResID()})
	return p.Cli.ResInstAllowed(pPermCtx.GetUsername(), actionID, resources, useCache)
}

func (p *IAMPerm) canMultiActions(ctx Ctx, perms map[string]bool) (bool, error) {
	messages := []string{}
	actionReqList := []ActionResourcesRequest{}

	for actionID, allow := range perms {
		if allow {
			continue
		}

		err := p.genIAMPermError(ctx, actionID).(*IAMPermError)
		messages = append(messages, err.Msg)
		actionReqList = append(actionReqList, err.ActionReqList...)
	}

	if len(messages) == 0 {
		return true, nil
	}
	return false, &IAMPermError{
		Code:          errcode.NoIAMPerm,
		Username:      ctx.GetUsername(),
		Msg:           strings.Join(messages, "; "),
		ActionReqList: actionReqList,
	}
}

// 生成权限中心校验异常错误，包含缺失权限及申请链接
func (p *IAMPerm) genIAMPermError(ctx Ctx, actionID string) error {
	resType := p.ResType
	resIDs := []string{}
	parentChain := []IAMRes{}

	if resID := ctx.GetResID(); resID != "" {
		resIDs = append(resIDs, resID)
		parentChain = ctx.GetParentChain()
	} else if p.hasParentRes() {
		resType = p.ParentResPerm.ResType
		pPermCtx := p.ParentResPerm.PermCtx.FromMap(ctx.ToMap())
		resIDs = append(resIDs, pPermCtx.GetResID())
		parentChain = pPermCtx.GetParentChain()
	}

	return NewIAMPermErr(ctx.GetUsername(), "no "+actionID+" permission", []ActionResourcesRequest{
		{actionID, resType, resIDs, parentChain},
	})
}
