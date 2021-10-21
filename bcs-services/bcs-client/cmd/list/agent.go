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

package list

import (
	"fmt"
	"net/url"
	"time"

	commonTypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"
)

func listAgent(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionIP); err != nil {
		return err
	}

	condition := url.Values{}
	condition.Add(filterNamespaceTag, c.Namespace())

	ipList := utils.GetIPList(c.String(utils.OptionIP))

	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	list, err := scheduler.ListAgentInfo(c.ClusterID(), ipList)
	if err != nil {
		return fmt.Errorf("failed to list agent: %v", err)
	}

	return printListAgent(list)
}

func printListAgent(list []*commonTypes.BcsClusterAgentInfo) error {
	if len(list) == 0 {
		fmt.Printf("Found no agent\n")
		return nil
	}

	fmt.Printf("%-5s %-20s %-30s %-40s %-40s\n",
		"INDEX",
		"IP",
		"Hostname",
		"Registered-Time",
		"Re-Registered-Time")

	for i, agent := range list {
		fmt.Printf("%-5d %-20s %-30s %-40s %-40s\n",
			i,
			agent.IP,
			agent.HostName,
			time.Unix(0, agent.RegisteredTime).String(),
			time.Unix(0, agent.ReRegisteredTime).String())
	}
	return nil
}
