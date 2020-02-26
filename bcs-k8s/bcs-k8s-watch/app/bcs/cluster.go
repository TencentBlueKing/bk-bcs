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

package bcs

import (
	"errors"
	"fmt"
	urlparse "net/url"
	"time"

	"github.com/json-iterator/go"
	"github.com/parnurzeal/gorequest"

	"bk-bcs/bcs-common/common/RegisterDiscover"
	glog "bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/ssl"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/options"
)

type ClusterKeeperResponse struct {
	Code    uint                   `json:"code"`
	Result  bool                   `json:"result"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

func getClusterKeeperServerAddress(zkHosts string) (string, error) {
	discovery := RegisterDiscover.NewRegDiscoverEx(zkHosts, 10*time.Second)
	if err := discovery.Start(); nil != err {
		return "", fmt.Errorf("start get storage zk service failed! Error: %v", err)
	}

	// zk
	path := fmt.Sprintf("%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_CLUSTERKEEPER)
	eventChan, err := discovery.DiscoverService(path)
	if err != nil {
		return "", fmt.Errorf("discover cluster keeper fail. Error: %s", err)
	}

	defer func() {
		if err := discovery.Stop(); err != nil {
			glog.Errorf("stop get cluster keeper  addr zk discover failed. reason: %v", err)
		}
	}()

	for {
		select {
		case data := <-eventChan:
			if data.Err != nil {
				return "", fmt.Errorf("get cluster keeper  api failed. reason: %s", data.Err.Error())
			}
			if len(data.Server) == 0 {
				return "", errors.New("get 0 cluster keeper  api address")
			}

			info := types.ServerInfo{}
			if err := jsoniter.Unmarshal([]byte(data.Server[0]), &info); nil != err {
				return "", fmt.Errorf("unmashal cluster keeper  server info failed. reason: %v", err)
			}
			if len(info.IP) == 0 || info.Port == 0 || len(info.Scheme) == 0 {
				return "", fmt.Errorf("get invalid cluster keeper  info: %s", data.Server[0])
			}

			url := fmt.Sprintf("%s://%s:%d", info.Scheme, info.IP, info.Port)
			glog.Infof("get valid cluster keeper  url: %s", url)

			return url, nil
		default:
			time.Sleep(1 * time.Second)
			glog.Info("try to get cluster keeper  api address, waiting for a second.")

		}
	}

}

func GetClusterID(zkHosts string, hostIP string, bcsTLSConfig options.TLS) (string, error) {
	clusterKeeperAddress, err := getClusterKeeperServerAddress(zkHosts)
	if err != nil {
		return "", fmt.Errorf("get clusterkeeper address fail. %v", err)
	}
	glog.Infof("get clusterkeeper address: %s", clusterKeeperAddress)

	r := ClusterKeeperResponse{}
	url := fmt.Sprintf("%s/%s?ip=%s", clusterKeeperAddress, "bcsclusterkeeper/v1/cluster/id/byip", hostIP)

	// support https
	request := gorequest.New()

	urlInfo, err := urlparse.Parse(url)
	if err != nil {
		return "", fmt.Errorf("invalid clusterkepper address: %s not an valid url", url)
	}

	// handler tls
	if urlInfo.Scheme == "https" {
		tlsConfig, err2 := ssl.ClientTslConfVerity(
			bcsTLSConfig.CAFile,
			bcsTLSConfig.CertFile,
			bcsTLSConfig.KeyFile,
			bcsTLSConfig.Password)
		if err2 != nil {
			return "", fmt.Errorf("init tls fail [clientConfig=%v, errors=%s]", tlsConfig, err2)
		}
		request = request.TLSClientConfig(tlsConfig)
	}
	glog.Infof("Get ClusterID from clusterkeeper: url=%s", url)

	resp, _, errs := request.Get(url).Timeout(10 * time.Second).EndStruct(&r)
	if errs != nil {
		glog.Errorf("Get ClusterID from clusterkeeper fail: [url=%s, resp=%v, errors=%s]", url, resp, errs)
		return "", errors.New("HTTP error")
	}

	if !r.Result {
		return "", fmt.Errorf("get clusterID from clusterkeeper fail: result=false [resp=%v]", r)
	}

	data := r.Data
	clusterID := data["clusterID"].(string)

	glog.Infof("Get ClusterID: %s", clusterID)
	return clusterID, nil
}

// GetStorageService returns storage InnerService object for discovery.
func GetStorageService(zkHosts string, bcsTLSConfig options.TLS, customEndpoints []string, isExternal bool) (*InnerService, *RegisterDiscover.RegDiscover, error) {
	discovery := RegisterDiscover.NewRegDiscoverEx(zkHosts, 5*time.Second)
	if err := discovery.Start(); err != nil {
		return nil, nil, fmt.Errorf("get storage service from ZK failed, %+v", err)
	}

	// e.g.
	// zk: 127.0.0.11
	// zknode: bcs/services/endpoints/storage
	path := fmt.Sprintf("%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_STORAGE)
	eventChan, err := discovery.DiscoverService(path)
	if err != nil {
		return nil, nil, fmt.Errorf("discover storage service failed, %+v", err)
	}

	storageService := NewInnerService(types.BCS_MODULE_STORAGE, eventChan, customEndpoints, isExternal)
	go storageService.Watch(bcsTLSConfig)

	return storageService, discovery, nil
}

// GetNetService returns netservice InnerService object for discovery.
func GetNetService(zkHosts string, bcsTLSConfig options.TLS, customEndpoints []string, isExternal bool) (*InnerService, *RegisterDiscover.RegDiscover, error) {
	discovery := RegisterDiscover.NewRegDiscoverEx(zkHosts, 5*time.Second)
	if err := discovery.Start(); err != nil {
		return nil, nil, fmt.Errorf("get netservice from ZK failed, %+v", err)
	}

	// e.g.
	// zk: 127.0.0.11
	// zknode: bcs/services/endpoints/netservice
	path := fmt.Sprintf("%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_NETSERVICE)
	eventChan, err := discovery.DiscoverService(path)
	if err != nil {
		discovery.Stop()
		return nil, nil, fmt.Errorf("discover netservice failed, %+v", err)
	}

	netService := NewInnerService(types.BCS_MODULE_NETSERVICE, eventChan, customEndpoints, isExternal)
	go netService.Watch(bcsTLSConfig)

	return netService, discovery, nil
}
