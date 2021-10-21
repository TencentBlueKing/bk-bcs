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

package agent

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"
)

func updateAgentSetting(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionKey, utils.OptionIP); err != nil {
		return err
	}

	scheduler := v4.NewBcsScheduler(utils.GetClientOption())

	ipList := utils.GetIPList(c.String(utils.OptionIP))
	key := c.String(utils.OptionKey)

	var err error
	if c.IsSet(utils.OptionString) {
		err = scheduler.UpdateStringAgentSetting(c.ClusterID(), ipList, key, c.String(utils.OptionString))
	} else if c.IsSet(utils.OptionScalar) {
		err = scheduler.UpdateScalarAgentSetting(c.ClusterID(), ipList, key, c.Float64(utils.OptionScalar))
	} else {
		return fmt.Errorf("string or scalar must be specified while updating")
	}

	if err != nil {
		return fmt.Errorf("failed to update agent setting: %v", err)
	}

	fmt.Printf("success to update agent setting\n")
	return nil
}
