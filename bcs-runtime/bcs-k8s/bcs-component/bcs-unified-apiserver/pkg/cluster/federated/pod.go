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

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/clientutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

type PodStor struct {
	members      []string
	k8sClientMap map[string]*kubernetes.Clientset
}

func NewPodStor(members []string) (*PodStor, error) {
	stor := &PodStor{members: members, k8sClientMap: make(map[string]*kubernetes.Clientset)}
	for _, k := range members {
		k8sClient, err := clientutil.GetKubeClientByClusterId(k)
		if err != nil {
			return nil, err
		}
		stor.k8sClientMap[k] = k8sClient
	}
	return stor, nil
}

func (p *PodStor) List(ctx context.Context, namespace string, opts *metav1.ListOptions) (*v1.PodList, error) {
	podList := &v1.PodList{}
	for k, v := range p.k8sClientMap {
		result, err := v.CoreV1().Pods(namespace).List(ctx, *opts)
		if err != nil {
			return nil, err
		}
		for _, item := range result.Items {
			if item.Annotations == nil {
				item.Annotations = map[string]string{"bcs_cluster_id": k}
			} else {
				item.Annotations["bcs_cluster_id"] = k
			}
		}

		podList.TypeMeta = result.TypeMeta
		podList.Items = append(podList.Items, result.Items...)
	}
	return podList, nil
}

func (p *PodStor) SelfLink(namespace string) string {
	return fmt.Sprintf("/api/v1/namespaces/%s/pods", namespace)
}

// ListAsTable kubectl 返回
func (p *PodStor) ListAsTable(ctx context.Context, namespace string, opts *metav1.ListOptions, accept string) (*metav1.Table, error) {
	typeMata := metav1.TypeMeta{APIVersion: "meta.k8s.io/v1", Kind: "Table"}
	listMeta := metav1.ListMeta{
		SelfLink:        p.SelfLink(namespace),
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
			VersionedParams(opts, scheme.ParameterCodec).
			Timeout(timeout).
			SetHeader("Accept", accept).
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
