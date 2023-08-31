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

package install

import (
	"time"
)

// Installer is the interface for app installer
type Installer interface {
	IsInstalled(clusterID string) (bool, error)
	Install(clusterID, values string) error
	Upgrade(clusterID, values string) error
	Uninstall(clusterID string) error
	// CheckAppStatus check app status. pre:true 前置检查；pre:false 后置检查
	CheckAppStatus(clusterID string, timeout time.Duration, pre bool) (bool, error)
	Close()
}

// InstallerType type
type InstallerType string

// String toString
func (it InstallerType) String() string {
	return string(it)
}

var (
	// DefaultCmdFlag xxx
	DefaultCmdFlag = []map[string]interface{}{{"--insecure-skip-tls-verify": ""}, {"--wait": true}}
	// DefaultArgsFlag xxx
	DefaultArgsFlag = []string{"--insecure-skip-tls-verify", "--wait"}
)
