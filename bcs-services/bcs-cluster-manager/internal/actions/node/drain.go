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

package node

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// DrainNodeAction action for drain node
type DrainNodeAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.DrainNodeRequest
	resp  *cmproto.DrainNodeResponse
	k8sOp *clusterops.K8SOperator

	cluster *cmproto.Cluster
	task    *cmproto.Task
	drainer *clusterops.DrainHelper
}

// NewDrainNodeAction create update action
func NewDrainNodeAction(model store.ClusterManagerModel, k8sOp *clusterops.K8SOperator) *DrainNodeAction {
	return &DrainNodeAction{
		model: model,
		k8sOp: k8sOp,
	}
}

func (ua *DrainNodeAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}
	// set default GracePeriodSeconds, negative value mean the default value specified in the pod will be used
	if ua.req.GracePeriodSeconds == 0 {
		ua.req.GracePeriodSeconds = -1
	}

	cluster, err := ua.model.GetCluster(ua.ctx, ua.req.ClusterID)
	if err == nil {
		ua.cluster = cluster
	}

	return nil
}

func (ua *DrainNodeAction) generateDrainHelper() error {
	// new drainer
	drainer := clusterops.DrainHelper{
		Force:                           true,
		GracePeriodSeconds:              int(ua.req.GracePeriodSeconds),
		IgnoreAllDaemonSets:             true,
		Timeout:                         int(ua.req.Timeout),
		DeleteLocalData:                 true,
		Selector:                        ua.req.Selector,
		PodSelector:                     ua.req.PodSelector,
		DisableEviction:                 true,
		DryRun:                          ua.req.DryRun,
		SkipWaitForDeleteTimeoutSeconds: int(ua.req.SkipWaitForDeleteTimeoutSeconds),
	}

	ua.drainer = &drainer

	// get node names
	if len(ua.req.Nodes) == 0 && len(ua.req.InnerIPs) > 0 {
		option := clusterops.ListNodeOption{ClusterID: ua.req.ClusterID, NodeIPs: ua.req.InnerIPs}
		nodes, err := ua.k8sOp.ListClusterNodesByIPsOrNames(ua.ctx, option)
		if err != nil {
			blog.Errorf("get nodename by ips failed in cluster %s, err %s", ua.req.ClusterID, err.Error())
			return fmt.Errorf("get nodename by ips failed in cluster %s, err %s", ua.req.ClusterID, err.Error())
		}
		for _, v := range nodes {
			ua.req.Nodes = append(ua.req.Nodes, v.Name)
		}
	}

	return nil
}

func (ua *DrainNodeAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handles node drain
func (ua *DrainNodeAction) Handle(ctx context.Context, req *cmproto.DrainNodeRequest, resp *cmproto.DrainNodeResponse) {
	if req == nil || resp == nil {
		blog.Errorf("drain cluster node failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.generateDrainHelper(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := ua.createDispatchTask(); err != nil {
		ua.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
		return
	}

	err := ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   ua.req.ClusterID,
		TaskID:       ua.task.TaskID,
		Message:      fmt.Sprintf("集群[%s]节点pod迁移", ua.req.ClusterID),
		OpUser:       auth.GetUserFromCtx(ua.ctx),
		CreateTime:   time.Now().UTC().Format(time.RFC3339),
		ClusterID:    ua.req.ClusterID,
		ProjectID:    ua.cluster.ProjectID,
		ResourceName: ua.cluster.ClusterName,
	})
	if err != nil {
		blog.Errorf("DrainNode[%s] CreateOperationLog failed: %v", ua.req.ClusterID, err)
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

func (ua *DrainNodeAction) createDispatchTask() error {
	task, err := ua.buildDrainPodTask()
	if err != nil {
		blog.Errorf("CreateDispatchTask BuildDebugSopsTask failed: %v", err)
		return err
	}

	ua.task = task

	err = CreateDispatchTask(ua.ctx, ua.model, task)
	if err != nil {
		blog.Errorf("CreateDispatchTask CreateDispatchTask failed: %v", err)
		return err
	}

	ua.resp.Data = task

	return nil
}

// BuildUpdateDesiredNodesTask build update desired nodes task
func (ua *DrainNodeAction) buildDrainPodTask() (*cmproto.Task, error) {
	// generate main task
	nowStr := time.Now().UTC().Format(time.RFC3339)
	task := &cmproto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.NodeDrainPodTask.String(),
		TaskName:       cloudprovider.NodeDrainPodTaskName.String(),
		Status:         cloudprovider.TaskStatusInit,
		Message:        "task initializing",
		Start:          nowStr,
		Steps:          make(map[string]*cmproto.Step),
		StepSequence:   make([]string, 0),
		ClusterID:      ua.cluster.ClusterID,
		ProjectID:      ua.cluster.ProjectID,
		Creator:        ua.req.Operator,
		Updater:        ua.req.Operator,
		LastUpdate:     nowStr,
		CommonParams:   make(map[string]string),
		ForceTerminate: false,
	}
	// generate taskName
	task.CommonParams[cloudprovider.TaskNameKey.String()] = "节点驱逐pod任务"

	err := ua.generateDrainPodStep(task)
	if err != nil {
		return nil, err
	}
	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildDispatchDebugSopsTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]

	return task, nil
}

func (ua *DrainNodeAction) generateDrainPodStep(task *cmproto.Task) error {
	now := time.Now().UTC().Format(time.RFC3339)

	stepName := cloudprovider.NodeDrainPodAction + "-" + utils.RandomString(8)
	step := &cmproto.Step{
		Name:   stepName,
		System: "clustermanager",
		Params: make(map[string]string),
		Retry:  0,
		Start:  now,
		Status: cloudprovider.TaskStatusNotStarted,
		// method name is registered name to taskServer
		TaskMethod: cloudprovider.NodeDrainPodAction,
		TaskName:   "节点驱逐pod",
	}

	step.Params[cloudprovider.NodeNamesKey.String()] = strings.Join(ua.req.Nodes, ",")

	drainer, err := json.Marshal(ua.drainer)
	if err != nil {
		blog.Errorf("generateBKopsStep failed: %v", err)
		return err
	}
	step.Params[cloudprovider.DrainHelperKey.String()] = string(drainer)
	step.Params[cloudprovider.ClusterIDKey.String()] = ua.req.ClusterID

	task.Steps[stepName] = step
	task.StepSequence = append(task.StepSequence, stepName)

	return nil
}

// CreateDispatchTask create and dispatch task
func CreateDispatchTask(ctx context.Context, model store.ClusterManagerModel, task *cmproto.Task) error {
	// create task and dispatch task
	if err := model.CreateTask(ctx, task); err != nil {
		return err
	}
	if err := taskserver.GetTaskServer().Dispatch(task); err != nil {
		return err
	}

	return nil
}

// CheckDrainNodeAction action for drain node
type CheckDrainNodeAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.CheckDrainNodeRequest
	resp  *cmproto.CheckDrainNodeResponse
	k8sOp *clusterops.K8SOperator

	cluster *cmproto.Cluster
	// node label filter pod
	selector labels.Selector
	// pod label filter
	podSelector labels.Selector
}

// NewCheckDrainNodeAction create update action
func NewCheckDrainNodeAction(model store.ClusterManagerModel, k8sOp *clusterops.K8SOperator) *CheckDrainNodeAction {
	return &CheckDrainNodeAction{
		model: model,
		k8sOp: k8sOp,
	}
}

const (
	maxGracePeriodSeconds int32 = 600
)

func (ua *CheckDrainNodeAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}
	// set default GracePeriodSeconds, negative value mean the default value specified in the pod will be used
	if ua.req.GracePeriodSeconds == 0 {
		ua.req.GracePeriodSeconds = -1
	}

	if ua.req.GracePeriodSeconds > maxGracePeriodSeconds {
		blog.Errorf("checkDrainClusterNodes gracePeriodSeconds greater than %d", maxGracePeriodSeconds)
		return fmt.Errorf("checkDrainClusterNodes gracePeriodSeconds must be less than or equal to %d", maxGracePeriodSeconds)
	}

	if ua.req.Selector != "" {
		selector, err := labels.Parse(ua.req.Selector)
		if err != nil {
			blog.Errorf("checkDrainClusterNodes parse Selector label filter failed, err: %s", err.Error())
			return fmt.Errorf("checkDrainClusterNodes parse Selector label filter failed , err: %s", err.Error())
		}

		ua.selector = selector
	}

	if ua.req.PodSelector != "" {
		podSelector, err := labels.Parse(ua.req.PodSelector)
		if err != nil {
			blog.Errorf("checkDrainClusterNodes parse PodSelector label filter failed, err: %s", err.Error())
			return fmt.Errorf("checkDrainClusterNodes parse PodSelector label filter failed , err: %s", err.Error())
		}

		ua.podSelector = podSelector
	}

	cluster, err := ua.model.GetCluster(ua.ctx, ua.req.ClusterID)
	if err != nil {
		blog.Errorf("checkDrainClusterNodes get cluster failed, err: %s", err.Error())
		return fmt.Errorf("checkDrainClusterNodes get cluster failed , err: %s", err.Error())
	}

	ua.cluster = cluster

	return nil
}

func (ua *CheckDrainNodeAction) checkDrainClusterNodes() error { // nolint
	// get node names
	nodeLabelsMap := make(map[string]map[string]string)

	if len(ua.req.InnerIPs) > 0 || len(ua.req.Nodes) > 0 {
		option := clusterops.ListNodeOption{ClusterID: ua.req.ClusterID, NodeIPs: ua.req.InnerIPs, NodeNames: ua.req.Nodes}
		nodes, err := ua.k8sOp.ListClusterNodesByIPsOrNames(ua.ctx, option)
		if err != nil {
			blog.Errorf("get nodename by ips failed in cluster %s, err %s", ua.req.ClusterID, err.Error())
			return fmt.Errorf("get nodename by ips failed in cluster %s, err %s", ua.req.ClusterID, err.Error())
		}

		if len(ua.req.Nodes) == 0 && len(ua.req.InnerIPs) > 0 {
			ua.req.Nodes = make([]string, 0, len(nodes))
			for _, node := range nodes {
				ua.req.Nodes = append(ua.req.Nodes, node.Name)
			}
		}

		for _, node := range nodes {
			nodeLabelsMap[node.Name] = node.Labels
		}
	}

	barrier := utils.NewRoutinePool(50)
	defer barrier.Close()

	dataCh := make(chan []*cmproto.CheckDrainNodeData, len(ua.req.Nodes))

	for i := range ua.req.Nodes {
		barrier.Add(1)
		go func(node string) {
			defer barrier.Done()

			if ua.selector != nil {
				nodeLabels := nodeLabelsMap[node]
				if !ua.selector.Matches(labels.Set(nodeLabels)) {
					return
				}
			}

			option := clusterops.GetNodePodsOption{
				ClusterID: ua.req.ClusterID,
				NodeName:  node,
			}

			ctx, cancel := context.WithTimeout(context.Background(), clusterops.DefaultTimeout)
			defer cancel()
			pods, err := ua.k8sOp.GetNodePods(ctx, option)
			if err != nil {
				blog.Errorf("checkDrainClusterNodes[%s] failed in cluster %s, err %s", pods, ua.req.ClusterID, err.Error())
				return
			}

			podCheckData := make([]*cmproto.CheckDrainNodeData, 0)

			for _, pod := range pods {
				if ua.podSelector != nil {
					podLabels := pod.GetLabels()
					if podLabels == nil {
						podLabels = make(map[string]string)
					}
					if !ua.podSelector.Matches(labels.Set(podLabels)) {
						continue
					}
				}

				// Use input if provided, else use default, if terminationGracePeriodSeconds is nil use 30 seconds
				var gracePeriodSeconds uint32 = 30
				if pod.Spec.TerminationGracePeriodSeconds != nil && *pod.Spec.TerminationGracePeriodSeconds > 0 {
					gracePeriodSeconds = uint32(math.Min(float64(*pod.Spec.TerminationGracePeriodSeconds), float64(math.MaxUint32)))
				}
				if ua.req.GracePeriodSeconds > 0 {
					gracePeriodSeconds = uint32(ua.req.GracePeriodSeconds)
				}

				evictionRisk, willBeEvicted, err := ua.getEvictionRisk(pod)
				if err != nil {
					blog.Warnf("getEvictionRisk[%s] failed in cluster %s, node %s, err %s",
						pod.Name, ua.req.ClusterID, node, err.Error())
				}

				podCheckData = append(podCheckData, &cmproto.CheckDrainNodeData{
					PodName:            pod.Name,
					NameSpace:          pod.Namespace,
					PodStatus:          string(pod.Status.Phase),
					PodServiceAccount:  pod.Spec.ServiceAccountName,
					Node:               node,
					GracePeriodSeconds: gracePeriodSeconds,
					EvictionRisk:       evictionRisk,
					WillBeEvicted:      willBeEvicted,
				})
			}

			dataCh <- podCheckData
		}(ua.req.Nodes[i])
	}
	barrier.Wait()
	close(dataCh)

	ua.resp.Data = make([]*cmproto.CheckDrainNodeData, 0)

	for v := range dataCh {
		ua.resp.Data = append(ua.resp.Data, v...)
	}
	return nil
}

const (
	// DaemonSetController daemonset controller name
	DaemonSetController = "DaemonSet"
	// ForceRiskParameter force risk parameter
	ForceRiskParameter = "--force"
	// ForceRiskDescription force risk description
	ForceRiskDescription = "未声明控制器的Pod"
	// IgnoreAllDaemonSetsRiskParameter ignore all daemonsets risk parameter
	IgnoreAllDaemonSetsRiskParameter = "--ignore-daemonsets"
	// IgnoreAllDaemonSetsRiskDescription ignore all daemonsets risk description
	IgnoreAllDaemonSetsRiskDescription = "忽略DaemonSet所控制的Pod"
	// DeleteLocalDataParameter delete local data parameter
	DeleteLocalDataParameter = "--delete-local-data"
	// DeleteLocalDataDescription delete local data description
	DeleteLocalDataDescription = "本地存储的Pod被驱逐后数据将丢失"
	// DisableEvictionParameter disable eviction parameter
	DisableEvictionParameter = "--disable-eviction"
	// DisableEvictionDescription disable eviction description
	DisableEvictionDescription = "受PodDisruptionBudget(PDB)策略保护的Pod"
)

// getEvictionRisk get pod eviction risk
func (ua *CheckDrainNodeAction) getEvictionRisk(pod *corev1.Pod) ([]*cmproto.EvictionRisk, bool, error) { // nolint
	evictionRisks := make([]*cmproto.EvictionRisk, 0)

	// ignore-daemonsets parameter needs to check if the pod has a controller
	// and if it's managed by a daemonset controller, it is the highest priority and only risk
	if ua.isDaemonSetPod(pod) {
		evictionRisks = append(evictionRisks, &cmproto.EvictionRisk{
			RiskParameter:   IgnoreAllDaemonSetsRiskParameter,
			RiskDescription: IgnoreAllDaemonSetsRiskDescription,
		})

		return evictionRisks, false, nil
	}

	// delete-local-data parameter checks if the pod has emptydir volumes
	if ua.hasEmptyDirVolume(pod) {
		evictionRisks = append(evictionRisks, &cmproto.EvictionRisk{
			RiskParameter:   DeleteLocalDataParameter,
			RiskDescription: DeleteLocalDataDescription,
		})
	}

	// force parameter needs to check if the pod has no controller declared
	if ua.req.Force {
		if ua.isOrphanPod(pod) {
			evictionRisks = append(evictionRisks, &cmproto.EvictionRisk{
				RiskParameter:   ForceRiskParameter,
				RiskDescription: ForceRiskDescription,
			})
		}
	}

	// disable-eviction parameter
	hasPDB, err := ua.hasPodPDB(context.Background(), ua.req.ClusterID, pod)
	if err != nil {
		blog.Warnf("CheckPodPDB[%s] failed in cluster %s, err %s", pod.Name, ua.req.ClusterID, err.Error())
		return evictionRisks, true, nil
	}

	if hasPDB {
		evictionRisks = append(evictionRisks, &cmproto.EvictionRisk{
			RiskParameter:   DisableEvictionParameter,
			RiskDescription: DisableEvictionDescription,
		})
	}

	return evictionRisks, true, nil
}

func (ua *CheckDrainNodeAction) isOrphanPod(pod *corev1.Pod) bool {
	return len(pod.GetOwnerReferences()) == 0
}

func (ua *CheckDrainNodeAction) isDaemonSetPod(pod *corev1.Pod) bool {
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == DaemonSetController {
			return true
		}
	}
	return false
}

func (ua *CheckDrainNodeAction) hasEmptyDirVolume(pod *corev1.Pod) bool {
	for _, volume := range pod.Spec.Volumes {
		if volume.EmptyDir != nil {
			return true
		}
	}
	return false
}

func (ua *CheckDrainNodeAction) hasPodPDB(ctx context.Context, clusterID string, pod *corev1.Pod) (bool, error) {
	option := clusterops.CheckPodPDBOption{
		ClusterID: clusterID,
		Pod:       *pod,
	}

	ctx, cancel := context.WithTimeout(ctx, clusterops.DefaultTimeout)
	defer cancel()

	return ua.k8sOp.CheckPodPDB(ctx, option)
}

func (ua *CheckDrainNodeAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handles check node drain
func (ua *CheckDrainNodeAction) Handle(ctx context.Context, req *cmproto.CheckDrainNodeRequest,
	resp *cmproto.CheckDrainNodeResponse) {
	if req == nil || resp == nil {
		blog.Errorf("check drain cluster node failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.checkDrainClusterNodes(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
