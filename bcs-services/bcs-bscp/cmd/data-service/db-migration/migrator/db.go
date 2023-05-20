/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package migrator

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/dal/sharding"
)

// NewSqlDB new a sql db instance
func NewSqlDB() (*sql.DB, error) {

	fmt.Println("Connecting to MySQL database...")

	dbConf := cc.DataService().Sharding.AdminDatabase
	db, err := sql.Open("mysql", sharding.URI(dbConf))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database, err: %s", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("unable to connect to database, err: %s", err)
	}

	fmt.Println("Database connected!")

	return db, nil
}

// NewGormDB new a gorm db instance from an existing sql db
func NewGormDB(sqlDB *sql.DB, debugGorm bool) (*gorm.DB, error) {
	if debugGorm {
		return gorm.Open(mysql.New(mysql.Config{
			Conn: sqlDB,
		}), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	}
	return gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
}
