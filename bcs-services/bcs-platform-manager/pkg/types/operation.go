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

// Package types pod types
package types

// ListOperationLogsReq list operation logs request
type ListOperationLogsReq struct {
	ResourceID   string `json:"resourceID" in:"query=resourceID"`
	ResourceName string `json:"resourceName" in:"query=resourceName"`
	Status       string `json:"status" in:"query=status"`
	OpUser       string `json:"opUser" in:"query=opUser"`
	StartTime    uint64 `json:"startTime" in:"query=startTime"`
	EndTime      uint64 `json:"endTime" in:"query=endTime"`
	ResourceType string `json:"resourceType" in:"query=resourceType"`
	Simple       bool   `json:"simple" in:"query=simple"`
	TaskIDNull   bool   `json:"taskIDNull" in:"query=taskIDNull"`
	ClusterID    string `json:"clusterID" in:"query=clusterID"`
	ProjectID    string `json:"projectID" in:"query=projectID"`
	TaskType     string `json:"taskType" in:"query=taskType"`
	V2           bool   `json:"v2" in:"query=v2"`
	IpList       string `json:"ipList" in:"query=ipList"`
	TaskID       string `json:"taskID" in:"query=taskID"`
	TaskName     string `json:"taskName" in:"query=taskName"`
	Limit        uint32 `json:"limit" in:"query=limit"`
	Page         uint32 `json:"page" in:"query=page"`
}

// ListOperationLogsResp list operation logs response
type ListOperationLogsResp struct {
	Count   uint32                `json:"count"`
	Results []*OperationLogDetail `json:"results"`
}

// OperationLogDetail operation log detail
type OperationLogDetail struct {
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceID"`
	TaskID       string `json:"taskID"`
	Message      string `json:"message"`
	OpUser       string `json:"opUser"`
	CreateTime   string `json:"createTime"`
	Task         *Task  `json:"task"`
	TaskType     string `json:"taskType"`
	Status       string `json:"status"`
	ResourceName string `json:"resourceName"`
	AllowRetry   bool   `json:"allowRetry"`
	AllowSkip    bool   `json:"allowSkip"`
}
