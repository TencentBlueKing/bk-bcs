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

package netservice

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/bcstypes"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/pkg/netservice/types"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	defaultDiscoveryPath = "/bcs/services/endpoints/netservice"
	defaultVersion       = "v1"
)

//Client define http client interface for bcs-ipam
type Client interface {
	GetNetService(zkHost []string) error                   //Get Netservice ip address from zookeeper
	RegisterPool(pool *types.NetPool) error                //register ip pool info to netservice
	UpdatePool(pool *types.NetPool) error                  //update info for pool
	GetPool(cluster, net string) ([]*types.NetPool, error) //Get pool info from netservice
	DeletePool(cluster, net string) error
	ListAllPool() ([]*types.NetPool, error)
	ListAllPoolWithCluster(cluster string) ([]*types.NetPool, error)
	RegisterHost(host *types.HostInfo) error //register host info attaching to ip pool in netservice
	DeleteHost(host string, ips []string) error
	GetHostInfo(host string, timeout int) (*types.HostInfo, error)                   //Get host info by host ip address
	LeaseIPAddr(lease *types.IPLease, timeout int) (*types.IPInfo, error)            //lease ip address from Netservice by containerId/Host/ipaddress(if needed)
	ReleaseIPAddr(release *types.IPRelease, ipInfo *types.IPInfo, timeout int) error // release ip address by containerId & host
	UpdateIPInstance(inst *types.IPInst) error
	TransferIPAttr(input *types.TranIPAttrInput) error
}

//####################################################################
// client response data structure
//####################################################################

//NewClient create new client with zookeeper
func NewClient() (Client, error) {
	c := &client{
		tlsConfig: nil,
		random:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	return c, nil
}

//NewTLSClient create tls client with root CA, private key, public key and password
func NewTLSClient(ca string, key string, crt string, passwd string) (Client, error) {
	config, err := ssl.ClientTslConfVerity(ca, crt, key, passwd)
	if err != nil {
		return nil, err
	}
	//create client
	c := &client{
		tlsConfig: config,
		random:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	return c, nil
}

//NewClientWithTLS create tls client with tls
func NewClientWithTLS(conf *tls.Config) (Client, error) {
	//create client
	c := &client{
		tlsConfig: conf,
		random:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	return c, nil
}

type client struct {
	tlsConfig *tls.Config //https config
	zkHost    []string    //zookeeper list
	netsvrs   []string    //NetService list
	random    *rand.Rand  //rand for connection
}

//GetNetService get bcs-netservice from zookeeper
func (c *client) GetNetService(zkHost []string) error {
	netSvr := os.Getenv("NETSVR_ADDR")
	if len(netSvr) != 0 {
		c.netsvrs = append(c.netsvrs, netSvr)
		return nil
	}
	if len(zkHost) == 0 {
		return fmt.Errorf("zookeeper host list empty")
	}
	c.zkHost = zkHost
	bcsClient := zkclient.NewZkClientWithoutLogger(zkHost)
	if err := bcsClient.ConnectEx(time.Second * 5); err != nil {
		return fmt.Errorf("zk connect err, %v", err)
	}
	rSvrList, stat, err := bcsClient.GetChildrenEx(defaultDiscoveryPath)
	if err != nil {
		return fmt.Errorf("get netservice err, %v", err)
	} else if stat == nil {
		return fmt.Errorf("zk status lost")
	}
	if len(rSvrList) == 0 {
		return fmt.Errorf("no bcs-netservice server node detected")
	}
	//get all datas
	for _, node := range rSvrList {
		child := filepath.Join(defaultDiscoveryPath, node)
		data, _, err := bcsClient.GetEx(child)
		if err != nil {
			continue
		}
		info := new(bcstypes.NetServiceInfo)
		if err := json.Unmarshal(data, info); err != nil {
			continue
		}
		svrInfo := info.IP + ":" + strconv.Itoa(int(info.Port))
		c.netsvrs = append(c.netsvrs, svrInfo)
	}
	if len(c.netsvrs) == 0 {
		return fmt.Errorf("parse %d bcs-netservice node json data fail", len(rSvrList))
	}
	return nil
}

//RegisterPool register pool info to bcs-netservice
func (c *client) RegisterPool(pool *types.NetPool) error {
	if len(pool.Cluster) == 0 {
		return fmt.Errorf("Lost cluster info")
	}
	if len(pool.Available) == 0 && len(pool.Reserved) == 0 {
		return fmt.Errorf("Lost ip address info")
	}
	//create net request
	netRequest := &types.NetRequest{
		Type: types.RequestType_POOL,
		Pool: pool,
	}
	reqDatas, err := json.Marshal(netRequest)
	if err != nil {
		return fmt.Errorf("RegisterPool encode failed, %s", err)
	}
	seq := c.random.Perm(len(c.netsvrs))
	httpClient, prefix := c.getHTTPClient(3)
	uri := prefix + c.netsvrs[seq[0]] + "/v1/pool"
	request, err := http.NewRequest("POST", uri, bytes.NewBuffer(reqDatas))
	if err != nil {
		return fmt.Errorf("RegisterPool create http request failed, %s", err.Error())
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("RegisterPool send request failed, %s", err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("RegisterPool Got err response: %s", response.Status)
	}
	defer response.Body.Close()
	resDatas, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("RegisterPool read response body failed, %s", err)
	}
	var netRes types.NetResponse
	if err := json.Unmarshal(resDatas, &netRes); err != nil {
		return fmt.Errorf("RegisterPool decode response failed, %s", err)
	}
	if netRes.Code != 0 {
		return fmt.Errorf("RegisterPool failed, %s", netRes.Message)
	}
	return nil
}

//UpdatePool update pool info
func (c *client) UpdatePool(pool *types.NetPool) error {
	if len(pool.Cluster) == 0 {
		return fmt.Errorf("Lost cluster info")
	}
	if len(pool.Available) == 0 && len(pool.Reserved) == 0 {
		return fmt.Errorf("Lost ip address info")
	}
	//create net request
	netRequest := &types.NetRequest{
		Type: types.RequestType_POOL,
		Pool: pool,
	}
	reqDatas, err := json.Marshal(netRequest)
	if err != nil {
		return fmt.Errorf("Updatepool encode failed, %s", err)
	}
	blog.Info("udpate pool string:%s", reqDatas)
	seq := c.random.Perm(len(c.netsvrs))
	httpClient, prefix := c.getHTTPClient(3)
	uri := prefix + c.netsvrs[seq[0]] + "/v1/pool/" + pool.Cluster + "/" + pool.Net
	request, err := http.NewRequest("PUT", uri, bytes.NewBuffer(reqDatas))
	if err != nil {
		return fmt.Errorf("UpdatePool create http request failed, %s", err.Error())
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("UpdatePool send request failed, %s", err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("UpdatePool Got err response: %s", response.Status)
	}
	defer response.Body.Close()
	resDatas, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("UpdatePool read response body failed, %s", err)
	}
	var netRes types.NetResponse
	if err := json.Unmarshal(resDatas, &netRes); err != nil {
		return fmt.Errorf("UpdatePool decode response failed, %s", err)
	}
	if netRes.Code != 0 {
		return fmt.Errorf("UpdatePool failed, %s", netRes.Message)
	}
	return nil
}

//GetPool get pool info from netservice
func (c *client) GetPool(cluster, net string) ([]*types.NetPool, error) {
	if len(cluster) == 0 || len(net) == 0 {
		return nil, fmt.Errorf("Lost cluster or network segment in request")
	}
	if len(c.netsvrs) == 0 {
		return nil, fmt.Errorf("no available bcs-netservice")
	}
	seq := c.random.Perm(len(c.netsvrs))
	httpClient, prefix := c.getHTTPClient(3)
	uri := prefix + c.netsvrs[seq[0]] + "/v1/pool/" + cluster + "/" + net
	request, err := http.NewRequest("GET", uri, nil)
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
		return nil, fmt.Errorf("http response %s from %s", response.Status, c.netsvrs[seq[0]])
	}
	defer response.Body.Close()
	datas, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read netservice response failed, %s", err.Error())
	}
	netRes := &types.NetResponse{}
	if err := json.Unmarshal(datas, netRes); err != nil {
		return nil, fmt.Errorf("decode netservice response failed, %s", err.Error())
	}
	if netRes.Type != types.ResponseType_POOL {
		return nil, fmt.Errorf("response data type expect %d, but got %d", types.ResponseType_HOST, netRes.Type)
	}
	if !netRes.IsSucc() {
		return nil, fmt.Errorf("pool request fialed: %s", netRes.Message)
	}
	if len(netRes.Pool) == 0 {
		return nil, nil
	}
	return netRes.Pool, nil
}

func (c *client) ListAllPool() ([]*types.NetPool, error) {
	if len(c.netsvrs) == 0 {
		return nil, fmt.Errorf("no available bcs-netservice")
	}
	seq := c.random.Perm(len(c.netsvrs))
	httpClient, prefix := c.getHTTPClient(3)
	uri := prefix + c.netsvrs[seq[0]] + "/v1/pool"
	request, err := http.NewRequest("GET", uri, nil)
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
		return nil, fmt.Errorf("http response %s from %s", response.Status, c.netsvrs[seq[0]])
	}
	defer response.Body.Close()
	datas, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read netservice response failed, %s", err.Error())
	}
	netRes := &types.NetResponse{}
	if err := json.Unmarshal(datas, netRes); err != nil {
		return nil, fmt.Errorf("decode netservice response failed, %s", err.Error())
	}
	if netRes.Type != types.ResponseType_POOL {
		return nil, fmt.Errorf("response data type expect %d, but got %d", types.ResponseType_HOST, netRes.Type)
	}
	if !netRes.IsSucc() {
		return nil, fmt.Errorf("pool request fialed: %s", netRes.Message)
	}
	if len(netRes.Pool) == 0 {
		return nil, nil
	}
	return netRes.Pool, nil
}

func (c *client) ListAllPoolWithCluster(cluster string) ([]*types.NetPool, error) {
	if len(c.netsvrs) == 0 {
		return nil, fmt.Errorf("no available bcs-netservice")
	}
	seq := c.random.Perm(len(c.netsvrs))
	httpClient, prefix := c.getHTTPClient(3)
	uri := prefix + c.netsvrs[seq[0]] + "/v1/pool/" + cluster + "?info=detail"
	request, err := http.NewRequest("GET", uri, nil)
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
		return nil, fmt.Errorf("http response %s from %s", response.Status, c.netsvrs[seq[0]])
	}
	defer response.Body.Close()
	datas, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read netservice response failed, %s", err.Error())
	}
	netRes := &types.NetResponse{}
	if err := json.Unmarshal(datas, netRes); err != nil {
		return nil, fmt.Errorf("decode netservice response failed, %s", err.Error())
	}
	if netRes.Type != types.ResponseType_POOL {
		return nil, fmt.Errorf("response data type expect %d, but got %d", types.ResponseType_HOST, netRes.Type)
	}
	if !netRes.IsSucc() {
		return nil, fmt.Errorf("pool request fialed: %s", netRes.Message)
	}
	if len(netRes.Pool) == 0 {
		return nil, nil
	}
	return netRes.Pool, nil
}

// DeletePool delete pool
func (c *client) DeletePool(cluster, net string) error {
	if len(cluster) == 0 || len(net) == 0 {
		return fmt.Errorf("neither cluster nor net can be empty")
	}
	seq := c.random.Perm(len(c.netsvrs))
	httpClient, prefix := c.getHTTPClient(250)
	uri := prefix + c.netsvrs[seq[0]] + "/v1/pool/" + cluster + "/" + net
	request, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return fmt.Errorf("create http request failed, %s", err.Error())
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("http client send request failed, %s", err.Error())
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("http response %s from %s", response.Status, c.netsvrs[seq[0]])
	}
	defer response.Body.Close()
	datas, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read netservice response failed, %s", err.Error())
	}
	netRes := &types.NetResponse{}
	if err := json.Unmarshal(datas, netRes); err != nil {
		return fmt.Errorf("decode netservice response failed, %s", err.Error())
	}
	if netRes.Code != 0 {
		return fmt.Errorf("netservice response code not zero, response data %s", string(datas))
	}
	return nil
}

//RegisterHost register host info to bcs-netservice
func (c *client) RegisterHost(host *types.HostInfo) error {
	if len(c.netsvrs) == 0 {
		return fmt.Errorf("no available bcs-netservice")
	}
	if len(host.Pool) == 0 || len(host.Cluster) == 0 {
		return fmt.Errorf("Host info is invalid")
	}
	seq := c.random.Perm(len(c.netsvrs))
	httpClient, prefix := c.getHTTPClient(3)
	uri := prefix + c.netsvrs[seq[0]] + "/v1/host"
	reqDatas, err := json.Marshal(host)
	if err != nil {
		return fmt.Errorf("Host json encode failed, %s", err)
	}
	request, err := http.NewRequest("POST", uri, bytes.NewBuffer(reqDatas))
	if err != nil {
		return fmt.Errorf("create host request failed, %s", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send request failed, %s", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("RegisterHost got response %s from %s", response.Status, c.netsvrs[seq[0]])
	}
	resDatas, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read netservice RegisterHost response failed, %s", err.Error())
	}
	var hostRes types.NetResponse
	if err := json.Unmarshal(resDatas, &hostRes); err != nil {
		return fmt.Errorf("decode RegisterHost response failed, %s", err)
	}
	if hostRes.Code == 0 {
		return nil
	}
	return fmt.Errorf(hostRes.Message)
}

//DeleteHost when host has container or any ip belongs to the host is active, it can't be deleted
func (c *client) DeleteHost(host string, ips []string) error {
	if len(c.netsvrs) == 0 {
		return fmt.Errorf("no available bcs-netservice")
	}
	if len(host) == 0 {
		return fmt.Errorf("bad request host ip")
	}
	seq := c.random.Perm(len(c.netsvrs))
	httpClient, prefix := c.getHTTPClient(3)
	uri := prefix + c.netsvrs[seq[0]] + "/v1/host/" + host

	//create net request
	netRequest := &types.NetRequest{
		Type: types.RequestType_HOST,
		IPs:  ips,
	}
	reqDatas, err := json.Marshal(netRequest)
	if err != nil {
		return fmt.Errorf("Delete host encode failed, %s", err)
	}

	request, err := http.NewRequest("DELETE", uri, bytes.NewBuffer(reqDatas))

	if err != nil {
		return fmt.Errorf("delete host request failed, %s", err.Error())
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send delete host request to net service failed, %s", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("DeleteHost got response %s from %s", response.Status, c.netsvrs[seq[0]])
	}
	resDatas, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read netservice DeleteHost response failed, %s", err.Error())
	}
	var hostRes types.NetResponse
	if err := json.Unmarshal(resDatas, &hostRes); err != nil {
		return fmt.Errorf("decode DeleteHost response failed, %s", err)
	}
	if hostRes.Code == 0 {
		return nil
	}
	return fmt.Errorf(hostRes.Message)
}

//GetHostInfo Get host info by host ip address
func (c *client) GetHostInfo(host string, timeout int) (*types.HostInfo, error) {
	if len(host) == 0 {
		return nil, fmt.Errorf("host ip address lost")
	}
	if timeout < 1 {
		return nil, fmt.Errorf("timeout must > 1")
	}
	seq := c.random.Perm(len(c.netsvrs))
	httpClient, prefix := c.getHTTPClient(timeout)
	var lastErr string
	for _, index := range seq {
		request, reqErr := http.NewRequest("GET", prefix+c.netsvrs[index]+"/v1/host/"+host, nil)
		if reqErr != nil {
			return nil, fmt.Errorf("create GetHostInfo request failed, %s", reqErr.Error())
		}
		request.Header.Set("Accept", "application/json")
		response, err := httpClient.Do(request)
		if err != nil {
			lastErr = err.Error()
			continue
		}
		if response.StatusCode != http.StatusOK {
			lastErr = fmt.Sprintf("NetService %s response code: %d", c.netsvrs[index], response.StatusCode)
			continue
		}
		defer response.Body.Close()
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		netRes := &types.NetResponse{}
		if err := json.Unmarshal(data, netRes); err != nil {
			return nil, err
		}
		if netRes.Type != types.ResponseType_HOST {
			return nil, fmt.Errorf("response data type expect %d, but got %d", types.ResponseType_HOST, netRes.Type)
		}
		if !netRes.IsSucc() {
			return nil, fmt.Errorf("request fialed: %s", netRes.Message)
		}
		if len(netRes.Host) == 0 {
			return nil, fmt.Errorf("response err, host info lost")
		}
		hostInfo := netRes.Host[0]
		if hostInfo.IPAddr != host {
			return nil, fmt.Errorf("response ip address expect %s, but got %s", host, hostInfo.IPAddr)
		}
		return hostInfo, nil
	}
	return nil, fmt.Errorf("all netservice failed, %s", lastErr)
}

//LeaseIPAddr lease one ip address from bcs-netservice
func (c *client) LeaseIPAddr(lease *types.IPLease, timeout int) (*types.IPInfo, error) {
	//create net request
	request := &types.NetRequest{
		Type:  types.RequestType_LEASE,
		Lease: lease,
	}
	if len(lease.Host) == 0 || len(lease.Container) == 0 {
		return nil, fmt.Errorf("host or container lost")
	}
	if len(c.netsvrs) == 0 {
		return nil, fmt.Errorf("get no online netservice")
	}
	seq := c.random.Perm(len(c.netsvrs))
	httpClient, prefix := c.getHTTPClient(timeout)
	requestData, jsonErr := json.Marshal(request)
	if jsonErr != nil {
		return nil, jsonErr
	}
	var lastErr string
	for _, index := range seq {
		httpRequest, reqErr := http.NewRequest("POST", prefix+c.netsvrs[index]+"/v1/allocator", bytes.NewBuffer(requestData))
		if reqErr != nil {
			return nil, fmt.Errorf("create request failed, %s", reqErr.Error())
		}
		httpRequest.Header.Set("Content-Type", "application/json")
		httpResponse, err := httpClient.Do(httpRequest)
		if err != nil {
			//do err, try next one
			lastErr = err.Error()
			continue
		}
		if httpResponse.StatusCode != http.StatusOK {
			lastErr = fmt.Sprintf("http response code: %d", httpResponse.StatusCode)
			continue
		}
		defer httpResponse.Body.Close()
		data, err := ioutil.ReadAll(httpResponse.Body)
		if err != nil {
			return nil, err
		}
		response := &types.NetResponse{}
		if err := json.Unmarshal(data, response); err != nil {
			return nil, err
		}
		if response.Type != types.ResponseType_LEASE {
			return nil, fmt.Errorf("response data type expect %d, but got %d", types.ResponseType_LEASE, response.Type)
		}
		if !response.IsSucc() {
			return nil, fmt.Errorf("request fialed: %s", response.Message)
		}
		if response.Info == nil || len(response.Info) == 0 {
			return nil, fmt.Errorf("response err, ip info lost")
		}
		ipInfo := response.Info[0]
		//check if response ip addr is what we need
		if lease.IPAddr != "" && lease.IPAddr != ipInfo.IPAddr {
			//todo(DeveloperJim): Get unexpect ip address, need to release
			return nil, fmt.Errorf("expect ipaddr %s, but got %s", lease.IPAddr, ipInfo.IPAddr)
		}
		if len(ipInfo.Gateway) == 0 || ipInfo.Mask == 0 {
			return nil, fmt.Errorf("ip lease failed, gateway/mask info lost")
		}
		return ipInfo, nil
	}
	return nil, fmt.Errorf("all netservice failed, %s", lastErr)
}

//ReleaseIPAddr release ip address to bcs-netservice
func (c *client) ReleaseIPAddr(release *types.IPRelease, ipInfo *types.IPInfo, timeout int) error {
	//create net request
	request := &types.NetRequest{
		Type:    types.RequestType_RELEASE,
		Release: release,
	}
	if len(release.Host) == 0 || len(release.Container) == 0 {
		return fmt.Errorf("host or container lost")
	}
	if len(c.netsvrs) == 0 {
		return fmt.Errorf("get no online netservice")
	}
	seq := c.random.Perm(len(c.netsvrs))
	httpClient, prefix := c.getHTTPClient(timeout)
	requestData, jsonErr := json.Marshal(request)
	if jsonErr != nil {
		return jsonErr
	}
	var lastErr string
	for _, index := range seq {
		httpRequest, reqErr := http.NewRequest("DELETE", prefix+c.netsvrs[index]+"/v1/allocator", bytes.NewBuffer(requestData))
		if reqErr != nil {
			return fmt.Errorf("create request failed, %s", reqErr.Error())
		}
		httpRequest.Header.Set("Content-Type", "application/json")
		httpResponse, err := httpClient.Do(httpRequest)
		if err != nil {
			//do err, try next one
			lastErr = err.Error()
			continue
		}
		if httpResponse.StatusCode != http.StatusOK {
			lastErr = fmt.Sprintf("http response code: %d", httpResponse.StatusCode)
			continue
		}
		defer httpResponse.Body.Close()
		data, err := ioutil.ReadAll(httpResponse.Body)
		if err != nil {
			return err
		}
		response := &types.NetResponse{}
		if err := json.Unmarshal(data, response); err != nil {
			return err
		}
		if response.Type != types.ResponseType_RELEASE {
			return fmt.Errorf("response data type expect %d, but got %d", types.ResponseType_RELEASE, response.Type)
		}
		if !response.IsSucc() {
			return fmt.Errorf("request fialed: %s", response.Message)
		}
		return nil
	}
	return fmt.Errorf("all netservice failed, %s", lastErr)
}

//UpdateIPInstance update ip instance info, expecially MacAddress
func (c *client) UpdateIPInstance(inst *types.IPInst) error {
	if inst == nil {
		return fmt.Errorf("Lost instance data")
	}
	if inst.Cluster == "" || inst.IPAddr == "" || inst.Pool == "" {
		return fmt.Errorf("ip instance lost key data")
	}
	seq := c.random.Perm(len(c.netsvrs))
	httpClient, prefix := c.getHTTPClient(3)
	uri := prefix + c.netsvrs[seq[0]] + "/v1/ipinstance"
	reqDatas, err := json.Marshal(inst)
	if err != nil {
		return fmt.Errorf("IP Instance json encode failed, %s", err)
	}
	request, err := http.NewRequest("PUT", uri, bytes.NewBuffer(reqDatas))
	if err != nil {
		return fmt.Errorf("Update IPInstance host request failed, %s", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send request failed, %s", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Update Instance got response %s from %s", response.Status, c.netsvrs[seq[0]])
	}
	resDatas, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read netservice UpdateInstance response failed, %s", err.Error())
	}
	var res types.SvcResponse
	if err := json.Unmarshal(resDatas, &res); err != nil {
		return fmt.Errorf("decode SvcResponse common response failed, %s", err)
	}
	if res.Code == 0 {
		return nil
	}
	return fmt.Errorf(res.Message)
}

func (c *client) TransferIPAttr(input *types.TranIPAttrInput) error {
	if input == nil {
		return fmt.Errorf("input can not be nil")
	}
	if len(input.Cluster) == 0 || len(input.Net) == 0 || len(input.IPList) == 0 ||
		len(input.SrcStatus) == 0 || len(input.DestStatus) == 0 {
		return fmt.Errorf("cluster, net, iplist, src, dest can not be empty")
	}
	seq := c.random.Perm(len(c.netsvrs))
	httpClient, prefix := c.getHTTPClient(3)
	uri := prefix + c.netsvrs[seq[0]] + "/v1/ipinstance/status"
	reqDatas, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("IP Instance status json encode failed, %s", err)
	}
	request, err := http.NewRequest("PUT", uri, bytes.NewBuffer(reqDatas))
	if err != nil {
		return fmt.Errorf("Transfer IPInstance status host request failed, %s", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send request failed, %s", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Update Instance got response %s from %s", response.Status, c.netsvrs[seq[0]])
	}
	resDatas, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read netservice UpdateInstance response failed, %s", err.Error())
	}
	var res types.SvcResponse
	if err := json.Unmarshal(resDatas, &res); err != nil {
		return fmt.Errorf("decode SvcResponse common response failed, %s", err)
	}
	if res.Code == 0 {
		return nil
	}
	return fmt.Errorf(res.Message)
}

// getHttpClient create http client according tlsconfig
func (c *client) getHTTPClient(timeout int) (*http.Client, string) {
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
