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

package inspect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	v4 "github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"
)

func inspectCustomResourceDefinition(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionName); err != nil {
		return err
	}
	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	crd, err := scheduler.GetCustomResourceDefinition(c.ClusterID(), c.String(utils.OptionName))
	if err != nil {
		return fmt.Errorf("failed to Get CustomResourceDefinition: %s", err.Error())
	}
	return printInspect(crd)
}

func inspectCustomResource(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionNamespace, utils.OptionName); err != nil {
		return err
	}
	namespace := c.String(utils.OptionNamespace)
	name := c.String(utils.OptionName)
	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	// validate command line option type
	apiVersion, plural, err := utils.GetCustomResourceType(scheduler, c.ClusterID(), c.String(utils.OptionType))
	if err != nil {
		return err
	}
	crd, err := scheduler.GetCustomResource(c.ClusterID(), apiVersion, plural, namespace, name)
	if err != nil {
		return fmt.Errorf("failed to Get %s: %v", plural, err)
	}
	utils.DebugPrintf("original CustomResource: %s", string(crd))
	var buffer bytes.Buffer
	if err := json.Indent(&buffer, crd, "", "  "); err != nil {
		return fmt.Errorf("pretty print CustomResource failed, %s", err.Error())
	}
	fmt.Println(buffer.String())
	return nil
}
