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

// Package migrator is the manager of db migrations
package migrator

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/db-migration" //nolint:goimports
)

const (
	// GormMode gorm mode
	GormMode string = "gorm"
	// SqlMode sql mode
	SqlMode string = "sql"
)

var allSQLFiles []string

// Migration is one specific db migration
type Migration struct {
	Version string
	Name    string
	Mode    string
	Up      func(*gorm.DB) error
	Down    func(*gorm.DB) error

	done bool
}

// Migrator is the controller for all migrations
type Migrator struct {
	db            *gorm.DB
	Versions      []string
	Migrations    map[string]*Migration
	MigrationSQLs map[string]string
}

var migrator = &Migrator{
	Versions:      []string{},
	Migrations:    map[string]*Migration{},
	MigrationSQLs: map[string]string{},
}

// AddMigration add one migration to migrator
func (m *Migrator) AddMigration(mg *Migration) {
	// Add the migration to the hash with version as key
	m.Migrations[mg.Version] = mg

	// Insert version into versions array using insertion sort
	index := 0
	for index < len(m.Versions) {
		if m.Versions[index] > mg.Version {
			break
		}
		index++
	}

	m.Versions = append(m.Versions, mg.Version)
	copy(m.Versions[index+1:], m.Versions[index:])
	m.Versions[index] = mg.Version
}

// Init create the db connection and get migrator
func Init(db *gorm.DB) (*Migrator, error) {
	migrator.db = db
	var err error

	// Create `schema_migrations` table to remember which migrations were executed.
	if result := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
		applied DATETIME ( 6 ) NOT NULL,
		version VARCHAR ( 255 )
	);`); result.Error != nil {
		fmt.Println("Unable to create `schema_migrations` table", result.Error)
		return migrator, result.Error
	}

	// Find out all the executed migrations
	rows, err := db.Raw("SELECT version FROM `schema_migrations`;").Rows()
	if err != nil {
		return migrator, err
	}
	defer func() {
		_ = rows.Close()
	}()

	// Mark the migrations as Done if it is already executed
	for rows.Next() {
		var version string
		err = rows.Scan(&version)
		if err != nil {
			return migrator, err
		}

		if migrator.Migrations[version] != nil {
			migrator.Migrations[version].done = true
		}
	}

	migrator.MigrationSQLs, err = getMigrationSQLs()
	if err != nil {
		return migrator, err
	}

	migrator.checkSQLFiles()

	return migrator, nil
}

// GetMigrator get the migrator
func GetMigrator() *Migrator {
	return migrator
}

// Up execute forward migration
func (m *Migrator) Up(step int) error {
	tx := m.db.Begin()

	count := 0
	for _, v := range m.Versions {
		if step > 0 && count == step {
			break
		}

		mg := m.Migrations[v]

		if mg.done {
			continue
		}

		fmt.Println("Running migration", mg.Version)
		if err := mg.Up(tx); err != nil {
			tx.Rollback()
			return err
		}

		if result := tx.Exec("INSERT INTO `schema_migrations` (applied, version) VALUES(?, ?)",
			time.Now().UTC(), mg.Version); result.Error != nil {
			tx.Rollback()
			return result.Error
		}
		fmt.Println("Finished running migration", mg.Version)

		count++
	}

	if err := tx.Commit().Error; err != nil {
		fmt.Printf("commit transaction failed, err: %s", err.Error())
	}

	return nil
}

// Down execute backward migration
func (m *Migrator) Down(step int) error {
	tx := m.db.Begin()

	count := 0
	for _, v := range reverse(m.Versions) {
		if step > 0 && count == step {
			break
		}

		mg := m.Migrations[v]

		if !mg.done {
			continue
		}

		fmt.Println("Reverting Migration", mg.Version)
		if err := mg.Down(tx); err != nil {
			tx.Rollback()
			return err
		}

		if result := tx.Exec("DELETE FROM `schema_migrations` WHERE version = ?", mg.Version); result.Error != nil {
			tx.Rollback()
			return result.Error
		}
		fmt.Println("Finished reverting migration", mg.Version)

		count++
	}

	if err := tx.Commit().Error; err != nil {
		fmt.Printf("commit transaction failed, err: %s", err.Error())
	}

	return nil
}

// MigrationStatus get the current migration status
func (m *Migrator) MigrationStatus() error {
	for _, v := range m.Versions {
		mg := m.Migrations[v]

		if mg.done {
			fmt.Printf("Migration %s completed\n", mg.Name)
		} else {
			fmt.Printf("Migration %s pending\n", mg.Name)
		}
	}

	return nil
}

// Create generate one migration template file to use
func Create(name, mode string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("the name used to create migration file can't be empty")
	}
	fileName, funcName := getName(name)

	version := time.Now().Format("20060102150405")

	in := struct {
		Version  string
		FuncName string
		FileName string
	}{
		Version:  version,
		FuncName: funcName,
		FileName: fileName,
	}

	switch mode {
	case GormMode:
		return genGormMigrationFile(version, fileName, in)
	case SqlMode:
		return genSqlMigrationFile(version, fileName, in)
	default:
		return fmt.Errorf("unsupported migration mode: %s", mode)
	}
}

// genGormMigrationFile gen migration file when using gorm
func genGormMigrationFile(version, fileName string, in interface{}) error {
	var out bytes.Buffer

	t := template.Must(template.ParseFiles("./cmd/data-service/db-migration/migrator/gorm_schema.tmpl"))
	err := t.Execute(&out, in)
	if err != nil {
		return errors.New("Unable to execute template:" + err.Error())
	}

	migrationFile, err := os.Create(fmt.Sprintf("./cmd/data-service/db-migration/migrations/%s_%s.go",
		version, fileName))
	if err != nil {
		return errors.New("Unable to create gorm migration file:" + err.Error())
	}
	defer migrationFile.Close()

	if _, err := migrationFile.WriteString(out.String()); err != nil {
		return errors.New("Unable to write to migration file:" + err.Error())
	}

	fmt.Printf("Generated new migration files:\n%s\n",
		migrationFile.Name())
	return nil
}

// genSqlMigrationFile gen migration file when using sql
func genSqlMigrationFile(version, fileName string, in interface{}) error {
	var out bytes.Buffer

	t := template.Must(template.ParseFiles("./cmd/data-service/db-migration/migrator/sql_schema.tmpl"))
	err := t.Execute(&out, in)
	if err != nil {
		return errors.New("Unable to execute template:" + err.Error())
	}

	migrationFile, err := os.Create(fmt.Sprintf("./cmd/data-service/db-migration/migrations/%s_%s.go",
		version, fileName))
	if err != nil {
		return errors.New("Unable to create sql migration file:" + err.Error())
	}
	defer migrationFile.Close()

	if _, err = migrationFile.WriteString(out.String()); err != nil {
		return errors.New("Unable to write to migration file:" + err.Error())
	}

	upSQLFile, err := os.Create(fmt.Sprintf("./cmd/data-service/db-migration/migrations/sql/%s_%s_up.sql",
		version, fileName))
	if err != nil {
		return errors.New("Unable to create up sql file:" + err.Error())
	}
	defer upSQLFile.Close()

	var downSQLFile *os.File
	downSQLFile, err = os.Create(fmt.Sprintf("./cmd/data-service/db-migration/migrations/sql/%s_%s_down.sql",
		version, fileName))
	if err != nil {
		return errors.New("Unable to create down sql file:" + err.Error())
	}
	defer downSQLFile.Close()

	fmt.Printf("Generated new migration files:\n%s\n%s\n%s\n",
		migrationFile.Name(), upSQLFile.Name(), downSQLFile.Name())
	return nil
}

// checkSQLFiles check if every migration has corresponding sql file
func (m *Migrator) checkSQLFiles() {
	for _, v := range m.Migrations {
		// not check for gore mode
		if v.Mode == GormMode {
			continue
		}
		// only check 'up sql' files, 'down sql' files are optional
		if _, ok := m.MigrationSQLs[GetUpSQLKey(v.Name)]; !ok {
			fmt.Printf("Warning: missing sql file for migration %s, please check!\n", v.Name)
		}
	}
}

// getName get file name and function name
// eg: test-mig-001 ==> (test_mig_001, TestMig001)
func getName(name string) (fileName string, funcName string) {
	fileName = strings.ReplaceAll(name, "-", "_")
	funcName = strings.ReplaceAll(cases.Title(language.English).String(strings.ReplaceAll(name, "_", "-")),
		"-", "")
	return
}

// getMigrationSQLs get migration sql from specific files
func getMigrationSQLs() (map[string]string, error) {
	migrationSQLs := make(map[string]string)
	dir := "migrations/sql"
	if err := getAllSQLFiles(dir); err != nil {
		fmt.Printf("get all sql fiiles(%s) err: %s", dir, err)
	}
	for _, file := range allSQLFiles {
		content, err := dbmigration.SQLFiles.ReadFile(file)
		if err != nil {
			fmt.Printf("read file %s err: %s", file, err)
			return nil, fmt.Errorf("read file %s err: %s", file, err)
		}
		filename := filepath.Base(file)
		migrationSQLs[strings.TrimSuffix(filename, path.Ext(filename))] = string(content)
	}

	return migrationSQLs, nil
}

func getAllSQLFiles(dir string) error {
	entries, err := dbmigration.SQLFiles.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		file := filepath.Join(dir, e.Name())
		if e.IsDir() {
			if err = getAllSQLFiles(file); err != nil {
				return err
			}
		} else if filepath.Ext(file) == ".sql" {
			allSQLFiles = append(allSQLFiles, file)
		}
	}
	return nil
}

func reverse(arr []string) []string {
	for i := 0; i < len(arr)/2; i++ {
		j := len(arr) - i - 1
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

// GetUpSQLKey get the down sql key for MigrationSQLs
func GetUpSQLKey(migrationName string) string {
	return migrationName + "_up"
}

// GetDownSQLKey get the up sql key for MigrationSQLs
func GetDownSQLKey(migrationName string) string {
	return migrationName + "_down"
}
