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

// Package template for template
package template

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tenant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// bksops task category
const (
	// UserBeforeInit before init
	UserBeforeInit = "userBeforeInit"
	// UserPostInit post init
	UserPostInit = "userAfterInit"
	// SystemInit bksops system init
	SystemInit = "系统初始化"
	// UserAfterInit bksops user after init
	UserAfterInit = "用户后置初始化"
	// UserPreInit bksops user pre init
	UserPreInit = "缩容节点清理"
	// NodeMixedInit mixed init
	NodeMixedInit = "nodeMixedInit"
	// NodeMixedInitCh mixed init ch
	NodeMixedInitCh = "混部集群节点初始化"
)

// using task commonName inject dynamic parameters when processing
var (
	// DynamicParameterInject inject parameter for bk-sops
	DynamicParameterInject = map[string]string{
		clusterMasterIPs:   "MasterNodeIPList",
		nodeIPList:         "NodeIPList",
		externalNodeScript: "ExternalNodeScript",
		clusterKubeConfig:  "KubeConfig",
	}
)

// ExtraInfo extra template values
type ExtraInfo struct {
	InstancePasswd     string
	NodeIPList         string
	MasterIPList       string
	NodeOperator       string
	ModuleID           string
	BusinessID         string
	Operator           string
	ExternalNodeScript string
	ClusterKubeConfig  string
	NodeGroupID        string
	ShowSopsUrl        bool
	TranslateMethod    string
	GroupCreator       string
	GroupColocation    bool
}

// BuildSopsFactory xxx
type BuildSopsFactory struct {
	StepName string
	Cluster  *proto.Cluster
	Extra    ExtraInfo
}

// BuildSopsStep build sops task
func (f BuildSopsFactory) BuildSopsStep(task *proto.Task, action *proto.Action, pre bool) error {
	step := &BkSopsStepAction{
		TaskName: f.StepName,
		Actions: func() []string {
			if pre {
				return action.PreActions
			}

			return action.PostActions
		}(),
		Plugins: action.Plugins,
	}
	err := step.BuildBkSopsStepAction(task, f.Cluster, f.Extra)
	if err != nil {
		return err
	}

	return nil
}

// BkSopsStepAction build bksops step action
type BkSopsStepAction struct {
	TaskName string
	Actions  []string
	Plugins  map[string]*proto.BKOpsPlugin
}

// BuildBkSopsStepAction build sops step action
func (sopStep *BkSopsStepAction) BuildBkSopsStepAction(task *proto.Task, cluster *proto.Cluster, info ExtraInfo) error {
	for _, name := range sopStep.Actions {
		plugin, ok := sopStep.Plugins[name]
		if ok {
			taskName := sopStep.TaskName
			if pluginName, ok := plugin.Params["template_name"]; ok && pluginName != "" {
				taskName = pluginName
			}

			if len(plugin.Params) == 0 {
				continue
			}

			stepName := cloudprovider.BKSOPTask + "-" + utils.RandomString(8)
			step, err := GenerateBKopsStep("", taskName, stepName, cluster, plugin, info)
			if err != nil {
				return fmt.Errorf("BuildBkSopsStepAction step failed: %v", err)
			}
			task.Steps[stepName] = step
			task.StepSequence = append(task.StepSequence, stepName)
		}
	}

	return nil
}

// GetPluginByAction get plugin by actionName
func GetPluginByAction(action *proto.Action, actionName string) *proto.BKOpsPlugin {
	if action != nil {
		plugin, ok := action.Plugins[actionName]
		if ok {
			return plugin
		}
	}
	return nil
}

// GenerateBKopsStep generate common bk-sops step
func GenerateBKopsStep(taskMethod, taskName, stepName string, cls *proto.Cluster, plugin *proto.BKOpsPlugin,
	info ExtraInfo) (*proto.Step, error) {
	now := time.Now().Format(time.RFC3339)

	if taskName == "" {
		taskName = SystemInit
	}
	if taskMethod == "" {
		taskMethod = cloudprovider.BKSOPTask
	}

	step := &proto.Step{
		Name:   stepName,
		System: plugin.System,
		Params: make(map[string]string),
		Retry:  0,
		Start:  now,
		Status: cloudprovider.TaskStatusNotStarted,
		// method name is registered name to taskServer
		TaskMethod:   taskMethod,
		TaskName:     taskName,
		SkipOnFailed: plugin.AllowSkipWhenFailed,
		Translate:    info.TranslateMethod,
	}
	step.Params[cloudprovider.BkSopsUrlKey.String()] = plugin.Link
	step.Params[cloudprovider.ShowSopsUrlKey.String()] = fmt.Sprintf("%v", info.ShowSopsUrl)

	constants := make(map[string]string)
	// 变量values值分3类: 1.标准参数  2.直接渲染参数 3.template渲染参数
	for k, v := range plugin.Params {
		switch {
		// 兼容自定义 业务ID/流程ID/流程user
		case strings.HasPrefix(k, template) && !strings.HasPrefix(v, prefix):
			step.Params[k] = v
		case len(v) == 0:
			continue
		case len(strings.Split(v, ".")) == 3 && strings.HasPrefix(v, prefix):
			tValue, err := getTemplateParameterByName(v, cls, info)
			if err != nil {
				blog.Errorf("%s GenerateBKopsStep failed: %v", taskName, err)
				return nil, err
			}
			// 兼容自定义 业务ID/业务源
			if strings.Contains(v, template) {
				step.Params[k] = tValue
				continue
			}
			constants[fmt.Sprintf("${%s}", k)] = tValue
		case renderTextContainTemplateVars(v):
			blog.Infof("renderTextContainTemplateVars %v, info %+v", v, info)
			render := NewRenderTemplateVars(cls, info.NodeIPList, info.NodeOperator)
			tValue, err := render.RenderTxtVars("", v)
			if err != nil {
				blog.Errorf("%s GenerateBKopsStep RenderTxtVars failed: %v", taskName, err)
				return nil, err
			}
			blog.Infof("renderTextContainTemplateVars %+v", tValue)
			constants[fmt.Sprintf("${%s}", k)] = tValue
		default:
			constants[fmt.Sprintf("${%s}", k)] = v
		}
	}
	constantsbyte, err := json.Marshal(&constants)
	if err != nil {
		blog.Errorf("%s GenerateBKopsStep failed: %v", taskName, err)
		return nil, err
	}
	step.Params["constants"] = string(constantsbyte)

	return step, nil
}

// NOCC:CCN_threshold(工具误报:),golint/fnsize(设计如此:)
func getTemplateParameterByName(name string, cluster *proto.Cluster, extra ExtraInfo) (string, error) { // nolint
	if cluster == nil {
		errMsg := fmt.Errorf("cluster is empty when getTemplateParameterByName")
		blog.Errorf(errMsg.Error())
		return "", errMsg
	}

	switch name {
	case clusterID:
		return cluster.GetClusterID(), nil
	case clusterBizOperator:
		biz, _ := strconv.Atoi(cluster.GetBusinessID())
		ctx, err := tenant.WithTenantIdByResourceForContext(context.Background(),
			tenant.ResourceMetaData{ProjectId: cluster.GetProjectID()})
		if err != nil {
			return "", err
		}
		maintainers := cloudprovider.GetBizMaintainers(ctx, biz)
		if len(maintainers) > 0 {
			return maintainers, nil
		}
		return strings.Join([]string{cluster.GetCreator(), extra.GroupCreator}, ","), nil
	case clusterGroupColocation:
		if extra.GroupColocation {
			return common.True, nil
		}
		return common.False, nil
	case clusterMasterIPs:
		if len(extra.MasterIPList) == 0 {
			return getClusterMasterIPs(cluster), nil
		}
		return clusterMasterIPs, nil
	case clusterRegion:
		return cluster.GetRegion(), nil
	case clusterVPC:
		return cluster.GetVpcID(), nil
	case clusterNetworkType:
		return cluster.GetNetworkType(), nil
	case clusterBizID:
		if extra.BusinessID == "" {
			return cluster.GetBusinessID(), nil
		}
		return extra.BusinessID, nil
	case clusterBizCCID:
		return extra.BusinessID, nil
	case clusterModuleID:
		return extra.ModuleID, nil
	case clusterExtraID:
		return getClusterType(cluster), nil
	case clusterManageEnv:
		return cluster.GetManageType(), nil
	case clusterExtraClusterID:
		return cluster.GetExtraClusterID(), nil
	case clusterProjectID:
		return cluster.GetProjectID(), nil
	case nodePasswd:
		return extra.InstancePasswd, nil
	case nodeCPUManagerPolicy:
		return defaultPolicy, nil
	// dynamic parameter
	case nodeIPList:
		if len(extra.NodeIPList) == 0 {
			return nodeIPList, nil
		}
		return extra.NodeIPList, nil
	case externalNodeScript:
		if len(extra.ExternalNodeScript) == 0 {
			return externalNodeScript, nil
		}
		return extra.ExternalNodeScript, nil
	case clusterKubeConfig:
		if len(extra.ClusterKubeConfig) == 0 {
			return clusterKubeConfig, nil
		}
		return extra.ExternalNodeScript, nil
	case nodeOperator:
		return extra.NodeOperator, nil
	case templateBusinessID:
		return extra.BusinessID, nil
	case templateOperator:
		return extra.Operator, nil
	// self builder cluster envs
	case clusterExtraEnv, addNodesExtraEnv:
		return getSelfBuilderClusterEnvs(cluster)
	case bcsCommonInfo:
		envs, err := getBcsEnvs(cluster)
		if err != nil {
			return "", nil
		}
		return envs, nil
	case nodeGroupID:
		return extra.NodeGroupID, nil
	case clusterCloudArea:
		return fmt.Sprintf("%v", getClusterCloudArea(cluster)), nil
	case clusterOsType:
		return getClusterOsType(cluster), nil
	case clusterK8sVersion:
		return cluster.GetClusterBasicSettings().GetVersion(), nil
	case clusterProvider:
		return cluster.GetProvider(), nil
	default:
	}

	return "", fmt.Errorf("getTemplateParameterByName unSupportType %s", name)
}
