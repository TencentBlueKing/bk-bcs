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

// Package pbrkv provides protocol definitions and conversion functions for releasedKv key-value related operations.
package pbrkv

import (
	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbkv "bscp.io/pkg/protocol/core/kv"
	"bscp.io/pkg/types"
)

// PbRKv convert table ReleasedKv to pb ReleasedKv
func PbRKv(k *table.ReleasedKv, kvType types.KvType, value string) *ReleasedKv {
	if k == nil {
		return nil
	}

	return &ReleasedKv{
		Id:         k.ID,
		ReleaseId:  k.ReleaseID,
		Spec:       pbkv.PbKvSpec(k.Spec, kvType, value),
		Attachment: pbkv.PbKvAttachment(k.Attachment),
		Revision:   pbbase.PbRevision(k.Revision),
	}
}
