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
	"encoding/json"

	comm "github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/util"
)

// DisableAgent setting agent unschedulable
func (b *backend) DisableAgent(IP string) error {
	util.Lock.Lock(commtypes.BcsClusterAgentSetting{}, IP)
	defer util.Lock.UnLock(commtypes.BcsClusterAgentSetting{}, IP)

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
	util.Lock.Lock(commtypes.BcsClusterAgentSetting{}, IP)
	defer util.Lock.UnLock(commtypes.BcsClusterAgentSetting{}, IP)

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

func (b *backend) TaintAgents(agents []*commtypes.BcsClusterAgentSetting) error {
	for _, o := range agents {
		err := b.taintAgent(o)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *backend) taintAgent(o *commtypes.BcsClusterAgentSetting) error {
	util.Lock.Lock(commtypes.BcsClusterAgentSetting{}, o.InnerIP)
	defer util.Lock.UnLock(commtypes.BcsClusterAgentSetting{}, o.InnerIP)

	agent, err := b.store.FetchAgentSetting(o.InnerIP)
	if err != nil {
		blog.Error("fetch agent setting(%s) from db fail:%s", o.InnerIP, err.Error())
		return err
	}
	by, _ := json.Marshal(o.NoSchedule)
	blog.Infof("taints agent(%s) %s", o.InnerIP, string(by))

	if agent != nil {
		agent.NoSchedule = o.NoSchedule
	} else {
		agent = &commtypes.BcsClusterAgentSetting{
			InnerIP:    o.InnerIP,
			NoSchedule: o.NoSchedule,
		}
	}

	err = b.store.SaveAgentSetting(agent)
	if err != nil {
		blog.Error("save agent(%s) in db fail:%s", o.InnerIP, err.Error())
		return err
	}

	return nil
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
		if err != nil {
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
/*func (b *backend) DeleteAgentSettingList(IPs []string) (int, error) {

	for _, IP := range IPs {

		blog.Infof("delete agent setting(%s) from zk", IP)

		err := b.store.DeleteAgentSetting(IP)
		if err != nil {
			blog.Error("delete agent setting(%s) from zk err: %s", IP, err.Error())
			return comm.BcsErrCommDeleteZkNodeFail, err
		}
	}

	return comm.BcsSuccess, nil
}*/

//SetAgentSettingList setting agent by detail info
func (b *backend) SetAgentSettingList(agents []*commtypes.BcsClusterAgentSetting) (int, error) {

	for _, agent := range agents {
		err := b.setAgentSetting(agent)
		if err != nil {
			return comm.BcsErrCommGetZkNodeFail, err
		}
	}

	return comm.BcsSuccess, nil
}

func (b *backend) setAgentSetting(agent *commtypes.BcsClusterAgentSetting) error {
	util.Lock.Lock(commtypes.BcsClusterAgentSetting{}, agent.InnerIP)
	defer util.Lock.UnLock(commtypes.BcsClusterAgentSetting{}, agent.InnerIP)

	o, err := b.store.FetchAgentSetting(agent.InnerIP)
	if err != nil {
		blog.Error("fetch agent setting(%s) from db fail:%s", agent.InnerIP, err.Error())
		return err
	}
	by, _ := json.Marshal(agent)
	blog.Infof("set agent(%s) setting(%s)", agent.InnerIP, string(by))
	if o == nil {
		o = agent
	} else {
		o.AttrScalars = agent.AttrScalars
		o.AttrStrings = agent.AttrStrings
	}
	err = b.store.SaveAgentSetting(o)
	if err != nil {
		blog.Error("save agent setting(%s) [%+v] to zk err:%s", o.InnerIP, o, err.Error())
		return err
	}

	return nil
}

func (b *backend) DisableAgentList(IPs []string) (int, error) {
	for _, IP := range IPs {
		err := b.DisableAgent(IP)
		if err != nil {
			return comm.BcsErrCommGetZkNodeFail, err
		}
	}

	return comm.BcsSuccess, nil
}

//EnableAgentList enable agent schedulable by IP address list
func (b *backend) EnableAgentList(IPs []string) (int, error) {

	for _, IP := range IPs {
		err := b.EnableAgent(IP)
		if err != nil {
			return comm.BcsErrCommGetZkNodeFail, err
		}
	}

	return comm.BcsSuccess, nil
}

func (b *backend) UpdateExtendedResources(ex *commtypes.ExtendedResource) error {
	util.Lock.Lock(commtypes.BcsClusterAgentSetting{}, ex.InnerIP)
	defer util.Lock.UnLock(commtypes.BcsClusterAgentSetting{}, ex.InnerIP)

	agent, err := b.store.FetchAgentSetting(ex.InnerIP)
	if err != nil {
		blog.Error("fetch agent setting(%s) from db fail:%s", ex.InnerIP, err.Error())
		return err
	}
	blog.Infof("update agent(%s) ExtendedResource %s", ex.InnerIP, ex.Name)

	if agent == nil {
		agent = &commtypes.BcsClusterAgentSetting{
			InnerIP:           ex.InnerIP,
			ExtendedResources: make(map[string]*commtypes.ExtendedResource),
		}
	}
	if agent.ExtendedResources == nil {
		agent.ExtendedResources = make(map[string]*commtypes.ExtendedResource)
	}
	agent.ExtendedResources[ex.Name] = ex
	return b.store.SaveAgentSetting(agent)
}

//UpdateAgentSettingList update agent setting by details
/*func (b *backend) UpdateAgentSettingList(update *commtypes.BcsClusterAgentSettingUpdate) (int, error) {

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
			if err != nil {
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
			if err != nil {
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
}*/
