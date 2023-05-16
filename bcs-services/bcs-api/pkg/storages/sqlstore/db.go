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

// Package sqlstore is main SQL database storage
package sqlstore

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/config"
	"github.com/jinzhu/gorm"
	// import empty mysql package
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

// GCoreDB xxx
var GCoreDB *gorm.DB

// InitCoreDatabase initialize the GLOBAL database object
func InitCoreDatabase(conf *config.ApiServConfig) error {
	if conf == nil {
		return fmt.Errorf("core_database config not init")
	}

	dsn := conf.BKE.DSN
	if dsn == "" {
		return fmt.Errorf("core_database dsn not configured")
	}
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		return err
	}
	db.DB().SetConnMaxLifetime(60 * time.Second)
	db.DB().SetMaxIdleConns(20)
	db.DB().SetMaxOpenConns(20)

	GCoreDB = db
	return nil
}
