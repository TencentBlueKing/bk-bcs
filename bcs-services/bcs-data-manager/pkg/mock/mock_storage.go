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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/stretchr/testify/mock"
)

// MockStorage mock storage
type MockStorage struct {
	mock.Mock
}

// QueryMesosTaskgroup search all taskgroup by clusterID
func (m *MockStorage) QueryMesosTaskgroup(cluster string) ([]*storage.Taskgroup, error) {}

// QueryK8SPod query all pod information in specified cluster
func (m *MockStorage) QueryK8SPod(cluster string) ([]*storage.Pod, error) {}

// GetIPPoolDetailInfo get all underlay ip information
func (m *MockStorage) GetIPPoolDetailInfo(clusterID string) ([]*storage.IPPool, error) {}

// ListCustomResource list custom resources, Unmarshalled to dest.
// dest should be a pointer to a struct of map[string]interface{}
func (m *MockStorage) ListCustomResource(resourceType string, filter map[string]string, dest interface{}) error {
}

// PutCustomResource put custom resources, support map or struct
func (m *MockStorage) PutCustomResource(resourceType string, data interface{}) error {}

// DeleteCustomResource delete custom resources, data is resource filter
func (m *MockStorage) DeleteCustomResource(resourceType string, data map[string]string) error {}

// CreateCustomResourceIndex create custom resources' index
func (m *MockStorage) CreateCustomResourceIndex(resourceType string, index drivers.Index) error {}

// DeleteCustomResourceIndex delete custom resources' index
func (m *MockStorage) DeleteCustomResourceIndex(resourceType string, indexName string) error {}

// QueryK8SNamespace query all namespace in specified cluster
func (m *MockStorage) QueryK8SNamespace(cluster string) ([]*storage.Namespace, error) {}

// QueryK8SDeployment query all deployment in specified cluster
func (m *MockStorage) QueryK8SDeployment(cluster string) ([]*storage.Deployment, error) {}

// QueryK8SDaemonSet QueryK8SDaemonset query all daemonset in specified cluster
func (m *MockStorage) QueryK8SDaemonSet(cluster string) ([]*storage.DaemonSet, error) {}

// QueryK8SStatefulSet query all statefulset in specified cluster
func (m *MockStorage) QueryK8SStatefulSet(cluster string) ([]*storage.StatefulSet, error) {}

// QueryK8SGameDeployment query all gamedeployment in specified cluster
func (m *MockStorage) QueryK8SGameDeployment(cluster string) ([]*storage.GameDeployment, error) {}

// QueryK8SGameStatefulSet query all gamestatefulset in specified cluster
func (m *MockStorage) QueryK8SGameStatefulSet(cluster string) ([]*storage.GameStatefulSet, error) {}

// QueryK8SNode query all node in specified cluster
func (m *MockStorage) QueryK8SNode(cluster string) ([]*storage.K8sNode, error) {}

// QueryMesosNamespace query all namespace in specified cluster
func (m *MockStorage) QueryMesosNamespace(cluster string) ([]*storage.Namespace, error) {}

// QueryMesosDeployment query all deployment in specified cluster
func (m *MockStorage) QueryMesosDeployment(cluster string) ([]*storage.Deployment, error) {}

// QueryMesosApplication query all application in specified cluster
func (m *MockStorage) QueryMesosApplication(cluster string) ([]*storage.MesosApplication, error) {}
