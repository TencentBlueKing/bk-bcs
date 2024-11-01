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
	"testing"

	"k8s.io/client-go/tools/clientcmd"
)

func TestGetK8sVersion(t *testing.T) {
	config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
	if err != nil {
		// 处理错误
	}

	clientset, err := GetClientsetByConfig(config)
	if err != nil {
		// 处理错误
	}

	version, err := GetK8sVersion(clientset)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("集群版本:", version)
}
