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

package ns

import (
	"context"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//CheckNamespace if namespace exists
func CheckNamespace(c cache.Cache, cli client.Client, name string) error {
	namespaceName := types.NamespacedName{
		Name: name,
	}
	ns := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		//todo(DeveloperJim): add spec & status after testing
	}
	err := c.Get(context.TODO(), namespaceName, ns)
	if err == nil {
		return nil
	}
	if errors.IsNotFound(err) {
		// Object not found, create new one directly
		createErr := cli.Create(context.TODO(), ns)
		if createErr == nil {
			blog.Infof("mesos-adaptor creat new namespace %s on success", namespaceName.String())
			return nil
		}
		if errors.IsAlreadyExists(createErr) {
			blog.Warnf("mesos-adaptor creat exist namespace %s, skip", namespaceName.String())
			return nil
		}
		blog.Errorf("mesos-adaptor create namespace %s failed, %s", namespaceName.String(), createErr.Error())
		return createErr
	}
	blog.Errorf("mesos-adaptor check namespace %s failed, %s", namespaceName.String(), err.Error())
	return err
}
