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

func NewRollBackCommand() cli.Command {
	return cli.Command{
		Name:  "rollback",
		Usage: "rollback application to last version",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "from-file, f",
				Usage: "Rollback with configuration `FILE`",
			},
			cli.StringFlag{
				Name:  "clusterid",
				Usage: "Cluster ID",
			},
			cli.StringFlag{
				Name:  "type, t",
				Usage: "Rollback type, app/process",
			},
		},
		Action: func(c *cli.Context) error {
			if err := rollBack(utils.NewClientContext(c)); err != nil {
				return err
			}
			return nil
		},
	}
}

func rollBack(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionType); err != nil {
		return err
	}

	resourceType := c.String(utils.OptionType)

	switch resourceType {
	case "app", "application":
		return rollBackApplication(c)
	case "process":
		return rollBackProcess(c)
	default:
		return fmt.Errorf("invalid type: %s", resourceType)
	}
}

func rollBackApplication(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID); err != nil {
		return err
	}

	data, err := c.FileData()
	if err != nil {
		return err
	}

	namespace, err := utils.ParseNamespaceFromJSON(data)
	if err != nil {
		return err
	}

	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	err = scheduler.RollBackApplication(c.ClusterID(), namespace, data)
	if err != nil {
		return fmt.Errorf("failed to roll back application: %v", err)
	}

	fmt.Printf("success to roll back application\n")
	return nil
}

func rollBackProcess(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID); err != nil {
		return err
	}

	data, err := c.FileData()
	if err != nil {
		return err
	}

	namespace, err := utils.ParseNamespaceFromJSON(data)
	if err != nil {
		return err
	}

	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	err = scheduler.RollBackProcess(c.ClusterID(), namespace, data)
	if err != nil {
		return fmt.Errorf("failed to roll back process: %v", err)
	}

	fmt.Printf("success to roll back process\n")
	return nil
}
