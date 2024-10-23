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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/validator"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// App defines an application's detail information
type App struct {
	// ID is an auto-increased value, which is an application's
	// unique identity.
	ID uint32 `json:"id" gorm:"primaryKey"`
	// BizID is the business is which this app belongs to
	BizID uint32 `json:"biz_id" gorm:"column:biz_id"`
	// Spec is a collection of app's specifics defined with user
	Spec *AppSpec `json:"spec" gorm:"embedded"`
	// Revision record this app's revision information
	Revision *Revision `json:"revision" gorm:"embedded"`
}

// TableName is the app's database table name.
func (a *App) TableName() string {
	return "applications"
}

// AppID AuditRes interface
func (a *App) AppID() uint32 {
	return a.ID
}

// ResID AuditRes interface
func (a *App) ResID() uint32 {
	return a.ID
}

// ResType AuditRes interface
func (a *App) ResType() string {
	return "app"
}

// ValidateCreate validate app's info when created.
func (a *App) ValidateCreate(kit *kit.Kit) error {
	if a.ID != 0 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "id can not be set"))
	}

	if a.BizID <= 0 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid biz id"))
	}

	if a.Spec == nil {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid spec, is nil"))
	}

	if err := a.Spec.ValidateCreate(kit); err != nil {
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
func (a *App) ValidateUpdate(kit *kit.Kit, configType ConfigType) error {
	if a.ID <= 0 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "id can not be set"))
	}

	if a.BizID <= 0 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid biz id"))
	}

	if a.Spec == nil {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid spec, is nil"))
	}

	if err := a.Spec.ValidateUpdate(kit, configType); err != nil {
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
func (a *App) ValidateDelete() error {
	if a.ID <= 0 {
		return errors.New("app id not set")
	}

	if a.BizID <= 0 {
		return errors.New("biz id not set")
	}

	return nil
}

// AppSpec is a collection of app's specifics defined with user
type AppSpec struct {
	// Name is application's name
	Name string `json:"name" gorm:"column:name"`
	// ConfigType defines which type is this configuration, different type has the
	// different ways to be consumed.
	ConfigType       ConfigType  `json:"config_type" gorm:"column:config_type"`
	Memo             string      `json:"memo" gorm:"column:memo"`
	Alias            string      `json:"alias" gorm:"alias"`
	DataType         DataType    `json:"data_type" gorm:"data_type"`
	LastConsumedTime *time.Time  `json:"last_consumed_time" gorm:"column:last_consumed_time"`
	ApproveType      ApproveType `json:"approve_type" gorm:"approve_type"`
	IsApprove        bool        `json:"is_approve" gorm:"is_approve"`
	Approver         string      `json:"approver" gorm:"approver"`
}

// ValidateCreate validate spec when created.
func (as *AppSpec) ValidateCreate(kit *kit.Kit) error {
	if as == nil {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "app spec is nil"))
	}

	if err := validator.ValidateAppName(kit, as.Name); err != nil {
		return err
	}

	if err := validator.ValidateAppAlias(kit, as.Alias); err != nil {
		return err
	}

	if err := as.ConfigType.Validate(kit); err != nil {
		return err
	}

	if err := validator.ValidateMemo(kit, as.Memo, false); err != nil {
		return err
	}

	if err := as.ApproveType.ValidateApproveType(); err != nil {
		return err
	}

	if as.IsApprove && (as.ApproveType == "" || as.Approver == "") {
		return errors.New("approve_type or approver cannot be empty")
	}

	switch as.ConfigType {
	case File:
	case KV:
		if err := as.DataType.ValidateApp(kit); err != nil {
			return err
		}
	case Table:
	default:
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "unknown config type: %s", as.ConfigType))
	}

	return nil
}

// ValidateUpdate validate spec when updated.
func (as *AppSpec) ValidateUpdate(kit *kit.Kit, configType ConfigType) error {
	if as == nil {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "app spec is nil"))
	}

	if err := validator.ValidateAppName(kit, as.Name); err != nil {
		return err
	}

	if len(as.ConfigType) > 0 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "app's type can not be updated"))
	}

	if err := validator.ValidateMemo(kit, as.Memo, false); err != nil {
		return err
	}

	switch configType {
	case File:
	case KV:
		if err := as.DataType.ValidateApp(kit); err != nil {
			return err
		}
	case Table:
	default:
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "unknown config type: %s", as.ConfigType))
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
func (c ConfigType) Validate(kit *kit.Kit) error {
	switch c {
	case KV:
	case File:
	case Table:
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "not support table config type for now"))
	default:
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "unsupported config type: %s", c))
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
func (rt AppReloadType) Validate(kit *kit.Kit) error {
	switch rt {
	case ReloadWithFile:
	default:
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "unsupported app reload type: %s", rt))
	}

	return nil
}

// ArchivedApp is used to record applications basic information
// which is used to purge resources related with this application
// asynchronously.
type ArchivedApp struct {
	ID        uint32    `json:"id" gorm:"primaryKey"`
	BizID     uint32    `json:"biz_id" gorm:"column:biz_id"`
	AppID     uint32    `json:"app_id" gorm:"column:app_id"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}

// TableName is the archived app's database table name.
func (a *ArchivedApp) TableName() string {
	return "archived_apps"
}

// DataType is the app's kv type
type DataType string

const (
	// KvAny 任意类型
	KvAny DataType = "any"
	// KvStr is the type for string kv
	KvStr DataType = "string"
	// KvNumber is the type for number kv
	KvNumber DataType = "number"
	// KvText is the type for text kv
	KvText DataType = "text"
	// KvJson is the type for json kv
	KvJson DataType = "json"
	// KvYAML is the type for yaml kv
	KvYAML DataType = "yaml"
	// KvXml is the type for xml kv
	KvXml DataType = "xml"
	// KvSecret is the type for secret kv
	KvSecret DataType = "secret"
)

// ValidateApp the kvType and value match
func (k DataType) ValidateApp(kit *kit.Kit) error {
	switch k {
	case KvAny:
	case KvStr:
	case KvNumber:
	case KvText:
	case KvJson:
	case KvYAML:
	case KvXml:
	case KvSecret:
	default:
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid data-type"))
	}
	return nil
}

// ApproveType is the app's config approval type
type ApproveType string

const (
	// CountSign counter sign
	CountSign ApproveType = "CountSign"
	// OrSign or sign
	OrSign ApproveType = "OrSign"
)

// ValidateApproveType validate approve type
func (a ApproveType) ValidateApproveType() error {
	switch a {
	case CountSign:
	case OrSign:
	case "":
	default:
		return fmt.Errorf("unsupported approve type: %s", a)
	}
	return nil
}
