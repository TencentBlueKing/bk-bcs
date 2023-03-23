/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

// Package suite NOTES
package suite

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/logs"
	"bscp.io/test/client"

	_ "github.com/go-sql-driver/mysql" // import mysql drive, used to create conn.
	"github.com/jmoiron/sqlx"
	_ "github.com/smartystreets/goconvey/convey" // import convey.
)

var (
	// cli is service request client.
	cli *client.Client
	// dbCfg is db config file.
	dbCfg dbConfig
	// db is db request client.
	db *sqlx.DB
	// SidecarStartCmd sidecar start cmd, must init data, then start sidecar. if sidecar bind app not exist,
	// sidecar will start failed.
	SidecarStartCmd = ""
	// logCfg is log config
	logCfg logConfig
)

type dbConfig struct {
	IP       string
	Port     int64
	User     string
	Password string
	DB       string
}

type logConfig struct {
	Verbosity uint
}

func init() {
	var clientCfg client.Config
	var concurrent int
	var sustainSeconds float64
	var totalRequest int64

	flag.StringVar(&clientCfg.ApiHost, "api-host", "http://127.0.0.1:8080", "api http server address")
	flag.StringVar(&clientCfg.CacheHost, "cache-host", "127.0.0.1:9514", "cache rpc service address")
	flag.StringVar(&clientCfg.FeedHost, "feed-host", "127.0.0.1:9510", "feed rpc server address")
	flag.IntVar(&concurrent, "concurrent", 1000, "concurrent request during the load test.")
	flag.Float64Var(&sustainSeconds, "sustain-seconds", 10, "the load test sustain time in seconds ")
	flag.Int64Var(&totalRequest, "total-request", 0, "the load test total request,it has higher priority than "+
		"SustainSeconds")
	flag.StringVar(&dbCfg.IP, "mysql-ip", "127.0.0.1", "mysql ip address")
	flag.Int64Var(&dbCfg.Port, "mysql-port", 3306, "mysql port")
	flag.StringVar(&dbCfg.User, "mysql-user", "root", "mysql login user")
	flag.StringVar(&dbCfg.Password, "mysql-passwd", "root", "mysql login password")
	flag.StringVar(&dbCfg.DB, "mysql-db", "bk_bscp_admin", "mysql database")
	flag.StringVar(&SidecarStartCmd, "sidecar-start-cmd", "", "sidecar start command")
	flag.UintVar(&logCfg.Verbosity, "log-verbosity", 0, "log verbosity")

	testing.Init()
	flag.Parse()

	dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=UTC",
		dbCfg.User, dbCfg.Password, dbCfg.IP, dbCfg.Port, dbCfg.DB)
	db = sqlx.MustConnect("mysql", dsn)
	db.SetMaxOpenConns(500)
	db.SetMaxIdleConns(5)

	var err error
	if cli, err = client.NewClient(clientCfg); err != nil {
		log.Printf("new suite test client err: %v", err)
		os.Exit(0)
	}

	setLogger()
}

func setLogger() {
	logs.InitLogger(
		logs.LogConfig{
			ToStdErr:       true,
			LogLineMaxSize: 2,
			Verbosity:      logCfg.Verbosity,
		},
	)
}

// ClearData clear data.
func ClearData() error {
	if _, err := db.Exec("truncate table " + table.AppTable.Name()); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + table.ArchivedAppTable.Name()); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + string(table.HookTable)); err != nil {
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
		" t1 set t1.max_id = 0 where resource != 'event'"); err != nil {
		return err
	}
	if _, err := db.Exec("update " + table.IDGeneratorTable.Name() +
		" t1 set t1.max_id = 500 where resource = 'event'"); err != nil {
		return err
	}

	return nil
}

// GetClient get suite-test client.
func GetClient() *client.Client {
	return cli
}
