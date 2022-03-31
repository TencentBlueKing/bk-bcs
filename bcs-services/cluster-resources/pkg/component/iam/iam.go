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

package iam

import (
	"context"

	bcsAuth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth"
	clusterAuth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	nsAuth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"

	crAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// ScopedResPermValidate ...
func ScopedResPermValidate(ctx context.Context, action, projectID, clusterID, namespace string) (err error) {
	permCtx := bcsAuth.ScopedResPermCtx{
		Username:  ctx.Value(ctxkey.UsernameKey).(string),
		ProjectID: projectID,
		ClusterID: clusterID,
		Namespace: namespace,
	}
	var perm bcsAuth.ScopedResPerm
	// 上游逻辑中已经确保命名空间域的资源，传入的 Namespace 必定不为空，因此这里直接根据命名空间判断权限类型即可
	if namespace == "" {
		perm = getClusterScopedPerm()
	} else {
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

func getNSScopedPerm() bcsAuth.ScopedResPerm {
	if runtime.RunMode == runmode.Dev || runtime.RunMode == runmode.UnitTest {
		return &MockScopedResPerm{}
	}
	return nsAuth.NewNamespaceScopedPerm(conf.G.IAM.Cli)
}

func getClusterScopedPerm() bcsAuth.ScopedResPerm {
	if runtime.RunMode == runmode.Dev || runtime.RunMode == runmode.UnitTest {
		return &MockScopedResPerm{}
	}
	return clusterAuth.NewClusterScopedPerm(conf.G.IAM.Cli)
}
