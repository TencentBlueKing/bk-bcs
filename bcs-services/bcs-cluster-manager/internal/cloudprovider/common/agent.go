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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/nodemgrv3"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tenant"
)

// installGseAgentStep 安装 GSE Agent 任务步骤定义
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
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start install gse agent")

	// get cluster/nodeGroup
	clusterIDString := step.Params[cloudprovider.ClusterIDKey.String()]
	groupIDString := step.Params[cloudprovider.NodeGroupIDKey.String()]
	// get bkBizID
	bkBizIDString := step.Params[cloudprovider.BKBizIDKey.String()]
	// get bkCloudID
	// bkCloudIDString := step.Params[cloudprovider.BKCloudIDKey.String()]
	// get nodeIPs
	nodeIPs := state.Task.CommonParams[cloudprovider.NodeIPsKey.String()]
	// get nodeIPv6s
	nodeIPv6s := state.Task.CommonParams[cloudprovider.NodeIPv6sKey.String()]
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
		retErr := fmt.Errorf("empty node ip")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterIDString,
		NodeGroupID: groupIDString,
	})
	if err != nil {
		blog.Infof("InstallGSEAgentTask GetClusterDependBasicInfo failed: %v", taskID, err)
		retErr := fmt.Errorf("installAgent getDependInfo failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
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
		retErr := fmt.Errorf("invalid bkCloudID, err %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	bkBizID, err := strconv.Atoi(bkBizIDString)
	if err != nil {
		blog.Errorf("InstallGSEAgentTask %s failed, invalid bkBizID, err %s", taskID, err.Error())
		retErr := fmt.Errorf("invalid bkBizID, err %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)
	ctx, err = tenant.WithTenantIdByResourceForContext(ctx,
		tenant.ResourceMetaData{ProjectId: dependInfo.Cluster.GetProjectID()})
	if err != nil {
		retErr := fmt.Errorf("WithTenantIdByResourceForContext %s failed", dependInfo.Cluster.GetProjectID())
		blog.Errorf("InstallGSEAgentTask %s failed: %s", taskID, retErr.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// v3 分支：根据版本开关决定是否走 v3 接口
	if nodemgrv3.IsEnabled() {
		v3Info := &gseAgentV3InstallInfo{
			ClusterID:   clusterIDString,
			NodeGroupID: groupIDString,
			BkBizID:     bkBizID,
			BkCloudID:   bkCloudID,
			NodeIPs:     nodeIPs,
			NodeIPv6s:   nodeIPv6s,
			User:        user,
			Passwd:      passwd,
			Port:        port,
			Secret:      secret,
			Allow:       allow,
			DependInfo:  dependInfo,
		}
		return installGSEAgentV3(ctx, taskID, stepName, start, state, step, v3Info) // nolint
	}

	// v2 分支：通过节点管理 V2 接口安装
	nodeManClient := nodeman.GetNodeManClient()
	if nodeManClient == nil {
		retErr := fmt.Errorf("nodeman client is not init")
		blog.Errorf("InstallGSEAgentTask %s failed: %s", taskID, retErr.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get apID from cloud list
	clouds, err := nodeManClient.CloudList(ctx)
	if err != nil {
		blog.Errorf("InstallGSEAgentTask %s get cloud list error, %s", taskID, err.Error())
		retErr := fmt.Errorf("get cloud list error, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	apID := getAPID(bkCloudID, clouds)

	// install gse agent
	hosts := make([]nodeman.JobInstallHost, 0)
	ips := strings.Split(nodeIPs, ",")
	var ipv6s []string
	if len(nodeIPv6s) > 0 {
		ipv6s = strings.Split(nodeIPv6s, ",")
	}

	// delete ips when install agent if hostIPs exist cmdb
	err = RemoveHostFromCmdb(ctx, bkBizID, nodeIPs)
	if err != nil {
		blog.Errorf("InstallGSEAgentTask %s RemoveHostFromCmdb error, %s", taskID, err.Error())
	}

	for i, v := range ips {
		// ipv6 与 ipv4 按下标一一对应，缺省时为空
		ipv6 := ""
		if i < len(ipv6s) {
			ipv6 = ipv6s[i]
		}
		hosts = append(hosts, nodeman.JobInstallHost{
			BKCloudID: bkCloudID,
			APID:      apID,
			BKBizID:   bkBizID,
			OSType:    nodeman.LinuxOSType,
			InnerIP:   v,
			InnerIPv6: ipv6,
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
			// master 节点使用集群初始化登录信息，worker 节点使用任务参数中的登录信息；均优先密钥认证
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
			// 凭据在存储中加密保存，提交前需解密为明文
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
			ForceUpdateAgentId: true,
		})
	}
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		fmt.Sprintf("install gse agent for biz %d, ipv4: %s, ipv6: %s", bkBizID, nodeIPs, nodeIPv6s))
	job, err := nodeManClient.JobInstall(ctx, nodeman.InstallAgentJob, hosts)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("install gse agent job failed [%s]", err))
		blog.Errorf("InstallGSEAgentTask %s install gse agent job error, %s", taskID, err.Error())
		retErr := fmt.Errorf("install gse agent job failed [%s]", err)
		_ = state.UpdateStepRetryOrFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("InstallGSEAgentTask %s install gse agent job(%d) url %s", taskID, job.JobID, job.JobURL)

	// 休眠10秒，避免任务过快查询导致节点还未开始安装返回成功
	time.Sleep(10 * time.Second)

	// check status
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	err = loop.LoopDoFunc(ctx, func() error {
		detail, errLocal := nodeManClient.JobDetails(ctx, job.JobID)
		if errLocal != nil {
			blog.Errorf("InstallGSEAgentTask %s failed, get job detail err %s", taskID, errLocal.Error())
			return errLocal
		}

		blog.Infof("InstallGSEAgentTask %s checking job ID[%d], detail[%#v]", taskID, job.JobID, detail)

		// 运行中继续轮询，成功结束轮询，失败或部分失败直接返回错误
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
		// 允许修正 Agent 时记录安装失败标记，供后续步骤处理
		if allow == icommon.True {
			step.Params[cloudprovider.InstallGseAgentKey.String()] = icommon.True
		}

		_ = state.UpdateStepRetryOrFailure(start, stepName, fmt.Errorf("check gse "+
			"agent install job status err: %s", err.Error()))
		return err
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"install gse agent job successful")

	// update step
	_ = state.UpdateStepSucc(start, stepName)

	return nil
}

// getAPID 从云区域列表中匹配 bkCloudID 对应的接入点 ID，未匹配时返回默认接入点
func getAPID(bkCloudID int, clouds []nodeman.CloudListData) int {
	apID := nodeman.DefaultAPID
	for _, v := range clouds {
		// 直连区域（0）使用默认接入点，不参与匹配
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

// getV3NetworkUnitID 参考 V2 getAPID(bkCloudID, clouds)：
// 调用 V3 管控单元列表接口，筛选 bk_networkarea_id 与 bkCloudID 相同的单元，
// 返回其中最小的 bk_networkunit_id
func getV3NetworkUnitID(ctx context.Context, v3Cli *nodemgrv3.Client, bkCloudID int) int {
	resp, err := v3Cli.NetworkUnitList(ctx, &nodemgrv3.NetworkUnitListRequest{
		Page: &nodemgrv3.NetworkUnitPage{Limit: 1000},
		ExactIncludeConditions: &nodemgrv3.NetworkUnitExactCondition{
			BkNetworkareaID: []int64{int64(bkCloudID)},
		},
	})
	if err != nil {
		blog.Errorf("getV3NetworkUnitID NetworkUnitList failed, err %s", err.Error())
		return 0
	}

	// 同一管控区域存在多个网络单元时取最小 ID，保证结果稳定
	var minUnitID int64
	matched := false
	for _, item := range resp.Items {
		if item.BkNetworkareaID != int64(bkCloudID) {
			continue
		}
		if !matched || item.BkNetworkunitID < minUnitID {
			minUnitID = item.BkNetworkunitID
			matched = true
		}
	}
	if !matched {
		return 0
	}
	return int(minUnitID)
}

// gseAgentV3InstallInfo installGSEAgentV3 入参集合，用于收敛函数参数
type gseAgentV3InstallInfo struct {
	ClusterID   string
	NodeGroupID string
	BkBizID     int
	BkCloudID   int
	NodeIPs     string
	NodeIPv6s   string
	User        string
	Passwd      string
	Port        string
	Secret      string
	Allow       string
	DependInfo  *cloudprovider.CloudDependBasicInfo
}

// installGSEAgentV3 v3 接口安装 GSE Agent 流程
func installGSEAgentV3(ctx context.Context, taskID string, stepName string, start time.Time,
	state *cloudprovider.TaskState, step *proto.Step,
	info *gseAgentV3InstallInfo) error {

	blog.Infof("InstallGSEAgentTask %s start install gse agent (v3), cluster %s, group %s, bizID %d, "+
		"bkCloudID %d, nodeIPs %s", taskID, info.ClusterID, info.NodeGroupID, info.BkBizID, info.BkCloudID, info.NodeIPs)
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		fmt.Sprintf("start install gse agent (v3), cluster %s, group %s, bizID %d, bkCloudID %d",
			info.ClusterID, info.NodeGroupID, info.BkBizID, info.BkCloudID))

	// v3 客户端未初始化时直接失败，不回退到 v2
	v3Cli := nodemgrv3.GetNodeManV3Client()
	if v3Cli == nil {
		retErr := fmt.Errorf("nodeman v3 client is not init")
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			"nodeman v3 client is not init (v3)")
		blog.Errorf("InstallGSEAgentTask %s failed: %s", taskID, retErr.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// 1. 构造安装主机参数
	ips := strings.Split(info.NodeIPs, ",")
	blog.Infof("InstallGSEAgentTask %s prepare install hosts (v3), ip count %d, nodeIPs %s, user %s, port %s",
		taskID, len(ips), info.NodeIPs, info.User, info.Port)
	cleanV3InstallHosts(ctx, v3Cli, taskID, stepName, info, ips)

	// 从管控区域列表匹配 bkCloudID，获取对应的网络单元（管控区域）ID
	networkUnitID := getV3NetworkUnitID(ctx, v3Cli, info.BkCloudID)
	blog.Infof("InstallGSEAgentTask %s get network unit id (v3) by bkCloudID %d: %d",
		taskID, info.BkCloudID, networkUnitID)

	// v3 凭据需使用节点管理服务器 RSA 公钥加密
	// RSA-OAEP(SHA256) -> 前置 0x01 -> base64
	publicKey, err := v3Cli.PublicKeyGet(ctx)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("get nodeman v3 public key error: %s", err.Error()))
		blog.Errorf("InstallGSEAgentTask %s get nodeman v3 public key error, %s", taskID, err.Error())
		retErr := fmt.Errorf("get nodeman v3 public key failed [%s]", err)
		_ = state.UpdateStepRetryOrFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("InstallGSEAgentTask %s get nodeman v3 public key success", taskID)

	// 2. 构造安装主机参数，凭据使用公钥加密
	hosts := buildV3InstallHosts(taskID, info, ips, publicKey, networkUnitID)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		fmt.Sprintf("install gse agent (v3) for biz %d, ipv4: %s, ipv6: %s", info.BkBizID, info.NodeIPs, info.NodeIPv6s))

	// 3. 调用 v3 Agent 安装接口
	workflowID, err := submitV3AgentInstall(ctx, v3Cli, taskID, stepName, hosts)
	if err != nil {
		_ = state.UpdateStepRetryOrFailure(start, stepName, err)
		return err
	}

	// 4. 休眠10秒，避免任务过快查询导致节点还未开始安装返回成功
	blog.Infof("InstallGSEAgentTask %s step 3: sleep 10s before polling workflow status (v3), workflow_id: %s",
		taskID, workflowID)
	time.Sleep(10 * time.Second)
	blog.Infof("InstallGSEAgentTask %s step 4: start polling workflow status (v3), workflow_id: %s",
		taskID, workflowID)

	// 5. 轮询工作流操作状态
	err = waitV3InstallWorkflow(ctx, v3Cli, taskID, stepName, workflowID)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("check gse agent install workflow status (v3) failed [%s]", err))
		blog.Errorf("InstallGSEAgentTask %s check gse agent install workflow status (v3) failed: %v",
			taskID, err)
		// 允许修正 Agent 时记录安装失败标记，供后续步骤处理
		if info.Allow == icommon.True {
			step.Params[cloudprovider.InstallGseAgentKey.String()] = icommon.True
		}

		_ = state.UpdateStepRetryOrFailure(start, stepName, fmt.Errorf("check gse "+
			"agent install workflow status (v3) err: %s", err.Error()))
		return err
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"install gse agent (v3) workflow successful")

	// update step
	_ = state.UpdateStepSucc(start, stepName)

	return nil
}

// cleanV3InstallHosts 安装前对比 v3 主机列表并清理 cmdb 中已存在的主机记录，失败仅记录日志
func cleanV3InstallHosts(ctx context.Context, v3Cli *nodemgrv3.Client, taskID, stepName string,
	info *gseAgentV3InstallInfo, ips []string) {
	blog.Infof("InstallGSEAgentTask %s step 0: get host list (v3) and compare with install ips", taskID)
	existHostIDs, err := v3Cli.GetHostIDByIPs(ctx, info.BkBizID, ips)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("get host list (v3) error: %s", err.Error()))
		blog.Errorf("InstallGSEAgentTask %s get host list (v3) error, %s", taskID, err.Error())
	} else {
		blog.Infof("InstallGSEAgentTask %s compare host list (v3) done, existHostCount %d, existHostIDs %v",
			taskID, len(existHostIDs), existHostIDs)
	}

	// 先清理 cmdb 中已存在的主机记录
	blog.Infof("InstallGSEAgentTask %s step 1: build install hosts (v3) and clean cmdb", taskID)
	err = RemoveHostFromCmdb(ctx, info.BkBizID, info.NodeIPs)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("RemoveHostFromCmdb (v3) error: %s", err.Error()))
		blog.Errorf("InstallGSEAgentTask %s RemoveHostFromCmdb error, %s", taskID, err.Error())
	}
	blog.Infof("InstallGSEAgentTask %s RemoveHostFromCmdb (v3) done", taskID)
}

// buildV3InstallHosts 构造 v3 Agent 安装接口的主机参数
func buildV3InstallHosts(taskID string, info *gseAgentV3InstallInfo, ips []string, publicKey string,
	networkUnitID int) []nodemgrv3.AgentInstallHost {
	var ipv6s []string
	if len(info.NodeIPv6s) > 0 {
		ipv6s = strings.Split(info.NodeIPv6s, ",")
	}

	// 登录端口
	loginPort := nodeman.DefaultPort
	if info.Port != "" {
		if dPort, errPort := strconv.Atoi(info.Port); errPort == nil {
			loginPort = dPort
		}
	}

	hosts := make([]nodemgrv3.AgentInstallHost, 0, len(ips))
	for i, v := range ips {
		// 内网 IP 数组化
		var innerIPv6List []string
		if i < len(ipv6s) && ipv6s[i] != "" {
			innerIPv6List = []string{ipv6s[i]}
		}

		loginMode, loginPassword, loginKeyFile := buildV3HostCredential(v, info, publicKey)
		hosts = append(hosts, nodemgrv3.AgentInstallHost{
			BkAddressing:             nodemgrv3.AddressingStatic,
			BkBizID:                  info.BkBizID,
			BkHostInnerIP:            []string{v},
			BkHostInnerIPv6:          innerIPv6List,
			LoginIP:                  v,
			LoginPort:                loginPort,
			LoginUser:                info.User,
			LoginMode:                loginMode,
			LoginPassword:            loginPassword,
			LoginKeyFile:             loginKeyFile,
			OsType:                   nodemgrv3.LinuxOSType,
			InstallMethod:            nodemgrv3.InstallMethodSSH,
			BkNetworkunitID:          networkUnitID,
			InstallPreOrderedPlugins: true,
		})
		blog.Infof("InstallGSEAgentTask %s build host (v3) #%d: ip %s, loginMode %s, bizID %d",
			taskID, i, v, loginMode, info.BkBizID)
	}
	return hosts
}

// buildV3HostCredential 计算主机登录方式与凭据，v3 使用 login_mode 替代 auth_type，
// 返回 loginMode、loginPassword、loginKeyFile，凭据均已用节点管理公钥加密
func buildV3HostCredential(ip string, info *gseAgentV3InstallInfo,
	publicKey string) (nodemgrv3.LoginMode, string, string) {
	if cloudprovider.IsMasterIp(ip, info.DependInfo.Cluster) {
		masterLogin := info.DependInfo.Cluster.GetNodeSettings().GetMasterLogin()
		// master 节点优先使用密钥
		if len(masterLogin.GetKeyPair().GetKeySecret()) > 0 {
			return nodemgrv3.LoginModeKeyFile, "", encryptV3Credit(publicKey, masterLogin.GetKeyPair().GetKeySecret())
		}
		return nodemgrv3.LoginModePassword, encryptV3Credit(publicKey, masterLogin.GetInitLoginPassword()), ""
	}

	// worker 节点
	if len(info.Secret) > 0 {
		return nodemgrv3.LoginModeKeyFile, "", encryptV3Credit(publicKey, info.Secret)
	}
	if len(info.Passwd) > 0 {
		return nodemgrv3.LoginModePassword, encryptV3Credit(publicKey, info.Passwd), ""
	}
	return nodemgrv3.LoginModePassword, "", ""
}

// submitV3AgentInstall 调用 v3 Agent 安装接口，返回工作流 ID
func submitV3AgentInstall(ctx context.Context, v3Cli *nodemgrv3.Client, taskID, stepName string,
	hosts []nodemgrv3.AgentInstallHost) (string, error) {
	blog.Infof("InstallGSEAgentTask %s calling AgentInstall (v3), host count %d", taskID, len(hosts))
	installResp, err := v3Cli.AgentInstall(ctx, &nodemgrv3.AgentInstallRequest{Host: hosts})
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("install gse agent (v3) failed [%s]", err))
		blog.Errorf("InstallGSEAgentTask %s install gse agent (v3) error, %s", taskID, err.Error())
		return "", fmt.Errorf("install gse agent (v3) failed [%s]", err)
	}
	workflowID := installResp.WorkflowID
	blog.Infof("InstallGSEAgentTask %s step 2: AgentInstall (v3) success, workflow_id: %s", taskID, workflowID)
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		fmt.Sprintf("install gse agent (v3) success, workflow_id: %s", workflowID))
	return workflowID, nil
}

// encryptV3Credit 解密本地存储的凭据后使用节点管理公钥重新加密
func encryptV3Credit(publicKey, encrypted string) string {
	plain, _ := encrypt.Decrypt(nil, encrypted)
	credit, _ := encryptCreditWithPublicKey(publicKey, plain)
	return credit
}

// waitV3InstallWorkflow 轮询 v3 安装工作流，全部成功返回 nil，存在失败操作或超时返回错误
func waitV3InstallWorkflow(ctx context.Context, v3Cli *nodemgrv3.Client, taskID, stepName,
	workflowID string) error {
	// 轮询整体超时 10 分钟，每 5 秒查询一次
	checkCtx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	defer cancel()
	return loop.LoopDoFunc(checkCtx, func() error {
		opResp, errLocal := v3Cli.WorkflowOperationList(ctx, &nodemgrv3.WorkflowOperationListRequest{
			WorkflowID: workflowID,
			Page:       &nodemgrv3.V3Page{Offset: 0, Limit: 200},
		})
		if errLocal != nil {
			cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
				fmt.Sprintf("get workflow operation list (v3) err: %s", errLocal.Error()))
			blog.Errorf("InstallGSEAgentTask %s get workflow operation list (v3) err %s",
				taskID, errLocal.Error())
			return errLocal
		}

		blog.Infof("InstallGSEAgentTask %s checking workflow %s (v3), operations count %d",
			taskID, workflowID, len(opResp.Operations))
		cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
			fmt.Sprintf("checking workflow (v3) %s, operations count %d", workflowID, len(opResp.Operations)))

		// 全部成功结束轮询；存在失败立即返回；其余情况继续等待
		allSuccess, hasFailed := summarizeV3Operations(taskID, workflowID, opResp.Operations)
		if allSuccess {
			blog.Infof("InstallGSEAgentTask %s all operations success (v3), workflow_id: %s",
				taskID, workflowID)
			cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
				fmt.Sprintf("all operations success (v3), workflow_id: %s", workflowID))
			return loop.EndLoop
		}

		if hasFailed {
			failedOps := collectV3FailedOperations(opResp.Operations)
			blog.Errorf("InstallGSEAgentTask %s workflow %s has failed operations: %v",
				taskID, workflowID, failedOps)
			cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
				fmt.Sprintf("workflow (v3) %s failed operations: %v", workflowID, failedOps))
			return fmt.Errorf("GSE Agent 安装失败, workflow_id: %s, failed operations: %v",
				workflowID, failedOps)
		}

		cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
			"checking workflow status (v3), waiting")
		return nil
	}, loop.LoopInterval(5*time.Second))
}

// summarizeV3Operations 遍历所有操作，判断整体状态；init / launched / running 视为未完成
func summarizeV3Operations(taskID, workflowID string, ops []nodemgrv3.Operation) (allSuccess, hasFailed bool) {
	allSuccess = true
	for _, op := range ops {
		opState := op.LatestOperInstBriefData.LifeCycle.State
		blog.Infof("InstallGSEAgentTask %s workflow %s operation %s state %s",
			taskID, workflowID, op.OperationID, opState)
		if opState == string(nodemgrv3.OperationStateSuccess) {
			continue
		}
		allSuccess = false
		if isV3OperationFailed(opState) {
			hasFailed = true
		}
	}
	return allSuccess, hasFailed
}

// collectV3FailedOperations 收集失败操作的具体原因（主机、失败动作、错误标签）
func collectV3FailedOperations(ops []nodemgrv3.Operation) []string {
	failedOps := make([]string, 0)
	for _, op := range ops {
		opState := op.LatestOperInstBriefData.LifeCycle.State
		if !isV3OperationFailed(opState) {
			continue
		}
		ips := op.NodeDeploymentInfo.BkHostInnerIPList
		// 无内网 IP 时以 bk_host_id 标识主机
		host := fmt.Sprintf("%v", ips)
		if len(ips) == 0 {
			host = fmt.Sprintf("bk_host_id=%d", op.NodeDeploymentInfo.BkHostID)
		}
		action := op.LatestOperInstBriefData.LatestActionInstBriefData.Name
		tags := op.LatestOperInstBriefData.LatestActionInstBriefData.Tags
		failedOps = append(failedOps, fmt.Sprintf("op=%s host=%s action=%s state=%s tags=%v",
			op.OperationID, host, action, opState, tags))
	}
	return failedOps
}

// isV3OperationFailed 判断操作是否处于失败终态（failed / timeout / terminated）
func isV3OperationFailed(opState string) bool {
	return opState == string(nodemgrv3.OperationStateFailed) ||
		opState == string(nodemgrv3.OperationStateTimeout) ||
		opState == string(nodemgrv3.OperationStateTerminated)
}
