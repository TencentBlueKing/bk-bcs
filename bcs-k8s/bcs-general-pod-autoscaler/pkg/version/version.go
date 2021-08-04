// Copyright 2021 The BCS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package version

import (
	"fmt"
	"runtime"

	"k8s.io/klog"
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
