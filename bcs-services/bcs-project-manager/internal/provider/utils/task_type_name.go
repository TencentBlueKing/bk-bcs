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

// Package utils xxx
package utils

import "fmt"

// TaskTypeName task type & name
type taskTypeName struct {
	taskType string
	taskName string
}

// GetTaskType 获取任务类型
func (t taskTypeName) GetTaskType(provider string) string {
	return fmt.Sprintf("%s-%s", provider, t.taskType)
}

// GetJobType 获取job类型
func (t taskTypeName) GetJobType() string {
	return t.taskType
}

// GetTaskName 获取任务名称
func (t taskTypeName) GetTaskName() string {
	return t.taskName
}

var (
	// TestExample example 测试任务
	TestExample = taskTypeName{
		taskType: "TestTask",
		taskName: "测试任务",
	}
	// CreateProjectQuota 创建项目配额
	CreateProjectQuota = taskTypeName{
		taskType: "CreateProjectQuota",
		taskName: "创建项目配额",
	}
	// DeleteProjectQuota 删除项目配额
	DeleteProjectQuota = taskTypeName{
		taskType: "DeleteProjectQuota",
		taskName: "删除项目配额",
	}
	// ScaleUpProjectQuota 调增项目配额
	ScaleUpProjectQuota = taskTypeName{
		taskType: "ScaleUpProjectQuota",
		taskName: "调增项目配额",
	}
	// ScaleDownProjectQuota 调减项目配额
	ScaleDownProjectQuota = taskTypeName{
		taskType: "ScaleDownProjectQuota",
		taskName: "调减项目配额",
	}
)
