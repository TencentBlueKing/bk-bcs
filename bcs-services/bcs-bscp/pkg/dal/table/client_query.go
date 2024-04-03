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
)

// ClientQuery is a client query
type ClientQuery struct {
	ID         uint32                 `gorm:"column:id" json:"id"`
	Attachment *ClientQueryAttachment `json:"attachment" gorm:"embedded"`
	Spec       *ClientQuerySpec       `json:"spec" gorm:"embedded"`
}

// ClientQuerySpec is a client query spec
type ClientQuerySpec struct {
	Creator         string     `gorm:"column:creator"`
	SearchName      string     `gorm:"column:search_name"`
	SearchType      SearchType `gorm:"column:search_type"`
	SearchCondition string     `gorm:"column:search_condition"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

// ClientQueryAttachment is a client query attachment
type ClientQueryAttachment struct {
	BizID uint32 `gorm:"column:biz_id"`
	AppID uint32 `gorm:"column:app_id"`
}

// SearchType define the search type structure
type SearchType string

const (
	// Recent 最近查询
	Recent SearchType = "recent"
	// Common 常用查询
	Common SearchType = "common"
)

// Validate the search type is valid or not.
func (st SearchType) Validate() error {
	switch st {
	case Recent:
	case Common:
	default:
		return fmt.Errorf("unknown %s search type", st)
	}

	return nil
}

// ValidateCreate validate client query info when created.
func (c *ClientQuery) ValidateCreate() error {
	if c.ID != 0 {
		return errors.New("id can not be set")
	}

	if c.Attachment.BizID <= 0 {
		return errors.New("invalid biz id")
	}

	if c.Attachment.AppID <= 0 {
		return errors.New("invalid app id")
	}

	if c.Spec == nil {
		return errors.New("invalid spec, is nil")
	}

	if err := c.Spec.SearchType.Validate(); err != nil {
		return err
	}

	if len(c.Spec.SearchCondition) <= 2 {
		return errors.New("invalid search condition, is nil")
	}

	if len(c.Spec.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}

// ValidateUpdate validate client query info when update.
func (c *ClientQuery) ValidateUpdate() error {
	if c.ID <= 0 {
		return errors.New("id is not set")
	}

	if c.Attachment.BizID <= 0 {
		return errors.New("invalid biz id")
	}

	if c.Attachment.AppID <= 0 {
		return errors.New("invalid app id")
	}

	if c.Spec == nil {
		return errors.New("invalid spec, is nil")
	}

	if len(c.Spec.SearchCondition) <= 2 {
		return errors.New("invalid search condition, is nil")
	}

	if len(c.Spec.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}

// ValidateDelete validate client query info when delete.
func (c *ClientQuery) ValidateDelete() error {
	if c.ID <= 0 {
		return errors.New("id is not set")
	}

	if c.Attachment.BizID <= 0 {
		return errors.New("invalid biz id")
	}

	if c.Attachment.AppID <= 0 {
		return errors.New("invalid app id")
	}

	return nil
}

// TableName is the client_search database table name.
func (c *ClientQuery) TableName() string {
	return "client_querys"
}

// AppID KvRes interface
func (c *ClientQuery) AppID() uint32 {
	return c.Attachment.AppID
}

// ResID KvRes interface
func (c *ClientQuery) ResID() uint32 {
	return c.ID
}

// ResType KvRes interface
func (c *ClientQuery) ResType() string {
	return "client_query"
}
