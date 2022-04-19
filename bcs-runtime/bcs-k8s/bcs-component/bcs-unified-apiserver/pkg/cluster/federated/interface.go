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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodInterface has methods to work with Pod resources.
type PodInterface interface {
	List(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.PodList, error)
	ListAsTable(ctx context.Context, namespace string, accept string, opts metav1.ListOptions) (*metav1.Table, error)
	Get(ctx context.Context, namespace string, name string, opts metav1.GetOptions) (*v1.Pod, error)
	GetAsTable(ctx context.Context, namespace string, name string, accept string, opts metav1.GetOptions) (*metav1.Table, error)
	Delete(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) (*v1.Pod, error)
}
