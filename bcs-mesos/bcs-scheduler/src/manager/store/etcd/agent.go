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

package etcd

import (
	"context"

	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	schStore "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckAgentExist check agent whether exist
func (store *managerStore) CheckAgentExist(agent *types.Agent) (string, bool) {
	obj, _ := store.fetchAgentInDB(agent.Key)
	if obj != nil {
		return obj.ResourceVersion, true
	}

	return "", false
}

// SaveAgent save agent
func (store *managerStore) SaveAgent(agent *types.Agent) error {
	client := store.BkbcsClient.Agents(DefaultNamespace)
	v2Agent := &v2.Agent{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdAgent,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      agent.Key,
			Namespace: DefaultNamespace,
		},
		Spec: v2.AgentSpec{
			Agent: *agent,
		},
	}

	var err error
	rv, exist := store.CheckAgentExist(agent)
	// if exist, then update
	if exist {
		v2Agent.ResourceVersion = rv
		v2Agent, err = client.Update(context.Background(), v2Agent, metav1.UpdateOptions{})
	} else {
		// else not exist, then create it
		v2Agent, err = client.Create(context.Background(), v2Agent, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	// update kube-apiserver ResourceVersion
	agent.ResourceVersion = v2Agent.ResourceVersion
	// save agent in cache
	saveCacheAgent(agent)
	return nil
}

// FetchAgent fetch agent for agent InnerIP
func (store *managerStore) FetchAgent(Key string) (*types.Agent, error) {
	//fetch agent in cache
	agent := getCacheAgent(Key)
	if agent == nil {
		return nil, schStore.ErrNoFound
	}
	return agent, nil
}

func (store *managerStore) fetchAgentInDB(Key string) (*types.Agent, error) {
	client := store.BkbcsClient.Agents(DefaultNamespace)
	// fetch agent in kube-apiserver
	v2Agent, err := client.Get(context.Background(), Key, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, schStore.ErrNoFound
		}
		return nil, err
	}

	obj := v2Agent.Spec.Agent
	obj.ResourceVersion = v2Agent.ResourceVersion
	return &obj, nil
}

// ListAllAgents list all agent list
func (store *managerStore) ListAllAgents() ([]*types.Agent, error) {
	if cacheMgr.isOK {
		return listCacheAgents()
	}

	client := store.BkbcsClient.Agents(DefaultNamespace)
	v2Agents, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	agents := make([]*types.Agent, 0, len(v2Agents.Items))
	for _, v2 := range v2Agents.Items {
		obj := v2.Spec.Agent
		obj.ResourceVersion = v2.ResourceVersion
		agents = append(agents, &obj)
	}
	return agents, nil
}

// ListAgentNodes list all agent ip list
func (store *managerStore) ListAgentNodes() ([]string, error) {
	agents, err := store.ListAllAgents()
	if err != nil {
		return nil, err
	}
	agentNodes := make([]string, 0, len(agents))
	for _, agent := range agents {
		agentNodes = append(agentNodes, agent.Key)
	}

	return agentNodes, nil
}

// DeleteAgent delete agent for innerip
func (store *managerStore) DeleteAgent(key string) error {
	client := store.BkbcsClient.Agents(DefaultNamespace)
	err := client.Delete(context.Background(), key, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	// delete agent in cache
	deleteCacheAgent(key)
	return nil
}

// CheckAgentSettingExist check agentsetting exist
func (store *managerStore) CheckAgentSettingExist(agent *commtypes.BcsClusterAgentSetting) (string, bool) {
	obj, _ := store.fetchAgentSettingInDB(agent.InnerIP)
	if obj != nil {
		return obj.ResourceVersion, true
	}

	return "", false
}

// SaveAgentSetting save agentsetting
func (store *managerStore) SaveAgentSetting(agent *commtypes.BcsClusterAgentSetting) error {
	client := store.BkbcsClient.BcsClusterAgentSettings(DefaultNamespace)
	v2Agent := &v2.BcsClusterAgentSetting{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdAgentSetting,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      agent.InnerIP,
			Namespace: DefaultNamespace,
		},
		Spec: v2.BcsClusterAgentSettingSpec{
			BcsClusterAgentSetting: *agent,
		},
	}

	var err error
	rv, exist := store.CheckAgentSettingExist(agent)
	if exist {
		v2Agent.ResourceVersion = rv
		v2Agent, err = client.Update(context.Background(), v2Agent, metav1.UpdateOptions{})
	} else {
		v2Agent, err = client.Create(context.Background(), v2Agent, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	agent.ResourceVersion = v2Agent.ResourceVersion
	saveCacheAgentsetting(agent)
	return nil
}

// FetchAgentSetting fetch agentsetting for innerip
func (store *managerStore) FetchAgentSetting(InnerIP string) (*commtypes.BcsClusterAgentSetting, error) {
	agentSetting := getCacheAgentsetting(InnerIP)
	if agentSetting == nil {
		return nil, schStore.ErrNoFound
	}
	return agentSetting, nil
}

// fetch agentsetting for innerip
func (store *managerStore) fetchAgentSettingInDB(InnerIP string) (*commtypes.BcsClusterAgentSetting, error) {
	client := store.BkbcsClient.BcsClusterAgentSettings(DefaultNamespace)
	v2Agent, err := client.Get(context.Background(), InnerIP, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	obj := v2Agent.Spec.BcsClusterAgentSetting
	obj.ResourceVersion = v2Agent.ResourceVersion
	return &obj, nil
}

// DeleteAgentSetting delete agentsetting
func (store *managerStore) DeleteAgentSetting(InnerIP string) error {
	client := store.BkbcsClient.BcsClusterAgentSettings(DefaultNamespace)
	err := client.Delete(context.Background(), InnerIP, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	deleteCacheAgentsetting(InnerIP)
	return nil
}

// ListAgentSettingNodes list agentsetting ip list
func (store *managerStore) ListAgentSettingNodes() ([]string, error) {
	agents, err := store.ListAgentsettings()
	if err != nil {
		return nil, err
	}

	nodes := make([]string, 0, len(agents))
	for _, agent := range agents {
		nodes = append(nodes, agent.InnerIP)
	}
	return nodes, nil
}

// ListAgentsettings list all agentsetting list
func (store *managerStore) ListAgentsettings() ([]*commtypes.BcsClusterAgentSetting, error) {
	if cacheMgr.isOK {
		return listCacheAgentsettings()
	}

	client := store.BkbcsClient.BcsClusterAgentSettings(DefaultNamespace)
	v2Agents, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	agents := make([]*commtypes.BcsClusterAgentSetting, 0, len(v2Agents.Items))
	for _, v2 := range v2Agents.Items {
		obj := v2.Spec.BcsClusterAgentSetting
		obj.ResourceVersion = v2.ResourceVersion
		agents = append(agents, &obj)
	}
	return agents, nil
}

// CheckAgentSchedInfoExist check agent schedinfo exist
func (store *managerStore) CheckAgentSchedInfoExist(agent *types.AgentSchedInfo) (string, bool) {
	client := store.BkbcsClient.AgentSchedInfos(DefaultNamespace)
	obj, err := client.Get(context.Background(), agent.HostName, metav1.GetOptions{})
	if err == nil {
		return obj.ResourceVersion, true
	}

	return "", false
}

// SaveAgentSchedInfo save agentschedinfo
func (store *managerStore) SaveAgentSchedInfo(agent *types.AgentSchedInfo) error {
	client := store.BkbcsClient.AgentSchedInfos(DefaultNamespace)
	v2Agent := &v2.AgentSchedInfo{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdAgentSchedInfo,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      agent.HostName,
			Namespace: DefaultNamespace,
		},
		Spec: v2.AgentSchedInfoSpec{
			AgentSchedInfo: *agent,
		},
	}

	var err error
	rv, exist := store.CheckAgentSchedInfoExist(agent)
	if exist {
		v2Agent.ResourceVersion = rv
		_, err = client.Update(context.Background(), v2Agent, metav1.UpdateOptions{})
	} else {
		_, err = client.Create(context.Background(), v2Agent, metav1.CreateOptions{})
	}
	return err
}

// FetchAgentSchedInfo fetch agentschedinfo
func (store *managerStore) FetchAgentSchedInfo(HostName string) (*types.AgentSchedInfo, error) {
	client := store.BkbcsClient.AgentSchedInfos(DefaultNamespace)
	v2Agent, err := client.Get(context.Background(), HostName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &v2Agent.Spec.AgentSchedInfo, nil
}

// DeleteAgentSchedInfo delete agent schedinfo
func (store *managerStore) DeleteAgentSchedInfo(HostName string) error {
	client := store.BkbcsClient.AgentSchedInfos(DefaultNamespace)
	err := client.Delete(context.Background(), HostName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}

// ListAgentSchedInfoNodes list agentschedinfo node
func (store *managerStore) ListAgentSchedInfoNodes() ([]string, error) {
	agentSchedInfos, err := store.ListAgentSchedInfo()
	if err != nil {
		return nil, err
	}

	nodes := make([]string, 0, len(agentSchedInfos))
	for _, agent := range agentSchedInfos {
		nodes = append(nodes, agent.HostName)
	}
	return nodes, nil
}

// ListAgentSchedInfo list agentschedinfo
func (store *managerStore) ListAgentSchedInfo() ([]*types.AgentSchedInfo, error) {
	client := store.BkbcsClient.AgentSchedInfos(DefaultNamespace)
	v2AgentschedInfos, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	agentSchedInfos := make([]*types.AgentSchedInfo, 0, len(v2AgentschedInfos.Items))
	for _, v2 := range v2AgentschedInfos.Items {
		obj := v2.Spec.AgentSchedInfo
		agentSchedInfos = append(agentSchedInfos, &obj)
	}
	return agentSchedInfos, nil
}
