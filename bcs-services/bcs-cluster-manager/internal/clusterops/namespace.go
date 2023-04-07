/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package clusterops

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// CreateNamespace create namespace
func (ko *K8SOperator) CreateNamespace(ctx context.Context, clusterID string, name string) error {
	if ko == nil {
		return ErrServerNotInit
	}
	clientInterface, err := ko.GetClusterClient(clusterID)
	if err != nil {
		blog.Errorf("CreateNamespace[%s] GetClusterClient failed: %v", clusterID, err)
		return err
	}

	_, err = clientInterface.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		blog.Errorf("CreateNamespace[%s:%s] getNamespace failed: %v", clusterID, name, err)
		return err
	}

	if err == nil {
		blog.Errorf("CreateNamespace[%s:%s] getNamespace success", clusterID, name)
		return nil
	}

	newNs := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	_, err = clientInterface.CoreV1().Namespaces().Create(ctx, newNs, metav1.CreateOptions{})
	if err != nil {
		blog.Errorf("CreateNamespace[%s:%s] createNamespace failed: %v", clusterID, name, err)
		return err
	}
	blog.Infof("CreateNamespace[%s:%s] createNamespace success", clusterID, name)

	return nil
}
