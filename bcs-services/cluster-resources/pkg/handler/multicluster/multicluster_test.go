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

package multicluster

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
)

func TestCheckAccess(t *testing.T) {
	runtime.RunMode = runmode.UnitTest
	ctx := context.Background()

	// 独立集群不做共享集群校验
	assert.Nil(t, checkAccess(ctx, "", "BCS-K8S-00000", resCsts.Secret, nil))

	// 共享集群中 SC 等集群域资源允许访问
	assert.Nil(t, checkAccess(ctx, "", envs.TestSharedClusterID, resCsts.SC, nil))

	// 共享集群中不允许的资源类型
	assert.NotNil(t, checkAccess(ctx, "default", envs.TestSharedClusterID, resCsts.ClusterRoleBinding, nil))

	// 共享集群中命名空间域资源，命名空间为空时拒绝
	assert.NotNil(t, checkAccess(ctx, "", envs.TestSharedClusterID, resCsts.Secret, nil))
	assert.NotNil(t, checkAccess(ctx, "", envs.TestSharedClusterID, resCsts.Secret, map[string]interface{}{
		"kind":     resCsts.Secret,
		"metadata": map[string]interface{}{"name": "test"},
	}))
}
