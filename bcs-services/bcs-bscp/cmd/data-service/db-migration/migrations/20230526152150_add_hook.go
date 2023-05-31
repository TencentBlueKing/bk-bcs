package migrations

import (
	"time"

	"gorm.io/gorm"

	"bscp.io/cmd/data-service/db-migration/migrator"
)

func init() {
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "mig20230526152150",
		Name:    "20230526152150_add_hook",
		Mode:    migrator.GormMode,
		Up:      mig20230526152150GormTestUp,
		Down:    mig20230526152150GormDown,
	})
}

func mig20230526152150GormTestUp(tx *gorm.DB) error {

	// ConfigHook : 配置脚本
	type ConfigHook struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		PreHookID         uint `gorm:"type:bigint(1) unsigned not null"`
		PreHookReleaseID  uint `gorm:"type:bigint(1) unsigned not null"`
		PostHookID        uint `gorm:"type:bigint(1) unsigned not null"`
		PostHookReleaseID uint `gorm:"type:bigint(1) unsigned not null"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_AppID:2"`
		APPID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_AppID:1"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// Release mapped from table <releases>
	type Release struct {
		PreHookID         uint `gorm:"type:bigint(1) unsigned not null"`
		PreHookReleaseID  uint `gorm:"type:bigint(1) unsigned not null"`
		PostHookID        uint `gorm:"type:bigint(1) unsigned not null"`
		PostHookReleaseID uint `gorm:"type:bigint(1) unsigned not null"`
	}

	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").
		AutoMigrate(&ConfigHook{}, &Release{}); err != nil {
		return err
	}

	now := time.Now()
	if result := tx.Create([]IDGenerators{
		{Resource: "config_hooks", MaxID: 0, UpdatedAt: now},
	}); result.Error != nil {
		return result.Error
	}

	return nil
}

func mig20230526152150GormDown(tx *gorm.DB) error {

	type ConfigHook struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		PreHookID         uint `gorm:"type:bigint(1) unsigned not null"`
		PreHookReleaseID  uint `gorm:"type:bigint(1) unsigned not null"`
		PostHookID        uint `gorm:"type:bigint(1) unsigned not null"`
		PostHookReleaseID uint `gorm:"type:bigint(1) unsigned not null"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null"`
		APPID uint `gorm:"type:bigint(1) unsigned not null uniqueIndex:idx_AppID:1"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// Release mapped from table <releases>
	type Release struct {
		PreHookID         uint `gorm:"type:bigint(1) unsigned not null"`
		PreHookReleaseID  uint `gorm:"type:bigint(1) unsigned not null"`
		PostHookID        uint `gorm:"type:bigint(1) unsigned not null"`
		PostHookReleaseID uint `gorm:"type:bigint(1) unsigned not null"`
	}

	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Migrator().DropTable(ConfigHook{}); err != nil {
		return err
	}

	if err := tx.Migrator().DropColumn(Release{}, "pre_hook_id"); err != nil {
		return err
	}
	if err := tx.Migrator().DropColumn(Release{}, "pre_hook_release_id"); err != nil {
		return err
	}
	if err := tx.Migrator().DropColumn(Release{}, "post_hook_id"); err != nil {
		return err
	}
	if err := tx.Migrator().DropColumn(Release{}, "post_hook_release_id"); err != nil {
		return err
	}

	if err := tx.Migrator().
		DropTable("template_spaces", "templates", "template_releases", "template_sets"); err != nil {
		return err
	}

	if result := tx.Where("resource = ?", "config_hooks").Delete(&IDGenerators{}); result.Error != nil {
		return result.Error
	}

	return nil
}
