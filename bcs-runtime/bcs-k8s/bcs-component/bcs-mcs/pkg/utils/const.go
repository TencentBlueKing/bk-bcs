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

package utils

const (
	// BcsMcsFinalizerName is the name of the finalizer
	BcsMcsFinalizerName = "mcs.bkbcs.tencent.com/finalizer"
	// EndpointsSliceResourceName is the name of the endpoints resource
	EndpointsSliceResourceName = "endpointslices"

	// ServiceExportKind TODO
	ServiceExportKind = "ServiceExport"
	// ServiceImportKind TODO
	ServiceImportKind = "ServiceImport"
	// EndpointSliceKind TODO
	EndpointSliceKind = "EndpointSlice"

	// ConfigGroupLabel TODO
	ConfigGroupLabel = "mcs.bkbcs.tencent.com/config.group"
	// ConfigVersionLabel TODO
	ConfigVersionLabel = "mcs.bkbcs.tencent.com/config.version"
	// ConfigKindLabel TODO
	ConfigKindLabel = "mcs.bkbcs.tencent.com/config.kind"
	// ConfigNameLabel TODO
	ConfigNameLabel = "mcs.bkbcs.tencent.com/config.name"
	// ConfigNamespaceLabel TODO
	ConfigNamespaceLabel = "mcs.bkbcs.tencent.com/config.namespace"
	// ConfigUIDLabel TODO
	ConfigUIDLabel = "mcs.bkbcs.tencent.com/config.uid"
	// ConfigClusterLabel TODO
	ConfigClusterLabel = "mcs.bkbcs.tencent.com/config.cluster"
	// ConfigCreatedBy TODO
	ConfigCreatedBy = "mcs.bkbcs.tencent.com/created-by"
)
