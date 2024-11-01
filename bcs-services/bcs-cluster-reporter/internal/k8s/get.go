/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package k8s xxx
package k8s

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
)

// GetK8sVersion get cluster k8s version
func GetK8sVersion(clientSet *kubernetes.Clientset) (string, error) {
	var versionStr string
	var versionError error
	var versionInfo *version.Info

	ctx, cancel := context.WithTimeout(context.Background(), 11*time.Second)
	go func() {
		versionInfo, versionError = clientSet.ServerVersion()
		if ctx.Err() != nil {
			return
		}
		cancel()
	}()

	select {
	case <-time.After(10 * time.Second):
		return "", fmt.Errorf("get k8s version timeout")

	case <-ctx.Done():
	}

	if versionError != nil {
		versionStr = ""
	} else if versionInfo.GitVersion == "" {
		versionStr = ""
		versionError = fmt.Errorf("get blank result")
	} else {
		versionStr = versionInfo.GitVersion
	}

	return versionStr, versionError
}

// GetK8sApi get apiserver api list
func GetK8sApi(clientSet *kubernetes.Clientset) ([]*v1.APIResourceList, error) {
	apiResources, err := clientSet.Discovery().ServerPreferredResources()
	if err != nil {
		return nil, err
	}

	return apiResources, nil
}
