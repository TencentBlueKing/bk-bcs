/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pkg

import (
	"testing"

	datamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"github.com/stretchr/testify/assert"
)

func TestNewClientWithConfiguration(t *testing.T) {
	client := NewDataManagerCli(&Config{
		APIServer: "",
		AuthToken: "",
	})
	rsp, err := client.GetClusterInfo(&datamanager.GetClusterInfoRequest{
		ClusterID: "BCS-K8S-15202",
		Dimension: "hour",
	})
	assert.Nil(t, err)
	assert.NotNil(t, rsp)

}

func Test_GetClusterInfo(t *testing.T) {
	client := NewDataManagerCli(&Config{
		APIServer: "",
		AuthToken: "",
	})
	rsp, err := client.GetClusterInfo(&datamanager.GetClusterInfoRequest{
		ClusterID: "BCS-K8S-15202",
		Dimension: "hour",
	})
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}

func Test_GetClusterInfoList(t *testing.T) {
	client := NewDataManagerCli(&Config{
		APIServer: "127.0.0.1:8081",
		AuthToken: "",
	})
	rsp, err := client.GetClusterInfoList(&datamanager.GetClusterInfoListRequest{
		ProjectID: "111",
		Dimension: "hour",
	})
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}

func Test_GetNamespaceInfo(t *testing.T) {
	client := NewDataManagerCli(&Config{
		APIServer: "127.0.0.1:8081",
		AuthToken: "",
	})
	rsp, err := client.GetNamespaceInfo(&datamanager.GetNamespaceInfoRequest{
		ClusterID: "111",
		Dimension: "hour",
	})
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}

func Test_GetNamespaceInfoList(t *testing.T) {
	client := NewDataManagerCli(&Config{
		APIServer: "127.0.0.1:8081",
		AuthToken: "",
	})
	rsp, err := client.GetNamespaceInfoList(&datamanager.GetNamespaceInfoListRequest{
		ClusterID: "111",
		Dimension: "hour",
	})
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}

func Test_GetProjectInfo(t *testing.T) {
	client := NewDataManagerCli(&Config{
		APIServer: "127.0.0.1:8081",
		AuthToken: "",
	})
	rsp, err := client.GetProjectInfo(&datamanager.GetProjectInfoRequest{
		ProjectID: "111",
		Dimension: "hour",
	})
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}

func Test_GetWorkloadInfo(t *testing.T) {
	client := NewDataManagerCli(&Config{
		APIServer: "127.0.0.1:8081",
		AuthToken: "",
	})
	rsp, err := client.GetWorkloadInfo(&datamanager.GetWorkloadInfoRequest{
		ClusterID: "111",
		Dimension: "hour",
	})
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}

func Test_GetWorkloadInfoList(t *testing.T) {
	client := NewDataManagerCli(&Config{
		APIServer: "127.0.0.1:8081",
		AuthToken: "",
	})
	rsp, err := client.GetWorkloadInfoList(&datamanager.GetWorkloadInfoListRequest{
		ClusterID: "111",
		Dimension: "hour",
	})
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}
