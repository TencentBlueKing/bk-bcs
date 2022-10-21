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

package printer

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/pretty"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// PrintProjectsListInTable prints the response that list projects
func PrintProjectsListInTable(flagOutput string, resp *bcsproject.ListProjectsResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("list projects output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"PROJECT_ID", "PROJECT_CODE", "NAME", "BUSINESS_ID", "CREATOR", "UPDATER", "CREATE", "UPDATE",
		}
	}())
	// 添加页脚
	tw.SetFooter([]string{"", "", "", "", "", "", "Total", strconv.Itoa(int(resp.Data.Total))})
	// 合并相同值的列
	//tw.SetAutoMergeCells(true)
	for _, item := range resp.Data.Results {
		tw.Append(func() []string {
			return []string{
				item.GetProjectID(),
				item.GetProjectCode(),
				item.GetName(),
				item.GetBusinessID(),
				item.GetCreator(),
				item.GetUpdater(),
				item.GetCreateTime(),
				item.GetUpdateTime(),
			}
		}())
	}
	tw.Render()
}

<<<<<<< HEAD
// PrintUpdateProjectInJSON prints the response that edit project
=======
>>>>>>> cd831f67dcced2448d87af2258a3604299f448fc
func PrintUpdateProjectInJSON(project *bcsproject.ProjectResponse) {
	if project == nil {
		return
	}

	var data []byte
	_ = encodeJSONWithIndent(4, project, &data)
	fmt.Println(string(pretty.Color(pretty.Pretty(data), nil)))
}
