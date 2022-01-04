/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resource

// k8s 资源类型
const (
	// Deploy Deployment
	Deploy = "deployment"
	// DS DaemonSet
	DS = "daemonset"
	// STS StatefulSet
	STS = "statefulset"
	// CJ CronJob
	CJ = "cronjob"
	// Job
	Job = "job"
	// Pod
	Pod = "pod"

	// Ing Ingress
	Ing = "ingress"
	// SVC Service
	SVC = "service"
	// EP Endpoints
	EP = "endpoints"

	// CM ConfigMap
	CM = "configmap"
	// Secret
	Secret = "secret"

	// PV PersistentVolume
	PV = "persistentvolume"
	// PVC PersistentVolumeClaim
	PVC = "persistentvolumeclaim"
	// SC StorageClass
	SC = "storageclass"

	// SA ServiceAccount
	SA = "serviceaccount"

	// HPA HorizontalPodAutoscaler
	HPA = "horizontalpodautoscaler"

	// CRD CustomResourceDefinition
	CRD = "customresourcedefinition"
	// CObj CustomObject
	CObj = "customobject"
)
