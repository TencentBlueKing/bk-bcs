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

package gw

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	rdiscover "bk-bcs/bcs-common/common/RegisterDiscover"
	"bk-bcs/bcs-common/common/blog"
)

// ServerInfo gw server information in zk path
type ServerInfo struct {
	IP         string `json:"ip"`
	Port       uint   `json:"port"`
	MetricPort uint   `json:"metric_port"`
	HostName   string `json:"hostname"`
	//http, https
	Scheme  string `json:"scheme"`
	Version string `json:"version"`
	Cluster string `json:"cluster"`
	Pid     int    `json:"pid"`
}

// Interface interface to call gw api
type Interface interface {
	Run() error
	Update(svcs []*Service) error
	Delete(svcs []*Service) error
}

// Client client for operate gw services
type Client struct {
	tlsConfig  *tls.Config
	Cluster    string
	ZkPath     string
	rdiscover  *rdiscover.RegDiscover
	ctx        context.Context
	masterInfo *ServerInfo
	masterMutx sync.Mutex
}

// NewClient create new client for gw server
func NewClient(ctx context.Context, cluster string, r *rdiscover.RegDiscover, zkPath string) *Client {
	return &Client{
		Cluster:   cluster,
		rdiscover: r,
		ZkPath:    zkPath,
		ctx:       ctx,
	}
}

// NewClientWithTLS create new client for tls gw server
func NewClientWithTLS(ctx context.Context, cluster string, r *rdiscover.RegDiscover, zkPath string, tlsConfig *tls.Config) *Client {
	return &Client{
		tlsConfig: tlsConfig,
		Cluster:   cluster,
		rdiscover: r,
		ZkPath:    zkPath,
		ctx:       ctx,
	}
}

// Run start discover gw server master
func (c *Client) Run() error {
	ch, err := c.rdiscover.DiscoverService(c.ZkPath)
	if err != nil {
		blog.Errorf("discover %s failed, err %s", c.ZkPath, err.Error())
	}
	for {
		select {
		case event := <-ch:
			serverInfo := new(ServerInfo)
			if event.Err != nil {
				blog.Errorf("get err in discovery event, err %s", event.Err.Error())
				c.masterMutx.Lock()
				serverInfo = nil
				c.masterMutx.Unlock()
				continue
			}
			if len(event.Server) == 0 {
				blog.Errorf("get no server in discovery event")
				c.masterMutx.Lock()
				serverInfo = nil
				c.masterMutx.Unlock()
				continue
			}
			err := json.Unmarshal([]byte(event.Server[0]), serverInfo)
			if err != nil {
				blog.Errorf("unmarshal str %s to serverInfo failed, err %s", event.Server[0], err.Error())
				continue
			}
			blog.Infof("get master gw server %s %d", serverInfo.IP, serverInfo.Port)
			c.masterMutx.Lock()
			c.masterInfo = serverInfo
			c.masterMutx.Unlock()
		case <-c.ctx.Done():
			blog.Infof("gw client received stop event")
			err := c.rdiscover.Stop()
			if err != nil {
				blog.Errorf("rdiscover stop failed, err %s", err.Error())
			}
		}
	}
}

// get the master gw server address and port
// if get error from zk, return error
func (c *Client) getMasterServer() (string, int, error) {
	var addr string
	var port int
	c.masterMutx.Lock()
	defer c.masterMutx.Unlock()
	if c.masterInfo == nil {
		return "", 0, fmt.Errorf("gw server master is nil")
	}
	addr = c.masterInfo.IP
	port = int(c.masterInfo.Port)
	return addr, port, nil
}

// getHttpClient create http client according tlsconfig
func (c *Client) getHTTPClient(timeout int) (*http.Client, string) {
	prefix := "http://"
	// extend from Default Transport in package net, and disable keepalive
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     true,
	}
	if c.tlsConfig != nil {
		prefix = "https://"
		transport.TLSClientConfig = c.tlsConfig
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   time.Second * time.Duration(int64(timeout)),
	}
	return httpClient, prefix
}

// do http request to gw server master with certain path, method and data
func (c *Client) doRequest(path, method string, data []byte) ([]byte, error) {
	addr, port, err := c.getMasterServer()
	if err != nil {
		return nil, fmt.Errorf("request failed, err %s", err.Error())
	}
	fullAddr := addr + ":" + strconv.Itoa(port)
	httpClient, prefix := c.getHTTPClient(5)
	url := prefix + fullAddr + path
	blog.Infof("request: method %s, url %s, data %s", method, url, string(data))
	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("create http request failed, %s", err.Error())
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("http client send request failed, %s", err.Error())
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http response %s from %s", response.Status, fullAddr)
	}
	defer response.Body.Close()
	datas, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read netservice response failed, %s", err.Error())
	}
	blog.Infof("response: %s", string(datas))
	return datas, nil
}

// Update call update api to gw concentrator
func (c *Client) Update(svcs []*Service) error {
	req := new(UpdateRequest)
	req.Cluster = c.Cluster
	req.ServiceList = svcs
	data, err := json.Marshal(req)
	if err != nil {
		blog.Errorf("json encode failed, err %s", err.Error())
		return fmt.Errorf("json encode failed, err %s", err.Error())
	}
	respData, err := c.doRequest("/stgw/services", "POST", data)
	if err != nil {
		blog.Errorf("do request failed, err %s", err.Error())
		return fmt.Errorf("do request failed, err %s", err.Error())
	}
	resp := new(UpdateResponse)
	err = json.Unmarshal(respData, resp)
	if err != nil {
		blog.Errorf("json decode failed, err %s", err.Error())
		return fmt.Errorf("json decode failed, err %s", err.Error())
	}
	if resp.Code != 200 {
		return fmt.Errorf("update failed, message %s, code %d", resp.Message, resp.Code)
	}
	return nil
}

// Delete call delete api to gw concentrator
func (c *Client) Delete(svcs []*Service) error {
	req := new(DeleteRequest)
	req.Cluster = c.Cluster
	req.ServiceList = svcs
	data, err := json.Marshal(req)
	if err != nil {
		blog.Errorf("json encode failed, err %s", err.Error())
		return fmt.Errorf("json encode failed, err %s", err.Error())
	}
	respData, err := c.doRequest("/stgw/services", "DELETE", data)
	if err != nil {
		blog.Errorf("do request failed, err %s", err.Error())
		return fmt.Errorf("do request failed, err %s", err.Error())
	}
	resp := new(DeleteResponse)
	err = json.Unmarshal(respData, resp)
	if err != nil {
		blog.Errorf("json decode failed, err %s", err.Error())
		return fmt.Errorf("json decode failed, err %s", err.Error())
	}
	if resp.Code != 200 {
		return fmt.Errorf("delete failed, message %s, code %d", resp.Message, resp.Code)
	}
	return nil
}
