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

package application

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"

	"github.com/urfave/cli"
)

func NewScaleCommand() cli.Command {
	return cli.Command{
		Name:  "scale",
		Usage: "Scale down or scale up applications",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "name, n",
				Usage: "Application/Process name",
			},
			cli.StringFlag{
				Name:  "type, t",
				Usage: "scale type, app/process",
				Value: "app",
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
			cli.IntFlag{
				Name:  "instance",
				Usage: "Instances to be scaled",
				Value: 1,
			},
		},
		Action: func(c *cli.Context) error {
			if err := scale(utils.NewClientContext(c)); err != nil {
				return err
			}
			return nil
		},
	}
}

func scale(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionType); err != nil {
		return err
	}

	resourceType := c.String(utils.OptionType)

	switch resourceType {
	case "app", "application":
		return scaleApplication(c)
	case "process":
		return scaleProcess(c)
	default:
		return fmt.Errorf("invalid type: %s", resourceType)
	}
}

func scaleApplication(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionNamespace, utils.OptionName, utils.OptionInstance); err != nil {
		return err
	}

	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	err := scheduler.ScaleApplication(c.ClusterID(), c.Namespace(), c.String(utils.OptionName), c.Int(utils.OptionInstance))
	if err != nil {
		return fmt.Errorf("failed to scale application: %v", err)
	}

	fmt.Printf("success to scale application\n")
	return nil
}

func scaleProcess(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionNamespace, utils.OptionName, utils.OptionInstance); err != nil {
		return err
	}

	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	err := scheduler.ScaleProcess(c.ClusterID(), c.Namespace(), c.String(utils.OptionName), c.Int(utils.OptionInstance))
	if err != nil {
		return fmt.Errorf("failed to scale process: %v", err)
	}

	fmt.Printf("success to scale process\n")
	return nil
}
