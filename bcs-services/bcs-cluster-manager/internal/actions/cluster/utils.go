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

package cluster

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	spb "google.golang.org/protobuf/types/known/structpb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	provider "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/passcc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
)

type clusterInfo struct {
	clusterName string
	clusterID   string
}

const (
	// Builder self builder cluster
	Builder = "builder"
	// Importer export external cluster
	Importer = "importer"

	// KubeConfig import
	KubeConfig = "kubeConfig"
	// Cloud import
	Cloud = "cloud"

	// Prod environment
	Prod = "prod"
)

const (
	// DefaultImageName default image name
	DefaultImageName = "Tencent Linux Release 2.2 (Final)"
)

const (
	// ManagerCluster manage cluster
	ManagerCluster = "MANAGED_CLUSTER"
	// IndependentCluster independent cluster
	IndependentCluster = "INDEPENDENT_CLUSTER"

	// NodeRoleMaster master nodes
	NodeRoleMaster = "node-role.kubernetes.io/master"
)

// ClusterManageTypeMap cluster manage type
var ClusterManageTypeMap = map[string]struct{}{
	"MANAGED_CLUSTER":     {},
	"INDEPENDENT_CLUSTER": {},
}

// ClusterTypeMap cluster type
var ClusterTypeMap = map[string]struct{}{
	"mesos": {},
	"k8s":   {},
}

// ClusterEnvMap cluster env map
var ClusterEnvMap = map[string]struct{}{
	"stag":  {},
	"debug": {},
	"prod":  {},
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
	_, ok := ClusterTypeMap[clusterType]
	if !ok {
		return 0, fmt.Errorf("clusterType[%s] failed", clusterType)
	}

	_, ok = ClusterEnvMap[env]
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
func getClusterList(model store.ClusterManagerModel) ([]proto.Cluster, error) {
	clusterStatus := []string{common.StatusInitialization, common.StatusRunning, common.StatusDeleting}
	condStatus := operator.NewLeafCondition(operator.In, operator.M{"status": clusterStatus})

	clusterList, err := model.ListCluster(context.Background(), condStatus, &storeopt.ListOption{All: true})
	if err != nil {
		blog.Errorf("getClusterList failed: %v", err)
		return nil, err
	}

	return clusterList, nil
}

// getAllMasterIPs get cluster masterIPs
func getAllMasterIPs(model store.ClusterManagerModel) map[string]clusterInfo {
	clusterList, err := getClusterList(model)
	if err != nil {
		blog.Errorf("getAllIPList ListCluster failed: %v", err)
		return nil
	}

	ipListInfo := make(map[string]clusterInfo)
	for i := range clusterList {
		for ip := range clusterList[i].Master {
			ipListInfo[ip] = clusterInfo{
				clusterName: clusterList[i].ClusterName,
				clusterID:   clusterList[i].ClusterID,
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

// GetUserClusterPermList get user cluster permission
func GetUserClusterPermList(iam iam.PermClient, user actions.PermInfo, clusterList []string) (
	map[string]map[string]interface{}, error) {
	permissions := make(map[string]map[string]interface{})
	clusterPerm := cluster.NewBCSClusterPermClient(iam)

	actionIDs := []string{cluster.ClusterView.String(), cluster.ClusterManage.String(), cluster.ClusterDelete.String()}
	perms, err := clusterPerm.GetMultiClusterMultiActionPermission(user.UserID, user.ProjectID, clusterList, actionIDs)
	if err != nil {
		return nil, err
	}

	for clusterID, perm := range perms {
		if permissions[clusterID] == nil {
			permissions[clusterID] = make(map[string]interface{})
		}
		for action, res := range perm {
			permissions[clusterID][action] = res
		}
	}

	return permissions, nil
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
		hostList, err := cmdb.GetCmdbClient().FetchAllHostsByBizID(bizID)
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
	hostList, err := cmdb.GetCmdbClient().FetchAllHostsByBizID(bizID)
	if err != nil {
		blog.Errorf("getUserHasPermHosts FetchAllHostsByBizID failed: %v", err)
		return nil
	}
	for i := range hostList {
		hostUsers := []string{hostList[i].Operator, hostList[i].BKBakOperator}
		if utils.StringInSlice(user, hostUsers) {
			hostIPs = append(hostIPs, hostList[i].BKHostInnerIP)
		}
	}

	return hostIPs
}

// importClusterExtraOperation extra operation (sync cluster to pass-cc)
func importClusterExtraOperation(cluster *proto.Cluster) {
	// sync cluster/cluster-snap info to pass-cc
	err := passcc.GetCCClient().CreatePassCCCluster(cluster)
	if err != nil {
		blog.Errorf("ImportClusterExtraOperation[%s] CreatePassCCCluster failed: %v",
			cluster.ClusterID, err)
	}
	err = passcc.GetCCClient().CreatePassCCClusterSnapshoot(cluster)
	if err != nil {
		blog.Errorf("ImportClusterExtraOperation CreatePassCCClusterSnapshoot[%s] failed: %v",
			cluster.ClusterID, err)
	}
}

// deleteClusterExtraOperation sync delete pass-cc cluster
func deleteClusterExtraOperation(cluster *proto.Cluster) {
	// sync delete clusterInfo info to pass-cc
	err := passcc.GetCCClient().DeletePassCCCluster(cluster.ProjectID, cluster.ClusterID)
	if err != nil {
		blog.Errorf("deleteClusterExtraOperation DeletePassCCCluster[%s] failed: %v", cluster.ClusterID, err)
		return
	}

	blog.V(4).Infof("deleteClusterExtraOperation DeletePassCCCluster[%s] successful", cluster.ClusterID)
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

// deleteClusterCredentialInfo sync delete cluster credential
func deleteClusterCredentialInfo(store store.ClusterManagerModel, clusterID string) {
	err := store.DeleteClusterCredential(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("deleteClusterCredentialInfo[%s] failed: %v", clusterID, err)
		return
	}

	blog.V(4).Infof("deleteClusterCredentialInfo[%s] successful", clusterID)
}

func asyncDeleteImportedClusterInfo(ctx context.Context, store store.ClusterManagerModel, cluster *proto.Cluster) {
	ctx = cloudprovider.WithTaskIDForContext(ctx,
		fmt.Sprintf("asyncDeleteImportedClusterInfo:%s", cluster.ClusterID))
	err := provider.DeleteWatchComponentByHelm(ctx, cluster.ProjectID, cluster.ClusterID)
	if err != nil {
		blog.Errorf("asyncDeleteImportedClusterInfo DeleteWatchComponentByHelm[%s] failed: %v",
			cluster.ClusterID, err)
	} else {
		blog.Errorf("asyncDeleteImportedClusterInfo DeleteWatchComponentByHelm[%s] successful", cluster.ClusterID)
	}

	deleteClusterExtraOperation(cluster)
	deleteClusterCredentialInfo(store, cluster.ClusterID)
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
	}
}

// 转换节点状态
func transNodeStatus(cmNodeStatus string, k8sNode *corev1.Node) string {
	if cmNodeStatus == common.StatusInitialization || cmNodeStatus == common.StatusAddNodesFailed ||
		cmNodeStatus == common.StatusDeleting || cmNodeStatus == common.StatusRemoveNodesFailed {
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

func filterNodesRole(k8sNodes []*corev1.Node, master bool) []*corev1.Node {
	nodes := make([]*corev1.Node, 0)
	for _, v := range k8sNodes {
		if _, ok := v.Labels[common.MasterRole]; ok == master {
			nodes = append(nodes, v)
		}
	}
	return nodes
}

// mergeClusterNodes merge k8s nodes and db nodes
// 1. 集群中不存在的节点，并且在cluster manager中状态处于初始化中、初始化失败、移除中、移除失败状态时，需要展示cluster manager中数据
// 2. 集群中存在的节点，则以集群中为准，注意状态的转换
// 3. 适配双栈, 通过nodeName 作为唯一值, 当前数据库nodeName 可能为空, 因此需要适配转换
func mergeClusterNodes(clusterID string, cmNodes []*proto.ClusterNode, k8sNodes []*corev1.Node) []*proto.ClusterNode {
	// cnNodes exist in k8s cluster and get nodeName
	GetCmNodeNames(cmNodes, k8sNodes)

	cmNodesMap := make(map[string]*proto.ClusterNode, 0)
	k8sNodesMap := make(map[string]*corev1.Node, 0)

	for i := range cmNodes {
		clusterID = cmNodes[i].ClusterID
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
				InnerIPv6: ipv6,
				NodeGroupName: n.NodeGroupName,
			})
		} else {
			nodes2 = append(nodes2, &proto.ClusterNode{
				InnerIP:   ipv4,
				Status:    transNodeStatus("", node),
				ClusterID: clusterID,
				NodeName:  node.Name,
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
	}
	sort.Sort(NodeSlice(nodes2))
	nodes = append(nodes, nodes2...)
	return nodes
}

// NodeSlice cluster node slice
type NodeSlice []*proto.ClusterNode

func (n NodeSlice) Len() int {
	return len(n)
}

func (n NodeSlice) Less(i, j int) bool {
	return n[i].NodeName < n[j].NodeName
}

func (n NodeSlice) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
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
			ipv4s, ipv6s := getNodeIPAddress(node)
			if utils.StringInSlice(ipv4, ipv4s) || utils.StringInSlice(ipv6, ipv6s) {
				cmNodes[i].NodeName = node.Name
			}
		}
	}
}

func getNodeIPAddress(node *corev1.Node) ([]string, []string) {
	ipv4Address := make([]string, 0)
	ipv6Address := make([]string, 0)

	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			switch {
			case util.IsIPv6(address.Address):
				ipv6Address = append(ipv6Address, address.Address)
			case util.IsIPv4(address.Address):
				ipv4Address = append(ipv4Address, address.Address)
			default:
				blog.Errorf("unsupported ip type")
			}
		}
	}

	return ipv4Address, ipv6Address
}

func getNodeDualAddress(node *corev1.Node) (string, string) {
	ipv4s, ipv6s := getNodeIPAddress(node)
	return utils.SliceToString(ipv4s), utils.SliceToString(ipv6s)
}

func shieldClusterInfo(cluster *proto.Cluster) *proto.Cluster {
	if cluster != nil {
		cluster.KubeConfig = ""
	}

	return cluster
}
