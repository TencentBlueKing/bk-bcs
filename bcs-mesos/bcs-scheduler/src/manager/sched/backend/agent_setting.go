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

package backend

import (
	comm "bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	commtypes "bk-bcs/bcs-common/common/types"
	"errors"

	"github.com/samuel/go-zookeeper/zk"
)

// DisableAgent setting agent unschedulable
func (b *backend) DisableAgent(IP string) error {

	agent, err := b.store.FetchAgentSetting(IP)
	if err != nil {
		blog.Error("fetch agent setting(%s) from db fail:%s", IP, err.Error())
		return err
	}

	blog.Infof("disable agent(%s)", IP)

	if agent != nil {
		agent.Disabled = true
		return b.store.SaveAgentSetting(agent)
	}

	agentNew := commtypes.BcsClusterAgentSetting{
		InnerIP:  IP,
		Disabled: true,
	}
	return b.store.SaveAgentSetting(&agentNew)
}

// EnableAgent enable Agent schedulable
func (b *backend) EnableAgent(IP string) error {

	agent, err := b.store.FetchAgentSetting(IP)
	if err != nil {
		blog.Error("fetch agent setting(%s) from db fail:%s", IP, err.Error())
		return err
	}

	blog.Infof("enable agent(%s)", IP)

	if agent != nil {
		agent.Disabled = false
		return b.store.SaveAgentSetting(agent)
	}

	agentNew := commtypes.BcsClusterAgentSetting{
		InnerIP:  IP,
		Disabled: false,
	}
	return b.store.SaveAgentSetting(&agentNew)
}

//QueryAgentSetting by IP address
func (b *backend) QueryAgentSetting(IP string) (*commtypes.BcsClusterAgentSetting, error) {

	agent, err := b.store.FetchAgentSetting(IP)
	if err != nil {
		blog.Error("fetch agent setting(%s) from db fail:%s", IP, err.Error())
		return nil, err
	}

	return agent, nil
}

//QueryAgentSettingList by IP address list
func (b *backend) QueryAgentSettingList(IPs []string) ([]*commtypes.BcsClusterAgentSetting, int, error) {

	IPList := IPs
	if IPs == nil || len(IPs) <= 0 {
		IPList, _ = b.store.ListAgentSettingNodes()
	}

	var agents []*commtypes.BcsClusterAgentSetting

	for _, IP := range IPList {
		agent, err := b.store.FetchAgentSetting(IP)
		if err != nil && err != zk.ErrNoNode {
			blog.Error("fetch agent setting(%s) from db fail:%s", IP, err.Error())
			return nil, comm.BcsErrCommGetZkNodeFail, err
		}

		blog.V(3).Infof("fetch agent setting(%s)[%+v] from zk", IP, agent)

		if agent != nil {
			agents = append(agents, agent)
		} else {
			blog.Warn("fetch agent setting(%s) from zk, not exist", IP)
		}
	}

	return agents, comm.BcsSuccess, nil
}

//DeleteAgentSettingList clean agent setting by IP address list
func (b *backend) DeleteAgentSettingList(IPs []string) (int, error) {

	for _, IP := range IPs {

		blog.Infof("delete agent setting(%s) from zk", IP)

		err := b.store.DeleteAgentSetting(IP)
		if err != nil && err != zk.ErrNoNode {
			blog.Error("delete agent setting(%s) from zk err: %s", IP, err.Error())
			return comm.BcsErrCommDeleteZkNodeFail, err
		}
	}

	return comm.BcsSuccess, nil
}

//SetAgentSettingList setting agent by detail info
func (b *backend) SetAgentSettingList(agents []*commtypes.BcsClusterAgentSetting) (int, error) {

	for _, agent := range agents {
		err := b.store.SaveAgentSetting(agent)
		if err != nil {
			blog.Error("save agent setting(%s) [%+v] to zk err:%s", agent.InnerIP, agent, err.Error())
			return comm.BcsErrCommCreateZkNodeFail, err
		}

		blog.Infof("save agent setting(%s) [%+v] to zk", agent.InnerIP, agent)
	}

	return comm.BcsSuccess, nil
}

func (b *backend) DisableAgentList(IPs []string) (int, error) {

	for _, IP := range IPs {

		agent, err := b.store.FetchAgentSetting(IP)
		if err != nil {
			blog.Error("fetch agent setting(%s) from db fail:%s", IP, err.Error())
			return comm.BcsErrCommGetZkNodeFail, err
		}

		blog.Infof("disable agent(%s)", IP)

		if agent != nil {
			agent.Disabled = true
			err = b.store.SaveAgentSetting(agent)
			if err != nil {
				blog.Error("save agent(%s)(%+v) fail", IP, agent)
				return comm.BcsErrCommGetZkNodeFail, err
			}
		} else {
			agent := commtypes.BcsClusterAgentSetting{
				InnerIP:  IP,
				Disabled: true,
			}
			err = b.store.SaveAgentSetting(&agent)
			if err != nil {
				blog.Error("save agent(%s)(%+v) fail", IP, agent)
				return comm.BcsErrCommGetZkNodeFail, err
			}
		}
	}

	return comm.BcsSuccess, nil
}

//EnableAgentList enable agent schedulable by IP address list
func (b *backend) EnableAgentList(IPs []string) (int, error) {

	for _, IP := range IPs {

		agent, err := b.store.FetchAgentSetting(IP)
		if err != nil {
			blog.Error("fetch agent setting(%s) from db fail:%s", IP, err.Error())
			return comm.BcsErrCommGetZkNodeFail, err
		}

		blog.Infof("enable agent(%s)", IP)

		if agent != nil {
			agent.Disabled = false
			err = b.store.SaveAgentSetting(agent)
			if err != nil {
				blog.Error("save agent(%s)(%+v) fail", IP, agent)
				return comm.BcsErrCommCreateZkNodeFail, err
			}
		} else {
			agent := commtypes.BcsClusterAgentSetting{
				InnerIP:  IP,
				Disabled: false,
			}
			err = b.store.SaveAgentSetting(&agent)
			if err != nil {
				blog.Error("save agent(%s)(%+v) fail", IP, agent)
				return comm.BcsErrCommCreateZkNodeFail, err
			}
		}
	}

	return comm.BcsSuccess, nil
}

//UpdateAgentSettingList update agent setting by details
func (b *backend) UpdateAgentSettingList(update *commtypes.BcsClusterAgentSettingUpdate) (int, error) {

	if len(update.IPs) <= 0 {
		return comm.BcsErrCommRequestDataErr, errors.New("no ips to update")
	}
	if update.SettingName == "" {
		return comm.BcsErrCommRequestDataErr, errors.New("no settingName to update")
	}

	if update.ValueType == commtypes.MesosValueType_Text {

		if update.ValueText == nil {
			blog.Error("update agentsetting, valueText is nil")
			return comm.BcsErrCommRequestDataErr, errors.New("text is nil")
		}

		for _, IP := range update.IPs {
			agent, err := b.store.FetchAgentSetting(IP)
			if err != nil && err != zk.ErrNoNode {
				blog.Error("fetch agent setting(%s) from db fail:%s", IP, err.Error())
				return comm.BcsErrCommGetZkNodeFail, err
			}

			if agent == nil {
				blog.Infof("update agentsetting, agent(%s) not exist, to create", IP)
				agent = &commtypes.BcsClusterAgentSetting{
					InnerIP:  IP,
					Disabled: false,
				}
			}

			if agent.AttrStrings == nil {
				blog.Infof("update agentsetting, agent(%s) attrStrings is nil, to create", IP)
				agent.AttrStrings = make(map[string]commtypes.MesosValue_Text)
			}

			blog.Infof("update agentsetting, agent(%s: %s -> %s) ", IP, update.SettingName, *update.ValueText)
			agent.AttrStrings[update.SettingName] = *update.ValueText
			err = b.store.SaveAgentSetting(agent)
			if err != nil {
				blog.Error("save agent(%s)(%+v) fail", IP, agent)
				return comm.BcsErrCommCreateZkNodeFail, err
			}
		}

	} else if update.ValueType == commtypes.MesosValueType_Scalar {

		if update.ValueScalar == nil {
			blog.Error("update agentsetting, valueScalar is nil")
			return comm.BcsErrCommRequestDataErr, errors.New("scalar is nil")
		}

		for _, IP := range update.IPs {
			agent, err := b.store.FetchAgentSetting(IP)
			if err != nil && err != zk.ErrNoNode {
				blog.Error("fetch agent setting(%s) from db fail:%s", IP, err.Error())
				return comm.BcsErrCommGetZkNodeFail, err
			}

			if agent == nil {
				blog.Infof("update agentsetting, agent(%s) not exist, to create", IP)
				agent = &commtypes.BcsClusterAgentSetting{
					InnerIP:  IP,
					Disabled: false,
				}
			}

			if agent.AttrScalars == nil {
				blog.Infof("update agentsetting, agent(%s) attrScalars is nil, to create", IP)
				agent.AttrScalars = make(map[string]commtypes.MesosValue_Scalar)
			}

			blog.Infof("update agentsetting, agent(%s: %s -> %d) ", IP, update.SettingName, *update.ValueScalar)
			agent.AttrScalars[update.SettingName] = *update.ValueScalar
			err = b.store.SaveAgentSetting(agent)
			if err != nil {
				blog.Error("save agent(%s)(%+v) fail", IP, agent)
				return comm.BcsErrCommCreateZkNodeFail, err
			}
		}
	} else {
		blog.Error("update agentsetting, value type(%d) error", update.ValueType)
		return comm.BcsErrCommRequestDataErr, errors.New("value type error")
	}

	return comm.BcsSuccess, nil
}
