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

package bcsapi

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage"
	restclient "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
)

const (
	storagePath             = "bcsstorage/v1"
	customResourcePath      = "bcsstorage/v1/dynamic/customresources/%s"
	customResourceIndexPath = "bcsstorage/v1/dynamic/customresources/%s/index/%s"
)

// Storage interface definition for bcs-storage
type Storage interface {
	// search all taskgroup by clusterID
	QueryMesosTaskgroup(cluster string) ([]*storage.Taskgroup, error)
	// query all pod information in specified cluster
	QueryK8SPod(cluster string) ([]*storage.Pod, error)
	// GetIPPoolDetailInfo get all underlay ip information
	GetIPPoolDetailInfo(clusterID string) ([]*storage.IPPool, error)
	// ListCustomResource list custom resources, Unmarshalled to dest.
	// dest should be a pointer to a struct of map[string]interface{}
	ListCustomResource(resourceType string, filter map[string]string, dest interface{}) error
	// PutCustomResource put custom resources, support map or struct
	PutCustomResource(resourceType string, data interface{}) error
	// DeleteCustomResource delete custom resources, data is resource filter
	DeleteCustomResource(resourceType string, data map[string]string) error
	// CreateCustomResourceIndex create custom resources' index
	CreateCustomResourceIndex(resourceType string, index drivers.Index) error
	// DeleteCustomResourceIndex delete custom resources' index
	DeleteCustomResourceIndex(resourceType string, indexName string) error
	// QueryK8SNamespace query all namespace in specified cluster
	QueryK8SNamespace(cluster string) ([]*storage.Namespace, error)
	// QueryK8SDeployment query all deployment in specified cluster
	QueryK8SDeployment(cluster string) ([]*storage.Deployment, error)
	// QueryK8SDaemonset query all daemonset in specified cluster
	QueryK8SDaemonSet(cluster string) ([]*storage.DaemonSet, error)
	// QueryK8SStatefulSet query all statefulset in specified cluster
	QueryK8SStatefulSet(cluster string) ([]*storage.StatefulSet, error)
	// QueryK8SGameDeployment query all gamedeployment in specified cluster
	QueryK8SGameDeployment(cluster string) ([]*storage.GameDeployment, error)
	// QueryK8SGameStatefulSet query all gamestatefulset in specified cluster
	QueryK8SGameStatefulSet(cluster string) ([]*storage.GameStatefulSet, error)
	// QueryK8SNode query all node in specified cluster
	QueryK8SNode(cluster string) ([]*storage.K8sNode, error)
	// QueryMesosNamespace query all namespace in specified cluster
	QueryMesosNamespace(cluster string) ([]*storage.Namespace, error)
	//QueryMesosDeployment query all deployment in specified cluster
	QueryMesosDeployment(cluster string) ([]*storage.Deployment, error)
	//QueryMesosApplication query all application in specified cluster
	QueryMesosApplication(cluster string) ([]*storage.MesosApplication, error)
}

// NewStorage create bcs-storage api implementation
func NewStorage(config *Config) Storage {
	c := &StorageCli{
		Config: config,
	}
	if config.TLSConfig != nil {
		c.Client = restclient.NewRESTClientWithTLS(config.TLSConfig)
	} else {
		c.Client = restclient.NewRESTClient()
	}
	if c.Config.Etcd.Feature {
		err := c.watchEndpoints()
		if err != nil {
			blog.Errorf("watch etcd of service storage failed: %s", err.Error())
			return nil
		}
	}
	return c
}

// StorageCli bcsf-storage client implementation
type StorageCli struct {
	Config   *Config
	Client   *restclient.RESTClient
	discover registry.Registry
}

func (c *StorageCli) QueryK8SNode(cluster string) ([]*storage.K8sNode, error) {
	subPath := "/query/k8s/dynamic/clusters/%s/node"
	response, err := c.query(cluster, subPath)
	if err != nil {
		return nil, err
	}
	var k8sNodes []*storage.K8sNode
	if err := json.Unmarshal(response.Data, &k8sNodes); err != nil {
		return nil, fmt.Errorf("k8sNode slice decode err: %s", err.Error())
	}
	if len(k8sNodes) == 0 {
		blog.V(5).Infof("query kubernetes empty k8sNodes in cluster %s", cluster)
		return nil, nil
	}
	return k8sNodes, nil
}

func (c *StorageCli) QueryMesosApplication(cluster string) ([]*storage.MesosApplication, error) {
	subPath := "/query/mesos/dynamic/clusters/%s/application"
	response, err := c.query(cluster, subPath)
	if err != nil {
		return nil, err
	}
	var applications []*storage.MesosApplication
	if err := json.Unmarshal(response.Data, &applications); err != nil {
		return nil, fmt.Errorf("application slice decode err: %s", err.Error())
	}
	if len(applications) == 0 {
		blog.V(5).Infof("query mesos empty application in cluster %s", cluster)
		return nil, nil
	}
	return applications, nil
}

func (c *StorageCli) QueryMesosDeployment(cluster string) ([]*storage.Deployment, error) {
	subPath := "/query/mesos/dynamic/clusters/%s/deployment"
	response, err := c.query(cluster, subPath)
	if err != nil {
		return nil, err
	}
	var deployments []*storage.Deployment
	if err := json.Unmarshal(response.Data, &deployments); err != nil {
		return nil, fmt.Errorf("deployment slice decode err: %s", err.Error())
	}
	if len(deployments) == 0 {
		blog.V(5).Infof("query mesos empty deployment in cluster %s", cluster)
		return nil, nil
	}
	return deployments, nil
}

func (c *StorageCli) QueryMesosNamespace(cluster string) ([]*storage.Namespace, error) {
	subPath := "/query/mesos/dynamic/clusters/%s/namespace"
	response, err := c.query(cluster, subPath)
	if err != nil {
		return nil, err
	}
	var namespaces []*storage.Namespace
	if err := json.Unmarshal(response.Data, &namespaces); err != nil {
		return nil, fmt.Errorf("namespace slice decode err: %s", err.Error())
	}
	if len(namespaces) == 0 {
		blog.V(5).Infof("query mesos empty namespace in cluster %s", cluster)
		return nil, nil
	}
	return namespaces, nil
}

func (c *StorageCli) QueryK8SGameStatefulSet(cluster string) ([]*storage.GameStatefulSet, error) {
	subPath := "/k8s/dynamic/all_resources/clusters/%s/GameStatefulSet"
	response, err := c.query(cluster, subPath)
	if err != nil {
		return nil, err
	}
	var gamestatefulsets []*storage.GameStatefulSet
	if err := json.Unmarshal(response.Data, &gamestatefulsets); err != nil {
		return nil, fmt.Errorf("gamestatefulset slice decode err: %s", err.Error())
	}
	if len(gamestatefulsets) == 0 {
		blog.V(5).Infof("query kubernetes empty gamestatefulsets in cluster %s", cluster)
		return nil, nil
	}
	return gamestatefulsets, nil
}

func (c *StorageCli) QueryK8SGameDeployment(cluster string) ([]*storage.GameDeployment, error) {
	subPath := "/k8s/dynamic/all_resources/clusters/%s/GameDeployment"
	response, err := c.query(cluster, subPath)
	if err != nil {
		return nil, err
	}
	var gamedeployments []*storage.GameDeployment
	if err := json.Unmarshal(response.Data, &gamedeployments); err != nil {
		return nil, fmt.Errorf("gamedeployment slice decode err: %s", err.Error())
	}
	if len(gamedeployments) == 0 {
		blog.V(5).Infof("query kubernetes empty gamedeployments in cluster %s", cluster)
		return nil, nil
	}
	return gamedeployments, nil
}

func (c *StorageCli) QueryK8SStatefulSet(cluster string) ([]*storage.StatefulSet, error) {
	subPath := "/query/k8s/dynamic/clusters/%s/statefulset"
	response, err := c.query(cluster, subPath)
	if err != nil {
		return nil, err
	}
	var statefulsets []*storage.StatefulSet
	if err := json.Unmarshal(response.Data, &statefulsets); err != nil {
		return nil, fmt.Errorf("statefulset slice decode err: %s", err.Error())
	}
	if len(statefulsets) == 0 {
		blog.V(5).Infof("query kubernetes empty statefulsets in cluster %s", cluster)
		return nil, nil
	}
	return statefulsets, nil
}

func (c *StorageCli) QueryK8SDaemonSet(cluster string) ([]*storage.DaemonSet, error) {
	subPath := "/query/k8s/dynamic/clusters/%s/daemonset"
	response, err := c.query(cluster, subPath)
	if err != nil {
		return nil, err
	}
	var daemonsets []*storage.DaemonSet
	if err := json.Unmarshal(response.Data, &daemonsets); err != nil {
		return nil, fmt.Errorf("daemonset slice decode err: %s", err.Error())
	}
	if len(daemonsets) == 0 {
		blog.V(5).Infof("query kubernetes empty daemonsets in cluster %s", cluster)
		return nil, nil
	}
	return daemonsets, nil
}

func (c *StorageCli) QueryK8SDeployment(cluster string) ([]*storage.Deployment, error) {
	subPath := "/query/k8s/dynamic/clusters/%s/deployment"
	response, err := c.query(cluster, subPath)
	if err != nil {
		return nil, err
	}
	var deployments []*storage.Deployment
	if err := json.Unmarshal(response.Data, &deployments); err != nil {
		return nil, fmt.Errorf("deployment slice decode err: %s", err.Error())
	}
	if len(deployments) == 0 {
		blog.V(5).Infof("query kubernetes empty deployments in cluster %s", cluster)
		return nil, nil
	}
	return deployments, nil
}

func (c *StorageCli) QueryK8SNamespace(cluster string) ([]*storage.Namespace, error) {
	subPath := "/query/k8s/dynamic/clusters/%s/namespace"
	response, err := c.query(cluster, subPath)
	if err != nil {
		return nil, err
	}
	var namespaces []*storage.Namespace
	if err := json.Unmarshal(response.Data, &namespaces); err != nil {
		return nil, fmt.Errorf("namespace slice decode err: %s", err.Error())
	}
	if len(namespaces) == 0 {
		blog.V(5).Infof("query kubernetes empty namespaces in cluster %s", cluster)
		return nil, nil
	}
	return namespaces, nil
}

func (c *StorageCli) query(cluster, subPath string) (*BasicResponse, error) {
	if len(cluster) == 0 {
		return nil, fmt.Errorf("lost cluster id")
	}
	var response BasicResponse
	err := bkbcsSetting(c.Client.Get(), c.Config).
		WithEndpoints(c.Config.Hosts).
		WithBasePath(c.getRequestPath()).
		SubPathf(subPath, cluster).
		Do().
		Into(&response)
	if err != nil {
		return nil, err
	}
	if !response.Result {
		return nil, fmt.Errorf(response.Message)
	}
	return &response, nil
}

// getRequestPath get storage query URL prefix
func (c *StorageCli) getRequestPath() string {
	if c.Config.Gateway {
		//format bcs-api-gateway path
		return fmt.Sprintf("%s%s/", gatewayPrefix, types.BCS_MODULE_STORAGE)
	}
	return fmt.Sprintf("/%s/", storagePath)
}

// QueryMesosTaskgroup search all taskgroup by clusterID
func (c *StorageCli) QueryMesosTaskgroup(cluster string) ([]*storage.Taskgroup, error) {
	var response BasicResponse
	err := bkbcsSetting(c.Client.Get(), c.Config).
		WithEndpoints(c.Config.Hosts).
		WithBasePath(c.getRequestPath()).
		SubPathf("/query/mesos/dynamic/clusters/%s/taskgroup", cluster).
		Do().
		Into(&response)
	if err != nil {
		return nil, err
	}
	if !response.Result {
		return nil, fmt.Errorf(response.Message)
	}
	var taskgroups []*storage.Taskgroup
	if err := json.Unmarshal(response.Data, &taskgroups); err != nil {
		return nil, fmt.Errorf("taskgroup slice decode err: %s", err.Error())
	}
	if len(taskgroups) == 0 {
		//No taskgroup data retrieve from bcs-storage
		blog.V(5).Infof("query mesos empty taskgroups in cluster %s", cluster)
		return nil, nil
	}
	return taskgroups, nil
}

// QueryK8SPod query all pod information in specified cluster
func (c *StorageCli) QueryK8SPod(cluster string) ([]*storage.Pod, error) {
	subPath := "/query/k8s/dynamic/clusters/%s/pod"
	response, err := c.query(cluster, subPath)
	if err != nil {
		return nil, err
	}
	//decode destination object
	var pods []*storage.Pod
	if err := json.Unmarshal(response.Data, &pods); err != nil {
		return nil, fmt.Errorf("pod slice decode err: %s", err.Error())
	}
	if len(pods) == 0 {
		//No taskgroup data retrieve from bcs-storage
		blog.V(5).Infof("query kubernetes empty pods in cluster %s", cluster)
		return nil, nil
	}
	return pods, nil
}

//GetIPPoolDetailInfo get all underlay ip information
func (c *StorageCli) GetIPPoolDetailInfo(clusterID string) ([]*storage.IPPool, error) {
	if len(clusterID) == 0 {
		return nil, fmt.Errorf("lost cluster Id")
	}
	var response BasicResponse
	err := bkbcsSetting(c.Client.Get(), c.Config).
		WithEndpoints(c.Config.Hosts).
		WithBasePath(c.getRequestPath()).
		SubPathf("/query/mesos/dynamic/clusters/%s/ippoolstaticdetail", clusterID).
		Do().
		Into(&response)
	if err != nil {
		return nil, err
	}
	if !response.Result {
		return nil, fmt.Errorf(response.Message)
	}
	//parse response.Data according to specified interface
	detailResponse := make([]*storage.IPPoolDetailResponse, 0)
	if err := json.Unmarshal(response.Data, &detailResponse); err != nil {
		return nil, fmt.Errorf("decode response data failed, %s", err.Error())
	}
	if len(detailResponse) == 0 {
		return nil, fmt.Errorf("empty response from storage even http request success")
	}
	if len(detailResponse[0].Datas) == 0 {
		return nil, nil
	}
	return detailResponse[0].Datas, nil
}

// ListCustomResource list custom resources, dest should be corresponding resource type or map[string]interface{}
func (c *StorageCli) ListCustomResource(resourceType string, filter map[string]string, dest interface{}) error {
	err := bkbcsSetting(c.Client.Get(), c.Config).
		WithEndpoints(c.Config.Hosts).
		WithBasePath("/").
		SubPathf(customResourcePath, resourceType).
		WithParams(filter).
		Do().
		Into(dest)
	if err != nil {
		return err
	}
	return nil
}

// PutCustomResource put cluster resource
func (c *StorageCli) PutCustomResource(resourceType string, data interface{}) error {
	resp := bkbcsSetting(c.Client.Put(), c.Config).
		WithEndpoints(c.Config.Hosts).
		WithBasePath("/").
		SubPathf(customResourcePath, resourceType).
		WithJSON(data).
		Do()
	if resp.Err != nil {
		return resp.Err
	}
	return nil
}

// DeleteCustomResource delete custom resource
func (c *StorageCli) DeleteCustomResource(resourceType string, data map[string]string) error {
	resp := bkbcsSetting(c.Client.Delete(), c.Config).
		WithEndpoints(c.Config.Hosts).
		WithBasePath("/").
		SubPathf(customResourcePath, resourceType).
		WithParams(data).
		Do()
	if resp.Err != nil {
		return resp.Err
	}
	return nil
}

// CreateCustomResourceIndex create custom resource index
func (c *StorageCli) CreateCustomResourceIndex(resourceType string, index drivers.Index) error {
	resp := bkbcsSetting(c.Client.Put(), c.Config).
		WithEndpoints(c.Config.Hosts).
		WithBasePath("/").
		SubPathf(customResourceIndexPath, resourceType, index.Name).
		WithJSON(index.Key).
		Do()
	if resp.Err != nil {
		return resp.Err
	}
	return nil
}

// DeleteCustomResourceIndex delete custom resource index
func (c *StorageCli) DeleteCustomResourceIndex(resourceType string, indexName string) error {
	resp := bkbcsSetting(c.Client.Delete(), c.Config).
		WithEndpoints(c.Config.Hosts).
		WithBasePath("/").
		SubPathf(customResourceIndexPath, resourceType, indexName).
		Do()
	if resp.Err != nil {
		return resp.Err
	}
	return nil
}

func (c *StorageCli) watchEndpoints() error {
	tlsconf, err := c.Config.Etcd.GetTLSConfig()
	if err != nil {
		blog.Errorf("Get tlsconfig for etcd failed: %s, ca: %s, cert: %s, key:%s",
			err.Error(), c.Config.Etcd.CA, c.Config.Etcd.Cert, c.Config.Etcd.Key)
		return err
	}
	options := &registry.Options{
		RegistryAddr: strings.Split(c.Config.Etcd.Address, ","),
		Name:         types.BCS_MODULE_STORAGE + "bkbcs.tencent.com",
		Version:      version.BcsVersion,
		Config:       tlsconf,
		EvtHandler:   c.handlerEtcdEvent,
	}
	c.discover = registry.NewEtcdRegistry(options)
	if c.discover == nil {
		blog.Errorf("NewEtcdRegistry for service (%s) discovery failed", types.BCS_MODULE_STORAGE)
		return fmt.Errorf("NewEtcdRegistry for service (%s) discovery failed", types.BCS_MODULE_STORAGE)
	}
	c.handlerEtcdEvent(options.Name)
	return nil
}

func (c *StorageCli) handlerEtcdEvent(svcName string) {
	svc, err := c.discover.Get(svcName)
	if err != nil {
		blog.Errorf("Get svc %s from etcd registry failed: %s", svcName, err.Error())
		return
	}
	if len(svc.Nodes) == 0 {
		blog.Warnf("Non service found from etcd named %s", svcName)
	}
	endpoints := make([]string, 0)
	for _, node := range svc.Nodes {
		endpoints = append(endpoints, node.Address)
	}
	c.Config.Hosts = endpoints
	blog.V(3).Infof("%d endpoints found for service %s in etcd registry: %+v", len(endpoints), svcName, endpoints)
}
