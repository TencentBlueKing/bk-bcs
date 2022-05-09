/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package entity

// Release 定义了chart的部署信息, 存储在helm-manager的数据库中, 用于对部署版本做记录
type Release struct {
	Name         string   `json:"name" bson:"name"`
	Namespace    string   `json:"namespace" bson:"namespace"`
	ClusterID    string   `json:"clusterID" bson:"clusterID"`
	ChartName    string   `json:"chartName" bson:"chartName"`
	ChartVersion string   `json:"chartVersion" bson:"chartVersion"`
	Revision     int      `json:"revision" bson:"revision"`
	Values       []string `json:"values" bson:"values"`
	RollbackTo   int      `json:"rollbackTo" bson:"rollbackTo"`
	CreateBy     string   `json:"createBy" bson:"createBy"`
	CreateTime   int64    `json:"createTime" bson:"createTime"`
}
