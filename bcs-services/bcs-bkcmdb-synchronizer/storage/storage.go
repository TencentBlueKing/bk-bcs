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
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ugorji/go/codec"

	discovery "github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/common"
)

// defaultJSONHandle default handle type for codec decoder
var defaultJSONHandle = codec.JsonHandle{MapKeyAsString: true}

const (
	// BcsStorageDynamicAllResourceURI uri for query dynamic resource by bcs storage
	BcsStorageDynamicAllResourceURI = "/bcsstorage/v1/%s/dynamic/all_resources/clusters/%s/%s"
	// BcsStorageDynamicWatchResourceURI uri for watch dynamic resource by bcs storage
	BcsStorageDynamicWatchResourceURI = "/bcsstorage/v1/dynamic/watch/%s/%s"
)

// Interface interface for storage
type Interface interface {
	ListResources(clusterType, clusterID, resourceType string) (*common.ListStorageResourceResult, error)
	WatchClusterResources(clusterID, resourceType string) (chan *common.StorageEvent, error)
}

// StorageClient client for bcs storage
type StorageClient struct {
	disc        *discovery.RegDiscover
	httpCli     *httpclient.HttpClient
	servers     []*types.ServerInfo
	tlsConfig   *tls.Config
	serversLock sync.Mutex
	chanStorage <-chan *discovery.DiscoverEvent
}

// NewStorageClient create storage client
func NewStorageClient(zkAddr string) (*StorageClient, error) {
	// create http client
	httpCli := httpclient.NewHttpClient()

	// create service module discovery to discover bcs-storage
	disc := discovery.NewRegDiscover(zkAddr)
	if err := disc.Start(); err != nil {
		blog.Errorf("failed to start register discovery, err %s", err.Error())
		return nil, err
	}

	event, err := disc.DiscoverService(types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_STORAGE)
	if err != nil {
		return nil, fmt.Errorf("failed to discovery storage, err %s", err.Error())
	}

	svcs := []*types.ServerInfo{}
	timeout := time.After(3 * time.Second)
	select {
	case <-timeout:
		return nil, fmt.Errorf("discover storage service falied timeout")
	case e := <-event:
		for _, serverStr := range e.Server {
			newSvc := new(types.ServerInfo)
			if err := json.Unmarshal([]byte(serverStr), newSvc); err != nil {
				blog.Warnf("failed to unmarshal %s, err %s", serverStr, err.Error())
				return nil, fmt.Errorf("failed to unmarshal %s, err %s", serverStr, err.Error())
			}
			svcs = append(svcs, newSvc)
		}
		break
	}

	return &StorageClient{
		servers:     svcs,
		chanStorage: event,
		disc:        disc,
		httpCli:     httpCli,
	}, nil
}

// SetTLSConfig set tls config
func (sc *StorageClient) SetTLSConfig(conf *tls.Config) {
	sc.tlsConfig = conf
	sc.httpCli.SetTlsVerityConfig(conf)
}

// Start start storage client
func (sc *StorageClient) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case e := <-sc.chanStorage:
				svcs := []*types.ServerInfo{}
				for _, serverStr := range e.Server {
					newSvc := new(types.ServerInfo)
					if err := json.Unmarshal([]byte(serverStr), newSvc); err != nil {
						blog.Warnf("failed to unmarshal %s, err %s", serverStr, err.Error())
						continue
					}
					svcs = append(svcs, newSvc)
				}
				sc.serversLock.Lock()
				sc.servers = svcs
				sc.serversLock.Unlock()
			case <-ctx.Done():
				blog.Infof("storage client context done")
				return
			}
		}
	}()
}

func (sc *StorageClient) getRandomServer() (*types.ServerInfo, error) {
	sc.serversLock.Lock()
	defer sc.serversLock.Unlock()
	if len(sc.servers) == 0 {
		return nil, fmt.Errorf("no storage servers found")
	}

	index := rand.Intn(len(sc.servers))
	return sc.servers[index], nil
}

// ListResources list resources
func (sc *StorageClient) ListResources(clusterType, clusterID, resourceType string) (
	*common.ListStorageResourceResult, error) {
	return sc.doQueryClusterResources(clusterType, clusterID, resourceType)
}

// WatchClusterResources watch cluster resources
func (sc *StorageClient) WatchClusterResources(clusterID, resourceType string) (chan *common.StorageEvent, error) {
	return sc.doWatchClusterResources(clusterID, resourceType)
}

func (sc *StorageClient) getStorageURL() (string, error) {
	storageInfo, err := sc.getRandomServer()
	if err != nil {
		return "", err
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

func (sc *StorageClient) doQueryClusterResources(clusterType, clusterID, resourceType string) (
	*common.ListStorageResourceResult, error) {
	url, err := sc.getStorageURL()
	if err != nil {
		return nil, err
	}
	realURL := url + fmt.Sprintf(BcsStorageDynamicAllResourceURI, clusterType, clusterID, resourceType)

	blog.Infof("do list storage %s", realURL)

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
	realURL := url + fmt.Sprintf(BcsStorageDynamicWatchResourceURI, clusterID, resourceType)

	blog.Infof("do watch storage %s", realURL)

	header := http.Header{
		"Content-Type": []string{"application/json"},
	}

	// create http client
	httpCli := httpclient.NewHttpClient()
	httpCli.SetTlsVerityConfig(sc.tlsConfig)
	body, err := httpCli.RequestStream(realURL, "POST", header, []byte("{}"))
	if err != nil {
		blog.Errorf("request stream failed, err %s", err.Error())
		return nil, fmt.Errorf("request stream failed, err %s", err.Error())
	}

	decoder := codec.NewDecoder(body, &defaultJSONHandle)
	ch := make(chan *common.StorageEvent)
	go func() {
		defer body.Close()
		for {
			event := new(common.StorageEvent)
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
