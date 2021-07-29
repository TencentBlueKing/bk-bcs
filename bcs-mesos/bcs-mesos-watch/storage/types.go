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

package storage

var (
	dataTypeApp                = "Application"
	dataTypeTaskGroup          = "TaskGroup"
	dataTypeCfg                = "Configmap"
	dataTypeSecret             = "Secret"
	dataTypeDeploy             = "Deployment"
	dataTypeSvr                = "Service"
	dataTypeExpSVR             = "ExportService"
	dataTypeEp                 = "Endpoint"
	dataTypeIPPoolStatic       = "IPPoolStatic"
	dataTypeIPPoolStaticDetail = "IPPoolStaticDetail"

	actionDelete = "DELETE"
	actionPut    = "PUT"
	// actionPost   = "POST"

	handlerClusterNamespaceTypeName = "mesos_cluster_namespace_type_name"
	handlerClusterNamespaceType     = "mesos_cluster_namespace_type"
	handlerClusterTypeName          = "mesos_cluster_type_name"
	handlerClusterClusterType       = "mesos_cluster_type"

	handlerAllClusterType = "mesos_all_cluster_type"

	handlerWatchClusterNamespaceTypeName = "mesos_watch_cluster_namespace_type_name"
	handlerEvent                         = "events"
)
