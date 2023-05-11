package migrations

import (
	"database/sql"
	"fmt"
	"strings"

	"bscp.io/cmd/data-service/db-migration/migrator"
)

const mig20230511114513 = "20230511114513_add_template"

func init() {
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20230511114513",
		Name:    "20230511114513_add_template",
		Up:      mig20230511114513AddTemplateUp,
		Down:    mig20230511114513AddTemplateDown,
	})
}

func mig20230511114513AddTemplateUp(tx *sql.Tx) error {
	sqlArr := strings.Split(migrator.GetMigrator().MigrationSQLs[migrator.GetUpSQLKey(mig20230511114513)], ";")
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

func mig20230511114513AddTemplateDown(tx *sql.Tx) error {
	sqlArr := strings.Split(migrator.GetMigrator().MigrationSQLs[migrator.GetDownSQLKey(mig20230511114513)], ";")
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
