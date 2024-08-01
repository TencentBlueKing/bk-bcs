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
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/validator"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// Kv defines a basic kv
type Kv struct {
	// ID is an auto-increased value, which is a unique identity of a kv.
	ID          uint32        `json:"id" gorm:"primaryKey"`
	KvState     KvState       `json:"kv_state" gorm:"column:kv_state"`
	Spec        *KvSpec       `json:"spec" gorm:"embedded"`
	Attachment  *KvAttachment `json:"attachment" gorm:"embedded"`
	Revision    *Revision     `json:"revision" gorm:"embedded"`
	ContentSpec *ContentSpec  `json:"content_spec" gorm:"embedded"`
}

// KvSpec is kv specific which is defined by user.
type KvSpec struct {
	Key     string   `json:"key" gorm:"column:key"`
	Memo    string   `json:"memo" gorm:"column:memo"`
	KvType  DataType `json:"kv_type" gorm:"column:kv_type"`
	Version uint32   `json:"version" gorm:"column:version"`
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
func (k Kv) ValidateCreate(kit *kit.Kit) error {

	if k.ID > 0 {
		return errors.New("id should not be set")
	}

	if k.KvState != KvStateAdd {
		return errors.New("KvState is not set to Add")

	}

	if k.Spec == nil {
		return errors.New("spec not set")
	}

	if err := k.Spec.ValidateCreate(kit); err != nil {
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
func (k KvSpec) ValidateCreate(kit *kit.Kit) error {
	if err := validator.ValidateName(kit, k.Key); err != nil {
		return err
	}

	if err := k.KvType.ValidateCreateKv(); err != nil {
		return err
	}

	return nil
}

// ValidateCreateKv the kvType and value match
func (k DataType) ValidateCreateKv() error {

	switch k {
	case KvStr:
	case KvNumber:
	case KvText:
	case KvJson:
	case KvYAML:
	case KvXml:
	default:
		return errors.New("invalid data-type")
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
		return errors.New("kv id should be set")
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

const (
	// MaxValueLength max value length 1MB
	MaxValueLength = 1 * 1024 * 1024
)

// ValidateValue the kvType and value match
func (k DataType) ValidateValue(value string) error {

	if value == "" {
		return errors.New("kv value is null")
	}

	if len(value) > MaxValueLength {
		return fmt.Errorf("the length of the value must not exceed %d MB", MaxValueLength)
	}

	switch k {
	case KvStr:
		if strings.Contains(value, "\n") {
			return errors.New("newline characters are not allowed in string-type values")
		}
		return nil
	case KvNumber:
		if !tools.IsNumber(value) {
			return fmt.Errorf("value is not a number")
		}
		return nil
	case KvText:
		return nil
	case KvJson:
		if !json.Valid([]byte(value)) {
			return fmt.Errorf("value is not a json")
		}
		return nil
	case KvYAML:
		var data interface{}
		if err := yaml.Unmarshal([]byte(value), &data); err != nil {
			return fmt.Errorf("value is not a yaml, err: %v", err)
		}
		return nil
	case KvXml:
		var v interface{}
		if err := xml.Unmarshal([]byte(value), &v); err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid key-value type")
	}
}

// KvState ....
type KvState string

const (
	// KvStateAdd 增加
	KvStateAdd KvState = "ADD"
	// KvStateDelete 删除
	KvStateDelete KvState = "DELETE"
	// KvStateRevise 修改
	KvStateRevise KvState = "REVISE"
	// KvStateUnchange 不变
	KvStateUnchange KvState = "UNCHANGE"
)

// String get string value of KvState
func (k KvState) String() string {
	return string(k)
}

// Validate validate kv state is valid or not.
func (k KvState) Validate() error {
	switch k {
	case KvStateAdd, KvStateDelete, KvStateRevise, KvStateUnchange:
		return nil
	default:
		return errors.New("invalid kv state")
	}
}
