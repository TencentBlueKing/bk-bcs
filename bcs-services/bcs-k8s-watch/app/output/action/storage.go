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

package action

import (
	"fmt"

	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/output/http"
)

// StorageAction is http action of storage service.
type StorageAction struct {
	clusterID      string
	name           string
	storageService *bcs.InnerService
}

// NewStorageAction creates a new StorageAction instance.
func NewStorageAction(clusterID, name string, storageService *bcs.InnerService) *StorageAction {
	return &StorageAction{
		clusterID:      clusterID,
		name:           name,
		storageService: storageService,
	}
}

// Add adds new resource data by http PUT.
func (act *StorageAction) Add(syncData *SyncData) error {
	return act.request("PUT", syncData)
}

// Delete deletes target resource data by http DELETE.
func (act *StorageAction) Delete(syncData *SyncData) error {
	return act.request("DELETE", syncData)
}

// Update updates old resource data by http PUT.
func (act *StorageAction) Update(syncData *SyncData) error {
	return act.request("PUT", syncData)
}

func (act *StorageAction) request(method string, syncData *SyncData) error {
	glog.Infof("calling request: %s %s %s/%s", method, syncData.Kind, syncData.Namespace, syncData.Name)

	targets := act.storageService.Servers()

	if len(targets) == 0 {
		glog.Errorf("storage server list is empty, got no address yet")
		return fmt.Errorf("storage server list is empty, got no address yet")
	}

	var client http.StorageClient
	var resp http.StorageResponse
	var err error

	for _, httpClientConfig := range targets {
		client = http.StorageClient{
			HTTPClientConfig: httpClientConfig,
			ClusterID:        act.clusterID,
			Namespace:        syncData.Namespace,
			ResourceType:     syncData.Kind,
			ResourceName:     syncData.Name,
		}

		switch method {
		case "GET":
			resp, err = client.GET()
		case "PUT":
			resp, err = client.PUT(syncData.Data)
		case "DELETE":
			resp, err = client.DELETE()
		}

		if err != nil {
			glog.Errorf("%s %s FAIL %s retry: [%s/%s]", method, syncData.Kind, err.Error(), syncData.Namespace, syncData.Name)
			continue
		}
		break
	}

	if !resp.Result {
		glog.Errorf("%s %s ERROR[%s]: [%s/%s]", method, syncData.Kind, resp.Message, syncData.Namespace, syncData.Name)
		return fmt.Errorf("%s %s ERROR[%s]: [%s/%s]", method, syncData.Kind, resp.Message, syncData.Namespace, syncData.Name)
	}

	glog.V(2).Infof("%s %s SUCCESS: [%s/%s]", method, syncData.Kind, syncData.Namespace, syncData.Name)
	return nil
}
