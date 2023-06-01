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
 *
 */

package util

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql" // import mysql drive, used to create conn.
	"github.com/jmoiron/sqlx"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/logs"
)

// LogConfig is log config
type LogConfig struct {
	Verbosity uint
}

// DBConfig is db config
type DBConfig struct {
	IP       string
	Port     int64
	User     string
	Password string
	DB       string
}

// SetLogger set logger
func SetLogger(logCfg LogConfig) {
	logs.InitLogger(
		logs.LogConfig{
			ToStdErr:       true,
			LogLineMaxSize: 2,
			Verbosity:      logCfg.Verbosity,
		},
	)
}

// NewDB new a db instance
func NewDB(dbCfg DBConfig) *sqlx.DB {
	dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=UTC",
		dbCfg.User, dbCfg.Password, dbCfg.IP, dbCfg.Port, dbCfg.DB)
	db := sqlx.MustConnect("mysql", dsn)
	db.SetMaxOpenConns(500)
	db.SetMaxIdleConns(5)
	return db
}

// ClearDB clear bscp database
func ClearDB(db *sqlx.DB) error {
	if _, err := db.Exec("truncate table " + table.AppTable.Name()); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + table.ArchivedAppTable.Name()); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + string(table.ContentTable)); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + table.ConfigItemTable.Name()); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + table.CommitsTable.Name()); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + table.ReleaseTable.Name()); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + table.ReleasedConfigItemTable.Name()); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + table.StrategySetTable.Name()); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + table.StrategyTable.Name()); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + table.CurrentPublishedStrategyTable.Name()); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + table.PublishedStrategyHistoryTable.Name()); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + table.CurrentReleasedInstanceTable.Name()); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + table.EventTable.Name()); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + table.AuditTable.Name()); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + table.ResourceLockTable.Name()); err != nil {
		return err
	}
	if _, err := db.Exec("update " + table.IDGeneratorTable.Name() +
		" t1 set t1.max_id = 0 where resource != 'events'"); err != nil {
		return err
	}
	if _, err := db.Exec("update " + table.IDGeneratorTable.Name() +
		" t1 set t1.max_id = 500 where resource = 'events'"); err != nil {
		return err
	}

	return nil
}

// TxnIsolationLevel is transaction isolation level
type TxnIsolationLevel string

const (
	RepeatableRead TxnIsolationLevel = "repeatable read"
	ReadCommitted  TxnIsolationLevel = "read committed"
)

// SetTxnIsolationLevel set transaction isolation level for stress test
func SetTxnIsolationLevel(db *sqlx.DB, level TxnIsolationLevel) error {
	if _, err := db.Exec("set global transaction isolation level " + string(level)); err != nil {
		return err
	}
	return nil
}
