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

// Package storage xxx
package storage

const (
	tablePreCheckTask = "bcs_gitops_precheck_task" // nolint
)

// Interface xxx interface
type Interface interface {
	Init() error
	CreatePreCheckTask(task *PreCheckTask) (*PreCheckTask, error)
	UpdatePreCheckTask(task *PreCheckTask) error
	ListPreCheckTask(query *PreCheckTaskQuery) ([]*PreCheckTask, error)
	GetPreCheckTask(id int, project string) (*PreCheckTask, error)
}
