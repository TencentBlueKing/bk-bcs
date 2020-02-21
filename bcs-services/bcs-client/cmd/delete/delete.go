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

package delete

import (
	"bk-bcs/bcs-services/bcs-client/cmd/utils"

	"github.com/urfave/cli"
)

//NewDeleteCommand delete sub command
func NewDeleteCommand() cli.Command {
	return cli.Command{
		Name:  "delete",
		Usage: "delete app/process/taskgroup/configmap/service/secret/deployment",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "type, t",
				Usage: "Delete type, app/taskgroup/configmap/service/secret/deployment/crd",
			},
			cli.StringFlag{
				Name:  "name, n",
				Usage: "resource name",
			},
			cli.StringFlag{
				Name:  "namespace, ns",
				Usage: "Namespace",
				Value: "defaultGroup",
			},
			cli.StringFlag{
				Name:  "clusterid",
				Usage: "Cluster ID",
			},
			cli.StringFlag{
				Name:  "enforce",
				Usage: "delete forcibly (1) or not (0)",
				Value: "0",
			},
		},
		Action: func(c *cli.Context) error {
			return deleteF(utils.NewClientContext(c))
		},
	}
}

func deleteF(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionType); err != nil {
		return err
	}

	resourceType := c.String(utils.OptionType)

	switch resourceType {
	case "app", "application":
		return deleteApplication(c)
	case "process":
		return deleteProcess(c)
	case "configmap":
		return deleteConfigMap(c)
	case "secret":
		return deleteSecret(c)
	case "service":
		return deleteService(c)
	case "deploy", "deployment":
		return deleteDeployment(c)
	case "crd", "customresourcedefinition":
		return deleteCustomResourceDefinition(c)
	default:
		//unkown type, try Custom Resource
		return deleteCustomResource(c)
	}
}
