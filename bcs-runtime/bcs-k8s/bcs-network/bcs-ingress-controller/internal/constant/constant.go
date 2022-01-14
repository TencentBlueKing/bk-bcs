/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package constant

const (
	// FinalizerNameBcsIngressController finalizer name of bcs ingress controller
	FinalizerNameBcsIngressController = "ingresscontroller.bkbcs.tencent.com"
	// CloudTencent tencent cloud
	CloudTencent = "tencentcloud"
	// CloudAWS aws cloud
	CloudAWS = "aws"

	// EnvNameIsTCPUDPPortReuse env name for option if the loadbalancer provider support tcp udp port reuse
	// if enabled, we will find protocol info in 4 layer listener name
	EnvNameIsTCPUDPPortReuse = "TCP_UDP_PORT_REUSE"
	// EnvNameIsBulkMode env name for option if use bulk interface for cloud lb
	EnvNameIsBulkMode = "IS_BULK_MODE"

	// DelimiterForLbID delimiter for lb id
	DelimiterForLbID = ":"

	// PortPoolItemStatusError error status of port pool item
	PortPoolItemStatusError = "Error"
	// PortPoolItemStatusInitialize initial status of port pool item
	PortPoolItemStatusInitialize = "Initialize"
	// PortPoolItemStatusReady ready status of port pool item
	PortPoolItemStatusReady = "Ready"
	// PortPoolItemStatusNotReady the status of port pool item is not ready
	PortPoolItemStatusNotReady = "NotReady"
	// PortPoolItemStatusDeleting deleting status of port pool item
	PortPoolItemStatusDeleting = "Deleting"

	// PortPoolItemMessageReady ready message for port pool item
	PortPoolItemMessageReady = "Ready"

	// PortBindingItemStatusInitializing the status of port binding item is initializing
	// means that binding info is not passed to listener yet
	PortBindingItemStatusInitializing = "Initializeing"
	// PortBindingItemStatusNotReady the status of port binding item is not ready
	PortBindingItemStatusNotReady = "NotReady"
	// PortBindingItemStatusReady the status of port binding item is ready
	PortBindingItemStatusReady = "Ready"
	// PortBindingItemStatusDeleting the port binding item is in deleting
	PortBindingItemStatusDeleting = "Deleting"
	// PortBindingItemStatusCleaned the listener of the port binding item is cleaned
	PortBindingItemStatusCleaned = "Cleaned"
	// PortBindingStatusNotReady the status of port binding is not ready
	PortBindingStatusNotReady = "NotReady"
	// PortBindingStatusReady the status of port binding is ready
	PortBindingStatusReady = "Ready"
	// PortBindingStatusCleaning the listener of the port binding is being cleaned
	PortBindingStatusCleaning = "Cleaning"
	// PortBindingStatusCleaned the listener of the port binding is all cleaned
	PortBindingStatusCleaned = "Cleaned"

	// AnnotationForPodStatusReady pod status ready
	AnnotationForPodStatusReady = "Ready"
	// AnnotationForPodStatusNotReady pod status not ready
	AnnotationForPodStatusNotReady = "NotReady"

	// AnnotationForPortPool annotation for claims for port pool
	AnnotationForPortPool = "portpools.networkextension.bkbcs.tencent.com"
	// AnnotationForPortPoolPorts annotation for port pool ports
	AnnotationForPortPoolPorts = "ports.portpools.networkextension.bkbcs.tencent.com"
	// AnnotationForPortPoolBindings annotation for port pool bindings
	AnnotationForPortPoolBindings = "poolbindings.portpool.networkextension.bkbcs.tencent.com"
	// AnnotationForPortPoolBindingStatus annotation for port pool ports binding status
	AnnotationForPortPoolBindingStatus = "status.portpools.networkextension.bkbcs.tencent.com"
	// AnnotationForPortPoolReadinessGate port pool readiness gate
	AnnotationForPortPoolReadinessGate = "readinessgate.portpools.networkextension.bkbcs.tencent.com"

	// ConditionTypeBcsIngressPortBinding readiness gate condition type for port binding of bcs-ingress-controller
	ConditionTypeBcsIngressPortBinding = "networkextension.bkbcs.tencent.com/portbinding-ready"
	// ConditionReasonReadyBcsIngressPortBinding ready reason for port binding condition
	ConditionReasonReadyBcsIngressPortBinding = "Ready"
	// ConditionMessageReadyBcsIngressPortBinding ready message for port binding condition
	ConditionMessageReadyBcsIngressPortBinding = "ports ares binded for the pod"

	// PatchOperationAdd patch add operation
	PatchOperationAdd = "add"
	// PatchOperationReplace patch replace operation
	PatchOperationReplace = "replace"
	// PatchOperationRemove patch remove operation
	PatchOperationRemove = "remove"
	// PatchPathPodAnnotations annotations path for patch operation
	PatchPathPodAnnotations = "/metadata/annotations"
	// PathPathContainerEnv container env path for patch operation
	PatchPathContainerEnv = "/spec/containers/%v/env"
	// PathPathInitContainerEnv init container env path for patch operation
	PathPathInitContainerEnv = "/spec/initContainers/%v/env"
	// PatchPathPodReadinessGate
	PatchPathPodReadinessGate = "/spec/readinessGates"

	// EnvVIPsPrefixForPortPoolPort env prefix for port in port pool
	EnvVIPsPrefixForPortPoolPort = "BCS_PORTPOOL_PORT_VIPLIST_"

	// MaxPortQuantityForEachLoadbalancer max port quantity for each loadbalancer
	MaxPortQuantityForEachLoadbalancer = 4000

	// PortPoolPortProtocolTCP protocol of port in pool is tcp
	PortPoolPortProtocolTCP = "TCP"
	// PortPoolPortProtocolUDP protocol of port in pool is udp
	PortPoolPortProtocolUDP = "UDP"
	// PortPoolPortProtocolTCPUDP protocol of port in pool is tcp&udp
	PortPoolPortProtocolTCPUDP = "TCP_UDP"

	//LoadBalanceCheckFormatWithAp regular expression for check lb format "ap-xxxxx:lb-xxxxx"
	LoadBalanceCheckFormatWithApLbID = "^ap-[A-Za-z0-9]+:lb-[A-Za-z0-9]+"
	//LoadBalanceCheckFormat regular expression for check lb format "lb-xxxxx"
	LoadBalanceCheckFormat = "^lb-[A-Za-z0-9]+"
	//LoadBalanceCheckFormatWithApLbName
	LoadBalanceCheckFormatWithApLbName = "^ap-[A-Za-z0-9]+:[A-Za-z0-9]+"
)
