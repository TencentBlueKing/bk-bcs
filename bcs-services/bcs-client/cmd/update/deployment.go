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

package update

import (
	"fmt"
	"net/url"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"
)

func updateDeployment(c *utils.ClientContext) error {
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

	updateResFlag := c.Bool(utils.OptionOnlyUpdateResource)
	extraValue := url.Values{}

	if updateResFlag {
		extraValue.Add("args", "resource")
	}

	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	err = scheduler.UpdateDeployment(c.ClusterID(), namespace, data, extraValue)
	if err != nil {
		return fmt.Errorf("failed to update deployment: %v", err)
	}

	fmt.Printf("success to update deployment\n")
	return nil
}
