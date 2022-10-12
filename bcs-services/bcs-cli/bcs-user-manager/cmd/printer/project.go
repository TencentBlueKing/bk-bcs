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
	"strconv"

	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/pkg"
)

// PrintProjectsListInTable prints the response that list projects
func PrintProjectsListInTable(flagOutput string, resp *pkg.GetAdminUserResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("list projects output json to stdout failed: %s", err.Error())
		}
	}
	tw := defaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NAME", "TYPE", "TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	user := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(user.ID)), user.Name, strconv.Itoa(int(user.UserType)), user.UserToken, user.CreatedBy,
			user.CreatedAt.Format(timeFormatter),
			user.UpdatedAt.Format(timeFormatter),
			user.ExpiresAt.Format(timeFormatter),
			user.DeletedAt.Format(timeFormatter),
		}
	}())
	tw.Render()
}
