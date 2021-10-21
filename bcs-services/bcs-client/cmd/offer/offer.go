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

package offer

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	v4 "github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"

	"github.com/urfave/cli"
)

func NewOfferCommand() cli.Command {
	return cli.Command{
		Name:  "offer",
		Usage: "list offers of clusters",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "clusterid",
				Usage: "Cluster ID",
			},
			cli.StringFlag{
				Name:  "ip",
				Usage: "IP of slaves",
			},
			cli.BoolFlag{
				Name:  "all, a",
				Usage: "get all agent raw offer data",
			},
		},
		Action: func(c *cli.Context) error {
			if err := offer(utils.NewClientContext(c)); err != nil {
				return err
			}
			return nil
		},
	}
}

func offer(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID); err != nil {
		return err
	}

	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	list, err := scheduler.GetOffer(c.ClusterID())
	if err != nil {
		return fmt.Errorf("failed to get offer: %v", err)
	}

	if c.IsSet(utils.OptionAll) {
		return printAllOffer(list)
	}

	if c.IsSet(utils.OptionIP) {
		return printOneOffer(list, c.String(utils.OptionIP))
	}

	return printListOffer(list)
}

func printAllOffer(list []*types.OfferWithDelta) error {
	fmt.Printf("%s\n", utils.TryIndent(list))
	return nil
}

func printOneOffer(list []*types.OfferWithDelta, ip string) error {
	var data *types.OfferWithDelta
	found := false
	for _, item := range list {
		for _, attr := range item.Attributes {
			if *attr.Name != "InnerIP" {
				break
			}
			if *attr.Text.Value == ip {
				found = true
				data = item
			}
		}
		if found {
			break
		}
	}

	if data == nil {
		return fmt.Errorf("offer does no exist")
	}

	fmt.Printf("%s\n", utils.TryIndent(data))
	return nil
}

func printListOffer(list []*types.OfferWithDelta) error {
	fmt.Printf("%-5s  %-17s  %-20s  %-4s  %-8s  %-10s %-10s %-12s %-30s\n",
		"INDEX",
		"IP",
		"Hostname",
		"CPUS",
		"MEM",
		"DISK",
		"CITY",
		"IPRESOURCES",
		"PORTS")

	for i, item := range list {
		var ip, city, ports string
		var ipResource, cpus, mem, disk float64

		var ipS, cityS, ipResourceS bool

		for _, attr := range item.Attributes {
			switch *attr.Name {
			case "InnerIP":
				if ipS {
					continue
				}
				ipS = true
				ip = *attr.Text.Value
			case "City":
				if cityS {
					continue
				}
				cityS = true
				city = *attr.Text.Value
			case "ip-resources":
				if ipResourceS {
					continue
				}
				ipResourceS = true
				ipResource = *attr.Scalar.Value
			}
		}

		for _, resource := range item.Resources {
			switch *resource.Name {
			case "cpus":
				cpus = *resource.Scalar.Value
			case "mem":
				mem = *resource.Scalar.Value
			case "disk":
				disk = *resource.Scalar.Value
			case "ports":
				for _, p := range resource.Ranges.Range {
					ports += fmt.Sprintf("%d-%d ", *p.Begin, *p.End)
				}
			}
		}

		if item.DeltaResource != nil {
			cpus = cpus - item.DeltaResource.Cpus
			mem = mem - item.DeltaResource.Mem
			disk = disk - item.DeltaResource.Disk
		}

		fmt.Printf("%-5d  %-17s  %-20s  %-4.2f  %-8.2f  %-10.2f %-10s %-12.0f %-30s\n",
			i+1,
			ip,
			*item.Hostname,
			cpus,
			mem,
			disk,
			city,
			ipResource,
			ports)
	}
	return nil
}
