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

package dbprivilege

const (
	// DBPrivilegePluginName plugin name for db privilege
	DBPrivilegePluginName = "dbpriv"
	// NetworkTypeOverlay overlay network
	NetworkTypeOverlay = "overlay"
	// NetworkTypeUnderlay underlay network
	NetworkTypeUnderlay = "underlay"
	// DbPrivilegeSecretName the name of secret to store db privilege info
	DbPrivilegeSecretName = "bcs-db-privilege" // nolint

	// BcsPodName podName
	BcsPodName = "io_tencent_bcs_pod_name"
	// BcsPodNamespace pod namespace
	BcsPodNamespace = "io_tencent_bcs_pod_namespace"
	// BcsPrivilegeServiceURL service 域名地址
	BcsPrivilegeServiceURL = "io_tencent_bcs_privilege_service_url"
	// BcsPrivilegeHost  service 域名地址
	BcsPrivilegeHost = "http://%s.%s.svc.cluster.local:%d"
	// BcsPrivilegeDbmOptimizeEnabled dbm优化开关
	BcsPrivilegeDbmOptimizeEnabled = "io_tencent_bcs_privilege_dbm_optimize_enabled"
	// BcsPrivilegeServiceTicketTimer 服务间隔时间单位 int
	BcsPrivilegeServiceTicketTimer = "io_tencent_bcs_privilege_service_ticket_timer"
	// BcsPrivilegeDBMAuthStatusDone dbm授权成功
	BcsPrivilegeDBMAuthStatusDone = "done"
	// BcsPrivilegeDBMAuthStatusPending  dbm授权未成功
	BcsPrivilegeDBMAuthStatusPending = "pending"
	// BcsPrivilegeDBMAuthStatusChanged  dbm授权信息已被修改
	BcsPrivilegeDBMAuthStatusChanged = "changed"
)
