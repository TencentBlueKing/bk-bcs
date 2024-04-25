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

package formatter

import "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"

// DefaultPruneFunc 默认的 PruneFunc
func DefaultPruneFunc(manifest map[string]interface{}) map[string]interface{} {
	return manifest
}

// CommonPrune 裁剪
func CommonPrune(manifest map[string]interface{}) map[string]interface{} {
	name, _ := mapx.GetItems(manifest, "metadata.name")
	namespace, _ := mapx.GetItems(manifest, "metadata.namespace")
	uid, _ := mapx.GetItems(manifest, "metadata.uid")
	labels, _ := mapx.GetItems(manifest, "metadata.labels")
	annotations, _ := mapx.GetItems(manifest, "metadata.annotations")
	newManifest := map[string]interface{}{
		"apiVersion": mapx.GetStr(manifest, "apiVersion"),
		"kind":       mapx.GetStr(manifest, "kind"),
		"metadata": map[string]interface{}{
			"name": name, "namespace": namespace, "uid": uid, "labels": labels, "annotations": annotations},
	}
	return newManifest
}

// PruneDeploy 裁剪 Deploy
func PruneDeploy(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonPrune(manifest)
	availableReplicas, _ := mapx.GetItems(manifest, "status.availableReplicas")
	readyReplicas, _ := mapx.GetItems(manifest, "status.readyReplicas")
	updatedReplicas, _ := mapx.GetItems(manifest, "status.updatedReplicas")
	replicas, _ := mapx.GetItems(manifest, "spec.replicas")
	strategy, _ := mapx.GetItems(manifest, "spec.strategy")
	ret["status"] = map[string]interface{}{"availableReplicas": availableReplicas, "readyReplicas": readyReplicas,
		"updatedReplicas": updatedReplicas}
	ret["spec"] = map[string]interface{}{"replicas": replicas, "strategy": strategy}
	return ret
}

// PruneSTS 裁剪 StatefulSets
func PruneSTS(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonPrune(manifest)
	readyReplicas, _ := mapx.GetItems(manifest, "status.readyReplicas")
	updatedReplicas, _ := mapx.GetItems(manifest, "status.updatedReplicas")
	replicas, _ := mapx.GetItems(manifest, "spec.replicas")
	updateStrategy, _ := mapx.GetItems(manifest, "spec.updateStrategy")
	podManagementPolicy, _ := mapx.GetItems(manifest, "spec.podManagementPolicy")
	ret["status"] = map[string]interface{}{"readyReplicas": readyReplicas, "updatedReplicas": updatedReplicas}
	ret["spec"] = map[string]interface{}{"replicas": replicas, "updateStrategy": updateStrategy,
		"podManagementPolicy": podManagementPolicy}
	return ret
}

// PruneDS 裁剪 DaemonSets
func PruneDS(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonPrune(manifest)
	desiredNumberScheduled, _ := mapx.GetItems(manifest, "status.desiredNumberScheduled")
	currentNumberScheduled, _ := mapx.GetItems(manifest, "status.currentNumberScheduled")
	numberReady, _ := mapx.GetItems(manifest, "status.numberReady")
	updatedNumberScheduled, _ := mapx.GetItems(manifest, "status.updatedNumberScheduled")
	numberAvailable, _ := mapx.GetItems(manifest, "status.numberAvailable")
	updateStrategy, _ := mapx.GetItems(manifest, "spec.updateStrategy")
	ret["status"] = map[string]interface{}{"desiredNumberScheduled": desiredNumberScheduled,
		"currentNumberScheduled": currentNumberScheduled, "numberReady": numberReady,
		"updatedNumberScheduled": updatedNumberScheduled, "numberAvailable": numberAvailable}
	ret["spec"] = map[string]interface{}{"updateStrategy": updateStrategy}
	return ret
}

// PruneJob 裁剪 Job
func PruneJob(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonPrune(manifest)
	status, _ := mapx.GetItems(manifest, "status")
	completions, _ := mapx.GetItems(manifest, "spec.completions")
	ret["status"] = status
	ret["spec"] = map[string]interface{}{"completions": completions}
	return ret
}

// PruneCJ 裁剪 CronJobs
func PruneCJ(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonPrune(manifest)
	schedule, _ := mapx.GetItems(manifest, "spec.schedule")
	suspend, _ := mapx.GetItems(manifest, "spec.suspend")
	ret["spec"] = map[string]interface{}{"schedule": schedule, "suspend": suspend}
	return ret
}

// PrunePod 裁剪 Pod
func PrunePod(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonPrune(manifest)
	hostIP, _ := mapx.GetItems(manifest, "status.hostIP")
	nodeName, _ := mapx.GetItems(manifest, "spec.nodeName")
	ret["status"] = map[string]interface{}{"hostIP": hostIP}
	ret["spec"] = map[string]interface{}{"nodeName": nodeName}
	return ret
}

// PruneIng 裁剪 Ingresses
func PruneIng(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonPrune(manifest)
	ingressClassName, _ := mapx.GetItems(manifest, "spec.ingressClassName")
	ret["spec"] = map[string]interface{}{"ingressClassName": ingressClassName}
	return ret
}

// PruneSVC 裁剪 Services
func PruneSVC(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonPrune(manifest)
	specType, _ := mapx.GetItems(manifest, "spec.type")
	ret["spec"] = map[string]interface{}{"type": specType}
	return ret
}

// PruneConfig 裁剪 Configmap 和 secret
func PruneConfig(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonPrune(manifest)
	return ret
}

// PrunePV 裁剪 PersistentVolumes
func PrunePV(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonPrune(manifest)
	phase, _ := mapx.GetItems(manifest, "status.phase")
	capacity, _ := mapx.GetItems(manifest, "spec.capacity")
	persistentVolumeReclaimPolicy, _ := mapx.GetItems(manifest, "spec.persistentVolumeReclaimPolicy")
	claimRef, _ := mapx.GetItems(manifest, "spec.claimRef")
	storageClassName, _ := mapx.GetItems(manifest, "spec.storageClassName")
	volumeMode, _ := mapx.GetItems(manifest, "spec.volumeMode")
	ret["status"] = map[string]interface{}{"phase": phase}
	ret["spec"] = map[string]interface{}{"capacity": capacity,
		"persistentVolumeReclaimPolicy": persistentVolumeReclaimPolicy, "claimRef": claimRef,
		"storageClassName": storageClassName, "volumeMode": volumeMode}
	return ret
}

// PrunePVC 裁剪 PersistentVolumeClaims
func PrunePVC(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonPrune(manifest)
	capacity, _ := mapx.GetItems(manifest, "status.capacity")
	phase, _ := mapx.GetItems(manifest, "status.phase")
	volumeName, _ := mapx.GetItems(manifest, "spec.volumeName")
	storageClassName, _ := mapx.GetItems(manifest, "spec.storageClassName")
	volumeMode, _ := mapx.GetItems(manifest, "spec.volumeMode")
	ret["status"] = map[string]interface{}{"capacity": capacity, "phase": phase}
	ret["spec"] = map[string]interface{}{
		"volumeName": volumeName, "storageClassName": storageClassName, "volumeMode": volumeMode}
	return ret
}

// PruneSC 裁剪 StorageClasses
func PruneSC(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonPrune(manifest)
	ret["provisioner"] = mapx.GetStr(manifest, "provisioner")
	ret["reclaimPolicy"] = mapx.GetStr(manifest, "reclaimPolicy")
	ret["volumeBindingMode"] = mapx.GetStr(manifest, "volumeBindingMode")
	return ret
}
