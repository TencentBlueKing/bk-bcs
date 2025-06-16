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

// Package version xxx
package version

import (
	"fmt"
	"runtime"

	"k8s.io/klog/v2"
)

var (
	// Version describes the components version
	Version = "default"
	// Commit describes the components commit
	Commit = "default"
	// BuildTime describes the components build time
	BuildTime = "unknow"
)

// Print prints the components version
func Print() {
	klog.Info(fmt.Sprintf("Version: %s", Version))
	klog.Info(fmt.Sprintf("Commit: %s", Commit))
	klog.Info(fmt.Sprintf("BuildTime: %s", BuildTime))
	klog.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	klog.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
}
