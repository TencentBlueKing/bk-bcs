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

package options

// BCSEdition version
type BCSEdition string

var (
	// InnerEdition inner
	InnerEdition BCSEdition = "inner_edition"
	// CommunicationEdition community
	CommunicationEdition BCSEdition = "communication_edition"
	// EnterpriseEdition enterprise
	EnterpriseEdition BCSEdition = "enterprise_edition"
)

// String toString
func (b BCSEdition) String() string {
	return string(b)
}

// IsInnerEdition innerVersion
func (b BCSEdition) IsInnerEdition() bool {
	if b == InnerEdition {
		return true
	}

	return false
}

// IsCommunicationEdition communityVersion
func (b BCSEdition) IsCommunicationEdition() bool {
	if b == CommunicationEdition {
		return true
	}

	return false
}

// IsEnterpriseEdition enterpriseVersion
func (b BCSEdition) IsEnterpriseEdition() bool {
	if b == EnterpriseEdition {
		return true
	}

	return false
}
