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
	"os"
	"strconv"

	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"

	"github.com/olekukonko/tablewriter"
)

// PrintReleaseInTable print release data in table format
func PrintReleaseInTable(wide bool, release *helmmanager.ReleaseListData) {
	if release == nil {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(func() []string {
		r := []string{"NAME", "NAMESPACE", "REVISION", "UPDATED", "STATUS", "CHART", "CHART_VERSION", "APP_VERSION"}
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

	for _, rl := range release.Data {
		table.Append(func() []string {
			r := []string{
				rl.GetName(),
				rl.GetNamespace(),
				strconv.Itoa(int(rl.GetRevision())),
				rl.GetUpdateTime(),
				rl.GetStatus(),
				rl.GetChart(),
				rl.GetChartVersion(),
				rl.GetAppVersion(),
			}

			if wide {
				// nothing to do
			}

			return r
		}())
	}
	table.Render()
}

// PrintReleaseDetailInTable print release detail data in table format
func PrintReleaseDetailInTable(wide bool, release []*helmmanager.ReleaseDetail) {
	if release == nil {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(func() []string {
		r := []string{"NAME", "NAMESPACE", "REVISION", "UPDATED", "STATUS", "CHART", "CHART_VERSION", "APP_VERSION"}
		if wide {
			r = append(r, "MESSAGE")
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

	for _, rl := range release {
		table.Append(func() []string {
			r := []string{
				rl.GetName(),
				rl.GetNamespace(),
				strconv.Itoa(int(rl.GetRevision())),
				rl.GetUpdateTime(),
				rl.GetStatus(),
				rl.GetChart(),
				rl.GetChartVersion(),
				rl.GetAppVersion(),
			}

			if wide {
				r = append(r, rl.GetMessage())
			}

			return r
		}())
	}
	table.Render()
}

// PrintReleaseHistoryInTable print release history data in table format
func PrintReleaseHistoryInTable(wide bool, release []*helmmanager.ReleaseHistory) {
	if release == nil {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(func() []string {
		r := []string{"REVISION", "UPDATED", "STATUS", "CHART", "CHART_VERSION", "APP_VERSION", "DESCRIPTION"}
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

	for _, rl := range release {
		table.Append(func() []string {
			r := []string{
				strconv.Itoa(int(rl.GetRevision())),
				rl.GetUpdateTime(),
				rl.GetStatus(),
				rl.GetChart(),
				rl.GetChartVersion(),
				rl.GetAppVersion(),
				rl.GetDescription(),
			}

			if wide {
				// nothing to do
			}

			return r
		}())
	}
	table.Render()
}
