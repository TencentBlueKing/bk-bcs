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

import (
	"bytes"
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
)

// ListNamespaceScopedRes 获取命名空间维度资源列表
func ListNamespaceScopedRes(
	conf *rest.Config,
	namespace string,
	resource schema.GroupVersionResource,
	opts metav1.ListOptions,
) (*unstructured.UnstructuredList, error) {
	client := newDynamicClient(conf)
	return client.Resource(resource).Namespace(namespace).List(context.TODO(), opts)
}

// GetNamespaceScopedRes 获取单个命名空间维度资源
func GetNamespaceScopedRes(
	conf *rest.Config,
	namespace string,
	name string,
	resource schema.GroupVersionResource,
	opts metav1.GetOptions,
) (*unstructured.Unstructured, error) {
	client := newDynamicClient(conf)
	return client.Resource(resource).Namespace(namespace).Get(context.TODO(), name, opts)
}

// CreateNamespaceScopedRes 创建命名空间维度资源
func CreateNamespaceScopedRes(
	conf *rest.Config,
	manifest map[string]interface{},
	resource schema.GroupVersionResource,
	opts metav1.CreateOptions,
) (*unstructured.Unstructured, error) {
	client := newDynamicClient(conf)
	namespace, err := util.GetItems(manifest, "metadata.namespace")
	if err != nil {
		return nil, fmt.Errorf("创建 %s 需要指定 metadata.namespace", resource.Resource)
	}
	return client.Resource(resource).Namespace(namespace.(string)).Create(
		context.TODO(), &unstructured.Unstructured{Object: manifest}, opts,
	)
}

// UpdateNamespaceScopedRes 更新单个命名空间维度资源
func UpdateNamespaceScopedRes(
	conf *rest.Config,
	namespace string,
	name string,
	manifest map[string]interface{},
	resource schema.GroupVersionResource,
	opts metav1.UpdateOptions,
) (*unstructured.Unstructured, error) {
	client := newDynamicClient(conf)
	// 检查 name 与 manifest.metadata.name 是否一致
	manifestName, err := util.GetItems(manifest, "metadata.name")
	if err != nil || name != manifestName {
		return nil, fmt.Errorf("metadata.name 必须指定且与准备编辑的资源名保持一致")
	}
	return client.Resource(resource).Namespace(namespace).Update(
		context.TODO(), &unstructured.Unstructured{Object: manifest}, opts,
	)
}

// DeleteNamespaceScopedRes 删除单个命名空间维度资源
func DeleteNamespaceScopedRes(
	conf *rest.Config,
	namespace string,
	name string,
	resource schema.GroupVersionResource,
	opts metav1.DeleteOptions,
) error {
	client := newDynamicClient(conf)
	return client.Resource(resource).Namespace(namespace).Delete(context.TODO(), name, opts)
}

// FetchPodManifest 获取指定 Pod Manifest
func FetchPodManifest(clusterID, namespace, podName string) (map[string]interface{}, error) {
	clusterConf := NewClusterConfig(clusterID)
	podRes, err := GetGroupVersionResource(clusterConf, clusterID, Po, "")
	if err != nil {
		return nil, err
	}

	var ret *unstructured.Unstructured
	ret, err = GetNamespaceScopedRes(clusterConf, namespace, podName, podRes, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ret.UnstructuredContent(), nil
}

// ExecCommand 在指定容器中执行命令，获取 stdout, stderr
func ExecCommand(clusterID, namespace, podName, containerName string, cmds []string) (string, string, error) {
	clusterConf := NewClusterConfig(clusterID)
	clientSet, err := kubernetes.NewForConfig(clusterConf)
	if err != nil {
		return "", "", err
	}

	req := clientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", containerName)

	opts := &v1.PodExecOptions{
		Command: cmds,
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
	}
	req.VersionedParams(opts, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(clusterConf, "POST", req.URL())
	if err != nil {
		return "", "", err
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return "", "", err
	}
	return stdout.String(), stderr.String(), err
}
