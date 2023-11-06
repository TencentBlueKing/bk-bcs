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

	"bscp.io/pkg/criteria/validator"
)

// Kv defines a basic kv
type Kv struct {
	// ID is an auto-increased value, which is a unique identity of a kv.
	ID         uint32        `json:"id" gorm:"primaryKey"`
	Spec       *KvSpec       `json:"spec" gorm:"embedded"`
	Attachment *KvAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision     `json:"revision" gorm:"embedded"`
}

// KvSpec is kv specific which is defined by user.
type KvSpec struct {
	Key     string `json:"key" gorm:"column:key"`
	Version uint32 `json:"version" gorm:"column:version"`
}

// KvAttachment is a kv attachment
type KvAttachment struct {
	BizID uint32 `db:"biz_id" gorm:"column:biz_id"`
	AppID uint32 `db:"app_id" gorm:"column:app_id"`
}

// TableName is the kv database table name.
func (k *Kv) TableName() string {
	return "kvs"
}

// AppID KvRes interface
func (k *Kv) AppID() uint32 {
	return k.Attachment.AppID
}

// ResID KvRes interface
func (k *Kv) ResID() uint32 {
	return k.ID
}

// ResType KvRes interface
func (k *Kv) ResType() string {
	return "kv"
}

// ValidateCreate validate kv is valid or not when create it.
func (k Kv) ValidateCreate() error {

	if k.ID > 0 {
		return errors.New("id should not be set")
	}

	if k.Spec == nil {
		return errors.New("spec not set")
	}

	if err := k.Spec.ValidateCreate(); err != nil {
		return err
	}

	if k.Attachment == nil {
		return errors.New("attachment not set")
	}

	if err := k.Attachment.Validate(); err != nil {
		return err
	}

	if k.Revision == nil {
		return errors.New("revision not set")
	}

	if err := k.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateCreate validate kv spec when it is created.
func (k KvSpec) ValidateCreate() error {
	if err := validator.ValidateName(k.Key); err != nil {
		return err
	}

	return nil
}

// Validate whether kv attachment is valid or not.
func (a KvAttachment) Validate() error {
	if a.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	if a.AppID <= 0 {
		return errors.New("invalid attachment app id")
	}

	return nil
}

// ValidateCreate validate kv spec when it is created.
func (a KvAttachment) ValidateCreate() error {
	return nil
}

// ValidateDelete validate the kv's info when delete it.
func (k *Kv) ValidateDelete() error {
	if k.ID <= 0 {
		return errors.New("credential id should be set")
	}

	if k.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if k.Attachment.AppID <= 0 {
		return errors.New("app id should be set")
	}

	return nil
}

// ValidateUpdate validate Kv is valid or not when update it.
func (k *Kv) ValidateUpdate() error {

	if k.ID <= 0 {
		return errors.New("id should be set")
	}

	if k.Spec == nil {
		return errors.New("spec should be set")
	}

	if k.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if k.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}
	if k.Attachment.AppID <= 0 {
		return errors.New("app id should be set")
	}

	if k.Revision == nil {
		return errors.New("revision not set")
	}

	return nil
}
