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

package healthz

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/metric"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
)

type Client interface {
	// get every cluster keeper members info form zookeeper
	DiscoveryClusterKeeperMembers(path string) ([]string, error)
	// retrival each members detail info
	RetrievalComponentMemberInfo(path string) (string, error)
	// get component children list from zookeeper
	GetComponentChildrenPathList(component, clusterid string) ([]string, error)
	// get each component healthz metric.
	GetMetric(addr, scheme string) (*metric.HealthInfo, error)
	// get all the platform component infos from clusterkeeper
	GetClusterInfos(addr, scheme string) (platform *types.DBDataItem, comp []*types.DBDataItem, err error)
}

func NewHealthzClient(zkHost []string, c conf.CertConfig) (Client, error) {

	var tlsConf *tls.Config
	var err error
	if len(c.CAFile) != 0 && len(c.ClientCertFile) != 0 && len(c.ClientKeyFile) != 0 {
		tlsConf, err = ssl.ClientTslConfVerity(c.CAFile, c.ClientCertFile, c.ClientKeyFile, static.ClientCertPwd)
		if err != nil {
			return nil, err
		}
	}
	cli := http.DefaultClient
	cli.Transport = &http.Transport{
		TLSHandshakeTimeout: 5 * time.Second,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 0 * time.Second,
		}).Dial,
		TLSClientConfig: tlsConf,
	}

	zkCli := zkclient.NewZkClient(zkHost)
	return &client{
		zkCli:   zkCli,
		httpCli: cli,
	}, nil
}

type client struct {
	zkCli   *zkclient.ZkClient
	httpCli *http.Client
}

func (c client) RetrievalComponentMemberInfo(path string) (string, error) {
	if err := c.zkCli.ConnectEx(5 * time.Second); err != nil {
		return "", err
	}
	// defer c.zkCli.Close()

	return c.zkCli.Get(path)
}

func (c client) GetComponentChildrenPathList(component, clusterid string) ([]string, error) {
	var basePath string
	if component == "health" {
		basePath = fmt.Sprintf("%s/%s/%s", types.BCS_SERV_BASEPATH, component, "master")
	} else {
		basePath = fmt.Sprintf("%s/%s", types.BCS_SERV_BASEPATH, component)
	}
	if len(clusterid) != 0 {
		basePath = fmt.Sprintf("%s/%s", basePath, clusterid)
	}

	return c.zkChildrenIterator(basePath)
}

func (c client) zkChildrenIterator(path string) ([]string, error) {
	isChild := func(child string) bool {
		return strings.HasPrefix(child, "_c_")
	}

	if err := c.zkCli.ConnectEx(5 * time.Second); err != nil {
		return []string{}, err
	}

	// defer c.zkCli.Close()

	children, err := c.zkCli.GetChildren(path)
	if err != nil {
		return []string{}, err
	}

	var list []string
	for _, child := range children {
		newPath := fmt.Sprintf("%s/%s", path, child)
		if isChild(child) {
			list = append(list, newPath)
			continue
		}

		itChildren, err := c.zkChildrenIterator(newPath)
		if err != nil {
			return []string{}, err
		}

		list = append(list, itChildren...)
	}

	return list, nil
}

func (c client) DiscoveryClusterKeeperMembers(path string) ([]string, error) {
	if err := c.zkCli.ConnectEx(5 * time.Second); err != nil {
		return []string{}, err
	}
	// defer c.zkCli.Close()

	var components []string
	children, err := c.zkCli.GetChildren(path)
	if err != nil {
		return components, err
	}

	for _, child := range children {
		subPath := fmt.Sprintf("%s/%s", path, child)
		com, err := c.zkCli.Get(subPath)
		if err != nil {
			return components, err
		}
		components = append(components, com)
	}

	return components, nil
}

func (c client) GetMetric(addr, scheme string) (*metric.HealthInfo, error) {
	url := fmt.Sprintf("%s://%s/%s", scheme, addr, "healthz")

	// TODO: switch http scheme, especially.
	resp, err := c.httpCli.Get(url)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info metric.HealthInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func (c client) GetClusterInfos(addr, scheme string) (platform *types.DBDataItem, comp []*types.DBDataItem, err error) {
	url := fmt.Sprintf("%s://%s/%s", scheme, addr, "bcsclusterkeeper/v1/clusterinfo/list/master")

	type Response struct {
		Code    int    `json:"code"`
		Result  bool   `json:"result"`
		Message string `json:"message"`
		Data    struct {
			ClusterInfo []*types.DBDataItem `json:"clusterinfo"`
		} `json:"data"`
	}

	// TODO: switch http scheme, especially https.
	resp, err := c.httpCli.Get(url)
	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	var r Response
	if err = json.Unmarshal(data, &r); err != nil {
		blog.Errorf("get cluster info, but unmarshal failed, err: %v, body: %s", err, data)
		return nil, nil, err
	}

	if !r.Result {
		return platform, comp, errors.New(r.Message)
	}

	comp = r.Data.ClusterInfo
	url = fmt.Sprintf("%s://%s/%s", scheme, addr, "bcsclusterkeeper/v1/serviceinfo")
	cresp, err := c.httpCli.Get(url)
	if err != nil {
		return nil, nil, err
	}

	body, err := ioutil.ReadAll(cresp.Body)
	if err != nil {
		return nil, nil, err
	}

	var cr Response
	if err = json.Unmarshal(body, &cr); err != nil {
		blog.Errorf("get platform service info, but unmarshal failed, err: %v, body: %s", err, body)
		return nil, nil, err
	}

	if !cr.Result {
		return platform, comp, errors.New(cr.Message)
	}

	if len(cr.Data.ClusterInfo) != 0 {
		platform = cr.Data.ClusterInfo[0]
	}

	return platform, comp, nil

}
