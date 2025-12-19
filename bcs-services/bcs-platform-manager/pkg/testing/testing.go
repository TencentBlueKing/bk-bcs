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

// Package testing xxx
package testing

import (
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
)

// GetTestProjectId 获取测试的项目Id
func GetTestProjectId() string {
	projectId := os.Getenv("TEST_PROJECT_ID")
	if projectId == "" {
		return "project00"
	}
	return projectId
}

// GetTestClusterId 获取测试的集群Id
func GetTestClusterId() string {
	clusterId := os.Getenv("TEST_CLUSTER_ID")
	if clusterId == "" {
		return "cluster00"
	}
	return clusterId
}

// GetTestUsername 获取测试的用户名
func GetTestUsername() string {
	username := os.Getenv("TEST_USERNAME")
	if username == "" {
		return "user00"
	}
	return username
}

// GetTestConfigFile 获取测试的配置文件路径
func GetTestConfigFile() string {
	configPath := os.Getenv("TEST_CONFIG_FILE")
	if configPath == "" {
		return "./etc/config_dev.yaml"
	}
	return configPath
}

func initConf() {
	// 初始化BCS配置
	bcsConfContentYaml, err := os.ReadFile(GetTestConfigFile())
	if err != nil {
		panic(err)
	}

	if err = config.G.ReadFrom(bcsConfContentYaml); err != nil {
		panic(err)
	}
}

func init() {
	_, filename, _, _ := runtime.Caller(0)
	// 切换到当前项目目录
	dir, err := filepath.Abs(path.Join(path.Dir(filename), "../../"))
	if err != nil {
		panic(err)
	}
	err = os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	initConf()
}
