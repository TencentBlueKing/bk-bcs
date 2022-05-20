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

package datajob

import (
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func Test_getWorkloadCount(t *testing.T) {
	opts1 := &common.JobCommonOpts{
		ObjectType:  common.NamespaceType,
		ClusterID:   "BCS-K8S-15091",
		ClusterType: common.Kubernetes,
		Namespace:   "bcs-system",
		Dimension:   "minute",
		CurrentTime: time.Time{},
	}

	storageCli := &mock.MockStorage{}
	clients := &Clients{
		bcsStorageCli: storageCli,
	}
	result := getWorkloadCount(opts1, clients)
	assert.Equal(t, int64(4), result)
	opts2 := &common.JobCommonOpts{
		ObjectType:  common.NamespaceType,
		ClusterID:   "BCS-K8S-15091",
		ClusterType: common.Kubernetes,
		Namespace:   "test",
		Dimension:   "minute",
		CurrentTime: time.Time{},
	}
	result2 := getWorkloadCount(opts2, clients)
	assert.Equal(t, int64(0), result2)
	opts3 := &common.JobCommonOpts{
		ObjectType:  common.ClusterType,
		ClusterID:   "BCS-K8S-15091",
		ClusterType: common.Kubernetes,
		Namespace:   "test",
		Dimension:   "minute",
		CurrentTime: time.Time{},
	}
	result3 := getWorkloadCount(opts3, clients)
	assert.Equal(t, int64(4), result3)
}
