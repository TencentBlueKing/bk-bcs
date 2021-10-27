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

package v4http

type OperateItem struct {
	Name     string `json:"name"`
	RunAs    string `json:"namespace"`
	SetId    string `json:"setid"`
	ModuleId string `json:"moduleid"`
}

type ScaleOpeParam struct {
	OperateItem
	Instance uint64 `json:"instance"`
}

type DeleteAppOpeParam struct {
	OperateItem
}

type DeleteTaskGroupsOpeParam struct {
	OperateItem
}

type DeleteTaskGroupOpeParam struct {
	OperateItem
	TaskGroupId string `json:"taskgroupid"`
}

type RollbackOpeParam struct {
	OperateItem
}

type FetchVersionOpeParam struct {
	OperateItem
	VersionId string `json:"versionid"`
}

type SendMsgOpeParam struct {
	OperateItem
	MsgType     string      `json:"msgtype"`
	MsgData     interface{} `json:"msgdata"`
	TaskGroupId string      `json:"taskgroupid"`
}
