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

// Package sqlite 是一个 SQLite 数据库的实现，它实现了 Store 接口
package sqlite

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SQLite 结构体包含一个 GORM 数据库实例和数据库文件路径
type SQLite struct {
	DB   *gorm.DB // GORM 数据库实例
	Path string   // 数据库文件路径
}

// New 是 SQLite 的构造函数，它接受一个数据库文件路径作为参数
// 并尝试打开数据库连接，如果成功则返回一个初始化的 SQLite 实例
func New(path string) *SQLite {
	db := Open(path) // 尝试打开数据库连接
	if db == nil {   // 如果打开失败，返回 nil
		return nil
	}
	return &SQLite{ // 返回初始化的 SQLite 实例
		DB:   db,
		Path: path,
	}
}

// Open 函数尝试使用给定的路径打开一个 SQLite 数据库连接
// 如果成功，它返回一个 GORM 数据库实例；如果失败，它会记录错误并返回 nil
func Open(path string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{}) // 尝试打开数据库连接
	if err != nil {                                         // 如果发生错误
		blog.Errorf("Failed to open the SQLite database: %v", err) // 记录错误
		return nil                                                 // 返回 nil
	}
	return db // 返回 GORM 数据库实例
}
