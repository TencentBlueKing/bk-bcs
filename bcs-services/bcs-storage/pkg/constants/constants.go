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

package constants

const (
	// ServerName 服务名
	ServerName   = "storage.bkbcs.tencent.com"
	ServerV4Name = "storagev4.bkbcs.tencent.com"

	// MicroMetaKeyHTTPPort http port in micro service meta
	MicroMetaKeyHTTPPort = "httpport"
)

const (
	DataTag             = "data"
	ExtraTag            = "extra"
	FieldTag            = "field"
	TypeTag             = "type"
	ClusterIDTag        = "clusterId"
	NamespaceTag        = "namespace"
	MessageTag          = "message"
	SourceTag           = "source"
	ModuleTag           = "module"
	OffsetTag           = "offset"
	LimitTag            = "length"
	TableName           = "alarm"
	CreateTimeTag       = "createTime"
	ReceivedTimeTag     = "receivedTime"
	TimeBeginTag        = "timeBegin"
	TimeEndTag          = "timeEnd"
	ServiceTag          = "service"
	UpdateTimeTag       = "updateTime"
	ResourceTypeTag     = "resourceType"
	ResourceNameTag     = "resourceName"
	IndexNameTag        = "indexName"
	ApplicationTypeName = "application"
	ProcessTypeName     = "process"
	KindTag             = "data.kind"
	EventResourceType   = "Event"
	LabelSelectorTag    = "labelSelector"
	UpdateTimeQueryTag  = "updateTimeBefore"
	EventTimeTag        = "eventTime"
	IpTag               = "ip"
	NameTag             = "name"
)

const (
	TaskGroup          = "taskgroup"
	Application        = "application"
	Deployment         = "deployment"
	Service            = "service"
	ConfigMap          = "configmap"
	Secret             = "secret"
	Endpoint           = "endpoint"
	ExportService      = "exportservice"
	IPPoolStatic       = "IPPoolStatic"
	IPPoolStaticDetail = "IPPoolStaticDetail"
	Pod                = "Pod"
	ReplicaSet         = "ReplicaSet"
	DeploymentK8S      = "Deployment"
	ServiceK8S         = "Service"
	ConfigMapK8S       = "ConfigMap"
	SecretK8S          = "Secret"
	EndpointsK8S       = "Endpoints"
	Ingress            = "Ingress"
	NamespaceK8S       = "Namespace"
	Node               = "Node"
	DaemonSet          = "DaemonSet"
	Job                = "Job"
	StatefulSet        = "StatefulSet"
	ContainerInfo      = "ContainerInfo"
)

const (
	IdTag        = "id"
	EnvTag       = "env"
	LevelTag     = "level"
	ComponentTag = "component"
)
