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

package agent

import (
	"fmt"
	"strconv"
	"strings"

	commonTypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"
)

func listAgentSetting(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID); err != nil {
		return err
	}

	ipList := utils.GetIPList(c.String(utils.OptionIP))
	scheduler := v4.NewBcsScheduler(utils.GetClientOption())

	infoList, err := scheduler.ListAgentInfo(c.ClusterID(), ipList)
	if err != nil {
		return fmt.Errorf("failed to list agent info: %v", err)
	}

	settingList, err := scheduler.ListAgentSetting(c.ClusterID(), ipList)
	if err != nil {
		return fmt.Errorf("failed to list agent setting: %v", err)
	}

	return printListAgentSetting(infoList, settingList)
}

func printListAgentSetting(infoList []*commonTypes.BcsClusterAgentInfo, settingList []*commonTypes.BcsClusterAgentSetting) error {
	stringKeys, scalarKeys := getKeys(infoList, settingList)

	base := "%-5s %-15s %-30s %-8s"
	columns := []interface{}{"INDEX", "IP", "Hostname", "disabled"}
	if len(stringKeys) > 0 {
		base += " %-8s"
		columns = append(columns, "|string:")
		for _, k := range stringKeys {
			base += " %-10s"
			columns = append(columns, k)
		}
	}

	if len(scalarKeys) > 0 {
		base += " %-8s"
		columns = append(columns, "|scalar:")
		for _, k := range scalarKeys {
			base += " %-9s"
			columns = append(columns, k)
		}
	}

	fmt.Printf(base+"\n", columns...)

	for i, agent := range infoList {
		agent.IP = strings.Split(agent.IP, ":")[0]
		var item *commonTypes.BcsClusterAgentSetting
		for _, attr := range settingList {
			if attr.InnerIP == agent.IP {
				item = attr
				break
			}
		}

		extra := []interface{}{strconv.Itoa(i), agent.IP, agent.HostName}
		if item != nil {
			extra = append(extra, strconv.FormatBool(item.Disabled))
		} else {
			extra = append(extra, "false")
		}

		if len(stringKeys) > 0 {
			extra = append(extra, "|")
		}
		for _, k := range stringKeys {
			if item == nil {
				extra = append(extra, "")
				continue
			}

			if v, ok := item.AttrStrings[k]; ok {
				extra = append(extra, v.Value)
			} else {
				extra = append(extra, "")
			}
		}
		if len(scalarKeys) > 0 {
			extra = append(extra, "|")
		}
		for _, k := range scalarKeys {
			if item == nil {
				extra = append(extra, "")
				continue
			}

			if v, ok := item.AttrScalars[k]; ok {
				extra = append(extra, fmt.Sprintf("%f", v.Value))
			} else {
				extra = append(extra, "")
			}
		}

		fmt.Printf(base+"\n", extra...)
	}
	return nil
}

func getKeys(infoList []*commonTypes.BcsClusterAgentInfo, settingList []*commonTypes.BcsClusterAgentSetting) (stringKeys, scalarKeys []string) {
	// get the columns
	stringKeys = make([]string, 0)
	scalarKeys = make([]string, 0)
	stringKeysMap := make(map[string]bool)
	scalarKeysMap := make(map[string]bool)

	for _, agent := range infoList {
		agent.IP = strings.Split(agent.IP, ":")[0]
		var attr *commonTypes.BcsClusterAgentSetting
		for _, a := range settingList {
			if a.InnerIP == agent.IP {
				attr = a
				break
			}
		}
		if attr == nil {
			continue
		}

		for k := range attr.AttrStrings {
			if _, ok := stringKeysMap[k]; ok {
				continue
			}
			stringKeysMap[k] = true
			stringKeys = append(stringKeys, k)
		}
		for k := range attr.AttrScalars {
			if _, ok := scalarKeysMap[k]; ok {
				continue
			}
			scalarKeysMap[k] = true
			scalarKeys = append(scalarKeys, k)
		}
	}
	return stringKeys, scalarKeys
}
