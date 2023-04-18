package migrations

import (
	"database/sql"
	"fmt"
	"strings"

	"bscp.io/cmd/data-service/db-migration/migrator"
)

const mig20230418102723 = "20230418102723_add_credential"

func init() {
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20230418102723",
		Name:    "20230418102723_add_credential",
		Up:      mig20230418102723AddCredentialUp,
		Down:    mig20230418102723AddCredentialDown,
	})
}

func mig20230418102723AddCredentialUp(tx *sql.Tx) error {
	sqlArr := strings.Split(migrator.GetMigrator().MigrationSQLs[migrator.GetUpSQLKey(mig20230418102723)], ";")
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

func mig20230418102723AddCredentialDown(tx *sql.Tx) error {
	sqlArr := strings.Split(migrator.GetMigrator().MigrationSQLs[migrator.GetDownSQLKey(mig20230418102723)], ";")
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
