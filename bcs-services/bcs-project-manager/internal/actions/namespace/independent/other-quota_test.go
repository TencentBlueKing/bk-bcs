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
	"testing"

	"github.com/stretchr/testify/require"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

func TestValidateOtherQuotaRequest(t *testing.T) {
	validQuota := &proto.ResourceQuota{CpuLimits: "2"}

	require.NoError(t, validateOtherQuotaRequest("test-ns", "extra-quota", validQuota))
	require.ErrorContains(t, validateOtherQuotaRequest("test-ns", "test-ns", validQuota),
		"其他配额名称不能与命名空间名称相同")
	require.ErrorContains(t, validateOtherQuotaRequest("test-ns", "extra-quota", nil),
		"quota cannot be empty")
	require.ErrorContains(t, validateOtherQuotaRequest("test-ns", "extra-quota", &proto.ResourceQuota{}),
		"quota resources cannot all be empty")
}
