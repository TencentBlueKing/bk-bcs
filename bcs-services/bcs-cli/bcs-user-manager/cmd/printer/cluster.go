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

// Package printer xxx
package printer

import (
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	klog "k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/pkg"
)

// PrintCreateClusterCmdResult prints the response that create cluster
func PrintCreateClusterCmdResult(flagOutput string, resp *pkg.CreateClusterResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("create cluster output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)

	tw.SetHeader(func() []string {
		return []string{
			"ID", "CLUSTER_TYPE", "TKE_CLUSTER_ID", "TKE_CLUSTER_REGION", "CREATOR_ID", "CREATED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			data.ID,
			strconv.Itoa(int(data.ClusterType)),
			data.TkeClusterId,
			data.TkeClusterRegion,
			strconv.Itoa(int(data.CreatorId)),
			data.CreatedAt.Format(timeFormatter),
		}
	}())
	tw.Render()
}

// PrintCreateRegisterTokenCmdResult prints the response that create register token
func PrintCreateRegisterTokenCmdResult(flagOutput string, resp *pkg.CreateRegisterTokenResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("create register token output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"ID", "CLUSTER_ID", "TOKEN", "CREATED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.ClusterId,
			data.Token,
			data.CreatedAt.Format(timeFormatter),
		}
	}())
	tw.Render()
}

// PrintGetRegisterTokenCmdResult prints the response that get register token
func PrintGetRegisterTokenCmdResult(flagOutput string, resp *pkg.GetRegisterTokenResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("get register token output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"ID", "CLUSTER_ID", "TOKEN", "CREATED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.ClusterId,
			data.Token,
			data.CreatedAt.Format(timeFormatter),
		}
	}())
	tw.Render()
}

// PrintGetCredentialsCmdResult prints the response that get credentials
func PrintGetCredentialsCmdResult(flagOutput string, resp *pkg.GetCredentialsResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("get register token output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"ID", "CLUSTER_ID", "SERVER_ADDRESSES", "CA_CERT_DATA", "USER_TOKEN", "CLUSTER_DOMAIN", "CREATED_AT", "UPDATED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.ClusterId,
			data.ServerAddresses,
			data.CaCertData,
			data.UserToken,
			data.ClusterDomain,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
		}
	}())
	tw.Render()
}

// PrintListCredentialsCmdResult prints the response that list credentials
func PrintListCredentialsCmdResult(flagOutput string, resp *pkg.ListCredentialsResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("list credentials output json to stdout failed: %s", err.Error())
		}
	}

	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"CLUSTER_ID", "SERVER_ADDRESSES", "CA_CERT_DATA", "USER_TOKEN", "CLUSTER_DOMAIN",
		}
	}())
	// 添加页脚
	tw.SetFooter([]string{"", "", "", "Total", strconv.Itoa(len(resp.Data))})
	// 合并相同值的列
	// tw.SetAutoMergeCells(true)
	for key, item := range resp.Data {
		tw.Append(func() []string {
			return []string{
				key,
				item.ServerAddresses,
				item.CaCertData,
				item.UserToken,
				item.ClusterDomain,
			}
		}())
	}
	tw.Render()
}

// PrintUpdateCredentialsCmdResult prints the response that update credentials
func PrintUpdateCredentialsCmdResult(flagOutput string, resp *pkg.UpdateCredentialsResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("update credentials output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"CLUSTER_ID", "SERVER_ADDRESSES", "CA_CERT_DATA", "USER_TOKEN", "CLUSTER_DOMAIN",
		}
	}())
	tw.Render()
}
