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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/validator"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// ShardingDB defines a mysql instance
type ShardingDB struct {
	// ID is an auto-increased value, which is a unique identity
	// of a sharding db instance.
	// if sharding mysql instance is used, then it can not be
	// deleted.
	ID       uint32          `db:"id" json:"id"`
	Spec     *ShardingDBSpec `db:"spec" json:"spec"`
	Revision *Revision       `db:"revision" json:"revision"`
}

// TableName is the sharding mysql instance's database table name.
func (s ShardingDB) TableName() Name {
	return ShardingDBTable
}

// ValidateCreate sharding db details
func (s ShardingDB) ValidateCreate(kit *kit.Kit) error {
	if s.ID > 0 {
		return errors.New("id can not set")
	}

	if s.Spec == nil {
		return errors.New("spec not set")
	}

	if err := s.Spec.Validate(kit); err != nil {
		return err
	}

	if err := s.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate sharding db details
func (s ShardingDB) ValidateUpdate(kit *kit.Kit) error {

	if s.ID <= 0 {
		return errors.New("id not set")
	}

	if s.Spec == nil {
		return errors.New("spec not set")
	}

	if err := s.Spec.Validate(kit); err != nil {
		return err
	}

	if err := s.Revision.ValidateUpdate(); err != nil {
		return err
	}

	return nil
}

const (
	// DefaultSharding means this is mysql instance is used as a
	// default database to store resources.
	DefaultSharding ShardingType = "default"

	// ResourceSharding means this is a resource mysql instance,
	// which is used to store all the biz related resources.
	ResourceSharding ShardingType = "resource"

	// AuditSharding means this mysql instance is used to
	// save audit logs.
	AuditSharding ShardingType = "audit"

	// EventSharding means this mysql instance is used to
	// save events
	EventSharding ShardingType = "event"
)

// ShardingType is sharding type, used to distinguish database types for different functions.
type ShardingType string

// Validate sharding type.
func (s ShardingType) Validate() error {
	switch s {
	case DefaultSharding:
	case ResourceSharding:
		return fmt.Errorf("%s sharding not support for now", s)
	case AuditSharding:
		return fmt.Errorf("%s sharding not support for now", s)
	case EventSharding:
		return fmt.Errorf("%s sharding not support for now", s)
	default:
		return fmt.Errorf("unsupported sharding type: %s", s)
	}

	return nil
}

// ShardingDBSpec is the sharding db instance's specifics
type ShardingDBSpec struct {
	// Type defines what type of this mysql instance belongs
	// to, different type have different usage.
	// Note: all the sharding types except "normal" type must
	// have only one mysql instance.
	Type     ShardingType `db:"type" json:"type"`
	Host     string       `db:"host" json:"host"`
	Port     uint32       `db:"port" json:"port"`
	User     string       `db:"user" json:"user"`
	Password string       `db:"password" json:"password"`
	Database string       `db:"database" json:"database"`
	Memo     string       `db:"memo" json:"memo"`
}

// Validate sharding db instance's specifics
func (s ShardingDBSpec) Validate(kit *kit.Kit) error {

	if err := s.Type.Validate(); err != nil {
		return err
	}

	if len(s.Host) == 0 {
		return errors.New("host not set")
	}

	if s.Port <= 0 {
		return errors.New("port not set")
	}

	if len(s.User) == 0 {
		return errors.New("user not set")
	}

	if len(s.Password) == 0 {
		return errors.New("passport not set")
	}

	if err := validator.ValidateMemo(kit, s.Memo, false); err != nil {
		return err
	}

	return nil
}

// ShardingBiz defines which db instance a biz used.
type ShardingBiz struct {
	ID       uint32           `db:"id" json:"id"`
	Spec     *ShardingBizSpec `db:"spec" json:"spec"`
	Revision *Revision        `db:"revision" json:"revision"`
}

// TableName is the sharding biz's database table name.
func (s ShardingBiz) TableName() Name {
	return ShardingBizTable
}

// ValidateCreate validate sharding biz details when create it
func (s ShardingBiz) ValidateCreate(kit *kit.Kit) error {

	if s.ID > 0 {
		return errors.New("id should not be set")
	}

	if s.Spec == nil {
		return errors.New("invalid spec")
	}

	if err := s.Spec.Validate(kit); err != nil {
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

// ValidateUpdate validate sharding biz details when update it
func (s ShardingBiz) ValidateUpdate(kit *kit.Kit) error {

	if s.ID <= 0 {
		return errors.New("invalid id")
	}

	if s.Spec == nil {
		return errors.New("spec not set")
	}

	if err := s.Spec.Validate(kit); err != nil {
		return err
	}

	if s.Revision == nil {
		return errors.New("revision not set")
	}

	if err := s.Revision.ValidateUpdate(); err != nil {
		return err
	}

	return nil
}

// ShardingBizSpec defines which sharding db instance a biz used.
type ShardingBizSpec struct {
	// ShardingDBID is the sharding db's identity id.
	// ShardingDBID is associated with foreign key ShardingDB.ID
	ShardingDBID uint32 `db:"sharding_db_id" json:"sharding_db_id"`

	// BizID is the biz id which used this ShardingDBID.
	// Note:
	// 1. one biz can only have one working mysql instance. which
	// means one biz id can only associate with only one ShardingDBID.
	// 2.if a biz is removed, then remove this record at the same time.
	// 3. BizID can not be edited.
	BizID uint32 `db:"biz_id" json:"biz_id"`
	Memo  string `db:"memo" json:"memo"`
}

// Validate sharding biz specifics
func (s ShardingBizSpec) Validate(kit *kit.Kit) error {
	if s.ShardingDBID <= 0 {
		return errors.New("invalid sharding db id")
	}

	if s.BizID <= 0 {
		return errors.New("invalid biz id")
	}

	if err := validator.ValidateMemo(kit, s.Memo, false); err != nil {
		return err
	}

	return nil
}

// IDGenerator defines all the specifics to generate resource's unique
// id list with different step.
type IDGenerator struct {
	ID uint32 `db:"id" json:"id" gorm:"primaryKey"`
	// Resource defines what kind of this id works for.
	// Resource should be unique.
	Resource  Name      `db:"resource" json:"resource" gorm:"column:resource"`
	MaxID     uint32    `db:"max_id" json:"max_id" gorm:"column:max_id"`
	UpdatedAt time.Time `db:"update_at" json:"update_time" gorm:"column:updated_at"`
}

// TableName is the resource id generator's database table name.
func (IDGenerator) TableName() Name {
	return "id_generators"
}
