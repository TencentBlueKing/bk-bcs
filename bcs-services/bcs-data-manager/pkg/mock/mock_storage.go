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
	"encoding/json"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/stretchr/testify/mock"
)

// MockStorage mock storage
type MockStorage struct {
	mock.Mock
}

// NewMockStorage new mock storage
func NewMockStorage() bcsapi.Storage {
	mockStorage := &MockStorage{}
	return mockStorage
}

// QueryMesosTaskgroup search all taskgroup by clusterID
func (m *MockStorage) QueryMesosTaskgroup(cluster string) ([]*storage.Taskgroup, error) {
	return nil, nil
}

// QueryK8SPod query all pod information in specified cluster
func (m *MockStorage) QueryK8SPod(cluster string) ([]*storage.Pod, error) {
	return nil, nil
}

// GetIPPoolDetailInfo get all underlay ip information
func (m *MockStorage) GetIPPoolDetailInfo(clusterID string) ([]*storage.IPPool, error) {
	return nil, nil
}

// ListCustomResource list custom resources, Unmarshalled to dest.
// dest should be a pointer to a struct of map[string]interface{}
func (m *MockStorage) ListCustomResource(resourceType string, filter map[string]string, dest interface{}) error {
	return nil
}

// PutCustomResource put custom resources, support map or struct
func (m *MockStorage) PutCustomResource(resourceType string, data interface{}) error {
	return nil
}

// DeleteCustomResource delete custom resources, data is resource filter
func (m *MockStorage) DeleteCustomResource(resourceType string, data map[string]string) error {
	return nil
}

// CreateCustomResourceIndex create custom resources' index
func (m *MockStorage) CreateCustomResourceIndex(resourceType string, index drivers.Index) error {
	return nil
}

// DeleteCustomResourceIndex delete custom resources' index
func (m *MockStorage) DeleteCustomResourceIndex(resourceType string, indexName string) error {
	return nil
}

// QueryK8SNamespace query all namespace in specified cluster
func (m *MockStorage) QueryK8SNamespace(cluster string) ([]*storage.Namespace, error) {
	rawNs := []byte("{\"result\":true,\"code\":0,\"message\":\"success\",\"data\":[{\"clusterId\":\"BCS-K8S-15091\",\"resourceName\":\"bcs-system\",\"resourceType\":\"Namespace\",\"createTime\":\"2021-05-28T02:23:26.943Z\",\"data\":{\"spec\":{\"finalizers\":[\"kubernetes\"]},\"status\":{\"phase\":\"Active\"},\"metadata\":{\"resourceVersion\":\"110257089\",\"selfLink\":\"/api/v1/namespaces/bcs-system\",\"uid\":\"510fe9bc-d148-11ea-82a9-5254007d033c\",\"creationTimestamp\":\"2020-07-29T03:05:11Z\",\"labels\":{\"bcs-webhook\":\"false\"},\"name\":\"bcs-system\"}},\"updateTime\":\"2021-12-29T06:22:49.705Z\",\"_id\":\"60b0541ea7431f3e7083fd2d\"},{\"data\":{\"spec\":{\"finalizers\":[\"kubernetes\"]},\"status\":{\"phase\":\"Active\"},\"metadata\":{\"resourceVersion\":\"70064883\",\"selfLink\":\"/api/v1/namespaces/dfdfdfdfdf\",\"uid\":\"4fd8fd38-4a85-11eb-a214-5254007d033c\",\"creationTimestamp\":\"2020-12-30T09:56:39Z\",\"name\":\"dfdfdfdfdf\"}},\"updateTime\":\"2021-12-29T06:22:51.748Z\",\"_id\":\"60b0541fa7431f3e7083fded\",\"clusterId\":\"BCS-K8S-15091\",\"resourceName\":\"dfdfdfdfdf\",\"resourceType\":\"Namespace\",\"createTime\":\"2021-05-28T02:23:27.037Z\"}]}")
	basicRsp := &bcsapi.BasicResponse{}
	json.Unmarshal(rawNs, basicRsp)
	var nsRsp []*storage.Namespace
	json.Unmarshal(basicRsp.Data, &nsRsp)
	m.On("QueryK8SNamespace", "BCS-K8S-15091").Return(nsRsp, nil)
	args := m.Called(cluster)
	return args.Get(0).([]*storage.Namespace), args.Error(1)
}

// QueryK8SDeployment query all deployment in specified cluster
func (m *MockStorage) QueryK8SDeployment(cluster string) ([]*storage.Deployment, error) {
	args := m.Called(cluster)
	return args.Get(0).([]*storage.Deployment), args.Error(1)
}

// QueryK8SDaemonSet QueryK8SDaemonset query all daemonset in specified cluster
func (m *MockStorage) QueryK8SDaemonSet(cluster string) ([]*storage.DaemonSet, error) {
	args := m.Called(cluster)
	return args.Get(0).([]*storage.DaemonSet), args.Error(1)
}

// QueryK8SStatefulSet query all statefulset in specified cluster
func (m *MockStorage) QueryK8SStatefulSet(cluster string) ([]*storage.StatefulSet, error) {
	args := m.Called(cluster)
	return args.Get(0).([]*storage.StatefulSet), args.Error(1)
}

// QueryK8SGameDeployment query all gamedeployment in specified cluster
func (m *MockStorage) QueryK8SGameDeployment(cluster string) ([]*storage.GameDeployment, error) {
	args := m.Called(cluster)
	return args.Get(0).([]*storage.GameDeployment), args.Error(1)
}

// QueryK8SGameStatefulSet query all gamestatefulset in specified cluster
func (m *MockStorage) QueryK8SGameStatefulSet(cluster string) ([]*storage.GameStatefulSet, error) {
	args := m.Called(cluster)
	return args.Get(0).([]*storage.GameStatefulSet), args.Error(1)
}

// QueryK8SNode query all node in specified cluster
func (m *MockStorage) QueryK8SNode(cluster string) ([]*storage.K8sNode, error) {
	args := m.Called(cluster)
	return args.Get(0).([]*storage.K8sNode), args.Error(1)
}

// QueryMesosNamespace query all namespace in specified cluster
func (m *MockStorage) QueryMesosNamespace(cluster string) ([]*storage.Namespace, error) {
	rawMesosNs := []byte("{\"result\":true,\"code\":0,\"message\":\"success\",\"data\":[\"marstest\",\"bcs-system\"]}")
	basicRsp := &bcsapi.BasicResponse{}
	json.Unmarshal(rawMesosNs, basicRsp)
	var mesosNsRsp []*storage.Namespace
	json.Unmarshal(basicRsp.Data, &mesosNsRsp)
	m.On("QueryMesosNamespace", "BCS-MESOS-10039").Return(mesosNsRsp, nil)
	args := m.Called(cluster)
	return args.Get(0).([]*storage.Namespace), args.Error(1)
}

// QueryMesosDeployment query all deployment in specified cluster
func (m *MockStorage) QueryMesosDeployment(cluster string) ([]*storage.Deployment, error) {
	args := m.Called(cluster)
	return args.Get(0).([]*storage.Deployment), args.Error(1)
}

// QueryMesosApplication query all application in specified cluster
func (m *MockStorage) QueryMesosApplication(cluster string) ([]*storage.MesosApplication, error) {
	args := m.Called(cluster)
	return args.Get(0).([]*storage.MesosApplication), args.Error(1)
}
