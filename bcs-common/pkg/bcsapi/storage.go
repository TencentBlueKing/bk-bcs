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

package bcsapi

import (
	"encoding/json"
	"fmt"
	"strings"

	blog "k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage"
	restclient "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	registry "github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
)

const (
	storagePath             = "bcsstorage/v1"
	customResourcePath      = "bcsstorage/v1/dynamic/customresources/%s"
	customResourceIndexPath = "bcsstorage/v1/dynamic/customresources/%s/index/%s"

	storageRequestLimit = 200
)

// Storage interface definition for bcs-storage
type Storage interface {
	// QueryMesosTaskgroup query mesos task groups
	// search all taskgroup by clusterID
	QueryMesosTaskgroup(cluster string) ([]*storage.Taskgroup, error)
	// QueryK8SPod query k8s pods
	// query all pod information in specified cluster
	QueryK8SPod(cluster, namespace string, pods ...string) ([]*storage.Pod, error)
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
	QueryK8SNamespace(cluster string, namespaces ...string) ([]*storage.Namespace, error)
	// QueryK8SDeployment query all deployment in specified cluster
	QueryK8SDeployment(cluster, namespace string) ([]*storage.Deployment, error)
	// QueryK8SDaemonSet query all daemonset in specified cluster
	QueryK8SDaemonSet(cluster, namespace string) ([]*storage.DaemonSet, error)
	// QueryK8SStatefulSet query all statefulset in specified cluster
	QueryK8SStatefulSet(cluster, namespace string) ([]*storage.StatefulSet, error)
	// QueryK8SGameDeployment query all gamedeployment in specified cluster
	QueryK8SGameDeployment(cluster, namespace string) ([]*storage.GameDeployment, error)
	// QueryK8SGameStatefulSet query all gamestatefulset in specified cluster
	QueryK8SGameStatefulSet(cluster, namespace string) ([]*storage.GameStatefulSet, error)
	// QueryK8SNode query all node in specified cluster
	QueryK8SNode(cluster string) ([]*storage.K8sNode, error)
	// QueryK8sHPA query all hpa in specified cluster and namespace
	QueryK8sHPA(cluster, namespace string) ([]*storage.Hpa, error)
	// QueryK8sGPA query all gpa in specified cluster and namespace
	QueryK8sGPA(cluster, namespace string) ([]*storage.Gpa, error)
	// QueryK8sPvc query pvc in specified cluster and namespace
	QueryK8sPvc(cluster, namespace string) ([]*storage.Pvc, error)
	// QueryK8sStorageClass query all StorageClass in specified cluster
	QueryK8sStorageClass(cluster string) ([]*storage.StorageClass, error)
	// QueryK8sResourceQuota query ResourceQuota in specified cluster and namespace
	QueryK8sResourceQuota(cluster, namespace string) ([]*storage.ResourceQuota, error)
	// QueryK8sReplicaSet query ReplicaSet in specified cluster and namespace
	QueryK8sReplicaSet(cluster, namespace, name string) ([]*storage.ReplicaSet, error)
	// QueryMesosNamespace query all namespace in specified cluster
	QueryMesosNamespace(cluster string) ([]*storage.MesosNamespace, error)
	// QueryMesosDeployment query all deployment in specified cluster
	QueryMesosDeployment(cluster string) ([]*storage.MesosDeployment, error)
	// QueryMesosApplication query all application in specified cluster
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

// QueryK8sReplicaSet query k8s replicaset in specified cluster and namespace
func (c *StorageCli) QueryK8sReplicaSet(cluster, namespace, name string) ([]*storage.ReplicaSet, error) {
	subPath := "/k8s/dynamic/namespace_resources/clusters/%s/namespaces/" + namespace + "/ReplicaSet?" // nolint
	if name != "" {
		subPath += "resourceName=" + name + "&"
	}
	var replicaSets []*storage.ReplicaSet
	offset := 0
	for {
		var replicaSetTmp []*storage.ReplicaSet
		path := fmt.Sprintf("%soffset=%d&limit=%d", subPath, offset, storageRequestLimit)
		response, err := c.query(cluster, path)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(response.Data, &replicaSetTmp); err != nil {
			return nil, fmt.Errorf("replicaSets slice decode err: %s", err.Error())
		}
		replicaSets = append(replicaSets, replicaSetTmp...)
		if len(replicaSetTmp) < storageRequestLimit {
			break
		}
	}

	if len(replicaSets) == 0 {
		blog.V(5).Infof("no replicaSets found in cluster %s", cluster)
		return nil, nil
	}
	return replicaSets, nil
}

// QueryK8sPvc query pvc in specified cluster and namespace
func (c *StorageCli) QueryK8sPvc(cluster, namespace string) ([]*storage.Pvc, error) {
	subPath := "/k8s/dynamic/namespace_resources/clusters/%s/namespaces/" + namespace + "/PersistentVolumeClaim"
	var pvcs []*storage.Pvc
	offset := 0
	for {
		var pvcsTmp []*storage.Pvc
		path := fmt.Sprintf("%s?offset=%d&limit=%d", subPath, offset, storageRequestLimit)
		response, err := c.query(cluster, path)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(response.Data, &pvcsTmp); err != nil {
			return nil, fmt.Errorf("pvcs slice decode err: %s", err.Error())
		}
		pvcs = append(pvcs, pvcsTmp...)
		if len(pvcsTmp) == storageRequestLimit {
			offset += storageRequestLimit
			continue
		}
		break
	}

	if len(pvcs) == 0 {
		// No pvc data retrieve from bcs-storage
		blog.V(5).Infof("query kubernetes empty pvcs in cluster %s", cluster)
		return nil, nil
	}
	return pvcs, nil
}

// QueryK8sStorageClass query all StorageClass in specified cluster
func (c *StorageCli) QueryK8sStorageClass(cluster string) ([]*storage.StorageClass, error) {
	subPath := "/k8s/dynamic/cluster_resources/clusters/%s/StorageClass"
	var scs []*storage.StorageClass
	offset := 0
	for {
		var scsTmp []*storage.StorageClass
		path := fmt.Sprintf("%s?offset=%d&limit=%d", subPath, offset, storageRequestLimit)
		response, err := c.query(cluster, path)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(response.Data, &scsTmp); err != nil {
			return nil, fmt.Errorf("scs slice decode err: %s", err.Error())
		}
		scs = append(scs, scsTmp...)
		if len(scsTmp) == storageRequestLimit {
			offset += storageRequestLimit
			continue
		}
		break
	}

	if len(scs) == 0 {
		blog.V(5).Infof("query kubernetes empty scs in cluster %s", cluster)
		return nil, nil
	}
	return scs, nil
}

// QueryK8sResourceQuota query ResourceQuota in specified cluster and namespace
func (c *StorageCli) QueryK8sResourceQuota(cluster, namespace string) ([]*storage.ResourceQuota, error) {
	subPath := "/k8s/dynamic/namespace_resources/clusters/%s/namespaces/" + namespace + "/ResourceQuota"
	var quotas []*storage.ResourceQuota
	offset := 0
	for {
		var quotasTmp []*storage.ResourceQuota
		path := fmt.Sprintf("%s?offset=%d&limit=%d", subPath, offset, storageRequestLimit)
		response, err := c.query(cluster, path)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(response.Data, &quotasTmp); err != nil {
			return nil, fmt.Errorf("quotas slice decode err: %s", err.Error())
		}
		quotas = append(quotas, quotasTmp...)
		if len(quotasTmp) == storageRequestLimit {
			offset += storageRequestLimit
			continue
		}
		break
	}

	if len(quotas) == 0 {
		// No quota data retrieve from bcs-storage
		blog.V(5).Infof("query kubernetes empty quotas in cluster %s", cluster)
		return nil, nil
	}
	return quotas, nil
}

// QueryK8sHPA query all hpa in specified cluster and namespace
func (c *StorageCli) QueryK8sHPA(cluster, namespace string) ([]*storage.Hpa, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace is empty")
	}
	subPath := "/k8s/dynamic/namespace_resources/clusters/%s/namespaces/" + namespace + "/HorizontalPodAutoscaler"

	var hpas []*storage.Hpa
	offset := 0
	for {
		var hpasTmp []*storage.Hpa
		path := fmt.Sprintf("%s?offset=%d&limit=%d", subPath, offset, storageRequestLimit)
		response, err := c.query(cluster, path)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(response.Data, &hpasTmp); err != nil {
			return nil, fmt.Errorf("hpa slice decode err: %s", err.Error())
		}
		hpas = append(hpas, hpasTmp...)
		if len(hpasTmp) == storageRequestLimit {
			offset += storageRequestLimit
			continue
		}
		break
	}

	if len(hpas) == 0 {
		blog.V(5).Infof("query kubernetes empty hpas in cluster %s", cluster)
		return nil, nil
	}
	return hpas, nil
}

// QueryK8sGPA query all gpa in specified cluster and namespace
func (c *StorageCli) QueryK8sGPA(cluster, namespace string) ([]*storage.Gpa, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace is empty")
	}
	subPath := "/k8s/dynamic/namespace_resources/clusters/%s/namespaces/" + namespace + "/GeneralPodAutoscaler"

	var gpas []*storage.Gpa
	offset := 0
	for {
		var gpasTmp []*storage.Gpa
		path := fmt.Sprintf("%s?offset=%d&limit=%d", subPath, offset, storageRequestLimit)
		response, err := c.query(cluster, path)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(response.Data, &gpasTmp); err != nil {
			return nil, fmt.Errorf("gpa slice decode err: %s", err.Error())
		}
		gpas = append(gpas, gpasTmp...)
		if len(gpasTmp) == storageRequestLimit {
			offset += storageRequestLimit
			continue
		}
		break
	}

	if len(gpas) == 0 {
		blog.V(5).Infof("query kubernetes empty gpas in cluster %s", cluster)
		return nil, nil
	}
	return gpas, nil
}

// QueryK8SNode query all node in specified cluster
func (c *StorageCli) QueryK8SNode(cluster string) ([]*storage.K8sNode, error) {
	subPath := "/k8s/dynamic/cluster_resources/clusters/%s/Node"
	var k8sNodes []*storage.K8sNode
	offset := 0
	for {
		var k8sNodesTmp []*storage.K8sNode
		path := fmt.Sprintf("%s?offset=%d&limit=%d", subPath, offset, storageRequestLimit)
		response, err := c.query(cluster, path)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(response.Data, &k8sNodesTmp); err != nil {
			return nil, fmt.Errorf("k8sNodes slice decode err: %s", err.Error())
		}
		k8sNodes = append(k8sNodes, k8sNodesTmp...)
		if len(k8sNodesTmp) == storageRequestLimit {
			offset += storageRequestLimit
			continue
		}
		break
	}

	if len(k8sNodes) == 0 {
		blog.V(5).Infof("query kubernetes empty k8sNodes in cluster %s", cluster)
		return nil, nil
	}
	return k8sNodes, nil
}

// QueryMesosApplication query all application in specified cluster
func (c *StorageCli) QueryMesosApplication(cluster string) ([]*storage.MesosApplication, error) {
	subPath := "/query/mesos/dynamic/clusters/%s/application"

	var applications []*storage.MesosApplication
	offset := 0
	for {
		var applicationsTmp []*storage.MesosApplication
		path := fmt.Sprintf("%s?offset=%d&limit=%d", subPath, offset, storageRequestLimit)
		response, err := c.query(cluster, path)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(response.Data, &applicationsTmp); err != nil {
			return nil, fmt.Errorf("applications slice decode err: %s", err.Error())
		}
		applications = append(applications, applicationsTmp...)
		if len(applicationsTmp) == storageRequestLimit {
			offset += storageRequestLimit
			continue
		}
		break
	}

	if len(applications) == 0 {
		blog.V(5).Infof("query mesos empty application in cluster %s", cluster)
		return nil, nil
	}
	return applications, nil
}

// QueryMesosDeployment query all mesos deployment in specified cluster
func (c *StorageCli) QueryMesosDeployment(cluster string) ([]*storage.MesosDeployment, error) {
	subPath := "/query/mesos/dynamic/clusters/%s/deployment"

	var deployments []*storage.MesosDeployment
	offset := 0
	for {
		var deploymentsTmp []*storage.MesosDeployment
		path := fmt.Sprintf("%s?offset=%d&limit=%d", subPath, offset, storageRequestLimit)
		response, err := c.query(cluster, path)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(response.Data, &deploymentsTmp); err != nil {
			return nil, fmt.Errorf("deployments slice decode err: %s", err.Error())
		}
		deployments = append(deployments, deploymentsTmp...)
		if len(deploymentsTmp) == storageRequestLimit {
			offset += storageRequestLimit
			continue
		}
		break
	}

	if len(deployments) == 0 {
		blog.V(5).Infof("query mesos empty deployment in cluster %s", cluster)
		return nil, nil
	}
	return deployments, nil
}

// QueryMesosNamespace query all mesos namespace in specified cluster
func (c *StorageCli) QueryMesosNamespace(cluster string) ([]*storage.MesosNamespace, error) {
	subPath := "/query/mesos/dynamic/clusters/%s/namespace"
	response, err := c.query(cluster, subPath)
	if err != nil {
		return nil, err
	}
	var namespaces []*storage.MesosNamespace
	if err := json.Unmarshal(response.Data, &namespaces); err != nil {
		return nil, fmt.Errorf("namespaces slice decode err: %s", err.Error())
	}

	if len(namespaces) == 0 {
		blog.V(5).Infof("query mesos empty namespace in cluster %s", cluster)
		return nil, nil
	}
	return namespaces, nil
}

// QueryK8SGameStatefulSet query all gamestatefulset in specified cluster
func (c *StorageCli) QueryK8SGameStatefulSet(cluster, namespace string) ([]*storage.GameStatefulSet, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace is empty")
	}
	subPath := "/k8s/dynamic/namespace_resources/clusters/%s/namespaces/" + namespace + "/GameStatefulSet"

	var gamestatefulsets []*storage.GameStatefulSet
	offset := 0
	for {
		var gamestatefulsetsTmp []*storage.GameStatefulSet
		path := fmt.Sprintf("%s?offset=%d&limit=%d", subPath, offset, storageRequestLimit)
		response, err := c.query(cluster, path)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(response.Data, &gamestatefulsetsTmp); err != nil {
			return nil, fmt.Errorf("gamestatefulset slice decode err: %s", err.Error())
		}
		gamestatefulsets = append(gamestatefulsets, gamestatefulsetsTmp...)
		if len(gamestatefulsetsTmp) == storageRequestLimit {
			offset += storageRequestLimit
			continue
		}
		break
	}

	if len(gamestatefulsets) == 0 {
		blog.V(5).Infof("query kubernetes empty gamestatefulsets in cluster %s", cluster)
		return nil, nil
	}
	return gamestatefulsets, nil
}

// QueryK8SGameDeployment query all gamedeployment in specified cluster
func (c *StorageCli) QueryK8SGameDeployment(cluster, namespace string) ([]*storage.GameDeployment, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace is empty")
	}
	subPath := "/k8s/dynamic/namespace_resources/clusters/%s/namespaces/" + namespace + "/GameDeployment"

	var gamedeployments []*storage.GameDeployment
	offset := 0
	for {
		var gamedeploymentsTmp []*storage.GameDeployment
		path := fmt.Sprintf("%s?offset=%d&limit=%d", subPath, offset, storageRequestLimit)
		response, err := c.query(cluster, path)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(response.Data, &gamedeploymentsTmp); err != nil {
			return nil, fmt.Errorf("gamedeployments slice decode err: %s", err.Error())
		}
		gamedeployments = append(gamedeployments, gamedeploymentsTmp...)
		if len(gamedeploymentsTmp) == storageRequestLimit {
			offset += storageRequestLimit
			continue
		}
		break
	}

	if len(gamedeployments) == 0 {
		blog.V(5).Infof("query kubernetes empty gamedeployments in cluster %s", cluster)
		return nil, nil
	}
	return gamedeployments, nil
}

// QueryK8SStatefulSet query all statefulset in specified cluster
func (c *StorageCli) QueryK8SStatefulSet(cluster, namespace string) ([]*storage.StatefulSet, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace is empty")
	}
	subPath := "/k8s/dynamic/namespace_resources/clusters/%s/namespaces/" + namespace + "/StatefulSet"

	var statefulsets []*storage.StatefulSet
	offset := 0
	for {
		var statefulsetsTmp []*storage.StatefulSet
		path := fmt.Sprintf("%s?offset=%d&limit=%d", subPath, offset, storageRequestLimit)
		response, err := c.query(cluster, path)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(response.Data, &statefulsetsTmp); err != nil {
			return nil, fmt.Errorf("statefulsets slice decode err: %s", err.Error())
		}
		statefulsets = append(statefulsets, statefulsetsTmp...)
		if len(statefulsetsTmp) == storageRequestLimit {
			offset += storageRequestLimit
			continue
		}
		break
	}

	if len(statefulsets) == 0 {
		blog.V(5).Infof("query kubernetes empty statefulsets in cluster %s", cluster)
		return nil, nil
	}
	return statefulsets, nil
}

// QueryK8SDaemonSet query all daemonset in specified cluster
func (c *StorageCli) QueryK8SDaemonSet(cluster, namespace string) ([]*storage.DaemonSet, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace is empty")
	}
	subPath := "/k8s/dynamic/namespace_resources/clusters/%s/namespaces/" + namespace + "/DaemonSet"

	var daemonsets []*storage.DaemonSet
	offset := 0
	for {
		var daemonsetsTmp []*storage.DaemonSet
		path := fmt.Sprintf("%s?offset=%d&limit=%d", subPath, offset, storageRequestLimit)
		response, err := c.query(cluster, path)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(response.Data, &daemonsetsTmp); err != nil {
			return nil, fmt.Errorf("daemonsets slice decode err: %s", err.Error())
		}
		daemonsets = append(daemonsets, daemonsetsTmp...)
		if len(daemonsetsTmp) == storageRequestLimit {
			offset += storageRequestLimit
			continue
		}
		break
	}

	if len(daemonsets) == 0 {
		blog.V(5).Infof("query kubernetes empty daemonsets in cluster %s", cluster)
		return nil, nil
	}
	return daemonsets, nil
}

// QueryK8SDeployment query all deployment in specified cluster
func (c *StorageCli) QueryK8SDeployment(cluster, namespace string) ([]*storage.Deployment, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace is empty")
	}
	subPath := "/k8s/dynamic/namespace_resources/clusters/%s/namespaces/" + namespace + "/Deployment"

	var deployments []*storage.Deployment
	offset := 0
	for {
		var deploymentsTmp []*storage.Deployment
		path := fmt.Sprintf("%s?offset=%d&limit=%d", subPath, offset, storageRequestLimit)
		response, err := c.query(cluster, path)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(response.Data, &deploymentsTmp); err != nil {
			return nil, fmt.Errorf("deployments slice decode err: %s", err.Error())
		}
		deployments = append(deployments, deploymentsTmp...)
		if len(deploymentsTmp) == storageRequestLimit {
			offset += storageRequestLimit
			continue
		}
		break
	}

	if len(deployments) == 0 {
		blog.V(5).Infof("query kubernetes empty deployments in cluster %s", cluster)
		return nil, nil
	}
	return deployments, nil
}

// QueryK8SNamespace query all namespace in specified cluster
func (c *StorageCli) QueryK8SNamespace(cluster string, namespaces ...string) ([]*storage.Namespace, error) {
	subPath := "/k8s/dynamic/cluster_resources/clusters/%s/Namespace"

	var nsList []*storage.Namespace
	if len(namespaces) != 0 {
		for _, ns := range namespaces {
			var nsTmp *storage.Namespace
			path := fmt.Sprintf("%s/%s?limit=%d", subPath, ns, storageRequestLimit)
			response, err := c.query(cluster, path)
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal(response.Data, &nsTmp); err != nil {
				return nil, fmt.Errorf("namespaces slice decode err: %s", err.Error())
			}
			nsList = append(nsList, nsTmp)
		}
	} else {
		offset := 0
		for {
			var nsListTmp []*storage.Namespace
			path := fmt.Sprintf("%s?offset=%d&limit=%d", subPath, offset, storageRequestLimit)
			response, err := c.query(cluster, path)
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal(response.Data, &nsListTmp); err != nil {
				return nil, fmt.Errorf("namespaces slice decode err: %s", err.Error())
			}
			nsList = append(nsList, nsListTmp...)
			if len(nsListTmp) == storageRequestLimit {
				offset += storageRequestLimit
				continue
			}
			break
		}
	}

	if len(nsList) == 0 {
		blog.V(5).Infof("query kubernetes empty namespaces in cluster %s", cluster)
		return nil, nil
	}
	return nsList, nil
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
		// format bcs-api-gateway path
		return fmt.Sprintf("%s%s/", gatewayPrefix, types.BCS_MODULE_STORAGE)
	}
	return fmt.Sprintf("/%s/", storagePath)
}

// QueryMesosTaskgroup search all taskgroup by clusterID
func (c *StorageCli) QueryMesosTaskgroup(cluster string) ([]*storage.Taskgroup, error) {
	subPath := "/query/mesos/dynamic/clusters/%s/taskgroup"

	var taskgroups []*storage.Taskgroup
	offset := 0
	for {
		var taskgroupsTmp []*storage.Taskgroup
		path := fmt.Sprintf("%s?offset=%d&limit=%d", subPath, offset, storageRequestLimit)
		response, err := c.query(cluster, path)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(response.Data, &taskgroupsTmp); err != nil {
			return nil, fmt.Errorf("taskgroups slice decode err: %s", err.Error())
		}
		taskgroups = append(taskgroups, taskgroupsTmp...)
		if len(taskgroupsTmp) == storageRequestLimit {
			offset += storageRequestLimit
			continue
		}
		break
	}

	if len(taskgroups) == 0 {
		// No taskgroup data retrieve from bcs-storage
		blog.V(5).Infof("query mesos empty taskgroups in cluster %s", cluster)
		return nil, nil
	}
	return taskgroups, nil
}

// QueryK8SPod query all pod information in specified cluster
func (c *StorageCli) QueryK8SPod(cluster, namespace string, pods ...string) ([]*storage.Pod, error) {
	subPath := "/k8s/dynamic/namespace_resources/clusters/%s/namespaces/" + namespace + "/Pod"
	var podList []*storage.Pod

	if len(pods) != 0 {
		for _, pod := range pods {
			var podsTmp *storage.Pod
			path := fmt.Sprintf("%s/%s?limit=%d", subPath, pod, storageRequestLimit)
			response, err := c.query(cluster, path)
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal(response.Data, &podsTmp); err != nil {
				return nil, fmt.Errorf("pods slice decode err: %s", err.Error())
			}
			podList = append(podList, podsTmp)
		}
	} else {
		offset := 0
		for {
			var podsTmp []*storage.Pod
			path := fmt.Sprintf("%s?offset=%d&limit=%d", subPath, offset, storageRequestLimit)
			response, err := c.query(cluster, path)
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal(response.Data, &podsTmp); err != nil {
				return nil, fmt.Errorf("pods slice decode err: %s", err.Error())
			}
			podList = append(podList, podsTmp...)
			if len(podsTmp) == storageRequestLimit {
				offset += storageRequestLimit
				continue
			}
			break
		}
	}

	if len(podList) == 0 {
		// No pod data retrieve from bcs-storage
		blog.V(5).Infof("query kubernetes empty pods in cluster %s", cluster)
		return nil, nil
	}
	return podList, nil
}

// GetIPPoolDetailInfo get all underlay ip information
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
	// parse response.Data according to specified interface
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
		blog.Warningf("Non service found from etcd named %s", svcName)
	}
	endpoints := make([]string, 0)
	for _, node := range svc.Nodes {
		if ipv6Address := node.Metadata[types.IPV6]; ipv6Address != "" {
			endpoints = append(endpoints, ipv6Address)
		}
		endpoints = append(endpoints, node.Address)
	}
	c.Config.Hosts = endpoints
	blog.V(3).Infof("%d endpoints found for service %s in etcd registry: %+v", len(endpoints), svcName, endpoints)
}
