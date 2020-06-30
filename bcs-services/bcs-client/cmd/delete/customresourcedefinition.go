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

package delete

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	v4 "github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"
)

func deleteCustomResourceDefinition(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionName); err != nil {
		return err
	}
	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	if err := scheduler.DeleteCustomResourceDefinition(c.ClusterID(), c.String(utils.OptionName)); err != nil {
		return fmt.Errorf("failed to delete customresourcedefinition: %v", err)
	}
	fmt.Printf("success to delete customresourcedefinition %s\n", c.String(utils.OptionName))
	return nil
}

func deleteCustomResource(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionName, utils.OptionNamespace); err != nil {
		return err
	}
	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	apiVersion, plural, err := utils.GetCustomResourceType(scheduler, c.ClusterID(), c.String(utils.OptionType))
	if err != nil {
		return err
	}
	if err := scheduler.DeleteCustomResource(c.ClusterID(), apiVersion, plural, c.Namespace(), c.String(utils.OptionName)); err != nil {
		return fmt.Errorf("failed to delete custom resource: %v", err)
	}
	fmt.Printf("success to delete %s: %s/%s\n", plural, c.Namespace(), c.String(utils.OptionName))
	return nil
}
