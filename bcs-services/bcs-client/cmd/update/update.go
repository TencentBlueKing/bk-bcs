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

package update

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"

	"github.com/urfave/cli"
)

//NewUpdateCommand update sub command
func NewUpdateCommand() cli.Command {
	return cli.Command{
		Name:  "update",
		Usage: "update application(app)/service/secret/configmap/deployment(deploy)/customresourcedefinition(crd)",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "from-file, f",
				Usage: "read request from file, like myrequest.json",
			},
			cli.StringFlag{
				Name:  "type, t",
				Usage: "update type, application(app)/service/secret/configmap/deployment(deploy)/customresourcedefinition(crd)",
			},
			cli.StringFlag{
				Name:  "clusterid",
				Usage: "Cluster ID",
			},
			cli.IntFlag{
				Name:  "instance",
				Usage: "Instances to update",
				Value: 1,
			},
			cli.BoolFlag{
				Name:  utils.OptionOnlyUpdateResource,
				Usage: "whether only update resource for application",
			},
		},
		Action: func(c *cli.Context) error {
			if err := update(utils.NewClientContext(c)); err != nil {
				return err
			}
			return nil
		},
	}
}

func update(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionType); err != nil {
		return err
	}

	resourceType := c.String(utils.OptionType)

	switch resourceType {
	case "app", "application":
		return updateApplication(c)
	case "process":
		return updateProcess(c)
	case "configmap":
		return updateConfigMap(c)
	case "secret":
		return updateSecret(c)
	case "service":
		return updateService(c)
	case "deploy", "deployment":
		return updateDeployment(c)
	case "crd", "customresourcedefinition":
		return updateCustomResourceDefinition(c)
	default:
		//unkown type, try CustomResource
		return updateCustomResource(c)
	}
}
