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

package resp

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
)

// BuildListAPIResp ...
func BuildListAPIResp(
	clusterID, resKind, groupVersion, namespace string, opts metav1.ListOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GetGroupVersionResource(clusterConf, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	var ret *unstructured.UnstructuredList
	ret, err = cli.NewResClient(clusterConf, k8sRes).List(namespace, opts)
	if err != nil {
		return nil, err
	}

	return genListResRespData(ret.UnstructuredContent(), resKind)
}

// BuildRetrieveAPIResp ...
func BuildRetrieveAPIResp(
	clusterID, resKind, groupVersion, namespace, name string, opts metav1.GetOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GetGroupVersionResource(clusterConf, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	var ret *unstructured.Unstructured
	ret, err = cli.NewResClient(clusterConf, k8sRes).Get(namespace, name, opts)
	if err != nil {
		return nil, err
	}

	manifest := ret.UnstructuredContent()
	respData := map[string]interface{}{
		"manifest": manifest, "manifestExt": formatter.GetFormatFunc(resKind)(manifest),
	}
	return pbstruct.Map2pbStruct(respData)
}

// BuildCreateAPIResp ...
func BuildCreateAPIResp(
	clusterID, resKind, groupVersion string, manifest *structpb.Struct, isNamespaceScoped bool, opts metav1.CreateOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GetGroupVersionResource(clusterConf, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	var ret *unstructured.Unstructured
	ret, err = cli.NewResClient(clusterConf, k8sRes).Create(manifest.AsMap(), isNamespaceScoped, opts)
	if err != nil {
		return nil, err
	}
	return pbstruct.Unstructured2pbStruct(ret), nil
}

// BuildUpdateAPIResp ...
func BuildUpdateAPIResp(
	clusterID, resKind, groupVersion, namespace, name string, manifest *structpb.Struct, opts metav1.UpdateOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GetGroupVersionResource(clusterConf, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	var ret *unstructured.Unstructured
	ret, err = cli.NewResClient(clusterConf, k8sRes).Update(namespace, name, manifest.AsMap(), opts)
	if err != nil {
		return nil, err
	}
	return pbstruct.Unstructured2pbStruct(ret), nil
}

// BuildDeleteAPIResp ...
func BuildDeleteAPIResp(
	clusterID, resKind, groupVersion, namespace, name string, opts metav1.DeleteOptions,
) error {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GetGroupVersionResource(clusterConf, resKind, groupVersion)
	if err != nil {
		return err
	}
	return cli.NewResClient(clusterConf, k8sRes).Delete(namespace, name, opts)
}

// BuildPodListAPIResp ...
func BuildPodListAPIResp(
	clusterID, namespace, ownerKind, ownerName string, opts metav1.ListOptions,
) (*structpb.Struct, error) {
	// 获取指定命名空间下的所有符合条件的 Pod
	ret, err := cli.NewPodCliByClusterID(clusterID).List(namespace, ownerKind, ownerName, opts)
	if err != nil {
		return nil, err
	}
	return genListResRespData(ret, res.Po)
}

// BuildListPodRelatedResResp ...
func BuildListPodRelatedResResp(clusterID, namespace, podName, resKind string) (*structpb.Struct, error) {
	relatedRes, err := cli.NewPodCliByClusterID(clusterID).ListPodRelatedRes(namespace, podName, resKind)
	if err != nil {
		return nil, err
	}
	return genListResRespData(relatedRes, resKind)
}

// 根据 ResList Manifest 生成获取某类资源列表的响应结果
func genListResRespData(manifest map[string]interface{}, resKind string) (*structpb.Struct, error) {
	manifestExt := map[string]interface{}{}
	formatFunc := formatter.GetFormatFunc(resKind)
	// 遍历列表中的每个资源，生成 manifestExt
	for _, item := range manifest["items"].([]interface{}) {
		uid, _ := mapx.GetItems(item.(map[string]interface{}), "metadata.uid")
		manifestExt[uid.(string)] = formatFunc(item.(map[string]interface{}))
	}

	// 组装数据，并转换为 structpb.Struct 格式
	respData := map[string]interface{}{"manifest": manifest, "manifestExt": manifestExt}
	return pbstruct.Map2pbStruct(respData)
}

// BuildListContainerAPIResp ...
func BuildListContainerAPIResp(clusterID, namespace, podName string) (*structpb.ListValue, error) {
	podManifest, err := cli.NewPodCliByClusterID(clusterID).GetManifest(namespace, podName)
	if err != nil {
		return nil, err
	}

	containers := []map[string]interface{}{}
	containerStatuses, _ := mapx.GetItems(podManifest, "status.containerStatuses")
	for _, cs := range containerStatuses.([]interface{}) {
		cs, _ := cs.(map[string]interface{})
		status, reason, message := "", "", ""
		// state 有且只有一对键值：running / terminated / waiting
		// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#containerstate-v1-core
		for k := range cs["state"].(map[string]interface{}) {
			status = k
			reason, _ = mapx.GetWithDefault(cs, []string{"state", k, "reason"}, k).(string)
			message, _ = mapx.GetWithDefault(cs, []string{"state", k, "message"}, k).(string)
		}
		containers = append(containers, map[string]interface{}{
			"containerID": extractContainerID(mapx.GetWithDefault(cs, "containerID", "").(string)),
			"image":       cs["image"].(string),
			"name":        cs["name"].(string),
			"status":      status,
			"reason":      reason,
			"message":     message,
		})
	}
	return pbstruct.MapSlice2ListValue(containers)
}

// BuildGetContainerAPIResp ...
func BuildGetContainerAPIResp(clusterID, namespace, podName, containerName string) (*structpb.Struct, error) {
	podManifest, err := cli.NewPodCliByClusterID(clusterID).GetManifest(namespace, podName)
	if err != nil {
		return nil, err
	}

	// 遍历查找指定容器的 Spec 及 Status，若其中某项不存在，则抛出错误
	var curContainerSpec, curContainerStatus map[string]interface{}
	containerSpec, _ := mapx.GetItems(podManifest, "spec.containers")
	for _, csp := range containerSpec.([]interface{}) {
		csp, _ := csp.(map[string]interface{})
		if containerName == csp["name"].(string) {
			curContainerSpec = csp
		}
	}
	containerStatuses, _ := mapx.GetItems(podManifest, "status.containerStatuses")
	for _, cs := range containerStatuses.([]interface{}) {
		cs, _ := cs.(map[string]interface{})
		if containerName == cs["name"].(string) {
			curContainerStatus = cs
		}
	}
	if len(curContainerSpec) == 0 || len(curContainerStatus) == 0 {
		return nil, fmt.Errorf("container %s spec or status not found", containerName)
	}

	// 各项容器数据组装
	containerInfo := map[string]interface{}{
		"hostName":      mapx.GetWithDefault(podManifest, "spec.nodeName", "--"),
		"hostIP":        mapx.GetWithDefault(podManifest, "status.hostIP", "--"),
		"containerIP":   mapx.GetWithDefault(podManifest, "status.podIP", "--"),
		"containerID":   extractContainerID(mapx.GetWithDefault(curContainerStatus, "containerID", "").(string)),
		"containerName": containerName,
		"image":         mapx.GetWithDefault(curContainerStatus, "image", "--"),
		"networkMode":   mapx.GetWithDefault(podManifest, "spec.dnsPolicy", "--"),
		"ports":         mapx.GetWithDefault(curContainerSpec, "ports", []interface{}{}),
		"volumes":       []map[string]interface{}{},
		"resources":     mapx.GetWithDefault(curContainerSpec, "resources", map[string]interface{}{}),
		"command": map[string]interface{}{
			"command": mapx.GetWithDefault(curContainerSpec, "command", []string{}),
			"args":    mapx.GetWithDefault(curContainerSpec, "args", []string{}),
		},
	}
	mounts := mapx.GetWithDefault(curContainerSpec, "volumeMounts", []map[string]interface{}{})
	for _, m := range mounts.([]interface{}) {
		m, _ := m.(map[string]interface{})
		containerInfo["volumes"] = append(containerInfo["volumes"].([]map[string]interface{}), map[string]interface{}{
			"name":      mapx.GetWithDefault(m, "name", "--"),
			"mountPath": mapx.GetWithDefault(m, "mountPath", "--"),
			"readonly":  mapx.GetWithDefault(m, "readOnly", "--"),
		})
	}

	return pbstruct.Map2pbStruct(containerInfo)
}

// BuildUpdateCObjAPIResp 构建更新自定义资源请求响应结果
func BuildUpdateCObjAPIResp(
	clusterID, resKind, groupVersion, namespace, name string, manifest *structpb.Struct, opts metav1.UpdateOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	cobjRes, err := res.GetGroupVersionResource(clusterConf, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	// CustomObject 需要自行更新到最新的 ResourceVersion，否则会更新失败
	cobjManifest, err := cli.GetCObjManifest(clusterConf, cobjRes, namespace, name)
	if err != nil {
		return nil, err
	}
	latestRV, err := mapx.GetItems(cobjManifest, "metadata.resourceVersion")
	if err != nil {
		return nil, err
	}
	newCObjManifest := manifest.AsMap()
	err = mapx.SetItems(newCObjManifest, "metadata.resourceVersion", latestRV)
	if err != nil {
		return nil, err
	}

	// 下发更新指令到集群
	var ret *unstructured.Unstructured
	ret, err = cli.NewResClient(clusterConf, cobjRes).Update(namespace, name, newCObjManifest, opts)
	if err != nil {
		return nil, err
	}
	return pbstruct.Unstructured2pbStruct(ret), nil
}

// 去除容器 ID 前缀，原格式：docker://[a-zA-Z0-9]{64}
func extractContainerID(rawContainerID string) string {
	return strings.Replace(rawContainerID, "docker://", "", 1)
}
