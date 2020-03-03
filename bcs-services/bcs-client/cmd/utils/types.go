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

package utils

import (
	v4 "bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"
	"fmt"
)

/*
command line option for bcs-client
*/
const (
	OptionClusterID     = "clusterid"
	OptionNamespace     = "namespace"
	OptionAllNamespace  = "all-namespaces"
	OptionName          = "name"
	OptionTaskGroupName = "tgname"
	OptionType          = "type"
	OptionClusterType   = "clustertype"
	OptionFile          = "from-file"
	OptionIP            = "ip"
	OptionList          = "list"
	OptionInspect       = "inspect"
	OptionUpdate        = "update"
	OptionUpsert        = "upsert"
	OptionSet           = "set"
	OptionDelete        = "delete"
	OptionInstance      = "instance"
	OptionEnforce       = "enforce"
	OptionKey           = "key"
	OptionString        = "string"
	OptionScalar        = "scalar"
	OptionAll           = "all"
)

//ValidateCustomResourceType check if speicifed CustomResource was registered before.
//return plural for api request, error if happened
func ValidateCustomResourceType(sche v4.Scheduler, clusterID, apiVersion, kind, cmdType string) (string, error) {
	list, err := sche.ListCustomResourceDefinition(clusterID)
	if err != nil {
		DebugPrintf("list all CustomResourceDefinition in cluster %s failed, %s", clusterID, err.Error())
		return "", err
	}
	if len(list.Items) == 0 {
		return "", fmt.Errorf("invalid type %s", cmdType)
	}
	DebugPrintf("##DEBUG##: %v", list.Items)
	//validate apiVersion, kind & type
	found := false
	plural := ""
	for _, item := range list.Items {
		dstAPIVersion := item.Spec.Group + "/" + item.Spec.Version
		if apiVersion != dstAPIVersion {
			continue
		}
		if item.Spec.Names.Singular == cmdType && item.Spec.Names.Kind == kind {
			found = true
			plural = item.Spec.Names.Plural
			break
		}
	}
	if !found {
		return "", fmt.Errorf("invalid type: %s, even not match in CustomResource", cmdType)
	}
	return plural, nil
}

//GetCustomResourceType get Custom Resource type from command line option
//returns: apiVersion, plural, errror if happened
func GetCustomResourceType(sche v4.Scheduler, cluster, cmdType string) (string, string, error) {
	list, err := sche.ListCustomResourceDefinition(cluster)
	if err != nil {
		DebugPrintf("list all CustomResourceDefinition failed, %s", err.Error())
		return "", "", err
	}
	if len(list.Items) == 0 {
		return "", "", fmt.Errorf("invalid type %s", cmdType)
	}
	DebugPrintf("##DEBUG##: %v", list.Items)
	//validate apiVersion, kind & type
	for _, item := range list.Items {
		if item.Spec.Names.Singular == cmdType {
			apiVersion := item.Spec.Group + "/" + item.Spec.Version
			return apiVersion, item.Spec.Names.Plural, nil
		}
	}
	return "", "", fmt.Errorf("invalid type: %s, even not match CustomResource", cmdType)
}

//GetCustomResourceTypeByKind get Custom Resource type by
//returns: apiVersion, plural, errror if happened
func GetCustomResourceTypeByKind(sche v4.Scheduler, cluster, kind string) (string, string, error) {
	list, err := sche.ListCustomResourceDefinition(cluster)
	if err != nil {
		DebugPrintf("list all CustomResourceDefinition failed, %s", err.Error())
		return "", "", err
	}
	if len(list.Items) == 0 {
		return "", "", fmt.Errorf("invalid kind %s", kind)
	}
	DebugPrintf("##DEBUG##: %v", list.Items)
	//validate apiVersion, kind & type
	for _, item := range list.Items {
		if item.Spec.Names.Kind == kind {
			apiVersion := item.Spec.Group + "/" + item.Spec.Version
			return apiVersion, item.Spec.Names.Plural, nil
		}
	}
	return "", "", fmt.Errorf("invalid kind: %s, even not match CustomResource", kind)
}
