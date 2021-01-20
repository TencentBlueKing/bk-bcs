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
	"encoding/json"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	types "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
	"path/filepath"
	"time"
)

//AddHost update host info
func (srv *NetService) AddHost(info *types.HostInfo) error {
	started := time.Now()
	info.Created = time.Now().Format("2006-01-02 15:04:05")
	info.Update = info.Created
	data, err := json.Marshal(info)
	if err != nil {
		reportMetrics("addHost", stateJSONFailure, started)
		return err
	}
	hostpath := filepath.Join(defaultHostInfoPath, info.IPAddr)
	if exist, _ := srv.store.Exist(hostpath); exist {
		blog.Errorf("Host %s is already exist, skip ADD", info.IPAddr)
		reportMetrics("addHost", stateNonExistFailure, started)
		return fmt.Errorf("host %s is already exist", info.IPAddr)
	}
	//check pool path first
	poolpath := filepath.Join(defaultPoolInfoPath, info.Cluster, info.Pool)
	if exist, _ := srv.store.Exist(poolpath); !exist {
		blog.Errorf("Host %s creat relation to pool %s failed, pool lost", info.IPAddr, info.Pool)
		reportMetrics("addHost", stateNonExistFailure, started)
		return fmt.Errorf("pool %s lost", info.Pool)
	}
	if err := srv.store.Add(hostpath, data); err != nil {
		blog.Errorf("host %s add failed, %v", hostpath, err)
		reportMetrics("addHost", stateStorageFailure, started)
		return fmt.Errorf("host %s add failed, %s", info.IPAddr, err.Error())
	}
	//create relation to pool
	phpath := filepath.Join(poolpath, "hosts", info.IPAddr)
	if err := srv.store.Add(phpath, data); err != nil {
		blog.Errorf("Host %s create node under pool %s/%s err, %v", info.IPAddr, info.Cluster, info.Pool, err)
		reportMetrics("addHost", stateStorageFailure, started)
		return fmt.Errorf("host %s node under pool failed, %s", info.IPAddr, err.Error())
	}
	reportMetrics("addHost", stateSuccess, started)
	blog.Info("Create Host %s node under pool %s%s success.", info.IPAddr, info.Cluster, info.Pool)
	return nil
}

//DeleteHost update host info
func (srv *NetService) DeleteHost(hostIP string, ipsDel []string) error {
	started := time.Now()
	hostpath := filepath.Join(defaultHostInfoPath, hostIP)
	if exist, _ := srv.store.Exist(hostpath); !exist {
		blog.Errorf("Host %s do not exist in deletion", hostIP)
		reportMetrics("deleteHost", stateNonExistFailure, started)
		return fmt.Errorf("host %s do not exist", hostIP)
	}
	//get host info
	hostData, hostErr := srv.store.Get(hostpath)
	if hostErr != nil {
		blog.Errorf("Get Host %s detail data failed, %v", hostIP, hostErr)
		reportMetrics("deleteHost", stateStorageFailure, started)
		return fmt.Errorf("Get Host %s failed, %s", hostIP, hostErr.Error())
	}
	var hostInfo types.HostInfo
	if err := json.Unmarshal(hostData, &hostInfo); err != nil {
		blog.Errorf("Host %s json unmarshal failed, %v", hostIP, err)
		reportMetrics("deleteHost", stateJSONFailure, started)
		return fmt.Errorf("Host %s get wrong json data format, %s", hostIP, err.Error())
	}
	if hostInfo.Pool == "" {
		blog.Errorf("Host %s lost ip pool info. delete failed", hostIP)
		reportMetrics("deleteHost", stateLogicFailure, started)
		return fmt.Errorf("delete failed: Host %s lost pool info", hostIP)
	}

	//try to lock
	lockpath := filepath.Join(defaultLockerPath, hostInfo.Cluster, hostInfo.Pool)
	poolLocker, lErr := srv.store.GetLocker(lockpath)
	if lErr != nil {
		blog.Errorf("create locker %s, %v", lockpath, lErr)
		reportMetrics("deleteHost", stateLogicFailure, started)
		return fmt.Errorf("create locker %s err, %s", lockpath, lErr.Error())
	}
	defer poolLocker.Unlock()
	if err := poolLocker.Lock(); err != nil {
		blog.Errorf("try to lock pool %s/%s err, %s", hostInfo.Cluster, hostInfo.Pool, err.Error())
		reportMetrics("deleteHost", stateLogicFailure, started)
		return fmt.Errorf("lock pool %s err, %s", hostInfo.Pool, err.Error())
	}

	//try to delete ips from pool
	if len(ipsDel) != 0 {
		if err := srv.cleanIPAssignToHost(&hostInfo, ipsDel); err != nil {
			reportMetrics("deleteHost", stateLogicFailure, started)
			return err
		}
	}

	containerList, err := srv.store.List(hostpath)
	if err != nil {
		blog.Errorf("check Host %s container list err, %v", hostIP, err)
		reportMetrics("deleteHost", stateStorageFailure, started)
		return fmt.Errorf("check host %s contianer list err %s", hostIP, err.Error())
	}
	if len(containerList) != 0 {
		blog.Errorf("Host %s is stil active, %d containers are deploying in it", hostIP, len(containerList))
		reportMetrics("deleteHost", stateLogicFailure, started)
		return fmt.Errorf("host %s is in active, %d container in it", hostIP, len(containerList))
	}
	if _, err := srv.store.Delete(hostpath); err != nil {
		blog.Errorf("host %s delete self node failed: %s", hostIP, err.Error())
		reportMetrics("deleteHost", stateStorageFailure, started)
		return err
	}

	//prepare delete host node in pool
	poolHostPath := filepath.Join(defaultPoolInfoPath, hostInfo.Cluster, hostInfo.Pool, "hosts", hostIP)
	if _, err := srv.store.Delete(poolHostPath); err != nil {
		blog.Errorf("Host %s delete node under pool %s/%s err, %s", hostIP, hostInfo.Cluster, hostInfo.Pool, err.Error())
		reportMetrics("deleteHost", stateStorageFailure, started)
		return fmt.Errorf("host %s delete node under pool %s failed, %s", hostIP, hostInfo.Pool, err.Error())
	}

	//if no host in the pool, we will delete the pool automatically
	/*poolPath := filepath.Join(defaultPoolInfoPath, hostInfo.Cluster, hostInfo.Pool, "hosts")
	hostList, err := srv.store.List(poolPath)
	if err != nil {
		blog.Errorf("list net path %s err, %s", poolPath, err.Error())
		return fmt.Errorf("list net path %s err, %s", poolPath, err.Error())
	}
	if len(hostList) == 0 {
		blog.Infof("there is no host in pool %s, begin to clean pool", poolPath)
		err := srv.DeletePool(hostInfo.Cluster + "/" + hostInfo.Pool)
		if err != nil {
			blog.Errorf("DeletePool error, %s", err.Error())
			return fmt.Errorf("DeletePool error, %s", err.Error())
		}
	}*/
	reportMetrics("deleteHost", stateSuccess, started)
	return nil
}

// cleanIPAssignToHost some IP addresses are assigned to host, and can only use
// in this host. when we delete this host from netservice, we need to clean these
// IP addresses.
func (srv *NetService) cleanIPAssignToHost(hostInfo *types.HostInfo, ips []string) error {
	poolKey := hostInfo.Cluster + "/" + hostInfo.Pool
	poolInfo, err := srv.ListPoolByKey(poolKey)
	if err != nil {
		blog.Errorf("get pool info error, err %s", err.Error())
		return err
	}
	var ipsDelReserved []string
	var ipsDelAvailable []string
	for _, ip := range ips {
		foundFlag := false
		for _, reservedIP := range poolInfo.Reserved {
			if reservedIP == ip {
				ipsDelReserved = append(ipsDelReserved, reservedIP)
				foundFlag = true
				break
			}
		}
		for _, availableIP := range poolInfo.Available {
			if availableIP == ip {
				ipsDelAvailable = append(ipsDelAvailable, availableIP)
				foundFlag = true
				break
			}
		}
		for _, activeIP := range poolInfo.Active {
			if activeIP == ip {
				blog.Errorf("ip %s is active, cannot be deleted", ip)
				return fmt.Errorf("ip %s is active, cannot be deleted", ip)
			}
		}
		//dose not find in pool, error
		//TODO: deal with delete failure
		if !foundFlag {
			blog.Errorf("ip %s is not in pool %s, cannot be deleted", ip, poolKey)
			//if the ip is not found in reserved or available, just show error log, don't return
			//So user can do a second chance when ip zk node deletion failed
			//return fmt.Errorf("ip %s is not in pool %s, cannot be deleted", ip, poolKey)
		}
	}
	for _, ip := range ipsDelReserved {
		ipPath := filepath.Join(defaultPoolInfoPath, hostInfo.Cluster, hostInfo.Pool, "reserved", ip)
		if _, err := srv.store.Delete(ipPath); err != nil {
			blog.Errorf("failed delete path %s, err: %s", ipPath, err.Error())
			return fmt.Errorf("failed delete path %s, err: %s", ipPath, err.Error())
		}
		blog.Infof("delete path %s successfully", ipPath)
	}
	for _, ip := range ipsDelAvailable {
		ipPath := filepath.Join(defaultPoolInfoPath, hostInfo.Cluster, hostInfo.Pool, "available", ip)
		if _, err := srv.store.Delete(ipPath); err != nil {
			blog.Errorf("failed delete path %s, err: %s", ipPath, err.Error())
			return fmt.Errorf("failed delete path %s, err: %s", ipPath, err.Error())
		}
		blog.Infof("delete path %s successfully", ipPath)
	}
	return nil
}

//UpdateHost update host info
func (srv *NetService) UpdateHost() error {
	return nil
}

//ListHost list all host info
func (srv *NetService) ListHost() ([]*types.HostInfo, error) {
	started := time.Now()
	paths, err := srv.store.List(defaultHostInfoPath)
	if err != nil {
		blog.Errorf("list all host path err, %v", err)
		reportMetrics("listHost", stateStorageFailure, started)
		return nil, fmt.Errorf("list all path err, %s", err.Error())
	}
	var hosts []*types.HostInfo
	if len(paths) == 0 {
		blog.Info("No host is active now.")
		reportMetrics("listHost", stateSuccess, started)
		return hosts, nil
	}
	for _, p := range paths {
		host, err := srv.ListHostByKey(p)
		if err != nil {
			blog.Errorf("Get host %s info failed, %v", p, err)
			reportMetrics("listHost", stateStorageFailure, started)
			return nil, fmt.Errorf("Get host info failed, %s", err.Error())
		}
		hosts = append(hosts, host)
	}
	reportMetrics("listHost", stateSuccess, started)
	return hosts, nil
}

//ListHostByKey list host by ip
func (srv *NetService) ListHostByKey(ip string) (*types.HostInfo, error) {
	started := time.Now()
	blog.Infof("try to get host %s info.", ip)
	hostpath := filepath.Join(defaultHostInfoPath, ip)
	hostData, err := srv.store.Get(hostpath)
	if err != nil {
		blog.Errorf("Get host %s by key err, %v", ip, err)
		reportMetrics("listHostByIP", stateStorageFailure, started)
		return nil, fmt.Errorf("get host %s by key err, %s", ip, err.Error())
	}
	host := &types.HostInfo{}
	if err := json.Unmarshal(hostData, host); err != nil {
		blog.Errorf("Host %s decode json data err: %v, origin data: %s", ip, err, string(hostData))
		reportMetrics("listHostByIP", stateJSONFailure, started)
		return nil, fmt.Errorf("decode host %s data err, %s", ip, err.Error())
	}
	//get container info under host
	containerList, listErr := srv.store.List(hostpath)
	if listErr != nil {
		blog.Errorf("Host %s list container node err, %s", ip, listErr.Error())
		reportMetrics("listHostByIP", stateStorageFailure, started)
		return nil, fmt.Errorf("host %s list container err, %s", ip, listErr.Error())
	}
	host.Containers = make(map[string]*types.IPInst)
	if len(containerList) > 0 {
		for _, containerID := range containerList {
			conpath := filepath.Join(defaultHostInfoPath, ip, containerID)
			data, err := srv.store.Get(conpath)
			if err != nil {
				blog.Errorf("Get container %s err, %v", containerID, err)
				reportMetrics("listHostByIP", stateStorageFailure, started)
				return nil, fmt.Errorf("get container %s err, %s", containerID, err.Error())
			}
			inst := &types.IPInst{}
			json.Unmarshal(data, inst)
			host.Containers[containerID] = inst
			blog.Infof("get container %s under host %s success.", containerID, ip)
		}
	}
	reportMetrics("listHostByIP", stateSuccess, started)
	blog.Infof("Get host %s info success, container size %d", ip, len(host.Containers))
	return host, nil
}

//GetHostInfo get host node content info, container info are exclude
//note: if err is nil & HostInfo is nil, no hostInfo in storage
func (srv *NetService) GetHostInfo(ip string) (*types.HostInfo, error) {
	started := time.Now()
	blog.Infof("try to get host %s info.", ip)
	hostpath := filepath.Join(defaultHostInfoPath, ip)
	exist, err := srv.store.Exist(hostpath)
	if err != nil {
		blog.Errorf("Check Host %s exist err, %v", ip, err)
		reportMetrics("getHostInfo", stateStorageFailure, started)
		return nil, err
	}
	if !exist {
		blog.Warnf("Host %s do not exist", ip)
		reportMetrics("getHostInfo", stateNonExistFailure, started)
		return nil, nil
	}
	hostData, getErr := srv.store.Get(hostpath)
	if getErr != nil {
		blog.Errorf("Get host %s by key err, %v", ip, getErr)
		reportMetrics("getHostInfo", stateStorageFailure, started)
		return nil, fmt.Errorf("get host %s by key err, %s", ip, getErr.Error())
	}
	host := &types.HostInfo{}
	if err := json.Unmarshal(hostData, host); err != nil {
		blog.Errorf("Host %s decode json data err: %v, origin data: %s", ip, err, string(hostData))
		reportMetrics("getHostInfo", stateJSONFailure, started)
		return nil, fmt.Errorf("decode host %s data err, %s", ip, err.Error())
	}
	reportMetrics("getHostInfo", stateSuccess, started)
	return host, nil
}
