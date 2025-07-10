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

package randhostport

const (
	pluginName                        = "randhostport"
	pluginAnnotationKey               = pluginName + ".webhook.bkbcs.tencent.com"
	pluginAnnotationValue             = "true"
	pluginPortsAnnotationKey          = "ports." + pluginAnnotationKey
	pluginContainerPortsAnnotationKey = "randcontainerport." + pluginAnnotationKey

	podHostportLabelFlagKey   = pluginAnnotationKey
	podHostportLabelFlagValue = pluginAnnotationValue

	podHostportLabelSuffix = "." + pluginAnnotationKey

	envRandHostportPrefix       = "BCS_RANDHOSTPORT_FOR_CONTAINER_PORT_"
	envRandHostportHostIP       = "BCS_RANDHOSTPORT_HOSTIP"
	envRandHostportPodName      = "BCS_RANDHOSTPORT_POD_NAME"
	envRandHostportPodNamespace = "BCS_RANDHOSTPORT_POD_NAMESPACE"

	annotationsRandHostportPrefix = pluginAnnotationKey + "."

	// PatchOperationAdd patch add operation
	PatchOperationAdd = "add"
	// PatchOperationReplace patch replace operation
	PatchOperationReplace = "replace"
	// PatchOperationRemove patch remove operation
	PatchOperationRemove = "remove"

	// PatchPathContainerHostPort path for patching container port
	PatchPathContainerHostPort = "/spec/containers/%v/ports/%v/hostPort"
	// PatchPathContainerContainerPort path for patching container port
	PatchPathContainerContainerPort = "/spec/containers/%v/ports/%v/containerPort"
	// PatchPathContainerEnv path for patching container env
	PatchPathContainerEnv = "/spec/containers/%v/env"
	// PatchPathInitContainerEnv path for patching init container env
	PatchPathInitContainerEnv = "/spec/initContainers/%v/env"
	// PatchPathPodLabel path for patching pod labels
	PatchPathPodLabel = "/metadata/labels"
	// PatchPathAffinity path for patching pod affinity
	PatchPathAffinity = "/spec/affinity"
	// PatchPathPodAnnotations path for patching pod annotations
	PatchPathPodAnnotations = "/metadata/annotations"
	// PatchPathAffinityPatchPath path for patching pod antiAffinity
	PatchPathAffinityPatchPath = "/podAntiAffinity/requiredDuringSchedulingIgnoredDuringExecution"
)
