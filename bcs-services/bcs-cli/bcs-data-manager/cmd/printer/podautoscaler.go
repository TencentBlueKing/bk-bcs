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

// PrintAutoscalerInTable print autoscaler data in table format
func PrintAutoscalerInTable(wide bool, podAutoscaler *bcsdatamanager.PodAutoscaler) {
	if podAutoscaler == nil {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(func() []string {
		r := []string{"NAME", "TYPE", "WL_TYPE", "WL_NAME", "METRIC_TIME", "TRIGGER_TIMES"}
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

	for _, metric := range podAutoscaler.Metrics {
		table.Append(func() []string {
			r := []string{
				podAutoscaler.GetPodAutoscalerName(), podAutoscaler.GetPodAutoscalerType(),
				podAutoscaler.GetWorkloadType(), podAutoscaler.GetWorkloadName(),
				metric.GetTime(), metric.GetTotalSuccessfulRescale(),
			}
			return r
		}())
	}
	table.Render()
}

// PrintAutoscalerListInTable print cluster list data in table format
func PrintAutoscalerListInTable(wide bool, autoscalerList []*bcsdatamanager.PodAutoscaler) {
	if len(autoscalerList) == 0 {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(func() []string {
		r := []string{"NAME", "TYPE", "WL_TYPE", "WL_NAME", "METRIC_TIME", "TRIGGER_TIMES"}
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

	for _, autoscaler := range autoscalerList {
		for _, metric := range autoscaler.Metrics {
			table.Append(func() []string {
				r := []string{
					autoscaler.GetPodAutoscalerName(), autoscaler.GetPodAutoscalerType(),
					autoscaler.GetWorkloadType(), autoscaler.GetWorkloadName(),
					metric.GetTime(), metric.GetTotalSuccessfulRescale(),
				}
				return r
			}())
		}
	}
	table.Render()
}

// PrintAutoscalerInJSON print chart data in json format
func PrintAutoscalerInJSON(autoscaler *bcsdatamanager.PodAutoscaler) {
	if autoscaler == nil {
		return
	}

	var data []byte
	_ = encodeJSONWithIndent(4, autoscaler, &data)
	fmt.Println(string(pretty.Color(pretty.Pretty(data), nil)))
}

// PrintAutoscalerListInJSON print chart data in json format
func PrintAutoscalerListInJSON(autoscalerList []*bcsdatamanager.PodAutoscaler) {
	if autoscalerList == nil {
		return
	}
	for _, autoscaler := range autoscalerList {
		var data []byte
		_ = encodeJSONWithIndent(4, autoscaler, &data)
		fmt.Println(string(pretty.Color(pretty.Pretty(data), nil)))
	}
}
