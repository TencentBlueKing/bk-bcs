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

package available

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"

	"github.com/urfave/cli"
)

// NewDisableCommand xxx
func NewDisableCommand() cli.Command {
	return cli.Command{
		Name:  "disable",
		Usage: "disable agent by ip",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "type, t",
				Usage: "Disable type, agent",
			},
			cli.StringFlag{
				Name:  "clusterid",
				Usage: "Cluster ID",
			},
			cli.StringFlag{
				Name:  "ip",
				Usage: "The ip of agent to disabled. Split by ,",
			},
		},
		Action: func(c *cli.Context) error {
			if err := disable(utils.NewClientContext(c)); err != nil {
				return err
			}
			return nil
		},
	}
}

func disable(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionType); err != nil {
		return err
	}

	resourceType := c.String(utils.OptionType)

	switch resourceType {
	case "agent":
		return disableAgent(c)
	default:
		return fmt.Errorf("invalid type: %s", resourceType)
	}
}

func disableAgent(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionIP); err != nil {
		return err
	}

	ipList := utils.GetIPList(c.String(utils.OptionIP))

	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	err := scheduler.DisableAgent(c.ClusterID(), ipList)
	if err != nil {
		return fmt.Errorf("failed to disable agent: %v", err)
	}

	fmt.Printf("success to disable agent\n")
	return nil
}
