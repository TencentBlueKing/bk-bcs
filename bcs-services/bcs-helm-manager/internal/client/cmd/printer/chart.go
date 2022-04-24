/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package printer

import (
	"fmt"
	"os"

	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"

	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/pretty"
)

// PrintChartInTable print chart data in table format
func PrintChartInTable(wide bool, chart *helmmanager.ChartListData) {
	if chart == nil {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(func() []string {
		r := []string{"NAME", "VERSION", "APP_VERSION", "DESCRIPTION"}
		if wide {
			r = append(r, "KEY", "CREATE_BY", "CREATE_TIME", "UPDATE_BY", "UPDATE_TIME")
		}
		return r
	}())
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("    ")
	table.SetNoWhiteSpace(true)

	for _, ct := range chart.Data {
		table.Append(func() []string {
			r := []string{
				ct.GetName(), ct.GetLatestVersion(), ct.GetLatestAppVersion(), cut(ct.GetLatestDescription(), 50),
			}

			if wide {
				r = append(r, ct.GetKey(), ct.GetCreateBy(), ct.GetCreateTime(), ct.GetUpdateBy(), ct.GetUpdateTime())
			}

			return r
		}())
	}
	table.Render()
}

// PrintChartInJson print chart data in json format
func PrintChartInJson(chart *helmmanager.ChartListData) {
	if chart == nil {
		return
	}

	for _, ct := range chart.Data {
		var data []byte
		_ = encodeJsonWithIndent(4, ct, &data)

		fmt.Println(string(pretty.Color(pretty.Pretty(data), nil)))
	}
}
