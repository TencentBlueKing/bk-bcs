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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/urfave/cli"
)

func NewAgentSettingCommand() cli.Command {
	return cli.Command{
		Name:  "as",
		Usage: "manage the agentsettings of nodes",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "list, l",
				Usage: "For get operation",
			},
			cli.BoolFlag{
				Name:  "update, u",
				Usage: "For update operation",
			},
			cli.BoolFlag{
				Name:  "set, s",
				Usage: "For set operation",
			},
			cli.BoolFlag{
				Name:  "delete, d",
				Usage: "For delete operation, it will delete all agentsettings of specific ips",
			},
			cli.StringFlag{
				Name:  "key, k",
				Usage: "attribute key",
			},
			cli.StringFlag{
				Name:  "string",
				Usage: "attribute string value",
			},
			cli.StringFlag{
				Name:  "from-file, f",
				Usage: "set attribute file",
			},
			cli.Float64Flag{
				Name:  "scalar",
				Usage: "attribute float value",
			},
			cli.StringFlag{
				Name:  "clusterid",
				Usage: "Cluster ID",
			},
			cli.StringFlag{
				Name:  "ip",
				Usage: "The ip of slaves. In list/update it support multi ips, split by comma",
			},
			cli.StringFlag{
				Name:  "labelSelector",
				Usage: "string selector selector of slaves. example: key1=value1,key2=value2",
			},
		},
		Action: func(c *cli.Context) error {
			if err := agentSetting(utils.NewClientContext(c)); err != nil {
				return err
			}
			return nil
		},
	}
}

func agentSetting(c *utils.ClientContext) error {
	// get basic command
	isList := c.Bool(utils.OptionList)
	isSet := c.Bool(utils.OptionSet)
	isUpdate := c.Bool(utils.OptionUpdate)
	isDelete := c.Bool(utils.OptionDelete)

	if isList {
		return listAgentSetting(c)
	}

	if isSet {
		return setAgentSetting(c)
	}

	if isUpdate {
		return updateAgentSetting(c)
	}

	if isDelete {
		return deleteAgentSetting(c)
	}

	return nil
}
