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

package independent

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// AllocateNamespace implement for AllocateNamespace interface
// 独立集群不存在「未分配」命名空间，分配/变更归属语义仅适用于共享集群
func (c *IndependentNamespaceAction) AllocateNamespace(ctx context.Context,
	req *proto.AllocateNamespaceRequest, resp *proto.AllocateNamespaceResponse) error {
	return errorx.NewReadableErr(errorx.ParamErr, "独立集群不支持分配命名空间")
}
