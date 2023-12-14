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
	"time"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/credential"
)

// CredentialScope defines CredentialScope's columns
type CredentialScope struct {
	// ID is an auto-increased value, which is a unique identity of a Credential.
	ID         uint32                     `json:"id" gorm:"primaryKey"`
	Spec       *CredentialScopeSpec       `json:"spec" gorm:"embedded"`
	Attachment *CredentialScopeAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision                  `json:"revision" gorm:"embedded"`
}

// TableName is the CredentialScope's database table name.
func (c *CredentialScope) TableName() string {
	return "credential_scopes"
}

// AppID AuditRes interface
func (c *CredentialScope) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (c *CredentialScope) ResID() uint32 {
	return c.ID
}

// ResType AuditRes interface
func (c *CredentialScope) ResType() string {
	return "credential_scope"
}

// ValidateCreate validate Credential is valid or not when create it.
func (c *CredentialScope) ValidateCreate() error {

	if c.ID > 0 {
		return errors.New("id should not be set")
	}

	if c.Spec == nil {
		return errors.New("spec not set")
	}

	if err := c.Spec.CredentialScope.Validate(); err != nil {
		return err
	}

	if c.Attachment == nil {
		return errors.New("attachment not set")
	}

	if c.Revision == nil {
		return errors.New("revision not set")
	}

	if err := c.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// CredentialScopeSpec defines credential scope's Spec
type CredentialScopeSpec struct {
	CredentialScope credential.Scope `json:"credential_scope" gorm:"column:credential_scope"`
	ExpiredAt       time.Time        `json:"expired_at" gorm:"column:expired_at"`
}

// CredentialScopeAttachment defines the credential scope attachments.
type CredentialScopeAttachment struct {
	BizID        uint32 `json:"biz_id" gorm:"column:biz_id"`
	CredentialId uint32 `json:"credential_id" gorm:"column:credential_id"`
}

// ValidateDelete credential scope validate
func (c *CredentialScope) ValidateDelete() error {
	if c.ID <= 0 {
		return errors.New("credential scope id should be set")
	}

	if c.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	return nil
}

// ValidateUpdate validate Credential is valid or not when update it.
func (c *CredentialScope) ValidateUpdate() error {

	if c.ID <= 0 {
		return errors.New("credential scope id should be set")
	}

	if c.Spec == nil {
		return errors.New("spec not set")
	}

	if err := c.Spec.CredentialScope.Validate(); err != nil {
		return err
	}

	if c.Attachment == nil {
		return errors.New("attachment not set")
	}

	if c.Revision == nil {
		return errors.New("revision not set")
	}

	return nil
}
