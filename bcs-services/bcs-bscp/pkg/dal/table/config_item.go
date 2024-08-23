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
	"strconv"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/enumor"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/validator"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// ConfigItemColumns defines config item's columns
var ConfigItemColumns = mergeColumns(ConfigItemColumnDescriptor)

// ConfigItemColumnDescriptor is ConfigItem's column descriptors.
var ConfigItemColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{{Column: "id", NamedC: "id", Type: enumor.Numeric}},
	mergeColumnDescriptors("spec", CISpecColumnDescriptor),
	mergeColumnDescriptors("attachment", CIAttachmentColumnDescriptor),
	mergeColumnDescriptors("revision", RevisionColumnDescriptor))

// ConfigItem defines a basic configuration item
type ConfigItem struct {
	// ID is an auto-increased value, which is a config item's
	// unique identity.
	ID         uint32                `db:"id" json:"id" gorm:"primaryKey"`
	Spec       *ConfigItemSpec       `db:"spec" json:"spec" gorm:"embedded"`
	Attachment *ConfigItemAttachment `db:"attachment" json:"attachment" gorm:"embedded"`
	Revision   *Revision             `db:"revision" json:"revision" gorm:"embedded"`
}

// AppID AuditRes interface
func (c *ConfigItem) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (c *ConfigItem) ResID() uint32 {
	return c.ID
}

// ResType AuditRes interface
func (c *ConfigItem) ResType() string {
	return "config_item"
}

// TableName is the config item's database table name.
func (c ConfigItem) TableName() Name {
	return ConfigItemTable
}

// ValidateCreate validate the config item's specific when create it.
func (c ConfigItem) ValidateCreate(kit *kit.Kit) error {
	if c.ID != 0 {
		return errors.New("config item id can not be set")
	}

	if c.Spec == nil {
		return errors.New("spec should be set")
	}

	if err := c.Spec.ValidateCreate(kit); err != nil {
		return err
	}

	if c.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if err := c.Attachment.Validate(); err != nil {
		return err
	}

	if c.Revision == nil {
		return errors.New("revision should be set")
	}

	if err := c.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate the config item's specific when update it.
func (c ConfigItem) ValidateUpdate(kit *kit.Kit) error {
	if c.ID <= 0 {
		return errors.New("config item id should be set")
	}

	if c.Spec != nil {
		if err := c.Spec.ValidateUpdate(kit); err != nil {
			return err
		}
	}

	if c.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if err := c.Attachment.Validate(); err != nil {
		return err
	}

	if c.Revision != nil {
		if err := c.Revision.ValidateUpdate(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateDelete validate the config item's info when delete it.
func (c ConfigItem) ValidateDelete() error {
	if c.ID <= 0 {
		return errors.New("config item id should be set")
	}

	if c.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if c.Attachment.AppID <= 0 {
		return errors.New("app id should be set")
	}

	return nil
}

// ValidateRecover validate the config item's specific when recover it.
func (c ConfigItem) ValidateRecover(kit *kit.Kit) error {
	if c.ID == 0 {
		return errors.New("config item id can not be set")
	}

	if c.Spec == nil {
		return errors.New("spec should be set")
	}

	if err := c.Spec.ValidateCreate(kit); err != nil {
		return err
	}

	if c.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if err := c.Attachment.Validate(); err != nil {
		return err
	}

	if c.Revision == nil {
		return errors.New("revision should be set")
	}

	if err := c.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ConfigItemSpecColumns defines commit attachment's columns
var ConfigItemSpecColumns = mergeColumns(CISpecColumnDescriptor)

// RCISpecColumnDescriptor is released ConfigItemSpec's column descriptors, remove memo field.
var RCISpecColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{
		{Column: "name", NamedC: "name", Type: enumor.String},
		{Column: "path", NamedC: "path", Type: enumor.String},
		{Column: "file_type", NamedC: "file_type", Type: enumor.String},
		{Column: "file_mode", NamedC: "file_mode", Type: enumor.String}},
	mergeColumnDescriptors("permission", ColumnDescriptorColumnDescriptor))

// CISpecColumnDescriptor is ConfigItemSpec's column descriptors.
var CISpecColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{
		{Column: "name", NamedC: "name", Type: enumor.String},
		{Column: "path", NamedC: "path", Type: enumor.String},
		{Column: "file_type", NamedC: "file_type", Type: enumor.String},
		{Column: "file_mode", NamedC: "file_mode", Type: enumor.String},
		{Column: "memo", NamedC: "memo", Type: enumor.String}},
	mergeColumnDescriptors("permission", ColumnDescriptorColumnDescriptor))

// ConfigItemSpec is config item's specific which is defined
// by user.
type ConfigItemSpec struct {
	// Name is the name of this config item
	Name string `db:"name" json:"name" gorm:"column:name"`
	// Path is where these configurations to save.
	// Note:
	// 1. KV config type do not need path.
	// 2. this path is a relevant path to the sidecar's system workspace path.
	// 3. this path is the absolute path for user's workspace path.
	Path string `db:"path" json:"path" gorm:"column:path"`
	// FileType is the file type of this configuration.
	FileType FileFormat `db:"file_type" json:"file_type" gorm:"column:file_type"`
	FileMode FileMode   `db:"file_mode" json:"file_mode" gorm:"column:file_mode"`
	Memo     string     `db:"memo" json:"memo" gorm:"column:memo"`
	// KV类型，不能有Permission
	Permission *FilePermission `db:"permission" json:"permission" gorm:"embedded"`
}

// ValidateCreate validate the config item's specifics
func (ci ConfigItemSpec) ValidateCreate(kit *kit.Kit) error {

	if err := validator.ValidateFileName(kit, ci.Name); err != nil {
		return err
	}

	if err := ci.FileType.Validate(kit); err != nil {
		return err
	}

	if err := ci.FileMode.Validate(kit); err != nil {
		return err
	}

	if err := ValidatePath(kit, ci.Path, ci.FileMode); err != nil {
		return err
	}

	if err := validator.ValidateMemo(kit, ci.Memo, false); err != nil {
		return err
	}

	if err := ci.Permission.Validate(kit, ci.FileMode); err != nil {
		return err
	}

	return nil
}

// ValidatePath validate path.
func ValidatePath(kit *kit.Kit, path string, fileMode FileMode) error {
	switch fileMode {
	case Windows:
		if err := validator.ValidateWinFilePath(kit, path); err != nil {
			return errors.New(i18n.T(kit, "verify Windows file paths failed, path: %s, err: %v", path, err))
		}
	case Unix:
		if err := validator.ValidateUnixFilePath(kit, path); err != nil {
			return errors.New(i18n.T(kit, "verify Unix file paths failed, path: %s, err: %v", path, err))
		}
	default:
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "unknown file mode "+string(fileMode)))
	}

	return nil
}

// Validate file permission.
func (f FilePermission) Validate(kit *kit.Kit, mode FileMode) error {
	switch mode {
	case Windows:
		return errors.New("windows file mode not supported at the moment")

	case Unix:
		if len(f.User) == 0 {
			return errors.New("invalid user")
		}

		if len(f.UserGroup) == 0 {
			return errors.New("invalid user group")
		}

		if len(f.Privilege) != 3 {
			return errors.New("invalid privilege, privilege length should be 3")
		}

		for i := 0; i < 3; i++ {
			p, err := strconv.ParseInt(string(f.Privilege[i]), 10, 64)
			if err != nil {
				return fmt.Errorf("invalid privilege, convert to int failed, err: %v", err)
			}

			if p < 0 || p > 7 {
				return errors.New("invalid privilege, correct permissions range 0-7")
			}
		}
		return nil

	default:
		return errors.New("unknown file mode " + string(mode))
	}
}

// ValidateUpdate validate the config item's specifics when update it.
func (ci ConfigItemSpec) ValidateUpdate(kit *kit.Kit) error {

	if len(ci.Name) != 0 {
		if err := validator.ValidateFileName(kit, ci.Name); err != nil {
			return err
		}
	}

	if len(ci.FileType) != 0 {
		if err := ci.FileType.Validate(kit); err != nil {
			return err
		}
	}

	if err := ci.FileMode.Validate(kit); err != nil {
		return err
	}

	if err := ValidatePath(kit, ci.Path, ci.FileMode); err != nil {
		return err
	}

	if len(ci.Memo) != 0 {
		if err := validator.ValidateMemo(kit, ci.Memo, false); err != nil {
			return err
		}
	}

	if err := ci.Permission.Validate(kit, ci.FileMode); err != nil {
		return err
	}

	return nil
}

// CIAttachmentColumns defines commit attachment's columns
var CIAttachmentColumns = mergeColumns(CIAttachmentColumnDescriptor)

// CIAttachmentColumnDescriptor is ConfigItemAttachment's column descriptors.
var CIAttachmentColumnDescriptor = ColumnDescriptors{
	{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
	{Column: "app_id", NamedC: "app_id", Type: enumor.Numeric}}

// ConfigItemAttachment is a configuration item attachment
type ConfigItemAttachment struct {
	BizID uint32 `db:"biz_id" json:"biz_id" gorm:"column:biz_id"`
	AppID uint32 `db:"app_id" json:"app_id" gorm:"column:app_id"`
}

// Validate config item attachment.
func (c ConfigItemAttachment) Validate() error {
	if c.BizID <= 0 {
		return errors.New("invalid biz id")
	}

	if c.AppID <= 0 {
		return errors.New("invalid app id")
	}

	return nil

}

// FilePermissionColumns defines file permission's columns
var FilePermissionColumns = mergeColumns(ColumnDescriptorColumnDescriptor)

// ColumnDescriptorColumnDescriptor is FilePermission's column descriptors.
var ColumnDescriptorColumnDescriptor = ColumnDescriptors{
	{Column: "user", NamedC: "user", Type: enumor.String},
	{Column: "user_group", NamedC: "user_group", Type: enumor.String},
	{Column: "privilege", NamedC: "privilege", Type: enumor.String}}

// FilePermission defines a config's permission details.
type FilePermission struct {
	User      string `db:"user" json:"user" gorm:"column:user"`
	UserGroup string `db:"user_group" json:"user_group" gorm:"column:user_group"`
	// config file's privilege
	Privilege string `db:"privilege" json:"privilege" gorm:"column:privilege"`
}

const (
	// Text file format
	Text FileFormat = "text"
	// Binary file format
	Binary FileFormat = "binary"
)

// FileFormat is config item format type
type FileFormat string

// Validate the file format is supported or not.
func (f FileFormat) Validate(kit *kit.Kit) error {
	switch f {
	case Text:
	case Binary:
	default:
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "unsupported file format: %s", f))
	}

	return nil
}

const (
	// Windows NOTES
	Windows FileMode = "win"
	// Unix NOTES
	Unix FileMode = "unix"
)

// FileMode NOTES
type FileMode string

// Validate the file mode is supported or not.
func (f FileMode) Validate(kit *kit.Kit) error {
	switch f {
	case Windows:
	case Unix:
	default:
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "unsupported file mode: %s", f))
	}

	return nil
}

// ListConfigItemCounts return data structure
type ListConfigItemCounts struct {
	AppId     uint32    `gorm:"column:app_id" json:"app_id"`
	Count     uint32    `gorm:"column:count" json:"count"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}
