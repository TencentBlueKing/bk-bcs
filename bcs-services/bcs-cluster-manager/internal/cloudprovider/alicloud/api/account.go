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

package api

import (
	"errors"

	"github.com/alibabacloud-go/tea/tea"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// ValidateAccount verifies long-lived AccessKeys without storing STS identity information.
// SecretID and SecretKey reuse the common encrypted account storage format.
func ValidateAccount(account *proto.Account) error {
	client, err := NewClient(&cloudprovider.CommonOption{Account: account})
	if err != nil {
		return err
	}
	identity, err := client.GetCallerIdentity()
	if err != nil {
		return err
	}
	identityType := tea.StringValue(identity.IdentityType)
	if identityType != "Account" && identityType != "RAMUser" {
		return errors.New("alicloud account requires a long-lived account or RAM user AccessKey")
	}
	return nil
}
