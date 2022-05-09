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

// PrintChartDetailInTable print chart detail data in table format
func PrintChartDetailInTable(wide bool, chartDetail *helmmanager.ChartDetail) {
	if chartDetail == nil {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(func() []string {
		r := []string{"FILE", "NAME", "IS_VALUES", "IS_README"}
		if wide {
			// nothing to do
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

	for _, ct := range chartDetail.Contents {
		table.Append(func() []string {
			r := []string{
				ct.GetPath(),
				ct.GetName(),
				func() string {
					for _, f := range chartDetail.GetValuesFile() {
						if ct.GetPath() == f {
							return "yes"
						}
					}
					return ""
				}(),
				func() string {
					if ct.GetPath() == chartDetail.GetReadme() {
						return "yes"
					}
					return ""
				}(),
			}

			if wide {
				// nothing to do
			}

			return r
		}())
	}
	table.Render()
}

// PrintChartDetailInJson print chart detail data in json format
func PrintChartDetailInJson(chart *helmmanager.ChartDetail) {
	if chart == nil {
		return
	}

	var data []byte
	_ = encodeJsonWithIndent(4, chart, &data)
	fmt.Println(string(pretty.Color(pretty.Pretty(data), nil)))
}
