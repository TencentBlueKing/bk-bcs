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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/validator"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// Credential defines a credential's detail information
type Credential struct {
	// ID is an auto-increased value, which is a unique identity of a Credential.
	ID         uint32                `json:"id" gorm:"primaryKey"`
	Spec       *CredentialSpec       `json:"spec" gorm:"embedded"`
	Attachment *CredentialAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision             `json:"revision" gorm:"embedded"`
}

// TableName  is the Credential's database table name.
func (c *Credential) TableName() string {
	return "credentials"
}

// AppID AuditRes interface
func (c *Credential) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (c *Credential) ResID() uint32 {
	return c.ID
}

// ResType AuditRes interface
func (c *Credential) ResType() string {
	return "credential"
}

// ValidateCreate validate Credential is valid or not when create it.
func (c *Credential) ValidateCreate(kit *kit.Kit) error {

	if c.ID > 0 {
		return errors.New("id should not be set")
	}

	if c.Spec == nil {
		return errors.New("spec not set")
	}

	if err := c.Spec.ValidateCreate(kit); err != nil {
		return err
	}

	if c.Attachment == nil {
		return errors.New("attachment not set")
	}

	if err := c.Attachment.Validate(); err != nil {
		return err
	}

	if c.Revision == nil {
		return errors.New("revision not set")
	}

	if err := c.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// CredentialSpec defines all the specifics for credential set by user.
type CredentialSpec struct {
	CredentialType CredentialType `json:"credential_type" gorm:"column:credential_type"`
	EncCredential  string         `json:"enc_credential" gorm:"column:enc_credential"`
	EncAlgorithm   string         `json:"enc_algorithm" gorm:"column:enc_algorithm"`
	Name           string         `json:"name" gorm:"column:name"`
	Memo           string         `json:"memo" gorm:"column:memo"`
	Enable         bool           `json:"enable" gorm:"column:enable"`
	ExpiredAt      time.Time      `json:"expired_at" gorm:"column:expired_at"`
}

const (
	// BearToken is the type default
	BearToken CredentialType = "bearToken"
)

// CredentialType is the type of credential
type CredentialType string

// Validate validate the credential type
func (s CredentialType) Validate() error {
	if s == "" {
		return nil
	}
	switch s {
	case BearToken:
	default:
		return fmt.Errorf("unsupported credential type: %s", s)
	}

	return nil
}

// String credential to string
func (s CredentialType) String() string {
	return string(s)
}

// ValidateCreate validate credential spec when it is created.
func (c *CredentialSpec) ValidateCreate(kit *kit.Kit) error {
	if err := c.CredentialType.Validate(); err != nil {
		return err
	}
	if err := validator.ValidateName(kit, c.Name); err != nil {
		return err
	}
	return nil
}

// ValidateUpdate validate credential spec when it is updated.
func (c *CredentialSpec) ValidateUpdate(kit *kit.Kit) error {
	if c.CredentialType != "" {
		return errors.New("credential type cannot be updated once created")
	}
	if c.EncAlgorithm != "" {
		return errors.New("enc algorithm cannot be updated once created")
	}
	if c.EncCredential != "" {
		return errors.New("enc credential cannot be updated once created")
	}
	if err := validator.ValidateName(kit, c.Name); err != nil {
		return err
	}
	return nil
}

// CredentialAttachment defines the credential attachments.
type CredentialAttachment struct {
	BizID uint32 `json:"biz_id" gorm:"column:biz_id"`
}

// IsEmpty test whether credential attachment is empty or not.
func (c *CredentialAttachment) IsEmpty() bool {
	return c.BizID == 0
}

// Validate whether credential attachment is valid or not.
func (c *CredentialAttachment) Validate() error {
	if c.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	return nil
}

// ValidateDelete validate the credential's info when delete it.
func (c *Credential) ValidateDelete() error {
	if c.ID <= 0 {
		return errors.New("credential id should be set")
	}

	if c.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	return nil
}

// ValidateUpdate validate Credential is valid or not when update it.
func (c *Credential) ValidateUpdate(kit *kit.Kit) error {

	if c.ID <= 0 {
		return errors.New("id should be set")
	}

	if c.Spec == nil {
		return errors.New("spec should be set")
	}

	if err := c.Spec.ValidateUpdate(kit); err != nil {
		return err
	}

	if c.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if c.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if c.Revision == nil {
		return errors.New("revision not set")
	}

	return nil
}
