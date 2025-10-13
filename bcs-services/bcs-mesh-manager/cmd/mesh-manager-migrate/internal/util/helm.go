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

// Package util helm 客户端工具
package util

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/yaml"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
)

// initHelmActionConfig 初始化helm action configuration
func initHelmActionConfig(kubeconfigPath, namespace string) (*action.Configuration, error) {
	// 创建 ConfigFlags
	flags := genericclioptions.NewConfigFlags(false)

	// 正确设置 KubeConfig：创建变量然后取其地址
	kubeconfig := kubeconfigPath
	flags.KubeConfig = &kubeconfig

	// 正确设置 Insecure：创建变量然后取其地址
	insecure := true
	flags.Insecure = &insecure

	// 初始化 action configuration
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(flags, namespace, "", blog.Infof); err != nil {
		return nil, fmt.Errorf("failed to init helm action config: %v", err)
	}

	return actionConfig, nil
}

// GetValues 获取 helm release values
func GetValues(
	clusterID,
	releaseName string,
	kubeconfigPath string,
) ([]byte, error) {
	// 初始化helm action configuration
	actionConfig, err := initHelmActionConfig(kubeconfigPath, common.IstioNamespace)
	if err != nil {
		return nil, err
	}

	// 创建GetValues action
	getValues := action.NewGetValues(actionConfig)
	// 设置 AllValues 为 true，对应实际的命令 helm get values --all
	getValues.AllValues = true

	// 获取values
	values, err := getValues.Run(releaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get helm values: %v", err)
	}

	// 将values转换为YAML格式
	yamlData, err := yaml.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal values to yaml: %v", err)
	}

	return yamlData, nil
}

// GetAppVersion 获取 helm release app version
func GetAppVersion(
	clusterID,
	releaseName string,
	kubeconfigPath string,
) (string, error) {
	// 初始化helm action configuration
	actionConfig, err := initHelmActionConfig(kubeconfigPath, common.IstioNamespace)
	if err != nil {
		return "", err
	}

	// 创建Get action
	get := action.NewGet(actionConfig)

	// 获取release信息
	release, err := get.Run(releaseName)
	if err != nil {
		return "", fmt.Errorf("failed to get helm release: %v", err)
	}

	return release.Chart.Metadata.AppVersion, nil
}
