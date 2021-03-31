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

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"bk-bscp/internal/database"
)

var (
	// USER user name.
	USER = "root"

	// PASSWD password.
	PASSWD = ""

	// HOST hostname.
	HOST = "localhost"

	// PORT port.
	PORT = 3306

	// DBNAME db name.
	DBNAME = "bscp_test"

	// SYSDBNAME system db name.
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
		tables = append(tables, &database.LocalAuth{})
		tables = append(tables, &database.System{})
		tables = append(tables, &database.Sharding{})
		tables = append(tables, &database.ShardingDB{})
		tables = append(tables, &database.ProcAttr{})
	} else {
		// business database.
		tables = append(tables, &database.App{})
		tables = append(tables, &database.TemplateBind{})
		tables = append(tables, &database.ConfigTemplate{})
		tables = append(tables, &database.ConfigTemplateVersion{})
		tables = append(tables, &database.VariableGroup{})
		tables = append(tables, &database.Variable{})
		tables = append(tables, &database.Config{})
		tables = append(tables, &database.Content{})
		tables = append(tables, &database.Commit{})
		tables = append(tables, &database.AppInstance{})
		tables = append(tables, &database.AppInstanceRelease{})
		tables = append(tables, &database.Release{})
		tables = append(tables, &database.Strategy{})
		tables = append(tables, &database.Audit{})
		tables = append(tables, &database.MultiCommit{})
		tables = append(tables, &database.MultiRelease{})
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&timeout=3s&readTimeout=3s&writeTimeout=3s&charset=%s",
		USER, PASSWD, HOST, PORT, DBNAME, database.BSCPCHARSET)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	for _, table := range tables {
		if err := db.Set("gorm:table_options", fmt.Sprintf("ENGINE=%s", table.DBEngineType())).
			AutoMigrate(table); err != nil {
			panic(err)
		}
	}
}
