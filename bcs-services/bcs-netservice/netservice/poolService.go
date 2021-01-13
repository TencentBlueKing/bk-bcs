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
	"net"
	"path/filepath"
	"strconv"
	"time"
)

func getAllHosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	var ips []string
	for checkip := ip.Mask(ipnet.Mask); ipnet.Contains(checkip); inc(checkip) {
		ips = append(ips, checkip.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func stringInSlice(str string, strs []string) bool {
	for _, item := range strs {
		if str == item {
			return true
		}
	}
	return false
}

//AddPool construct pool info
func (srv *NetService) AddPool(pool *types.NetPool) error {
	started := time.Now()
	//check & construct data
	pool.Created = time.Now().Format("2006-01-02 15:04:05")
	pool.Update = pool.Created
	poolPath := filepath.Join(defaultPoolInfoPath, pool.GetKey())
	blog.Info("try to Add pool %s", poolPath)
	//creat netpool node with pool data
	poolData, jerr := json.Marshal(pool)
	if jerr != nil {
		blog.Errorf("Pool %s json err: %s", pool.GetKey(), jerr.Error())
		reportMetrics("addpool", stateJSONFailure, started)
		return jerr
	}
	if exist, _ := srv.store.Exist(poolPath); exist {
		blog.Errorf("Pool %s is exist, skip create", pool.GetKey())
		reportMetrics("addpool", stateNonExistFailure, started)
		return fmt.Errorf("pool %s exist, create failed", pool.GetKey())
	}
	if err := srv.initPoolPath(pool, poolPath, poolData); err != nil {
		reportMetrics("addpool", stateLogicFailure, started)
		return err
	}
	for _, ip := range pool.Reserved {
		inst := &types.IPInst{
			IPAddr:     ip,
			Pool:       pool.Net,
			Mask:       pool.Mask,
			Gateway:    pool.Gateway,
			LastStatus: "created",
			Status:     types.IPStatus_RESERVED,
			Update:     time.Now().Format("2006-01-02 15:04:05"),
			Container:  "",
			Host:       "",
		}
		instData, _ := json.Marshal(inst)
		instPath := filepath.Join(poolPath, "reserved", inst.GetKey())
		if err := srv.store.Add(instPath, instData); err != nil {
			blog.Errorf("Pool create reserved node %s err: %v", inst.GetKey(), err)
			reportMetrics("addpool", stateLogicFailure, started)
			return err
		}
		blog.Info("Pool create reserved node %s success.", inst.GetKey())
	}
	//create available node
	if err := srv.store.Add(poolPath+"/available", []byte("available")); err != nil {
		blog.Errorf("Pool %s create reserved node err: %s", pool.GetKey(), err)
		reportMetrics("addpool", stateLogicFailure, started)
		return err
	}
	//create available ip instance with certain ip
	blog.Info("Pool create available ip instance with certain ip")
	for _, ip := range pool.Available {
		inst := &types.IPInst{
			IPAddr:     ip,
			Pool:       pool.Net,
			Mask:       pool.Mask,
			Gateway:    pool.Gateway,
			LastStatus: "created",
			Status:     types.IPStatus_AVAILABLE,
			Update:     time.Now().Format("2006-01-02 15:04:05"),
			Container:  "",
			Host:       "",
		}
		instData, _ := json.Marshal(inst)
		instPath := filepath.Join(poolPath, "available", inst.GetKey())
		if err := srv.store.Add(instPath, instData); err != nil {
			blog.Errorf("Pool create available node %s err: %v", inst.GetKey(), err)
			reportMetrics("addpool", stateLogicFailure, started)
			return err
		}
		blog.Info("Pool create available node %s success.", inst.GetKey())
	}

	//如果Reserved和Available都没有填，默认就是整个网段都是Available
	if len(pool.Reserved) == 0 && len(pool.Available) == 0 {
		blog.Info("Pool create available ip instance with net pool")
		allIps, _ := getAllHosts(pool.Net + "/" + strconv.Itoa(pool.Mask))
		for _, ip := range allIps {
			//fix(DeveloperJim): clean gateway in available ip address
			if ip == pool.Gateway {
				continue
			}
			if stringInSlice(ip, pool.Reserved) {
				continue
			}
			inst := &types.IPInst{
				IPAddr:     ip,
				Pool:       pool.Net,
				Mask:       pool.Mask,
				Gateway:    pool.Gateway,
				LastStatus: "created",
				Status:     types.IPStatus_AVAILABLE,
				Update:     time.Now().Format("2006-01-02 15:04:05"),
				Container:  "",
				Host:       "",
			}
			instData, _ := json.Marshal(inst)
			instPath := filepath.Join(poolPath, "available", inst.GetKey())
			if err := srv.store.Add(instPath, instData); err != nil {
				blog.Errorf("Pool create available node %s err: %v", inst.GetKey(), err)
				reportMetrics("addpool", stateLogicFailure, started)
				return err
			}
			blog.Info("Pool create available node %s success.", inst.GetKey())
		}
	}

	//create host node for query
	for _, host := range pool.Hosts {
		pHostPath := filepath.Join(poolPath, "hosts", host)
		hostInfo := &types.HostInfo{
			IPAddr:  host,
			Pool:    pool.Net,
			Cluster: pool.Cluster,
			Created: time.Now().Format("2006-01-02 15:04:05"),
			Update:  time.Now().Format("2006-01-02 15:04:05"),
		}
		hostData, _ := json.Marshal(hostInfo)
		if err := srv.store.Add(pHostPath, hostData); err != nil {
			blog.Errorf("Pool create query host node %s err: %v", host, err)
			reportMetrics("addpool", stateLogicFailure, started)
			return err
		}
		hostPath := filepath.Join(defaultHostInfoPath, host)
		if err := srv.store.Add(hostPath, hostData); err != nil {
			blog.Errorf("Create Host Node %s err: %v", hostPath, err)
		} else {
			blog.Info("Create Host node %s under pool %s success", host, pool.GetKey())
		}
	}
	reportMetrics("addpool", stateSuccess, started)
	return nil
}

func (srv *NetService) initPoolPath(pool *types.NetPool, poolPath string, poolData []byte) error {
	if err := srv.store.Add(poolPath, poolData); err != nil {
		blog.Errorf("Pool %s create data node %v err: %s", pool.GetKey(), pool, err.Error())
		return err
	}
	if err := srv.store.Add(poolPath+"/active", []byte("active")); err != nil {
		blog.Errorf("Pool %s create active node err: %s", pool.GetKey(), err)
		return err
	}
	if err := srv.store.Add(poolPath+"/hosts", []byte("hosts")); err != nil {
		blog.Errorf("Pool %s create hosts node err: %s", pool.GetKey(), err)
		return err
	}
	//creat reserved node
	if err := srv.store.Add(poolPath+"/reserved", []byte("reserved")); err != nil {
		blog.Errorf("Pool %s create reserved node err: %s", pool.GetKey(), err)
		return err
	}
	return nil
}

//DeletePool update Pool info
func (srv *NetService) DeletePool(netKey string) error {
	started := time.Now()
	//check pool existence
	poolPath := filepath.Join(defaultPoolInfoPath, netKey)
	blog.Info("try to delete pool %s", poolPath)
	if exist, _ := srv.store.Exist(poolPath); !exist {
		blog.Errorf("Delete pool %s error, no such Pool Info", netKey)
		reportMetrics("deletePool", stateNonExistFailure, started)
		return fmt.Errorf("pool %s do not exist", netKey)
	}
	//try to lock
	lockPath := filepath.Join(defaultLockerPath, netKey)
	distLocker, err := srv.store.GetLocker(lockPath)
	if err != nil {
		blog.Errorf("create pool %s lock err in delete pool, %v", netKey, err)
		reportMetrics("deletePool", stateLogicFailure, started)
		return fmt.Errorf("create pool %s lock err, %v", lockPath, err)
	}
	defer distLocker.Unlock()
	if err := distLocker.Lock(); err != nil {
		blog.Errorf("try to lock pool %s error, %v", netKey, err)
		reportMetrics("deletePool", stateLogicFailure, started)
		return fmt.Errorf("lock pool %s err, %v", netKey, err)
	}
	blog.Info("Lock pool %s in deletion success.", lockPath)
	//check active node is empty
	activepath := filepath.Join(defaultPoolInfoPath, netKey, "active")
	activeIPList, listErr := srv.store.List(activepath)
	if listErr != nil {
		blog.Errorf("check pool %s active ip address info in deletion err, %v", netKey, listErr)
		reportMetrics("deletePool", stateLogicFailure, started)
		return fmt.Errorf("check active ip list err, %v", listErr)
	}
	//check delete permision
	if len(activeIPList) != 0 {
		blog.Errorf("pool %s has %d ip in active, delete is deny", netKey, len(activeIPList))
		reportMetrics("deletePool", stateLogicFailure, started)
		return fmt.Errorf("delete pool %s is deny, pool is still active", netKey)
	}
	//clean active node
	if _, err := srv.store.Delete(activepath); err != nil {
		blog.Errorf("Delete pool %s active path err, %v", activepath, err)
		reportMetrics("deletePool", stateLogicFailure, started)
		return fmt.Errorf("delete pool %s active path err, %s", netKey, err.Error())
	}
	blog.Info("delete pool %s active path %s success", lockPath, activepath)
	//clean reserved node
	reservedpath := filepath.Join(defaultPoolInfoPath, netKey, "reserved")
	if err := srv.cleanChildrenNode(reservedpath); err != nil {
		blog.Errorf("delete pool %s reserved node err, %v", netKey, err)
		reportMetrics("deletePool", stateLogicFailure, started)
		return fmt.Errorf("delete pool %s reserved err, %s", netKey, err.Error())
	}
	//clean available node
	availablepath := filepath.Join(defaultPoolInfoPath, netKey, "available")
	if err := srv.cleanChildrenNode(availablepath); err != nil {
		blog.Errorf("delete pool %s available node err, %v", netKey, err)
		reportMetrics("deletePool", stateLogicFailure, started)
		return fmt.Errorf("delete pool %s available err, %s", netKey, err.Error())
	}
	//clean host query node
	hostpath := filepath.Join(defaultPoolInfoPath, netKey, "hosts")
	if err := srv.cleanChildrenNode(hostpath); err != nil {
		blog.Errorf("delete pool %s host node err, %v", netKey, err)
		reportMetrics("deletePool", stateLogicFailure, started)
		return fmt.Errorf("delete pool %s host err, %s", netKey, err.Error())
	}
	//clean pool node
	if _, err := srv.store.Delete(poolPath); err != nil {
		blog.Errorf("delete pool %s node err, %v", netKey, err)
		reportMetrics("deletePool", stateLogicFailure, started)
		return fmt.Errorf("delete pool %s err, %s", netKey, err.Error())
	}
	reportMetrics("deletePool", stateSuccess, started)
	return nil
}

//UpdatePool update Pool info
func (srv *NetService) UpdatePool(pool *types.NetPool, netKey string) error {
	started := time.Now()
	//check & construct data
	poolPath := filepath.Join(defaultPoolInfoPath, pool.GetKey())
	blog.Infof("try to Update pool %s", poolPath)
	//update netpool node with pool data
	if exist, _ := srv.store.Exist(poolPath); !exist {
		blog.Errorf("Pool %s is not exist, can not update", pool.GetKey())
		reportMetrics("updatePool", stateNonExistFailure, started)
		return fmt.Errorf("pool %s not exist, can not update", pool.GetKey())
	}
	if len(pool.Active) > 0 {
		blog.Errorf("Error Update Request, can not update Pool %s with Active Info", pool.GetKey())
		reportMetrics("updatePool", stateLogicFailure, started)
		return fmt.Errorf("can not update pool %s with Active info", pool.GetKey())
	}
	//try to lock
	lockPath := filepath.Join(defaultLockerPath, netKey)
	distLocker, lockErr := srv.store.GetLocker(lockPath)
	if lockErr != nil {
		blog.Errorf("create pool %s lock err in delete pool, %v", netKey, lockErr)
		reportMetrics("updatePool", stateLogicFailure, started)
		return fmt.Errorf("create pool %s lock err, %v", lockPath, lockErr)
	}
	defer distLocker.Unlock()
	if err := distLocker.Lock(); err != nil {
		blog.Errorf("try to lock pool %s error, %v", netKey, err)
		reportMetrics("updatePool", stateStorageFailure, started)
		return fmt.Errorf("lock pool %s err, %v", netKey, err)
	}
	blog.Infof("Lock pool %s in update success.", lockPath)
	//Get pool info
	oldPool, poolErr := srv.ListPoolByKey(netKey)
	if poolErr != nil {
		blog.Errorf("get old pool %s info failed, %v", netKey, poolErr)
		reportMetrics("updatePool", stateStorageFailure, started)
		return fmt.Errorf("check old pool info failed, %v", poolErr)
	}
	//filter duplicated ips, pool is left data for add,
	//oldPool is full data for update
	if err := srv.filterDuplicateIP(pool, oldPool); err != nil {
		blog.Errorf("filtDuplicateIP failed: %v", err)
		reportMetrics("updatePool", stateLogicFailure, started)
		return fmt.Errorf("filter info for %s failed, %v", netKey, err)
	}
	if len(pool.Available) == 0 && len(pool.Reserved) == 0 {
		blog.Warnf("No IP address can be update for pool %s, skip update work flow", netKey)
		reportMetrics("updatePool", stateSuccess, started)
		return nil
	}
	//append Reserved info
	for _, ip := range pool.Reserved {
		inst := &types.IPInst{
			IPAddr:     ip,
			Pool:       pool.Net,
			Mask:       pool.Mask,
			Gateway:    pool.Gateway,
			LastStatus: "created",
			Status:     types.IPStatus_RESERVED,
			Update:     time.Now().Format("2006-01-02 15:04:05"),
			Cluster:    pool.Cluster,
			Container:  "",
			Host:       "",
		}
		instData, _ := json.Marshal(inst)
		instPath := filepath.Join(poolPath, "reserved", inst.GetKey())
		if err := srv.store.Add(instPath, instData); err != nil {
			blog.Errorf("Pool %s create reserved node %s err: %v", pool.GetKey(), inst.GetKey(), err)
			reportMetrics("updatePool", stateStorageFailure, started)
			return err
		}
		blog.Infof("Pool %s append reserved node %s success.", pool.GetKey(), inst.GetKey())
	}

	for _, ip := range pool.Available {
		inst := &types.IPInst{
			IPAddr:     ip,
			Pool:       pool.Net,
			Mask:       pool.Mask,
			Gateway:    pool.Gateway,
			LastStatus: "created",
			Status:     types.IPStatus_AVAILABLE,
			Update:     time.Now().Format("2006-01-02 15:04:05"),
			Container:  "",
			Host:       "",
		}
		instData, _ := json.Marshal(inst)
		instPath := filepath.Join(poolPath, "available", inst.GetKey())
		if err := srv.store.Add(instPath, instData); err != nil {
			blog.Errorf("Pool %s create available node %s err: %v", pool.GetKey(), inst.GetKey(), err)
			reportMetrics("updatePool", stateStorageFailure, started)
			return err
		}
		blog.Infof("Pool %s append available node %s success.", pool.GetKey(), inst.GetKey())
	}

	//todo here create host node for query
	for _, host := range pool.Hosts {
		hostPath := filepath.Join(defaultHostInfoPath, host)
		exist, eErr := srv.store.Exist(hostPath)
		if eErr != nil {
			blog.Errorf("Pool %s check host %s failed, %v", pool.GetKey(), hostPath, eErr)
			reportMetrics("updatePool", stateStorageFailure, started)
			return fmt.Errorf("ip append success, but host %s check failed, %v", host, eErr)
		}
		if exist {
			blog.Errorf("Host %s exist, return", hostPath)
			reportMetrics("updatePool", stateLogicFailure, started)
			return fmt.Errorf("Host %s exist", host)
		}
		pHostPath := filepath.Join(poolPath, "hosts", host)
		hostInfo := &types.HostInfo{
			IPAddr:  host,
			Pool:    pool.Net,
			Cluster: pool.Cluster,
			Created: time.Now().Format("2006-01-02 15:04:05"),
			Update:  time.Now().Format("2006-01-02 15:04:05"),
		}
		hostData, _ := json.Marshal(hostInfo)
		if err := srv.store.Add(hostPath, hostData); err != nil {
			blog.Errorf("Create Host Node %s err: %v", hostPath, err)
			reportMetrics("updatePool", stateStorageFailure, started)
			return err
		}
		if err := srv.store.Add(pHostPath, hostData); err != nil {
			blog.Errorf("Pool %s append host %s err: %v", pool.GetKey(), host, err)
			reportMetrics("updatePool", stateStorageFailure, started)
			return err
		}
		blog.Infof("Create Host node %s under pool %s success", host, pool.GetKey())
	}
	//update pool node
	newPoolData, _ := json.Marshal(oldPool)
	if err := srv.store.Update(poolPath, newPoolData); err != nil {
		blog.Errorf("Update Pool %s value failed:%v", pool.GetKey(), err)
		reportMetrics("updatePool", stateStorageFailure, started)
		return err
	}
	reportMetrics("updatePool", stateSuccess, started)
	return nil
}

func (srv *NetService) filterDuplicateIP(update *types.NetPool, old *types.NetPool) error {
	existIPs := make(map[string]bool)
	for _, a := range old.Active {
		existIPs[a] = true
	}
	for _, ava := range old.Available {
		existIPs[ava] = true
	}
	for _, r := range old.Reserved {
		existIPs[r] = true
	}
	existHosts := make(map[string]bool)
	for _, h := range old.Hosts {
		existHosts[h] = true
	}
	var hostList []string
	for _, host := range update.Hosts {
		if _, ok := existHosts[host]; ok {
			blog.Warnf("Update pool %s host %s using now, skip", update.GetKey(), host)
			continue
		}
		hostList = append(hostList, host)
	}
	update.Hosts = hostList

	//check reserved ip info duplicated or not
	var tempReservedList []string
	for _, ip := range update.Reserved {
		if _, ok := existIPs[ip]; ok {
			blog.Warnf("Update pool %s ip %s in ReservedList are using now, skip", update.GetKey(), ip)
			continue
		}
		old.Reserved = append(old.Reserved, ip)
		tempReservedList = append(tempReservedList, ip)
	}
	//filter out reserved
	update.Reserved = tempReservedList

	var availableList []string
	for _, ip := range update.Available {
		if _, ok := existIPs[ip]; ok {
			blog.Warnf("Update pool %s ip %s in Available List are using now, skip", update.GetKey(), ip)
			continue
		}
		old.Available = append(old.Available, ip)
		availableList = append(availableList, ip)
	}
	update.Available = availableList
	return nil
}

//ListPool update Pool info
func (srv *NetService) ListPool() ([]*types.NetPool, error) {
	started := time.Now()
	//list all pool from defaultPoolPath
	clusters, err := srv.store.List(defaultPoolInfoPath)
	if err != nil {
		blog.Errorf("list all net pool err, %v", err)
		reportMetrics("updatePool", stateStorageFailure, started)
		return nil, fmt.Errorf("list all pool err, %s", err.Error())
	}
	var pools []*types.NetPool
	if len(clusters) == 0 {
		blog.Info("No Cluseter here now")
		reportMetrics("updatePool", stateSuccess, started)
		return pools, nil
	}
	for _, c := range clusters {
		list, err := srv.ListPoolByCluster(c)
		if err != nil {
			reportMetrics("updatePool", stateStorageFailure, started)
			return nil, err
		}
		if len(list) == 0 {
			continue
		}
		pools = append(pools, list...)
	}
	reportMetrics("updatePool", stateSuccess, started)
	return pools, nil
}

//ListPoolByKey Get pool info by net
func (srv *NetService) ListPoolByKey(net string) (*types.NetPool, error) {
	started := time.Now()
	blog.Info("try to get pool %s info.", net)
	poolpath := filepath.Join(defaultPoolInfoPath, net)
	poolData, gerr := srv.store.Get(poolpath)
	if gerr != nil {
		blog.Errorf("Get pool %s by key err, %v", net, gerr)
		reportMetrics("listPoolByKey", stateStorageFailure, started)
		return nil, fmt.Errorf("get pool %s by key err, %s", net, gerr.Error())
	}
	pool := &types.NetPool{}
	if err := json.Unmarshal(poolData, pool); err != nil {
		blog.Errorf("Pool %s decode json data err: %v, origin data: %s", net, err, string(poolData))
		reportMetrics("listPoolByKey", stateJSONFailure, started)
		return nil, fmt.Errorf("decode pool %s data err, %s", net, err.Error())
	}
	//get hosts
	hostpath := filepath.Join(defaultPoolInfoPath, net, "hosts")
	hosts, err := srv.store.List(hostpath)
	if err != nil {
		blog.Errorf("list pool %s path %s failed, %v", net, hostpath, err)
		reportMetrics("listPoolByKey", stateStorageFailure, started)
		return nil, fmt.Errorf("list pool %s host path failed, %v", net, err)
	}
	pool.Hosts = hosts
	//get reserved list
	rpath := filepath.Join(defaultPoolInfoPath, net, "reserved")
	reserved, err := srv.store.List(rpath)
	if err != nil {
		blog.Errorf("list pool %s path %s err, %v", net, rpath, err)
		reportMetrics("listPoolByKey", stateStorageFailure, started)
		return nil, fmt.Errorf("list pool %s reserved path err, %s", net, err.Error())
	}
	pool.Reserved = reserved
	apath := filepath.Join(defaultPoolInfoPath, net, "available")
	available, err := srv.store.List(apath)
	if err != nil {
		blog.Errorf("list pool %s path %s err, %v", net, apath, err)
		reportMetrics("listPoolByKey", stateStorageFailure, started)
		return nil, fmt.Errorf("list pool %s available path err, %s", net, err.Error())
	}
	pool.Available = available
	activepath := filepath.Join(defaultPoolInfoPath, net, "active")
	active, err := srv.store.List(activepath)
	if err != nil {
		blog.Errorf("list pool %s path %s err, %v", net, activepath, err)
		reportMetrics("listPoolByKey", stateStorageFailure, started)
		return nil, fmt.Errorf("list pool %s active path err, %s", net, err.Error())
	}
	pool.Active = active
	blog.Info("Get pool %s info success", net)
	reportMetrics("listPoolByKey", stateSuccess, started)
	return pool, nil
}

//ListPoolByCluster list all pool under cluster
func (srv *NetService) ListPoolByCluster(cluster string) ([]*types.NetPool, error) {
	started := time.Now()
	//list all pool from defaultPoolPath
	clusterpath := filepath.Join(defaultPoolInfoPath, cluster)
	ok, err := srv.store.Exist(clusterpath)
	if err != nil {
		blog.Errorf("check cluster %s exist failed, %v", clusterpath, err)
		reportMetrics("listPoolByCluster", stateStorageFailure, started)
		return nil, err
	}
	if !ok {
		blog.Warnf("####No Cluster %s creatd###", cluster)
		reportMetrics("listPoolByCluster", stateSuccess, started)
		return nil, nil
	}
	nets, err := srv.store.List(clusterpath)
	if err != nil {
		blog.Errorf("list all net pool under %s err, %v", cluster, err)
		reportMetrics("listPoolByCluster", stateStorageFailure, started)
		return nil, fmt.Errorf("list all pool err, %s", err.Error())
	}
	var pools []*types.NetPool
	if len(nets) == 0 {
		blog.Infof("No netpool under cluster %s now", cluster)
		reportMetrics("listPoolByCluster", stateSuccess, started)
		return pools, nil
	}
	for _, net := range nets {
		pool, err := srv.ListPoolByKey(cluster + "/" + net)
		if err != nil {
			blog.Errorf("Get pool %s/%s info err, %v", cluster, net, err)
			reportMetrics("listPoolByCluster", stateStorageFailure, started)
			return nil, err
		}
		pools = append(pools, pool)
	}
	reportMetrics("listPoolByCluster", stateSuccess, started)
	return pools, nil
}

//cleanChildrenNode clean all children node and node itself
func (srv *NetService) cleanChildrenNode(nodepath string) error {
	nodeList, err := srv.store.List(nodepath)
	if err != nil {
		blog.Errorf("List %s all sub node err, %s", nodepath, err.Error())
		return err
	}
	for _, node := range nodeList {
		sub := filepath.Join(nodepath, node)
		if _, err := srv.store.Delete(sub); err != nil {
			blog.Errorf("Delete sub %s err, %v", sub, err)
			return err
		}
	}
	//clean self node
	if _, err := srv.store.Delete(nodepath); err != nil {
		blog.Errorf("Delete node %s err, %v", nodepath, err)
		return err
	}
	return nil
}

//GetPoolAvailable get pool available ip list only
func (srv *NetService) GetPoolAvailable(net string) (*types.NetPool, error) {
	started := time.Now()
	blog.Info("try to get pool %s info.", net)
	poolpath := filepath.Join(defaultPoolInfoPath, net)
	poolData, getErr := srv.store.Get(poolpath)
	if getErr != nil {
		blog.Errorf("Get pool %s by key err, %v", net, getErr)
		reportMetrics("getPoolAvailable", stateStorageFailure, started)
		return nil, fmt.Errorf("get pool %s by key err, %s", net, getErr.Error())
	}
	pool := &types.NetPool{}
	if err := json.Unmarshal(poolData, pool); err != nil {
		blog.Errorf("Pool %s decode json data err: %v, origin data: %s", net, err, string(poolData))
		reportMetrics("getPoolAvailable", stateJSONFailure, started)
		return nil, fmt.Errorf("decode pool %s data err, %s", net, err.Error())
	}
	//get available node
	apath := filepath.Join(defaultPoolInfoPath, net, "available")
	available, err := srv.store.List(apath)
	if err != nil {
		blog.Errorf("list pool %s path %s err, %v", net, apath, err)
		reportMetrics("getPoolAvailable", stateStorageFailure, started)
		return nil, fmt.Errorf("list pool %s available path err, %s", net, err.Error())
	}
	pool.Available = available
	reportMetrics("getPoolAvailable", stateSuccess, started)
	return pool, nil
}
