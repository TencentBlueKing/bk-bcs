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

package zk

import (
	"encoding/json"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	schStore "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"

	"github.com/samuel/go-zookeeper/zk"
)

func getAgentRootPath() string {
	return "/" + bcsRootNode + "/" + agentNode
}

func getAgentSettingRootPath() string {
	return "/" + bcsRootNode + "/" + agentSettingNode
}

func getAgentSchedInfoRootPath() string {
	return "/" + bcsRootNode + "/" + agentSchedInfoNode
}

// SaveAgent save agent object
func (store *managerStore) SaveAgent(agent *types.Agent) error {

	data, err := json.Marshal(agent)
	if err != nil {
		return err
	}

	path := getAgentRootPath() + "/" + agent.Key

	return store.Db.Insert(path, string(data))
}

// FetchAgent get agent object
func (store *managerStore) FetchAgent(Key string) (*types.Agent, error) {

	path := getAgentRootPath() + "/" + Key

	data, err := store.Db.Fetch(path)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil, schStore.ErrNoFound
		}
		return nil, err
	}

	agent := &types.Agent{}
	if err := json.Unmarshal(data, agent); err != nil {
		blog.Error("fail to unmarshal agent(%s). err:%s", string(data), err.Error())
		return nil, err
	}

	return agent, nil
}

// ListAgentNodes list agent object nodes
func (store *managerStore) ListAgentNodes() ([]string, error) {

	path := getAgentRootPath()

	agentNodes, err := store.Db.List(path)
	if err != nil {
		blog.Error("fail to list agents(%s), err:%s", path, err.Error())
		return nil, err
	}

	return agentNodes, nil
}

// DeleteAgent delete agent by key
func (store *managerStore) DeleteAgent(key string) error {

	path := getAgentRootPath() + "/" + key
	if err := store.Db.Delete(path); err != nil {
		blog.Error("fail to delete agent(%s) err:%s", path, err.Error())
		return err
	}

	return nil
}

// SaveAgentSetting save agent setting info
func (store *managerStore) SaveAgentSetting(agent *commtypes.BcsClusterAgentSetting) error {

	data, err := json.Marshal(agent)
	if err != nil {
		return err
	}

	path := getAgentSettingRootPath() + "/" + agent.InnerIP

	return store.Db.Insert(path, string(data))
}

// FetchAgentSetting get agent setting info
func (store *managerStore) FetchAgentSetting(InnerIP string) (*commtypes.BcsClusterAgentSetting, error) {

	path := getAgentSettingRootPath() + "/" + InnerIP

	data, err := store.Db.Fetch(path)

	if err == zk.ErrNoNode {
		blog.V(3).Infof("agentSetting(%s) not exist", path)
		return nil, schStore.ErrNoFound
	}

	if err != nil {
		return nil, err
	}

	agent := &commtypes.BcsClusterAgentSetting{}
	if err := json.Unmarshal(data, agent); err != nil {
		blog.Error("fail to unmarshal agentSetting(%s). err:%s", string(data), err.Error())
		return nil, err
	}

	return agent, nil
}

// DeleteAgentSetting delete agent setting by inner IP
func (store *managerStore) DeleteAgentSetting(InnerIP string) error {

	path := getAgentSettingRootPath() + "/" + InnerIP
	if err := store.Db.Delete(path); err != nil {
		if err == zk.ErrNoNode {
			return nil
		}
		blog.Error("fail to delete agentSetting(%s) err:%s", path, err.Error())
		return err
	}
	return nil
}

// ListAgentSettingNodes list agent setting node names
func (store *managerStore) ListAgentSettingNodes() ([]string, error) {

	path := getAgentSettingRootPath()

	agentNodes, err := store.Db.List(path)
	if err != nil {
		blog.Error("fail to list agentsettings(%s), err:%s", path, err.Error())
		return nil, err
	}

	return agentNodes, nil
}

// ListAgentsettings get agent setting list
func (store *managerStore) ListAgentsettings() ([]*commtypes.BcsClusterAgentSetting, error) {
	nodes, err := store.ListAgentSettingNodes()
	if err != nil {
		return nil, err
	}

	settings := make([]*commtypes.BcsClusterAgentSetting, 0, len(nodes))
	for _, node := range nodes {
		setting, err := store.FetchAgentSetting(node)
		if err != nil {
			return nil, err
		}

		settings = append(settings, setting)
	}

	return settings, nil
}

// SaveAgentSchedInfo save agent schedule info
func (store *managerStore) SaveAgentSchedInfo(agent *types.AgentSchedInfo) error {

	data, err := json.Marshal(agent)
	if err != nil {
		return err
	}

	path := getAgentSchedInfoRootPath() + "/" + agent.HostName

	return store.Db.Insert(path, string(data))
}

// FetchAgentSchedInfo get agent schedule info
func (store *managerStore) FetchAgentSchedInfo(HostName string) (*types.AgentSchedInfo, error) {

	path := getAgentSchedInfoRootPath() + "/" + HostName

	data, err := store.Db.Fetch(path)

	if err == zk.ErrNoNode {
		blog.V(3).Infof("agentSchedInfo(%s) not exist", path)
		return nil, schStore.ErrNoFound
	}

	if err != nil {
		return nil, err
	}

	agent := &types.AgentSchedInfo{}
	if err := json.Unmarshal(data, agent); err != nil {
		blog.Error("fail to unmarshal agentSchedInfo(%s). err:%s", string(data), err.Error())
		return nil, err
	}

	return agent, nil
}

// DeleteAgentSchedInfo delete agent schedule info
func (store *managerStore) DeleteAgentSchedInfo(HostName string) error {

	path := getAgentSchedInfoRootPath() + "/" + HostName
	if err := store.Db.Delete(path); err != nil {
		blog.Error("fail to delete agentSchedInfo(%s) err:%s", path, err.Error())
		return err
	}
	return nil
}

// ListAgentSchedInfoNodes list agent schedule info node names
func (store *managerStore) ListAgentSchedInfoNodes() ([]string, error) {
	path := getAgentSchedInfoRootPath()

	agentSchedInfoNodes, err := store.Db.List(path)
	if err != nil {
		blog.Error("fail to list agentschedinfo(%s), err:%s", path, err.Error())
		return nil, err
	}

	return agentSchedInfoNodes, nil
}

// ListAgentSchedInfo list agent schedule info
func (store *managerStore) ListAgentSchedInfo() ([]*types.AgentSchedInfo, error) {
	nodes, err := store.ListAgentSchedInfoNodes()
	if err != nil {
		return nil, err
	}

	schedinfos := make([]*types.AgentSchedInfo, 0, len(nodes))
	for _, node := range nodes {
		info, err := store.FetchAgentSchedInfo(node)
		if err != nil {
			return nil, err
		}

		schedinfos = append(schedinfos, info)
	}

	return schedinfos, nil
}

// ListAllAgents list all agents
func (store *managerStore) ListAllAgents() ([]*types.Agent, error) {
	nodes, err := store.ListAgentNodes()
	if err != nil {
		return nil, err
	}

	var agents []*types.Agent
	for _, node := range nodes {
		agent, err := store.FetchAgent(node)
		if err != nil {
			blog.Errorf("fetch agent %s error %s", node, err.Error())
			continue
		}

		agents = append(agents, agent)
	}

	return agents, nil
}
