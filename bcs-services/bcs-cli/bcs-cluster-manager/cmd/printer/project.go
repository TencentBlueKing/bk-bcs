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
 */

package printer

import (
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	klog "k8s.io/klog/v2"
)

// PrintClustersListInTable prints the response that list projects
func PrintClustersListInTable(flagOutput string, resp *clustermanager.ListClusterResp) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("list projects output json to stdout failed: %s", err.Error())
		}
	}
	tw := defaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"CLUSTER", "NAME", "FEDERATION_ID", "REGION", "VPC", "PROJECT", "BUSINESS", "ENV",
			"EXCLUSIVE", "TYPE", "CREATOR", "CREATE",
		}
	}())
	tw.SetAutoMergeCells(true)
	for _, item := range resp.Data {
		tw.Append(func() []string {
			return []string{
				item.GetClusterID(), item.GetClusterName(), item.GetRegion(), item.GetVpcID(),
				item.GetProjectID(), item.GetBusinessID(), item.GetEnvironment(),
				strconv.FormatBool(item.GetIsExclusive()), item.GetClusterType(), item.GetCreator(),
				item.GetCreateTime(),
			}
		}())
	}
	tw.Render()
}
