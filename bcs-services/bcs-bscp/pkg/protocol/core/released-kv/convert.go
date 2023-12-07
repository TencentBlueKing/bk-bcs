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
	"time"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbkv "bscp.io/pkg/protocol/core/kv"
)

// PbRKv convert table ReleasedKv to pb ReleasedKv
func PbRKv(k *table.ReleasedKv, kvType table.DataType, value string) *ReleasedKv {
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

// PbKvState convert kv state
func PbKvState(kvs []*table.Kv, kvRelease []*table.ReleasedKv) []*pbkv.Kv {

	releaseMap := make(map[string]*table.ReleasedKv, len(kvRelease))
	for _, release := range kvRelease {
		releaseMap[release.Spec.Key] = release
	}

	result := make([]*pbkv.Kv, 0)
	for _, kv := range kvs {
		var kvState string
		if len(kvRelease) == 0 {
			kvState = constant.KvStateAdd
		} else {
			if _, ok := releaseMap[kv.Spec.Key]; ok {
				if kv.Revision.UpdatedAt.After(releaseMap[kv.Spec.Key].Revision.CreatedAt) {
					kvState = constant.KvStateRevise
				} else {
					kvState = constant.KvStateUnchange
				}
				delete(releaseMap, kv.Spec.Key)
			}
		}
		if len(kvState) == 0 {
			kvState = constant.KvStateAdd
		}

		result = append(result, pbkv.PbKv(kv, "", "", kvState))
	}
	for _, kv := range releaseMap {
		kv.ID = 0
		result = append(result, PbKv(kv, constant.KvStateDelete))
	}
	return sortKvsByState(result)
}

// sortKvsByState sort as add > revise > unchange > delete
func sortKvsByState(kvs []*pbkv.Kv) []*pbkv.Kv {
	result := make([]*pbkv.Kv, 0)
	add := make([]*pbkv.Kv, 0)
	del := make([]*pbkv.Kv, 0)
	revise := make([]*pbkv.Kv, 0)
	unchange := make([]*pbkv.Kv, 0)
	for _, kv := range kvs {
		switch kv.KvState {
		case constant.KvStateAdd:
			add = append(add, kv)
		case constant.KvStateDelete:
			del = append(del, kv)
		case constant.KvStateRevise:
			revise = append(revise, kv)
		case constant.KvStateUnchange:
			unchange = append(unchange, kv)
		}
	}
	result = append(result, add...)
	result = append(result, revise...)
	result = append(result, unchange...)
	result = append(result, del...)
	return result
}

// PbKv convert table ReleasedKv to pb Kv
func PbKv(rkv *table.ReleasedKv, kvState string) *pbkv.Kv {
	if rkv == nil {
		return nil
	}

	return &pbkv.Kv{
		Id:         rkv.ID,
		KvState:    kvState,
		Spec:       pbkv.PbKvSpec(rkv.Spec, "", ""),
		Attachment: pbkv.PbKvAttachment(rkv.Attachment),
		Revision: &pbbase.Revision{
			Creator:  rkv.Revision.Creator,
			Reviser:  rkv.Revision.Creator,
			CreateAt: rkv.Revision.CreatedAt.Format(time.RFC3339),
			UpdateAt: rkv.Revision.CreatedAt.Format(time.RFC3339),
		},
	}
}
