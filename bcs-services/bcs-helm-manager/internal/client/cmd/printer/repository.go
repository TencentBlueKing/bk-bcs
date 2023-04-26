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
	"encoding/json"
	"fmt"
	"os"

	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"

	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/pretty"
)

// PrintRepositoryInTable print repository data in table format
func PrintRepositoryInTable(wide bool, repository []*helmmanager.Repository) {
	if repository == nil {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(func() []string {
		r := []string{"NAME", "PROJECT_ID", "TYPE", "REMOTE"}
		if wide {
			r = append(r, "CREATE_BY", "CREATE_TIME", "UPDATE_BY", "UPDATE_TIME")
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

	for _, repo := range repository {
		table.Append(func() []string {
			r := []string{
				repo.GetName(), repo.GetProjectCode(), repo.GetType(), fmt.Sprintf("%t", repo.GetRemote()),
			}

			if wide {
				r = append(r, repo.GetCreateBy(), repo.GetCreateTime(), repo.GetUpdateBy(), repo.GetUpdateTime())
			}

			return r
		}())
	}
	table.Render()
}

// PrintRepositoryInJSON print repository data in json format
func PrintRepositoryInJSON(repository []*helmmanager.Repository) {
	if repository == nil {
		return
	}

	for _, repo := range repository {
		data, _ := json.Marshal(repo)

		fmt.Println(string(pretty.Color(pretty.Pretty(data), nil)))
	}
}
