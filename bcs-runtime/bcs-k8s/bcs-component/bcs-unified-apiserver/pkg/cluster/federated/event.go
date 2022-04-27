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
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/clientutil"
)

// EventStor EventInterface 实现
type EventStor struct {
	members      []string
	k8sClientMap map[string]*kubernetes.Clientset
}

// NewEventStor
func NewEventStor(masterClientId string, members []string) (*EventStor, error) {
	stor := &EventStor{members: members, k8sClientMap: make(map[string]*kubernetes.Clientset)}
	for _, k := range members {
		k8sClient, err := clientutil.GetKubeClientByClusterId(k)
		if err != nil {
			return nil, err
		}
		stor.k8sClientMap[k] = k8sClient
	}
	return stor, nil
}

// List 查询Event列表, Json格式返回
func (p *EventStor) List(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.EventList, error) {
	listMeta := metav1.ListMeta{
		SelfLink:        p.selfLink(namespace, ""),
		ResourceVersion: "0",
	}

	EventList := &v1.EventList{
		ListMeta: listMeta,
		Items:    []v1.Event{},
	}
	for k, v := range p.k8sClientMap {
		result, err := v.CoreV1().Events(namespace).List(ctx, opts)
		if err != nil {
			return nil, err
		}
		for _, item := range result.Items {
			if item.Annotations == nil {
				item.Annotations = map[string]string{ClusterIdLabel: k}
			} else {
				item.Annotations[ClusterIdLabel] = k
			}
			EventList.Items = append(EventList.Items, item)
		}
	}
	return EventList, nil
}

// ListAsTable 查询 Event 列表, kubectl格式返回
func (p *EventStor) ListAsTable(ctx context.Context, namespace string, acceptHeader string, opts metav1.ListOptions) (*metav1.Table, error) {
	listMeta := metav1.ListMeta{
		SelfLink:        p.selfLink(namespace, ""),
		ResourceVersion: "0",
	}

	result := &metav1.Table{
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
			Resource("events").
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

func (p *EventStor) selfLink(namespace, name string) string {
	if name == "" {
		return fmt.Sprintf("/api/v1/namespaces/%s/events", namespace)
	}
	return fmt.Sprintf("/api/v1/namespaces/%s/events/%s", namespace, name)
}

// Get 获取单个Event, Json格式返回
func (p *EventStor) Get(ctx context.Context, namespace string, name string, opts metav1.GetOptions) (*v1.Event, error) {
	for k, v := range p.k8sClientMap {
		Event, err := v.CoreV1().Events(namespace).Get(ctx, name, opts)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		if Event.Annotations == nil {
			Event.Annotations = map[string]string{ClusterIdLabel: k}
		} else {
			Event.Annotations[ClusterIdLabel] = k
		}
		return Event, nil
	}
	return nil, apierrors.NewNotFound(v1.Resource("Events"), name)
}

// GetAsTable 获取单个Event, Table格式返回
func (p *EventStor) GetAsTable(ctx context.Context, namespace string, name string, acceptHeader string, opts metav1.GetOptions) (*metav1.Table, error) {
	listMeta := metav1.ListMeta{
		SelfLink:        p.selfLink(namespace, name),
		ResourceVersion: "0",
	}

	result := &metav1.Table{
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
			Resource("events").
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
