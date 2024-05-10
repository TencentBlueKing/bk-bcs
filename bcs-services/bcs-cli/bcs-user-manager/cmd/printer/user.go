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

// PrintCreateAdminUserCmdResult prints the response that create admin user
func PrintCreateAdminUserCmdResult(flagOutput string, resp *pkg.CreateAdminUserResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("create admin user output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NAME", "USER_TYPE", "USER_TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.Name,
			strconv.Itoa(int(data.UserType)),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
			data.ExpiresAt.Format(timeFormatter),
			data.GetDeletedAtStr(),
		}
	}())
	tw.Render()
}

// PrintCreateSaasUserCmdResult prints the response that create saas user
func PrintCreateSaasUserCmdResult(flagOutput string, resp *pkg.CreateSaasUserResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("create saas user output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NAME", "USER_TYPE", "USER_TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.Name,
			strconv.Itoa(int(data.UserType)),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
			data.ExpiresAt.Format(timeFormatter),
			data.GetDeletedAtStr(),
		}
	}())
	tw.Render()
}

// PrintCreatePlainUserCmdResult prints the response that create plain user
func PrintCreatePlainUserCmdResult(flagOutput string, resp *pkg.CreatePlainUserResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("create plain user output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NAME", "USER_TYPE", "USER_TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.Name,
			strconv.Itoa(int(data.UserType)),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
			data.ExpiresAt.Format(timeFormatter),
			data.GetDeletedAtStr(),
		}
	}())
	tw.Render()
}

// PrintGetAdminUserCmdResult prints the response that get admin user
func PrintGetAdminUserCmdResult(flagOutput string, resp *pkg.GetAdminUserResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("get admin user output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NAME", "USER_TYPE", "USER_TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.Name,
			strconv.Itoa(int(data.UserType)),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
			data.ExpiresAt.Format(timeFormatter),
			data.GetDeletedAtStr(),
		}
	}())
	tw.Render()
}

// PrintGetSaasUserCmdResult prints the response that get saas user
func PrintGetSaasUserCmdResult(flagOutput string, resp *pkg.GetSaasUserResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("get saas user output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NAME", "USER_TYPE", "USER_TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.Name,
			strconv.Itoa(int(data.UserType)),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
			data.ExpiresAt.Format(timeFormatter),
			data.GetDeletedAtStr(),
		}
	}())
	tw.Render()
}

// PrintGetPlainUserCmdResult prints the response that get plain user
func PrintGetPlainUserCmdResult(flagOutput string, resp *pkg.GetPlainUserResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("get plain user output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NAME", "USER_TYPE", "USER_TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.Name,
			strconv.Itoa(int(data.UserType)),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
			data.ExpiresAt.Format(timeFormatter),
			data.GetDeletedAtStr(),
		}
	}())
	tw.Render()
}

// PrintRefreshSaasTokenCmdResult prints the response that refresh saas user token
func PrintRefreshSaasTokenCmdResult(flagOutput string, resp *pkg.RefreshSaasTokenResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("refresh saas user token output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NAME", "USER_TYPE", "USER_TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.Name,
			strconv.Itoa(int(data.UserType)),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
			data.ExpiresAt.Format(timeFormatter),
			data.GetDeletedAtStr(),
		}
	}())
	tw.Render()
}

// PrintRefreshPlainTokenCmdResult prints the response that refresh plain user token
func PrintRefreshPlainTokenCmdResult(flagOutput string, resp *pkg.RefreshPlainTokenResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("refresh saas user token output json to stdout failed: %s", err.Error())
		}
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NAME", "USER_TYPE", "USER_TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.Name,
			strconv.Itoa(int(data.UserType)),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
			data.ExpiresAt.Format(timeFormatter),
			data.GetDeletedAtStr(),
		}
	}())
	tw.Render()
}
