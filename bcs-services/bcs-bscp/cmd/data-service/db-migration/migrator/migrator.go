package migrator

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"bscp.io/cmd/data-service/db-migration"
	"bscp.io/pkg/criteria/constant"
)

var allSQLFiles []string

// Migration is one specific db migration
type Migration struct {
	Version string
	Name    string
	Up      func(*sql.Tx) error
	Down    func(*sql.Tx) error

	done bool
}

// Migrator is the controller for all migrations
type Migrator struct {
	db            *sql.DB
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
func Init(db *sql.DB) (*Migrator, error) {
	migrator.db = db

	// Create `schema_migrations` table to remember which migrations were executed.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
		applied DATETIME ( 6 ) NOT NULL,
		version VARCHAR ( 255 )
	);`); err != nil {
		fmt.Println("Unable to create `schema_migrations` table", err)
		return migrator, err
	}

	// Find out all the executed migrations
	rows, err := db.Query("SELECT version FROM `schema_migrations`;")
	if err != nil {
		return migrator, err
	}

	defer rows.Close()

	// Mark the migrations as Done if it is already executed
	for rows.Next() {
		var version string
		err := rows.Scan(&version)
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
	tx, err := m.db.BeginTx(context.TODO(), &sql.TxOptions{})
	if err != nil {
		return err
	}

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

		if _, err := tx.Exec("INSERT INTO `schema_migrations` (applied, version) VALUES(?, ?)",
			time.Now().Format(constant.TimeStdFormat), mg.Version); err != nil {
			tx.Rollback()
			return err
		}
		fmt.Println("Finished running migration", mg.Version)

		count++
	}

	tx.Commit()

	return nil
}

// Down execute backward migration
func (m *Migrator) Down(step int) error {
	tx, err := m.db.BeginTx(context.TODO(), &sql.TxOptions{})
	if err != nil {
		return err
	}

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

		if _, err := tx.Exec("DELETE FROM `schema_migrations` WHERE version = ?", mg.Version); err != nil {
			tx.Rollback()
			return err
		}
		fmt.Println("Finished reverting migration", mg.Version)

		count++
	}

	tx.Commit()

	return nil
}

// MigrationStatus get the current migration status
func (m *Migrator) MigrationStatus() error {
	for _, v := range m.Versions {
		mg := m.Migrations[v]

		if mg.done {
			fmt.Println(fmt.Sprintf("Migration %s completed", mg.Name))
		} else {
			fmt.Println(fmt.Sprintf("Migration %s pending", mg.Name))
		}
	}

	return nil
}

// Create generate one migration template file to use
func Create(name string) error {
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

	var out bytes.Buffer

	t := template.Must(template.ParseFiles("./cmd/data-service/db-migration/migrator/schema.tmpl"))
	err := t.Execute(&out, in)
	if err != nil {
		return errors.New("Unable to execute template:" + err.Error())
	}

	migrationFile, err := os.Create(fmt.Sprintf("./cmd/data-service/db-migration/migrations/%s_%s.go",
		version, fileName))
	if err != nil {
		return errors.New("Unable to create migration file:" + err.Error())
	}
	defer migrationFile.Close()

	if _, err := migrationFile.WriteString(out.String()); err != nil {
		return errors.New("Unable to write to migration file:" + err.Error())
	}

	upSQLFile, err := os.Create(fmt.Sprintf("./cmd/data-service/db-migration/migrations/sql/%s_%s_up.sql",
		version, fileName))
	if err != nil {
		return errors.New("Unable to create up sql file:" + err.Error())
	}
	defer upSQLFile.Close()

	downSQLFile, err := os.Create(fmt.Sprintf("./cmd/data-service/db-migration/migrations/sql/%s_%s_down.sql",
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
	funcName = strings.ReplaceAll(strings.Title(strings.ReplaceAll(name, "_", "-")), "-", "")
	return
}

// getMigrationSQLs get migration sql from specific files
func getMigrationSQLs() (map[string]string, error) {
	migrationSQLs := make(map[string]string)
	dir := "migrations/sql"
	getAllSQLFiles(dir)
	for _, file := range allSQLFiles {
		content, err := db_migration.SQLFiles.ReadFile(file)
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
	entries, err := db_migration.SQLFiles.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		file := filepath.Join(dir, e.Name())
		if e.IsDir() {
			getAllSQLFiles(file)
		} else {
			if filepath.Ext(file) == ".sql" {
				allSQLFiles = append(allSQLFiles, file)
			}
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
