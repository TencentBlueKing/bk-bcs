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

// Package tasks include all tasks for bcs-federation-manager
package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
	steps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps/sync_namespace_steps"
)

var (
	// HandleNamespaceQuotaTaskName step name for create cluster
	HandleNamespaceQuotaTaskName = TaskNames{
		Name: "handle quota and subCluster namespace",
		Type: "HANDLE QUOTA AND SUBCLUSTER NAMESPACE",
	}
)

// NewHandleNamespaceQuotaTask new federation task
func NewHandleNamespaceQuotaTask(opt *HandleNamespaceQuotaOptions) *HandleNamespaceQuota {
	return &HandleNamespaceQuota{
		opt: opt,
	}
}

// HandleNamespaceQuotaOptions handle quota options
type HandleNamespaceQuotaOptions struct {
	HandleType    string // create or update
	FedClusterId  string // 联邦proxy集群id
	HostClusterId string // 联邦host集群id
	Namespace     string // 命名空间
	Parameter     string
}

// HandleNamespaceQuota federation task
type HandleNamespaceQuota struct {
	opt *HandleNamespaceQuotaOptions
}

// Name return name of task
func (i *HandleNamespaceQuota) Name() string {
	return HandleNamespaceQuotaTaskName.Name
}

// Type 任务类型
func (i *HandleNamespaceQuota) Type() string {
	return HandleNamespaceQuotaTaskName.Type
}

// Steps build steps for task
func (i *HandleNamespaceQuota) Steps() []*types.Step {
	allSteps := make([]*types.Step, 0)

	checkNamespaceQuotaParamStep := steps.NewCheckNamespaceQuotaStep().BuildStep([]task.KeyValue{
		{Key: fedsteps.ParameterKey, Value: i.opt.Parameter},
	})

	allSteps = append(allSteps, checkNamespaceQuotaParamStep)

	reqMap := make(map[string]string)
	_ = json.Unmarshal([]byte(i.opt.Parameter), &reqMap)

	for key, reqListStr := range reqMap {
		switch key {
		case fedsteps.ClusterQuotaKey:
			handleFederationNamespaceQuotaStep := steps.NewHandleFederationNamespaceQuotaStep().BuildStep([]task.KeyValue{
				{Key: fedsteps.HandleTypeKey, Value: i.opt.HandleType},
				{Key: fedsteps.NamespaceKey, Value: i.opt.Namespace},
				{Key: fedsteps.HostClusterIdKey, Value: i.opt.HostClusterId},
				{Key: fedsteps.ParameterKey, Value: reqListStr},
			})
			allSteps = append(allSteps, handleFederationNamespaceQuotaStep)
		}
	}

	// 更新状态step
	updateFederationNamespaceStatusStep := steps.NewUpdateFederationNamespaceStatusStep().BuildStep([]task.KeyValue{
		{Key: fedsteps.HostClusterIdKey, Value: i.opt.HostClusterId},
		{Key: fedsteps.NamespaceKey, Value: i.opt.Namespace},
	})

	allSteps = append(allSteps, updateFederationNamespaceStatusStep)

	return allSteps
}

// BuildTask build task with steps
func (i *HandleNamespaceQuota) BuildTask(creator string, opts ...types.TaskOption) (*types.Task, error) {
	if i.opt.HostClusterId == "" || i.opt.Parameter == "" {
		return nil, fmt.Errorf("handle namespace and quota task parameter is empty")
	}

	t := types.NewTask(&types.TaskInfo{
		TaskIndex: fmt.Sprintf("%s/%s", i.opt.HostClusterId, i.opt.Namespace),
		TaskType:  i.Type(),
		TaskName:  i.Name(),
		Creator:   creator,
	}, opts...)
	if len(i.Steps()) == 0 {
		return nil, fmt.Errorf("task steps empty")
	}

	for _, step := range i.Steps() {
		t.Steps[step.GetName()] = step
		t.StepSequence = append(t.StepSequence, step.GetName())
	}

	t.CurrentStep = t.StepSequence[0]

	// record task params
	t.AddCommonParams(fedsteps.FedClusterIdKey, i.opt.FedClusterId).
		AddCommonParams(fedsteps.HostClusterIdKey, i.opt.HostClusterId).
		AddCommonParams(fedsteps.NamespaceKey, i.opt.Namespace)

	return t, nil
}

var (
	// SyncTjNamespaceQuotaTaskName step name  sync taiji namespace quota
	SyncTjNamespaceQuotaTaskName = TaskNames{
		Name: "sync_tj_namespace_quota",
		Type: "SYNC_TJ_NAMESPACE_QUOTA",
	}
)

// NewSyncTjNamespaceQuotaTask new sync namespace quota task
func NewSyncTjNamespaceQuotaTask(opt *SyncTjNamespaceQuotaOptions) *SyncTjNamespaceQuota {
	return &SyncTjNamespaceQuota{
		opt: opt,
	}
}

// SyncTjNamespaceQuotaOptions sync namespace quota options
type SyncTjNamespaceQuotaOptions struct {
	Namespace     string // 命名空间
	HostClusterID string // 联邦集群hostID
	FederationID  string // 联邦集群federationID
}

// SyncTjNamespaceQuota handle namespace quota
type SyncTjNamespaceQuota struct {
	opt *SyncTjNamespaceQuotaOptions
}

// Name return name of task
func (s *SyncTjNamespaceQuota) Name() string {
	return SyncTjNamespaceQuotaTaskName.Name
}

// Type 任务类型
func (s *SyncTjNamespaceQuota) Type() string {
	return SyncTjNamespaceQuotaTaskName.Type
}

// Steps build steps for task
func (s *SyncTjNamespaceQuota) Steps() []*types.Step {
	allSteps := make([]*types.Step, 0)
	getNamespaceQuotaStep := steps.NewGetNamespaceQuotaStep().BuildStep([]task.KeyValue{})
	allSteps = append(allSteps, getNamespaceQuotaStep)
	checkInTaijiStep := steps.NewCheckInTaijiStep().BuildStep([]task.KeyValue{})
	allSteps = append(allSteps, checkInTaijiStep)
	// 更新状态step
	updateFederationNamespaceStatusStep := steps.NewUpdateFederationNamespaceStatusStep().BuildStep([]task.KeyValue{
		{Key: fedsteps.HostClusterIdKey, Value: s.opt.HostClusterID},
		{Key: fedsteps.NamespaceKey, Value: s.opt.Namespace},
	})
	allSteps = append(allSteps, updateFederationNamespaceStatusStep)

	return allSteps
}

// BuildTask build task with steps
func (s *SyncTjNamespaceQuota) BuildTask(creator string, opts ...types.TaskOption) (*types.Task, error) {
	if s.opt.HostClusterID == "" {
		blog.Errorf("syncNamespaceQuota task build failed, HostClusterId is empty, opt: %v", s.opt)
		return nil, fmt.Errorf("syncNamespaceQuota task parameter is empty")
	}

	t := types.NewTask(&types.TaskInfo{
		TaskIndex: fmt.Sprintf("%s/%s", s.opt.HostClusterID, s.opt.Namespace),
		TaskType:  s.Type(),
		TaskName:  s.Name(),
		Creator:   creator,
	}, opts...)
	if len(s.Steps()) == 0 {
		blog.Errorf("syncNamespaceQuota task steps empty")
		return nil, fmt.Errorf("task steps empty")
	}

	for _, step := range s.Steps() {
		t.Steps[step.GetName()] = step
		t.StepSequence = append(t.StepSequence, step.GetName())
	}

	t.CurrentStep = t.StepSequence[0]
	t.AddCommonParams(fedsteps.NamespaceKey, s.opt.Namespace)
	t.AddCommonParams(fedsteps.HostClusterIdKey, s.opt.HostClusterID)
	blog.Infof("syncNamespaceQuota task build success, taskID: %s, taskName: %s", t.TaskID, t.TaskName)
	return t, nil
}

var (
	// SyncSlNamespaceQuotaTaskName step name  sync suanli namespace quota
	SyncSlNamespaceQuotaTaskName = TaskNames{
		Name: "sync_sl_namespace_quota",
		Type: "SYNC_SL_NAMESPACE_QUOTA",
	}
)

// NewSyncSlNamespaceQuotaTask new sync namespace quota task
func NewSyncSlNamespaceQuotaTask(opt *SyncSlNamespaceQuotaOptions) *SyncSlNamespaceQuota {
	return &SyncSlNamespaceQuota{
		opt: opt,
	}
}

// SyncSlNamespaceQuotaOptions sync namespace quota options
type SyncSlNamespaceQuotaOptions struct {
	Namespace     string // 命名空间
	HostClusterID string // 联邦集群hostID
}

// SyncSlNamespaceQuota handle namespace quota
type SyncSlNamespaceQuota struct {
	opt *SyncSlNamespaceQuotaOptions
}

// Name return name of task
func (s *SyncSlNamespaceQuota) Name() string {
	return SyncSlNamespaceQuotaTaskName.Name
}

// Type 任务类型
func (s *SyncSlNamespaceQuota) Type() string {
	return SyncSlNamespaceQuotaTaskName.Type
}

// Steps build steps for task
func (s *SyncSlNamespaceQuota) Steps() []*types.Step {
	allSteps := make([]*types.Step, 0)
	getNamespaceQuotaStep := steps.NewGetNamespaceQuotaStep().BuildStep([]task.KeyValue{})
	allSteps = append(allSteps, getNamespaceQuotaStep)
	checkInSuanliStep := steps.NewCheckInSuanliStep().BuildStep([]task.KeyValue{})
	allSteps = append(allSteps, checkInSuanliStep)
	// 更新状态step
	updateFedNamespaceStatusStep := steps.NewUpdateFederationNamespaceStatusStep().BuildStep([]task.KeyValue{
		{Key: fedsteps.HostClusterIdKey, Value: s.opt.HostClusterID},
		{Key: fedsteps.NamespaceKey, Value: s.opt.Namespace},
	})
	allSteps = append(allSteps, updateFedNamespaceStatusStep)
	return allSteps
}

// BuildTask build task with steps
func (s *SyncSlNamespaceQuota) BuildTask(creator string, opts ...types.TaskOption) (*types.Task, error) {
	if s.opt.HostClusterID == "" {
		return nil, fmt.Errorf("handle namespace and quota task parameter is empty")
	}

	t := types.NewTask(&types.TaskInfo{
		TaskIndex: fmt.Sprintf("%s/%s", s.opt.HostClusterID, s.opt.Namespace),
		TaskType:  s.Type(),
		TaskName:  s.Name(),
		Creator:   creator,
	}, opts...)
	if len(s.Steps()) == 0 {
		return nil, fmt.Errorf("task steps empty")
	}

	for _, step := range s.Steps() {
		t.Steps[step.GetName()] = step
		t.StepSequence = append(t.StepSequence, step.GetName())
	}

	t.CurrentStep = t.StepSequence[0]
	t.AddCommonParams(fedsteps.NamespaceKey, s.opt.Namespace)
	t.AddCommonParams(fedsteps.HostClusterIdKey, s.opt.HostClusterID)

	return t, nil
}

var (
	// SyncHbNamespaceQuotaTaskName step name  sync 混部 namespace quota
	SyncHbNamespaceQuotaTaskName = TaskNames{
		Name: "sync_hb_namespace_quota",
		Type: "SYNC_HB_NAMESPACE_QUOTA",
	}
)

// NewSyncHbNamespaceQuotaTask new sync namespace quota task
func NewSyncHbNamespaceQuotaTask(opt *SyncHbNamespaceQuotaOptions) *SyncHbNamespaceQuota {
	return &SyncHbNamespaceQuota{
		opt: opt,
	}
}

// SyncHbNamespaceQuotaOptions sync namespace quota options
type SyncHbNamespaceQuotaOptions struct {
	Namespace     string // 命名空间
	HostClusterID string // 联邦集群hostID
	SubClusterID  string // 子集群id
	Labels        string
}

// SyncHbNamespaceQuota handle namespace quota
type SyncHbNamespaceQuota struct {
	opt *SyncHbNamespaceQuotaOptions
}

// Name return name of task
func (s *SyncHbNamespaceQuota) Name() string {
	return SyncHbNamespaceQuotaTaskName.Name
}

// Type 任务类型
func (s *SyncHbNamespaceQuota) Type() string {
	return SyncHbNamespaceQuotaTaskName.Type
}

// Steps build steps for task
func (s *SyncHbNamespaceQuota) Steps() []*types.Step {
	allSteps := make([]*types.Step, 0)

	checkInHunbuStep := steps.NewCheckInHunbuStep().BuildStep([]task.KeyValue{})
	allSteps = append(allSteps, checkInHunbuStep)
	// 更新状态step
	updateFedNamespaceStatusStep := steps.NewUpdateFederationNamespaceStatusStep().BuildStep([]task.KeyValue{
		{Key: fedsteps.HostClusterIdKey, Value: s.opt.HostClusterID},
		{Key: fedsteps.NamespaceKey, Value: s.opt.Namespace},
	})
	allSteps = append(allSteps, updateFedNamespaceStatusStep)

	return allSteps
}

// BuildTask build task with steps
func (s *SyncHbNamespaceQuota) BuildTask(creator string, opts ...types.TaskOption) (*types.Task, error) {
	if s.opt.HostClusterID == "" {
		return nil, fmt.Errorf("handle namespace and quota task parameter is empty")
	}

	t := types.NewTask(&types.TaskInfo{
		TaskIndex: fmt.Sprintf("%s/%s", s.opt.HostClusterID, s.opt.Namespace),
		TaskType:  s.Type(),
		TaskName:  s.Name(),
		Creator:   creator,
	}, opts...)
	if len(s.Steps()) == 0 {
		return nil, fmt.Errorf("task steps empty")
	}

	for _, step := range s.Steps() {
		t.Steps[step.GetName()] = step
		t.StepSequence = append(t.StepSequence, step.GetName())
	}

	t.CurrentStep = t.StepSequence[0]
	t.AddCommonParams(fedsteps.NamespaceKey, s.opt.Namespace)
	t.AddCommonParams(fedsteps.HostClusterIdKey, s.opt.HostClusterID)
	t.AddCommonParams(fedsteps.SubClusterIdKey, s.opt.SubClusterID)
	t.AddCommonParams(fedsteps.ManagedClusterLabelsKey, s.opt.Labels)

	return t, nil
}

var (
	// SyncNormalNamespaceQuotaTaskName step name  sync 普通 namespace quota
	SyncNormalNamespaceQuotaTaskName = TaskNames{
		Name: "sync_normal_namespace_quota",
		Type: "SYNC_NORMAL_NAMESPACE_QUOTA",
	}
)

// NewSyncNormalNamespaceQuotaTask new sync namespace quota task
func NewSyncNormalNamespaceQuotaTask(opt *SyncNormalNamespaceQuotaOptions) *SyncNormalNamespaceQuota {
	return &SyncNormalNamespaceQuota{
		opt: opt,
	}
}

// SyncNormalNamespaceQuotaOptions sync namespace quota options
type SyncNormalNamespaceQuotaOptions struct {
	Namespace     string // 命名空间
	HostClusterID string // 联邦集群hostID
	SubClusterID  string // 子集群id
}

// SyncNormalNamespaceQuota handle namespace quota
type SyncNormalNamespaceQuota struct {
	opt *SyncNormalNamespaceQuotaOptions
}

// Name return name of task
func (s *SyncNormalNamespaceQuota) Name() string {
	return SyncNormalNamespaceQuotaTaskName.Name
}

// Type 任务类型
func (s *SyncNormalNamespaceQuota) Type() string {
	return SyncNormalNamespaceQuotaTaskName.Type
}

// Steps build steps for task
func (s *SyncNormalNamespaceQuota) Steps() []*types.Step {
	allSteps := make([]*types.Step, 0)

	getNamespaceQuotaStep := steps.NewGetNamespaceQuotaStep().BuildStep([]task.KeyValue{})
	allSteps = append(allSteps, getNamespaceQuotaStep)
	checkInNormalStep := steps.NewCheckInNormalStep().BuildStep([]task.KeyValue{})
	allSteps = append(allSteps, checkInNormalStep)
	// 更新状态step
	updateFedNamespaceStatusStep := steps.NewUpdateFederationNamespaceStatusStep().BuildStep([]task.KeyValue{
		{Key: fedsteps.HostClusterIdKey, Value: s.opt.HostClusterID},
		{Key: fedsteps.NamespaceKey, Value: s.opt.Namespace},
	})
	allSteps = append(allSteps, updateFedNamespaceStatusStep)

	return allSteps
}

// BuildTask build task with steps
func (s *SyncNormalNamespaceQuota) BuildTask(creator string, opts ...types.TaskOption) (*types.Task, error) {
	if s.opt.HostClusterID == "" {
		return nil, fmt.Errorf("handle namespace and quota task parameter is empty")
	}

	t := types.NewTask(&types.TaskInfo{
		TaskIndex: fmt.Sprintf("%s/%s", s.opt.HostClusterID, s.opt.Namespace),
		TaskType:  s.Type(),
		TaskName:  s.Name(),
		Creator:   creator,
	}, opts...)
	if len(s.Steps()) == 0 {
		return nil, fmt.Errorf("task steps empty")
	}

	for _, step := range s.Steps() {
		t.Steps[step.GetName()] = step
		t.StepSequence = append(t.StepSequence, step.GetName())
	}

	t.CurrentStep = t.StepSequence[0]
	t.AddCommonParams(fedsteps.NamespaceKey, s.opt.Namespace)
	t.AddCommonParams(fedsteps.HostClusterIdKey, s.opt.HostClusterID)
	t.AddCommonParams(fedsteps.SubClusterIdKey, s.opt.SubClusterID)

	return t, nil
}
