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

// Package pbkv provides protocol definitions and conversion functions for key-value related operations.
package pbkv

import (
	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
)

// Kv convert pb Kv to table Kv
func (k *Kv) Kv() (*table.Kv, error) {
	if k == nil {
		return nil, nil
	}

	return &table.Kv{
		ID:         k.Id,
		Spec:       k.Spec.KvSpec(),
		Attachment: k.Attachment.KvAttachment(),
	}, nil
}

// KvSpec convert pb kv to table KvSpec
func (k *KvSpec) KvSpec() *table.KvSpec {
	if k == nil {
		return nil
	}

	return &table.KvSpec{
		Key: k.Key,
	}
}

// KvAttachment convert pb KvAttachment to table KvAttachment
func (k *KvAttachment) KvAttachment() *table.KvAttachment {
	if k == nil {
		return nil
	}

	return &table.KvAttachment{
		BizID: k.BizId,
		AppID: k.AppId,
	}
}

// PbKv convert table kv to pb kv
func PbKv(k *table.Kv, kvType table.DataType, value string) *Kv {
	if k == nil {
		return nil
	}

	return &Kv{
		Id:         k.ID,
		Spec:       PbKvSpec(k.Spec, kvType, value),
		Attachment: PbKvAttachment(k.Attachment),
		Revision:   pbbase.PbRevision(k.Revision),
	}
}

// PbKvSpec convert table KvSpec to pb KvSpec
//
//nolint:revive
func PbKvSpec(spec *table.KvSpec, kvType table.DataType, value string) *KvSpec {
	if spec == nil {
		return nil
	}

	return &KvSpec{
		Key:    spec.Key,
		KvType: string(kvType),
		Value:  value,
	}
}

// PbKvAttachment convert table KvAttachment to pb KvAttachment
//
//nolint:revive
func PbKvAttachment(ka *table.KvAttachment) *KvAttachment {
	if ka == nil {
		return nil
	}

	return &KvAttachment{
		BizId: ka.BizID,
		AppId: ka.AppID,
	}
}
