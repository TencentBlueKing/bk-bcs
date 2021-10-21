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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	v4 "github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"
)

func updateCustomResourceDefinition(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID); err != nil {
		return err
	}

	data, err := c.FileData()
	if err != nil {
		return err
	}

	version, _, err := utils.ParseAPIVersionAndKindFromJSON(data)
	if err != nil {
		return err
	}
	if version != "v4" {
		return fmt.Errorf("custom resource definition only support v4 `apiVersion`")
	}
	name, err := utils.ParseNameFromJSON(data)
	if err != nil {
		return err
	}
	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	err = scheduler.UpdateCustomResourceDefinition(c.ClusterID(), name, data)
	if err != nil {
		return fmt.Errorf("failed to update CustomResourceDefinition: %v", err)
	}

	fmt.Printf("success to update CustomResourceDefinition: %s\n", name)
	return nil
}

func updateCustomResource(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionType); err != nil {
		return err
	}

	data, err := c.FileData()
	if err != nil {
		return err
	}

	version, kind, err := utils.ParseAPIVersionAndKindFromJSON(data)
	if err != nil {
		return err
	}
	namespace, name, err := utils.ParseNamespaceNameFromJSON(data)
	if err != nil {
		return err
	}
	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	//validate command line option type
	plural, err := utils.ValidateCustomResourceType(scheduler, c.ClusterID(), version, kind, c.String(utils.OptionType))
	if err != nil {
		return err
	}
	err = scheduler.UpdateCustomResource(c.ClusterID(), version, plural, namespace, name, data)
	if err != nil {
		return fmt.Errorf("failed to update %s: %v", plural, err)
	}

	fmt.Printf("success to update %s: %s\n", plural, name)
	return nil
}
