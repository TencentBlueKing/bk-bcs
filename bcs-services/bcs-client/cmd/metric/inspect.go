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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/metric/v1"
)

func inspectMetric(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionNamespace, utils.OptionName); err != nil {
		return err
	}

	metric := v1.NewBcsMetric(utils.GetClientOption())
	single, err := metric.Inspect(c.ClusterID(), c.Namespace(), c.String(utils.OptionName))
	if err != nil {
		return fmt.Errorf("failed to inspect metric: %v", err)
	}

	return printInspect(single)
}

func inspectMetricTask(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionNamespace, utils.OptionName); err != nil {
		return err
	}

	metric := v1.NewBcsMetric(utils.GetClientOption())
	single, err := metric.InspectTask(c.ClusterID(), c.Namespace(), c.String(utils.OptionName))
	if err != nil {
		return fmt.Errorf("failed to inspect metric task: %v", err)
	}

	return printInspect(single)
}

func printInspect(single interface{}) error {
	fmt.Printf("%s\n", utils.TryIndent(single))
	return nil
}
