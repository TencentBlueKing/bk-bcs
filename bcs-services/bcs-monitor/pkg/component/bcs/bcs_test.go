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

package bcs

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/stretchr/testify/assert"
)

func initConf() {
	_, filename, _, _ := runtime.Caller(0)
	dir, err := filepath.Abs(path.Join(path.Dir(filename), "../../../"))
	if err != nil {
		panic(err)
	}
	err = os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	// 初始化BCS配置
	bcsConfContentYaml, err := ioutil.ReadFile("./etc/config_dev.yaml")
	if err != nil {
		panic(err)
	}

	if err = config.G.ReadFrom(bcsConfContentYaml); err != nil {
		panic(err)
	}
}

func getTestProjectId() string {
	projectId := os.Getenv("TEST_PROJECT_ID")
	if projectId == "" {
		return "project00"
	}
	return projectId
}

func getTestClusterId() string {
	clusterId := os.Getenv("TEST_CLUSTER_ID")
	if clusterId == "" {
		return "cluster00"
	}
	return clusterId
}

func getTestUsername() string {
	username := os.Getenv("TEST_USERNAME")
	if username == "" {
		return "user00"
	}
	return username
}

func TestListClusters(t *testing.T) {
	initConf()
	ctx := context.Background()

	clusters, err := ListClusters(ctx, config.G.BCS, getTestProjectId())
	assert.NoError(t, err)
	assert.Equal(t, len(clusters), 0)
}

func TestGetCluster(t *testing.T) {
	initConf()
	ctx := context.Background()

	cluster, err := GetCluster(ctx, config.G.BCS, getTestProjectId(), getTestClusterId())
	assert.NoError(t, err)
	assert.Equal(t, cluster.ProjectId, getTestProjectId())
}
