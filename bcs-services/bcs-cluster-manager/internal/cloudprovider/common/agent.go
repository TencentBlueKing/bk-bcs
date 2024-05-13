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

// Package common xxx
package common

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/nodeman"
)

var (
	installGseAgentStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.InstallGseAgentAction,
		StepName:   "安装 GSE Agent",
	}
)

// GseInstallInfo xxx
type GseInstallInfo struct {
	ClusterId   string
	NodeGroupId string
	BusinessId  string

	CloudArea *proto.CloudArea

	User    string
	Passwd  string
	KeyInfo *proto.KeyInfo

	Port               string
	AllowReviseCloudId string
}

// BuildInstallGseAgentTaskStep build common watch step
func BuildInstallGseAgentTaskStep(task *proto.Task, gseInfo *GseInstallInfo, options ...cloudprovider.StepOption) {
	installGseStep := cloudprovider.InitTaskStep(installGseAgentStep, options...)

	installGseStep.Params[cloudprovider.ClusterIDKey.String()] = gseInfo.ClusterId     // nolint
	installGseStep.Params[cloudprovider.NodeGroupIDKey.String()] = gseInfo.NodeGroupId // nolint

	installGseStep.Params[cloudprovider.BKBizIDKey.String()] = gseInfo.BusinessId // nolint
	if gseInfo != nil && gseInfo.CloudArea != nil {                               // nolint
		installGseStep.Params[cloudprovider.BKCloudIDKey.String()] = strconv.Itoa(int(gseInfo.CloudArea.BkCloudID))
	}
	installGseStep.Params[cloudprovider.UsernameKey.String()] = gseInfo.User
	installGseStep.Params[cloudprovider.PasswordKey.String()] = gseInfo.Passwd
	installGseStep.Params[cloudprovider.SecretKey.String()] = gseInfo.KeyInfo.GetKeySecret()
	installGseStep.Params[cloudprovider.PortKey.String()] = gseInfo.Port

	if gseInfo.AllowReviseCloudId == "" {
		gseInfo.AllowReviseCloudId = icommon.False
	}
	installGseStep.Params[cloudprovider.AllowReviseAgent.String()] = gseInfo.AllowReviseCloudId

	task.Steps[installGseAgentStep.StepMethod] = installGseStep
	task.StepSequence = append(task.StepSequence, installGseAgentStep.StepMethod)
}

// InstallGSEAgentTask install gse agent task
func InstallGSEAgentTask(taskID string, stepName string) error { // nolint
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start install gse agent")
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}
	// get cluster/nodeGroup
	clusterIDString := step.Params[cloudprovider.ClusterIDKey.String()]
	groupIDString := step.Params[cloudprovider.NodeGroupIDKey.String()]
	// get bkBizID
	bkBizIDString := step.Params[cloudprovider.BKBizIDKey.String()]
	// get bkCloudID
	// bkCloudIDString := step.Params[cloudprovider.BKCloudIDKey.String()]
	// get nodeIPs
	nodeIPs := state.Task.CommonParams[cloudprovider.NodeIPsKey.String()]
	// get password
	passwd := step.Params[cloudprovider.PasswordKey.String()]
	// get user
	user := step.Params[cloudprovider.UsernameKey.String()]
	// get port
	port := step.Params[cloudprovider.PortKey.String()]
	// allow check revise cloudId
	allow := step.Params[cloudprovider.AllowReviseAgent.String()]
	// update connect cluster status when task retry
	install, ok := step.Params[cloudprovider.InstallGseAgentKey.String()]
	if allow == icommon.True && ok && install == icommon.True {
		step.Params[cloudprovider.InstallGseAgentKey.String()] = icommon.False
	}

	if len(user) == 0 {
		user = nodeman.RootAccount
	}
	// get secretKey
	secret := step.Params[cloudprovider.SecretKey.String()]

	if len(nodeIPs) == 0 {
		blog.Infof("InstallGSEAgentTask %s skip, cause of empty node", taskID)
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("empty node ip"))
		return nil
	}

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterIDString,
		NodeGroupID: groupIDString,
	})
	if err != nil {
		blog.Infof("InstallGSEAgentTask GetClusterDependBasicInfo failed: %v", taskID, err)
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("installAgent getDependInfo failed"))
		return nil
	}
	if bkBizIDString == "" {
		bkBizIDString = dependInfo.Cluster.GetBusinessID()
	}
	cloudAreaID := func() string {
		if dependInfo.NodeGroup != nil && dependInfo.NodeGroup.GetArea() != nil {
			return strconv.Itoa(int(dependInfo.NodeGroup.GetArea().GetBkCloudID()))
		}

		return strconv.Itoa(int(dependInfo.Cluster.GetClusterBasicSettings().GetArea().GetBkCloudID()))
	}()

	bkCloudID, err := strconv.Atoi(cloudAreaID)
	if err != nil {
		blog.Errorf("InstallGSEAgentTask %s failed, invalid bkCloudID, err %s", taskID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("invalid bkCloudID, err %s", err.Error()))
		return nil
	}
	bkBizID, err := strconv.Atoi(bkBizIDString)
	if err != nil {
		blog.Errorf("InstallGSEAgentTask %s failed, invalid bkBizID, err %s", taskID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("invalid bkBizID, err %s", err.Error()))
		return nil
	}

	nodeManClient := nodeman.GetNodeManClient()
	if nodeManClient == nil {
		blog.Errorf("nodeman client is not init")
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("nodeman client is not init"))
		return nil
	}

	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)

	// get apID from cloud list
	clouds, err := nodeManClient.CloudList(context.Background())
	if err != nil {
		blog.Errorf("InstallGSEAgentTask %s get cloud list error, %s", taskID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("get cloud list error, %s", err.Error()))
		return nil
	}
	apID := getAPID(bkCloudID, clouds)

	// install gse agent
	hosts := make([]nodeman.JobInstallHost, 0)
	ips := strings.Split(nodeIPs, ",")

	// delete ips when install agent if hostIPs exist cmdb
	err = RemoveHostFromCmdb(ctx, bkBizID, nodeIPs)
	if err != nil {
		blog.Errorf("InstallGSEAgentTask %s RemoveHostFromCmdb error, %s", taskID, err.Error())
	}

	for _, v := range ips {
		hosts = append(hosts, nodeman.JobInstallHost{
			BKCloudID: bkCloudID,
			APID:      apID,
			BKBizID:   bkBizID,
			OSType:    nodeman.LinuxOSType,
			InnerIP:   v,
			LoginIP:   v,
			Account:   user,
			Port: func() int {
				if port == "" {
					return nodeman.DefaultPort
				}
				dPort, err := strconv.Atoi(port) // nolint
				if err != nil {
					return nodeman.DefaultPort
				}

				return dPort
			}(),
			AuthType: func() nodeman.AuthType {
				if cloudprovider.IsMasterIp(v, dependInfo.Cluster) {
					if len(dependInfo.Cluster.GetNodeSettings().GetMasterLogin().GetKeyPair().GetKeySecret()) > 0 {
						return nodeman.KeyAuthType
					}
					return nodeman.PasswordAuthType
				}

				if len(secret) > 0 {
					return nodeman.KeyAuthType
				}
				return nodeman.PasswordAuthType
			}(),
			Password: func() string {
				if cloudprovider.IsMasterIp(v, dependInfo.Cluster) &&
					len(dependInfo.Cluster.GetNodeSettings().GetMasterLogin().GetInitLoginPassword()) > 0 {
					pwd, _ := encrypt.Decrypt(nil,
						dependInfo.Cluster.GetNodeSettings().GetMasterLogin().GetInitLoginPassword())

					return pwd
				}

				if len(passwd) > 0 {
					pwd, _ := encrypt.Decrypt(nil, passwd)
					return pwd
				}
				return ""
			}(),
			Key: func() string {
				if cloudprovider.IsMasterIp(v, dependInfo.Cluster) &&
					len(dependInfo.Cluster.GetNodeSettings().GetMasterLogin().GetKeyPair().GetKeySecret()) > 0 {
					secretStr, _ := encrypt.Decrypt(nil,
						dependInfo.Cluster.GetNodeSettings().GetMasterLogin().GetKeyPair().GetKeySecret())
					return secretStr
				}

				if len(secret) > 0 {
					secretStr, _ := encrypt.Decrypt(nil, secret)
					return secretStr
				}
				return ""
			}(),
		})
	}
	job, err := nodeManClient.JobInstall(nodeman.InstallAgentJob, hosts)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("install gse agent job failed [%s]", err))
		blog.Errorf("InstallGSEAgentTask %s install gse agent job error, %s", taskID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("install gse agent job error, %s", err.Error()))
		return nil
	}
	blog.Infof("InstallGSEAgentTask %s install gse agent job(%d) url %s", taskID, job.JobID, job.JobURL)

	// check status
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	defer cancel()
	err = loop.LoopDoFunc(ctx, func() error {
		var err error // nolint
		detail, err := nodeManClient.JobDetails(job.JobID)
		if err != nil {
			blog.Errorf("InstallGSEAgentTask %s failed, get job detail err %s", taskID, err.Error())
			return err
		}
		switch detail.Status {
		case nodeman.JobRunning:
			cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
				"checking job status, waiting")
			blog.Infof("InstallGSEAgentTask %s checking job status, waiting", taskID)
			return nil
		case nodeman.JobSuccess:
			return loop.EndLoop
		case nodeman.JobFailed, nodeman.JobPartFailed:
			return fmt.Errorf("GSE Agent 安装失败，详情查看: %s", job.JobURL)
		}
		return nil
	}, loop.LoopInterval(5*time.Second))
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("check gse agent install job status failed [%s]", err))
		blog.Errorf("InstallGSEAgentTask %s check gse agent install job status failed: %v", taskID, err)
		if allow == icommon.True {
			step.Params[cloudprovider.InstallGseAgentKey.String()] = icommon.True
		}

		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("check gse "+
			"agent install job status err: %s", err.Error()))
		return nil
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"install gse agent job successful")

	// update step
	_ = state.UpdateStepSucc(start, stepName)

	return nil
}

func getAPID(bkCloudID int, clouds []nodeman.CloudListData) int {
	apID := nodeman.DefaultAPID
	for _, v := range clouds {
		if v.BKCloudID == 0 {
			continue
		}
		if v.BKCloudID == bkCloudID {
			apID = v.APID
			break
		}
	}
	return apID
}
