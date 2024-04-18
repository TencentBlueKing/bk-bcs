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

package bcsapi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	bcstypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
	blog "k8s.io/klog/v2"

	types "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
)

const (
	netserviceDiscoveryPath = bcstypes.BCS_SERV_BASEPATH + "/" + bcstypes.BCS_MODULE_NETSERVICE
	netserviceVersion       = "v1"
	envVarNameNetservice    = "NETSVR_ADDR"
)

// Netservice define http client for bcs-netservice. The interface used to operate
// netservice with the crud of pool.
type Netservice interface {
	RegisterPool(pool *types.NetPool) error
	UpdatePool(pool *types.NetPool) error
	GetPool(cluster, net string) ([]*types.NetPool, error)
	DeletePool(cluster, net string) error
	ListAllPool() ([]*types.NetPool, error)
	ListAllPoolWithCluster(cluster string) ([]*types.NetPool, error)
	RegisterHost(host *types.HostInfo) error
	DeleteHost(host string, ips []string) error
	GetHostInfo(host string, timeout int) (*types.HostInfo, error)
	LeaseIPAddr(lease *types.IPLease, timeout int) (*types.IPInfo, error)
	ReleaseIPAddr(release *types.IPRelease, ipInfo *types.IPInfo, timeout int) error
	UpdateIPInstance(inst *types.IPInst) error
	TransferIPAttr(input *types.TranIPAttrInput) error
}

// NetserviceCli netservice http client, will handle all operations with netservice
// It can be used as the cli to handle every operation of netservice
type NetserviceCli struct {
	httpClientTimeout int
	tlsConfig         *tls.Config
	// netservice addresses
	netSvrs []string
	random  *rand.Rand
}

// NewNetserviceCli create new client for netservice cli
func NewNetserviceCli() *NetserviceCli {
	return &NetserviceCli{
		httpClientTimeout: 3,
		random:            rand.New(rand.NewSource(time.Now().UnixNano())), // nolint
	}
}

// NewNetserviceCliWithTimeout create new client with timeout
func NewNetserviceCliWithTimeout(timeoutSeconds int) *NetserviceCli {
	return &NetserviceCli{
		httpClientTimeout: timeoutSeconds,
		random:            rand.New(rand.NewSource(time.Now().UnixNano())), // nolint
	}
}

// SetCerts set tls client with root CA, private key, public key and password
func (nc *NetserviceCli) SetCerts(ca, key, crt, passwd string) error {
	config, err := ssl.ClientTslConfVerity(ca, crt, key, passwd)
	if err != nil {
		return err
	}
	nc.tlsConfig = config
	return nil
}

// SetTLS set tls config
func (nc *NetserviceCli) SetTLS(config *tls.Config) {
	nc.tlsConfig = config
}

// SetHosts set netservice server addresses
func (nc *NetserviceCli) SetHosts(svrs []string) {
	nc.netSvrs = svrs
}

// GetNetService get netservice server addresses from zookeeper or from envs. It
// will return error if netservice not exist.
func (nc *NetserviceCli) GetNetService(zkHost []string) error {
	// get netservice addresses from env
	netSvrStr := os.Getenv(envVarNameNetservice)
	if len(netSvrStr) != 0 {
		netSvrs := strings.Split(
			strings.ReplaceAll(
				strings.TrimSpace(netSvrStr), ";", ","), ",")
		nc.netSvrs = append(nc.netSvrs, netSvrs...)
		return nil
	}
	// get netservice address from zookeeper
	zkCli := zkclient.NewZkClientWithoutLogger(zkHost)
	if err := zkCli.ConnectEx(time.Second * 5); err != nil {
		return fmt.Errorf("zk connnect failed, zk %+v, err %s", zkHost, err.Error())
	}
	rSvrList, stat, err := zkCli.GetChildrenEx(netserviceDiscoveryPath)
	if err != nil {
		return fmt.Errorf("get zookeeper path children failed, path %s, err %s", netserviceDiscoveryPath, err.Error())
	} else if stat == nil {
		return fmt.Errorf("zookeeper status lost, path %s", netserviceDiscoveryPath)
	}
	if len(rSvrList) == 0 {
		return fmt.Errorf("no bcs-netservice server node detected")
	}
	// get all data
	for _, node := range rSvrList {
		child := filepath.Join(netserviceDiscoveryPath, node)
		data, _, err := zkCli.GetEx(child)
		if err != nil {
			blog.Warningf("get zookeeper data failed, path %s, err %s", child, err.Error())
			continue
		}
		info := new(bcstypes.NetServiceInfo)
		if err := json.Unmarshal(data, info); err != nil {
			blog.Warningf("unmarshal zookeeper data failed, data %s, err %s", string(data), err.Error())
			continue
		}
		svrInfo := info.IP + ":" + strconv.Itoa(int(info.Port))
		nc.netSvrs = append(nc.netSvrs, svrInfo)
	}
	if len(nc.netSvrs) == 0 {
		return fmt.Errorf("get no netservice node, raw svr list %+v", rSvrList)
	}
	return nil
}

// RegisterPool register pool info to bcs-netservice, will register pool info to netservice
// The pool inf will be saved in store.
func (nc *NetserviceCli) RegisterPool(pool *types.NetPool) error {
	if len(pool.Cluster) == 0 {
		return fmt.Errorf("lost cluster info")
	}
	if len(pool.Available) == 0 && len(pool.Reserved) == 0 {
		return fmt.Errorf("lost ip address info")
	}
	netRequest := &types.NetRequest{
		Type: types.RequestType_POOL,
		Pool: pool,
	}
	reqDatas, err := json.Marshal(netRequest)
	if err != nil {
		return fmt.Errorf("register pool encode failed, err %s", err.Error())
	}
	seq := nc.random.Perm(len(nc.netSvrs))
	httpClient, prefix := nc.getHTTPClient(nc.httpClientTimeout)
	uri := prefix + nc.netSvrs[seq[0]] + "/v1/pool"
	request, err := http.NewRequest("POST", uri, bytes.NewBuffer(reqDatas))
	if err != nil {
		return fmt.Errorf("register pool create http request failed, %s", err.Error())
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("register pool send request failed, %s", err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("register pool Got err response: %s", response.Status)
	}
	defer response.Body.Close()
	resDatas, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("register pool read response body failed, %s", err)
	}
	var netRes types.NetResponse
	if err := json.Unmarshal(resDatas, &netRes); err != nil {
		return fmt.Errorf("register pool decode response failed, %s", err)
	}
	if netRes.Code != 0 {
		return fmt.Errorf("register pool failed, %s", netRes.Message)
	}
	return nil
}

// UpdatePool update pool info, will update pool information to netservice
// will update poll info to store.
func (nc *NetserviceCli) UpdatePool(pool *types.NetPool) error {
	if len(pool.Cluster) == 0 {
		return fmt.Errorf("lost cluster info")
	}
	if len(pool.Available) == 0 && len(pool.Reserved) == 0 {
		return fmt.Errorf("lost ip address info")
	}
	// create net request
	netRequest := &types.NetRequest{
		Type: types.RequestType_POOL,
		Pool: pool,
	}
	reqDatas, err := json.Marshal(netRequest)
	if err != nil {
		return fmt.Errorf("update pool encode failed, %s", err)
	}
	blog.Info("udpate pool string:%s", reqDatas)
	seq := nc.random.Perm(len(nc.netSvrs))
	httpClient, prefix := nc.getHTTPClient(nc.httpClientTimeout)
	uri := prefix + nc.netSvrs[seq[0]] + "/v1/pool/" + pool.Cluster + "/" + pool.Net // nolint
	request, err := http.NewRequest("PUT", uri, bytes.NewBuffer(reqDatas))
	if err != nil {
		return fmt.Errorf("update pool create http request failed, %s", err.Error())
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("udpate pool send request failed, %s", err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("udpate pool Got err response: %s", response.Status)
	}
	defer response.Body.Close()
	resDatas, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("udpate pool read response body failed, %s", err)
	}
	var netRes types.NetResponse
	if err := json.Unmarshal(resDatas, &netRes); err != nil {
		return fmt.Errorf("udpate pool decode response failed, %s", err)
	}
	if netRes.Code != 0 {
		return fmt.Errorf("udpate pool failed, %s", netRes.Message)
	}
	return nil
}

// GetPool get pool info from netservice, will return pool information from netservice
func (nc *NetserviceCli) GetPool(cluster, net string) ([]*types.NetPool, error) {
	if len(cluster) == 0 || len(net) == 0 {
		return nil, fmt.Errorf("Lost cluster or network segment in request")
	}
	if len(nc.netSvrs) == 0 {
		return nil, fmt.Errorf("no available bcs-netservice")
	}
	seq := nc.random.Perm(len(nc.netSvrs))
	httpClient, prefix := nc.getHTTPClient(nc.httpClientTimeout)
	uri := prefix + nc.netSvrs[seq[0]] + "/v1/pool/" + cluster + "/" + net
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("get pool create http request failed, %s", err.Error())
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("get pool http client send request failed, %s", err.Error())
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get pool http response %s from %s", response.Status, nc.netSvrs[seq[0]])
	}
	defer response.Body.Close()
	datas, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("get pool read netservice response failed, %s", err.Error())
	}
	netRes := &types.NetResponse{}
	if err := json.Unmarshal(datas, netRes); err != nil {
		return nil, fmt.Errorf("get pool decode netservice response failed, %s", err.Error())
	}
	if netRes.Type != types.ResponseType_POOL {
		return nil, fmt.Errorf("get pool response data type expect %d, but got %d",
			types.ResponseType_HOST, netRes.Type)
	}
	if !netRes.IsSucc() {
		return nil, fmt.Errorf("get pool request fialed: %s", netRes.Message)
	}
	if len(netRes.Pool) == 0 {
		return nil, nil
	}
	return netRes.Pool, nil
}

// ListAllPool list all pools, it will return all pools from netservice
func (nc *NetserviceCli) ListAllPool() ([]*types.NetPool, error) {
	if len(nc.netSvrs) == 0 {
		return nil, fmt.Errorf("no available bcs-netservice")
	}
	seq := nc.random.Perm(len(nc.netSvrs))
	httpClient, prefix := nc.getHTTPClient(nc.httpClientTimeout)
	uri := prefix + nc.netSvrs[seq[0]] + "/v1/pool"
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("create http request failed, %s", err.Error())
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("list all pool http client send request failed, %s", err.Error())
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list all pool http response %s from %s", response.Status, nc.netSvrs[seq[0]])
	}
	defer response.Body.Close()
	datas, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("list all pool read netservice response failed, %s", err.Error())
	}
	netRes := &types.NetResponse{}
	if err := json.Unmarshal(datas, netRes); err != nil {
		return nil, fmt.Errorf("list all pool decode netservice response failed, %s", err.Error())
	}
	if netRes.Type != types.ResponseType_POOL {
		return nil, fmt.Errorf("list all pool response data type expect %d, but got %d",
			types.ResponseType_HOST, netRes.Type)
	}
	if !netRes.IsSucc() {
		return nil, fmt.Errorf("list all pool pool request fialed: %s", netRes.Message)
	}
	if len(netRes.Pool) == 0 {
		return nil, nil
	}
	return netRes.Pool, nil
}

// ListAllPoolWithCluster list all pool with cluster, it will list all pools by clusterid
func (nc *NetserviceCli) ListAllPoolWithCluster(cluster string) ([]*types.NetPool, error) {
	if len(nc.netSvrs) == 0 {
		return nil, fmt.Errorf("no available bcs-netservice")
	}
	seq := nc.random.Perm(len(nc.netSvrs))
	httpClient, prefix := nc.getHTTPClient(nc.httpClientTimeout)
	uri := prefix + nc.netSvrs[seq[0]] + "/v1/pool/" + cluster + "?info=detail"
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("list all pool with cluster create http request failed, %s", err.Error())
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("list all pool with cluster http client send request failed, %s", err.Error())
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list all pool with cluster http response %s from %s",
			response.Status, nc.netSvrs[seq[0]])
	}
	defer response.Body.Close()
	datas, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("list all pool with cluster read netservice response failed, %s", err.Error())
	}
	netRes := &types.NetResponse{}
	if err := json.Unmarshal(datas, netRes); err != nil {
		return nil, fmt.Errorf("list all pool with cluster decode netservice response failed, %s", err.Error())
	}
	if netRes.Type != types.ResponseType_POOL {
		return nil, fmt.Errorf("list all pool with cluster response data type expect %d, but got %d",
			types.ResponseType_HOST, netRes.Type)
	}
	if !netRes.IsSucc() {
		return nil, fmt.Errorf("list all pool with cluster pool request fialed: %s", netRes.Message)
	}
	if len(netRes.Pool) == 0 {
		return nil, nil
	}
	return netRes.Pool, nil
}

// DeletePool delete pool, it will delete pool by cluster and net info
func (nc *NetserviceCli) DeletePool(cluster, net string) error {
	if len(cluster) == 0 || len(net) == 0 {
		return fmt.Errorf("neither cluster nor net can be empty")
	}
	seq := nc.random.Perm(len(nc.netSvrs))
	httpClient, prefix := nc.getHTTPClient(250)
	uri := prefix + nc.netSvrs[seq[0]] + "/v1/pool/" + cluster + "/" + net
	request, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return fmt.Errorf("delete pool create http request failed, %s", err.Error())
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("delete pool http client send request failed, %s", err.Error())
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("delete pool http response %s from %s", response.Status, nc.netSvrs[seq[0]])
	}
	defer response.Body.Close()
	datas, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("delete pool read netservice response failed, %s", err.Error())
	}
	netRes := &types.NetResponse{}
	if err := json.Unmarshal(datas, netRes); err != nil {
		return fmt.Errorf("delete pool decode netservice response failed, %s", err.Error())
	}
	if netRes.Code != 0 {
		return fmt.Errorf("delete pool netservice response code not zero, response data %s", string(datas))
	}
	return nil
}

// RegisterHost register host info to bcs-netservice
// It will register host to netservice
func (nc *NetserviceCli) RegisterHost(host *types.HostInfo) error {
	if len(nc.netSvrs) == 0 {
		return fmt.Errorf("no available bcs-netservice")
	}
	if len(host.Pool) == 0 || len(host.Cluster) == 0 {
		return fmt.Errorf("Host info is invalid")
	}
	seq := nc.random.Perm(len(nc.netSvrs))
	httpClient, prefix := nc.getHTTPClient(nc.httpClientTimeout)
	uri := prefix + nc.netSvrs[seq[0]] + "/v1/host"
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
		return fmt.Errorf("register host got response %s from %s", response.Status, nc.netSvrs[seq[0]])
	}
	resDatas, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("register host  read netservice response failed, %s", err.Error())
	}
	var hostRes types.NetResponse
	if err := json.Unmarshal(resDatas, &hostRes); err != nil {
		return fmt.Errorf("register host decode response failed, %s", err)
	}
	if hostRes.Code == 0 {
		return nil
	}
	return fmt.Errorf(hostRes.Message)
}

// DeleteHost when host has container or any ip belongs to the host is active, it can't be deleted
// It will delete hosts from netservice
func (nc *NetserviceCli) DeleteHost(host string, ips []string) error {
	if len(nc.netSvrs) == 0 {
		return fmt.Errorf("no available bcs-netservice")
	}
	if len(host) == 0 {
		return fmt.Errorf("bad request host ip")
	}
	seq := nc.random.Perm(len(nc.netSvrs))
	httpClient, prefix := nc.getHTTPClient(nc.httpClientTimeout)
	uri := prefix + nc.netSvrs[seq[0]] + "/v1/host/" + host

	// create net request
	netRequest := &types.NetRequest{
		Type: types.RequestType_HOST,
		IPs:  ips,
	}
	reqDatas, err := json.Marshal(netRequest)
	if err != nil {
		return fmt.Errorf("delete host encode failed, %s", err)
	}

	request, err := http.NewRequest("DELETE", uri, bytes.NewBuffer(reqDatas))

	if err != nil {
		return fmt.Errorf("delete host request failed, %s", err.Error())
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("delete host send request to net service failed, %s", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("delete host got response %s from %s", response.Status, nc.netSvrs[seq[0]])
	}
	resDatas, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("delete host read netservice response failed, %s", err.Error())
	}
	var hostRes types.NetResponse
	if err := json.Unmarshal(resDatas, &hostRes); err != nil {
		return fmt.Errorf("delete host decode response failed, %s", err)
	}
	if hostRes.Code == 0 {
		return nil
	}
	return fmt.Errorf(hostRes.Message)
}

// GetHostInfo Get host info by host ip address. It will get host info from netservice
func (nc *NetserviceCli) GetHostInfo(host string, timeout int) (*types.HostInfo, error) {
	if len(host) == 0 {
		return nil, fmt.Errorf("host ip address lost")
	}
	if timeout < 1 {
		return nil, fmt.Errorf("timeout must > 1")
	}
	seq := nc.random.Perm(len(nc.netSvrs))
	httpClient, prefix := nc.getHTTPClient(timeout)
	var lastErr string
	for _, index := range seq {
		request, reqErr := http.NewRequest("GET", prefix+nc.netSvrs[index]+"/v1/host/"+host, nil)
		if reqErr != nil {
			return nil, fmt.Errorf("get host info create request failed, %s", reqErr.Error())
		}
		request.Header.Set("Accept", "application/json")
		response, err := httpClient.Do(request)
		if err != nil {
			lastErr = err.Error()
			continue
		}
		if response.StatusCode != http.StatusOK {
			lastErr = fmt.Sprintf("get host info netService %s response code: %d", nc.netSvrs[index], response.StatusCode)
			continue
		}
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		netRes := &types.NetResponse{}
		if err := json.Unmarshal(data, netRes); err != nil {
			return nil, err
		}
		if netRes.Type != types.ResponseType_HOST {
			return nil, fmt.Errorf("get host info response data type expect %d, but got %d",
				types.ResponseType_HOST, netRes.Type)
		}
		if !netRes.IsSucc() {
			return nil, fmt.Errorf("get host info request fialed: %s", netRes.Message)
		}
		if len(netRes.Host) == 0 {
			return nil, fmt.Errorf("get host info response err, host info lost")
		}
		hostInfo := netRes.Host[0]
		if hostInfo.IPAddr != host {
			return nil, fmt.Errorf("get host info response ip address expect %s, but got %s", host, hostInfo.IPAddr)
		}
		return hostInfo, nil
	}
	return nil, fmt.Errorf("get host info all netservice failed, %s", lastErr)
}

// LeaseIPAddr lease one ip address from bcs-netservice. It will lease ip address from netservice
func (nc *NetserviceCli) LeaseIPAddr(lease *types.IPLease, timeout int) (*types.IPInfo, error) {
	// create net request
	request := &types.NetRequest{
		Type:  types.RequestType_LEASE,
		Lease: lease,
	}
	if len(lease.Host) == 0 || len(lease.Container) == 0 {
		return nil, fmt.Errorf("lease ip addr host or container lost")
	}
	if len(nc.netSvrs) == 0 {
		return nil, fmt.Errorf("lease ip addr get no online netservice")
	}
	seq := nc.random.Perm(len(nc.netSvrs))
	httpClient, prefix := nc.getHTTPClient(timeout)
	requestData, jsonErr := json.Marshal(request)
	if jsonErr != nil {
		return nil, jsonErr
	}
	var lastErr string
	for _, index := range seq {
		httpRequest, reqErr := http.NewRequest("POST",
			prefix+nc.netSvrs[index]+"/v1/allocator",
			bytes.NewBuffer(requestData))
		if reqErr != nil {
			return nil, fmt.Errorf("lease ip addr create request failed, %s", reqErr.Error())
		}
		httpRequest.Header.Set("Content-Type", "application/json")
		httpResponse, err := httpClient.Do(httpRequest)
		if err != nil {
			// do err, try next one
			lastErr = err.Error()
			continue
		}
		if httpResponse.StatusCode != http.StatusOK {
			lastErr = fmt.Sprintf("lease ip addr http response code: %d", httpResponse.StatusCode)
			continue
		}
		defer httpResponse.Body.Close()
		data, err := io.ReadAll(httpResponse.Body)
		if err != nil {
			return nil, err
		}
		response := &types.NetResponse{}
		if err := json.Unmarshal(data, response); err != nil {
			return nil, err
		}
		if response.Type != types.ResponseType_LEASE {
			return nil, fmt.Errorf("lease ip addr response data type expect %d, but got %d",
				types.ResponseType_LEASE, response.Type)
		}
		if !response.IsSucc() {
			return nil, fmt.Errorf("lease ip addr request fialed: %s", response.Message)
		}
		if response.Info == nil || len(response.Info) == 0 {
			return nil, fmt.Errorf("lease ip addr response err, ip info lost")
		}
		ipInfo := response.Info[0]
		// check if response ip addr is what we need
		if lease.IPAddr != "" && lease.IPAddr != ipInfo.IPAddr {
			return nil, fmt.Errorf("lease ip addr expect ipaddr %s, but got %s", lease.IPAddr, ipInfo.IPAddr)
		}
		if len(ipInfo.Gateway) == 0 || ipInfo.Mask == 0 {
			return nil, fmt.Errorf("lease ip addr ip lease failed, gateway/mask info lost")
		}
		return ipInfo, nil
	}
	return nil, fmt.Errorf("lease ip addr all netservice failed, %s", lastErr)
}

// ReleaseIPAddr release ip address to bcs-netservice. It will release the ip address
// from netservice
func (nc *NetserviceCli) ReleaseIPAddr(release *types.IPRelease, ipInfo *types.IPInfo, timeout int) error {
	// create net request
	request := &types.NetRequest{
		Type:    types.RequestType_RELEASE,
		Release: release,
	}
	if len(release.Host) == 0 || len(release.Container) == 0 {
		return fmt.Errorf("host or container lost")
	}
	if len(nc.netSvrs) == 0 {
		return fmt.Errorf("get no online netservice")
	}
	seq := nc.random.Perm(len(nc.netSvrs))
	httpClient, prefix := nc.getHTTPClient(timeout)
	requestData, jsonErr := json.Marshal(request)
	if jsonErr != nil {
		return jsonErr
	}
	var lastErr string
	for _, index := range seq {
		httpRequest, reqErr := http.NewRequest("DELETE",
			prefix+nc.netSvrs[index]+"/v1/allocator",
			bytes.NewBuffer(requestData))
		if reqErr != nil {
			return fmt.Errorf("release ip create request failed, %s", reqErr.Error())
		}
		httpRequest.Header.Set("Content-Type", "application/json")
		httpResponse, err := httpClient.Do(httpRequest)
		if err != nil {
			// do err, try next one
			lastErr = err.Error()
			continue
		}
		if httpResponse.StatusCode != http.StatusOK {
			lastErr = fmt.Sprintf("release ip http response code: %d", httpResponse.StatusCode)
			continue
		}
		defer httpResponse.Body.Close()
		data, err := io.ReadAll(httpResponse.Body)
		if err != nil {
			return err
		}
		response := &types.NetResponse{}
		if err := json.Unmarshal(data, response); err != nil {
			return err
		}
		if response.Type != types.ResponseType_RELEASE {
			return fmt.Errorf("release ip response data type expect %d, but got %d",
				types.ResponseType_RELEASE, response.Type)
		}
		if !response.IsSucc() {
			return fmt.Errorf("release ip request fialed: %s", response.Message)
		}
		return nil
	}
	return fmt.Errorf("all netservice failed, %s", lastErr)
}

// UpdateIPInstance update ip instance info, especially MacAddress
func (nc *NetserviceCli) UpdateIPInstance(inst *types.IPInst) error {
	if inst == nil {
		return fmt.Errorf("Lost instance data")
	}
	if inst.Cluster == "" || inst.IPAddr == "" || inst.Pool == "" {
		return fmt.Errorf("ip instance lost key data")
	}
	seq := nc.random.Perm(len(nc.netSvrs))
	httpClient, prefix := nc.getHTTPClient(nc.httpClientTimeout)
	uri := prefix + nc.netSvrs[seq[0]] + "/v1/ipinstance"
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
		return fmt.Errorf("update instance got response %s from %s", response.Status, nc.netSvrs[seq[0]])
	}
	resDatas, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("update instance read netservice response failed, %s", err.Error())
	}
	var res types.SvcResponse
	if err := json.Unmarshal(resDatas, &res); err != nil {
		return fmt.Errorf("update instance decode common response failed, %s", err)
	}
	if res.Code == 0 {
		return nil
	}
	return fmt.Errorf(res.Message)
}

// TransferIPAttr transfer ip attribution. It will transfer the ip status for source to target status
func (nc *NetserviceCli) TransferIPAttr(input *types.TranIPAttrInput) error {
	if input == nil {
		return fmt.Errorf("input can not be nil")
	}
	if len(input.Cluster) == 0 || len(input.Net) == 0 || len(input.IPList) == 0 ||
		len(input.SrcStatus) == 0 || len(input.DestStatus) == 0 {
		return fmt.Errorf("cluster, net, iplist, src, dest can not be empty")
	}
	seq := nc.random.Perm(len(nc.netSvrs))
	httpClient, prefix := nc.getHTTPClient(nc.httpClientTimeout)
	uri := prefix + nc.netSvrs[seq[0]] + "/v1/ipinstance/status"
	reqDatas, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("tranfer ip attr json encode failed, %s", err)
	}
	request, err := http.NewRequest("PUT", uri, bytes.NewBuffer(reqDatas))
	if err != nil {
		return fmt.Errorf("tranfer ip attr host request failed, %s", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("tranfer ip attr send request failed, %s", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("tranfer ip attr got response %s from %s", response.Status, nc.netSvrs[seq[0]])
	}
	resDatas, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("tranfer ip attr read netservice response failed, %s", err.Error())
	}
	var res types.SvcResponse
	if err := json.Unmarshal(resDatas, &res); err != nil {
		return fmt.Errorf("tranfer ip attr decode common response failed, %s", err)
	}
	if res.Code == 0 {
		return nil
	}
	return fmt.Errorf(res.Message)
}

// getHTTPClient create http client according tls config
func (nc *NetserviceCli) getHTTPClient(timeout int) (*http.Client, string) {
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
	if nc.tlsConfig != nil {
		prefix = "https://"
		transport.TLSClientConfig = nc.tlsConfig
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   time.Second * time.Duration(int64(timeout)),
	}
	return httpClient, prefix
}
