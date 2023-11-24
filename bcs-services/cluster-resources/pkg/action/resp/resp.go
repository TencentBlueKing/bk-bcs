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

// Package resp xxx
package resp

import (
	"context"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/timex"
)

// BuildListAPIResp xxx
func BuildListAPIResp(
	ctx context.Context, params ListParams, opts metav1.ListOptions,
) (*structpb.Struct, error) {
	// NOTE 部分逻辑需要保留 map[string]interface{} 格式以生成 webAnnotations，因此分离出 BuildListAPIRespData
	respData, err := BuildListAPIRespData(ctx, params, opts)
	if err != nil {
		return nil, err
	}
	return pbstruct.Map2pbStruct(respData)
}

// BuildRetrieveAPIResp xxx
func BuildRetrieveAPIResp(
	ctx context.Context, params GetParams, opts metav1.GetOptions,
) (*structpb.Struct, error) {
	respData, err := BuildRetrieveAPIRespData(ctx, params, opts)
	if err != nil {
		return nil, err
	}
	return pbstruct.Map2pbStruct(respData)
}

// BuildCreateAPIResp xxx
func BuildCreateAPIResp(
	ctx context.Context,
	clusterID, resKind, groupVersion string,
	manifest map[string]interface{},
	isNSScoped bool,
	opts metav1.CreateOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConf(clusterID)
	k8sRes, err := res.GetGroupVersionResource(ctx, clusterConf, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	var ret *unstructured.Unstructured
	ret, err = cli.NewResClient(clusterConf, k8sRes).Create(ctx, manifest, isNSScoped, opts)
	if err != nil {
		return nil, err
	}
	return pbstruct.Unstructured2pbStruct(ret), nil
}

// BuildUpdateAPIResp xxx
func BuildUpdateAPIResp(
	ctx context.Context,
	clusterID, resKind, groupVersion, namespace, name string,
	manifest map[string]interface{},
	opts metav1.UpdateOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConf(clusterID)
	k8sRes, err := res.GetGroupVersionResource(ctx, clusterConf, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	var ret *unstructured.Unstructured
	ret, err = cli.NewResClient(clusterConf, k8sRes).Update(ctx, namespace, name, manifest, opts)
	if err != nil {
		return nil, err
	}
	return pbstruct.Unstructured2pbStruct(ret), nil
}

// BuildPatchAPIResp xxx
func BuildPatchAPIResp(
	ctx context.Context,
	clusterID, resKind, groupVersion, namespace, name string,
	pt types.PatchType,
	data []byte,
	opts metav1.PatchOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConf(clusterID)
	k8sRes, err := res.GetGroupVersionResource(ctx, clusterConf, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	var ret *unstructured.Unstructured
	ret, err = cli.NewResClient(clusterConf, k8sRes).Patch(ctx, namespace, name, pt, data, opts)
	if err != nil {
		return nil, err
	}
	return pbstruct.Unstructured2pbStruct(ret), nil
}

// BuildDeleteAPIResp xxx
func BuildDeleteAPIResp(
	ctx context.Context, clusterID, resKind, groupVersion, namespace, name string, opts metav1.DeleteOptions,
) error {
	clusterConf := res.NewClusterConf(clusterID)
	k8sRes, err := res.GetGroupVersionResource(ctx, clusterConf, resKind, groupVersion)
	if err != nil {
		return err
	}
	return cli.NewResClient(clusterConf, k8sRes).Delete(ctx, namespace, name, opts)
}

// BuildListPodRelatedResResp xxx
func BuildListPodRelatedResResp(
	ctx context.Context, clusterID, namespace, podName, format, resKind string,
) (*structpb.Struct, error) {
	relatedRes, err := cli.NewPodCliByClusterID(ctx, clusterID).ListPodRelatedRes(ctx, namespace, podName, resKind)
	if err != nil {
		return nil, err
	}
	respDataBuilder, err := NewRespDataBuilder(
		ctx, DataBuilderParams{Manifest: relatedRes, Kind: resKind, Format: format},
	)
	if err != nil {
		return nil, err
	}
	respData, err := respDataBuilder.BuildList()
	if err != nil {
		return nil, err
	}
	return pbstruct.Map2pbStruct(respData)
}

// BuildListContainerAPIResp xxx
func BuildListContainerAPIResp(ctx context.Context, clusterID, namespace, podName string) (*structpb.ListValue, error) {
	podManifest, err := cli.NewPodCliByClusterID(ctx, clusterID).GetManifest(ctx, namespace, podName)
	if err != nil {
		return nil, err
	}

	containers := []map[string]interface{}{}
	// 获取container statuses
	for _, containerStatus := range mapx.GetList(podManifest, "status.containerStatuses") {
		containers = append(containers, getContainerStatuses(containerStatus, constants.Containers)...)
	}

	// 获取initContainer statuses
	for _, initContainerStatus := range mapx.GetList(podManifest, "status.initContainerStatuses") {
		containers = append(containers, getContainerStatuses(initContainerStatus, constants.InitContainers)...)
	}
	return pbstruct.MapSlice2ListValue(containers)
}

// BuildGetContainerAPIResp xxx
func BuildGetContainerAPIResp(
	ctx context.Context,
	clusterID, namespace, podName, containerName string,
) (*structpb.Struct, error) {
	podManifest, err := cli.NewPodCliByClusterID(ctx, clusterID).GetManifest(ctx, namespace, podName)
	if err != nil {
		return nil, err
	}

	// 遍历查找指定容器的 Spec 及 Status，若其中某项不存在，则抛出错误
	var curContainerSpec, curContainerStatus map[string]interface{}
	for _, csp := range mapx.GetList(podManifest, "spec.containers") {
		spec, _ := csp.(map[string]interface{})
		if containerName == spec["name"].(string) {
			curContainerSpec = spec
		}
	}
	for _, containerStatus := range mapx.GetList(podManifest, "status.containerStatuses") {
		cs, _ := containerStatus.(map[string]interface{})
		if containerName == cs["name"].(string) {
			curContainerStatus = cs
		}
	}
	if len(curContainerSpec) == 0 || len(curContainerStatus) == 0 {
		return nil, errorx.New(errcode.General, "container %s spec or status not found", containerName)
	}

	// 转换时间格式
	startedAt := ""
	state := mapx.GetMap(curContainerStatus, []string{"state"})
	for i := range state {
		if value, ok := state[i].(map[string]interface{}); ok {
			startedAt, _ = timex.NormalizeDatetime(mapx.GetStr(value, "startedAt"))
		}
	}
	// 转换时间格式lastState格式，lastState有好几种状态但是是单一的，无法确定key，只能遍历
	lastState := mapx.GetMap(curContainerStatus, "lastState")
	for i := range lastState {
		if value, ok := lastState[i].(map[string]interface{}); ok {
			lastStateStartedAt, _ := timex.NormalizeDatetime(mapx.GetStr(value, "startedAt"))
			lastStateFinishedAt, _ := timex.NormalizeDatetime(mapx.GetStr(value, "finishedAt"))
			// 有才赋值转换时间格式，没有直接原样返回
			if lastStateStartedAt != "" {
				lastState[i].(map[string]interface{})["startedAt"] = lastStateStartedAt
			}
			if lastStateFinishedAt != "" {
				lastState[i].(map[string]interface{})["finishedAt"] = lastStateFinishedAt
			}
		}
	}

	// 各项容器数据组装
	containerInfo := map[string]interface{}{
		"hostName":      mapx.Get(podManifest, "spec.nodeName", "N/A"),
		"hostIP":        mapx.Get(podManifest, "status.hostIP", "N/A"),
		"containerIP":   mapx.Get(podManifest, "status.podIP", "N/A"),
		"containerID":   extractContainerID(mapx.GetStr(curContainerStatus, "containerID")),
		"containerName": containerName,
		"image":         mapx.Get(curContainerStatus, "image", "N/A"),
		"networkMode":   mapx.Get(podManifest, "spec.dnsPolicy", "N/A"),
		"ports":         mapx.GetList(curContainerSpec, "ports"),
		"startedAt":     startedAt,
		"restartCnt":    mapx.GetInt64(curContainerStatus, "restartCount"),
		"lastState":     lastState,
		"resources":     mapx.Get(curContainerSpec, "resources", map[string]interface{}{}),
		"command": map[string]interface{}{
			"command": mapx.GetList(curContainerSpec, "command"),
			"args":    mapx.GetList(curContainerSpec, "args"),
		},
	}
	mounts := mapx.GetList(curContainerSpec, "volumeMounts")
	volumes := []map[string]interface{}{}
	for _, mount := range mounts {
		m, _ := mount.(map[string]interface{})
		volumes = append(volumes, map[string]interface{}{
			"name":      mapx.Get(m, "name", "N/A"),
			"mountPath": mapx.Get(m, "mountPath", "N/A"),
			"readonly":  mapx.Get(m, "readOnly", "N/A"),
		})
	}
	containerInfo["volumes"] = volumes

	return pbstruct.Map2pbStruct(containerInfo)
}

// BuildUpdateCObjAPIResp 构建更新自定义资源请求响应结果
func BuildUpdateCObjAPIResp(
	ctx context.Context,
	clusterID, resKind, groupVersion, namespace, name string,
	manifest map[string]interface{},
	opts metav1.UpdateOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConf(clusterID)
	cobjRes, err := res.GetGroupVersionResource(ctx, clusterConf, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	// CustomObject 需要自行更新到最新的 ResourceVersion，否则会更新失败
	cobjManifest, err := cli.GetCObjManifest(ctx, clusterConf, cobjRes, namespace, name)
	if err != nil {
		return nil, err
	}
	latestRV, err := mapx.GetItems(cobjManifest, "metadata.resourceVersion")
	if err != nil {
		return nil, err
	}
	err = mapx.SetItems(manifest, "metadata.resourceVersion", latestRV)
	if err != nil {
		return nil, err
	}

	// 下发更新指令到集群
	var ret *unstructured.Unstructured
	ret, err = cli.NewResClient(clusterConf, cobjRes).Update(ctx, namespace, name, manifest, opts)
	if err != nil {
		return nil, err
	}
	return pbstruct.Unstructured2pbStruct(ret), nil
}

// extractContainerID 去除容器 ID 前缀，原格式：docker://[a-zA-Z0-9]{64}
func extractContainerID(rawContainerID string) string {
	return strings.Replace(rawContainerID, "docker://", "", 1)
}

// getContainerStatuses 获取container statuses 想关参数
func getContainerStatuses(containerStatus interface{}, containerType string) (containers []map[string]interface{}) {
	cs, ok := containerStatus.(map[string]interface{})
	// 取不出则退出
	if !ok {
		return
	}
	status, reason, message, startedAt := "", "", "", ""
	// state 有且只有一对键值：running / terminated / waiting
	// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#containerstate-v1-core
	for k := range cs["state"].(map[string]interface{}) {
		status = k
		reason, _ = mapx.Get(cs, []string{"state", k, "reason"}, k).(string)
		message, _ = mapx.Get(cs, []string{"state", k, "message"}, k).(string)
		startedTime, _ := mapx.Get(cs, []string{"state", k, "startedAt"}, k).(string)
		startedAt, _ = timex.NormalizeDatetime(startedTime)
	}
	containers = append(containers, map[string]interface{}{
		"containerID":    extractContainerID(mapx.GetStr(cs, "containerID")),
		"image":          cs["image"].(string),
		"name":           cs["name"].(string),
		"container_type": containerType,
		"status":         status,
		"reason":         reason,
		"message":        message,
		"restartCnt":     mapx.GetInt64(cs, "restartCount"),
		"startedAt":      startedAt,
	})
	return containers
}
