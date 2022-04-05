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

package client

import (
	"context"

	crAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam"
	clusterAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/cluster"
	nsAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// PermValidate ...
func PermValidate(ctx context.Context, res, action, projectID, clusterID, namespace string) (err error) {
	permCtx := iam.NewPermCtx(ctx.Value(ctxkey.UsernameKey).(string), projectID, clusterID, namespace)
	var perm iam.Perm
	// 上游逻辑中已经确保命名空间域的资源，传入的 Namespace 必定不为空，
	// 因此这里直接根据命名空间是否为空判断权限类型即可（命名空间类型除外）
	switch {
	case res == "namespaces":
		perm = getNSPerm()
	case namespace == "":
		perm = getClusterScopedPerm()
	default:
		perm = getNSScopedPerm()
	}
	switch action {
	case crAction.View:
		_, err = perm.CanView(permCtx)
	case crAction.Create:
		_, err = perm.CanCreate(permCtx)
	case crAction.Update:
		_, err = perm.CanUpdate(permCtx)
	case crAction.Delete:
		_, err = perm.CanDelete(permCtx)
	case crAction.Use:
		_, err = perm.CanUse(permCtx)
	default:
		return errorx.New(
			errcode.Unsupported, "Action %s in scoped resource perm validate unsupported", action,
		)
	}
	return err
}

func getNSPerm() iam.Perm {
	if runtime.RunMode == runmode.Dev || runtime.RunMode == runmode.UnitTest {
		return &iam.MockPerm{}
	}
	return nsAuth.NewPerm()
}

func getNSScopedPerm() iam.Perm {
	if runtime.RunMode == runmode.Dev || runtime.RunMode == runmode.UnitTest {
		return &iam.MockPerm{}
	}
	return nsAuth.NewScopedPerm()
}

func getClusterScopedPerm() iam.Perm {
	if runtime.RunMode == runmode.Dev || runtime.RunMode == runmode.UnitTest {
		return &iam.MockPerm{}
	}
	return clusterAuth.NewScopedPerm()
}
