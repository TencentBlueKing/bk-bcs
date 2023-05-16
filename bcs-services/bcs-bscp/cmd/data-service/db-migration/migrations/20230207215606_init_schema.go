package migrations

import (
	"database/sql"
	"fmt"
	"strings"

	"bscp.io/cmd/data-service/db-migration/migrator"
)

const mig20230207215606 = "20230207215606_init_schema"

func init() {
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20230207215606",
		Name:    "20230207215606_init_schema",
		Up:      mig20230207215606InitSchemaUp,
		Down:    mig20230207215606InitSchemaDown,
	})
}

func mig20230207215606InitSchemaUp(tx *sql.Tx) error {
	sqlArr := strings.Split(migrator.GetMigrator().MigrationSQLs[migrator.GetUpSQLKey(mig20230207215606)], ";")
	for _, sql := range sqlArr {
		sql = strings.TrimSpace(sql)
		if sql == "" {
			continue
		}
		_, err := tx.Exec(sql)
		if err != nil {
			return fmt.Errorf("exec sql [%s] err: %s", sql, err)
		}
	}
	return nil

}

func mig20230207215606InitSchemaDown(tx *sql.Tx) error {
	sqlArr := strings.Split(migrator.GetMigrator().MigrationSQLs[migrator.GetDownSQLKey(mig20230207215606)], ";")
	for _, sql := range sqlArr {
		sql = strings.TrimSpace(sql)
		if sql == "" {
			continue
		}
		_, err := tx.Exec(sql)
		if err != nil {
			return fmt.Errorf("exec sql [%s] err: %s", sql, err)
		}
	}
	return nil
}
