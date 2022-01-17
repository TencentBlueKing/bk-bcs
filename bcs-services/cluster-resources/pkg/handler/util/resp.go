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

package util

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
)

func BuildListApiResp(
	clusterID, resKind, groupVersion, namespace string, opts metav1.ListOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GetGroupVersionResource(clusterConf, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	var ret *unstructured.UnstructuredList
	if namespace != "" {
		ret, err = cli.NewNsScopedResClient(clusterConf, k8sRes).List(namespace, opts)
	} else {
		// TODO 支持集群域资源
		panic("cluster scoped resource unsupported")
	}
	if err != nil {
		return nil, err
	}

	return genListResRespData(ret.UnstructuredContent(), resKind)
}

func BuildRetrieveApiResp(
	clusterID, resKind, groupVersion, namespace, name string, opts metav1.GetOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GetGroupVersionResource(clusterConf, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	var ret *unstructured.Unstructured
	if namespace != "" {
		ret, err = cli.NewNsScopedResClient(clusterConf, k8sRes).Get(namespace, name, opts)
	} else {
		// TODO 支持集群域资源
		panic("cluster scoped resource unsupported")
	}
	if err != nil {
		return nil, err
	}

	manifest := ret.UnstructuredContent()
	respData := map[string]interface{}{
		"manifest": manifest, "manifestExt": formatter.Kind2FormatFuncMap[resKind](manifest),
	}
	return util.Map2pbStruct(respData)
}

func BuildCreateApiResp(
	clusterID, resKind, groupVersion string, manifest *structpb.Struct, isNamespaceScoped bool, opts metav1.CreateOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GetGroupVersionResource(clusterConf, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	var ret *unstructured.Unstructured
	if isNamespaceScoped {
		ret, err = cli.NewNsScopedResClient(clusterConf, k8sRes).Create(manifest.AsMap(), opts)
	} else {
		// TODO 支持集群域资源
		panic("cluster scoped resource unsupported")
	}
	if err != nil {
		return nil, err
	}
	return util.Unstructured2pbStruct(ret), nil
}

func BuildUpdateApiResp(
	clusterID, resKind, groupVersion, namespace, name string, manifest *structpb.Struct, opts metav1.UpdateOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GetGroupVersionResource(clusterConf, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	var ret *unstructured.Unstructured
	if namespace != "" {
		ret, err = cli.NewNsScopedResClient(clusterConf, k8sRes).Update(namespace, name, manifest.AsMap(), opts)
	} else {
		// TODO 支持集群域资源
		panic("cluster scoped resource unsupported")
	}
	if err != nil {
		return nil, err
	}
	return util.Unstructured2pbStruct(ret), nil
}

func BuildDeleteApiResp(
	clusterID, resKind, groupVersion, namespace, name string, opts metav1.DeleteOptions,
) error {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GetGroupVersionResource(clusterConf, resKind, groupVersion)
	if err != nil {
		return err
	}
	if namespace != "" {
		return cli.NewNsScopedResClient(clusterConf, k8sRes).Delete(namespace, name, opts)
	}
	// TODO 支持集群域资源
	panic("cluster scoped resource unsupported")
}

func BuildPodListApiResp(
	clusterID, namespace, ownerKind, ownerName string, opts metav1.ListOptions,
) (*structpb.Struct, error) {
	// 获取指定命名空间下的所有符合条件的 Pod
	ret, err := cli.NewPodResCliByClusterID(clusterID).List(namespace, ownerKind, ownerName, opts)
	if err != nil {
		return nil, err
	}

	return genListResRespData(ret, res.Po)
}

func BuildListPodRelatedResResp(clusterID, namespace, podName, resKind string) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	podManifest, err := cli.NewPodResClient(clusterConf).FetchManifest(namespace, podName)
	if err != nil {
		return nil, err
	}

	// Pod 配置中资源类型为驼峰式，需要将 Resource Kind 首字母小写
	kind, resNameKey := util.Decapitalize(resKind), res.Volume2ResNameKeyMap[resKind]
	// 获取与指定 Pod 相关联的 某种资源 的资源名称列表
	resNameList := []string{}
	volumes, _ := util.GetItems(podManifest, "spec.volumes")
	for _, vol := range volumes.([]interface{}) {
		if v, ok := vol.(map[string]interface{})[kind]; ok {
			resNameList = append(resNameList, v.(map[string]interface{})[resNameKey].(string))
		}
	}

	// 获取同命名空间下指定资源列表
	relatedRes, err := res.GetGroupVersionResource(clusterConf, resKind, "")
	if err != nil {
		return nil, err
	}
	ret, err := cli.NewNsScopedResClient(clusterConf, relatedRes).List(namespace, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	manifest := ret.UnstructuredContent()

	// 按照名称匹配过滤
	relatedItems := []interface{}{}
	for _, item := range manifest["items"].([]interface{}) {
		name, _ := util.GetItems(item.(map[string]interface{}), "metadata.name")
		if util.StringInSlice(name.(string), resNameList) {
			relatedItems = append(relatedItems, item)
		}
	}
	manifest["items"] = relatedItems
	return genListResRespData(manifest, resKind)
}

// genListResRespData 根据 ResList Manifest 生成获取某类资源列表的响应结果
func genListResRespData(manifest map[string]interface{}, resKind string) (*structpb.Struct, error) {
	manifestExt := map[string]interface{}{}
	// 遍历列表中的每个资源，生成 manifestExt
	for _, item := range manifest["items"].([]interface{}) {
		uid, _ := util.GetItems(item.(map[string]interface{}), "metadata.uid")
		manifestExt[uid.(string)] = formatter.Kind2FormatFuncMap[resKind](item.(map[string]interface{}))
	}

	// 组装数据，并转换为 structpb.Struct 格式
	respData := map[string]interface{}{"manifest": manifest, "manifestExt": manifestExt}
	return util.Map2pbStruct(respData)
}

func BuildListContainerApiResp(clusterID, namespace, podName string) (*structpb.ListValue, error) {
	podManifest, err := cli.NewPodResCliByClusterID(clusterID).FetchManifest(namespace, podName)
	if err != nil {
		return nil, err
	}

	containers := []map[string]interface{}{}
	containerStatuses, _ := util.GetItems(podManifest, "status.containerStatuses")
	for _, cs := range containerStatuses.([]interface{}) {
		cs, _ := cs.(map[string]interface{})
		status, reason, message := "", "", ""
		// state 有且只有一对键值：running / terminated / waiting
		// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#containerstate-v1-core
		for k := range cs["state"].(map[string]interface{}) {
			status = k
			reason, _ = util.GetWithDefault(cs, []string{"state", k, "reason"}, k).(string)
			message, _ = util.GetWithDefault(cs, []string{"state", k, "message"}, k).(string)
		}
		containers = append(containers, map[string]interface{}{
			"containerId": extractContainerID(util.GetWithDefault(cs, "containerID", "").(string)),
			"image":       cs["image"].(string),
			"name":        cs["name"].(string),
			"status":      status,
			"reason":      reason,
			"message":     message,
		})
	}
	return util.MapSlice2ListValue(containers)
}

func BuildGetContainerApiResp(clusterID, namespace, podName, containerName string) (*structpb.Struct, error) {
	podManifest, err := cli.NewPodResCliByClusterID(clusterID).FetchManifest(namespace, podName)
	if err != nil {
		return nil, err
	}

	// 遍历查找指定容器的 Spec 及 Status，若其中某项不存在，则抛出错误
	var curContainerStatus map[string]interface{}
	var curContainerSpec map[string]interface{}
	containerSpec, _ := util.GetItems(podManifest, "spec.containers")
	for _, csp := range containerSpec.([]interface{}) {
		csp, _ := csp.(map[string]interface{})
		if containerName == csp["name"].(string) {
			curContainerSpec = csp
		}
	}
	containerStatuses, _ := util.GetItems(podManifest, "status.containerStatuses")
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
		"hostName":      util.GetWithDefault(podManifest, "spec.nodeName", "--"),
		"hostIP":        util.GetWithDefault(podManifest, "status.hostIP", "--"),
		"containerIP":   util.GetWithDefault(podManifest, "status.podIP", "--"),
		"containerID":   extractContainerID(util.GetWithDefault(curContainerStatus, "containerID", "").(string)),
		"containerName": containerName,
		"image":         util.GetWithDefault(curContainerStatus, "image", "--"),
		"networkMode":   util.GetWithDefault(podManifest, "spec.dnsPolicy", "--"),
		"ports":         util.GetWithDefault(curContainerSpec, "ports", []interface{}{}),
		"volumes":       []map[string]interface{}{},
		"resources":     util.GetWithDefault(curContainerSpec, "resources", map[string]interface{}{}),
		"command": map[string]interface{}{
			"command": util.GetWithDefault(curContainerSpec, "command", []string{}),
			"args":    util.GetWithDefault(curContainerSpec, "args", []string{}),
		},
	}
	mounts := util.GetWithDefault(curContainerSpec, "volumeMounts", []map[string]interface{}{})
	for _, m := range mounts.([]interface{}) {
		m, _ := m.(map[string]interface{})
		containerInfo["volumes"] = append(containerInfo["volumes"].([]map[string]interface{}), map[string]interface{}{
			"name":      util.GetWithDefault(m, "name", "--"),
			"mountPath": util.GetWithDefault(m, "mountPath", "--"),
			"readonly":  util.GetWithDefault(m, "readOnly", "--"),
		})
	}

	return util.Map2pbStruct(containerInfo)
}

// 去除容器 ID 前缀，原格式：docker://[a-zA-Z0-9]{64}
func extractContainerID(rawContainerID string) string {
	return strings.Replace(rawContainerID, "docker://", "", 1)
}
