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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/passcc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"k8s.io/apimachinery/pkg/util/sets"
)

type clusterInfo struct {
	clusterName string
	clusterID string
}

const (
	// Builder self builder cluster
	Builder = "builder"
	// Importer export external cluster
	Importer = "importer"

	// KubeConfig import
	KubeConfig = "kubeConfig"
	// Cloud import
	Cloud      = "cloud"

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

// getAllMasterIPs get cluster masterIPs
func getAllMasterIPs(model store.ClusterManagerModel) map[string]clusterInfo {
	clusterStatus := []string{common.StatusInitialization, common.StatusRunning, common.StatusDeleting}
	condStatus := operator.NewLeafCondition(operator.In, operator.M{"status": clusterStatus})

	clusterList, err := model.ListCluster(context.Background(), condStatus, &storeopt.ListOption{All: true})
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

// UserInfo for perm request
type UserInfo struct {
	ProjectID string
	UserID    string
}

// GetUserClusterPermList get user cluster permission
func GetUserClusterPermList(user UserInfo, clusterList []string) (map[string]*proto.Permission, error) {
	permissions := make(map[string]*proto.Permission)
	cli := &cluster.BCSClusterPerm{}

	actionIDs := []string{cluster.ClusterView.String(), cluster.ClusterManage.String(), cluster.ClusterDelete.String()}
	perms, err := cli.GetMultiClusterMultiActionPermission(user.UserID, user.ProjectID, clusterList, actionIDs)
	if err != nil {
		return nil, err
	}

	for clusterID, perm := range perms {
		permissions[clusterID] = &proto.Permission{
			Policy: perm,
		}
	}

	return permissions, nil
}

// GetUserPermListByProjectAndCluster get user cluster permissions
func GetUserPermListByProjectAndCluster(user UserInfo, clusterList []string, filterUse bool) (map[string]*proto.Permission, error) {
	permissions := make(map[string]*proto.Permission)

	// policyCode resourceType  clusterList
	for _, clusterID := range clusterList {
		defaultPerm := auth.GetInitPerm(true)

		permissions[clusterID] = &proto.Permission{
			Policy: defaultPerm,
		}
	}

	return permissions, nil
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

// GetClusterCreatePerm get cluster create permission by user
func GetClusterCreatePerm(user UserInfo) map[string]bool {
	permissions := make(map[string]bool)

	// attention: v0 permission only support project
	permissions["test"] = true
	permissions["prod"] = true
	permissions["create"] = true

	return permissions
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

