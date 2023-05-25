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
	"bscp.io/pkg/criteria/enumor"
	"errors"
)

// HookReleaseColumns defines Hook's columns
var HookReleaseColumns = mergeColumns(HookReleaseColumnDescriptor)

// HookReleaseColumnDescriptor is Hook's column descriptors.
var HookReleaseColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{{Column: "id", NamedC: "id", Type: enumor.Numeric}},
	mergeColumnDescriptors("spec", HookReleaseSpecColumnDescriptor),
	mergeColumnDescriptors("attachment", HookReleaseAttachmentColumnDescriptor),
	mergeColumnDescriptors("revision", RevisionColumnDescriptor))

// HookRelease 脚本版本
type HookRelease struct {
	// ID is an auto-increased value, which is a unique identity of a hook.
	ID uint32 `db:"id" json:"id"`

	Spec       *HookReleaseSpec       `json:"spec" gorm:"embedded"`
	Attachment *HookReleaseAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision              `json:"revision" gorm:"embedded"`
}

// HookReleaseSpecColumns defines HookReleaseSpec's columns
var HookReleaseSpecColumns = mergeColumns(HookSpecColumnDescriptor)

// HookReleaseSpecColumnDescriptor is HookSpec's column descriptors.
var HookReleaseSpecColumnDescriptor = ColumnDescriptors{
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "contents", NamedC: "contents", Type: enumor.String},
	{Column: "release_log", NamedC: "release_log", Type: enumor.String},
	{Column: "state", NamedC: "state", Type: enumor.Boolean},
}

// HookReleaseAttachmentColumnDescriptor is HookReleaseAttachment's column descriptors.
var HookReleaseAttachmentColumnDescriptor = ColumnDescriptors{
	{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
	{Column: "hook_id", NamedC: "hook_id", Type: enumor.Numeric},
}

// HookReleaseSpec defines all the specifics for hook set by user.
type HookReleaseSpec struct {
	Name       string        `json:"name" gorm:"column:name"`
	PublishNum uint32        `json:"publish_num" gorm:"column:publish_num"`
	PubState   ReleaseStatus `json:"pub_state" gorm:"column:pub_state"`
	Contents   string        `json:"contents" gorm:"column:contents"`
	Memo       string        `json:"memo" gorm:"column:memo"`
}

// HookReleaseAttachment defines the hook attachments.
type HookReleaseAttachment struct {
	BizID  uint32 `json:"biz_id" gorm:"column:biz_id"`
	HookID uint32 `json:"hook_id" gorm:"column:hook_id"`
}

// TableName is the hook's database table name.
func (r *HookRelease) TableName() Name {
	return "hook_releases"
}

// AppID AuditRes interface
func (r *HookRelease) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (r *HookRelease) ResID() uint32 {
	return r.ID
}

// ResType AuditRes interface
func (r *HookRelease) ResType() string {
	return "hook_releases"
}

// ValidateCreate validate hook is valid or not when create it.
func (r *HookRelease) ValidateCreate() error {

	if r.ID > 0 {
		return errors.New("id should not be set")
	}

	if r.Spec == nil {
		return errors.New("spec not set")
	}

	//if err := r.Spec.ValidateCreate(); err != nil {
	//	return err
	//}

	if r.Attachment == nil {
		return errors.New("attachment not set")
	}

	//if err := r.Attachment.Validate(); err != nil {
	//	return err
	//}

	if r.Revision == nil {
		return errors.New("revision not set")
	}

	if err := r.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateDelete validate the hook release info when delete it.
func (r HookRelease) ValidateDelete() error {
	if r.ID <= 0 {
		return errors.New("hook release id should be set")
	}

	if r.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if r.Attachment.HookID <= 0 {
		return errors.New("hook id should be set")
	}

	return nil
}

func (r HookRelease) ValidatePublish() error {

	if r.ID <= 0 {
		return errors.New("hook release id should be set")
	}

	if r.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if r.Attachment.HookID <= 0 {
		return errors.New("hook id should be set")
	}

	return nil
}
