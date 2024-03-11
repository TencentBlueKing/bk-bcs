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

package migrations

import (
	"time"

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/db-migration/migrator"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/vault"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

func init() {
	// add current migration to migrator
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20240125175500",
		Name:    "20240125175500_kv_add_signature",
		Mode:    migrator.GormMode,
		Up:      mig20240125175500Up,
		Down:    mig20240125175500Down,
	})
}

// Kvs20240125175500 kv
type Kvs20240125175500 struct {
	ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

	// Spec is specifics of the resource defined with user
	Key     string `gorm:"type:varchar(255) not null;uniqueIndex:idx_bizID_appID_key_kvState,priority:1"`
	Version uint   `gorm:"type:bigint(1) unsigned not null;"`
	KvType  string `gorm:"type:varchar(64) not null"`
	KvState string `gorm:"type:varchar(64) not null;uniqueIndex:idx_bizID_appID_key_kvState,priority:2"`

	// Attachment is attachment info of the resource
	BizID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID_key_kvState,priority:3"`
	APPID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID_key_kvState,priority:4"`

	// Revision is revision info of the resource
	Creator   string    `gorm:"type:varchar(64) not null"`
	Reviser   string    `gorm:"type:varchar(64) not null"`
	CreatedAt time.Time `gorm:"type:datetime(6) not null"`
	UpdatedAt time.Time `gorm:"type:datetime(6) not null"`

	Signature string `gorm:"type:varchar(64) not null"`
	Md5       string `gorm:"type:varchar(64) not null"`
	ByteSize  uint   `gorm:"type:bigint(1) unsigned not null"`
}

// TableName gorm table name
func (Kvs20240125175500) TableName() string {
	t := &table.Kv{}
	return t.TableName()
}

// ReleasedKvs20240125175500 已生成版本的kv
type ReleasedKvs20240125175500 struct {
	ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

	// Spec is specifics of the resource defined with user
	Key       string `gorm:"type:varchar(255) not null;uniqueIndex:relID_key,priority:1"`
	Version   uint   `gorm:"type:bigint(1) unsigned not null;"`
	ReleaseID uint   `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_ID,priority:3;uniqueIndex:relID_key,priority:2"` //nolint:lll
	KvType    string `gorm:"type:varchar(64) not null"`

	// Attachment is attachment info of the resource
	BizID uint `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_ID,priority:1"`
	AppID uint `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_ID,priority:2"`

	// Revision is revision info of the resource
	Creator   string    `gorm:"type:varchar(64) not null"`
	Reviser   string    `gorm:"type:varchar(64) not null"`
	CreatedAt time.Time `gorm:"type:datetime(6) not null"`
	UpdatedAt time.Time `gorm:"type:datetime(6) not null"`

	Signature string `gorm:"type:varchar(64) not null"`
	Md5       string `gorm:"type:varchar(64) not null"`
	ByteSize  uint   `gorm:"type:bigint(1) unsigned not null"`
}

// TableName gorm table name
func (ReleasedKvs20240125175500) TableName() string {
	t := &table.ReleasedKv{}
	return t.TableName()
}

func syncSignature(tx *gorm.DB) error {
	// set default value
	var kvs []table.Kv

	tx.Model(&table.Kv{}).Find(&kvs)
	cli, err := vault.NewSet(cc.DataService().Vault)
	if err != nil {
		return err
	}

	for _, kv := range kvs {
		if kv.ContentSpec.Signature != "" && kv.ContentSpec.Md5 != "" {
			continue
		}

		opt := &types.GetLastKvOpt{BizID: kv.Attachment.BizID, AppID: kv.Attachment.AppID, Key: kv.Spec.Key}
		_, value, err := cli.GetLastKv(kit.New(), opt)
		if err != nil {
			return err
		}

		kv.ContentSpec.Signature = tools.SHA256(value)
		kv.ContentSpec.Md5 = tools.MD5(value)
		kv.ContentSpec.ByteSize = uint64(len(value))

		// 只更新必须的字段, 不刷新updated_at
		tx.Select("Signature", "ByteSize", "Md5").UpdateColumns(&kv)
	}

	return nil
}

func syncReleaseSignature(tx *gorm.DB) error {
	// set default value
	var kvs []table.ReleasedKv

	tx.Model(&table.ReleasedKv{}).Find(&kvs)
	cli, err := vault.NewSet(cc.DataService().Vault)
	if err != nil {
		return err
	}

	for _, kv := range kvs {
		if kv.ContentSpec.Signature != "" && kv.ContentSpec.Md5 != "" {
			continue
		}

		// 获取 release 版本的值
		opt := &types.GetRKvOption{
			BizID:      kv.Attachment.BizID,
			AppID:      kv.Attachment.AppID,
			Key:        kv.Spec.Key,
			ReleasedID: kv.ReleaseID,
			Version:    int(kv.Spec.Version),
		}
		_, value, err := cli.GetRKv(kit.New(), opt)
		if err != nil {
			return err
		}

		kv.ContentSpec.Signature = tools.SHA256(value)
		kv.ContentSpec.ByteSize = uint64(len(value))
		kv.ContentSpec.Md5 = tools.MD5(value)

		// 只更新必须的字段, 不刷新updated_at
		tx.Select("Signature", "ByteSize", "Md5").UpdateColumns(&kv)
	}

	return nil
}

// mig20240125175500Up for up migration
func mig20240125175500Up(tx *gorm.DB) error {
	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").
		AutoMigrate(&Kvs20240125175500{}); err != nil {
		return err
	}

	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").
		AutoMigrate(&ReleasedKvs20240125175500{}); err != nil {
		return err
	}

	if err := syncSignature(tx); err != nil {
		return err
	}

	if err := syncReleaseSignature(tx); err != nil {
		return err
	}

	return nil
}

// mig20240125175500Down for down migration
func mig20240125175500Down(tx *gorm.DB) error {
	// delete kvs old column
	if tx.Migrator().HasColumn(&Kvs20240125175500{}, "Signature") {
		if err := tx.Migrator().DropColumn(&Kvs20240125175500{}, "Signature"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&Kvs20240125175500{}, "ByteSize") {
		if err := tx.Migrator().DropColumn(&Kvs20240125175500{}, "ByteSize"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&Kvs20240125175500{}, "Md5") {
		if err := tx.Migrator().DropColumn(&Kvs20240125175500{}, "Md5"); err != nil {
			return err
		}
	}

	// delete release_kvs old column
	if tx.Migrator().HasColumn(&ReleasedKvs20240125175500{}, "Signature") {
		if err := tx.Migrator().DropColumn(&ReleasedKvs20240125175500{}, "Signature"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&ReleasedKvs20240125175500{}, "ByteSize") {
		if err := tx.Migrator().DropColumn(&ReleasedKvs20240125175500{}, "ByteSize"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&ReleasedKvs20240125175500{}, "Md5") {
		if err := tx.Migrator().DropColumn(&Kvs20240125175500{}, "Md5"); err != nil {
			return err
		}
	}

	return nil
}
