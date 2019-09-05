/*
Copyright (C) 2019 The BlueKing Authors. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package ns

import (
	"bk-bcs/bcs-common/common/blog"
	"context"

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
