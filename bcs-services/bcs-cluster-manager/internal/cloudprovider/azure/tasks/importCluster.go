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
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// ImportClusterNodesTask call aksInterface or kubeConfig import cluster nodes
func ImportClusterNodesTask(taskID string, stepName string) error {
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

	basicInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("ImportClusterNodesTask[%s]: getClusterDependBasicInfo failed: %v", taskID, err)
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// import cluster instances
	err = importClusterInstances(basicInfo)
	if err != nil {
		blog.Errorf("ImportClusterNodesTask[%s]: importClusterInstances failed: %v", taskID, err)
		retErr := fmt.Errorf("importClusterInstances failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// 导入集群nodeResourceGroup
	if err = importNodeResourceGroup(basicInfo); err != nil {
		blog.Errorf("ImportClusterNodesTask[%s]: importNodeResourceGroup failed: %v", taskID, err)
		retErr := fmt.Errorf("importNodeResourceGroup failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	// 导入vpcID
	if err = importVpcID(basicInfo); err != nil {
		blog.Errorf("ImportClusterNodesTask[%s]: importVpcID failed: %v", taskID, err)
		retErr := fmt.Errorf("importVpcID failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update cluster masterNodes info
	err = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), basicInfo.Cluster)
	if err != nil {
		return err
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("ImportClusterNodesTask[%s] %s update to storage fatal", taskID, stepName)
		return err
	}
	return nil
}

// RegisterClusterKubeConfigTask register cluster kubeConfig connection
func RegisterClusterKubeConfigTask(taskID string, stepName string) error {
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

	basicInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("RegisterClusterKubeConfigTask[%s]: getClusterDependBasicInfo failed: %v", taskID, err)
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	err = importClusterCredential(ctx, basicInfo)
	if err != nil {
		blog.Errorf("RegisterClusterKubeConfigTask[%s]: importClusterCredential failed: %v", taskID, err)
		retErr := fmt.Errorf("importClusterCredential failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RegisterClusterKubeConfigTask[%s] %s update to storage fatal", taskID, stepName)
		return err
	}
	return nil
}

func importClusterCredential(ctx context.Context, data *cloudprovider.CloudDependBasicInfo) error {
	cli, err := api.NewAksServiceImplWithCommonOption(data.CmOption)
	if err != nil {
		return err
	}

	credentials, err := cli.GetClusterAdminCredentials(ctx, data)
	if err != nil {
		return err
	}
	if len(credentials) == 0 {
		return fmt.Errorf("credentials not found")
	}

	// save cluster kubeConfig
	kubeConfig := string(credentials[0].Value)
	data.Cluster.KubeConfig = kubeConfig
	_ = cloudprovider.UpdateCluster(data.Cluster)

	config, err := types.GetKubeConfigFromYAMLBody(false, types.YamlInput{
		YamlContent: kubeConfig,
	})
	if err != nil {
		return err
	}

	err = cloudprovider.UpdateClusterCredentialByConfig(data.Cluster.ClusterID, config)
	if err != nil {
		return err
	}

	return nil
}

func importClusterInstances(data *cloudprovider.CloudDependBasicInfo) error {
	// get cluster nodes
	kubeRet := base64.StdEncoding.EncodeToString([]byte(data.Cluster.KubeConfig))
	kubeCli, err := clusterops.NewKubeClient(kubeRet)
	if err != nil {
		return fmt.Errorf("new kube client failed, %s", err.Error())
	}
	nodes, err := kubeCli.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list nodes failed, %s", err.Error())
	}

	// get container runtime info here
	if len(nodes.Items) > 0 {
		crv := strings.Split(nodes.Items[0].Status.NodeInfo.ContainerRuntimeVersion, "://")
		if len(crv) == 2 {
			data.Cluster.ClusterAdvanceSettings = &proto.ClusterAdvanceSetting{
				ContainerRuntime: crv[0],
				RuntimeVersion:   crv[1],
			}
			err = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), data.Cluster)
			if err != nil {
				blog.Errorf("importClusterInstances update cluster[%s] failed: %v", data.Cluster.ClusterName, err)
			}
		}
	}

	err = importClusterNodesToCM(context.Background(), nodes.Items, data.Cluster.ClusterID)
	if err != nil {
		return err
	}

	return nil
}

// importNodeResourceGroup 导入nodeResourceGroup
func importNodeResourceGroup(info *cloudprovider.CloudDependBasicInfo) error {
	cluster := info.Cluster
	if cluster.ExtraInfo != nil && len(cluster.ExtraInfo[api.NodeResourceGroup]) > 0 {
		return nil
	}
	client, err := api.NewAksServiceImplWithCommonOption(info.CmOption)
	if err != nil {
		return errors.Wrapf(err, "create AksService failed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	managedCluster, err := client.GetCluster(ctx, info)
	if err != nil {
		return errors.Wrapf(err, "call GetCluster falied")
	}
	if cluster.ExtraInfo == nil {
		cluster.ExtraInfo = make(map[string]string)
	}

	cluster.ExtraInfo[api.NodeResourceGroup] = *managedCluster.Properties.NodeResourceGroup
	return nil
}

// importVpcID 导入vpc id
func importVpcID(info *cloudprovider.CloudDependBasicInfo) error {
	cluster := info.Cluster
	client, err := api.NewAksServiceImplWithCommonOption(info.CmOption)
	if err != nil {
		return errors.Wrapf(err, "create AksService failed")
	}

	nodeResourceGroup := cluster.ExtraInfo[api.NodeResourceGroup]
	blog.Infof("importVpcID nodeResourceGroup:%s", nodeResourceGroup)

	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	list, err := client.ListVirtualNetwork(ctx, nodeResourceGroup)
	if err != nil {
		return errors.Wrapf(err, "call ListVirtualNetwork failed")
	}

	// blog.Infof("importVpcID list:%s", toPrettyJsonString(list))
	if len(list) > 0 {
		cluster.VpcID = *list[0].Name
	}

	return nil
}
