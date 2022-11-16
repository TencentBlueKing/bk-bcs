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

// StrategySetColumns defines StrategySet's columns
var StrategySetColumns = mergeColumns(StrategySetColumnDescriptor)

// StrategySetColumnDescriptor is StrategySet's column descriptors.
var StrategySetColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{{Column: "id", NamedC: "id", Type: enumor.Numeric}},
	mergeColumnDescriptors("spec", StrategySetSpecColumnDescriptor),
	mergeColumnDescriptors("state", StrategySetStateColumnDescriptor),
	mergeColumnDescriptors("attachment", StrategySetAttachColumnDescriptor),
	mergeColumnDescriptors("revision", RevisionColumnDescriptor))

// maxStrategySetsLimitForApp defines the max limit of strategy set for an app for user to create.
const maxStrategySetsLimitForApp = 1

// ValidateAppStrategySetNumber verify whether the current number of app strategies have reached the maximum.
func ValidateAppStrategySetNumber(count uint32) error {
	if count >= maxStrategySetsLimitForApp {
		return errf.New(errf.InvalidParameter, fmt.Sprintf("an application only create %d strategy set",
			maxStrategySetsLimitForApp))
	}
	return nil
}

// StrategySet defines all the information for a strategy set.
type StrategySet struct {
	// ID is an auto-increased value, which is a unique identity
	// of a strategy set.
	ID         uint32                 `db:"id" json:"id"`
	Spec       *StrategySetSpec       `db:"spec" json:"spec"`
	State      *StrategySetState      `db:"state" json:"state"`
	Attachment *StrategySetAttachment `db:"attachment" json:"attachment"`
	Revision   *Revision              `db:"revision" json:"revision"`
}

// TableName is the strategy set's database table name.
func (s StrategySet) TableName() Name {
	return StrategySetTable
}

// ValidateCreate the strategy set's information when create it.
func (s StrategySet) ValidateCreate() error {
	if s.ID != 0 {
		return errors.New("id should not set")
	}

	if s.Spec == nil {
		return errors.New("spec is empty")
	}

	if err := s.Spec.ValidateCreate(); err != nil {
		return err
	}

	if s.State == nil {
		return errors.New("state not set")
	}

	if err := s.State.Validate(); err != nil {
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

	if err := s.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate the strategy set's information when update it.
func (s StrategySet) ValidateUpdate() error {
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

	if s.State != nil {
		if err := s.State.Validate(); err != nil {
			return err
		}
	}

	if s.Revision == nil {
		return errors.New("revision is empty")
	}

	if err := s.Revision.ValidateUpdate(); err != nil {
		return err
	}

	return nil
}

// ValidateDelete validate the strategy set's info when delete it.
func (s StrategySet) ValidateDelete() error {
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

const (
	// Enabled means this strategy set is enabled.
	Enabled StrategySetStatusType = "enabled"
	// Disabled means this strategy set is disabled.
	Disabled StrategySetStatusType = "disabled"
)

// StrategySetStatusType is a type which describe a strategy set's status.
type StrategySetStatusType string

// Validate whether a strategy set's status type is valid or not.
func (s StrategySetStatusType) Validate() error {
	switch s {
	case Enabled:
	case Disabled:
	default:
		return fmt.Errorf("unsupported strategy set status type: %s", s)
	}

	return nil
}

// StrategySetStateColumns defines StrategySetState's columns
var StrategySetStateColumns = mergeColumns(StrategySetStateColumnDescriptor)

// StrategySetStateColumnDescriptor is StrategySetState's column descriptors.
var StrategySetStateColumnDescriptor = ColumnDescriptors{{Column: "status", NamedC: "status", Type: enumor.String}}

// StrategySetState describe the strategy set's state.
type StrategySetState struct {
	// 是否启用该策略，
	Status StrategySetStatusType `db:"status" json:"status"`
}

// Validate the strategy set is valid or not.
func (s StrategySetState) Validate() error {
	return s.Status.Validate()
}

// StrategySetAttachColumns defines StrategySetAttachment's columns
var StrategySetAttachColumns = mergeColumns(StrategySetAttachColumnDescriptor)

// StrategySetAttachColumnDescriptor is StrategySetAttachment's column descriptors.
var StrategySetAttachColumnDescriptor = ColumnDescriptors{
	{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
	{Column: "app_id", NamedC: "app_id", Type: enumor.Numeric}}

// StrategySetAttachment strategy set attachment info.
type StrategySetAttachment struct {
	BizID uint32 `db:"biz_id" json:"biz_id"`
	AppID uint32 `db:"app_id" json:"app_id"`
}

// Validate validate strategy set's attachment.
func (s StrategySetAttachment) Validate() error {
	if s.BizID <= 0 {
		return errors.New("invalid biz id")
	}

	if s.AppID <= 0 {
		return errors.New("invalid app id")
	}

	return nil
}

// StrategySetSpecColumns defines StrategySetSpec's columns
var StrategySetSpecColumns = mergeColumns(StrategySetSpecColumnDescriptor)

// StrategySetSpecColumnDescriptor is StrategySetSpec's column descriptors.
var StrategySetSpecColumnDescriptor = ColumnDescriptors{
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "mode", NamedC: "mode", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String}}

// StrategySetSpec defines all the specifics for a strategy set, which
// is set by user.
type StrategySetSpec struct {
	Name string `db:"name" json:"name"`
	// Mode defines what mode of this strategy set works at, it is succeeded from
	// this strategy set's app's mode.
	// it can not be updated once it is created.
	Mode AppMode `db:"mode" json:"mode"`
	Memo string  `db:"memo" json:"memo"`
}

// ValidateCreate the strategy set specifics.
func (s StrategySetSpec) ValidateCreate() error {
	if err := validator.ValidateName(s.Name); err != nil {
		return err
	}

	if err := validator.ValidateMemo(s.Memo, false); err != nil {
		return err
	}

	if err := s.Mode.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate the strategy set specifics.
func (s StrategySetSpec) ValidateUpdate() error {
	if err := validator.ValidateName(s.Name); err != nil {
		return err
	}

	if err := validator.ValidateMemo(s.Memo, false); err != nil {
		return err
	}

	if len(s.Mode) != 0 {
		return errors.New("strategy set's mode can not be updated")
	}

	return nil
}
