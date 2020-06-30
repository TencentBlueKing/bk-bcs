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
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"sort"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/metric/v1"
)

func listMetric(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID); err != nil {
		return err
	}

	metric := v1.NewBcsMetric(utils.GetClientOption())
	list, err := metric.List(c.ClusterID())
	if err != nil {
		return fmt.Errorf("failed to list metric: %v", err)
	}

	return printListMetric(list)
}

func printListMetric(list v1.MetricList) error {
	fmt.Printf("%-5s %-15s %-20s %-7s %-6s %-5s %-30s %-30s\n",
		"INDEX", "NAME", "NAMESPACE", "VERSION", "DATAID", "PORT", "URI", "SELECTOR")

	sort.Sort(list)
	for i, m := range list {
		var selector []byte
		_ = codec.EncJson(m.Selector, &selector)
		fmt.Printf("%-5d %-15s %-20s %-7s %-6d %-5d %-30s %-30s\n",
			i, m.Name, m.Namespace, m.Version, m.DataID, m.Port, m.URI, selector)
	}
	return nil
}

func listMetricTask(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID); err != nil {
		return err
	}

	metric := v1.NewBcsMetric(utils.GetClientOption())
	list, err := metric.ListTask(c.ClusterID())
	if err != nil {
		return fmt.Errorf("failed to list metric task: %v", err)
	}

	return printListMetricTask(list)
}

func printListMetricTask(list v1.MetricTaskList) error {
	fmt.Printf("%-5s %-15s %-20s %-30s\n",
		"INDEX", "NAME", "NAMESPACE", "SELECTOR")

	sort.Sort(list)
	for i, m := range list {
		var selector []byte
		_ = codec.EncJson(m.Selector, &selector)
		fmt.Printf("%-5d %-15s %-20s %-30s\n",
			i, m.Name, m.Namespace, selector)
	}
	return nil
}
