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

package daemon

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	syncClusterNodeCMDBMetaLockKey = "syncClusterNodeCMDBMeta"
)

// syncClusterNodeCMDBMeta 集群中节点与蓝鲸CMDB系统的元数据联动，
// 动态获取集群中的节点数据信息，并与CMDB中的数据进行对比处理：
// 1. 集群中存在，CMDB不存在 → 创建主机到资源池并转移到空闲机器池，若配置特性模块则进行转移
// 2. 集群中存在，CMDB存在 → 更新HOST元数据信息（暂不处理）
// 3. 集群中不存在，CMDB存在 → 转移到空闲机池→资源池，删除主机
// 4. 集群中不存在，CMDB不存在 → 无需处理
func (d *Daemon) syncClusterNodeCMDBMeta(errChan chan<- error) {
	blog.Infof("syncClusterNodeCMDBMeta start")
	d.lock.Lock(syncClusterNodeCMDBMetaLockKey,
		[]lock.LockOption{lock.LockTTL(time.Second * 60)}...) // nolint
	defer d.lock.Unlock(syncClusterNodeCMDBMetaLockKey)       // nolint

	defer utils.RecoverPrintStack("syncClusterNodeCMDBMeta")

	cmdbClient := cmdb.GetCmdbClient()
	if cmdbClient == nil {
		blog.Errorf("syncClusterNodeCMDBMeta cmdb client not initialized")
		return
	}

	// 获取所有运行中的集群
	cond := operator.NewBranchCondition(operator.And,
		operator.NewLeafCondition(operator.In, operator.M{
			"status": []string{common.StatusRunning, common.StatusConnectClusterFailed},
		}),
	)

	clusterList, err := d.model.ListCluster(d.ctx, cond, &storeopt.ListOption{All: true})
	if err != nil {
		blog.Errorf("syncClusterNodeCMDBMeta ListCluster failed: %v", err)
		errChan <- err
		return
	}

	if len(clusterList) == 0 {
		blog.Infof("syncClusterNodeCMDBMeta no running clusters found")
		return
	}
	blog.Infof("syncClusterNodeCMDBMeta found %d running clusters", len(clusterList))

	// 汇总所有集群的节点与CMDB主机数据，并按集群保留完整同步上下文，取完后统一同步
	allClusterNodesMap := make(map[string]*cmproto.Node)
	allCmdbNodesMap := make(map[string]*cmdb.HostData)
	clusterSyncInfoMap := make(map[string]*clusterSyncInfo)

	concurrency := utils.NewRoutinePool(10)
	defer concurrency.Close()

	// 保护汇总map的并发写入
	var mu sync.Mutex

	for i := range clusterList {
		concurrency.Add(1)
		go func(cluster *cmproto.Cluster) {
			defer concurrency.Done()

			// 检查集群是否在白名单中（Labels中syncImportCluster为true）
			if cluster.GetLabels()[common.SyncImportClusterLabel] != "true" {
				blog.Infof("syncClusterNodeCMDBMeta cluster[%s] not in whitelist, skip",
					cluster.GetClusterID())
				return
			}

			if cluster.GetBusinessID() == "" {
				blog.Warnf("syncClusterNodeCMDBMeta cluster[%s] has no businessID, skip",
					cluster.GetClusterID())
				return
			}

			clusterNodesMap, cmdbNodesMap, errLocal := d.getClusterSyncNodesInfo(cluster)
			if errLocal != nil {
				blog.Errorf("syncClusterNodeCMDBMeta cluster[%s] getClusterSyncNodesInfo failed: %v",
					cluster.GetClusterID(), errLocal)
				errChan <- errLocal
				return
			}

			bizID, errLocal := strconv.Atoi(cluster.GetBusinessID())
			if errLocal != nil {
				blog.Errorf("syncClusterNodeCMDBMeta cluster[%s] parse businessID[%s] failed: %v",
					cluster.GetClusterID(), cluster.GetBusinessID(), errLocal)
				errChan <- errLocal
				return
			}

			mu.Lock()
			defer mu.Unlock()

			clusterID := cluster.GetClusterID()
			if _, find := clusterSyncInfoMap[clusterID]; !find {
				clusterSyncInfoMap[clusterID] = &clusterSyncInfo{
					cluster:          cluster,
					bizID:            bizID,
					clusterNodeIPMap: clusterNodesMap,
					cmdbHostIPMap:    cmdbNodesMap,
				}
			}

			for ip, node := range clusterNodesMap {
				if _, find := allClusterNodesMap[ip]; !find {
					allClusterNodesMap[ip] = node
				}
			}
			for ip, host := range cmdbNodesMap {
				if _, find := allCmdbNodesMap[ip]; !find {
					allCmdbNodesMap[ip] = host
				}
			}
		}(clusterList[i])
	}

	concurrency.Wait()

	// 所有集群数据取完后，统一进行对比同步
	if errLocal := d.syncClusterNodeWithCMDB(cmdbClient, allClusterNodesMap, allCmdbNodesMap,
		clusterSyncInfoMap); errLocal != nil {
		blog.Errorf("syncClusterNodeCMDBMeta syncClusterNodeWithCMDB failed: %v", errLocal)
		errChan <- errLocal
	}
}

// clusterSyncInfo 汇总阶段保留的单集群完整同步上下文：
// 集群对象、业务ID、该集群的节点与CMDB主机数据，同步阶段据此直接调用现有handler
type clusterSyncInfo struct {
	cluster          *cmproto.Cluster
	bizID            int
	clusterNodeIPMap map[string]*cmproto.Node
	cmdbHostIPMap    map[string]*cmdb.HostData
}

// syncClusterNodeWithCMDB 集群中节点与CMDB主机数据对比同步：
// 1. 集群中存在，CMDB不存在 → 创建主机到资源池并转移到空闲机器池，若配置特性模块则进行转移
// 2. 集群中存在，CMDB存在 → 更新HOST元数据信息（暂不处理）
// 3. 集群中不存在，CMDB存在 → 转移到空闲机池→资源池，删除主机
// 4. 集群中不存在，CMDB不存在 → 无需处理
func (d *Daemon) syncClusterNodeWithCMDB(cmdbClient *cmdb.Client,
	allClusterNodesMap map[string]*cmproto.Node, allCmdbNodesMap map[string]*cmdb.HostData,
	clusterSyncInfoMap map[string]*clusterSyncInfo) error {

	if cmdbClient == nil {
		return fmt.Errorf("cmdb client not initialized")
	}

	blog.Infof("syncClusterNodeWithCMDB start, clusterNodes[%d] cmdbHosts[%d] clusters[%d]",
		len(allClusterNodesMap), len(allCmdbNodesMap), len(clusterSyncInfoMap))

	// 集群中存在，CMDB不存在 → 创建主机到资源池并转移到空闲机器池
	// 按集群维度处理：判断节点是否已存在于CMDB，需与该业务的CMDB全量主机比对
	bizID := 0
	for clusterID, syncInfo := range clusterSyncInfoMap {
		bizID = syncInfo.bizID

		if errLocal := d.handleNodesInClusterNotInCMDB(syncInfo.cluster, syncInfo.bizID, cmdbClient,
			syncInfo.clusterNodeIPMap, syncInfo.cmdbHostIPMap); errLocal != nil {
			blog.Errorf("syncClusterNodeWithCMDB cluster[%s] handleNodesInClusterNotInCMDB failed: %v",
				clusterID, errLocal)
			return errLocal
		}

		blog.Infof("syncClusterNodeWithCMDB cluster[%s] compare done", clusterID)
	}

	// 集群中存在，CMDB存在 → 更新HOST元数据信息（暂不处理该情况）

	// 集群中不存在，CMDB存在 → 转移到空闲机池→资源池，删除主机
	// CMDB主机是业务维度数据（含该业务下所有集群的节点），必须基于全部集群的汇总节点做差集，
	// 只删除在所有集群都不存在的主机；若放在上面的集群循环内，会把其他集群的节点误判为待删除
	blog.Infof("syncClusterNodeWithCMDB cleanup bizID[%d] clusterNodes[%d] cmdbHosts[%d]",
		bizID, len(allClusterNodesMap), len(allCmdbNodesMap))

	if errLocal := d.handleNodesNotInClusterInCMDB(bizID, cmdbClient,
		allClusterNodesMap, allCmdbNodesMap); errLocal != nil {
		blog.Errorf("syncClusterNodeWithCMDB bizID[%d] handleNodesNotInClusterInCMDB failed: %v",
			bizID, errLocal)
		return errLocal
	}

	blog.Infof("syncClusterNodeWithCMDB all clusters compare done")

	return nil
}

// getClusterSyncNodesInfo 同步单个集群节点与CMDB数据
func (d *Daemon) getClusterSyncNodesInfo(cluster *cmproto.Cluster) (map[string]*cmproto.Node, map[string]*cmdb.HostData, error) {
	clusterID := cluster.GetClusterID()
	bizIDStr := cluster.GetBusinessID()
	blog.Infof("syncClusterNodeWithCMDB cluster[%s] provider[%s] bizID[%s] start",
		clusterID, cluster.GetProvider(), bizIDStr)

	bizID, err := strconv.Atoi(bizIDStr)
	if err != nil {
		return nil, nil, fmt.Errorf("parse businessID[%s] failed: %v", bizIDStr, err)
	}

	// 获取CMDB客户端
	cmdbClient := cmdb.GetCmdbClient()
	if cmdbClient == nil {
		return nil, nil, fmt.Errorf("cmdb client not initialized")
	}

	// 从k8s API获取集群中所有节点IP列表
	nodeIPs, err := d.listClusterNodeIPsFromK8s(cluster)
	if err != nil {
		return nil, nil, fmt.Errorf("listClusterNodeIPsFromK8s[%s] failed: %v", clusterID, err)
	}
	blog.Infof("syncClusterNodeWithCMDB cluster[%s] listClusterNodeIPsFromK8s success, nodes[%d]",
		clusterID, len(nodeIPs))

	// 通过cloudprovider获取节点详细信息
	clusterNodeIPMap, err := d.getCloudNodesInfo(cluster, nodeIPs)
	if err != nil {
		blog.Warnf("syncClusterNodeWithCMDB cluster[%s] getCloudNodesInfo failed: %v, "+
			"fallback to k8s node info", clusterID, err)
		return nil, nil, fmt.Errorf("getCloudNodesInfo[%s] failed: %v", clusterID, err)
	}

	blog.Infof("syncClusterNodeWithCMDB cluster[%s] getCloudNodesInfo success, nodes[%d]",
		clusterID, len(clusterNodeIPMap))

	// 仅同步GPU节点
	beforeFilter := len(clusterNodeIPMap)
	clusterNodeIPMap = filterGPUMapNodes(clusterNodeIPMap)
	blog.Infof("syncClusterNodeWithCMDB cluster[%s] filterGPUNodes, before[%d] after[%d]",
		clusterID, beforeFilter, len(clusterNodeIPMap))

	// 获取CMDB中该业务下所有主机
	cmdbHosts, err := cmdbClient.FetchAllHostsByBizID(d.ctx, bizID, false)
	if err != nil {
		return nil, nil, fmt.Errorf("FetchAllHostsByBizID[%d] failed: %v", bizID, err)
	}
	blog.Infof("syncClusterNodeWithCMDB cluster[%s] FetchAllHostsByBizID bizID[%d] success, hosts[%d]",
		clusterID, bizID, len(cmdbHosts))

	// 构建CMDB主机IP到主机的映射，过滤非GPU主机
	cmdbHostIPMap := buildCMDBHostIPMap(cmdbHosts)

	blog.Infof("syncClusterNodeWithCMDB cluster[%s] provider[%s] bizID[%d] "+
		"clusterNodes[%d] cmdbHosts[%d]", clusterID, cluster.GetProvider(), bizID,
		len(clusterNodeIPMap), len(cmdbHostIPMap))

	return clusterNodeIPMap, cmdbHostIPMap, nil
}

// filterGPUMapNodes 根据GPU判断逻辑过滤出GPU节点
func filterGPUMapNodes(nodeMap map[string]*cmproto.Node) map[string]*cmproto.Node {
	gpuMap := make(map[string]*cmproto.Node)
	for ip, node := range nodeMap {
		if node.GetGPU() > 0 {
			gpuMap[ip] = node
		}
	}
	return gpuMap
}

// listClusterNodeIPsFromK8s 从k8s API获取集群中所有节点的IP列表
func (d *Daemon) listClusterNodeIPsFromK8s(cluster *cmproto.Cluster) ([]string, error) {
	clusterID := cluster.GetClusterID()

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), d.model)
	blog.Infof("listClusterNodeIPsFromK8s cluster[%s] start ListClusterNodes", clusterID)
	k8sNodes, err := k8sOperator.ListClusterNodes(d.ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("ListClusterNodes[%s] failed: %v", clusterID, err)
	}

	var ips []string
	for _, k8sNode := range k8sNodes {
		ipv4s, _ := utils.GetNodeIPAddress(k8sNode)
		if len(ipv4s) > 0 {
			ips = append(ips, ipv4s[0])
		}
	}

	return ips, nil
}

// getCloudNodesInfo 通过cloudprovider的ListImportClusterNodes获取节点详细信息
// 各云厂商有差异化的k8s label配置和providerID解析逻辑
// 返回 IP -> Node 的映射
func (d *Daemon) getCloudNodesInfo(cluster *cmproto.Cluster, nodeIPs []string) (map[string]*cmproto.Node, error) {
	if len(nodeIPs) == 0 {
		return nil, fmt.Errorf("no node IPs provided")
	}

	// 通过cluster.Provider(cloudID)查找cloud记录
	blog.Infof("getCloudNodesInfo cluster[%s] try GetCloud by provider[%s]",
		cluster.GetClusterID(), cluster.GetProvider())
	cloud, err := d.model.GetCloud(d.ctx, cluster.GetProvider())
	if err != nil {
		blog.Errorf("getCloudNodesInfo cluster[%s] GetCloud by provider[%s] failed: %v",
			cluster.GetClusterID(), cluster.GetProvider(), err)
		return nil, fmt.Errorf("GetCloud by provider[%s] failed: %v", cluster.GetProvider(), err)
	}
	blog.Infof("getCloudNodesInfo cluster[%s] GetCloud success, cloudID[%s] cloudProvider[%s]",
		cluster.GetClusterID(), cloud.GetCloudID(), cloud.GetCloudProvider())

	nodeMgr, err := cloudprovider.GetNodeMgr(cloud.GetCloudProvider())
	if err != nil {
		blog.Errorf("getCloudNodesInfo cluster[%s] GetNodeMgr by cloudName[%s] failed: %v",
			cluster.GetClusterID(), cloud.GetCloudProvider(), err)
		return nil, fmt.Errorf("GetNodeMgr by cloudName[%s] failed: %v", cloud.GetCloudProvider(), err)
	}
	blog.Infof("getCloudNodesInfo cluster[%s] GetNodeMgr success, will ListImportClusterNodes with "+
		"region[%s] nodeIPs[%d]", cluster.GetClusterID(), cluster.GetRegion(), len(nodeIPs))

	cloudNodes, err := nodeMgr.ListImportClusterNodes(nodeIPs, &cloudprovider.ListNodesOption{
		ClusterID: cluster.GetClusterID(),
		Common: &cloudprovider.CommonOption{
			Region: cluster.GetRegion(),
		},
	})
	if err != nil {
		blog.Errorf("getCloudNodesInfo cluster[%s] ListImportClusterNodes failed: %v",
			cluster.GetClusterID(), err)
		return nil, fmt.Errorf("ListImportClusterNodes failed: %v", err)
	}
	blog.Infof("getCloudNodesInfo cluster[%s] ListImportClusterNodes returned[%d] nodes",
		cluster.GetClusterID(), len(cloudNodes))

	cloudNodesMap := make(map[string]*cmproto.Node)
	for _, cn := range cloudNodes {
		if ip := cn.GetInnerIP(); ip != "" {
			cloudNodesMap[ip] = cn
		}
	}

	blog.Infof("getCloudNodesInfo cluster[%s] provider[%s] requested[%d] found[%d]",
		cluster.GetClusterID(), cluster.GetProvider(), len(nodeIPs), len(cloudNodesMap))

	return cloudNodesMap, nil
}

// buildCMDBHostIPMap 构建CMDB主机IP到主机的映射，仅保留GPU主机
func buildCMDBHostIPMap(hosts []cmdb.HostData) map[string]*cmdb.HostData {
	ipMap := make(map[string]*cmdb.HostData)
	for i := range hosts {
		if hosts[i].BKHostInnerIP != "" {
			if hosts[i].IsGPU != true {
				continue
			}
			ipMap[hosts[i].BKHostInnerIP] = &hosts[i]
		}
	}
	return ipMap
}

// cmdbCloudVendorMap BCS云厂商标识到CMDB云厂商枚举值的映射
// CMDB枚举值：亚马逊云(1)、腾讯云(2)、GCP(3)、微软云(4)、企业私有云(5)、
// 雨花石云(6)、Oracle云(7)、IBM云(8)、阿里云(9)、中国电信(10)、
// 优云(11)、美团云(12)、金山云(13)、百度云(14)、华为云(15)、
// 首都云(16)、腾讯自研云(17)、Zenlayer(18)
var cmdbCloudVendorMap = map[string]string{
	"awsCloud":     "1",  // 亚马逊云
	"tencentCloud": "2",  // 腾讯云
	"gcpCloud":     "3",  // GCP
	"azureCloud":   "4",  // 微软云
	"eopCloud":     "5",  // 企业私有云
	"oracleCloud":  "7",  // Oracle云
	"ibmCloud":     "8",  // IBM云
	"huaweiCloud":  "15", // 华为云
}

// getCMDBCloudVendorCode 将BCS云厂商标识转换为CMDB云厂商枚举值
func getCMDBCloudVendorCode(provider string) string {
	if code, ok := cmdbCloudVendorMap[provider]; ok {
		return code
	}
	return provider
}

// buildAddHostInfo 根据集群和节点信息构建创建主机到CMDB的主机信息
func buildAddHostInfo(cluster *cmproto.Cluster, node *cmproto.Node, ip string) cmdb.AddHostToBusinessIdleHostInfo {
	isGPU := false
	if node.GetGPU() > 0 {
		isGPU = true
	}

	hostInfo := cmdb.AddHostToBusinessIdleHostInfo{
		BkHostInnerIP: ip,
		BkCloudID:     0,
		BkAddressing:  "dynamic",
		// CMDB主机扩展属性
		IsGPU:         isGPU,
		BkCloudVendor: getCMDBCloudVendorCode(cluster.GetProvider()),
		BkCloudInstID: node.GetNodeID(),
		BkCloudRegion: cluster.GetRegion(),
		BkCloudZone:   node.GetZoneName(),
		InstanceType:  node.GetInstanceType(),
		OnshelfDate:   cluster.GetCreateTime(),
		Operator:      cluster.GetCreator(),
		BkBakOperator: cluster.GetCreator(),
	}

	return hostInfo
}

// buildUpdateHostData 根据集群和节点信息构建更新CMDB主机属性的数据
func buildUpdateHostData(cluster *cmproto.Cluster, node *cmproto.Node) *cmdb.UpdateHostData {
	isGPU := false
	if node.GetGPU() > 0 {
		isGPU = true
	}

	return &cmdb.UpdateHostData{
		BkHostName:    node.GetNodeName(),
		IsGPU:         isGPU,
		BkCloudVendor: getCMDBCloudVendorCode(cluster.GetProvider()),
		BkCloudInstID: node.GetNodeID(),
		BkCloudRegion: cluster.GetRegion(),
		BkCloudZone:   node.GetZoneName(),
		InstanceType:  node.GetInstanceType(),
		OnshelfDate:   cluster.GetCreateTime(),
		Operator:      cluster.GetCreator(),
		BkBakOperator: cluster.GetCreator(),
	}
}

// handleNodesInClusterNotInCMDB 集群中存在，CMDB不存在 → 创建主机到业务空闲机
// 若用户配置了特性模块（ExtraInfo中的cmdbFeatureModuleID），则进一步转移到特性模块
func (d *Daemon) handleNodesInClusterNotInCMDB(cluster *cmproto.Cluster, bizID int,
	cmdbClient *cmdb.Client, clusterNodeIPMap map[string]*cmproto.Node, cmdbHostIPMap map[string]*cmdb.HostData) error {

	clusterID := cluster.GetClusterID()

	for ip, node := range clusterNodeIPMap {
		if _, exists := cmdbHostIPMap[ip]; exists {
			continue
		}

		blog.Infof("syncClusterNodeWithCMDB cluster[%s] node[%s][%s] exists in cluster but not in CMDB, "+
			"will add to business idle, provider[%s]", clusterID, node.GetNodeName(), ip,
			cluster.GetProvider())

		// 构建主机信息，创建主机到业务空闲机（包含CMDB扩展属性）
		hostInfo := buildAddHostInfo(cluster, node, ip)

		hostIDs, err := cmdbClient.AddHostToBusinessIdle(d.ctx, bizID,
			[]cmdb.AddHostToBusinessIdleHostInfo{hostInfo})
		if err != nil {
			blog.Errorf("syncClusterNodeWithCMDB handleNodesInClusterNotInCMDB AddHostToBusinessIdle cluster[%s] node[%s] failed: %v",
				clusterID, ip, err)
			continue
		}

		blog.Infof("syncClusterNodeWithCMDB handleNodesInClusterNotInCMDB AddHostToBusinessIdle cluster[%s] node[%s] success, "+
			"hostIDs[%v]", clusterID, ip, hostIDs)

		moduleIDStr := cluster.GetClusterBasicSettings().GetModule().WorkerModuleID
		// 若用户配置了特性模块，则调用 TransferHostModule 将主机转移到特性模块
		if moduleIDStr != "" && len(hostIDs) > 0 {
			moduleID, err := strconv.Atoi(moduleIDStr)
			if err != nil {
				blog.Errorf("syncClusterNodeWithCMDB handleNodesInClusterNotInCMDB parse moduleID[%s] failed: %v", moduleIDStr, err)
			} else {
				intHostIDs := make([]int, 0, len(hostIDs))
				for _, hid := range hostIDs {
					intHostIDs = append(intHostIDs, int(hid))
				}
				if err := cmdbClient.TransferHostModule(d.ctx, bizID, intHostIDs, []int{moduleID}, false); err != nil {
					blog.Errorf("syncClusterNodeWithCMDB handleNodesInClusterNotInCMDB TransferHostModule cluster[%s] "+
						"moduleID[%d] hostIDs[%v] failed: %v", clusterID, moduleID, intHostIDs, err)
				} else {
					blog.Infof("syncClusterNodeWithCMDB handleNodesInClusterNotInCMDB TransferHostModule cluster[%s] "+
						"moduleID[%d] hostIDs[%v] success", clusterID, moduleID, intHostIDs)
				}
			}
		}
	}

	return nil
}

// handleNodesNotInClusterInCMDB 集群中不存在，CMDB存在 → 转移到空闲机池→资源池，删除主机
func (d *Daemon) handleNodesNotInClusterInCMDB(bizID int, cmdbClient *cmdb.Client,
	clusterNodeIPMap map[string]*cmproto.Node, cmdbHostIPMap map[string]*cmdb.HostData) error {

	var hostIDsToDelete []int

	for ip, cmdbHost := range cmdbHostIPMap {
		if _, exists := clusterNodeIPMap[ip]; exists {
			continue
		}

		blog.Infof("syncClusterNodeWithCMDB bizID[%d] host[%s][%d] exists in CMDB but not in cluster, "+
			"will transfer to idle then resource pool and delete", bizID, ip, cmdbHost.BKHostID)

		// Step1: 转移到空闲机池
		if err := cmdbClient.TransferHostToIdleModule(d.ctx, bizID, []int{int(cmdbHost.BKHostID)}); err != nil {
			blog.Errorf("syncClusterNodeWithCMDB handleNodesNotInClusterInCMDB TransferHostToIdleModule bizID[%d] host[%s] failed: %v",
				bizID, ip, err)
			continue
		}

		// Step2: 转移到资源池
		if err := cmdbClient.TransferHostToResourceModule(d.ctx, bizID, []int{int(cmdbHost.BKHostID)}); err != nil {
			blog.Errorf("syncClusterNodeWithCMDB handleNodesNotInClusterInCMDB TransferHostToResourceModule bizID[%d] host[%s] failed: %v",
				bizID, ip, err)
			continue
		}

		// Step3: 记录待删除的主机ID
		hostIDsToDelete = append(hostIDsToDelete, int(cmdbHost.BKHostID))
	}

	// Step4: 批量删除主机
	if len(hostIDsToDelete) > 0 {
		if err := cmdbClient.DeleteHost(d.ctx, hostIDsToDelete); err != nil {
			blog.Errorf("syncClusterNodeWithCMDB handleNodesNotInClusterInCMDB DeleteHost bizID[%d] hostIDs[%v] failed: %v",
				bizID, hostIDsToDelete, err)
			return err
		}
		blog.Infof("syncClusterNodeWithCMDB handleNodesNotInClusterInCMDB DeleteHost bizID[%d] hostIDs[%v] success",
			bizID, hostIDsToDelete)
	}

	return nil
}
