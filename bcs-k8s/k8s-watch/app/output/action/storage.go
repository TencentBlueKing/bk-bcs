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

package action

// Use http protocol to sync data to BCSStorage Service
import (
	glog "bk-bcs/bcs-common/common/blog"

	"bk-bcs/bcs-k8s/k8s-watch/app/bcs"
	"bk-bcs/bcs-k8s/k8s-watch/app/output/http"
)

type StorageAction struct {
	Name           string
	ClusterID      string
	StorageService *bcs.StorageService
}

func (storageAction *StorageAction) Add(syncData *SyncData) {
	storageAction.request("PUT", syncData)
}

func (storageAction *StorageAction) Delete(syncData *SyncData) {
	storageAction.request("DELETE", syncData)
}

func (storageAction *StorageAction) Update(syncData *SyncData) {
	storageAction.request("PUT", syncData)
}

func (storageAction *StorageAction) request(method string, syncData *SyncData) {

	//glog.Infof("current servers: %s", storageAction.StorageService.Servers)
	//glog.Infof("calling request: %s %s %s/%s", method, syncData.Kind, syncData.Namespace, syncData.Name)
	if len(storageAction.StorageService.Servers) == 0 {
		// the process get address from zk not finished yet or there is no storage server on zk
		//glog.Errorf("storage server list is empty! got no address yet")
		return
	}

	var client http.StorageClient
	var resp http.StorageResponse
	var err error
	for _, httpClientConfig := range storageAction.StorageService.Servers {

		client = http.StorageClient{
			HTTPClientConfig: httpClientConfig,
			ClusterID:        storageAction.ClusterID,
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
			glog.Errorf("%s %s FAIL %s retry: [%s/%s]", method, syncData.Kind, err.Error(),syncData.Namespace, syncData.Name)
			continue
		}
		break
	}
	if !resp.Result {
		glog.Errorf("%s %s ERROR: [%s/%s]", method, syncData.Kind, syncData.Namespace, syncData.Name)
		return
	}

	glog.V(2).Infof("%s %s SUCCESS: [%s/%s]", method, syncData.Kind, syncData.Namespace, syncData.Name)
	return

}
