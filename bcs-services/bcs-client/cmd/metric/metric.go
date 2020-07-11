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

package metric

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"

	"github.com/urfave/cli"
)

func NewMetricCommand() cli.Command {
	return cli.Command{
		Name:  "metric",
		Usage: "manage bcs metric",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "from-file, f",
				Usage: "upsert with configuration `FILE`",
			},
			cli.StringFlag{
				Name:  "type, t",
				Usage: "metric type, metric/task",
				Value: "metric",
			},
			cli.BoolFlag{
				Name:  "list, l",
				Usage: "For get operation",
			},
			cli.BoolFlag{
				Name:  "inspect, i",
				Usage: "For inspect operation",
			},
			cli.BoolFlag{
				Name:  "upsert, u",
				Usage: "For insert or update operation",
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
				Name:  "namespace, ns",
				Usage: "Namespace",
			},
			cli.StringFlag{
				Name:  "name, n",
				Usage: "Name",
			},
			cli.StringFlag{
				Name:  "clustertype, ct",
				Usage: "cluster type, mesos/k8s",
				Value: "mesos",
			},
		},
		Action: func(c *cli.Context) error {
			if err := metric(utils.NewClientContext(c)); err != nil {
				return err
			}
			return nil
		},
	}
}

func metric(c *utils.ClientContext) error {
	// get basic command
	isList := c.Bool(utils.OptionList)
	isUpsert := c.Bool(utils.OptionUpsert)
	isInspect := c.Bool(utils.OptionInspect)
	isDelete := c.Bool(utils.OptionDelete)

	metricType := c.String(utils.OptionType)
	switch metricType {
	case "metric":
		if isList {
			return listMetric(c)
		}

		if isUpsert {
			return upsertMetric(c)
		}

		if isInspect {
			return inspectMetric(c)
		}

		if isDelete {
			return deleteMetric(c)
		}
	case "task":
		if isList {
			return listMetricTask(c)
		}

		if isUpsert {
			return upsertMetricTask(c)
		}

		if isInspect {
			return inspectMetricTask(c)
		}

		if isDelete {
			return deleteMetricTask(c)
		}
	default:
		return fmt.Errorf("unknown type: %s", metricType)
	}
	return nil
}
