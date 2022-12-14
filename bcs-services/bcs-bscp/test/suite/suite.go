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
)

type dbConfig struct {
	IP       string
	Port     int64
	User     string
	Password string
	DB       string
}

func init() {
	var clientCfg client.Config
	var concurrent int
	var sustainSeconds float64
	var totalRequest int64

	flag.StringVar(&clientCfg.ApiHost, "api-host", "http://127.0.0.1:8080", "api http server address")
	flag.StringVar(&clientCfg.CacheHost, "cache-host", "127.0.0.1:8081", "cache rpc service address")
	flag.StringVar(&clientCfg.FeedHost, "feed-host", "127.0.0.1:9091", "feed rpc server address")
	flag.IntVar(&concurrent, "concurrent", 1000, "concurrent request during the load test.")
	flag.Float64Var(&sustainSeconds, "sustain-seconds", 10, "the load test sustain time in seconds ")
	flag.Int64Var(&totalRequest, "total-request", 0, "the load test total request,it has higher priority than "+
		"SustainSeconds")
	flag.StringVar(&dbCfg.IP, "mysql-ip", "127.0.0.1", "mysql ip address")
	flag.Int64Var(&dbCfg.Port, "mysql-port", 3306, "mysql port")
	flag.StringVar(&dbCfg.User, "mysql-user", "root", "mysql login user")
	flag.StringVar(&dbCfg.Password, "mysql-passwd", "admin", "mysql login password")
	flag.StringVar(&dbCfg.DB, "mysql-db", "bk_bscp_admin", "mysql database")
	flag.StringVar(&SidecarStartCmd, "sidecar-start-cmd", "", "sidecar start command")
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
}

// ClearData clear data.
func ClearData() error {
	if _, err := db.Exec("truncate table " + string(table.AppTable)); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + string(table.ArchivedAppTable)); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + string(table.ContentTable)); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + string(table.ConfigItemTable)); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + string(table.CommitsTable)); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + string(table.ReleaseTable)); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + string(table.ReleasedConfigItemTable)); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + string(table.StrategySetTable)); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + string(table.StrategyTable)); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + string(table.CurrentPublishedStrategyTable)); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + string(table.PublishedStrategyHistoryTable)); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + string(table.CurrentReleasedInstanceTable)); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + string(table.EventTable)); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + string(table.AuditTable)); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table " + string(table.ResourceLockTable)); err != nil {
		return err
	}
	if _, err := db.Exec("update " + string(table.IDGeneratorTable) +
		" t1 set t1.max_id = 0 where resource != 'event'"); err != nil {
		return err
	}
	if _, err := db.Exec("update " + string(table.IDGeneratorTable) +
		" t1 set t1.max_id = 500 where resource = 'event'"); err != nil {
		return err
	}

	return nil
}

// GetClient get suite-test client.
func GetClient() *client.Client {
	return cli
}
