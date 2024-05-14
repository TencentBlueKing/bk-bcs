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

// Package k8s xxx
package k8s

import (
	"fmt"
	"time"

	version2 "k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
)

// GetK8sVersion get cluster k8s version
func GetK8sVersion(clientSet *kubernetes.Clientset) (string, error) {
	versionChann := make(chan int)
	var version string
	var versionError error
	var versionInfo *version2.Info
	go func() {
		versionInfo, versionError = clientSet.ServerVersion()
		if versionError != nil {
			version = ""
		} else if versionInfo.GitVersion == "" {
			version = ""
			versionError = fmt.Errorf("get blank result")
		} else {
			version = versionInfo.GitVersion
		}
		versionChann <- 0
	}()

	select {
	case <-time.After(10 * time.Second):
		return "", fmt.Errorf("get k8s version timeout")

	case <-versionChann:
	}

	return version, versionError
}
