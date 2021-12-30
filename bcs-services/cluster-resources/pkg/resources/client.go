/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resources

import (
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// 创建 k8s local client
func newLocalResourceClient() dynamic.Interface {
	kubeConfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, _ := clientcmd.BuildConfigFromFlags("", kubeConfig)
	client, _ := dynamic.NewForConfig(config)
	return client
}

func newResourceClient() dynamic.Interface {
	return newLocalResourceClient()
}
