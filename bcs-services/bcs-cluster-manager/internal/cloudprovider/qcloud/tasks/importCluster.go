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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// ImportClusterNodesTask call tkeInterface or kubeConfig import cluster nodes
func ImportClusterNodesTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start import cluster nodes")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("ImportClusterNodesTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("ImportClusterNodesTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

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

	// update cluster masterNodes info
	err = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), basicInfo.Cluster)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("update cluster failed [%s]", err))
		return err
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"update cluster successful")

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"import cluster instances successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("ImportClusterNodesTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

// RegisterClusterKubeConfigTask register cluster kubeConfig connection
func RegisterClusterKubeConfigTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start register cluster kubeconfig connection")
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("RegisterClusterKubeConfigTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("RegisterClusterKubeConfigTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

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

	// 社区版本 TKE公有云导入获取集群kubeConfig并进行配置
	err = registerTKEClusterEndpoint(ctx, basicInfo, api.ClusterEndpointConfig{
		IsExtranet: func() bool {
			importType, ok := basicInfo.Cluster.GetExtraInfo()[icommon.ImportType]
			if !ok || importType == icommon.ExternalImport {
				return true
			}

			return false
		}(),
	})
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("register tke cluster endpoint failed [%s]", err))
		blog.Errorf("RegisterClusterKubeConfigTask[%s]: getTKEExternalClusterEndpoint failed: %v", taskID, err)
		retErr := fmt.Errorf("getTKEExternalClusterEndpoint failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"register tke cluster endpoint successful")

	err = importClusterCredential(ctx, basicInfo, true, true, "", "")
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("import cluster credential failed [%s]", err))
		blog.Errorf("RegisterClusterKubeConfigTask[%s]: importClusterCredential failed: %v", taskID, err)
		retErr := fmt.Errorf("importClusterCredential failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"import cluster credential successful")

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"register cluster kubeconfig connection successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RegisterClusterKubeConfigTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}
	return nil
}

// registerTKEClusterEndpoint 开启内网或外网访问config: err = nil 已开启内/外网访问; err != nil 开启失败
func registerTKEClusterEndpoint(ctx context.Context, data *cloudprovider.CloudDependBasicInfo,
	config api.ClusterEndpointConfig) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	tkeCli, err := api.NewTkeClient(data.CmOption)
	if err != nil {
		return err
	}

	if data.Cluster.SystemID == "" {
		return fmt.Errorf("taskID[%s] cluster[%s] systemID is null", taskID, data.Cluster.ClusterID)
	}

	endpointStatus, err := tkeCli.GetClusterEndpointStatus(data.Cluster.SystemID, config.IsExtranet)
	if err != nil {
		return fmt.Errorf("taskID[%s] registerTKEClusterEndpoint[%s] failed: %v",
			taskID, data.Cluster.ClusterID, err.Error())
	}

	blog.Infof("taskID[%s] registerTKEClusterEndpoint endpointStatus[%s]",
		taskID, endpointStatus.Status())

	switch {
	case endpointStatus.Created():
		return nil
	case endpointStatus.NotFound(), endpointStatus.Deleted():
		err = tkeCli.CreateClusterEndpoint(data.Cluster.SystemID, config)
		if err != nil {
			return err
		}
		err = checkClusterEndpointStatus(ctx, data, config.IsExtranet)
		if err != nil {
			return err
		}

		return nil
	case endpointStatus.Creating():
		err = checkClusterEndpointStatus(ctx, data, config.IsExtranet)
		if err != nil {
			return err
		}
		return nil
	default:
	}

	return fmt.Errorf("taskID[%s] GetClusterEndpointStatus not support status[%s]", taskID, endpointStatus)
}

func checkClusterEndpointStatus(ctx context.Context, data *cloudprovider.CloudDependBasicInfo, isExtranet bool) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	cli, err := api.NewTkeClient(data.CmOption)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	err = loop.LoopDoFunc(ctx, func() error {
		status, errStatus := cli.GetClusterEndpointStatus(data.Cluster.SystemID, isExtranet)
		if errStatus != nil {
			blog.Errorf("taskID[%s] GetClusterEndpointStatus[%s] failed: %v",
				taskID, data.Cluster.SystemID, errStatus)
			return nil
		}
		switch {
		case status.Creating():
			blog.Infof("taskID[%s] GetClusterEndpointStatus[%s] still creating, status[%s]",
				taskID, data.Cluster.SystemID, status)
			return nil
		case status.Created():
			blog.Infof("taskID[%s] GetClusterEndpointStatus[%s] status[%s]",
				taskID, data.Cluster.SystemID, status)
			return loop.EndLoop
		default:
			return nil
		}
	}, loop.LoopInterval(20*time.Second))
	if err != nil {
		blog.Errorf("taskID[%s] GetClusterEndpointStatus failed: %v", taskID, err)
		return err
	}

	return nil
}

// importClusterCredential import cluster kubeconfig to clustercredential
func importClusterCredential(ctx context.Context, data *cloudprovider.CloudDependBasicInfo,
	isExtranet bool, syncCluster bool, token string, newKubeConfig string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	var (
		kubeConfig string
		err        error
	)

	cli, err := api.NewTkeClient(data.CmOption)
	if err != nil {
		blog.Errorf("importClusterCredential[%s] NewTkeClient failed: %v", taskID, err)
		return err
	}

	if newKubeConfig != "" {
		kubeConfig = newKubeConfig
	} else {
		kubeConfig, err = cli.GetTKEClusterKubeConfig(data.Cluster.SystemID, isExtranet)
		if err != nil {
			blog.Errorf("importClusterCredential[%s] GetTKEClusterKubeConfig failed: %v", taskID, err)
			return err
		}
	}

	kubeRet, err := base64.StdEncoding.DecodeString(kubeConfig)
	if err != nil {
		return err
	}
	blog.Infof("importClusterCredential[%s] kubeConfig[%s]", taskID, string(kubeRet))

	// syncCluster sync kubeconfig to cluster
	if syncCluster {
		// save cluster kubeConfig
		data.Cluster.KubeConfig, _ = encrypt.Encrypt(nil, string(kubeRet))
		err = cloudprovider.UpdateCluster(data.Cluster)
		if err != nil {
			return err
		}
	}

	config, err := types.GetKubeConfigFromYAMLBody(false, types.YamlInput{
		YamlContent: string(kubeRet),
	})
	if err != nil {
		return err
	}

	blog.Infof("importClusterCredential[%s] kubeConfig token[%s]", taskID, token)
	if len(token) > 0 && len(config.AuthInfos) > 0 {
		config.AuthInfos[0].AuthInfo.Token = token
	}

	err = cloudprovider.UpdateClusterCredentialByConfig(data.Cluster.ClusterID, config)
	if err != nil {
		return err
	}

	return nil
}

// InstanceInfo instance info
type InstanceInfo struct {
	// InstanceIP ip
	InstanceIP string
	// InstanceId id
	InstanceId string
	// InstanceStatus status
	InstanceStatus string
}

func getClusterInstancesByClusterID(data *cloudprovider.CloudDependBasicInfo) ([]InstanceInfo, []InstanceInfo, error) {
	tkeCli, err := api.NewTkeClient(data.CmOption)
	if err != nil {
		return nil, nil, err
	}

	instancesList, err := tkeCli.QueryTkeClusterAllInstances(context.Background(), data.Cluster.SystemID, nil)
	if err != nil {
		return nil, nil, err
	}

	var (
		masterIPs, nodeIPs = make([]InstanceInfo, 0), make([]InstanceInfo, 0)
	)
	for _, ins := range instancesList {
		switch ins.InstanceRole {
		case api.MASTER_ETCD.String():
			masterIPs = append(masterIPs, InstanceInfo{
				InstanceIP:     ins.InstanceIP,
				InstanceId:     ins.InstanceID,
				InstanceStatus: ins.InstanceState,
			})
		case api.WORKER.String():
			nodeIPs = append(nodeIPs, InstanceInfo{
				InstanceIP:     ins.InstanceIP,
				InstanceId:     ins.InstanceID,
				InstanceStatus: ins.InstanceState,
			})
		default:
			continue
		}
	}

	return masterIPs, nodeIPs, nil
}
