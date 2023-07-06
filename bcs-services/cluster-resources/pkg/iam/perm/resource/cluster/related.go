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

package cluster

import (
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// RelatedClusterCanViewPerm 关联集群查看权限，用于诸如命名空间查看，命名空间域操作等权限
func RelatedClusterCanViewPerm(ctx perm.Ctx, subAllow bool, subErr error) (bool, error) {
	clusterPermCtx := PermCtx{}
	if subErr != nil {
		// 按照权限中心的建议，无论关联资源操作是否有权限，统一按照无权限返回，目的是生成最终的 ApplyURL
		if se, ok := subErr.(*perm.IAMPermError); ok {
			clusterPermCtx.SetForceRaise()
			// 强制无权限，因此 err 必定不等于 nil
			_, err := NewPerm(ctx.GetProjID()).CanView(clusterPermCtx.FromMap(ctx.ToMap()))
			if e, ok := err.(*perm.IAMPermError); ok {
				actionReqList := append(se.ActionReqList, e.ActionReqList...)
				return false, perm.NewIAMPermErr(ctx.GetUsername(), se.Msg+"; "+e.Msg, actionReqList)
			}
			return false, err
		}
		return false, subErr
	}
	// 若没有错误，但是 allow 仍然是 false，说明 iam 不存在对应的策略
	if !subAllow {
		return false, errorx.New(errcode.NoIAMPerm, "not iam permission for current operate")
	}
	// TODO 考虑下要不要加上 View 类型的，跳过鉴权？
	return NewPerm(ctx.GetProjID()).CanView(clusterPermCtx.FromMap(ctx.ToMap()))
}
