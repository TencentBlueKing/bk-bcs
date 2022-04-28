/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/cmanager"
	"time"

	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	DataJobQueue = "dataJob"
)

// extract type
const (
	ProjectType   = "project"
	ClusterType   = "cluster"
	NamespaceType = "namespace"
	WorkloadType  = "workload"
	PublicType    = "public"
)

// extract dimension
const (
	DimensionMinute = "minute"
	DimensionHour   = "hour"
	DimensionDay    = "day"
)

// db table name
const (
	DataTableNamePrefix = "bcsdatamanager_"
	ClusterTableName    = "cluster"
	ProjectTableName    = "project"
	NamespaceTableName  = "namespace"
	WorkloadTableName   = "workload"
	PublicTableName     = "public"
)

// cluster type
const (
	Kubernetes = "k8s"
	Mesos      = "mesos"
)

const (
	DeploymentType       = "deployment"
	StatefulSetType      = "statefulSet"
	DaemonSetType        = "daemonSet"
	GameDeploymentType   = "gameDeployment"
	GameStatefulSetType  = "gameStatefulSet"
	MesosApplicationType = "application"
	MesosDeployment      = "deployment"
)

const (
	// ServiceDomain domain name for service
	ServiceDomain = "datamanager.bkbcs.tencent.com"
	// MicroMetaKeyHTTPPort http port in micro service meta
	MicroMetaKeyHTTPPort = "httpport"
)

const (
	MonthTimeFormat  = "2006-01"
	DayTimeFormat    = "2006-01-02"
	HourTimeFormat   = "2006-01-02 15:00:00"
	MinuteTimeFormat = "2006-01-02 15:04:00"
	SecondTimeFormat = "2006-01-02 15:04:05"
)

// ProjectMeta meta for project
type ProjectMeta struct {
	ProjectID string `json:"projectID"`
}

// ClusterMeta meta for cluster
type ClusterMeta struct {
	ProjectID   string `json:"projectID"`
	ClusterID   string `json:"clusterID"`
	ClusterType string `json:"clusterType"`
}

// NamespaceMeta meta for namespace
type NamespaceMeta struct {
	ProjectID   string `json:"projectID"`
	ClusterID   string `json:"clusterID"`
	ClusterType string `json:"clusterType"`
	Name        string `json:"name"`
}

// WorkloadMeta meta for workload
type WorkloadMeta struct {
	ProjectID    string `json:"projectID"`
	ClusterID    string `json:"clusterID"`
	ClusterType  string `json:"clusterType"`
	Namespace    string `json:"namespace"`
	ResourceType string `json:"resourceType"`
	Name         string `json:"name"`
}

// PublicData for public table
type PublicData struct {
	CreateTime   primitive.DateTime `json:"create_time" bson:"create_time"`
	UpdateTime   primitive.DateTime `json:"update_time" bson:"update_time"`
	ObjectType   string             `json:"object_type" bson:"object_type"`
	ProjectID    string             `json:"projectID" bson:"project_id"`
	ClusterID    string             `json:"clusterID" bson:"cluster_id"`
	ClusterType  string             `json:"clusterType" bson:"cluster_type"`
	Namespace    string             `json:"namespace" bson:"namespace"`
	WorkloadType string             `json:"workloadType" bson:"workload_type"`
	WorkloadName string             `json:"workloadName" bson:"workload_name"`
	Metrics      interface{}        `json:"metrics" bson:"metrics"`
}

// WorkloadData for workload table
type WorkloadData struct {
	CreateTime         primitive.DateTime             `json:"createTime" bson:"create_time"`
	UpdateTime         primitive.DateTime             `json:"updateTime" bson:"update_time"`
	BucketTime         string                         `json:"bucketTime" bson:"bucket_time"`
	Dimension          string                         `json:"dimension" bson:"dimension"`
	ProjectID          string                         `json:"projectID" bson:"project_id"`
	ClusterID          string                         `json:"clusterID" bson:"cluster_id"`
	ClusterType        string                         `json:"clusterType" bson:"cluster_type"`
	Namespace          string                         `json:"namespace" bson:"namespace"`
	WorkloadType       string                         `json:"workloadType" bson:"workload_type"`
	Name               string                         `json:"workloadName" bson:"workload_name"`
	MaxCPUUsageTime    *bcsdatamanager.ExtremumRecord `json:"maxCPUUsageTime"  bson:"max_cpu_usage_time"`
	MinCPUUsageTime    *bcsdatamanager.ExtremumRecord `json:"minCPUUsageTime" bson:"min_cpu_usage_time"`
	MaxMemoryUsageTime *bcsdatamanager.ExtremumRecord `json:"maxMemoryUsageTime" bson:"max_memory_usage_time"`
	MinMemoryUsageTime *bcsdatamanager.ExtremumRecord `json:"minMemoryUsageTime" bson:"min_memory_usage_time"`
	MaxCPUTime         *bcsdatamanager.ExtremumRecord `json:"maxCPUTime"  bson:"max_cpu_time"`
	MinCPUTime         *bcsdatamanager.ExtremumRecord `json:"minCPUTime" bson:"min_cpu_time"`
	MaxMemoryTime      *bcsdatamanager.ExtremumRecord `json:"maxMemoryTime" bson:"max_memory_time"`
	MinMemoryTime      *bcsdatamanager.ExtremumRecord `json:"minMemoryTime" bson:"min_memory_time"`
	MinInstanceTime    *bcsdatamanager.ExtremumRecord `json:"minInstanceTime" bson:"min_instance_time"`
	MaxInstanceTime    *bcsdatamanager.ExtremumRecord `json:"maxInstanceTime" bson:"max_instance_time"`
	Metrics            []*WorkloadMetrics             `json:"metrics" bson:"metrics"`
}

// ProjectData for project table
type ProjectData struct {
	CreateTime primitive.DateTime             `json:"createTime" bson:"create_time"`
	UpdateTime primitive.DateTime             `json:"updateTime" bson:"update_time"`
	BucketTime string                         `json:"bucketTime" bson:"bucket_time"`
	Dimension  string                         `json:"dimension" bson:"dimension"`
	ProjectID  string                         `json:"projectID" bson:"project_id"`
	MinNode    *bcsdatamanager.ExtremumRecord `json:"minNode,omitempty" bson:"min_node"`
	MaxNode    *bcsdatamanager.ExtremumRecord `json:"maxNode,omitempty" bson:"max_node"`
	Metrics    []*ProjectMetrics              `json:"metrics" bson:"metrics"`
}

// NamespaceData for namespace table
type NamespaceData struct {
	CreateTime         primitive.DateTime             `json:"createTime" bson:"create_time"`
	UpdateTime         primitive.DateTime             `json:"updateTime" bson:"update_time"`
	BucketTime         string                         `json:"bucketTime" bson:"bucket_time"`
	Dimension          string                         `json:"dimension" bson:"dimension"`
	ProjectID          string                         `json:"projectID" bson:"project_id"`
	ClusterID          string                         `json:"clusterID" bson:"cluster_id"`
	ClusterType        string                         `json:"clusterType" bson:"cluster_type"`
	Namespace          string                         `json:"namespace" bson:"namespace"`
	MaxCPUUsageTime    *bcsdatamanager.ExtremumRecord `json:"maxCPUUsageTime" bson:"max_cpu_usage_time"`
	MinCPUUsageTime    *bcsdatamanager.ExtremumRecord `json:"minCPUUsageTime" bson:"min_cpu_usage_time"`
	MaxMemoryUsageTime *bcsdatamanager.ExtremumRecord `json:"maxMemoryUsageTime" bson:"max_memory_usage_time"`
	MinMemoryUsageTime *bcsdatamanager.ExtremumRecord `json:"minMemoryUsageTime" bson:"min_memory_usage_time"`
	MinInstanceTime    *bcsdatamanager.ExtremumRecord `json:"minInstanceTime" bson:"min_instance_time"`
	MaxInstanceTime    *bcsdatamanager.ExtremumRecord `json:"maxInstanceTime" bson:"max_instance_time"`
	MinWorkloadUsage   *bcsdatamanager.ExtremumRecord `json:"minWorkloadUsage" bson:"min_workload_usage"`
	MaxWorkloadUsage   *bcsdatamanager.ExtremumRecord `json:"maxWorkloadUsage" bson:"max_workload_usage"`
	Metrics            []*NamespaceMetrics            `json:"metrics" bson:"metrics"`
}

// ClusterData for cluster table
type ClusterData struct {
	CreateTime  primitive.DateTime             `json:"createTime" bson:"create_time"`
	UpdateTime  primitive.DateTime             `json:"updateTime" bson:"update_time"`
	BucketTime  string                         `json:"bucketTime" bson:"bucket_time"`
	Dimension   string                         `json:"dimension" bson:"dimension"`
	ProjectID   string                         `json:"projectID" bson:"project_id"`
	ClusterID   string                         `json:"clusterID" bson:"cluster_id"`
	ClusterType string                         `json:"clusterType" bson:"cluster_type"`
	MinNode     *bcsdatamanager.ExtremumRecord `json:"minNode,omitempty" bson:"min_node"`
	MaxNode     *bcsdatamanager.ExtremumRecord `json:"maxNode,omitempty" bson:"max_node"`
	MinInstance *bcsdatamanager.ExtremumRecord `json:"minInstance,omitempty" bson:"min_instance"`
	MaxInstance *bcsdatamanager.ExtremumRecord `json:"maxInstance,omitempty" bson:"max_instance"`
	Metrics     []*ClusterMetrics              `json:"metrics" bson:"metrics"`
}

// ProjectPublicMetrics public for project
type ProjectPublicMetrics struct {
	ClusterCount int64 `json:"cluster_count"`
}

// ClusterPublicMetrics public for cluster
type ClusterPublicMetrics struct {
}

// NamespacePublicMetrics public for namespace
type NamespacePublicMetrics struct {
	ResourceLimit *bcsdatamanager.ResourceLimit `json:"resource_limit"`
	SuggestCPU    float64                       `json:"suggest_cpu"`
	SuggestMemory float64                       `json:"suggest_memory"`
}

// WorkloadPublicMetrics public for workload
type WorkloadPublicMetrics struct {
	SuggestCPU    float64 `json:"suggest_cpu"`
	SuggestMemory float64 `json:"suggest_memory"`
}

// ProjectMetrics project metric
type ProjectMetrics struct {
	Index              int                            `json:"index"`
	Time               primitive.DateTime             `json:"time,omitempty"`
	ClustersCount      int64                          `json:"clustersCount,omitempty"`
	TotalCPU           float64                        `json:"totalCPU,omitempty"`
	TotalMemory        int64                          `json:"totalMemory,omitempty"`
	TotalLoadCPU       float64                        `json:"totalLoadCPU,omitempty"`
	TotalLoadMemory    int64                          `json:"totalLoadMemory,omitempty"`
	AvgLoadCPU         float64                        `json:"avgLoadCPU,omitempty"`
	AvgLoadMemory      int64                          `json:"avgLoadMemory,omitempty"`
	CPUUsage           float64                        `json:"CPUUsage,omitempty"`
	MemoryUsage        float64                        `json:"MemoryUsage,omitempty"`
	NodeCount          int64                          `json:"nodeCount,omitempty"`
	AvailableNodeCount int64                          `json:"availableNodeCount,omitempty"`
	MinNode            *bcsdatamanager.ExtremumRecord `json:"minNode,omitempty"`
	MaxNode            *bcsdatamanager.ExtremumRecord `json:"maxNode,omitempty"`
}

// ClusterMetrics cluster metric
type ClusterMetrics struct {
	Index              int                            `json:"index"`
	Time               primitive.DateTime             `json:"time,omitempty"`
	NodeCount          int64                          `json:"nodeCount,omitempty"`
	AvailableNodeCount int64                          `json:"availableNodeCount,omitempty"`
	MinNode            *bcsdatamanager.ExtremumRecord `json:"minNode,omitempty"`
	MaxNode            *bcsdatamanager.ExtremumRecord `json:"maxNode,omitempty"`
	NodeQuantile       []*bcsdatamanager.NodeQuantile `json:"nodeQuantile,omitempty"`
	MinUsageNode       string                         `json:"minUsageNode,omitempty"`
	TotalCPU           float64                        `json:"totalCPU,omitempty"`
	TotalMemory        int64                          `json:"totalMemory,omitempty"`
	TotalLoadCPU       float64                        `json:"totalLoadCPU,omitempty"`
	TotalLoadMemory    int64                          `json:"totalLoadMemory,omitempty"`
	AvgLoadCPU         float64                        `json:"avgLoadCPU,omitempty"`
	AvgLoadMemory      int64                          `json:"avgLoadMemory,omitempty"`
	CPUUsage           float64                        `json:"CPUUsage,omitempty"`
	MemoryUsage        float64                        `json:"MemoryUsage,omitempty"`
	WorkloadCount      int64                          `json:"workloadCount,omitempty"`
	InstanceCount      int64                          `json:"instanceCount,omitempty"`
	MinInstance        *bcsdatamanager.ExtremumRecord `json:"minInstance,omitempty"`
	MaxInstance        *bcsdatamanager.ExtremumRecord `json:"maxInstance,omitempty"`
	CpuRequest         float64                        `json:"cpuRequest,omitempty"`
	MemoryRequest      int64                          `json:"memoryRequest,omitempty"`
}

// NamespaceMetrics namespace metric
type NamespaceMetrics struct {
	Index              int                            `json:"index"`
	Time               primitive.DateTime             `json:"time"`
	CPURequest         float64                        `json:"CPURequest"`
	MemoryRequest      int64                          `json:"memoryRequest"`
	CPUUsageAmount     float64                        `json:"CPUUsageAmount"`
	MemoryUsageAmount  int64                          `json:"memoryUsageAmount"`
	CPUUsage           float64                        `json:"CPUUsage"`
	MemoryUsage        float64                        `json:"MemoryUsage"`
	WorkloadCount      int64                          `json:"workloadCount"`
	InstanceCount      int64                          `json:"instanceCount"`
	MaxCPUUsageTime    *bcsdatamanager.ExtremumRecord `json:"maxCPUUsageTime"`
	MinCPUUsageTime    *bcsdatamanager.ExtremumRecord `json:"minCPUUsageTime"`
	MaxMemoryUsageTime *bcsdatamanager.ExtremumRecord `json:"maxMemoryUsageTime"`
	MinMemoryUsageTime *bcsdatamanager.ExtremumRecord `json:"minMemoryUsageTime"`
	MinInstanceTime    *bcsdatamanager.ExtremumRecord `json:"minInstanceTime"`
	MaxInstanceTime    *bcsdatamanager.ExtremumRecord `json:"maxInstanceTime"`
	MinWorkloadUsage   *bcsdatamanager.ExtremumRecord `json:"minWorkloadUsage"`
	MaxWorkloadUsage   *bcsdatamanager.ExtremumRecord `json:"maxWorkloadUsage"`
}

// WorkloadMetrics workload metric
type WorkloadMetrics struct {
	Index              int                            `json:"index"`
	Time               primitive.DateTime             `json:"time"`
	CPURequest         float64                        `json:"CPURequest"`
	MemoryRequest      int64                          `json:"memoryRequest"`
	CPUUsageAmount     float64                        `json:"CPUUsageAmount"`
	MemoryUsageAmount  int64                          `json:"memoryUsageAmount"`
	CPUUsage           float64                        `json:"CPUUsage"`
	MemoryUsage        float64                        `json:"MemoryUsage"`
	InstanceCount      int64                          `json:"instanceCount"`
	MaxCPUUsageTime    *bcsdatamanager.ExtremumRecord `json:"maxCPUUsageTime"`
	MinCPUUsageTime    *bcsdatamanager.ExtremumRecord `json:"minCPUUsageTime"`
	MaxMemoryUsageTime *bcsdatamanager.ExtremumRecord `json:"maxMemoryUsageTime"`
	MinMemoryUsageTime *bcsdatamanager.ExtremumRecord `json:"minMemoryUsageTime"`
	MinInstanceTime    *bcsdatamanager.ExtremumRecord `json:"minInstanceTime"`
	MaxInstanceTime    *bcsdatamanager.ExtremumRecord `json:"maxInstanceTime"`
	MaxCPUTime         *bcsdatamanager.ExtremumRecord `json:"maxCPUTime" `
	MinCPUTime         *bcsdatamanager.ExtremumRecord `json:"minCPUTime" `
	MaxMemoryTime      *bcsdatamanager.ExtremumRecord `json:"maxMemoryTime"`
	MinMemoryTime      *bcsdatamanager.ExtremumRecord `json:"minMemoryTime" `
}

// JobCommonOpts data job common opts
type JobCommonOpts struct {
	ObjectType   string
	ProjectID    string
	ClusterID    string
	ClusterType  string
	Namespace    string
	WorkloadType string
	Name         string
	Dimension    string
	CurrentTime  time.Time
}

// Clients clients for dataJob
type Clients struct {
	MonitorClient   bcsmonitor.ClientInterface
	K8sStorageCli   bcsapi.Storage
	MesosStorageCli bcsapi.Storage
	CmCli           *cmanager.ClusterManagerClientWithHeader
}

// NewClients init dataJob clients
func NewClients(monitorClient bcsmonitor.ClientInterface, k8sStorageCli, mesosStorageCli bcsapi.Storage,
	cmCli *cmanager.ClusterManagerClientWithHeader) *Clients {
	return &Clients{MonitorClient: monitorClient, CmCli: cmCli,
		K8sStorageCli: k8sStorageCli, MesosStorageCli: mesosStorageCli}
}
