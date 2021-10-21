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

package v4

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	commonTypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
)

func (bs *bcsScheduler) ListAgentInfo(clusterID string, ipList []string) ([]*commonTypes.BcsClusterAgentInfo, error) {
	return bs.listAgentInfo(clusterID, ipList)
}

func (bs *bcsScheduler) ListAgentSetting(clusterID string, ipList []string) ([]*commonTypes.BcsClusterAgentSetting, error) {
	return bs.listAgentSetting(clusterID, ipList)
}

func (bs *bcsScheduler) UpdateStringAgentSetting(clusterID string, ipList []string, key, value string) error {
	return bs.updateStringAgentSetting(clusterID, ipList, key, value)
}

func (bs *bcsScheduler) UpdateScalarAgentSetting(clusterID string, ipList []string, key string, value float64) error {
	return bs.updateScalarAgentSetting(clusterID, ipList, key, value)
}

func (bs *bcsScheduler) UpdateAgentSetting(clusterID string, data []byte) error {
	return bs.updateAgentSetting(clusterID, data)
}

func (bs *bcsScheduler) SetAgentSetting(clusterID string, data []byte) error {
	return bs.setAgentSetting(clusterID, data)
}

func (bs *bcsScheduler) DeleteAgentSetting(clusterID string, ipList []string) error {
	return bs.deleteAgentSetting(clusterID, ipList)
}

func (bs *bcsScheduler) EnableAgent(clusterID string, ipList []string) error {
	return bs.enableAgent(clusterID, ipList)
}

func (bs *bcsScheduler) DisableAgent(clusterID string, ipList []string) error {
	return bs.disableAgent(clusterID, ipList)
}

func (bs *bcsScheduler) listAgentInfo(clusterID string, ipList []string) ([]*commonTypes.BcsClusterAgentInfo, error) {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerClusterResourceURI, bs.bcsAPIAddress),
		http.MethodGet,
		nil,
		getClusterIDHeader(clusterID),
	)

	if err != nil {
		return nil, err
	}

	code, msg, data, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	if code != 0 {
		return nil, fmt.Errorf("list agent info failed: %s", msg)
	}

	var resource commonTypes.BcsClusterResource
	if err = codec.DecJson(data, &resource); err != nil {
		return nil, err
	}

	result := make([]*commonTypes.BcsClusterAgentInfo, 0)
	for i, item := range resource.Agents {
		item.IP = strings.Split(item.IP, ":")[0]
		if len(ipList) == 0 || inList(item.IP, ipList) {
			result = append(result, &resource.Agents[i])
		}
	}
	return result, nil
}

func (bs *bcsScheduler) listAgentSetting(clusterID string, ipList []string) ([]*commonTypes.BcsClusterAgentSetting, error) {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerAgentSettingURI, bs.bcsAPIAddress, strings.Join(ipList, ",")),
		http.MethodGet,
		nil,
		getClusterIDHeader(clusterID),
	)

	if err != nil {
		return nil, err
	}

	code, msg, data, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	if code != 0 {
		return nil, fmt.Errorf("list agent setting failed: %s", msg)
	}

	var result []*commonTypes.BcsClusterAgentSetting
	err = codec.DecJson(data, &result)
	return result, err
}

func (bs *bcsScheduler) updateStringAgentSetting(clusterID string, ipList []string, key, value string) error {
	var agentSetting commonTypes.BcsClusterAgentSettingUpdate
	agentSetting.IPs = ipList
	agentSetting.SettingName = key
	agentSetting.ValueType = commonTypes.MesosValueType_Text
	agentSetting.ValueText = &commonTypes.MesosValue_Text{Value: value}
	var data []byte
	_ = codec.EncJson(agentSetting, &data)

	return bs.updateAgentSetting(clusterID, data)
}

func (bs *bcsScheduler) updateScalarAgentSetting(clusterID string, ipList []string, key string, value float64) error {
	var agentSetting commonTypes.BcsClusterAgentSettingUpdate
	agentSetting.IPs = ipList
	agentSetting.SettingName = key
	agentSetting.ValueType = commonTypes.MesosValueType_Scalar
	agentSetting.ValueScalar = &commonTypes.MesosValue_Scalar{Value: value}
	var data []byte
	_ = codec.EncJson(agentSetting, &data)

	return bs.updateAgentSetting(clusterID, data)
}

func (bs *bcsScheduler) updateAgentSetting(clusterID string, data []byte) error {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerUpdateAgentSettingURI, bs.bcsAPIAddress),
		http.MethodPost,
		data,
		getClusterIDHeader(clusterID),
	)

	if err != nil {
		return err
	}

	code, msg, _, err := parseResponse(resp)
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("update agent setting failed: %s", msg)
	}

	return nil
}

func (bs *bcsScheduler) setAgentSetting(clusterID string, data []byte) error {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerSetAgentSettingURI, bs.bcsAPIAddress),
		http.MethodPost,
		data,
		getClusterIDHeader(clusterID),
	)

	if err != nil {
		return err
	}

	code, msg, _, err := parseResponse(resp)
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("set agent setting failed: %s", msg)
	}

	return nil
}

func (bs *bcsScheduler) deleteAgentSetting(clusterID string, ipList []string) error {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerAgentSettingURI, bs.bcsAPIAddress, strings.Join(ipList, ",")),
		http.MethodDelete,
		nil,
		getClusterIDHeader(clusterID),
	)

	if err != nil {
		return err
	}

	code, msg, _, err := parseResponse(resp)
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("delete agent setting failed: %s", msg)
	}

	return nil
}

func (bs *bcsScheduler) enableAgent(clusterID string, ipList []string) error {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerEnableAgentURI, bs.bcsAPIAddress, strings.Join(ipList, ",")),
		http.MethodPost,
		nil,
		getClusterIDHeader(clusterID),
	)

	if err != nil {
		return err
	}

	code, msg, _, err := parseResponse(resp)
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("enable agent failed: %s", msg)
	}

	return nil
}

func (bs *bcsScheduler) disableAgent(clusterID string, ipList []string) error {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerDisableAgentURI, bs.bcsAPIAddress, strings.Join(ipList, ",")),
		http.MethodPost,
		nil,
		getClusterIDHeader(clusterID),
	)

	if err != nil {
		return err
	}

	code, msg, _, err := parseResponse(resp)
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("disable agent failed: %s", msg)
	}

	return nil
}
