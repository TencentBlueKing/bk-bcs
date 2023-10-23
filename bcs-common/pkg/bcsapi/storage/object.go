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

package storage

import (
	"time"

	appv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"

	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage/tkex/gamedeployment/v1alpha1"
	gsv1alpha1 "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage/tkex/gamestatefulset/v1alpha1"
	gpav1alpha1 "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage/tkex/generalpodautoscaler/v1alpha1"
	schedtype "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/types"
)

// Namespace is k8s namespace
type Namespace struct {
	CommonDataHeader
	Data *corev1.Namespace
}

// Deployment is k8s deployment
type Deployment struct {
	CommonDataHeader
	Data *appv1.Deployment
}

// DaemonSet is k8s daemonset
type DaemonSet struct {
	CommonDataHeader
	Data *appv1.DaemonSet
}

// StatefulSet is k8s statefulset
type StatefulSet struct {
	CommonDataHeader
	Data *appv1.StatefulSet
}

// GameDeployment is bcs gamedeployment
type GameDeployment struct {
	CommonDataHeader
	Data *gdv1alpha1.GameDeployment
}

// GameStatefulSet is bcs gamestatefulset
type GameStatefulSet struct {
	CommonDataHeader
	Data *gsv1alpha1.GameStatefulSet
}

// MesosApplication is mesos application
type MesosApplication struct {
	CommonDataHeader
	Data *Application
}

// MesosDeployment is mesos deployment
type MesosDeployment struct {
	CommonDataHeader
	Data *schedtype.Deployment
}

// MesosNamespace is mesos namespace
type MesosNamespace string

// K8sNode is k8s node
type K8sNode struct {
	CommonDataHeader
	Data *corev1.Node
}

// Hpa is k8s hpa
type Hpa struct {
	CommonDataHeader
	Data *autoscalingv1.HorizontalPodAutoscaler
}

// Gpa is bcs generalpodautoscaler
type Gpa struct {
	CommonDataHeader
	Data *gpav1alpha1.GeneralPodAutoscaler
}

// Application is mesos application
type Application struct {
	ID              string
	Name            string
	Metadata        commtypes.ObjectMeta
	DefineInstances uint64
	Instance        uint64
	RunningInstance uint64
	BuildedInstance int64
	RunAs           string
	ClusterId       string
	Status          string
	LastStatus      string
	CreateTime      time.Time
	UpdateTime      time.Time
	Mode            string
	LastUpdateTime  time.Time
	ReportTime      time.Time

	// we should replace the next three BcsXXX, using ObjectMeta.Labels directly
	BcsAppID    string
	BcsSetID    string
	BcsModuleID string

	Message string
	Pods    []*commtypes.BcsPodIndex
}

// Pvc is k8s pvc
type Pvc struct {
	CommonDataHeader
	Data *corev1.PersistentVolumeClaim
}

// StorageClass is k8s storageclass
type StorageClass struct { // nolint
	CommonDataHeader
	Data *storagev1.StorageClass
}

// ResourceQuota is k8s resourcequota
type ResourceQuota struct {
	CommonDataHeader
	Data *corev1.ResourceQuota
}

// ReplicaSet is k8s replicaset
type ReplicaSet struct {
	CommonDataHeader
	Data *appv1.ReplicaSet
}
