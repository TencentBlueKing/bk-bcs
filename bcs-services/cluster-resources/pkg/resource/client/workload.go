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
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam"
	clusterAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/cluster"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

// PodClient xxx
type PodClient struct {
	ResClient
}

// NewPodClient xxx
func NewPodClient(ctx context.Context, conf *res.ClusterConf) *PodClient {
	podRes, _ := res.GetGroupVersionResource(ctx, conf, res.Po, "")
	return &PodClient{ResClient{NewDynamicClient(conf), conf, podRes}}
}

// NewPodCliByClusterID xxx
func NewPodCliByClusterID(ctx context.Context, clusterID string) *PodClient {
	return NewPodClient(ctx, res.NewClusterConfig(clusterID))
}

// List xxx
func (c *PodClient) List(
	ctx context.Context, namespace, ownerKind, ownerName string, opts metav1.ListOptions,
) (map[string]interface{}, error) {
	ret, err := c.ResClient.List(ctx, namespace, opts)
	if err != nil {
		return nil, err
	}
	manifest := ret.UnstructuredContent()
	// 只有指定 OwnerReferences 信息才会再过滤
	if ownerKind == "" || ownerName == "" {
		return manifest, nil
	}

	// 找到当前指定资源关联的 Pod 的 OwnerReferences 信息
	podOwnerRefs, err := c.getPodOwnerRefs(ctx, c.conf, namespace, ownerKind, ownerName, opts)
	if err != nil {
		return nil, err
	}
	manifest["items"] = c.filterByOwnerRefs(manifest["items"].([]interface{}), podOwnerRefs)
	return manifest, nil
}

// ListAllPods 获取所有命名空间的 Pod，要求有集群管理权限
func (c *PodClient) ListAllPods(
	ctx context.Context, projectID, clusterID string, opts metav1.ListOptions,
) (map[string]interface{}, error) {
	// 权限控制为集群管理
	permCtx := clusterAuth.NewPermCtx(
		ctx.Value(ctxkey.UsernameKey).(string), projectID, clusterID,
	)
	if allow, err := iam.NewClusterPerm(projectID).CanManage(permCtx); err != nil {
		return nil, err
	} else if !allow {
		return nil, errorx.New(errcode.NoIAMPerm, i18n.GetMsg(ctx, "无查看指定节点上运行的 Pod 的权限"))
	}

	ret, err := c.ResClient.cli.Resource(c.res).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	return ret.UnstructuredContent(), nil
}

// getPodOwnerRefs 非直接关联 Pod 的资源，找到下层直接关联的子资源
func (c *PodClient) getPodOwnerRefs(
	ctx context.Context, clusterConf *res.ClusterConf, namespace, ownerKind, ownerName string, opts metav1.ListOptions,
) ([]map[string]string, error) {
	subOwnerRefs := []map[string]string{{"kind": ownerKind, "name": ownerName}}
	if !slice.StringInSlice(ownerKind, []string{res.Deploy, res.CJ}) {
		return subOwnerRefs, nil
	}

	// Deployment/CronJob 不直接关联 Pod，而是通过 ReplicaSet/Job 间接关联，需要向下钻取 Pod 的 OwnerReferences 信息
	subResKind := map[string]string{res.Deploy: res.RS, res.CJ: res.Job}[ownerKind]
	subRes, err := res.GetGroupVersionResource(ctx, clusterConf, subResKind, "")
	if err != nil {
		return nil, err
	}
	ret, err := NewResClient(clusterConf, subRes).List(ctx, namespace, opts)
	if err != nil {
		return nil, err
	}
	ownerRefs := []map[string]string{}
	for _, r := range c.filterByOwnerRefs(ret.UnstructuredContent()["items"].([]interface{}), subOwnerRefs) {
		resName, _ := mapx.GetItems(r.(map[string]interface{}), "metadata.name")
		ownerRefs = append(ownerRefs, map[string]string{"kind": subResKind, "name": resName.(string)})
	}
	return ownerRefs, nil
}

// filterByOwnerRefs 根据 OwnerReferences 过滤关联的子资源
func (c *PodClient) filterByOwnerRefs(subResItems []interface{}, ownerRefs []map[string]string) []interface{} {
	rets := []interface{}{}
	for _, subRes := range subResItems {
		resOwnerRefs, err := mapx.GetItems(subRes.(map[string]interface{}), "metadata.ownerReferences")
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

// GetManifest 获取指定 Pod Manifest
func (c *PodClient) GetManifest(ctx context.Context, namespace, podName string) (map[string]interface{}, error) {
	ret, err := c.Get(ctx, namespace, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ret.UnstructuredContent(), nil
}

// ListPodRelatedRes 获取 Pod 关联的某种资源列表
func (c *PodClient) ListPodRelatedRes(
	ctx context.Context, namespace, podName, resKind string,
) (map[string]interface{}, error) {
	// 获取同命名空间下指定的关联资源列表
	relatedRes, err := res.GetGroupVersionResource(ctx, c.conf, resKind, "")
	if err != nil {
		return nil, err
	}
	ret, err := NewResClient(c.conf, relatedRes).List(ctx, namespace, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	manifest := ret.UnstructuredContent()

	// 获取 Pod 关联的某种资源的名称列表，匹配过滤
	resNameList, err := c.getPodRelatedResNameList(ctx, namespace, podName, resKind)
	if err != nil {
		return nil, err
	}
	relatedItems := []interface{}{}
	for _, item := range manifest["items"].([]interface{}) {
		name, _ := mapx.GetItems(item.(map[string]interface{}), "metadata.name")
		if slice.StringInSlice(name.(string), resNameList) {
			relatedItems = append(relatedItems, item)
		}
	}
	manifest["items"] = relatedItems
	return manifest, nil
}

// getPodRelatedResNameList 获取 Pod 关联的某种资源的名称列表
func (c *PodClient) getPodRelatedResNameList(
	ctx context.Context, namespace, podName, resKind string,
) ([]string, error) {
	podManifest, err := c.GetManifest(ctx, namespace, podName)
	if err != nil {
		return nil, err
	}
	// Pod 配置中资源类型为驼峰式，需要将 Resource Kind 首字母小写
	kind, resNameKey := stringx.Decapitalize(resKind), res.Volume2ResNameKeyMap[resKind]
	// 获取与指定 Pod 相关联的 某种资源 的资源名称列表
	resNameList := []string{}
	volumes, _ := mapx.GetItems(podManifest, "spec.volumes")
	for _, vol := range volumes.([]interface{}) {
		if v, ok := vol.(map[string]interface{})[kind]; ok {
			resNameList = append(resNameList, v.(map[string]interface{})[resNameKey].(string))
		}
	}
	return resNameList, nil
}

// ExecCommand 在指定容器中执行命令，获取 stdout, stderr
func (c *PodClient) ExecCommand(
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

// BatchDelete 批量删除 Pod（需为同一命名空间下的）
// NOTE 由于 DeleteCollection 是基于 LabelSelector 等而非 PodName，因此这里还是单个单个 Pod 删除
// TODO 针对较多 Pod 同时删除的场景进行性能优化，比如拆分任务使用 goroutine 执行？
func (c *PodClient) BatchDelete(
	ctx context.Context, namespace string, podNames []string, opts metav1.DeleteOptions,
) (err error) {
	if len(podNames) == 0 {
		return nil
	}
	if err = c.permValidate(ctx, action.Delete, namespace); err != nil {
		return err
	}
	for _, pn := range podNames {
		if err = c.cli.Resource(c.res).Namespace(namespace).Delete(ctx, pn, opts); err != nil {
			return err
		}
	}
	return nil
}
