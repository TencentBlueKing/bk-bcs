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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/nodemgr_v3"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tenant"
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
	if nodemgr_v3.IsEnabled() {
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

// getV3NetworkUnitID 参考 V2 getAPID(bkCloudID, clouds)：
// 调用 V3 管控单元列表接口，筛选 bk_networkarea_id 与 bkCloudID 相同的单元，
// 返回其中最小的 bk_networkunit_id
func getV3NetworkUnitID(ctx context.Context, v3Cli *nodemgr_v3.Client, bkCloudID int) int {
	resp, err := v3Cli.NetworkUnitList(ctx, &nodemgr_v3.NetworkUnitListRequest{
		Page: &nodemgr_v3.NetworkUnitPage{Limit: 1000},
		ExactIncludeConditions: &nodemgr_v3.NetworkUnitExactCondition{
			BkNetworkareaID: []int64{int64(bkCloudID)},
		},
	})
	if err != nil {
		blog.Errorf("getV3NetworkUnitID NetworkUnitList failed, err %s", err.Error())
		return 0
	}

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

	blog.Infof("InstallGSEAgentTask %s start install gse agent (v3), cluster %s, group %s, bizID %d, bkCloudID %d, nodeIPs %s",
		taskID, info.ClusterID, info.NodeGroupID, info.BkBizID, info.BkCloudID, info.NodeIPs)
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		fmt.Sprintf("start install gse agent (v3), cluster %s, group %s, bizID %d, bkCloudID %d",
			info.ClusterID, info.NodeGroupID, info.BkBizID, info.BkCloudID))

	v3Cli := nodemgr_v3.GetNodeManV3Client()
	if v3Cli == nil {
		retErr := fmt.Errorf("nodeman v3 client is not init")
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("nodeman v3 client is not init (v3)"))
		blog.Errorf("InstallGSEAgentTask %s failed: %s", taskID, retErr.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// 1. 构造安装主机参数
	ips := strings.Split(info.NodeIPs, ",")
	var ipv6s []string
	if len(info.NodeIPv6s) > 0 {
		ipv6s = strings.Split(info.NodeIPv6s, ",")
	}

	blog.Infof("InstallGSEAgentTask %s prepare install hosts (v3), ip count %d, nodeIPs %s, user %s, port %s",
		taskID, len(ips), info.NodeIPs, info.User, info.Port)

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

	hosts := make([]nodemgr_v3.AgentInstallHost, 0, len(ips))
	for i, v := range ips {
		ipv6 := ""
		if i < len(ipv6s) {
			ipv6 = ipv6s[i]
		}

		// 内网 IP 数组化
		innerIPList := []string{v}
		var innerIPv6List []string
		if ipv6 != "" {
			innerIPv6List = []string{ipv6}
		}

		// 登录端口
		loginPort := nodeman.DefaultPort
		if info.Port != "" {
			if dPort, errPort := strconv.Atoi(info.Port); errPort == nil {
				loginPort = dPort
			}
		}

		// 认证方式：v3 使用 login_mode 替代 auth_type
		loginMode := nodemgr_v3.LoginModePassword
		var loginPassword, loginKeyFile string

		if cloudprovider.IsMasterIp(v, info.DependInfo.Cluster) {
			// master 节点优先使用密钥
			if len(info.DependInfo.Cluster.GetNodeSettings().GetMasterLogin().GetKeyPair().GetKeySecret()) > 0 {
				loginMode = nodemgr_v3.LoginModeKeyFile
				plainSecret, _ := encrypt.Decrypt(nil,
					info.DependInfo.Cluster.GetNodeSettings().GetMasterLogin().GetKeyPair().GetKeySecret())
				secretStr, _ := encryptCreditWithPublicKey(publicKey, plainSecret)
				loginKeyFile = secretStr
			} else {
				loginMode = nodemgr_v3.LoginModePassword
				plainPwd, _ := encrypt.Decrypt(nil,
					info.DependInfo.Cluster.GetNodeSettings().GetMasterLogin().GetInitLoginPassword())
				pwd, _ := encryptCreditWithPublicKey(publicKey, plainPwd)
				loginPassword = pwd
			}
		} else {
			// worker 节点
			if len(info.Secret) > 0 {
				loginMode = nodemgr_v3.LoginModeKeyFile
				plainSecret, _ := encrypt.Decrypt(nil, info.Secret)
				secretStr, _ := encryptCreditWithPublicKey(publicKey, plainSecret)
				loginKeyFile = secretStr
			} else {
				loginMode = nodemgr_v3.LoginModePassword
				if len(info.Passwd) > 0 {
					plainPwd, _ := encrypt.Decrypt(nil, info.Passwd)
					pwd, _ := encryptCreditWithPublicKey(publicKey, plainPwd)
					loginPassword = pwd
				}
			}
		}

		host := nodemgr_v3.AgentInstallHost{
			BkAddressing:             nodemgr_v3.AddressingStatic,
			BkBizID:                  info.BkBizID,
			BkHostInnerIP:            innerIPList,
			BkHostInnerIPv6:          innerIPv6List,
			LoginIP:                  v,
			LoginPort:                loginPort,
			LoginUser:                info.User,
			LoginMode:                loginMode,
			LoginPassword:            loginPassword,
			LoginKeyFile:             loginKeyFile,
			OsType:                   nodemgr_v3.LinuxOSType,
			InstallMethod:            nodemgr_v3.InstallMethodSSH,
			BkNetworkunitID:          networkUnitID,
			InstallPreOrderedPlugins: true,
		}
		hosts = append(hosts, host)
		blog.Infof("InstallGSEAgentTask %s build host (v3) #%d: ip %s, loginMode %s, bizID %d",
			taskID, i, v, loginMode, info.BkBizID)
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		fmt.Sprintf("install gse agent (v3) for biz %d, ipv4: %s, ipv6: %s", info.BkBizID, info.NodeIPs, info.NodeIPv6s))

	// 3. 调用 v3 Agent 安装接口
	installReq := &nodemgr_v3.AgentInstallRequest{
		Host: hosts,
	}
	blog.Infof("InstallGSEAgentTask %s calling AgentInstall (v3), host count %d", taskID, len(hosts))
	installResp, err := v3Cli.AgentInstall(ctx, installReq)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("install gse agent (v3) failed [%s]", err))
		blog.Errorf("InstallGSEAgentTask %s install gse agent (v3) error, %s", taskID, err.Error())
		retErr := fmt.Errorf("install gse agent (v3) failed [%s]", err)
		_ = state.UpdateStepRetryOrFailure(start, stepName, retErr)
		return retErr
	}
	workflowID := installResp.WorkflowID
	blog.Infof("InstallGSEAgentTask %s step 2: AgentInstall (v3) success, workflow_id: %s", taskID, workflowID)
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		fmt.Sprintf("install gse agent (v3) success, workflow_id: %s", workflowID))

	// 4. 休眠10秒，避免任务过快查询导致节点还未开始安装返回成功
	blog.Infof("InstallGSEAgentTask %s step 3: sleep 10s before polling workflow status (v3), workflow_id: %s",
		taskID, workflowID)
	time.Sleep(10 * time.Second)
	blog.Infof("InstallGSEAgentTask %s step 4: start polling workflow status (v3), workflow_id: %s",
		taskID, workflowID)

	// 5. 轮询工作流操作状态
	checkCtx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	defer cancel()
	err = loop.LoopDoFunc(checkCtx, func() error {
		opReq := &nodemgr_v3.WorkflowOperationListRequest{
			WorkflowID: workflowID,
			Page: &nodemgr_v3.V3Page{
				Offset: 0,
				Limit:  200,
			},
		}
		opResp, errLocal := v3Cli.WorkflowOperationList(ctx, opReq)
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

		// 遍历所有操作，判断整体状态
		allSuccess := true
		hasFailed := false
		for _, op := range opResp.Operations {
			opState := op.LatestOperInstBriefData.LifeCycle.State
			blog.Infof("InstallGSEAgentTask %s workflow %s operation %s state %s",
				taskID, workflowID, op.OperationID, opState)
			switch opState {
			case string(nodemgr_v3.OperationStateSuccess):
				continue
			case string(nodemgr_v3.OperationStateFailed),
				string(nodemgr_v3.OperationStateTimeout):
				hasFailed = true
				allSuccess = false
			case string(nodemgr_v3.OperationStateTerminated):
				hasFailed = true
				allSuccess = false
			default:
				// init / launched / running
				allSuccess = false
			}
		}

		if allSuccess {
			blog.Infof("InstallGSEAgentTask %s all operations success (v3), workflow_id: %s",
				taskID, workflowID)
			cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
				fmt.Sprintf("all operations success (v3), workflow_id: %s", workflowID))
			return loop.EndLoop
		}

		if hasFailed {
			// 收集失败操作的具体原因（主机、失败动作、错误标签）
			failedOps := make([]string, 0)
			for _, op := range opResp.Operations {
				opState := op.LatestOperInstBriefData.LifeCycle.State
				if opState == string(nodemgr_v3.OperationStateFailed) ||
					opState == string(nodemgr_v3.OperationStateTimeout) ||
					opState == string(nodemgr_v3.OperationStateTerminated) {
					ips := op.NodeDeploymentInfo.BkHostInnerIPList
					host := fmt.Sprintf("%v", ips)
					if len(ips) == 0 {
						host = fmt.Sprintf("bk_host_id=%d", op.NodeDeploymentInfo.BkHostID)
					}
					action := op.LatestOperInstBriefData.LatestActionInstBriefData.Name
					tags := op.LatestOperInstBriefData.LatestActionInstBriefData.Tags
					failedOps = append(failedOps, fmt.Sprintf(
						"op=%s host=%s action=%s state=%s tags=%v",
						op.OperationID, host, action, opState, tags))
				}
			}
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
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("check gse agent install workflow status (v3) failed [%s]", err))
		blog.Errorf("InstallGSEAgentTask %s check gse agent install workflow status (v3) failed: %v",
			taskID, err)
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
