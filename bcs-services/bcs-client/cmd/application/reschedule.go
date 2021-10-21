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
	"net/url"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/storage/v1"

	"github.com/urfave/cli"
)

func NewRescheduleCommand() cli.Command {
	return cli.Command{
		Name:  "reschedule",
		Usage: "reschedule taskgroup",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "type, t",
				Usage: "reschedule type, taskgroup",
			},
			cli.StringFlag{
				Name:  "name, n",
				Usage: "Application name",
				Value: "",
			},
			cli.StringFlag{
				Name:  "namespace, ns",
				Usage: "Namespace",
				Value: "",
			},
			cli.StringFlag{
				Name:  "tgname",
				Usage: "Taskgroup name",
				Value: "",
			},
			cli.StringFlag{
				Name:  "clusterid",
				Usage: "Cluster ID",
			},
			cli.StringFlag{
				Name:  "ip",
				Usage: "The ip of taskgroup. Split by ,",
			},
		},
		Action: func(c *cli.Context) error {
			if err := reschedule(utils.NewClientContext(c)); err != nil {
				return err
			}
			return nil
		},
	}
}

func reschedule(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionType); err != nil {
		return err
	}

	resourceType := c.String(utils.OptionType)

	switch resourceType {
	case "tg", "taskgroup":
		if c.IsSet(utils.OptionIP) {
			return rescheduleTaskGroupByIP(c)
		}
		return rescheduleTaskGroup(c)
	default:
		return fmt.Errorf("invalid type: %s", resourceType)
	}
}

func rescheduleTaskGroupByIP(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionName); err != nil {
		return err
	}

	condition := url.Values{}
	condition.Add("hostIp", c.String(utils.OptionIP))
	storage := v1.NewBcsStorage(utils.GetClientOption())

	list, err := storage.ListTaskGroup(c.ClusterID(), condition)
	if err != nil {
		return fmt.Errorf("failed to list taskgroup: %v", err)
	}

	successMsg := ""
	failureMsg := ""
	successCnt := 0
	failureCnt := 0

	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	for _, single := range list {
		if err := scheduler.RescheduleTaskGroup(c.ClusterID(), single.Data.NameSpace, single.Data.RcName, single.Data.Name); err != nil {
			successMsg += fmt.Sprintf("success to reschedule taskgroup %s\n", single.Data.Name)
			successCnt++
		} else {
			failureMsg += fmt.Sprintf("failed to reschedule taskgroup %s: %v\n", single.Data.Name, err)
			failureCnt++
		}
	}

	fmt.Printf("total success taskgroups: %d\n", successCnt)
	fmt.Printf(successMsg)
	fmt.Println()
	fmt.Printf("total failure taskgroups: %d\n", failureCnt)
	fmt.Printf(failureMsg)
	return nil
}

func rescheduleTaskGroup(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionNamespace, utils.OptionName, utils.OptionTaskGroupName); err != nil {
		return err
	}

	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	err := scheduler.RescheduleTaskGroup(c.ClusterID(), c.Namespace(), c.String(utils.OptionName), c.String(utils.OptionTaskGroupName))
	if err != nil {
		return fmt.Errorf("failed to reschedule taskgroup: %v", err)
	}

	fmt.Printf("success to reschedule taskgroup\n")
	return nil
}
