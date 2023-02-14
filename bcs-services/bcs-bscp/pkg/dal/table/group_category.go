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
	"fmt"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/criteria/validator"
)

// GroupCategoryColumns defines GroupCategory's columns
var GroupCategoryColumns = mergeColumns(GroupCategoryColumnDescriptor)

// GroupCategoryColumnDescriptor is GroupCategory's column descriptors.
var GroupCategoryColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{{Column: "id", NamedC: "id", Type: enumor.Numeric}},
	mergeColumnDescriptors("spec", GroupCategorySpecColumnDescriptor),
	mergeColumnDescriptors("attachment", GroupCategoryAttachColumnDescriptor),
	mergeColumnDescriptors("revision", CreatedRevisionColumnDescriptor))

// maxGroupCategorysLimitForApp defines the max limit of strategy set for an app for user to create.
const maxGroupCategorysLimitForApp = 1

// ValidateAppGroupCategoryNumber verify whether the current number of app strategies have reached the maximum.
func ValidateAppGroupCategoryNumber(count uint32) error {
	if count >= maxGroupCategorysLimitForApp {
		return errf.New(errf.InvalidParameter, fmt.Sprintf("an application only create %d strategy set",
			maxGroupCategorysLimitForApp))
	}
	return nil
}

// GroupCategory defines all the information for a strategy set.
type GroupCategory struct {
	// ID is an auto-increased value, which is a unique identity
	// of a strategy set.
	ID         uint32                   `db:"id" json:"id"`
	Spec       *GroupCategorySpec       `db:"spec" json:"spec"`
	Attachment *GroupCategoryAttachment `db:"attachment" json:"attachment"`
	Revision   *CreatedRevision         `db:"revision" json:"revision"`
}

// TableName is the strategy set's database table name.
func (s GroupCategory) TableName() Name {
	return GroupCategoryTable
}

// ValidateCreate the strategy set's information when create it.
func (s GroupCategory) ValidateCreate() error {
	if s.ID != 0 {
		return errors.New("id should not set")
	}

	if s.Spec == nil {
		return errors.New("spec is empty")
	}

	if err := s.Spec.ValidateCreate(); err != nil {
		return err
	}

	if s.Attachment == nil {
		return errors.New("attachment is empty")
	}

	if err := s.Attachment.Validate(); err != nil {
		return err
	}

	if s.Revision == nil {
		return errors.New("revision is empty")
	}

	if err := s.Revision.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate the strategy set's information when update it.
func (s GroupCategory) ValidateUpdate() error {
	if s.ID <= 0 {
		return errors.New("id not set")
	}

	if s.Spec == nil {
		return errors.New("spec is empty")
	}

	if err := s.Spec.ValidateUpdate(); err != nil {
		return err
	}

	if s.Attachment == nil {
		return errors.New("attachment is empty")
	}

	if s.Attachment.BizID <= 0 {
		return errors.New("attachment's biz id can not be update")
	}

	if s.Attachment.AppID <= 0 {
		return errors.New("attachment's app id can not be update")
	}

	if s.Revision == nil {
		return errors.New("revision is empty")
	}

	if err := s.Revision.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateDelete validate the strategy set's info when delete it.
func (s GroupCategory) ValidateDelete() error {
	if s.ID <= 0 {
		return errors.New("strategy set id should be set")
	}

	if s.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if s.Attachment.AppID <= 0 {
		return errors.New("app id should be set")
	}

	return nil
}

// GroupCategoryAttachColumns defines GroupCategoryAttachment's columns
var GroupCategoryAttachColumns = mergeColumns(GroupCategoryAttachColumnDescriptor)

// GroupCategoryAttachColumnDescriptor is GroupCategoryAttachment's column descriptors.
var GroupCategoryAttachColumnDescriptor = ColumnDescriptors{
	{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
	{Column: "app_id", NamedC: "app_id", Type: enumor.Numeric},
}

// GroupCategoryAttachment strategy set attachment info.
type GroupCategoryAttachment struct {
	BizID uint32 `db:"biz_id" json:"biz_id"`
	AppID uint32 `db:"app_id" json:"app_id"`
}

// Validate validate strategy set's attachment.
func (s GroupCategoryAttachment) Validate() error {
	if s.BizID <= 0 {
		return errors.New("invalid biz id")
	}

	if s.AppID <= 0 {
		return errors.New("invalid app id")
	}

	return nil
}

// GroupCategorySpecColumns defines GroupCategorySpec's columns
var GroupCategorySpecColumns = mergeColumns(GroupCategorySpecColumnDescriptor)

// GroupCategorySpecColumnDescriptor is GroupCategorySpec's column descriptors.
var GroupCategorySpecColumnDescriptor = ColumnDescriptors{
	{Column: "name", NamedC: "name", Type: enumor.String},
}

// GroupCategorySpec defines all the specifics for a strategy set, which
// is set by user.
type GroupCategorySpec struct {
	Name string `db:"name" json:"name"`
}

// ValidateCreate the strategy set specifics.
func (s GroupCategorySpec) ValidateCreate() error {
	if err := validator.ValidateName(s.Name); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate the strategy set specifics.
func (s GroupCategorySpec) ValidateUpdate() error {
	if err := validator.ValidateName(s.Name); err != nil {
		return err
	}

	return nil
}
