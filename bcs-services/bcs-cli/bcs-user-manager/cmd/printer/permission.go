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
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/pkg"
)

// PrintGrantPermissionCmdResult prints the response that grant permissions
func PrintGrantPermissionCmdResult(flagOutput string, resp *pkg.GrantPermissionResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("grant permissions output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"RESOURCE_TYPE", "RESOURCE", "ROLE",
		}
	}())
	tw.Render()
}

// PrintRevokePermissionCmdResult prints the response that revoke permissions
func PrintRevokePermissionCmdResult(flagOutput string, resp *pkg.RevokePermissionResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("revoke permissions output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"RESOURCE_TYPE", "RESOURCE", "ROLE",
		}
	}())
	tw.Render()
}

// PrintGetPermissionCmdResult prints the response that get permissions
func PrintGetPermissionCmdResult(flagOutput string, resp *pkg.GetPermissionResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("get permissions output json to stdout failed: %s", err.Error())
		}
	}

	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"RESOURCE_TYPE", "RESOURCE", "ROLE",
		}
	}())
	// 添加页脚
	tw.SetFooter([]string{"", "Total", strconv.Itoa(len(resp.Data))})
	// 合并相同值的列
	//tw.SetAutoMergeCells(true)
	for _, item := range resp.Data {
		tw.Append(func() []string {
			return []string{
				item.ResourceType,
				item.Resource,
				item.Role,
			}
		}())
	}
	tw.Render()
}

// PrintVerifyPermissionCmdResult prints the response that verify permissions
func PrintVerifyPermissionCmdResult(flagOutput string, resp *pkg.VerifyPermissionResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("verify permissions output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"ALLOWED", "MESSAGE",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.FormatBool(data.Allowed),
			data.Message,
		}
	}())
	tw.Render()
}

// PrintVerifyPermissionV2CmdResult prints the response that verify permissions v2
func PrintVerifyPermissionV2CmdResult(flagOutput string, resp *pkg.VerifyPermissionV2Response) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("verify permissions v2 output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"ALLOWED", "MESSAGE",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.FormatBool(data.Allowed),
			data.Message,
		}
	}())
	tw.Render()
}
