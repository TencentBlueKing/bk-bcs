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
	"net/url"
	"sort"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/storage/v1"
)

func listConfigMap(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID, utils.OptionNamespace); err != nil {
		return err
	}

	storage := v1.NewBcsStorage(utils.GetClientOption())

	// get namespace
	condition := url.Values{}
	condition.Add(filterNamespaceTag, c.Namespace())

	if c.IsAllNamespace() {
		var err error
		if condition, err = getNamespaceFilter(storage, c.ClusterID()); err != nil {
			return err
		}
	}

	list, err := storage.ListConfigMap(c.ClusterID(), condition)
	if err != nil {
		return fmt.Errorf("failed to list configmap: %v", err)
	}

	sort.Sort(list)
	return printListConfigMap(list)
}

func printListConfigMap(list v1.ConfigMapList) error {
	if len(list) == 0 {
		fmt.Printf("Found no configmap\n")
		return nil
	}

	fmt.Printf("%-50s  %-30s\n",
		"NAME",
		"NAMESPACE")
	for _, status := range list {
		fmt.Printf("%-50s %-30s\n",
			status.Data.ObjectMeta.Name,
			status.Data.ObjectMeta.NameSpace)
	}
	return nil
}
