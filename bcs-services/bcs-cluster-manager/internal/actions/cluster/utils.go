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

package cluster

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	spb "google.golang.org/protobuf/types/known/structpb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	autils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	provider "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/passcc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ClusterInfo info
// nolint revive
type ClusterInfo struct {
	ClusterName string
	ClusterID   string
}

func generateClusterID(cls *proto.Cluster, model store.ClusterManagerModel) (string, int, error) {
	clusterEnv := cls.Environment
	clusterEngine := cls.EngineType

	clusterNum, err := getClusterMaxNum(clusterEngine, clusterEnv, model)
	if err != nil {
		return "", 0, err
	}

	envTypeStart := common.ClusterIDRange[fmt.Sprintf("%v-%v", clusterEngine, clusterEnv)]
	clusterNum = int(math.Max(float64(clusterNum), float64(envTypeStart[0])))
	clusterID := fmt.Sprintf("BCS-%v-%v", strings.ToUpper(clusterEngine), clusterNum)
	return clusterID, clusterNum, nil
}

func getClusterMaxNum(clusterType string, env string, model store.ClusterManagerModel) (int, error) {
	_, ok := common.EngineTypeLookup[clusterType]
	if !ok {
		return 0, fmt.Errorf("clusterType[%s] failed", clusterType)
	}

	_, ok = common.ClusterEnvMap[env]
	if !ok {
		return 0, fmt.Errorf("cluster env[%s] failed", env)
	}

	condM := make(operator.M)
	condM["enginetype"] = clusterType
	condM["environment"] = env
	cond := operator.NewLeafCondition(operator.Eq, condM)
	clusterList, err := model.ListCluster(context.Background(), cond, &storeopt.ListOption{All: true})
	if err != nil {
		return 0, err
	}

	clusterNumIDs := make([]int, 0)
	for i := range clusterList {
		clusterStrs := strings.Split(clusterList[i].ClusterID, "-")
		if len(clusterStrs) != 3 {
			continue
		}

		id, _ := strconv.Atoi(clusterStrs[2])
		clusterNumIDs = append(clusterNumIDs, id)
	}
	sort.Ints(clusterNumIDs)

	if len(clusterNumIDs) == 0 {
		return 1, nil
	}

	return clusterNumIDs[len(clusterNumIDs)-1] + 1, nil
}

// getClusterList get all cm clusters
func getClusterList(model store.ClusterManagerModel) ([]*proto.Cluster, error) {
	clusterStatus := []string{common.StatusInitialization, common.StatusRunning, common.StatusDeleting}
	condStatus := operator.NewLeafCondition(operator.In, operator.M{"status": clusterStatus})

	clusterList, err := model.ListCluster(context.Background(), condStatus, &storeopt.ListOption{All: true})
	if err != nil {
		blog.Errorf("getClusterList failed: %v", err)
		return nil, err
	}

	return clusterList, nil
}

// VClusterHostFilterInfo xxx
type VClusterHostFilterInfo struct {
	Provider string
	Region   string
	Version  string
}

// selectVclusterHostCluster for select vcluster host cluster
func selectVclusterHostCluster(model store.ClusterManagerModel, filter VClusterHostFilterInfo) (string, error) {
	condCluster := operator.NewLeafCondition(operator.Eq, operator.M{
		"isshared": true,
		"region":   filter.Region,
		"provider": filter.Provider,
	})

	filterHostClusters := make([]*proto.Cluster, 0)

	condStatus := operator.NewLeafCondition(operator.Ne, operator.M{"status": common.StatusDeleted})
	branchCond := operator.NewBranchCondition(operator.And, condCluster, condStatus)

	// region filter
	clusterList, err := model.ListCluster(context.Background(), branchCond, &storeopt.ListOption{})
	if err != nil {
		return "", err
	}
	// version filter
	for i := range clusterList {
		if clusterList[i].GetClusterBasicSettings().Version == filter.Version {
			filterHostClusters = append(filterHostClusters, clusterList[i])
		}
	}

	if len(filterHostClusters) == 0 {
		return "", fmt.Errorf("region[%s]无可用匹配的共享集群列表,请联系管理员", filter.Region)
	}

	rand.Seed(time.Now().Unix())                                                 // nolint
	return filterHostClusters[rand.Intn(len(filterHostClusters))].ClusterID, nil // nolint
}

// GetAllMasterIPs get cluster masterIPs
func GetAllMasterIPs(model store.ClusterManagerModel) map[string]ClusterInfo {
	clusterStatus := []string{common.StatusInitialization, common.StatusRunning,
		common.StatusDeleting, common.StatusDeleteClusterFailed, common.StatusCreateClusterFailed}
	condStatus := operator.NewLeafCondition(operator.In, operator.M{"status": clusterStatus})

	clusterList, err := model.ListCluster(context.Background(), condStatus, &storeopt.ListOption{All: true})
	if err != nil {
		blog.Errorf("getAllIPList ListCluster failed: %v", err)
		return nil
	}

	ipListInfo := make(map[string]ClusterInfo)
	for i := range clusterList {
		for ip := range clusterList[i].Master {
			ipListInfo[ip] = ClusterInfo{
				ClusterName: clusterList[i].ClusterName,
				ClusterID:   clusterList[i].ClusterID,
			}
		}
	}

	return ipListInfo
}

// getAllIPList get mongo all IPList
func getAllIPList(provider string, model store.ClusterManagerModel) map[string]struct{} {
	condProd := operator.NewLeafCondition(operator.Eq, operator.M{"provider": provider})
	clusterStatus := []string{common.StatusInitialization, common.StatusRunning, common.StatusDeleting}
	condStatus := operator.NewLeafCondition(operator.In, operator.M{"status": clusterStatus})
	cond := operator.NewBranchCondition(operator.And, condProd, condStatus)
	clusterList, err := model.ListCluster(context.Background(), cond, &storeopt.ListOption{All: true})
	if err != nil {
		blog.Errorf("getAllIPList ListCluster failed: %v", err)
		return nil
	}

	ipList := make(map[string]struct{})
	for i := range clusterList {
		for ip := range clusterList[i].Master {
			ipList[ip] = struct{}{}
		}
	}

	condIP := make(operator.M)
	condIP["status"] = common.StatusRunning
	condP := operator.NewLeafCondition(operator.Eq, condIP)
	nodes, err := model.ListNode(context.Background(), condP, &storeopt.ListOption{All: true})
	if err != nil {
		blog.Errorf("getAllIPList ListNode failed: %v", err)
		return nil
	}

	for i := range nodes {
		ipList[nodes[i].InnerIP] = struct{}{}
	}

	return ipList
}

// GetProjectCommonClustersPerm get userPerm for cluster
func GetProjectCommonClustersPerm(clusterList []string) (map[string]*spb.Struct, error) {
	// trans result for adapt front
	v3ResultPerm := make(map[string]*spb.Struct)
	for _, clsID := range clusterList {
		actionPerm, err := spb.NewStruct(auth.GetV3SharedClusterPerm())
		if err != nil {
			return nil, err
		}

		v3ResultPerm[clsID] = actionPerm
	}

	return v3ResultPerm, nil
}

// PermRequest for perm request
type PermRequest struct {
	// ResourceCode (clusterID)
	ResourceCode string
	// ResourceType (cluster_test/cluster_prod)
	ResourceType string
	// PolicyCode (create/view/delete/use)
	PolicyCode string
}

// CheckUseNodesPermForUser check user use nodes permission
func CheckUseNodesPermForUser(businessID string, user string, nodes []string) bool {
	bizID, err := strconv.Atoi(businessID)
	if err != nil {
		errMsg := fmt.Errorf("strconv BusinessID to int failed: %v", err)
		blog.Errorf(errMsg.Error())
		return false
	}
	return canUseHosts(bizID, user, nodes)
}

// checkUserHasPerm check user has perm
func checkUserHasPerm(businessID string, user string) bool {
	bizID, err := strconv.Atoi(businessID)
	if err != nil {
		errMsg := fmt.Errorf("strconv BusinessID to int failed: %v", err)
		blog.Errorf(errMsg.Error())
		return false
	}

	// query biz hosts
	businessData, err := cmdb.GetCmdbClient().GetBusinessMaintainer(bizID)
	if err != nil {
		blog.Errorf("getUserHasPermHosts GetBusinessMaintainer failed: %v", err)
		return false
	}

	return utils.StringInSlice(user, strings.Split(businessData.BKBizMaintainer, ","))
}

func canUseHosts(bizID int, user string, hostIPList []string) bool {
	hasPermHosts := getUserHasPermHosts(bizID, user)

	hostAllString := sets.NewString(hasPermHosts...)
	return hostAllString.HasAll(hostIPList...)
}

func getUserHasPermHosts(bizID int, user string) []string {
	hostIPs := make([]string, 0)
	// query biz hosts
	businessData, err := cmdb.GetCmdbClient().GetBusinessMaintainer(bizID)
	if err != nil {
		blog.Errorf("getUserHasPermHosts GetBusinessMaintainer failed: %v", err)
		return nil
	}
	maintainers := strings.Split(businessData.BKBizMaintainer, ",")

	// 如果是业务运维，查询全量主机
	if utils.StringInSlice(user, maintainers) {
		var hostList []cmdb.HostData
		hostList, err = cmdb.GetCmdbClient().FetchAllHostsByBizID(bizID, false)
		if err != nil {
			blog.Errorf("getUserHasPermHosts FetchAllHostsByBizID failed: %v", err)
			return nil
		}
		for i := range hostList {
			hostIPs = append(hostIPs, hostList[i].BKHostInnerIP)
		}

		return hostIPs
	}

	// 查询有主机负责人权限的主机
	hostList, err := cmdb.GetCmdbClient().FetchAllHostsByBizID(bizID, false)
	if err != nil {
		blog.Errorf("getUserHasPermHosts FetchAllHostsByBizID failed: %v", err)
		return nil
	}
	for i := range hostList {
		mainOperators := strings.Split(hostList[i].Operator, ",")
		bakOperators := strings.Split(hostList[i].BKBakOperator, ",")

		var hostUsers = make([]string, 0)
		if len(mainOperators) > 0 {
			hostUsers = append(hostUsers, mainOperators...)
		}
		if len(bakOperators) > 0 {
			hostUsers = append(hostUsers, bakOperators...)
		}
		if utils.StringInSlice(user, hostUsers) {
			hostIPs = append(hostIPs, hostList[i].BKHostInnerIP)
		}
	}

	return hostIPs
}

// importClusterExtraOperation extra operation (1. v0 perm register cluster resource 2. sync cluster to pass-cc)
func importClusterExtraOperation(cluster *proto.Cluster) {
	// sync cluster/cluster-snap info to pass-cc
	err := passcc.GetCCClient().CreatePassCCCluster(cluster)
	if err != nil {
		blog.Errorf("importClusterExtraOperation[%s] CreatePassCCCluster failed: %v",
			cluster.ClusterID, err)
	}
	err = passcc.GetCCClient().CreatePassCCClusterSnapshoot(cluster)
	if err != nil {
		blog.Errorf("importClusterExtraOperation[%s] CreatePassCCClusterSnapshoot failed: %v",
			cluster.ClusterID, err)
	}
}

// updatePassCCClusterInfo update cc clusterInfo when update cm cluster
func updatePassCCClusterInfo(cluster *proto.Cluster) {
	err := passcc.GetCCClient().UpdatePassCCCluster(cluster)
	if err != nil {
		blog.Errorf("updatePassCCClusterInfo[%s] failed: %v", cluster.ClusterID, err)
		return
	}

	blog.V(4).Infof("updatePassCCClusterInfo[%s] successful", cluster.ClusterID)
}

func deleteClusterExtraOperation(cluster *proto.Cluster) {
	// sync delete clusterInfo info to pass-cc
	err := passcc.GetCCClient().DeletePassCCCluster(cluster.ProjectID, cluster.ClusterID)
	if err != nil {
		blog.Errorf("deleteClusterExtraOperation DeletePassCCCluster[%s] failed: %v", cluster.ClusterID, err)
		return
	}

	blog.V(4).Infof("deleteClusterExtraOperation DeletePassCCCluster[%s] successful", cluster.ClusterID)
}

// importClusterNode record cluster node
func importClusterNode(model store.ClusterManagerModel, node *proto.Node) error {
	dstCls, err := model.GetNodeByIP(context.Background(), node.InnerIP)
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return err
	}

	// db not exist cluster
	if dstCls == nil {
		err = model.CreateNode(context.Background(), node)
		if err != nil {
			return err
		}

		return nil
	}

	err = model.UpdateNode(context.Background(), node)
	if err != nil {
		return err
	}

	return nil
}

// importClusterData record cluster data
func importClusterData(model store.ClusterManagerModel, cls *proto.Cluster) error {
	if cls.ClusterID == "" {
		err := model.CreateCluster(context.Background(), cls)
		if err != nil {
			return err
		}

		return nil
	}

	dstCls, err := model.GetCluster(context.Background(), cls.ClusterID)
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return err
	}

	// db not exist cluster
	if dstCls == nil {
		err = model.CreateCluster(context.Background(), cls)
		if err != nil {
			return err
		}

		return nil
	}

	err = model.UpdateCluster(context.Background(), cls)
	if err != nil {
		return err
	}

	return nil
}

// deleteClusterCredentialInfo sync delete cluster credential
func deleteClusterCredentialInfo(store store.ClusterManagerModel, clusterID string) {
	err := store.DeleteClusterCredential(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("deleteClusterCredentialInfo[%s] failed: %v", clusterID, err)
		return
	}

	blog.V(4).Infof("deleteClusterCredentialInfo[%s] successful", clusterID)
}

func removeNodeSensitiveInfo(nodes []*proto.Node) {
	for i := range nodes {
		nodes[i].Passwd = ""
	}
}

func transNodeToClusterNode(model store.ClusterManagerModel, node *proto.Node) *proto.ClusterNode {
	var (
		nodeGroupName = ""
	)
	if node.NodeGroupID != "" {
		group, err := model.GetNodeGroup(context.Background(), node.NodeGroupID)
		if err != nil {
			blog.Warnf("transNodeToClusterNode GetNodeGroup[%s] failed: %v", node.NodeGroupID, err)
		} else {
			nodeGroupName = group.Name
		}
	}

	return &proto.ClusterNode{
		NodeID:        node.NodeID,
		InnerIP:       node.InnerIP,
		InstanceType:  node.InstanceType,
		CPU:           node.CPU,
		Mem:           node.Mem,
		GPU:           node.GPU,
		Status:        node.Status,
		ZoneID:        node.ZoneID,
		NodeGroupID:   node.NodeGroupID,
		ClusterID:     node.ClusterID,
		VPC:           node.VPC,
		Region:        node.Region,
		Passwd:        node.Passwd,
		Zone:          node.Zone,
		DeviceID:      node.DeviceID,
		NodeName:      node.NodeName,
		NodeGroupName: nodeGroupName,
		InnerIPv6:     node.InnerIPv6,
		TaskID:        node.TaskID,
		ZoneName:      node.ZoneName,
		FailedReason:  node.FailedReason,
	}
}

// transK8sNodesToClusterNodes parse master nodes to cluster nodes
func transK8sNodesToClusterNodes(clusterID string, k8sNodes []*corev1.Node) []*proto.ClusterNode {
	nodes := make([]*proto.ClusterNode, 0)

	k8sNodesMap := make(map[string]*corev1.Node, 0)
	for i := range k8sNodes {
		k8sNodesMap[k8sNodes[i].Name] = k8sNodes[i]
	}

	for name, node := range k8sNodesMap {
		ipv4, ipv6 := getNodeDualAddress(node)

		nodes = append(nodes, &proto.ClusterNode{
			InnerIP:   ipv4,
			Status:    transNodeStatus(common.StatusRunning, node),
			ClusterID: clusterID,
			NodeName:  name,
			Labels:    node.Labels,
			Taints:    actions.K8sTaintToTaint(node.Spec.Taints),
			UnSchedulable: func(u bool) uint32 {
				if u {
					return 1
				}
				return 0
			}(node.Spec.Unschedulable),
			InnerIPv6: ipv6,
		})
	}

	return nodes
}

// 转换节点状态
func transNodeStatus(cmNodeStatus string, k8sNode *corev1.Node) string {
	if cmNodeStatus == common.StatusInitialization || cmNodeStatus == common.StatusAddNodesFailed ||
		cmNodeStatus == common.StatusDeleting || cmNodeStatus == common.StatusRemoveNodesFailed ||
		cmNodeStatus == common.StatusRemoveCANodesFailed {
		return cmNodeStatus
	}
	for _, v := range k8sNode.Status.Conditions {
		if v.Type != corev1.NodeReady {
			continue
		}
		if v.Status == corev1.ConditionTrue {
			if k8sNode.Spec.Unschedulable {
				return common.StatusNodeRemovable
			}
			return common.StatusRunning
		}
		return common.StatusNodeNotReady
	}

	return common.StatusNodeUnknown
}

// mergeClusterNodes merge k8s nodes and db nodes
// 1. 集群中不存在的节点，并且在cluster manager中状态处于初始化中、初始化失败、移除中、移除失败状态时，需要展示cluster manager中数据
// 2. 集群中存在的节点，则以集群中为准，注意状态的转换
// 3. 适配双栈, 通过nodeName 作为唯一值, 当前数据库nodeName 可能为空, 因此需要适配转换
func mergeClusterNodes(cluster *proto.Cluster, cmNodes []*proto.ClusterNode, k8sNodes []*corev1.Node) []*proto.ClusterNode { // nolint
	// cnNodes exist in k8s cluster and get nodeName
	GetCmNodeNames(cmNodes, k8sNodes)

	cmNodesMap := make(map[string]*proto.ClusterNode, 0)
	k8sNodesMap := make(map[string]*corev1.Node, 0)
	for i := range cmNodes {
		cmNodesMap[cmNodes[i].NodeName] = cmNodes[i]
	}
	for i := range k8sNodes {
		k8sNodesMap[k8sNodes[i].Name] = k8sNodes[i]
	}

	// 处理在 cluster manager 中的节点，但是状态为非正常状态数据，非正常数据放在列表前面
	nodes := make([]*proto.ClusterNode, 0)
	for _, v := range cmNodes {
		if _, ok := k8sNodesMap[v.NodeName]; ok {
			continue
		}
		if v.Status == common.StatusRunning || v.Status == common.StatusNodeRemovable {
			continue
		}
		nodes = append(nodes, v)
	}

	nodes2 := make([]*proto.ClusterNode, 0)
	// 集群中存在的节点，则以集群中为准
	for name, node := range k8sNodesMap {
		ipv4, ipv6 := getNodeDualAddress(node)
		// 集群中存在节点且存在cm数据库, 提取其他信息
		if n, ok := cmNodesMap[name]; ok {
			nodes2 = append(nodes2, &proto.ClusterNode{
				NodeID:       n.NodeID,
				InnerIP:      ipv4,
				InstanceType: n.InstanceType,
				CPU:          n.CPU,
				Mem:          n.Mem,
				GPU:          n.GPU,
				Status:       transNodeStatus(n.Status, node),
				ZoneID:       n.ZoneID,
				NodeGroupID:  n.NodeGroupID,
				ClusterID:    n.ClusterID,
				VPC:          n.VPC,
				Region:       n.Region,
				Passwd:       n.Passwd,
				Zone:         n.Zone,
				DeviceID:     n.DeviceID,
				NodeName:     node.Name,
				Labels:       node.Labels,
				Taints:       actions.K8sTaintToTaint(node.Spec.Taints),
				UnSchedulable: func(u bool) uint32 {
					if u {
						return 1
					}
					return 0
				}(node.Spec.Unschedulable),
				InnerIPv6:     ipv6,
				NodeGroupName: n.NodeGroupName,
				Annotations:   node.Annotations,
				ZoneName: func() string {
					if autils.IsKubeConfigImportCluster(cluster) {
						return ""
					}
					if n.ZoneName != "" {
						return n.ZoneName
					}
					zoneName, ok := node.Labels[utils.ZoneTopologyFlag]
					if ok {
						return zoneName
					}
					return ""
				}(),
				FailedReason: n.FailedReason,
			})
		} else {
			nodes2 = append(nodes2, &proto.ClusterNode{
				InnerIP:   ipv4,
				Status:    transNodeStatus("", node),
				ClusterID: cluster.ClusterID,
				NodeName:  node.Name,
				Labels:    node.Labels,
				Taints:    actions.K8sTaintToTaint(node.Spec.Taints),
				UnSchedulable: func(u bool) uint32 {
					if u {
						return 1
					}
					return 0
				}(node.Spec.Unschedulable),
				InnerIPv6:   ipv6,
				Annotations: node.Annotations,
				ZoneName: func() string {
					if autils.IsKubeConfigImportCluster(cluster) {
						return ""
					}

					zoneName, ok := node.Labels[utils.ZoneTopologyFlag]
					if ok {
						return zoneName
					}
					return ""
				}(),
				ZoneID: func() string {
					if autils.IsKubeConfigImportCluster(cluster) {
						return ""
					}
					zoneName, ok := node.Labels[utils.ZoneTopologyFlag]
					if ok {
						return zoneName
					}
					return ""
				}(),
				Region: func() string {
					region, ok := node.Labels[utils.RegionLabelKey]
					if ok {
						return region
					}
					return ""
				}(),
				InstanceType: func() string {
					insType, ok := node.Labels[utils.NodeInstanceTypeFlag]
					if ok {
						return insType
					}
					return ""
				}(),
			})
		}
	}
	sort.Sort(utils.NodeSlice(nodes2))
	nodes = append(nodes, nodes2...)
	return nodes
}

// GetCmNodeNames get node name
func GetCmNodeNames(cmNodes []*proto.ClusterNode, k8sNodes []*corev1.Node) {
	for i := range cmNodes {
		ipv4 := cmNodes[i].InnerIP
		ipv6 := cmNodes[i].InnerIPv6

		if ipv4 == "" && ipv6 == "" {
			continue
		}

		for _, node := range k8sNodes {
			ipv4s, ipv6s := utils.GetNodeIPAddress(node)
			if utils.StringInSlice(ipv4, ipv4s) || utils.StringInSlice(ipv6, ipv6s) {
				cmNodes[i].NodeName = node.Name
			}
		}
	}
}

func getNodeDualAddress(node *corev1.Node) (string, string) {
	ipv4s, ipv6s := utils.GetNodeIPAddress(node)
	return utils.SliceToString(ipv4s), utils.SliceToString(ipv6s)
}

// asyncDeleteImportedClusterInfo async delete depend info, because deleteWatchComponent need to sync wait
func asyncDeleteImportedClusterInfo(ctx context.Context, store store.ClusterManagerModel, cluster *proto.Cluster) {
	ctx = cloudprovider.WithTaskIDForContext(ctx,
		fmt.Sprintf("asyncDeleteImportedClusterInfo:%s", cluster.ClusterID))

	if options.GetEditionInfo().IsEnterpriseEdition() || options.GetEditionInfo().IsCommunicationEdition() {
		err := provider.DeleteWatchComponentByHelm(ctx, cluster.ProjectID, cluster.ClusterID,
			options.GetGlobalCMOptions().ComponentDeploy.Watch.ReleaseNamespace)
		if err != nil {
			blog.Errorf("asyncDeleteImportedClusterInfo DeleteWatchComponentByHelm[%s] failed: %v",
				cluster.ClusterID, err)
		} else {
			blog.Infof("asyncDeleteImportedClusterInfo DeleteWatchComponentByHelm[%s] successful",
				cluster.ClusterID)
		}
	}
	if options.GetGlobalCMOptions().ComponentDeploy.ImagePullSecret.AddonName != "" {
		err := provider.DeleteImagePullSecretByAddon(ctx, cluster.ProjectID, cluster.ClusterID,
			options.GetGlobalCMOptions().ComponentDeploy.ImagePullSecret.AddonName)
		if err != nil {
			blog.Errorf("asyncDeleteImportedClusterInfo DeleteImagePullSecretByAddon[%s] failed: %v",
				cluster.ClusterID, err)
		} else {
			blog.Infof("asyncDeleteImportedClusterInfo DeleteImagePullSecretByAddon[%s] successful",
				cluster.ClusterID)
		}
	}

	deleteClusterExtraOperation(cluster)
	deleteClusterCredentialInfo(store, cluster.ClusterID)
}

// IsSupportAutoScale support autoscale feat
func IsSupportAutoScale(store store.ClusterManagerModel, cls *proto.Cluster) bool {
	cloudId := cls.GetProvider()
	if cloudId == "" {
		return false
	}

	cloud, err := store.GetCloud(context.Background(), cloudId)
	if err != nil || cloud == nil {
		return false
	}

	if cloud.GetConfInfo().GetDisableNodeGroup() {
		return false
	}

	if cls.ClusterType == common.ClusterTypeVirtual {
		return false
	}

	if cls.ClusterCategory == common.Importer && cls.ImportCategory == common.KubeConfigImport {
		return false
	}

	if cls.Provider == common.GcpCloudProvider && cls.ExtraInfo[api.GKEClusterType] == api.Autopilot {
		return false
	}

	return true
}

func shieldClusterInfo(cluster *proto.Cluster) *proto.Cluster {
	if cluster != nil {
		cluster.KubeConfig = ""
	}

	return cluster
}

// GetClusterStatusNodes get cluster status nodes
func GetClusterStatusNodes(
	store store.ClusterManagerModel, cls *proto.Cluster, status []string) ([]*proto.Node, error) {
	clusterCond := operator.NewLeafCondition(operator.Eq, operator.M{"clusterid": cls.ClusterID})
	statusCond := operator.NewLeafCondition(operator.In, operator.M{"status": status})
	cond := operator.NewBranchCondition(operator.And, clusterCond, statusCond)

	nodes, err := store.ListNode(context.Background(), cond, &storeopt.ListOption{})
	if err != nil {
		blog.Errorf("get Cluster %s all Nodes failed when AddNodesToCluster, %s", cls.ClusterID, err.Error())
		return nil, err
	}

	return nodes, nil
}

const (
	master = "master"
	worker = "worker"
)

func updateClusterModule(cluster *proto.Cluster, category, moduleID, moduleName string) {
	if cluster.GetClusterBasicSettings().GetModule() == nil {
		cluster.GetClusterBasicSettings().Module = &proto.ClusterModule{}
	}
	if moduleID == "" {
		return
	}

	bkBizID, _ := strconv.Atoi(cluster.GetBusinessID())
	bkModuleID, _ := strconv.Atoi(moduleID)

	switch category {
	case master:
		cluster.GetClusterBasicSettings().GetModule().MasterModuleID = moduleID
		if moduleName != "" {
			cluster.GetClusterBasicSettings().GetModule().MasterModuleName = moduleName
			break
		}

		cluster.GetClusterBasicSettings().GetModule().MasterModuleName = cloudprovider.GetModuleName(bkBizID, bkModuleID)
	case worker:
		cluster.GetClusterBasicSettings().GetModule().WorkerModuleID = moduleID
		if moduleName != "" {
			cluster.GetClusterBasicSettings().GetModule().WorkerModuleName = moduleName
			break
		}
		cluster.GetClusterBasicSettings().GetModule().WorkerModuleName = cloudprovider.GetModuleName(bkBizID, bkModuleID)
	default:
	}
}

func updateAutoScalingModule(cluster *proto.Cluster, option *proto.ClusterAutoScalingOption,
	moduleID, moduleName string) {
	if option.GetModule() == nil {
		option.Module = &proto.ModuleInfo{}
	}
	if moduleID == "" {
		return
	}
	bkBizID, _ := strconv.Atoi(cluster.GetBusinessID())
	bkModuleID, _ := strconv.Atoi(moduleID)
	option.Module.ScaleOutBizID = cluster.GetBusinessID()
	option.Module.ScaleOutModuleID = moduleID

	if moduleName != "" {
		option.Module.ScaleOutModuleName = moduleName
	} else {
		option.Module.ScaleOutModuleName = cloudprovider.GetModuleName(bkBizID, bkModuleID)
	}
}

func transClusterNodes(model store.ClusterManagerModel, cluster *proto.Cluster, category, moduleID string) error {
	bkBizID, _ := strconv.Atoi(cluster.GetBusinessID())
	bkModuleID, _ := strconv.Atoi(moduleID)

	var (
		err     error
		nodeIps []string
		nodes   []*proto.ClusterNode
	)

	switch category {
	case master:
		nodes, err = getClusterMasterNodes(model, cluster)
		if err != nil {
			blog.Errorf("transClusterNodes[%s] getClusterMasterNodes failed: %v", cluster.ClusterID, err)
			return err
		}
	case worker:
		nodes, err = getClusterNodes(model, cluster)
		if err != nil {
			blog.Errorf("transClusterNodes[%s] getClusterNodes failed: %v", cluster.ClusterID, err)
			return err
		}
	default:
		return fmt.Errorf("not support %s", category)
	}

	nodeIps = getCmNodeIps(nodes)
	if len(nodeIps) == 0 {
		blog.Errorf("transClusterNodes[%s] nodeIps empty", cluster.ClusterID)
		return nil
	}

	err = provider.TransBizNodeModule(context.Background(), bkBizID, bkModuleID, nodeIps)
	if err != nil {
		blog.Errorf("transClusterNodes[%s] TransBizNodeModule failed: %v", cluster.ClusterID, err)
		return err
	}

	return nil
}

func getClusterNodes(model store.ClusterManagerModel, cls *proto.Cluster) ([]*proto.ClusterNode, error) {
	condM := make(operator.M)
	condM["clusterid"] = cls.ClusterID
	cond := operator.NewLeafCondition(operator.Eq, condM)
	nodes, err := model.ListNode(context.Background(), cond, &storeopt.ListOption{})
	if err != nil {
		blog.Errorf("list nodes in cluster %s failed, %s", cls.ClusterID, err.Error())
		return nil, err
	}

	cmNodes := make([]*proto.ClusterNode, 0)
	for i := range nodes {
		cmNodes = append(cmNodes, transNodeToClusterNode(model, nodes[i]))
	}

	return cmNodes, nil
}

// getClusterNodes for get cluster nodes
func getClusterMasterNodes(model store.ClusterManagerModel, cls *proto.Cluster) ([]*proto.ClusterNode, error) {
	if cls == nil || len(cls.GetMaster()) == 0 {
		return nil, fmt.Errorf("cluster %s master nodes empty", cls.ClusterID)
	}

	cmNodes := make([]*proto.ClusterNode, 0)
	for i := range cls.GetMaster() {
		cmNodes = append(cmNodes, transNodeToClusterNode(model, cls.GetMaster()[i]))
	}

	return cmNodes, nil
}

// getCmNodeIps for get cm node ips
func getCmNodeIps(cmNodes []*proto.ClusterNode) []string {
	var ips = make([]string, 0)
	for i := range cmNodes {
		if cmNodes[i].InnerIP == "" {
			continue
		}
		ips = append(ips, cmNodes[i].InnerIP)
	}
	return ips
}

// checkHighAvailabilityMasterNodes for check master node number and zoneID
func checkHighAvailabilityMasterNodes(cls *proto.Cluster, cloud *proto.Cloud, nodes []*proto.Node) error {
	clsMgr, err := cloudprovider.GetClusterMgr(cloud.GetCloudProvider())
	if err != nil {
		blog.Errorf("checkHighAvailabilityMasterNodes[%s] failed: %v", cls.ClusterID, err)
		return err
	}

	return clsMgr.CheckHighAvailabilityMasterNodes(cls, nodes, &cloudprovider.CheckHaMasterNodesOption{Cloud: cloud})
}

// clusterToClusterBasicInfo for convert cluster to clusterBasicInfo
func clusterToClusterBasicInfo(cluster *proto.Cluster) *proto.ClusterBasicInfo {
	return &proto.ClusterBasicInfo{
		ClusterID:       cluster.ClusterID,
		ClusterName:     cluster.ClusterName,
		Provider:        cluster.Provider,
		Region:          cluster.Region,
		VpcID:           cluster.VpcID,
		ProjectID:       cluster.ProjectID,
		BusinessID:      cluster.BusinessID,
		Environment:     cluster.Environment,
		EngineType:      cluster.EngineType,
		ClusterType:     cluster.ClusterType,
		Labels:          cluster.Labels,
		Creator:         cluster.Creator,
		CreateTime:      cluster.CreateTime,
		UpdateTime:      cluster.UpdateTime,
		SystemID:        cluster.SystemID,
		ManageType:      cluster.ManageType,
		Status:          cluster.Status,
		Updater:         cluster.Updater,
		NetworkType:     cluster.NetworkType,
		ModuleID:        cluster.ModuleID,
		IsCommonCluster: cluster.IsCommonCluster,
		Description:     cluster.Description,
		ClusterCategory: cluster.ClusterCategory,
		IsShared:        cluster.IsShared,
	}
}
