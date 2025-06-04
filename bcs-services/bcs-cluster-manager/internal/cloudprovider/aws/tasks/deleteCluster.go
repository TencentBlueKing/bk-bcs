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

package tasks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"
	"github.com/aws/aws-sdk-go/service/eks"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/aws/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
)

// DeleteEKSClusterTask delete cluster task
func DeleteEKSClusterTask(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("DeleteEKSClusterTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("DeleteEKSClusterTask GetClusterDependBasicInfo failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = deleteEKSCluster(ctx, dependInfo)
	if err != nil {
		blog.Errorf("DeleteEKSClusterTask[%s]: deleteEKSCluster failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("DeleteEKSClusterTask deleteEKSCluster failed: %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("DeleteEKSClusterTask[%s]: %s update to storage fatal", taskID, stepName)
		return err
	}
	return nil
}

func deleteEKSCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	cluster := info.Cluster

	if cluster.GetSystemID() == "" {
		return nil
	}

	// get aws client
	cli, err := api.NewEksClient(info.CmOption)
	if err != nil {
		blog.Errorf("deleteEKSCluster[%s]: get aws client for cluster[%s] failed, %s", taskID, cluster.ClusterID,
			err.Error())
		return fmt.Errorf("get aws client failed, %s", err.Error())
	}

	// check cluster if exist
	cloudCluster, err := cli.GetEksCluster(cluster.SystemID)
	if err != nil {
		if strings.Contains(err.Error(), eks.ErrCodeResourceNotFoundException) {
			return nil
		}
		return err
	}

	// check cluster if node group exist, and batch delete nodegroups first, or the cluster can't be deleted
	ngList, err := cli.ListNodegroups(cluster.SystemID)
	if err != nil {
		blog.Errorf("deleteEKSCluster[%s]: call aws ListNodegroups failed: %v", taskID, err)
		return fmt.Errorf("call aws ListNodegroups failed: %s", err.Error())
	}
	for _, ng := range ngList {
		err = retry.Do(func() error {
			_, err = cli.DeleteNodegroup(&eks.DeleteNodegroupInput{
				NodegroupName: ng,
				ClusterName:   &cluster.SystemID})
			if err != nil {
				return err
			}
			return nil
		}, retry.Attempts(3))
		if err != nil {
			blog.Errorf("deleteEKSCluster[%s] DeleteNodegroup[%s] failed: %v", taskID, *ng, err)
			return err
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	err = loop.LoopDoFunc(ctx, func() error {
		ngList, err = cli.ListNodegroups(cluster.SystemID)
		if err != nil {
			blog.Errorf("deleteEKSCluster[%s]: call aws ListNodegroups failed: %v", taskID, err)
			return fmt.Errorf("call aws ListNodegroups failed: %s", err.Error())
		}

		if len(ngList) != 0 {
			blog.Infof("deleteEKSCluster[%s] %d nodegroups to be deleted", taskID, len(ngList))
			return nil
		}

		blog.Infof("deleteEKSCluster[%s] all nodegroups have been deleted", taskID)
		return loop.EndLoop
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("deleteEKSCluster[%s] delete nodegroups failed: %v", taskID, err)
		return err
	}

	// delete cluster
	_, err = cli.DeleteEksCluster(cluster.SystemID)
	if err != nil {
		blog.Errorf("deleteEKSCluster[%s]: call aws DeleteEKSCluster failed: %v", taskID, err)
		return fmt.Errorf("call aws DeleteEKSCluster failed: %s", err.Error())
	}

	clusterCtx, clusterCancel := context.WithTimeout(ctx, 5*time.Minute)
	defer clusterCancel()
	err = loop.LoopDoFunc(clusterCtx, func() error {
		_, retErr := cli.GetEksCluster(cluster.SystemID)
		if retErr == nil {
			blog.Infof("deleteEKSCluster[%s] %s cluster is being deleted", taskID, cluster.SystemID)
			return nil
		}

		// 集群已经被删除
		if strings.Contains(retErr.Error(), eks.ErrCodeResourceNotFoundException) {
			return loop.EndLoop
		}

		blog.Errorf("deleteEKSCluster[%s]: call aws GetEksCluster failed: %v", taskID, retErr)
		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("deleteEKSCluster[%s] delete cluster failed: %v", taskID, err)
		return err
	}

	ecCli, err := api.NewEC2Client(info.CmOption)
	if err != nil {
		blog.Errorf("deleteEKSCluster[%s]: get aws ec client for cluster[%s] failed, %s", taskID, cluster.ClusterID,
			err.Error())
		return fmt.Errorf("get aws ec client failed, %s", err.Error())
	}

	// 删除创建集群时自动创建的子网
	for _, subnetId := range cloudCluster.ResourcesVpcConfig.SubnetIds {
		err = retry.Do(func() error {
			if errLocal := ecCli.DeleteSubnet(subnetId); errLocal != nil {
				return errLocal
			}

			return nil
		}, retry.Attempts(3), retry.DelayType(retry.FixedDelay), retry.Delay(time.Second))
		if err != nil {
			blog.Errorf("deleteEKSCluster[%s] delete cluster subnet %s failed: %v", taskID, *subnetId, err)
		}

		blog.Infof("deleteEKSCluster[%s] delete cluster subnet %s successful", taskID, *subnetId)
	}

	return nil
}

// CleanClusterDBInfoTask clean cluster DB info
func CleanClusterDBInfoTask(taskID string, stepName string) error {
	// delete node && nodeGroup && cluster
	// get relative nodes by clusterID
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: get cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("get cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// delete nodes
	err = cloudprovider.GetStorageModel().DeleteNodesByClusterID(context.Background(), cluster.ClusterID)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: delete nodes for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("delete node for %s failed, %s", clusterID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CleanClusterDBInfoTask[%s]: delete nodes for cluster[%s] in DB successful", taskID, clusterID)

	// delete nodeGroup
	err = cloudprovider.GetStorageModel().DeleteNodeGroupByClusterID(context.Background(), cluster.ClusterID)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: delete nodeGroups for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("delete nodeGroups for %s failed, %s", clusterID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CleanClusterDBInfoTask[%s]: delete nodeGroups for cluster[%s] in DB successful", taskID, clusterID)

	// delete cluster
	cluster.Status = icommon.StatusDeleting
	err = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: delete cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("delete cluster for %s failed, %s", clusterID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CleanClusterDBInfoTask[%s]: delete cluster[%s] in DB successful", taskID, clusterID)

	utils.SyncDeletePassCCCluster(taskID, cluster)
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: %s update to storage fatal", taskID, stepName)
		return err
	}
	return nil
}
