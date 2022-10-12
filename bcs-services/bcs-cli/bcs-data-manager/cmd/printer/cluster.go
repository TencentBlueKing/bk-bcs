/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package printer

import (
	"fmt"
	"os"

	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/pretty"
)

// PrintClusterInTable print cluster data in table format
func PrintClusterInTable(wide bool, cluster *bcsdatamanager.Cluster) {
	if cluster == nil {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(func() []string {
		r := []string{"ID", "METRIC_TIME", "NODE_CNT", "WORKLOAD_CNT", "INSTANCE_CNT",
			"TOTAL_CPU", "LOAD_CPU", "CPU_USAGE", "TOTAL_MM", "LOAD_MM", "MM_USAGE"}
		if wide {
			r = append(r, "AVG_CPU")
		}
		return r
	}())
	// table.SetAutoWrapText(false)
	// table.SetAutoFormatHeaders(true)
	// table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	// table.SetAlignment(tablewriter.ALIGN_LEFT)
	// table.SetCenterSeparator("")
	// table.SetColumnSeparator("")
	// table.SetRowSeparator("")
	// table.SetHeaderLine(false)
	// table.SetBorder(false)
	// table.SetTablePadding("")
	// table.SetNoWhiteSpace(true)
	table.SetAutoMergeCellsByColumnIndex([]int{0})
	table.SetRowLine(true)

	for _, metric := range cluster.Metrics {
		table.Append(func() []string {
			r := []string{
				cluster.GetClusterID(), metric.GetTime(),
				metric.GetNodeCount(),
				metric.GetWorkloadCount(),
				metric.GetInstanceCount(),
				metric.GetTotalCPU(),
				metric.GetTotalLoadCPU(),
				metric.GetCPUUsage(),
				metric.GetTotalMemory(),
				metric.GetTotalLoadMemory(),
				metric.GetMemoryUsage(),
			}
			if wide {
				r = append(r, metric.GetAvgLoadCPU())
			}

			return r
		}())
	}
	table.Render()
}

// PrintClusterListInTable print cluster list data in table format
func PrintClusterListInTable(wide bool, clusterList []*bcsdatamanager.Cluster) {
	if len(clusterList) == 0 {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(func() []string {
		r := []string{"ID", "METRIC_TIME", "NODE_CNT", "WORKLOAD_CNT", "INSTANCE_CNT",
			"TOTAL_CPU", "LOAD_CPU", "CPU_USAGE", "TOTAL_MM", "LOAD_MM", "MM_USAGE"}
		if wide {
			r = append(r, "AVG_CPU")
		}
		return r
	}())
	// table.SetAutoWrapText(false)
	// table.SetAutoFormatHeaders(true)
	// table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	// table.SetAlignment(tablewriter.ALIGN_LEFT)
	// table.SetCenterSeparator("")
	// table.SetColumnSeparator("")
	// table.SetRowSeparator("")
	// table.SetHeaderLine(false)
	// table.SetBorder(false)
	// table.SetTablePadding("")
	// table.SetNoWhiteSpace(true)
	table.SetAutoMergeCellsByColumnIndex([]int{0})
	table.SetRowLine(true)

	for _, cluster := range clusterList {
		for _, metric := range cluster.Metrics {
			table.Append(func() []string {
				r := []string{
					cluster.GetClusterID(), metric.GetTime(),
					metric.GetNodeCount(), metric.GetWorkloadCount(),
					metric.GetInstanceCount(),
					metric.GetTotalCPU(),
					metric.GetTotalLoadCPU(),
					metric.GetCPUUsage(),
					metric.GetTotalMemory(),
					metric.GetTotalLoadMemory(),
					metric.GetMemoryUsage(),
				}
				if wide {
					r = append(r, metric.GetAvgLoadCPU())
				}

				return r
			}())
		}
	}
	table.Render()
}

// PrintClusterInJSON print chart data in json format
func PrintClusterInJSON(cluster *bcsdatamanager.Cluster) {
	if cluster == nil {
		return
	}

	var data []byte
	_ = encodeJSONWithIndent(4, cluster, &data)
	fmt.Println(string(pretty.Color(pretty.Pretty(data), nil)))
}

// PrintClusterListInJSON print chart data in json format
func PrintClusterListInJSON(clusterList []*bcsdatamanager.Cluster) {
	if clusterList == nil {
		return
	}
	for _, cluster := range clusterList {
		var data []byte
		_ = encodeJSONWithIndent(4, cluster, &data)
		fmt.Println(string(pretty.Color(pretty.Pretty(data), nil)))
	}
}
