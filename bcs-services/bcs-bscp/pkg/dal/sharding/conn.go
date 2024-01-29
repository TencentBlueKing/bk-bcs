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

package sharding

import (
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // import mysql drive, used to create conn.
	"github.com/jmoiron/sqlx"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
)

// connect to mysql
func connect(opt cc.Database) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", URI(opt))
	if err != nil {
		return nil, fmt.Errorf("connect to mysql failed, err: %v", err)
	}

	db.SetMaxOpenConns(int(opt.MaxOpenConn))
	db.SetMaxIdleConns(int(opt.MaxIdleConn))
	db.SetConnMaxLifetime(time.Duration(opt.MaxIdleTimeoutMin) * time.Minute)

	return db, nil
}

// URI generate the standard db connection string format uri.
func URI(opt cc.Database) string {

	return fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?parseTime=true&loc=UTC&timeout=%ds&readTimeout=%ds&writeTimeout=%ds&charset=%s",
		opt.User,
		opt.Password,
		strings.Join(opt.Endpoints, ","),
		opt.Database,
		opt.DialTimeoutSec,
		opt.ReadTimeoutSec,
		opt.WriteTimeoutSec,
		"utf8mb4",
	)
}
