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

package app

import (
	"reflect"

	"bk-bcs/bcs-common/common/blog"
	bcsv1 "bk-bcs/bcs-services/bcs-webhook-server/pkg/apis/bk-bcs/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// create crd of BcsDbPrivConfig
func createBcsDbPrivConfig(clientset apiextensionsclient.Interface) (bool, error) {
	bcsDbPrivConfigPlural := "bcsdbprivconfigs"

	bcsDbPrivConfigFullName := "bcsdbprivconfigs" + "." + bcsv1.SchemeGroupVersion.Group

	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: bcsDbPrivConfigFullName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   bcsv1.SchemeGroupVersion.Group,   // BcsDbPrivConfigsGroup,
			Version: bcsv1.SchemeGroupVersion.Version, // BcsDbPrivConfigsVersion,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   bcsDbPrivConfigPlural,
				Kind:     reflect.TypeOf(bcsv1.BcsDbPrivConfig{}).Name(),
				ListKind: reflect.TypeOf(bcsv1.BcsDbPrivConfigList{}).Name(),
			},
		},
	}

	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			blog.Infof("crd is already exists: %s", err)
			return false, nil
		}
		blog.Errorf("create crd failed: %s", err)
		return false, err
	}
	return true, nil
}
