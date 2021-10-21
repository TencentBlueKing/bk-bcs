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

package inspect

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"

	"github.com/urfave/cli"
)

//NewInspectCommand inspect sub command, get specified resource details
func NewInspectCommand() cli.Command {
	return cli.Command{
		Name:  "inspect",
		Usage: "show detailed information of application, taskgroup, service, configmap, deployment, secret and etc.",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "type, t",
				Usage: "Inspect type, application(app)/process/taskgroup/service/configmap/secret/deployment(deploy)/endpoint/customresourcedefinition(crd)/meshcluster",
			},
			cli.StringFlag{
				Name:  "clusterid",
				Usage: "Cluster ID",
			},
			cli.StringFlag{
				Name:  "namespace, ns",
				Usage: "Namespace",
				Value: "",
			},
			cli.StringFlag{
				Name:  "name, n",
				Usage: "Inspect name according to type",
			},
		},
		Action: func(c *cli.Context) error {
			return inspect(utils.NewClientContext(c))
		},
	}
}

func inspect(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionType); err != nil {
		return err
	}

	resourceType := c.String(utils.OptionType)

	switch resourceType {
	case "app", "application":
		return inspectApplication(c)
	case "process":
		return inspectProcess(c)
	case "tg", "taskgroup":
		return inspectTaskGroup(c)
	case "configmap":
		return inspectConfigMap(c)
	case "secret":
		return inspectSecret(c)
	case "service":
		return inspectService(c)
	case "deploy", "deployment":
		return inspectDeployment(c)
	case "endpoint":
		return inspectEndpoint(c)
	case "crd", "customresourcedefinition":
		return inspectCustomResourceDefinition(c)
	case "meshcluster":
		return inspectMeshCluster(c)
	default:
		//unkown type, try Custom Resource
		return inspectCustomResource(c)
	}
}

func printInspect(single interface{}) error {
	fmt.Printf("%s\n", utils.TryIndent(single))
	return nil
}
