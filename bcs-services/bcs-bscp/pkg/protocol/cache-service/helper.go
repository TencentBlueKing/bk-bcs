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

package pbcs

import (
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
)

// Validate BenchAppMetaReq.
func (m *BenchAppMetaReq) Validate() error {
	if m.BizId == 0 {
		return errf.New(errf.InvalidParameter, "invalid biz_id, biz_id should > 0")
	}

	if len(m.AppIds) == 0 {
		return errf.New(errf.InvalidParameter, "invalid app_ids, app_ids need at least one")
	}

	return nil
}

// Validate BenchReleasedCIReq.
func (m *BenchReleasedCIReq) Validate() error {
	if m.BizId == 0 {
		return errf.New(errf.InvalidParameter, "invalid biz_id, biz_id should > 0")
	}

	if m.ReleaseId == 0 {
		return errf.New(errf.InvalidParameter, "invalid release_id, release_id should > 0")
	}

	return nil
}
