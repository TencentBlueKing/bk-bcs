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
	"context"

	crAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	criam "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam"
	iamPerm "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm"
	clusterAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/cluster"
	nsAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// Validate 权限中心 V3 资源权限校验，支持 命名空间，命名空间域资源，集群域资源权限校验
func Validate(ctx context.Context, res, action, projectID, clusterID, namespace string) error {
	username := ctx.Value(ctxkey.UsernameKey).(string)
	p, pCtx := genPermAndCtx(res, action, username, projectID, clusterID, namespace)
	if allow, err := canAction(p, pCtx, action); err != nil {
		return err
	} else if !allow {
		return errorx.New(errcode.NoIAMPerm, "无指定操作权限")
	}
	return nil
}

// 生成权限中心鉴权所需的 Perm && Ctx
func genPermAndCtx(res, action, username, projectID, clusterID, namespace string) (iamPerm.Perm, iamPerm.Ctx) {
	// 上游逻辑中已经确保命名空间域的资源，传入的 Namespace 必定不为空，
	// 因此这里直接根据命名空间是否为空判断权限类型即可（命名空间类型除外）
	switch {
	case res == "namespaces":
		// 获取命名空间列表 / 创建命名空间，务必确保命名空间是空
		if action == crAction.List || action == crAction.Create {
			namespace = ""
		}
		return criam.NewNSPerm(projectID, clusterID), nsAuth.NewPermCtx(username, projectID, clusterID, namespace)
	case namespace == "":
		return criam.NewClusterScopedPerm(projectID), clusterAuth.NewPermCtx(username, projectID, clusterID)
	default:
		return criam.NewNSScopedPerm(projectID, clusterID), nsAuth.NewPermCtx(username, projectID, clusterID, namespace)
	}
}

// 根据指定动作进行权限校验
func canAction(perm iamPerm.Perm, permCtx iamPerm.Ctx, action string) (allow bool, err error) {
	switch action {
	case crAction.List:
		return perm.CanList(permCtx)
	case crAction.View:
		return perm.CanView(permCtx)
	case crAction.Create:
		return perm.CanCreate(permCtx)
	case crAction.Update:
		return perm.CanUpdate(permCtx)
	case crAction.Delete:
		return perm.CanDelete(permCtx)
	case crAction.Use:
		return perm.CanUse(permCtx)
	default:
		return false, errorx.New(errcode.Unsupported, "Action %s in perm validate unsupported", action)
	}
}
