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

package resource

import (
	types "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
)

// IPDriver driver for applying/releasing ip source
type IPDriver interface {
	GetIPAddr(host, containerID, requestIP string) (*types.IPInfo,
		error) // GetIPAddr get available ip resource for contaienr
	ReleaseIPAddr(host string, containerID string,
		ipInfo *types.IPInfo) error // ReleaseIPAddr release ip address for container
	GetHostInfo(host string) (*types.HostInfo, error) // Get host info from driver
}
