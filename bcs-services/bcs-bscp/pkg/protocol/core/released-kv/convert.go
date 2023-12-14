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
	"fmt"
	"time"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbkv "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/kv"
)

// PbRKv convert table ReleasedKv to pb ReleasedKv
func PbRKv(k *table.ReleasedKv, value string) *ReleasedKv {
	if k == nil {
		return nil
	}

	return &ReleasedKv{
		Id:         k.ID,
		ReleaseId:  k.ReleaseID,
		Spec:       pbkv.PbKvSpec(k.Spec, value),
		Attachment: pbkv.PbKvAttachment(k.Attachment),
		Revision:   pbbase.PbRevision(k.Revision),
	}
}

// RKvs convert pb kvs to table Rkvs
func RKvs(kvs []*pbkv.Kv, versionMap map[string]int, releaseID uint32) ([]*table.ReleasedKv, error) {

	var rkvs []*table.ReleasedKv

	for _, kv := range kvs {

		createdAt, err := time.Parse(time.RFC3339, kv.Revision.UpdateAt)
		if err != nil {
			return nil, fmt.Errorf("parse time from createAt string failed, err: %v", err)
		}

		rkv := &table.ReleasedKv{
			ReleaseID:  releaseID,
			Spec:       kv.Spec.KvSpec(),
			Attachment: kv.Attachment.KvAttachment(),
			Revision: &table.Revision{
				Creator:   kv.Revision.Reviser,
				CreatedAt: createdAt,
			},
		}
		rkv.Spec.Version = uint32(versionMap[kv.Spec.Key])
		rkvs = append(rkvs, rkv)
	}

	return rkvs, nil
}
