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
func (m *MockStorage) QueryK8SPod(cluster, namespace string) ([]*storage.Pod, error) {
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
func (m *MockStorage) QueryK8SDeployment(cluster, namespace string) ([]*storage.Deployment, error) {
	rawDeploy := []byte("{\"result\":true,\"code\":0,\"message\":\"success\",\"data\":[{\"namespace\":\"bcs-system\",\"updateTime\":\"2021-12-29T06:22:57.711Z\",\"clusterId\":\"BCS-K8S-15091\",\"createTime\":\"2021-10-15T12:21:31.022Z\",\"_id\":\"60b0541ea7431f3e7083f8e5\",\"resourceName\":\"bcs-k8s-watch\",\"data\":{\"metadata\":{\"generation\":1,\"resourceVersion\":\"184848343\",\"namespace\":\"bcs-system\",\"selfLink\":\"/apis/apps/v1/namespaces/bcs-system/deployments/bcs-k8s-watch\",\"uid\":\"60e1e762-2db2-11ec-810c-5254007d033c\",\"creationTimestamp\":\"2021-10-15T12:21:09Z\",\"annotations\":{\"deployment.kubernetes.io/revision\":\"1\",\"meta.helm.sh/release-name\":\"15091-bcs-k8s\",\"meta.helm.sh/release-namespace\":\"bcs-system\"},\"labels\":{\"app.kubernetes.io/version\":\"1.21.1\",\"helm.sh/chart\":\"bcs-k8s-watch-1.21.1\",\"app.kubernetes.io/instance\":\"15091-bcs-k8s\",\"app.kubernetes.io/managed-by\":\"Helm\",\"app.kubernetes.io/name\":\"bcs-k8s-watch\",\"app.kubernetes.io/platform\":\"bk-bcs\"},\"name\":\"bcs-k8s-watch\"},\"spec\":{\"revisionHistoryLimit\":10,\"selector\":{\"matchLabels\":{\"app.kubernetes.io/instance\":\"15091-bcs-k8s\",\"app.kubernetes.io/name\":\"bcs-k8s-watch\",\"app.kubernetes.io/platform\":\"bk-bcs\"}},\"strategy\":{\"rollingUpdate\":{\"maxSurge\":\"25%\",\"maxUnavailable\":\"25%\"},\"type\":\"RollingUpdate\"},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app.kubernetes.io/platform\":\"bk-bcs\",\"app.kubernetes.io/instance\":\"15091-bcs-k8s\",\"app.kubernetes.io/name\":\"bcs-k8s-watch\"}},\"spec\":{\"serviceAccountName\":\"bcs-k8s-watch\",\"restartPolicy\":\"Always\",\"containers\":[{\"terminationMessagePolicy\":\"File\",\"volumeMounts\":[{\"mountPath\":\"/data/bcs/cert/bcs\",\"name\":\"bcs-certs\"},{\"mountPath\":\"/data/bcs/bcs-k8s-watch/filter.json\",\"name\":\"filter-config\",\"subPath\":\"filter.json\"}],\"args\":[\"-f\",\"/data/bcs/bcs-k8s-watch/bcs-k8s-watch.json\"],\"imagePullPolicy\":\"Always\",\"name\":\"bcs-k8s-watch\",\"resources\":{},\"terminationMessagePath\":\"/dev/termination-log\",\"command\":[\"/data/bcs/bcs-k8s-watch/container-start.sh\"],\"env\":[{\"name\":\"clusterId\",\"value\":\"BCS-K8S-15091\"},{\"value\":\":2181,:2181,:2181\",\"name\":\"bcsZkHost\"},{\"name\":\"caFile\",\"value\":\"/data/bcs/cert/bcs/bcs-ca.crt\"},{\"name\":\"clientCertFile\",\"value\":\"/data/bcs/cert/bcs/bcs-client.crt\"},{\"name\":\"clientKeyFile\",\"value\":\"/data/bcs/cert/bcs/bcs-client.key\"},{\"name\":\"clientKeyPassword\",\"value\":\"Q5PNRjEZ7ri9vFGo\"},{\"value\":\"false\",\"name\":\"kubeWatchExternal\"},{\"name\":\"kubeMaster\"},{\"name\":\"customStorage\",\"value\":\"https://:31024\"},{\"name\":\"customNetService\"},{\"name\":\"customNetServiceZK\"},{\"valueFrom\":{\"fieldRef\":{\"fieldPath\":\"status.podIP\",\"apiVersion\":\"v1\"}},\"name\":\"localIp\"},{\"name\":\"writerQueueLen\",\"value\":\"10240\"},{\"value\":\"30\",\"name\":\"podQueueNum\"},{\"name\":\"bcsK8sWatchPort\",\"value\":\"10240\"},{\"name\":\"bcsK8sWatchMetricPort\",\"value\":\"10241\"},{\"value\":\"/data/bcs/cert/bcs/bcs-server.crt\",\"name\":\"serverCertFile\"},{\"value\":\"/data/bcs/cert/bcs/bcs-server.key\",\"name\":\"serverKeyFile\"},{\"name\":\"bcsK8sWatchDebug\",\"value\":\"true\"},{\"name\":\"log_dir\",\"value\":\"/data/bcs/logs/bcs\"},{\"value\":\"./\",\"name\":\"pid_dir\"},{\"name\":\"alsotostderr\",\"value\":\"true\"},{\"name\":\"log_level\",\"value\":\"3\"},{\"name\":\"BCS_CONFIG_TYPE\",\"value\":\"render\"}],\"image\":\"mirrors.tencent.com/bcs/bcs-k8s-watch:v1.21.3-filter\"}],\"dnsPolicy\":\"ClusterFirst\",\"terminationGracePeriodSeconds\":3,\"securityContext\":{},\"serviceAccount\":\"bcs-k8s-watch\",\"schedulerName\":\"default-scheduler\",\"volumes\":[{\"name\":\"bcs-certs\",\"secret\":{\"defaultMode\":420,\"secretName\":\"bk-bcs-certs\"}},{\"name\":\"filter-config\",\"configMap\":{\"name\":\"filter-config-bcs-k8s-watch\",\"defaultMode\":420,\"items\":[{\"key\":\"filter.json\",\"path\":\"filter.json\"}]}}]}},\"progressDeadlineSeconds\":600,\"replicas\":1},\"status\":{\"conditions\":[{\"reason\":\"NewReplicaSetAvailable\",\"status\":\"True\",\"type\":\"Progressing\",\"lastTransitionTime\":\"2021-10-15T12:21:09Z\",\"lastUpdateTime\":\"2021-10-15T12:21:23Z\",\"message\":\"ReplicaSet \\\"bcs-k8s-watch-7fb97997cc\\\" has successfully progressed.\"},{\"status\":\"True\",\"type\":\"Available\",\"lastTransitionTime\":\"2021-12-29T06:22:57Z\",\"lastUpdateTime\":\"2021-12-29T06:22:57Z\",\"message\":\"Deployment has minimum availability.\",\"reason\":\"MinimumReplicasAvailable\"}],\"observedGeneration\":1,\"readyReplicas\":1,\"replicas\":1,\"updatedReplicas\":1,\"availableReplicas\":1}},\"resourceType\":\"Deployment\"}]}")
	basicRsp := &bcsapi.BasicResponse{}
	json.Unmarshal(rawDeploy, basicRsp)
	var deployRsp []*storage.Deployment
	json.Unmarshal(basicRsp.Data, &deployRsp)
	m.On("QueryK8SDeployment", "BCS-K8S-15091").Return(deployRsp, nil)
	args := m.Called(cluster, namespace)
	return args.Get(0).([]*storage.Deployment), args.Error(1)
}

// QueryK8SDaemonSet QueryK8SDaemonset query all daemonset in specified cluster
func (m *MockStorage) QueryK8SDaemonSet(cluster, namespace string) ([]*storage.DaemonSet, error) {
	args := m.Called(cluster, namespace)
	return args.Get(0).([]*storage.DaemonSet), args.Error(1)
}

// QueryK8SStatefulSet query all statefulset in specified cluster
func (m *MockStorage) QueryK8SStatefulSet(cluster, namespace string) ([]*storage.StatefulSet, error) {
	args := m.Called(cluster, namespace)
	return args.Get(0).([]*storage.StatefulSet), args.Error(1)
}

// QueryK8SGameDeployment query all gamedeployment in specified cluster
func (m *MockStorage) QueryK8SGameDeployment(cluster, namespace string) ([]*storage.GameDeployment, error) {
	args := m.Called(cluster)
	return args.Get(0).([]*storage.GameDeployment), args.Error(1)
}

// QueryK8SGameStatefulSet query all gamestatefulset in specified cluster
func (m *MockStorage) QueryK8SGameStatefulSet(cluster, namespace string) ([]*storage.GameStatefulSet, error) {
	args := m.Called(cluster)
	return args.Get(0).([]*storage.GameStatefulSet), args.Error(1)
}

// QueryK8SNode query all node in specified cluster
func (m *MockStorage) QueryK8SNode(cluster string) ([]*storage.K8sNode, error) {
	args := m.Called(cluster)
	return args.Get(0).([]*storage.K8sNode), args.Error(1)
}

// QueryMesosNamespace query all namespace in specified cluster
func (m *MockStorage) QueryMesosNamespace(cluster string) ([]*storage.MesosNamespace, error) {
	rawMesosNs := []byte("{\"result\":true,\"code\":0,\"message\":\"success\",\"data\":[\"marstest\",\"bcs-system\"]}")
	basicRsp := &bcsapi.BasicResponse{}
	json.Unmarshal(rawMesosNs, basicRsp)
	var mesosNsRsp []*storage.Namespace
	json.Unmarshal(basicRsp.Data, &mesosNsRsp)
	m.On("QueryMesosNamespace", "BCS-MESOS-10039").Return(mesosNsRsp, nil)
	args := m.Called(cluster)
	return args.Get(0).([]*storage.MesosNamespace), args.Error(1)
}

// QueryMesosDeployment query all deployment in specified cluster
func (m *MockStorage) QueryMesosDeployment(cluster string) ([]*storage.MesosDeployment, error) {
	args := m.Called(cluster)
	return args.Get(0).([]*storage.MesosDeployment), args.Error(1)
}

// QueryMesosApplication query all application in specified cluster
func (m *MockStorage) QueryMesosApplication(cluster string) ([]*storage.MesosApplication, error) {
	rawMesosApp := []byte("{\"result\":true,\"code\":0,\"message\":\"success\",\"data\":[{\"_id\":\"61373373a7431f3e700bfab5\",\"namespace\":\"marstest\",\"resourceType\":\"application\",\"clusterId\":\"BCS-MESOS-10039\",\"resourceName\":\"deployment-1-v1625744136\",\"updateTime\":\"2021-09-13T13:05:38.82Z\",\"createTime\":\"2021-09-07T09:40:03.655Z\",\"data\":{\"kind\":\"application\",\"pods\":[{\"name\":\"0.deployment-1-v1625744136.marstest.10039.1625744136627629912\"},{\"name\":\"1.deployment-1-v1625744136.marstest.10039.1625744162889508616\"}],\"lastUpdateTime\":\"2021-07-08T19:36:05+08:00\",\"instance\":2,\"runningInstance\":2,\"metadata\":{\"labels\":{\"io.tencent.bcs.app.appid\":\"100148\",\"io.tencent.bcs.cluster\":\"BCS-MESOS-10039\",\"io.tencent.paas.application.mars-resource-update\":\"mars-resource-update\",\"io.tencent.paas.source_type\":\"template\",\"io.tencent.bcs.clusterid\":\"BCS-MESOS-10039\",\"io.tencent.bkdata.baseall.dataid\":\"6566\",\"io.tencent.bcs.controller.name\":\"deployment-1\",\"io.tencent.paas.templateid\":\"1901\",\"io.tencent.bcs.monitor.level\":\"general\",\"io.tencent.bcs.custom.labels\":\"{}\",\"io.tencent.bcs.controller.type\":\"deployment\",\"io.tencent.paas.version\":\"v1-update-res\",\"io.tencent.bkdata.container.stdlog.dataid\":\"15004\",\"io.tencent.bcs.namespace\":\"marstest\",\"io.tencent.bcs.kind\":\"Mesos\",\"io.tencent.paas.instanceid\":\"3757\",\"io.tencent.paas.projectid\":\"ab2b254938e84f6b86b466cc22e730b1\"},\"annotations\":{\"io.tencent.paas.creator\":\"bellkeyang\",\"io.tencent.paas.versionid\":\"11501\",\"io.tencent.bkdata.container.stdlog.dataid\":\"15004\",\"io.tencent.bcs.cluster\":\"BCS-MESOS-10039\",\"io.tencent.paas.webCache\":\"{\\\"remarkListCache\\\": [{\\\"key\\\": \\\"\\\", \\\"value\\\": \\\"\\\"}], \\\"labelListCache\\\": [{\\\"key\\\": \\\"\\\", \\\"value\\\": \\\"\\\"}], \\\"logLabelListCache\\\": [{\\\"key\\\": \\\"\\\", \\\"value\\\": \\\"\\\"}], \\\"isMetric\\\": false, \\\"metricIdList\\\": [], \\\"volumeUsers\\\": {\\\"container-1\\\": {}}}\",\"io.tencent.paas.version\":\"v1-update-res\",\"io.tencent.bcs.clusterid\":\"BCS-MESOS-10039\",\"io.tencent.paas.createTime\":\"2021-07-08 19:35:36\",\"io.tencent.paas.templateid\":\"1901\",\"io.tencent.bcs.controller.name\":\"deployment-1\",\"io.tencent.bkdata.baseall.dataid\":\"6566\",\"io.tencent.paas.instanceid\":\"3757\",\"io.tencent.bcs.namespace\":\"marstest\",\"io.tencent.bcs.controller.type\":\"deployment\",\"io.tencent.bcs.app.appid\":\"100148\",\"io.tencent.paas.source_type\":\"template\",\"io.tencent.paas.updator\":\"bellkeyang\",\"io.tencent.bcs.kind\":\"Mesos\",\"io.tencent.paas.updateTime\":\"2021-07-08 19:35:36\",\"io.tencent.paas.projectid\":\"ab2b254938e84f6b86b466cc22e730b1\"},\"name\":\"deployment-1-v1625744136\",\"namespace\":\"marstest\",\"creationTimestamp\":\"0001-01-01T00:00:00Z\"},\"reportTime\":\"2021-09-13T21:05:38.745977407+08:00\",\"status\":\"Running\",\"buildedInstance\":2,\"createTime\":\"2021-07-08T19:35:36+08:00\",\"lastStatus\":\"RollingUpdate\"}},{\"data\":{\"reportTime\":\"2021-09-13T21:06:53.01073613+08:00\",\"instance\":2,\"pods\":[{\"name\":\"0.marstest.marstest.10039.1624377266583567421\"},{\"name\":\"1.marstest.marstest.10039.1625739411004273114\"}],\"kind\":\"application\",\"createTime\":\"2021-06-22T23:54:26+08:00\",\"metadata\":{\"namespace\":\"marstest\",\"creationTimestamp\":\"0001-01-01T00:00:00Z\",\"labels\":{\"io.tencent.paas.projectid\":\"ab2b254938e84f6b86b466cc22e730b1\",\"io.tencent.paas.application.python\":\"python\",\"io.tencent.paas.instanceid\":\"3726\",\"io.tencent.bcs.custom.labels\":\"{}\",\"io.tencent.bcs.namespace\":\"marstest\",\"io.tencent.bkdata.container.stdlog.dataid\":\"15004\",\"io.tencent.bcs.clusterid\":\"BCS-MESOS-10039\",\"io.tencent.bcs.controller.name\":\"marstest\",\"io.tencent.paas.version\":\"v1\",\"io.tencent.bcs.cluster\":\"BCS-MESOS-10039\",\"io.tencent.paas.source_type\":\"template\",\"io.tencent.bkdata.baseall.dataid\":\"6566\",\"io.tencent.bcs.controller.type\":\"deployment\",\"io.tencent.paas.templateid\":\"1897\",\"io.tencent.bcs.app.appid\":\"100148\",\"io.tencent.bcs.kind\":\"Mesos\",\"io.tencent.bcs.monitor.level\":\"general\"},\"annotations\":{\"io.tencent.paas.webCache\":\"{\\\"remarkListCache\\\": [{\\\"key\\\": \\\"\\\", \\\"value\\\": \\\"\\\"}], \\\"labelListCache\\\": [{\\\"key\\\": \\\"\\\", \\\"value\\\": \\\"\\\"}], \\\"logLabelListCache\\\": [{\\\"key\\\": \\\"\\\", \\\"value\\\": \\\"\\\"}], \\\"isMetric\\\": false, \\\"metricIdList\\\": [], \\\"volumeUsers\\\": {\\\"python\\\": {}}}\",\"io.tencent.bcs.app.appid\":\"100148\",\"io.tencent.bkdata.baseall.dataid\":\"6566\",\"io.tencent.paas.projectid\":\"ab2b254938e84f6b86b466cc22e730b1\",\"io.tencent.paas.updateTime\":\"2021-06-22 23:54:26\",\"io.tencent.bcs.namespace\":\"marstest\",\"io.tencent.bcs.kind\":\"Mesos\",\"io.tencent.bcs.clusterid\":\"BCS-MESOS-10039\",\"io.tencent.bcs.controller.type\":\"deployment\",\"io.tencent.paas.instanceid\":\"3726\",\"io.tencent.paas.creator\":\"marsjma\",\"io.tencent.paas.source_type\":\"template\",\"io.tencent.paas.updator\":\"marsjma\",\"io.tencent.paas.version\":\"v1\",\"io.tencent.paas.createTime\":\"2021-06-22 23:54:26\",\"io.tencent.paas.versionid\":\"11395\",\"io.tencent.paas.templateid\":\"1897\",\"io.tencent.bkdata.container.stdlog.dataid\":\"15004\",\"io.tencent.bcs.controller.name\":\"marstest\",\"io.tencent.bcs.cluster\":\"BCS-MESOS-10039\"},\"name\":\"marstest\"},\"lastUpdateTime\":\"2021-07-08T18:16:52+08:00\",\"status\":\"Running\",\"message\":\"application is running\",\"runningInstance\":2,\"buildedInstance\":2,\"lastStatus\":\"Deploying\"},\"_id\":\"60d207b2a7431f3e70b58825\",\"resourceType\":\"application\",\"createTime\":\"2021-06-22T15:54:26.585Z\",\"namespace\":\"marstest\",\"resourceName\":\"marstest\",\"updateTime\":\"2021-09-13T13:06:53.095Z\",\"clusterId\":\"BCS-MESOS-10039\"}]}")
	basicRsp := &bcsapi.BasicResponse{}
	json.Unmarshal(rawMesosApp, basicRsp)
	var mesosApp []*storage.MesosApplication
	json.Unmarshal(basicRsp.Data, &mesosApp)
	m.On("QueryMesosApplication", "BCS-MESOS-10039").Return(mesosApp, nil)
	args := m.Called(cluster)
	return args.Get(0).([]*storage.MesosApplication), args.Error(1)
}
