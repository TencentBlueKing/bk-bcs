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

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/clientutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
			item.Labels["bcs_cluster_id"] = k
		}

		podList.TypeMeta = result.TypeMeta
		podList.Items = append(podList.Items, result.Items...)
	}
	return podList, nil
}

func addTypeInformationToObject(obj runtime.Object) error {
	gvks, _, err := scheme.Scheme.ObjectKinds(obj)
	if err != nil {
		return fmt.Errorf("missing apiVersion or kind and cannot assign it; %w", err)
	}

	for _, gvk := range gvks {
		if len(gvk.Kind) == 0 {
			continue
		}
		if len(gvk.Version) == 0 || gvk.Version == runtime.APIVersionInternal {
			continue
		}
		obj.GetObjectKind().SetGroupVersionKind(gvk)
		break
	}

	return nil
}

// ListAsTable kubectl 返回
func (p *PodStor) ListAsTable(ctx context.Context, namespace string, opts *metav1.ListOptions, accept string) (*metav1.Table, error) {
	potTable := &metav1.Table{}
	isSucc := false
	for _, v := range p.k8sClientMap {
		result := &metav1.Table{}
		err := v.CoreV1().RESTClient().Get().
			Namespace(namespace).
			Resource("pods").
			VersionedParams(opts, scheme.ParameterCodec).
			SetHeader("Accept", accept).
			Do(ctx).
			Into(result)
		if err != nil {
			return nil, err
		}
		fmt.Println("kind", result.APIVersion, result.Kind)
		if !isSucc {
			potTable = result
			isSucc = true
		} else {
			result.Rows = append(result.Rows, potTable.Rows...)
			potTable = result
		}
	}
	potTable.Kind = "Table"
	potTable.APIVersion = "meta.k8s.io/v1"
	// err := addTypeInformationToObject(potTable)
	// if err != nil {
	// 	return nil, err
	// }
	return potTable, nil
}
