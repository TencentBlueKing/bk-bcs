/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package transaction

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/urfave/cli"
)

func NewTransactionCommand() cli.Command {
	return cli.Command{
		Name:  "trans",
		Usage: "list and operate scheduler internal data --- transaction",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "list, l",
				Usage: "For get operation",
			},
			cli.BoolFlag{
				Name:  "delete, d",
				Usage: "For delete operation, it will delete all agentsettings of specific ips",
			},
			cli.StringFlag{
				Name:  "clusterid",
				Usage: "Cluster ID",
			},
			cli.StringFlag{
				Name:  "objkind",
				Usage: "For list operation, the kind of object related to the transaction",
			},
			cli.StringFlag{
				Name:  "objname",
				Usage: "For list operation, the name of object related to the transaction",
			},
			cli.StringFlag{
				Name:  "name, n",
				Usage: "For delete operation, the name of transaction",
			},
			cli.StringFlag{
				Name:  "namespace, ns",
				Usage: "For delete operation, the namespace of transaction",
			},
		},
		Action: func(c *cli.Context) error {
			if err := transaction(utils.NewClientContext(c)); err != nil {
				return err
			}
			return nil
		},
	}
}

func transaction(c *utils.ClientContext) error {
	// get basic command
	isList := c.Bool(utils.OptionList)
	isDelete := c.Bool(utils.OptionDelete)

	if isList {
		return listTransaction(c)
	}
	if isDelete {
		return deleteTransaction(c)
	}
	return nil
}
