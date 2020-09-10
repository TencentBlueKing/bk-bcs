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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage"
	restclient "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
)

// Storage interface definition for bcs-storage
type Storage interface {
	// search all taskgroup by clusterID
	QueryMesosTaskgroup(cluster string) ([]*storage.Taskgroup, error)
	// query all pod information in specified cluster
	QueryK8SPod(cluster string) ([]*storage.Pod, error)
	// GetIPPoolDetailInfo get all underlay ip information
	GetIPPoolDetailInfo(clusterID string) ([]*storage.IPPool, error)
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
	return c
}

// StorageCli bcsf-storage client implementation
type StorageCli struct {
	Config *Config
	Client *restclient.RESTClient
}

// getRequestPath get storage query URL prefix
func (c *StorageCli) getRequestPath() string {
	if c.Config.Gateway {
		//format bcs-api-gateway path
		return fmt.Sprintf("%s%s/", gatewayPrefix, types.BCS_MODULE_STORAGE)
	}
	return fmt.Sprintf("/%s/", types.BCS_MODULE_STORAGE)
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
	if len(cluster) == 0 {
		return nil, fmt.Errorf("lost cluster id")
	}
	var response BasicResponse
	err := bkbcsSetting(c.Client.Get(), c.Config).
		WithEndpoints(c.Config.Hosts).
		WithBasePath(c.getRequestPath()).
		SubPathf("/query/k8s/dynamic/clusters/%s/pod", cluster).
		Do().
		Into(&response)
	if err != nil {
		return nil, err
	}
	if !response.Result {
		return nil, fmt.Errorf(response.Message)
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
