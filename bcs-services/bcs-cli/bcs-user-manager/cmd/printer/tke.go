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
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	klog "k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/pkg"
)

// PrintApplyTkeCidrCmdResult prints the response that apply tke cidrs
func PrintApplyTkeCidrCmdResult(flagOutput string, resp *pkg.ApplyTkeCidrResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("apply tke cidrs output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"VPC", "CIDR", "IP_NUMBER", "STATUS",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			data.Vpc,
			data.Cidr,
			strconv.Itoa(int(data.IpNumber)),
			data.Status,
		}
	}())
	tw.Render()
}

// PrintAddTkeCidrCmdResult prints the response that add tkecidrs
func PrintAddTkeCidrCmdResult(flagOutput string, resp *pkg.AddTkeCidrResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("add tkecidrs output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"COUNT", "VPC", "IP_NUMBER", "STATUS",
		}
	}())
	tw.Render()
}

// PrintListTkeCidrCmdResult prints the response that list TkeCidr
func PrintListTkeCidrCmdResult(flagOutput string, resp *pkg.ListTkeCidrResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("list TkeCidr output json to stdout failed: %s", err.Error())
		}
	}

	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"COUNT", "VPC", "IP_NUMBER", "STATUS",
		}
	}())
	// 添加页脚
	tw.SetFooter([]string{"", "Total", strconv.Itoa(len(resp.Data))})
	// 合并相同值的列
	// tw.SetAutoMergeCells(true)
	for _, item := range resp.Data {
		tw.Append(func() []string {
			return []string{
				strconv.Itoa(item.Count),
				item.Vpc,
				strconv.Itoa(int(item.IpNumber)),
				item.Status,
			}
		}())
	}
	tw.Render()
}

// PrintReleaseTkeCidrCmdResult prints the response that release tkecidrs
func PrintReleaseTkeCidrCmdResult(flagOutput string, resp *pkg.ReleaseTkeCidrResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("release tkecidrs output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"VPC", "CIDR", "IP_NUMBER", "STATUS",
		}
	}())
	tw.Render()
}

// PrintSyncCredentialsResult prints the response that sync cluster tkecidrs
func PrintSyncCredentialsResult(flagOutput string, resp *pkg.SyncTkeClusterCredentialsResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("sync cluster tkecidrs output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"VPC", "CIDR", "IP_NUMBER", "STATUS",
		}
	}())
	tw.Render()
}
