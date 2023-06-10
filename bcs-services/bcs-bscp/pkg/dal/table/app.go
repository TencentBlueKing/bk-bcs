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
	"time"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/validator"
)

// AppColumns defines all the app table's columns.
var AppColumns = mergeColumns(AppColumnDescriptor)

// AppColumnDescriptor is App's column descriptors.
var AppColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{
		{Column: "id", NamedC: "id", Type: enumor.Numeric},
		{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
	},
	mergeColumnDescriptors("spec", AppSpecColumnDescriptor),
	mergeColumnDescriptors("revision", RevisionColumnDescriptor))

// App defines an application's detail information
type App struct {
	// ID is an auto-increased value, which is an application's
	// unique identity.
	ID uint32 `db:"id" json:"id" gorm:"primaryKey"`
	// BizID is the business is which this app belongs to
	BizID uint32 `db:"biz_id" json:"biz_id" gorm:"primaryKey"`
	// Spec is a collection of app's specifics defined with user
	Spec *AppSpec `db:"spec" json:"spec" gorm:"embedded"`
	// Revision record this app's revision information
	Revision *Revision `db:"revision" gorm:"embedded"`
}

// TableName is the Hook's database table name.
func (h *App) TableName() string {
	return "applications"
}

// AppID HookRes interface
func (h *App) AppID() uint32 {
	return h.ID
}

// ResID HookRes interface
func (h *App) ResID() uint32 {
	return h.ID
}

// ResType HookRes interface
func (h *App) ResType() string {
	return "application"
}

// ValidateCreate validate app's info when created.
func (a App) ValidateCreate() error {
	if a.ID != 0 {
		return errors.New("id can not be set")
	}

	if a.BizID <= 0 {
		return errors.New("invalid biz id")
	}

	if a.Spec == nil {
		return errors.New("invalid spec, is nil")
	}

	if err := a.Spec.ValidateCreate(); err != nil {
		return err
	}

	if a.Revision == nil {
		return errors.New("invalid revision, is nil")
	}

	if err := a.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate app's info when update.
func (a App) ValidateUpdate(configType ConfigType) error {
	if a.ID <= 0 {
		return errors.New("id is not set")
	}

	if a.BizID <= 0 {
		return errors.New("biz id not set")
	}

	if a.Spec == nil {
		return errors.New("invalid spec, is nil")
	}

	if err := a.Spec.ValidateUpdate(configType); err != nil {
		return err
	}

	if a.Revision == nil {
		return errors.New("invalid revision, is nil")
	}

	if err := a.Revision.ValidateUpdate(); err != nil {
		return err
	}

	return nil
}

// ValidateDelete validate app's info when delete.
func (a App) ValidateDelete() error {
	if a.ID <= 0 {
		return errors.New("app id not set")
	}

	if a.BizID <= 0 {
		return errors.New("biz id not set")
	}

	return nil
}

// AppSpecColumns defines all the app spec's columns.
var AppSpecColumns = mergeColumns(AppSpecColumnDescriptor)

// AppSpecColumnDescriptor is AppSpec's column descriptors.
var AppSpecColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{
		{Column: "name", NamedC: "name", Type: enumor.String},
		{Column: "config_type", NamedC: "config_type", Type: enumor.String},
		{Column: "mode", NamedC: "mode", Type: enumor.String},
		{Column: "memo", NamedC: "memo", Type: enumor.String},
	},
	mergeColumnDescriptors("reload", ReloadColumnDescriptor))

// AppSpec is a collection of app's specifics defined with user
type AppSpec struct {
	// Name is application's name
	Name string `db:"name" json:"name"`
	// ConfigType defines which type is this configuration, different type has the
	// different ways to be consumed.
	ConfigType ConfigType `db:"config_type" json:"config_type" gorm:"embedded"`
	// Mode defines what mode of this app works at.
	// Mode can not be updated once it is created.
	Mode   AppMode `db:"mode" json:"mode"`
	Memo   string  `db:"memo" json:"memo"`
	Reload *Reload `db:"reload" json:"reload" gorm:"embedded"`
}

const (
	// Normal means this is a normal app, and configuration
	// items can be consumed directly.
	Normal AppMode = "normal"

	// Namespace means that this app runs in the namespace
	// mode, which means user must consume app's configuration
	// item with namespace information.
	Namespace AppMode = "namespace"
)

// AppMode is the mode of an app works at, different mode has the
// different way or restricts to consume this strategy's configurations.
type AppMode string

// Validate strategy set type.
func (s AppMode) Validate() error {
	switch s {
	case Normal:
	case Namespace:
	default:
		return fmt.Errorf("unsupported app working mode: %s", s)
	}

	return nil
}

// ValidateCreate validate spec when created.
func (as *AppSpec) ValidateCreate() error {
	if as == nil {
		return errors.New("app spec is nil")
	}

	if err := validator.ValidateAppName(as.Name); err != nil {
		return err
	}

	if err := as.ConfigType.Validate(); err != nil {
		return err
	}

	if err := as.Mode.Validate(); err != nil {
		return err
	}

	if err := validator.ValidateMemo(as.Memo, false); err != nil {
		return err
	}

	switch as.ConfigType {
	case File:
		if err := as.Reload.ValidateCreate(); err != nil {
			return err
		}
	case KV, Table:
		if err := as.Reload.IsEmpty(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown config type: %s", as.ConfigType)
	}

	return nil
}

// ValidateUpdate validate spec when updated.
func (as *AppSpec) ValidateUpdate(configType ConfigType) error {
	if as == nil {
		return errors.New("app spec is nil")
	}

	if err := validator.ValidateAppName(as.Name); err != nil {
		return err
	}

	if len(as.ConfigType) > 0 {
		return errors.New("app's type can not be updated")
	}

	if len(as.Mode) > 0 {
		return errors.New("app's mode can not be updated")
	}

	if err := validator.ValidateMemo(as.Memo, false); err != nil {
		return err
	}

	switch configType {
	case File:
		if as.Reload != nil {
			if err := as.Reload.ValidateUpdate(); err != nil {
				return nil
			}
		}
	case KV, Table:
		if as.Reload != nil {
			if err := as.Reload.IsEmpty(); err != nil {
				return nil
			}
		}
	default:
		return fmt.Errorf("unknown config type: %s", as.ConfigType)
	}

	return nil
}

// ReloadColumns defines all the app reload spec's columns.
var ReloadColumns = mergeColumns(ReloadColumnDescriptor)

// ReloadColumnDescriptor is Reload's column descriptors.
var ReloadColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{
		{Column: "reload_type", NamedC: "reload_type", Type: enumor.String},
	},
	mergeColumnDescriptors("file_reload_spec", FileReloadSpecColumnDescriptor))

// Reload is a collection of app reload specifics defined with user. only is used when this app is file config type.
// Reload is used to control how bscp sidecar notifies applications to go to reload config files.
type Reload struct {
	ReloadType     AppReloadType   `db:"reload_type" json:"reload_type"`
	FileReloadSpec *FileReloadSpec `db:"file_reload_spec" json:"file_reload_spec" gorm:"embedded"`
}

// IsEmpty reload.
func (r *Reload) IsEmpty() error {
	if r == nil {
		return nil
	}

	if len(r.ReloadType) != 0 {
		return errors.New("reload type is not nil")
	}

	if r.FileReloadSpec != nil {
		if err := r.FileReloadSpec.IsEmpty(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateCreate reload spec when create.
func (r *Reload) ValidateCreate() error {
	if r == nil {
		return errors.New("reload spec is required")
	}

	if len(r.ReloadType) == 0 {
		return errors.New("reload type is required")
	}

	if err := r.ReloadType.Validate(); err != nil {
		return err
	}

	switch r.ReloadType {
	case ReloadWithFile:
		if err := r.FileReloadSpec.Validate(); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown app reload type: %s", r.ReloadType)
	}

	return nil
}

// ValidateUpdate reload spec when update.
func (r *Reload) ValidateUpdate() error {
	if r == nil {
		return errors.New("reload spec is required")
	}

	if len(r.ReloadType) != 0 {
		if err := r.ReloadType.Validate(); err != nil {
			return err
		}

		switch r.ReloadType {
		case ReloadWithFile:
			if err := r.FileReloadSpec.Validate(); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown app reload type: %s", r.ReloadType)
		}
	}

	return nil
}

// FileReloadSpecColumns defines all the app file reload spec's columns.
var FileReloadSpecColumns = mergeColumns(FileReloadSpecColumnDescriptor)

// FileReloadSpecColumnDescriptor is Reload's column descriptors.
var FileReloadSpecColumnDescriptor = ColumnDescriptors{
	{Column: "reload_file_path", NamedC: "reload_file_path", Type: enumor.String},
}

// FileReloadSpec is a collection of file reload spec's specifics defined with user.
type FileReloadSpec struct {
	ReloadFilePath string `db:"reload_file_path" json:"reload_file_path"`
}

// IsEmpty file reload spec.
func (f *FileReloadSpec) IsEmpty() error {
	if f == nil {
		return nil
	}

	if len(f.ReloadFilePath) != 0 {
		return errors.New("reload file path is not nil")
	}

	return nil
}

// Validate file reload spec.
func (f *FileReloadSpec) Validate() error {
	if f == nil {
		return errors.New("file reload spec is required")
	}

	if err := validator.ValidateReloadFilePath(f.ReloadFilePath); err != nil {
		return err
	}

	return nil
}

const (
	// KV is kv configuration type
	KV ConfigType = "kv"
	// File is file configuration type
	File ConfigType = "file"
	// Table is table configuration type
	Table ConfigType = "table"
)

// ConfigType is the app's config item's type
type ConfigType string

// Validate the config type is supported or not.
func (c ConfigType) Validate() error {
	switch c {
	case KV:
		return errors.New("not support kv config type for now")
	case File:
	case Table:
		return errors.New("not support table config type for now")
	default:
		return fmt.Errorf("unsupported config type: %s", c)
	}

	return nil
}

const (
	// ReloadWithFile the app's sidecar instance will write the downloaded configuration release information to the
	// reload file, then the application instance uses this reload file to determine whether has a new configuration
	// need to load.
	ReloadWithFile AppReloadType = "file"
)

// AppReloadType is the app's sidecar instance to notify application reload config files way.
type AppReloadType string

// Validate app reload type
func (rt AppReloadType) Validate() error {
	switch rt {
	case ReloadWithFile:
	default:
		return fmt.Errorf("unsupported app reload type: %s", rt)
	}

	return nil
}

// ArchivedAppColumns defines ArchivedApp's columns
var ArchivedAppColumns = mergeColumns(ArchivedAppColumnDescriptor)

// ArchivedAppColumnDescriptor is ArchivedApp's column descriptors.
var ArchivedAppColumnDescriptor = ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.Numeric},
	{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
	{Column: "app_id", NamedC: "app_id", Type: enumor.Numeric},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time}}

// ArchivedApp is used to record applications basic information
// which is used to purge resources related with this application
// asynchronously.
type ArchivedApp struct {
	ID        uint32    `db:"id" json:"id"`
	BizID     uint32    `db:"biz_id" json:"biz_id"`
	AppID     uint32    `db:"app_id" json:"app_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// TableName is the archived app's database table name.
func (a ArchivedApp) TableName() Name {
	return ArchivedAppTable
}
