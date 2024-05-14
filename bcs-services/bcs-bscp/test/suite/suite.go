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

// Package suite NOTES
package suite

import (
	"flag"
	"log"
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/util"

	_ "github.com/go-sql-driver/mysql" // import mysql drive, used to create conn.
	"github.com/jmoiron/sqlx"
	_ "github.com/smartystreets/goconvey/convey" // import convey.

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/client"
)

var (
	// cli is service request client.
	cli *client.Client
	// dbCfg is db config file.
	dbCfg util.DBConfig
	// DB is db request client.
	DB *sqlx.DB
	// SidecarStartCmd sidecar start cmd, must init data, then start sidecar. if sidecar bind app not exist,
	// sidecar will start failed.
	SidecarStartCmd = ""
	// logCfg is log config
	logCfg util.LogConfig
)

func init() {
	var clientCfg client.Config
	var concurrent int
	var sustainSeconds float64
	var totalRequest int64

	flag.StringVar(&clientCfg.ApiHost, "api-host", "http://127.0.0.1:8080", "api http server address")
	flag.StringVar(&clientCfg.CacheHost, "cache-host", "127.0.0.1:9514", "cache rpc service address")
	flag.StringVar(&clientCfg.FeedHost, "feed-host", "http://127.0.0.1:9610", "feed http server address")
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

	DB = util.NewDB(dbCfg)

	var err error
	if cli, err = client.NewClient(clientCfg); err != nil {
		log.Printf("new suite test client err: %v", err)
		os.Exit(0)
	}
	util.SetLogger(logCfg)
}

// GetClient get suite-test client.
func GetClient() *client.Client {
	return cli
}
