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

package mock

import (
	"fmt"

	"github.com/stretchr/testify/mock"
)

// MockRequester define mock requester
type MockRequester struct {
	mock.Mock
}

// NewMockRequester new MockRequester
func NewMockRequester() *MockRequester {
	return &MockRequester{}
}

// DoGetRequest do request
func (m *MockRequester) DoGetRequest(url string, header map[string]string) ([]byte, error) {
	rawNodeList := []byte("{\"kind\":\"NodeList\",\"apiVersion\":\"v1\",\"metadata\":{\"selfLink\":\"/api/v1/nodes\",\"resourceVersion\":\"47752935\"},\"items\":[{\"metadata\":{\"name\":\"1.1.1.1\",\"selfLink\":\"/api/v1/nodes/1.1.1.1\",\"uid\":\"952b3dc0-c9d1-47d2-94b8-e71af222b4c7\",\"resourceVersion\":\"47752848\",\"creationTimestamp\":\"2022-05-10T10:26:05Z\",\"labels\":{\"beta.kubernetes.io/arch\":\"amd64\",\"beta.kubernetes.io/instance-type\":\"S5.LARGE16\",\"beta.kubernetes.io/os\":\"linux\",\"cloud.tencent.com/node-instance-id\":\"ins-d3cwxd6w\",\"failure-domain.beta.kubernetes.io/region\":\"nj\",\"failure-domain.beta.kubernetes.io/zone\":\"330001\",\"kubernetes.io/arch\":\"amd64\",\"kubernetes.io/hostname\":\"1.1.1.1\",\"kubernetes.io/os\":\"linux\",\"module\":\"prometheus\",\"testupdate\":\"testupdate\",\"testupdate2\":\"testupdate3\",\"testupdate3\":\"testupdate3\"},\"annotations\":{\"node.alpha.kubernetes.io/ttl\":\"0\",\"volumes.kubernetes.io/controller-managed-attach-detach\":\"true\"}},\"spec\":{\"podCIDR\":\"1.1.1.1/27\",\"podCIDRs\":[\"1.1.1.1/27\"],\"providerID\":\"qcloud:///330001/ins-d3cwxd6w\"},\"status\":{\"capacity\":{\"cpu\":\"4\",\"ephemeral-storage\":\"103079844Ki\",\"hugepages-1Gi\":\"0\",\"hugepages-2Mi\":\"0\",\"memory\":\"16128872Ki\",\"pods\":\"29\"},\"allocatable\":{\"cpu\":\"3920m\",\"ephemeral-storage\":\"94998384074\",\"hugepages-1Gi\":\"0\",\"hugepages-2Mi\":\"0\",\"memory\":\"14695272Ki\",\"pods\":\"29\"},\"conditions\":[],\"addresses\":[{\"type\":\"InternalIP\",\"address\":\"1.1.1.1\"},{\"type\":\"Hostname\",\"address\":\"1.1.1.1\"}],\"daemonEndpoints\":{\"kubeletEndpoint\":{\"Port\":10250}},\"nodeInfo\":{},\"images\":[]}},{\"metadata\":{\"name\":\"2.2.2.2\",\"selfLink\":\"/api/v1/nodes/2.2.2.2\",\"uid\":\"a23b157d-03dd-49bd-a0fc-e7d2d8ec17a6\",\"resourceVersion\":\"47752784\",\"creationTimestamp\":\"2022-06-15T15:48:09Z\",\"labels\":{\"beta.kubernetes.io/arch\":\"amd64\",\"beta.kubernetes.io/instance-type\":\"SA3.2XLARGE64\",\"beta.kubernetes.io/os\":\"linux\",\"cloud.tencent.com/node-instance-id\":\"ins-6xv97uyw\",\"failure-domain.beta.kubernetes.io/region\":\"nj\",\"failure-domain.beta.kubernetes.io/zone\":\"330003\",\"kubernetes.io/arch\":\"amd64\",\"kubernetes.io/hostname\":\"2.2.2.2\",\"kubernetes.io/os\":\"linux\"},\"annotations\":{\"node.alpha.kubernetes.io/ttl\":\"0\",\"volumes.kubernetes.io/controller-managed-attach-detach\":\"true\"}},\"spec\":{\"podCIDR\":\"11.157.243.96/27\",\"podCIDRs\":[\"11.157.243.96/27\"],\"providerID\":\"qcloud:///330003/ins-6xv97uyw\"},\"status\":{\"capacity\":{\"cpu\":\"8\",\"ephemeral-storage\":\"103079844Ki\",\"hugepages-1Gi\":\"0\",\"hugepages-2Mi\":\"0\",\"memory\":\"64567304Ki\",\"pods\":\"29\"},\"allocatable\":{\"cpu\":\"7910m\",\"ephemeral-storage\":\"94998384074\",\"hugepages-1Gi\":\"0\",\"hugepages-2Mi\":\"0\",\"memory\":\"60946440Ki\",\"pods\":\"29\"},\"conditions\":[],\"addresses\":[{\"type\":\"InternalIP\",\"address\":\"2.2.2.2\"},{\"type\":\"Hostname\",\"address\":\"2.2.2.2\"}],\"daemonEndpoints\":{\"kubeletEndpoint\":{\"Port\":10250}},\"nodeInfo\":{},\"images\":[]}}]}\n")
	m.On("DoGetRequest", fmt.Sprintf("/clusters/%s/api/v1/nodes?labelSelector=node-role.kubernetes.io/master!=true", "BCS-K8S-15202")).
		Return(rawNodeList, nil)
	args := m.Called(url)
	return args.Get(0).([]byte), args.Error(1)
}

// DoPostRequest do post request
func (m *MockRequester) DoPostRequest(url string, header map[string]string, data []byte) ([]byte, error) {
	return nil, nil
}

// DoPutRequest do put request
func (m *MockRequester) DoPutRequest(url string, header map[string]string, data []byte) ([]byte, error) {
	return nil, nil
}

// DoPatchRequest do patch request
func (m *MockRequester) DoPatchRequest(url string, header map[string]string, data []byte) ([]byte, error) {
	m.On("DoPatchRequest", fmt.Sprintf("/clusters/%s/api/v1/nodes/%s", "BCS-K8S-15202", "2.2.2.2")).
		Return([]byte(""), nil)
	m.On("DoPatchRequest", fmt.Sprintf("/clusters/%s/api/v1/nodes/%s", "BCS-K8S-15202", "1.1.1.1")).
		Return([]byte(""), nil)
	args := m.Called(url)
	return args.Get(0).([]byte), args.Error(1)
}

// DoDeleteRequest do delete request
func (m *MockRequester) DoDeleteRequest(url string, header map[string]string) ([]byte, error) {
	return nil, nil
}
