/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package federated

import (
	"context"
	"errors"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/clientutil"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/proxy"
)

// BCS 集群ID Label
const ClusterIdLabel = "bkbcs.tencent.com/cluster-id"

// PodStor PodInterface 实现
type PodStor struct {
	members         []string
	k8sClientMap    map[string]*kubernetes.Clientset
	proxyHandlerMap map[string]*proxy.ProxyHandler
}

// NewPodStor
func NewPodStor(members []string) (*PodStor, error) {
	stor := &PodStor{
		members:         members,
		k8sClientMap:    make(map[string]*kubernetes.Clientset),
		proxyHandlerMap: make(map[string]*proxy.ProxyHandler),
	}

	for _, k := range members {
		k8sClient, err := clientutil.GetKubeClientByClusterId(k)
		if err != nil {
			return nil, err
		}
		stor.k8sClientMap[k] = k8sClient

		proxyHandler, err := proxy.NewProxyHandler(k)
		if err != nil {
			return nil, err
		}
		stor.proxyHandlerMap[k] = proxyHandler
	}

	return stor, nil
}

// List 查询Pod列表, Json格式返回
func (p *PodStor) List(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.PodList, error) {
	typeMata := metav1.TypeMeta{APIVersion: "v1", Kind: "PodList"}
	listMeta := metav1.ListMeta{
		SelfLink:        p.selfLink(namespace, ""),
		ResourceVersion: "0",
	}

	podList := &v1.PodList{
		TypeMeta: typeMata,
		ListMeta: listMeta,
		Items:    []v1.Pod{},
	}
	for k, v := range p.k8sClientMap {
		result, err := v.CoreV1().Pods(namespace).List(ctx, opts)
		if err != nil {
			return nil, err
		}
		for _, item := range result.Items {
			if item.Annotations == nil {
				item.Annotations = map[string]string{ClusterIdLabel: k}
			} else {
				item.Annotations[ClusterIdLabel] = k
			}
			podList.Items = append(podList.Items, item)
		}
	}
	return podList, nil
}

func (p *PodStor) selfLink(namespace, name string) string {
	if name == "" {
		return fmt.Sprintf("/api/v1/namespaces/%s/pods", namespace)
	}
	return fmt.Sprintf("/api/v1/namespaces/%s/pods/%s", namespace, name)
}

// ListAsTable 查询Pod列表, kubectl格式返回
func (p *PodStor) ListAsTable(ctx context.Context, namespace string, acceptHeader string, opts metav1.ListOptions) (*metav1.Table, error) {
	typeMata := metav1.TypeMeta{APIVersion: "meta.k8s.io/v1", Kind: "Table"}
	listMeta := metav1.ListMeta{
		SelfLink:        p.selfLink(namespace, ""),
		ResourceVersion: "0",
	}

	result := &metav1.Table{
		TypeMeta: typeMata,
		ListMeta: listMeta,
	}

	// 联邦集群添加集群Id一列
	clusterId := metav1.TableColumnDefinition{
		Name:        "Cluster Id",
		Type:        "string",
		Format:      "",
		Description: "bcs cluster id",
	}
	columns := []metav1.TableColumnDefinition{clusterId}

	rows := []metav1.TableRow{}

	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	for clusterId, v := range p.k8sClientMap {
		resultTemp := &metav1.Table{}
		err := v.CoreV1().RESTClient().Get().
			Namespace(namespace).
			Resource("pods").
			VersionedParams(&opts, scheme.ParameterCodec).
			Timeout(timeout).
			SetHeader("Accept", acceptHeader).
			Do(ctx).
			Into(resultTemp)
		if err != nil {
			return nil, err
		}

		if len(columns) == 1 {
			columns = append(columns, resultTemp.ColumnDefinitions...)
		}

		for idx, row := range resultTemp.Rows {
			cells := []interface{}{clusterId}
			cells = append(cells, row.Cells...)
			resultTemp.Rows[idx].Cells = cells
			rows = append(rows, resultTemp.Rows[idx])
		}
	}

	result.ColumnDefinitions = columns
	result.Rows = rows

	return result, nil
}

// Get 获取单个Pod, Json格式返回
func (p *PodStor) Get(ctx context.Context, namespace string, name string, opts metav1.GetOptions) (*v1.Pod, error) {
	typeMata := metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"}

	for k, v := range p.k8sClientMap {
		pod, err := v.CoreV1().Pods(namespace).Get(ctx, name, opts)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		if pod.Annotations == nil {
			pod.Annotations = map[string]string{ClusterIdLabel: k}
		} else {
			pod.Annotations[ClusterIdLabel] = k
		}

		pod.TypeMeta = typeMata
		return pod, nil
	}
	return nil, apierrors.NewNotFound(v1.Resource("pods"), name)
}

// GetAsTable 获取单个Pod, Table格式返回
func (p *PodStor) GetAsTable(ctx context.Context, namespace string, name string, acceptHeader string, opts metav1.GetOptions) (*metav1.Table, error) {
	typeMata := metav1.TypeMeta{APIVersion: "meta.k8s.io/v1", Kind: "Table"}
	listMeta := metav1.ListMeta{
		SelfLink:        p.selfLink(namespace, name),
		ResourceVersion: "0",
	}

	result := &metav1.Table{
		TypeMeta: typeMata,
		ListMeta: listMeta,
	}

	// 联邦集群添加集群Id一列
	clusterId := metav1.TableColumnDefinition{
		Name:        "Cluster Id",
		Type:        "string",
		Format:      "",
		Description: "bcs cluster id",
	}
	columns := []metav1.TableColumnDefinition{clusterId}

	rows := []metav1.TableRow{}

	for clusterId, v := range p.k8sClientMap {
		resultTemp := &metav1.Table{}
		err := v.CoreV1().RESTClient().Get().
			Namespace(namespace).
			Resource("pods").
			VersionedParams(&opts, scheme.ParameterCodec).
			SetHeader("Accept", acceptHeader).
			SubResource(name).
			Do(ctx).
			Into(resultTemp)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return nil, err
		}

		if len(columns) == 1 {
			columns = append(columns, resultTemp.ColumnDefinitions...)
		}

		for idx, row := range resultTemp.Rows {
			cells := []interface{}{clusterId}
			cells = append(cells, row.Cells...)
			resultTemp.Rows[idx].Cells = cells
			rows = append(rows, resultTemp.Rows[idx])
		}
	}

	result.ColumnDefinitions = columns
	result.Rows = rows

	return result, nil
}

func (p *PodStor) getClientByPod(pod *v1.Pod) (*kubernetes.Clientset, error) {
	clusterId, ok := pod.Annotations[ClusterIdLabel]
	if !ok {
		return nil, errors.New("cluter_id not in annotions")
	}
	client, ok := p.k8sClientMap[clusterId]
	if !ok {
		return nil, errors.New("cluter_id not in annotions")
	}
	return client, nil
}

// Delete 删除单个Pod
func (p *PodStor) Delete(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) (*v1.Pod, error) {
	pod, err := p.Get(ctx, namespace, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	client, err := p.getClientByPod(pod)
	if err != nil {
		return nil, err
	}
	if err = client.CoreV1().Pods(namespace).Delete(ctx, name, opts); err != nil {
		return nil, err
	}

	return pod, nil
}

// Watch 获取 Pod 列表
func (p *PodStor) Watch(ctx context.Context, namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true

	for _, v := range p.k8sClientMap {
		watch, err := v.CoreV1().RESTClient().Get().
			Namespace(namespace).
			Resource("pods").
			VersionedParams(&opts, scheme.ParameterCodec).
			Timeout(timeout).
			SetHeader("Accept", "application/json").
			Watch(ctx)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return nil, err
		}

		return watch, nil
	}
	return nil, apierrors.NewNotFound(v1.Resource("pods"), "")
}

// GetLogs kubectl logs 命令
func (p *PodStor) GetLogs(ctx context.Context, namespace string, name string, opts *v1.PodLogOptions) (*restclient.Request, error) {
	for _, v := range p.k8sClientMap {
		_, err := v.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
		}
		return v.CoreV1().Pods(namespace).GetLogs(name, opts), nil
	}
	return nil, apierrors.NewNotFound(v1.Resource("pods"), "")
}

// Exec kubectl exec 命令
func (p *PodStor) Exec(ctx context.Context, namespace string, name string, opts metav1.GetOptions) (*proxy.ProxyHandler, error) {
	for k, v := range p.k8sClientMap {
		_, err := v.CoreV1().Pods(namespace).Get(ctx, name, opts)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return nil, err
		}

		return p.proxyHandlerMap[k], nil
	}
	return nil, apierrors.NewNotFound(v1.Resource("pods"), name)
}
