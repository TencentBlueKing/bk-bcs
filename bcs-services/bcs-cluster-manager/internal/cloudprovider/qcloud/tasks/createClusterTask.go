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

package tasks

import (
	"context"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

const (
	// KubeAPIServer cluster apiserver key
	KubeAPIServer = "KubeAPIServer"
	// KubeController cluster controller key
	KubeController = "KubeController"
	// KubeScheduler cluster scheduler key
	KubeScheduler = "KubeScheduler"
	// Etcd cluster etcd key
	Etcd = "Etcd"
	// Kubelet cluster kubelet key
	Kubelet = "kubelet"
)

const (
	// DockerGraphPath default docker graphPath
	DockerGraphPath = "/data/bcs/service/docker"
	// MountTarget default mountTarget
	MountTarget = "/data"
)

// as far as possible to keep every operation unit simple

func stringToInt(str string) (int, error) {
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}

	return num, nil
}

func generateClusterCIDRInfo(cluster *proto.Cluster) *api.ClusterCIDRSettings {
	cidrInfo := &api.ClusterCIDRSettings{
		ClusterCIDR:          cluster.NetworkSettings.ClusterIPv4CIDR,
		MaxNodePodNum:        uint64(cluster.NetworkSettings.MaxNodePodNum),
		MaxClusterServiceNum: uint64(cluster.NetworkSettings.MaxServiceNum),
		ServiceCIDR:          cluster.NetworkSettings.ServiceIPv4CIDR,
	}

	return cidrInfo
}

func generateTags(bizID int64, operator string) map[string]string {
	cli := cmdb.GetCmdbClient()
	if cli == nil {
		return nil
	}

	return nil
}

func generateClusterBasicInfo(cluster *proto.Cluster, imageID, operator string) *api.ClusterBasicSettings {
	basicInfo := &api.ClusterBasicSettings{
		ClusterOS:      imageID,
		ClusterVersion: cluster.ClusterBasicSettings.Version,
		ClusterName:    cluster.ClusterID,
		VpcID:          cluster.VpcID,
	}

	basicInfo.TagSpecification = make([]*api.TagSpecification, 0)
	// build qcloud tag info
	if len(cluster.ClusterBasicSettings.ClusterTags) > 0 {
		tags := make([]*api.Tag, 0)
		for k, v := range cluster.ClusterBasicSettings.ClusterTags {
			tags = append(tags, &api.Tag{
				Key:   common.StringPtr(k),
				Value: common.StringPtr(v),
			})
		}
		basicInfo.TagSpecification = append(basicInfo.TagSpecification, &api.TagSpecification{
			ResourceType: "cluster",
			Tags:         tags,
		})
	} else { // according to cloud different realization to adapt
		bizID, _ := strconv.Atoi(cluster.BusinessID)
		cloudTags := generateTags(int64(bizID), operator)
		tags := make([]*api.Tag, 0)
		if len(cloudTags) > 0 {
			for k, v := range cloudTags {
				tags = append(tags, &api.Tag{
					Key:   common.StringPtr(k),
					Value: common.StringPtr(v),
				})
			}

			basicInfo.TagSpecification = append(basicInfo.TagSpecification, &api.TagSpecification{
				ResourceType: "cluster",
				Tags:         tags,
			})
		}
	}

	return basicInfo
}

func generateClusterAdvancedInfo(cluster *proto.Cluster) *api.ClusterAdvancedSettings {
	advancedInfo := &api.ClusterAdvancedSettings{
		IPVS:             cluster.ClusterAdvanceSettings.IPVS,
		ContainerRuntime: cluster.ClusterAdvanceSettings.ContainerRuntime,
		RuntimeVersion:   cluster.ClusterAdvanceSettings.RuntimeVersion,
		ExtraArgs:        &api.ClusterExtraArgs{},
	}

	if len(cluster.ClusterAdvanceSettings.ExtraArgs) > 0 {
		if apiserver, ok := cluster.ClusterAdvanceSettings.ExtraArgs[KubeAPIServer]; ok {
			paras := strings.Split(apiserver, ";")
			advancedInfo.ExtraArgs.KubeAPIServer = common.StringPtrs(paras)
		}

		if controller, ok := cluster.ClusterAdvanceSettings.ExtraArgs[KubeController]; ok {
			paras := strings.Split(controller, ";")
			advancedInfo.ExtraArgs.KubeControllerManager = common.StringPtrs(paras)
		}

		if scheduler, ok := cluster.ClusterAdvanceSettings.ExtraArgs[KubeScheduler]; ok {
			paras := strings.Split(scheduler, ";")
			advancedInfo.ExtraArgs.KubeScheduler = common.StringPtrs(paras)
		}

		if etcd, ok := cluster.ClusterAdvanceSettings.ExtraArgs[Etcd]; ok {
			paras := strings.Split(etcd, ";")
			advancedInfo.ExtraArgs.Etcd = common.StringPtrs(paras)
		}
	}

	return advancedInfo
}

func generateInstanceAdvanceInfo(cluster *proto.Cluster) *api.InstanceAdvancedSettings {
	if cluster.NodeSettings.MountTarget == "" {
		cluster.NodeSettings.MountTarget = MountTarget
	}
	if cluster.NodeSettings.DockerGraphPath == "" {
		cluster.NodeSettings.DockerGraphPath = DockerGraphPath
	}

	mountTarget := cluster.NodeSettings.MountTarget
	if len(cluster.ExtraClusterID) > 0 {
		mountTarget = ""
	}

	advanceInfo := &api.InstanceAdvancedSettings{
		MountTarget:     mountTarget,
		DockerGraphPath: cluster.NodeSettings.DockerGraphPath,
		Unschedulable:   common.Int64Ptr(int64(cluster.NodeSettings.UnSchedulable)),
	}

	// cluster node common labels
	if len(cluster.NodeSettings.Labels) > 0 {
		for key, value := range cluster.NodeSettings.Labels {
			advanceInfo.Labels = append(advanceInfo.Labels, &api.KeyValue{
				Name:  key,
				Value: value,
			})
		}
	}

	// Kubelet start params
	if len(cluster.NodeSettings.ExtraArgs) > 0 {
		advanceInfo.ExtraArgs = &api.InstanceExtraArgs{}

		if kubelet, ok := cluster.NodeSettings.ExtraArgs[Kubelet]; ok {
			paras := strings.Split(kubelet, ";")
			advanceInfo.ExtraArgs.Kubelet = paras
		}
	}

	return advanceInfo
}

func generateExistedInstance(passwd string, instanceIDs []string) *api.ExistedInstancesForNode {
	existedInstance := &api.ExistedInstancesForNode{
		NodeRole: api.MASTER_ETCD.String(),
		ExistedInstancesPara: &api.ExistedInstancesPara{
			InstanceIDs:   instanceIDs,
			LoginSettings: &api.LoginSettings{Password: passwd},
		},
	}

	return existedInstance
}

func disksToCVMDisks(disks []*proto.DataDisk) []*cvm.DataDisk {
	if len(disks) == 0 {
		return nil
	}

	cvmDisks := make([]*cvm.DataDisk, 0)
	for i := range disks {
		size, _ := stringToInt(disks[i].DiskSize)
		cvmDisks = append(cvmDisks, &cvm.DataDisk{
			DiskSize: common.Int64Ptr(int64(size)),
			DiskType: common.StringPtr(disks[i].DiskType),
		})
	}

	return cvmDisks
}

func generateRunInstance(cluster *proto.Cluster, passwd string) *api.RunInstancesForNode {
	runInstance := &api.RunInstancesForNode{
		NodeRole: api.MASTER_ETCD.String(),
	}

	for i := range cluster.Template {
		systemDiskSize, _ := stringToInt(cluster.Template[i].SystemDisk.DiskSize)
		req := &cvm.RunInstancesRequest{
			Placement: &cvm.Placement{
				Zone: common.StringPtr(cluster.Template[i].Zone),
			},
			InstanceType: common.StringPtr(cluster.Template[i].InstanceType),
			ImageId:      common.StringPtr(cluster.Template[i].ImageInfo.ImageID),
			SystemDisk: &cvm.SystemDisk{
				DiskType: common.StringPtr(cluster.Template[i].SystemDisk.DiskType),
				DiskSize: common.Int64Ptr(int64(systemDiskSize)),
			},
			DataDisks: disksToCVMDisks(cluster.Template[i].DataDisks),
			VirtualPrivateCloud: &cvm.VirtualPrivateCloud{
				VpcId:    common.StringPtr(cluster.Template[i].VpcID),
				SubnetId: common.StringPtr(cluster.Template[i].SubnetID),
			},

			InstanceCount: common.Int64Ptr(int64(cluster.Template[i].ApplyNum)),
			LoginSettings: &cvm.LoginSettings{
				Password: common.StringPtr(passwd),
			},
		}

		requestStr := req.ToJsonString()
		runInstance.RunInstancesPara = append(runInstance.RunInstancesPara, common.StringPtr(requestStr))
	}

	return runInstance
}

// CreateClusterShieldAlarmTask call alarm interface to shield alarm
func CreateClusterShieldAlarmTask(taskID string, stepName string) error {
	start := time.Now()
	//get task information and validate
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("CreateClusterShieldAlarmTask[%s]: task %s get detail task information from storage failed, %s. task retry", taskID, taskID, err.Error())
		return err
	}

	state := &cloudprovider.TaskState{Task: task, JobResult: cloudprovider.NewJobSyncResult(task)}
	if state.IsTerminated() {
		blog.Errorf("CreateClusterShieldAlarmTask[%s]: task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("CreateClusterShieldAlarmTask[%s]: task %s not turn to run step %s, err %s", taskID, taskID, stepName, err.Error())
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CreateClusterShieldAlarmTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CreateClusterShieldAlarmTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params["ClusterID"]
	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("CreateClusterShieldAlarmTask[%s]: get cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("get cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	masterIPs := make([]string, 0)
	for masterIP := range cluster.Master {
		masterIPs = append(masterIPs, masterIP)
	}

	if len(masterIPs) == 0 {
		blog.Errorf("CreateClusterShieldAlarmTask[%s]: get cluster masterIPs empty", taskID)
		retErr := fmt.Errorf("CreateClusterShieldAlarmTask: get cluster masterIPs empty")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// attention: call client to shieldAlarm

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateClusterShieldAlarmTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

// CreateTkeClusterTask call qcloud interface to create cluster
func CreateTkeClusterTask(taskID string, stepName string) error {
	start := time.Now()
	//get task information and validate
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("CreateTkeClusterTask[%s]: task %s get detail task information from storage failed, %s. task retry", taskID, taskID, err.Error())
		return err
	}
	//defer utils.RecoverPrintStack("CreateTkeClusterTask")

	state := &cloudprovider.TaskState{Task: task, JobResult: cloudprovider.NewJobSyncResult(task)}
	if state.IsTerminated() {
		blog.Errorf("CreateTkeClusterTask[%s]: task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("CreateTkeClusterTask[%s]: task %s not turn to run step %s, err %s", taskID, taskID, stepName, err.Error())
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CreateTkeClusterTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CreateTkeClusterTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params["ClusterID"]
	cloudID := step.Params["CloudID"]

	operator := state.Task.CommonParams["operator"]

	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("CreateTkeClusterTask[%s]: get cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("get cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cloud, project, err := actions.GetProjectAndCloud(cloudprovider.GetStorageModel(), cluster.ProjectID, cloudID)
	if err != nil {
		blog.Errorf("CreateTkeClusterTask[%s]: get cloud/project for cluster %s in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get dependency resource for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(project, cloud)
	if err != nil {
		blog.Errorf("CreateTkeClusterTask[%s]: get credential for cluster %s in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud credential err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption.Region = cluster.Region

	// get qcloud client
	tkeCli, err := api.NewTkeClient(cmOption)
	if err != nil {
		blog.Errorf("CreateTkeClusterTask[%s]: get tke client for cluster[%s] in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	masterIPs := make([]string, 0)
	for masterIP := range cluster.Master {
		masterIPs = append(masterIPs, masterIP)
	}
	instanceIDs, err := transIPsToInstanceID(&cloudprovider.ListNodesOption{
		Common:       cmOption,
		ClusterVPCID: cluster.VpcID,
	}, masterIPs)
	if err != nil || len(instanceIDs) == 0 {
		blog.Errorf("CreateTkeClusterTask[%s]: transIPsToInstanceID for cluster[%s] in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err)
		retErr := fmt.Errorf("CreateTkeClusterTask transIPsToInstanceID err, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	imageID, err := transImageNameToImageID(cmOption, cluster.ClusterBasicSettings.OS)
	if err != nil {
		blog.Errorf("CreateTkeClusterTask[%s]: transImageNameToImageID for cluster[%s] in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err)
		retErr := fmt.Errorf("CreateTkeClusterTask transImageNameToImageID err, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	passwd := task.CommonParams["Password"]
	req := &api.CreateClusterRequest{
		AddNodeMode:             cluster.AutoGenerateMasterNodes,
		Region:                  cluster.Region,
		ClusterType:             cluster.ManageType,
		ClusterCIDR:             generateClusterCIDRInfo(cluster),
		ClusterBasic:            generateClusterBasicInfo(cluster, imageID, operator),
		ClusterAdvanced:         generateClusterAdvancedInfo(cluster),
		InstanceAdvanced:        generateInstanceAdvanceInfo(cluster),
		ExistedInstancesForNode: nil,
		RunInstancesForNode:     nil,
	}

	// auto generate machine
	if req.AddNodeMode {
		req.RunInstancesForNode = []*api.RunInstancesForNode{
			generateRunInstance(cluster, passwd),
		}
	} else {
		req.ExistedInstancesForNode = []*api.ExistedInstancesForNode{
			generateExistedInstance(passwd, instanceIDs),
		}
	}

	systemID := cluster.SystemID
	if cluster.SystemID != "" {
		tkeCluster, err := tkeCli.GetTKECluster(cluster.SystemID)
		if err != nil {
			blog.Errorf("CreateTkeClusterTask[%s]: call GetTKECluster[%s] api in task %s step %s failed, %s",
				taskID, clusterID, taskID, stepName, err.Error())
			retErr := fmt.Errorf("call GetTKECluster[%s] api err, %s", clusterID, err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
		systemID = *tkeCluster.ClusterId
	} else {
		resp, err := tkeCli.CreateTKECluster(req)
		if err != nil {
			blog.Errorf("CreateTkeClusterTask[%s]: call CreateTKECluster[%s] api in task %s step %s failed, %s",
				taskID, clusterID, taskID, stepName, err.Error())
			retErr := fmt.Errorf("call CreateTKECluster[%s] api err, %s", clusterID, err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
		blog.Infof("CreateTkeClusterTask[%s]: call CreateTKECluster interface successful", taskID)

		// update cluster systemID
		err = updateClusterSystemID(clusterID, resp.ClusterID)
		if err != nil {
			blog.Errorf("CreateTkeClusterTask[%s]: updateClusterSystemID[%s] in task %s step %s failed, %s",
				taskID, clusterID, taskID, stepName, err.Error())
			retErr := fmt.Errorf("call CreateTKECluster updateClusterSystemID[%s] api err, %s", clusterID, err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
		blog.Infof("CreateTkeClusterTask[%s]: call CreateTKECluster updateClusterSystemID successful", taskID)
		systemID = resp.ClusterID
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	state.Task.CommonParams["SystemID"] = systemID
	state.Task.CommonParams["MasterIPs"] = strings.Join(masterIPs, ",")
	state.Task.CommonParams["InstanceIDs"] = strings.Join(instanceIDs, ",")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateTkeClusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

// CheckTkeClusterStatusTask check cluster create status
func CheckTkeClusterStatusTask(taskID string, stepName string) error {
	start := time.Now()
	//get task information and validate
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("CheckTkeClusterStatusTask[%s]: task %s get detail task information from storage failed, %s. task retry", taskID, taskID, err.Error())
		return err
	}

	state := &cloudprovider.TaskState{Task: task, JobResult: cloudprovider.NewJobSyncResult(task)}
	if state.IsTerminated() {
		blog.Errorf("CheckTkeClusterStatusTask[%s]: task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("CheckTkeClusterStatusTask[%s]: task %s not turn to run step %s, err %s", taskID, taskID, stepName, err.Error())
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckTkeClusterStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckTkeClusterStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params["ClusterID"]
	cloudID := step.Params["CloudID"]
	systemID := state.Task.CommonParams["SystemID"]

	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("CheckTkeClusterStatusTask[%s]: get cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("get cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cloud, project, err := actions.GetProjectAndCloud(cloudprovider.GetStorageModel(), cluster.ProjectID, cloudID)
	if err != nil {
		blog.Errorf("CheckTkeClusterStatusTask[%s]: get cloud/project for cluster %s in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get dependency resource for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(project, cloud)
	if err != nil {
		blog.Errorf("CheckTkeClusterStatusTask[%s]: get credential for cluster %s in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud credential err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption.Region = cluster.Region

	// get qcloud client
	cli, err := api.NewTkeClient(cmOption)
	if err != nil {
		blog.Errorf("CheckTkeClusterStatusTask[%s]: get tke client for cluster[%s] in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	var (
		timeOut  = false
		abnormal = false
	)
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			blog.Infof("CheckTkeClusterStatusTask[%s] GetTKECluster[%s] timeout", taskID, clusterID)
			timeOut = true
		}

		// timeOut quit
		if timeOut {
			break
		}

		cluster, err := cli.GetTKECluster(systemID)
		if err != nil {
			continue
		}

		blog.Infof("CheckTkeClusterStatusTask[%s] cluster[%s] current status[%s]", taskID, clusterID, *cluster.ClusterStatus)
		// check cluster status
		if *cluster.ClusterStatus == "Running" {
			break
		}
		if *cluster.ClusterStatus == "Abnormal" {
			abnormal = true
			break
		}
	}

	if timeOut || abnormal {
		blog.Errorf("CheckTkeClusterStatusTask[%s]: call CreateTKECluster[%s] api in task %s step %s timeout|abnormal",
			taskID, clusterID, taskID, stepName)
		retErr := fmt.Errorf("call CheckTkeClusterStatusTask[%s] api timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckTkeClusterStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// EnableTkeClusterVpcCniTask enable on vpc-cni networkMode
func EnableTkeClusterVpcCniTask(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("EnableTkeClusterVpcCniTask[%s]: task %s get detail task information from storage failed, %s. task retry", taskID, taskID, err.Error())
		return err
	}

	state := &cloudprovider.TaskState{Task: task, JobResult: cloudprovider.NewJobSyncResult(task)}
	if state.IsTerminated() {
		blog.Errorf("EnableTkeClusterVpcCniTask[%s]: task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("EnableTkeClusterVpcCniTask[%s]: task %s not turn to run step %s, err %s", taskID, taskID, stepName, err.Error())
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("EnableTkeClusterVpcCniTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}

	blog.Infof("EnableTkeClusterVpcCniTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params["ClusterID"]
	cloudID := step.Params["CloudID"]
	systemID := state.Task.CommonParams["SystemID"]

	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("EnableTkeClusterVpcCniTask[%s]: get cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("get cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cloud, project, err := actions.GetProjectAndCloud(cloudprovider.GetStorageModel(), cluster.ProjectID, cloudID)
	if err != nil {
		blog.Errorf("EnableTkeClusterVpcCniTask[%s]: get cloud/project for cluster %s in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get dependency resource for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(project, cloud)
	if err != nil {
		blog.Errorf("EnableTkeClusterVpcCniTask[%s]: get credential for cluster %s in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud credential err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption.Region = cluster.Region

	cli, err := api.NewTkeClient(cmOption)
	if err != nil {
		blog.Errorf("EnableTkeClusterVpcCniTask[%s]: get tke client for cluster[%s] in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	blog.Infof("EnableTkeClusterVpcCniTask[%s]: enableVPCCni %v", taskID, cluster.NetworkSettings.EnableVPCCni)
	if cluster.NetworkSettings.EnableVPCCni {
		err = cli.EnableTKEVpcCniMode(&api.EnableVpcCniInput{
			TkeClusterID:   systemID,
			VpcCniType:     api.TKEDirectEni,
			SubnetsIDs:     cluster.NetworkSettings.EniSubnetIDs,
			EnableStaticIP: cluster.NetworkSettings.IsStaticIpMode,
			ExpiredSeconds: int(cluster.NetworkSettings.ClaimExpiredSeconds),
		})
		if err != nil {
			blog.Errorf("EnableTkeClusterVpcCniTask[%s]: tke EnableTKEVpcCniMode for cluster[%s] in task %s step %s failed, %s",
				taskID, clusterID, taskID, stepName, err.Error())
			retErr := fmt.Errorf("EnableTKEVpcCniMode err, %s", err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}

		var (
			timeOut, abnormal = false, false
		)
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
		defer cancel()

		for {
			select {
			case <-ticker.C:
			case <-ctx.Done():
				blog.Infof("CheckTkeClusterStatusTask[%s] GetTKECluster[%s] timeout", taskID, clusterID)
				timeOut = true
			}

			// timeOut quit
			if timeOut {
				break
			}

			status, err := cli.GetEnableVpcCniProgress(systemID)
			if err != nil {
				continue
			}

			blog.Infof("EnableTkeClusterVpcCniTask[%s]: GetEnableVpcCniProgress current status[%s]",
				taskID, status.Status)
			if status.Status == string(api.Succeed) {
				break
			}

			if status.Status == string(api.Failed) {
				abnormal = true
				break
			}
		}

		if timeOut || abnormal {
			blog.Errorf("EnableTkeClusterVpcCniTask[%s]: call GetEnableVpcCniProgress status timeout|abnormal",
				taskID)
			retErr := fmt.Errorf("call GetEnableVpcCniProgress[%s] api timeout|abnormal", clusterID)
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("EnableTkeClusterVpcCniTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// UpdateCreateClusterDBInfoTask update cluster DB info
func UpdateCreateClusterDBInfoTask(taskID string, stepName string) error {
	start := time.Now()
	//get task information and validate
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s]: task %s get detail task information from storage failed, %s. task retry", taskID, taskID, err.Error())
		return err
	}

	state := &cloudprovider.TaskState{Task: task, JobResult: cloudprovider.NewJobSyncResult(task)}
	if state.IsTerminated() {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s]: task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s]: task %s not turn to run step %s, err %s", taskID, taskID, stepName, err.Error())
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UpdateCreateClusterDBInfoTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}

	blog.Infof("UpdateCreateClusterDBInfoTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params["ClusterID"]
	SystemID := state.Task.CommonParams["SystemID"]

	// need to generate master Nodes and update DB if auto generate machines

	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s]: get cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("get cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cluster.SystemID = SystemID
	cluster.Status = icommon.StatusInitialization

	err = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s]: update cluster systemID for %s failed", taskID, clusterID)
	}

	// sync clusterData to pass-cc
	utils.SyncClusterInfoToPassCC(taskID, cluster)

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}
