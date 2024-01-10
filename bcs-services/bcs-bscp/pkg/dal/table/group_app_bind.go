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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/enumor"
)

// GroupAppBindColumns defines group app's columns
var GroupAppBindColumns = mergeColumns(GroupAppBindColumnDescriptor)

// GroupAppBindColumnDescriptor is GroupApp's column descriptors.
var GroupAppBindColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{
		{Column: "id", NamedC: "id", Type: enumor.Numeric},
		{Column: "group_id", NamedC: "group_id", Type: enumor.Numeric},
		{Column: "app_id", NamedC: "app_id", Type: enumor.Numeric},
		{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
	})

// GroupAppBind defines a basic configuration item
type GroupAppBind struct {
	// ID is an auto-increased value, which is a group app's
	// unique identity.
	ID      uint32 `db:"id" json:"id" gorm:"primaryKey"`
	GroupID uint32 `db:"group_id" json:"group_id" gorm:"column:group_id"`
	AppID   uint32 `db:"app_id" json:"app_id" gorm:"column:app_id"`
	BizID   uint32 `db:"biz_id" json:"biz_id" gorm:"column:biz_id"`
}

// TableName is the group app's database table name.
func (c GroupAppBind) TableName() string {
	return "group_app_binds"
}

// ValidateCreate validate the group app's specific when create it.
func (c GroupAppBind) ValidateCreate() error {
	if c.ID != 0 {
		return errors.New("group app id can not be set")
	}

	if c.GroupID <= 0 {
		return errors.New("group id should be set")
	}
	if c.AppID <= 0 {
		return errors.New("app id should be set")
	}
	if c.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	return nil
}

// ValidateUpdate validate the group app's specific when update it.
func (c GroupAppBind) ValidateUpdate() error {
	if c.ID <= 0 {
		return errors.New("group app id should be set")
	}

	return nil
}

// ValidateDelete validate the group app's info when delete it.
func (c GroupAppBind) ValidateDelete() error {
	if c.ID <= 0 {
		return errors.New("group app id should be set")
	}

	return nil
}
