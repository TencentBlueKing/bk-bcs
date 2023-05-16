/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package table

import (
	"errors"
	"time"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/runtime/selector"
)

// ReleasedGroupColumns defines group app's columns
var ReleasedGroupColumns = mergeColumns(ReleasedGroupColumnDescriptor)

// ReleasedGroupColumnDescriptor is CurrentRelease's column descriptors.
var ReleasedGroupColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{
		{Column: "id", NamedC: "id", Type: enumor.Numeric},
		{Column: "group_id", NamedC: "group_id", Type: enumor.Numeric},
		{Column: "app_id", NamedC: "app_id", Type: enumor.Numeric},
		{Column: "release_id", NamedC: "release_id", Type: enumor.Numeric},
		{Column: "strategy_id", NamedC: "strategy_id", Type: enumor.Numeric},
		{Column: "mode", NamedC: "mode", Type: enumor.String},
		{Column: "selector", NamedC: "selector", Type: enumor.String},
		{Column: "uid", NamedC: "uid", Type: enumor.String},
		{Column: "edited", NamedC: "edited", Type: enumor.Boolean},
		{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
		{Column: "reviser", NamedC: "reviser", Type: enumor.String},
		{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
	})

// ReleasedGroup defines a basic configuration item
type ReleasedGroup struct {
	// ID is an auto-increased value, which is a group app's
	// unique identity.
	ID         uint32             `db:"id" json:"id"`
	GroupID    uint32             `db:"group_id" json:"group_id"`
	AppID      uint32             `db:"app_id" json:"app_id"`
	ReleaseID  uint32             `db:"release_id" json:"release_id"`
	StrategyID uint32             `db:"strategy_id" json:"strategy_id"`
	Mode       GroupMode          `db:"mode" json:"mode"`
	Selector   *selector.Selector `db:"selector" json:"selector"`
	UID        string             `db:"uid" json:"uid"`
	Edited     bool               `db:"edited" json:"edited"`
	BizID      uint32             `db:"biz_id" json:"biz_id"`
	Reviser    string             `db:"reviser" json:"reviser"`
	UpdatedAt  time.Time          `db:"updated_at" json:"updated_at"`
}

// TableName is the group app's database table name.
func (c ReleasedGroup) TableName() Name {
	return ReleasedGroupTable
}

// ValidateCreate validate the group app's specific when create it.
func (c ReleasedGroup) ValidateCreate() error {
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
	if err := c.Mode.Validate(); err != nil {
		return err
	}
	if c.Mode == Custom && c.Selector == nil {
		return errors.New("selector should be set when mode is custom")
	}
	if c.Mode == Debug && c.UID == "" {
		return errors.New("uid should be set when mode is debug")
	}
	if c.UpdatedAt.IsZero() {
		return errors.New("updated_at should be set")
	}

	return nil
}

// ValidateUpdate validate the group app's specific when update it.
func (c ReleasedGroup) ValidateUpdate() error {
	if c.ID <= 0 {
		return errors.New("group app id should be set")
	}
	if c.BizID <= 0 {
		return errors.New("biz id should be set")
	}
	if c.UpdatedAt.IsZero() {
		return errors.New("updated_at should be set")
	}

	return nil
}

// ValidateDelete validate the group app's info when delete it.
func (c ReleasedGroup) ValidateDelete() error {
	if c.ID <= 0 {
		return errors.New("group app id should be set")
	}
	if c.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	return nil
}
