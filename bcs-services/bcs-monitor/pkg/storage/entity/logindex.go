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

package entity

// LogIndex for log index
type LogIndex struct {
	ProjectID      string `json:"project_id" bson:"projectID" gorm:"column:project_id"`
	BizID          int    `json:"biz_id" bson:"bizID" gorm:"column:cc_app_id"`
	STDDataID      int    `json:"std_data_id" bson:"stdDataID" gorm:"column:std_data_id"`
	FileDataID     int    `json:"file_data_id" bson:"fileDataID" gorm:"column:file_data_id"`
	STDIndexSetID  int    `json:"std_index_set_id" bson:"stdIndexSetID" gorm:"column:std_index_set_id"`
	FileIndexSetID int    `json:"file_index_set_id" bson:"fileIndexSetID" gorm:"column:file_index_set_id"`
}

// TableName return log index table name
func (LogIndex) TableName() string {
	return "datalog_datalogplan"
}
