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

// Package helm 获取helm manager client
package helm

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
)

// GetClient 获取helm manager client
func GetClient() (*helmmanager.HelmClientWrapper, func(), error) {
	helmManagerClient, closeFunc, err := helmmanager.GetClient(common.ServiceDomain)
	if err != nil {
		return nil, nil, err
	}
	return helmManagerClient, closeFunc, nil
}
