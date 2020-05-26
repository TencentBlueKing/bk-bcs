/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package storage

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ugorji/go/codec"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/http/httpclient"
	"bk-bcs/bcs-common/common/types"
	mdiscovery "bk-bcs/bcs-common/pkg/module-discovery"
	"bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/common"
)

// defaultJSONHandle default handle type for codec decoder
var defaultJSONHandle = codec.JsonHandle{MapKeyAsString: true}

const (
	BCS_STORAGE_DYNAMIC_ALL_RESOURCE_URI   = "/bcsstorage/v1/%s/dynamic/all_resources/clusters/%s/%s"
	BCS_STORAGE_DYNAMIC_WATCH_RESOURCE_URI = "/bcsstorage/v1/dynamic/watch/%s/%s"
)

// Interface interface for storage
type Interface interface {
	ListResources(clusterType, clusterID, resourceType string) (*common.ListStorageResourceResult, error)
	WatchClusterResources(clusterID, resourceType string) (chan *common.StorageEvent, error)
}

// StorageClient client for bcs storage
type StorageClient struct {
	moduleDiscovery mdiscovery.ModuleDiscovery
	httpCli         *httpclient.HttpClient
}

// NewStorageClient create storage client
func NewStorageClient(zkAddr string) (*StorageClient, error) {
	// create http client
	httpCli := httpclient.NewHttpClient()

	// create service module discovery to discover bcs-storage
	moduleDiscovery, err := mdiscovery.NewServiceDiscovery(zkAddr)
	if err != nil {
		return nil, err
	}

	return &StorageClient{
		moduleDiscovery: moduleDiscovery,
		httpCli:         httpCli,
	}, nil
}

// SetTLSConfig set tls config
func (sc *StorageClient) SetTLSConfig(conf *tls.Config) {
	sc.httpCli.SetTlsVerityConfig(conf)
}

// ListResources list resources
func (sc *StorageClient) ListResources(clusterType, clusterID, resourceType string) (*common.ListStorageResourceResult, error) {
	return sc.doQueryClusterResources(clusterType, clusterID, resourceType)
}

// WatchClusterResources watch cluster resources
func (sc *StorageClient) WatchClusterResources(clusterID, resourceType string) (chan *common.StorageEvent, error) {
	return sc.doWatchClusterResources(clusterID, resourceType)
}

func (sc *StorageClient) getStorageURL() (string, error) {
	storageSvrData, err := sc.moduleDiscovery.GetRandModuleServer(types.BCS_MODULE_STORAGE)
	if err != nil {
		blog.Errorf("get rand storage server failed, err %s", err.Error())
		return "", fmt.Errorf("get rand storage server failed, err %s", err.Error())
	}
	storageInfo, ok := storageSvrData.(*types.BcsStorageInfo)
	if !ok {
		blog.Errorf("get storage server info failed, invalid data %+v", storageSvrData)
		return "", fmt.Errorf("get storage server info failed, invalid data %+v", storageSvrData)
	}

	if storageInfo == nil {
		blog.Errorf("storageInfo empty")
		return "", fmt.Errorf("storageInfo empty")
	}

	scheme := storageInfo.Scheme
	if scheme != "http" && scheme != "https" {
		blog.Errorf("invalid server scheme %s", scheme)
		return "", fmt.Errorf("invalid server scheme %s", scheme)
	}

	url := scheme + "://" + storageInfo.IP + ":" + strconv.Itoa(int(storageInfo.Port))
	return url, nil
}

func (sc *StorageClient) doQueryClusterResources(clusterType, clusterID, resourceType string) (*common.ListStorageResourceResult, error) {
	url, err := sc.getStorageURL()
	if err != nil {
		return nil, err
	}
	realURL := url + fmt.Sprintf(BCS_STORAGE_DYNAMIC_ALL_RESOURCE_URI, clusterType, clusterID, resourceType)

	header := http.Header{
		"Content-Type": []string{"application/json"},
	}

	retData, err := sc.httpCli.GET(realURL, header, nil)
	if err != nil {
		blog.Errorf("call GET %s failed, err %s", realURL, err.Error())
		return nil, err
	}

	ret := &common.ListStorageResourceResult{}
	err = json.Unmarshal(retData, ret)
	if err != nil {
		blog.Errorf("Decode data %+v failed, err %s", retData, err.Error())
		return nil, err
	}
	return ret, nil
}

func (sc *StorageClient) doWatchClusterResources(clusterID, resourceType string) (chan *common.StorageEvent, error) {
	url, err := sc.getStorageURL()
	if err != nil {
		return nil, err
	}
	realURL := url + fmt.Sprintf(BCS_STORAGE_DYNAMIC_WATCH_RESOURCE_URI, clusterID, resourceType)

	header := http.Header{
		"Content-Type": []string{"application/json"},
	}

	body, err := sc.httpCli.RequestStream(realURL, "POST", header, []byte("{}"))
	if err != nil {
		blog.Errorf("request stream failed, err %s", err.Error())
		return nil, fmt.Errorf("request stream failed, err %s", err.Error())
	}

	decoder := codec.NewDecoder(body, &defaultJSONHandle)
	ch := make(chan *common.StorageEvent)
	event := new(common.StorageEvent)
	go func() {
		defer body.Close()
		for {
			if err := decoder.Decode(event); err != nil {
				blog.Errorf("decode watch response failed, err %s", err.Error())
				return
			}
			if event.Type == common.Brk {
				blog.Infof("watch %s/%s done", clusterID, resourceType)
				return
			}
			ch <- event
		}
	}()

	return ch, nil
}
