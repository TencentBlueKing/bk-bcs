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
	"fmt"
	"os"

	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/pretty"
)

// PrintWorkloadInTable print workload data in table format
func PrintWorkloadInTable(wide bool, workload *bcsdatamanager.Workload) {
	if workload == nil {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(func() []string {
		r := []string{"NAME", "METRIC_TIME", "INSTANCE_CNT",
			"CPU_REQ", "LOAD_CPU", "CPU_USAGE", "MM_REQ", "LOAD_CPU", "MM_USAGE"}
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

	for _, metric := range workload.Metrics {
		table.Append(func() []string {
			r := []string{
				workload.GetWorkloadName(), metric.GetTime(),
				metric.GetInstanceCount(),
				metric.GetCPURequest(),
				metric.GetCPUUsageAmount(),
				metric.GetCPUUsage(),
				metric.GetMemoryRequest(),
				metric.GetMemoryUsageAmount(),
				metric.GetMemoryUsage(),
			}
			return r
		}())
	}
	table.Render()
}

// PrintWorkloadListInTable print cluster list data in table format
func PrintWorkloadListInTable(wide bool, workloadList []*bcsdatamanager.Workload) {
	if len(workloadList) == 0 {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(func() []string {
		r := []string{"NAME", "METRIC_TIME", "INSTANCE_CNT",
			"CPU_REQ", "LOAD_CPU", "CPU_USAGE", "MM_REQ", "LOAD_MM", "MM_USAGE"}
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

	for _, workload := range workloadList {
		for _, metric := range workload.Metrics {
			table.Append(func() []string {
				r := []string{
					workload.GetWorkloadName(), metric.GetTime(),
					metric.GetInstanceCount(),
					metric.GetCPURequest(),
					metric.GetCPUUsageAmount(),
					metric.GetCPUUsage(),
					metric.GetMemoryRequest(),
					metric.GetMemoryUsageAmount(),
					metric.GetMemoryUsage(),
				}
				return r
			}())
		}
	}
	table.Render()
}

// PrintWorkloadInJSON print chart data in json format
func PrintWorkloadInJSON(workload *bcsdatamanager.Workload) {
	if workload == nil {
		return
	}

	var data []byte
	_ = encodeJSONWithIndent(4, workload, &data)
	fmt.Println(string(pretty.Color(pretty.Pretty(data), nil)))
}

// PrintWorkloadListInJSON print chart data in json format
func PrintWorkloadListInJSON(workloadList []*bcsdatamanager.Workload) {
	if workloadList == nil {
		return
	}
	for _, workload := range workloadList {
		var data []byte
		_ = encodeJSONWithIndent(4, workload, &data)
		fmt.Println(string(pretty.Color(pretty.Pretty(data), nil)))
	}
}
