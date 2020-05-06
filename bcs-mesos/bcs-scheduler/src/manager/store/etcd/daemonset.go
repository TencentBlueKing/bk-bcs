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
	commtypes "bk-bcs/bcs-common/common/types"
	schStore "bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"bk-bcs/bcs-mesos/pkg/apis/bkbcs/v2"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//check agent whether exist
func (store *managerStore) CheckDaemonsetExist(agent *types.Agent) (string, bool) {
	obj, _ := store.FetchAgent(agent.Key)
	if obj != nil {
		return obj.ResourceVersion, true
	}

	return "", false
}

//save agent
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
	//if exist, then update
	if exist {
		v2Agent.ResourceVersion = rv
		v2Agent, err = client.Update(v2Agent)
		//else not exist, then create it
	} else {
		v2Agent, err = client.Create(v2Agent)
	}
	if err != nil {
		return err
	}

	//update kube-apiserver ResourceVersion
	agent.ResourceVersion = v2Agent.ResourceVersion
	//save agent in cache
	saveCacheAgent(agent)
	return nil
}

//fetch agent for agent InnerIP
func (store *managerStore) FetchAgent(Key string) (*types.Agent, error) {
	//fetch agent in cache
	if cacheMgr.isOK {
		agent := getCacheAgent(Key)
		if agent == nil {
			return nil, schStore.ErrNoFound
		}
		return agent, nil
	}

	client := store.BkbcsClient.Agents(DefaultNamespace)
	//fetch agent in kube-apiserver
	v2Agent, err := client.Get(Key, metav1.GetOptions{})
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

//list all agent list
func (store *managerStore) ListAllAgents() ([]*types.Agent, error) {
	if cacheMgr.isOK {
		return listCacheAgents()
	}

	client := store.BkbcsClient
	v2Agents, err := client.List(metav1.ListOptions{})
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

//list all agent ip list
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

//delete agent for innerip
func (store *managerStore) DeleteAgent(key string) error {
	client := store.BkbcsClient.Agents(DefaultNamespace)
	err := client.Delete(key, &metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	//delete agent in cache
	deleteCacheAgent(key)
	return nil
}