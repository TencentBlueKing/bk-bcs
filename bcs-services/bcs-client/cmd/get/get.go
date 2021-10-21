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

package get

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/storage/v1"
	userV1 "github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/usermanager/v1"
	"net/http"

	"github.com/urfave/cli"
)

func NewGetCommand() cli.Command {
	return cli.Command{
		Name:  "get",
		Usage: "get the original definition of application/process/deployment/ippoolstatic/ippoolstatic-detail/user",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "type, t",
				Usage: "Get type, application(app)/process/deployment(deploy)/ippoolstatic(ipps)/ippoolstatic-detail(ippsd)/user",
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
				Usage: "Name",
			},
			cli.StringFlag{
				Name:  "usertype",
				Usage: "user type, value can be admin/saas/plain",
			},
			cli.StringFlag{
				Name:  "username",
				Usage: "user name",
			},
			cli.StringFlag{
				Name:  "resourcetype",
				Usage: "resource type, value can be cluster/storage/network-detection...",
			},
		},
		Action: func(c *cli.Context) error {
			if err := get(utils.NewClientContext(c)); err != nil {
				return err
			}
			return nil
		},
	}
}

func get(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionType); err != nil {
		return err
	}

	resourceType := c.String(utils.OptionType)

	switch resourceType {
	case "app", "application":
		return getApplication(c)
	case "process":
		return getProcess(c)
	case "deploy", "deployment":
		return getDeployment(c)
	case "ipps", "ippoolstatic":
		return getIPPoolStatic(c)
	case "ippsd", "ippoolstatic-detail":
		return getIPPoolStaticDetail(c)
	case "user":
		return getUser(c)
	default:
		return fmt.Errorf("invalid type: %s", resourceType)
	}
}

func getApplication(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionNamespace, utils.OptionName); err != nil {
		return err
	}

	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	result, err := scheduler.GetApplicationDefinition(c.ClusterID(), c.Namespace(), c.String(utils.OptionName))
	if err != nil {
		return fmt.Errorf("failed to get application definition: %v", err)
	}

	return printGet(result)
}

func getProcess(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionNamespace, utils.OptionName); err != nil {
		return err
	}

	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	result, err := scheduler.GetProcessDefinition(c.ClusterID(), c.Namespace(), c.String(utils.OptionName))
	if err != nil {
		return fmt.Errorf("failed to get process definition: %v", err)
	}

	return printGet(result)
}

func getDeployment(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionNamespace, utils.OptionName); err != nil {
		return err
	}

	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	result, err := scheduler.GetDeploymentDefinition(c.ClusterID(), c.Namespace(), c.String(utils.OptionName))
	if err != nil {
		return fmt.Errorf("failed to get deployment definition: %v", err)
	}

	return printGet(result)
}

func getIPPoolStatic(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID); err != nil {
		return err
	}

	storage := v1.NewBcsStorage(utils.GetClientOption())

	result, err := storage.ListIPPoolStatic(c.ClusterID(), nil)
	if err != nil {
		return fmt.Errorf("failed to get ippoolstatic: %v", err)
	}

	if len(result) == 0 {
		fmt.Println("Resource Not Found.")
		return nil
	}
	return printGet(result[0].Data)
}

func getIPPoolStaticDetail(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID); err != nil {
		return err
	}

	storage := v1.NewBcsStorage(utils.GetClientOption())

	result, err := storage.ListIPPoolStaticDetail(c.ClusterID(), nil)
	if err != nil {
		return fmt.Errorf("failed to get ippoolstatic-detail : %v", err)
	}

	if len(result) == 0 || len(result[0].Data) == 0 {
		fmt.Println("Resource Not Found.")
		return nil
	}
	return printGet(result[0].Data)
}

func getUser(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionUserName, utils.OptionUserType); err != nil {
		return err
	}

	userManager := userV1.NewBcsUserManager(utils.GetClientOption())
	user, err := userManager.CreateOrGetUser(c.String(utils.OptionUserType), c.String(utils.OptionUserName), http.MethodGet)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return printGet(user)
}

func printGet(single interface{}) error {
	fmt.Printf("%s\n", utils.TryIndent(single))
	return nil
}
