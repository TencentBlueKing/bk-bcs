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

// Package model is the model of bcs cmdb
package model

import "gorm.io/gorm"

// Sync 结构体用于定义模型同步的相关信息
type Sync struct {
	// ID 是同步记录的唯一标识符，对应数据库中的主键
	ID int64 `gorm:"column:id;primary_key" json:"id,omitempty"`
	// FullSyncTime 记录全量同步的时间戳
	FullSyncTime int64 `gorm:"column:full_sync_time" json:"full_sync_time,omitempty"`
}

// SyncMigrate 函数用于自动迁移数据库表结构以匹配Sync模型。
// 参数db是一个指向gorm.DB的指针，它提供了与数据库交互的能力。
// 函数调用db.AutoMigrate方法，传入Sync模型的实例指针，实现表的自动迁移。
func SyncMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Sync{})
}

// TableName 方法用于指定Sync模型对应的数据库表名。
// 返回的字符串"sync"表示Sync模型映射到数据库中的表名为"sync"。
func (d Sync) TableName() string {
	return "sync"
}
