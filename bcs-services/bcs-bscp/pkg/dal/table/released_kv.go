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

package table

import (
	"errors"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// ReleasedKv 已生成版本的kv
type ReleasedKv struct {
	ID uint32 `json:"id" gorm:"primaryKey"`

	// ReleaseID is this app's config item's release id
	ReleaseID uint32 `json:"release_id" gorm:"column:release_id"`

	Spec        *KvSpec       `json:"spec" gorm:"embedded"`
	Attachment  *KvAttachment `json:"attachment" gorm:"embedded"`
	Revision    *Revision     `json:"revision" gorm:"embedded"`
	ContentSpec *ContentSpec  `json:"content_spec" gorm:"embedded"`
}

// TableName is the ReleasedKv's database table name.
func (r *ReleasedKv) TableName() string {
	return "released_kvs"
}

// AppID AuditRes interface
func (r *ReleasedKv) AppID() uint32 {
	return r.Attachment.AppID
}

// ResID AuditRes interface
func (r *ReleasedKv) ResID() uint32 {
	return r.ID
}

// ResType AuditRes interface
func (r *ReleasedKv) ResType() string {
	return "released_kv"
}

// RkvList is released kvs
type RkvList []*ReleasedKv

// AppID AuditRes interface
func (rs RkvList) AppID() uint32 {
	if len(rs) > 0 {
		return rs[0].Attachment.AppID
	}
	return 0
}

// ResID AuditRes interface
func (rs RkvList) ResID() uint32 {
	if len(rs) > 0 {
		return rs[0].ID
	}
	return 0
}

// ResType AuditRes interface
func (rs RkvList) ResType() string {
	return "released_kv"
}

// ValidateCreate validate ReleasedKv is valid or not when create ir.
func (r *ReleasedKv) ValidateCreate(kit *kit.Kit) error {
	if r.ID > 0 {
		return errors.New("id should not be set")
	}

	if r.Spec == nil {
		return errors.New("spec not set")
	}

	if err := r.Spec.ValidateCreate(kit); err != nil {
		return err
	}

	if r.Attachment == nil {
		return errors.New("attachment not set")
	}

	if err := r.Attachment.Validate(); err != nil {
		return err
	}

	if r.Revision == nil {
		return errors.New("revision not set")
	}

	return nil
}
