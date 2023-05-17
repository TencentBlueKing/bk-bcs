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

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/validator"
)

// TemplateSpaceColumns defines TemplateSpace's columns
var TemplateSpaceColumns = mergeColumns(TemplateSpaceColumnDescriptor)

// TemplateSpaceColumnDescriptor is TemplateSpace's column descriptors.
var TemplateSpaceColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{{Column: "id", NamedC: "id", Type: enumor.Numeric}},
	mergeColumnDescriptors("spec", TemplateSpaceSpecColumnDescriptor),
	mergeColumnDescriptors("attachment", TemplateSpaceAttachmentColumnDescriptor),
	mergeColumnDescriptors("revision", RevisionColumnDescriptor))

// TemplateSpace defines a TemplateSpace for an app to publish.
// it contains the selector to define the scope of the matched instances.
type TemplateSpace struct {
	// ID is an auto-increased value, which is a unique identity of a TemplateSpace.
	ID         uint32                   `db:"id" json:"id"`
	Spec       *TemplateSpaceSpec       `db:"spec" json:"spec"`
	Attachment *TemplateSpaceAttachment `db:"attachment" json:"attachment"`
	Revision   *Revision                `db:"revision" json:"revision"`
}

// TableName is the TemplateSpace's database table name.
func (s TemplateSpace) TableName() Name {
	return TemplateSpaceTable
}

// ValidateCreate validate TemplateSpace is valid or not when create it.
func (s TemplateSpace) ValidateCreate() error {

	if s.ID > 0 {
		return errors.New("id should not be set")
	}

	if s.Spec == nil {
		return errors.New("spec not set")
	}

	if err := s.Spec.ValidateCreate(); err != nil {
		return err
	}

	if s.Attachment == nil {
		return errors.New("attachment not set")
	}

	if err := s.Attachment.Validate(); err != nil {
		return err
	}

	if s.Revision == nil {
		return errors.New("revision not set")
	}

	if err := s.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate TemplateSpace is valid or not when update it.
func (s TemplateSpace) ValidateUpdate() error {

	if s.ID <= 0 {
		return errors.New("id should be set")
	}

	changed := false
	if s.Spec != nil {
		changed = true
		if err := s.Spec.ValidateUpdate(); err != nil {
			return err
		}
	}

	if s.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if s.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if !changed {
		return errors.New("nothing is found to be change")
	}

	if s.Revision == nil {
		return errors.New("revision not set")
	}

	if err := s.Revision.ValidateUpdate(); err != nil {
		return err
	}

	return nil
}

// ValidateDelete validate the TemplateSpace's info when delete it.
func (s TemplateSpace) ValidateDelete() error {
	if s.ID <= 0 {
		return errors.New("TemplateSpace id should be set")
	}

	if s.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	return nil
}

// TemplateSpaceSpecColumns defines TemplateSpaceSpec's columns
var TemplateSpaceSpecColumns = mergeColumns(TemplateSpaceSpecColumnDescriptor)

// TemplateSpaceSpecColumnDescriptor is TemplateSpaceSpec's column descriptors.
var TemplateSpaceSpecColumnDescriptor = ColumnDescriptors{
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
}

// TemplateSpaceSpec defines all the specifics for TemplateSpace set by user.
type TemplateSpaceSpec struct {
	Name string `db:"name" json:"name"`
	Memo string `db:"memo" json:"memo"`
}

// TemplateSpaceType is the type of TemplateSpace
type TemplateSpaceType string

// ValidateCreate validate TemplateSpace spec when it is created.
func (s TemplateSpaceSpec) ValidateCreate() error {
	if err := validator.ValidateName(s.Name); err != nil {
		return err
	}

	if err := validator.ValidateAppName(s.Name); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate TemplateSpace spec when it is updated.
func (s TemplateSpaceSpec) ValidateUpdate() error {
	if err := validator.ValidateName(s.Name); err != nil {
		return err
	}

	if err := validator.ValidateMemo(s.Memo, false); err != nil {
		return err
	}

	return nil
}

// TemplateSpaceAttachmentColumns defines TemplateSpaceAttachment's columns
var TemplateSpaceAttachmentColumns = mergeColumns(TemplateSpaceAttachmentColumnDescriptor)

// TemplateSpaceAttachmentColumnDescriptor is TemplateSpaceAttachment's column descriptors.
var TemplateSpaceAttachmentColumnDescriptor = ColumnDescriptors{
	{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric}}

// TemplateSpaceAttachment defines the TemplateSpace attachments.
type TemplateSpaceAttachment struct {
	BizID uint32 `db:"biz_id" json:"biz_id"`
}

// IsEmpty test whether TemplateSpace attachment is empty or not.
func (s TemplateSpaceAttachment) IsEmpty() bool {
	return s.BizID == 0
}

// Validate whether TemplateSpace attachment is valid or not.
func (s TemplateSpaceAttachment) Validate() error {
	if s.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	return nil
}
