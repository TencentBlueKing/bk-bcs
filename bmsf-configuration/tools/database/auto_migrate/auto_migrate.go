/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	"bk-bscp/internal/database"
)

var (
	// USER.
	USER = "root"

	// PASSWD.
	PASSWD = ""

	// HOST.
	HOST = "localhost"

	// PORT.
	PORT = 3306

	// DBNAME.
	DBNAME = "bscp_test"

	// SYSDBNAME.
	SYSDBNAME = "bscpdb"
)

var (
	isSysMigrate = false
)

func main() {
	// all tables;
	tables := []database.Table{}

	if isSysMigrate {
		DBNAME = SYSDBNAME

		// not in business database, but in bscpdb.
		tables = append(tables, &database.Business{})
		tables = append(tables, &database.Sharding{})
		tables = append(tables, &database.ShardingDB{})
	} else {
		// business database.
		tables = append(tables, &database.App{})
		tables = append(tables, &database.Cluster{})
		tables = append(tables, &database.Zone{})
		tables = append(tables, &database.ConfigSet{})
		tables = append(tables, &database.ConfigSetLock{})
		tables = append(tables, &database.Configs{})
		tables = append(tables, &database.Commit{})
		tables = append(tables, &database.AppInstance{})
		tables = append(tables, &database.AppInstanceRelease{})
		tables = append(tables, &database.Release{})
		tables = append(tables, &database.Strategy{})
		tables = append(tables, &database.Audit{})
		tables = append(tables, &database.MultiCommit{})
		tables = append(tables, &database.MultiRelease{})
		tables = append(tables, &database.ProcAttr{})
	}

	db, err := gorm.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&timeout=3s&readTimeout=3s&writeTimeout=3s&charset=%s",
			USER, PASSWD, HOST, PORT, DBNAME, database.BSCPCHARSET))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	for _, table := range tables {
		if db.AutoMigrate(table).Error != nil {
			panic(err)
		}
	}
}
