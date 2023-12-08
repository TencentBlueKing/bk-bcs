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

// Package bcs
package bcs

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/rest"
	k8srest "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// GetClusters get clustermanager clusters by ids
func (cm *ClusterManager) GetClusters(clusterIds []string) ([]cmproto.Cluster, error) {
	klog.V(6).Infof("start ClusterManager request")
	var (
		rt         http.RoundTripper
		httpClient *http.Client
	)

	rt = &BcsTransport{token: cm.token}
	httpClient = &http.Client{Transport: rt}

	if len(clusterIds) == 0 {
		clusterIds = append(clusterIds, "")
	}

	resultList := make([]cmproto.Cluster, 0, 0)
	for _, clusterId := range clusterIds {
		svcUrl, _ := url.Parse(cm.url + _urlMap["GetClusters"])
		if clusterId != "" {
			svcUrl.Path = path.Join(svcUrl.Path, clusterId)
		}

		req := rest.NewRequest(httpClient, "GET", svcUrl, nil)
		data, err := req.Do()
		if err != nil {
			return nil, err
		}

		result := rest.BaseResponse{}
		err = json.Unmarshal(data, &result)
		if err != nil {
			return nil, err
		}

		if !result.Result {
			e := errors.New(fmt.Sprintf("cluster response result failed: %s", result.Msg))
			klog.V(3).Info(e.Error())
			return nil, e
		}

		clusterData, _ := json.Marshal(result.Data)
		clusterList := make([]cmproto.Cluster, 0, 0)
		err = json.Unmarshal(clusterData, &clusterList)
		if err != nil {
			cluster := cmproto.Cluster{}
			err = json.Unmarshal(clusterData, &cluster)
			if err != nil {
				e := errors.New(fmt.Sprintf("Unmarshal cluster response failed %s", err.Error()))
				klog.V(3).Info(e.Error())
				return nil, e
			} else {
				clusterList = append(clusterList, cluster)
			}
		}
		resultList = append(resultList, clusterList...)
	}
	return resultList, nil
}

// GetNodesByClusterId get cluster nodes
func (cm *ClusterManager) GetNodesByClusterId(clusterId string) ([]cmproto.Node, error) {
	if clusterId == "" {
		return nil, errors.New("ClusterId cannot be blank")
	}
	svcUrl, _ := url.Parse(cm.url + fmt.Sprintf(_urlMap["GetNodesByClusterId"], clusterId))
	klog.V(6).Infof("start ClusterManager request %s", svcUrl.String())

	var (
		rt         http.RoundTripper
		httpClient *http.Client
	)

	rt = &BcsTransport{token: cm.token}
	httpClient = &http.Client{Transport: rt, Timeout: 10 * time.Second}

	req := rest.NewRequest(httpClient, "GET", svcUrl, nil)
	data, err := req.Do()
	if err != nil {
		return nil, err
	}

	result := rest.BaseResponse{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	if !result.Result {
		e := errors.New(fmt.Sprintf("cluster response result failed: %s", result.Msg))
		klog.V(3).Info(e.Error())
		return nil, e
	}

	nodeData, _ := json.Marshal(result.Data)
	nodeList := make([]cmproto.Node, 0, 0)
	err = json.Unmarshal(nodeData, &nodeList)
	if err != nil {
		e := errors.New(fmt.Sprintf("Unmarshal cluster response failed %s", err.Error()))
		klog.V(3).Info(e.Error())
		return nil, e
	}
	return nodeList, nil
}

// GetNode get node detail
func (cm *ClusterManager) GetNode(ip string) (*cmproto.Node, error) {
	svcUrl, _ := url.Parse(cm.url + fmt.Sprintf(_urlMap["GetNode"], ip))
	klog.V(6).Infof("start GetNode request %s", svcUrl.String())

	var (
		rt         http.RoundTripper
		httpClient *http.Client
	)

	rt = &BcsTransport{token: cm.token}
	httpClient = &http.Client{Transport: rt}

	req := rest.NewRequest(httpClient, "GET", svcUrl, nil)
	data, err := req.Do()
	if err != nil {
		return nil, err
	}

	result := rest.BaseResponse{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	if !result.Result {
		e := errors.New(fmt.Sprintf("getnode response result failed: %s", result.Msg))
		klog.V(3).Info(e.Error())
		return nil, e
	}

	nodeData, _ := json.Marshal(result.Data)
	nodeList := make([]cmproto.Node, 0, 0)
	err = json.Unmarshal(nodeData, &nodeList)
	if err != nil {
		e := errors.New(fmt.Sprintf("Unmarshal getnode response failed %s", err.Error()))
		klog.V(3).Info(e.Error())
		return nil, e
	}

	if len(nodeList) != 1 {
		e := errors.New(fmt.Sprintf("getnode result num wrong %s", err.Error()))
		klog.V(3).Info(e.Error())
		return nil, e
	}

	return &nodeList[0], nil
}

// GetKubeconfig get cluster kubeconfig
func (cm *ClusterManager) GetKubeconfig(clusterID string) *k8srest.Config {
	config := &k8srest.Config{
		Host:        fmt.Sprintf(cm.apiGatewayUrl, clusterID),
		BearerToken: cm.apiGatewayToken,
		TLSClientConfig: k8srest.TLSClientConfig{
			Insecure: true,
		},
	}

	return config
}
