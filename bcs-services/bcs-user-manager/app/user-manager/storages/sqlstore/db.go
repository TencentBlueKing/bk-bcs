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

package sqlstore

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // mysql
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/jinzhu/gorm/splunkgorm"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
)

// GCoreDB global DB client
var GCoreDB *gorm.DB

// InitCoreDatabase set DB client
func InitCoreDatabase(conf *config.UserMgrConfig) error {
	if conf == nil {
		return fmt.Errorf("core_database config not init")
	}

	dsn := conf.DSN
	if dsn == "" {
		return fmt.Errorf("core_database dsn not configured")
	}
	db, err := splunkgorm.Open("mysql", dsn)
	if err != nil {
		return err
	}
	db.DB().SetConnMaxLifetime(time.Hour)
	db.DB().SetMaxIdleConns(50)
	db.DB().SetMaxOpenConns(100)

	GCoreDB = db
	return nil
}
