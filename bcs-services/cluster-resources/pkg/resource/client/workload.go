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

package client

import (
	"bytes"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
)

type PodResClient struct {
	NsScopedResClient
}

// NewPodResClient ...
func NewPodResClient(conf *res.ClusterConf) *PodResClient {
	podRes, _ := res.GetGroupVersionResource(conf, res.Po, "")
	return &PodResClient{NsScopedResClient{NewDynamicClient(conf), conf, podRes}}
}

func (c *PodResClient) List(
	namespace, ownerKind, ownerName string, opts metav1.ListOptions,
) (map[string]interface{}, error) {
	ret, err := c.NsScopedResClient.List(namespace, opts)
	if err != nil {
		return nil, err
	}
	manifest := ret.UnstructuredContent()

	// 找到当前指定资源关联的 Pod 的 OwnerReferences 信息
	podOwnerRefs, err := c.getPodOwnerRefs(c.conf, namespace, ownerKind, ownerName)
	if err != nil {
		return nil, err
	}
	manifest["items"] = c.filterByOwnerRefs(manifest["items"].([]interface{}), podOwnerRefs)
	return manifest, nil
}

// 非直接关联 Pod 的资源，找到下层直接关联的子资源
func (c *PodResClient) getPodOwnerRefs(
	clusterConf *res.ClusterConf, namespace, ownerKind, ownerName string,
) ([]map[string]string, error) {
	subOwnerRefs := []map[string]string{{"kind": ownerKind, "name": ownerName}}
	if !util.StringInSlice(ownerKind, []string{res.Deploy, res.CJ}) {
		return subOwnerRefs, nil
	}

	// Deployment/CronJob 不直接关联 Pod，而是通过 ReplicaSet/Job 间接关联，需要向下钻取 Pod 的 OwnerReferences 信息
	subResKind := map[string]string{res.Deploy: res.RS, res.CJ: res.Job}[ownerKind]
	subRes, err := res.GetGroupVersionResource(clusterConf, subResKind, "")
	if err != nil {
		return nil, err
	}
	ret, err := NewNsScopedResClient(clusterConf, subRes).List(namespace, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	ownerRefs := []map[string]string{}
	for _, res := range c.filterByOwnerRefs(ret.UnstructuredContent()["items"].([]interface{}), subOwnerRefs) {
		resName, _ := util.GetItems(res.(map[string]interface{}), "metadata.name")
		ownerRefs = append(ownerRefs, map[string]string{"kind": subResKind, "name": resName.(string)})
	}
	return ownerRefs, nil
}

// 根据 owner_references 过滤关联的子资源
func (c *PodResClient) filterByOwnerRefs(subResItems []interface{}, ownerRefs []map[string]string) []interface{} {
	rets := []interface{}{}
	for _, subRes := range subResItems {
		resOwnerRefs, err := util.GetItems(subRes.(map[string]interface{}), "metadata.ownerReferences")
		if err != nil {
			continue
		}
		for _, resOwnerRef := range resOwnerRefs.([]interface{}) {
			for _, ref := range ownerRefs {
				kind, name := "", ""
				if r, ok := resOwnerRef.(map[string]interface{}); ok {
					kind, name = r["kind"].(string), r["name"].(string)
				}
				if kind == ref["kind"] && name == ref["name"] {
					rets = append(rets, subRes)
					break
				}
			}
		}
	}
	return rets
}

// FetchManifest 获取指定 Pod Manifest
func (c *PodResClient) FetchManifest(namespace, podName string) (map[string]interface{}, error) {
	ret, err := c.Get(namespace, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ret.UnstructuredContent(), nil
}

// ExecCommand 在指定容器中执行命令，获取 stdout, stderr
func (c *PodResClient) ExecCommand(
	namespace, podName, containerName string, cmds []string,
) (string, string, error) {
	clientSet, err := kubernetes.NewForConfig(c.conf.Rest)
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
	exec, err := remotecommand.NewSPDYExecutor(c.conf.Rest, "POST", req.URL())
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
