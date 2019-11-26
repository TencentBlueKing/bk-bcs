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

package list

import (
	"fmt"
	"strings"

	"bk-bcs/bcs-services/bcs-client/cmd/utils"
	v4 "bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"
)

//listCustomResourceDefinition list all CRDs from mesos-driver
func listCustomResourceDefinition(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID); err != nil {
		return err
	}
	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	crdList, err := scheduler.ListCustomResourceDefinition(c.ClusterID())
	if err != nil {
		return fmt.Errorf("failed to List all CustomResourceDefinition: %v", err)
	}
	if len(crdList.Items) == 0 {
		fmt.Printf("Found no customresourcedefinition\n")
		return nil
	}
	//print all datas
	//Name - ShortName - apiVersion - Kind - CreatedTime
	fmt.Printf(
		"%-50s %-10s %-50s %-20s %-21s\n",
		"NAME",
		"SHORTNAME",
		"APIVERSION",
		"KIND",
		"CRAETEDTIME",
	)
	for _, item := range crdList.Items {
		apiVersion := item.Spec.Group + "/" + item.Spec.Version
		shortNames := strings.Join(item.Spec.Names.ShortNames, ",")
		fmt.Printf(
			"%-50s %-10s %-50s %-20s\n",
			item.GetName(),
			shortNames,
			apiVersion,
			item.Spec.Names.Kind,
			item.GetCreationTimestamp(),
		)
	}
	return nil
}

func listCustomResource(c *utils.ClientContext) error {
	return nil
}
