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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	v4 "github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"

	simplejson "github.com/bitly/go-simplejson"
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
	//print all datas
	//Name - ShortName - apiVersion - Kind - CreatedTime
	fmt.Printf(
		"%-50s %-20s %-25s %-20s %-21s\n",
		"NAME",
		"CMDTYPE",
		"APIVERSION",
		"KIND",
		"CRAETEDTIME",
	)
	for _, item := range crdList.Items {
		apiVersion := item.Spec.Group + "/" + item.Spec.Version
		fmt.Printf(
			"%-50s %-20s %-25s %-20s %-21s\n",
			item.GetName(),
			item.Spec.Names.Singular,
			apiVersion,
			item.Spec.Names.Kind,
			item.GetCreationTimestamp(),
		)
	}
	return nil
}

func listCustomResource(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionType); err != nil {
		return err
	}
	namespace := c.String(utils.OptionNamespace)
	allNamespaces := c.Bool(utils.OptionAllNamespace)
	if namespace == "" && !allNamespaces {
		return fmt.Errorf("namespace or all-namespace must be specified")
	}
	if allNamespaces {
		namespace = v4.AllNamespace
	}
	scheduler := v4.NewBcsScheduler(utils.GetClientOption())
	//validate command line option type
	apiVersion, plural, err := utils.GetCustomResourceType(scheduler, c.ClusterID(), c.String(utils.OptionType))
	if err != nil {
		return err
	}
	allBytes, err := scheduler.ListCustomResource(c.ClusterID(), apiVersion, plural, namespace)
	if err != nil {
		return fmt.Errorf("failed to List %s, %v", plural, err)
	}
	//parse item list formation
	json, err := simplejson.NewJson(allBytes)
	if err != nil {
		return fmt.Errorf("list %s failed, response is not expected json format: %s", plural, err.Error())
	}
	items := json.Get("items")
	dataList, err := items.Array()
	if err != nil {
		return fmt.Errorf("list %s failed, No items array response, %s", plural, err.Error())
	}
	len := len(dataList)
	if len == 0 {
		fmt.Printf("Found No Resources\n")
		return nil
	}
	//print simple information
	fmt.Printf(
		"%-30s %-20s %-25s\n",
		"NAME",
		"NAMESPACE",
		"CRAETEDTIME",
	)
	for i := 0; i < len; i++ {
		meta := items.GetIndex(i).Get("metadata")
		destName, _ := meta.Get("name").String()
		destNS, _ := meta.Get("namespace").String()
		createdTime, _ := meta.Get("creationTimestamp").String()
		fmt.Printf(
			"%-30s %-20s %-25s\n",
			destName,
			destNS,
			createdTime,
		)
	}
	return nil
}
