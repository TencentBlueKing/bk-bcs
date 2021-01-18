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

//IPLean lease ip address from net
func (srv *NetService) IPLean(lease *types.IPLease) (*types.IPInfo, error) {
	//check host info
	started := time.Now()
	blog.Info("try to get ip address for container %s in host %s", lease.Container, lease.Host)
	hostpath := filepath.Join(defaultHostInfoPath, lease.Host)
	//check pool node
	if exist, _ := srv.store.Exist(hostpath); !exist {
		blog.Errorf("host %s do not exist, no ip can be asigned for container %s", hostpath, lease.Container)
		reportMetrics("ipLean", stateNonExistFailure, started)
		return nil, fmt.Errorf("host %s do not exist", lease.Host)
	}
	hostData, gerr := srv.store.Get(hostpath)
	if gerr != nil {
		blog.Errorf("get host %s info for container %s failed, %v", lease.Host, lease.Container, gerr)
		reportMetrics("ipLean", stateStorageFailure, started)
		return nil, gerr
	}
	host := &types.HostInfo{}
	if err := json.Unmarshal(hostData, host); err != nil {
		blog.Errorf("json decode host %s data err, %s", lease.Host, err.Error())
		reportMetrics("ipLean", stateJSONFailure, started)
		return nil, err
	}
	poolpath := filepath.Join(defaultPoolInfoPath, host.Cluster, host.Pool)
	if exist, _ := srv.store.Exist(poolpath); !exist {
		blog.Errorf("get no ip resource in host %s for container %s, pool %s/%s is lost", lease.Host, lease.Container, host.Cluster, host.Pool)
		reportMetrics("ipLean", stateNonExistFailure, started)
		return nil, fmt.Errorf("pool resource %s for host %s lost", host.Pool, lease.Host)
	}
	//try to lock
	lockpath := filepath.Join(defaultLockerPath, host.Cluster, host.Pool)
	poolLocker, lErr := srv.store.GetLocker(lockpath)
	if lErr != nil {
		blog.Errorf("create locker %s for container %s in host %s err, %v", lockpath, lease.Container, lease.Host, lErr)
		reportMetrics("ipLean", stateLogicFailure, started)
		return nil, fmt.Errorf("create locker %s err, %s", lockpath, lErr.Error())
	}
	defer poolLocker.Unlock()
	if err := poolLocker.Lock(); err != nil {
		blog.Errorf("try to lock pool %s/%s err, %s", host.Cluster, host.Pool, err.Error())
		reportMetrics("ipLean", stateStorageFailure, started)
		return nil, fmt.Errorf("lock pool %s err, %s", host.Pool, err.Error())
	}
	blog.Info("lock pool %s success in ip lease for container %s in host %s", lockpath, lease.Container, lease.Host)
	//construct ip address path
	var ippath, destpath, lastStatus string
	if len(lease.IPAddr) == 0 {
		//random select from available ip node
		blog.Info("random select ip address for container %s in host %s", lease.Container, lease.Host)
		pNode := filepath.Join(poolpath, "available")
		allNodes, err := srv.store.List(pNode)
		if len(allNodes) == 0 {
			blog.Errorf("no available ip address for container %s in host %s", lease.Container, lease.Host)
			reportMetrics("ipLean", stateLogicFailure, started)
			return nil, fmt.Errorf("no available ip resource in pool %s", host.Pool)
		}
		if err != nil {
			blog.Errorf("random select addr for container %s in host %s err, %v", lease.Container, lease.Host, err)
			reportMetrics("ipLean", stateStorageFailure, started)
			return nil, fmt.Errorf("random select ip in pool %s failed, %s", host.Pool, err.Error())
		}
		ippath = filepath.Join(poolpath, "available", allNodes[0])
		destpath = filepath.Join(poolpath, "active", allNodes[0])
		lastStatus = types.IPStatus_AVAILABLE
	} else {
		blog.Info("select reserved ip %s for container %s in host %s", lease.IPAddr, lease.Container, lease.Host)
		ippath = filepath.Join(poolpath, "reserved", lease.IPAddr)
		destpath = filepath.Join(poolpath, "active", lease.IPAddr)
		lastStatus = types.IPStatus_RESERVED
	}
	//get ip resource data, move node
	ipData, ipErr := srv.store.Get(ippath)
	if ipErr != nil {
		blog.Errorf("get %s data err, %v", ippath, ipErr)
		reportMetrics("ipLean", stateStorageFailure, started)
		return nil, fmt.Errorf("get ip %s data failed, %s", ippath, ipErr.Error())
	}
	ipInst := &types.IPInst{}
	if err := json.Unmarshal(ipData, ipInst); err != nil {
		blog.Errorf("Fatal error in ip lease, %s data decode json err, %v", ippath, err)
		reportMetrics("ipLean", stateJSONFailure, started)
		return nil, fmt.Errorf("ip %s data json decode err, %s", ippath, err.Error())
	}
	ipInst.LastStatus = lastStatus
	ipInst.Host = lease.Host
	ipInst.Container = lease.Container
	ipInst.Status = types.IPStatus_ACTIVE
	ipInst.Update = time.Now().Format("2006-01-02 15:04:05")
	ipInst.Cluster = host.Cluster
	if lease.App != "" {
		ipInst.App = lease.App
	}
	//delete old node
	if _, err := srv.store.Delete(ippath); err != nil {
		blog.Errorf("clean ip %s for container %s in host %s err, %v", ippath, lease.Container, lease.Host, err)
		reportMetrics("ipLean", stateStorageFailure, started)
		return nil, fmt.Errorf("lease ip %s resource err, %s", filepath.Base(ippath), err.Error())
	}
	data, _ := json.Marshal(ipInst)
	if err := srv.store.Add(destpath, data); err != nil {
		//todo(DeveloperJim): recover ip resource when error
		blog.Errorf("move ip %s to active err, %v", ippath, err)
		reportMetrics("ipLean", stateStorageFailure, started)
		return nil, fmt.Errorf("active ip %s resource err, %s", filepath.Base(ippath), err.Error())
	}
	containerpath := filepath.Join(defaultHostInfoPath, lease.Host, lease.Container)
	if err := srv.store.Add(containerpath, data); err != nil {
		//todo(DeveloperJim): recover ip resource when error
		blog.Errorf("create container query node %s err, %v", containerpath, err)
		reportMetrics("ipLean", stateStorageFailure, started)
		return nil, fmt.Errorf("active ip %s in host %s err, %s", filepath.Base(ippath), lease.Host, err.Error())
	}
	ipInfo := &types.IPInfo{
		IPAddr:  ipInst.IPAddr,
		MacAddr: ipInst.MacAddr,
		Pool:    ipInst.Pool,
		Mask:    ipInst.Mask,
		Gateway: ipInst.Gateway,
	}
	blog.Infof("asigned ip %s for container %s in host %s success.", ippath, lease.Container, lease.Host)
	reportMetrics("ipLean", stateSuccess, started)
	return ipInfo, nil
}

//IPRelease release ip resource to net pool
func (srv *NetService) IPRelease(release *types.IPRelease) error {
	started := time.Now()
	//check container path
	blog.Info("try to release ip resource in container %s ", release.Container)
	containerpath := filepath.Join(defaultHostInfoPath, release.Host, release.Container)
	if exist, _ := srv.store.Exist(containerpath); !exist {
		blog.Errorf("No ip release for container %s in host %s", release.Container, release.Host)
		reportMetrics("ipRelease", stateNonExistFailure, started)
		return fmt.Errorf("container %s do not exist in host %s", release.Container, release.Host)
	}
	ipInst := &types.IPInst{}
	data, err := srv.store.Get(containerpath)
	if err != nil {
		blog.Errorf("Release conatainer %s resource in host %s err, %v", release.Container, release.Host, err)
		reportMetrics("ipRelease", stateStorageFailure, started)
		return fmt.Errorf("verify ip resource for container %s in host %s err, %s", release.Container, release.Host, err.Error())
	}
	json.Unmarshal(data, ipInst)
	//check active path
	ippath := filepath.Join(defaultPoolInfoPath, ipInst.Cluster, ipInst.Pool, "active", ipInst.IPAddr)
	if exist, _ := srv.store.Exist(ippath); !exist {
		blog.Errorf("ip resource %s lost for container %s in host %s", ipInst.IPAddr, release.Container, release.Host)
		reportMetrics("ipRelease", stateNonExistFailure, started)
		return fmt.Errorf("ip resource %s lost for container %s", ipInst.IPAddr, release.Container)
	}
	destpath := filepath.Join(defaultPoolInfoPath, ipInst.Cluster, ipInst.Pool, ipInst.LastStatus, ipInst.IPAddr)
	//move to reserved path/available path
	ipInst.Status = ipInst.LastStatus
	ipInst.LastStatus = types.IPStatus_ACTIVE
	ipInst.Update = time.Now().Format("2006-01-02 15:04:05")
	ipdata, _ := json.Marshal(ipInst)
	if _, err := srv.store.Delete(containerpath); err != nil {
		blog.Errorf("clean container %s err, %v", containerpath, err)
		reportMetrics("ipRelease", stateStorageFailure, started)
		return fmt.Errorf("clean container %s err, %s", release.Container, err.Error())
	}
	if _, err := srv.store.Delete(ippath); err != nil {
		//todo(DeveloperJim): delete force
		blog.Errorf("clean ip resource %s err, %v", ippath, err)
		reportMetrics("ipRelease", stateStorageFailure, started)
		return fmt.Errorf("clean ip resource %s err, %s", ippath, err.Error())
	}
	if err := srv.store.Add(destpath, ipdata); err != nil {
		blog.Errorf("return ip %s in container %s err, %v", destpath, release.Container, err)
		reportMetrics("ipRelease", stateStorageFailure, started)
		return fmt.Errorf("release ip %s err, %s", ipInst.IPAddr, err.Error())
	}
	blog.Infof("Release ip %s/%s for container %s in host %s success", ipInst.Cluster, ipInst.IPAddr, release.Container, release.Host)
	reportMetrics("ipRelease", stateSuccess, started)
	return nil
}

//HostVIPRelease to release host's vip
func (srv *NetService) HostVIPRelease(hostIP string) error {
	started := time.Now()
	blog.Info("try to release ip resource in host %s ", hostIP)
	hostPath := filepath.Join(defaultHostInfoPath, hostIP)
	if exist, _ := srv.store.Exist(hostPath); !exist {
		blog.Errorf("No host node %s zookeeper", hostIP)
		reportMetrics("hostResourceRelease", stateNonExistFailure, started)
		return fmt.Errorf("No host node %s zookeeper", hostIP)
	}
	containerIDs, err := srv.store.List(hostPath)
	if err != nil {
		blog.Errorf("list host %s containerids err, %v", hostIP, err)
		reportMetrics("hostResourceRelease", stateStorageFailure, started)
		return err
	}
	for _, item := range containerIDs {
		ipRel := &types.IPRelease{
			Host:      hostIP,
			Container: item,
		}

		if err := srv.IPRelease(ipRel); err != nil {
			reportMetrics("hostResourceRelease", stateLogicFailure, started)
			blog.Errorf("IPRelease container %s in host %s,failed:%v", item, hostIP, err)
			return err
		}
	}
	reportMetrics("hostResourceRelease", stateSuccess, started)
	blog.Info("release ip resource in host %s success", hostIP)
	return nil
}
