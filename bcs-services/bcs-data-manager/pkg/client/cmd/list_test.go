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

package cmd

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/client/pkg"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"github.com/stretchr/testify/assert"
)

func TestListCluster(t *testing.T) {

}

func TestListNamespace(t *testing.T) {
	client := pkg.NewDataManagerCli(&pkg.Config{
		APIServer: "",
		AuthToken: "",
	})
	rsp, err := client.GetNamespaceInfoList(&bcsdatamanager.GetNamespaceInfoListRequest{
		ClusterID: "BCS-K8S-15091",
		Dimension: "hour",
		Page:      1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}

func TestListWorkload(t *testing.T) {

}
