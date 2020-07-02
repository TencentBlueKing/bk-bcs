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

package etcd

import (
	"fmt"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/pkg/apis/bkbcs/v2"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (store *managerStore) CheckCustomResourceRegisterExist(crr *commtypes.Crr) (string, bool) {
	client := store.BkbcsClient.Crrs(DefaultNamespace)
	obj, err := client.Get(crr.Spec.Names.Kind, metav1.GetOptions{})
	if err == nil {
		return obj.ResourceVersion, true
	}

	return "", false
}

//save custom resource register
func (store *managerStore) SaveCustomResourceRegister(crr *commtypes.Crr) error {
	client := store.BkbcsClient.Crrs(DefaultNamespace)
	v2Crr := &v2.Crr{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdCrr,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      crr.Spec.Names.Kind,
			Namespace: DefaultNamespace,
		},
		Spec: v2.CrrSpec{
			Crr: *crr,
		},
	}

	var err error
	rv, exist := store.CheckCustomResourceRegisterExist(crr)
	if exist {
		v2Crr.ResourceVersion = rv
		_, err = client.Update(v2Crr)
	} else {
		_, err = client.Create(v2Crr)
	}
	return err
}

func (store *managerStore) DeleteCustomResourceRegister(name string) error {
	client := store.BkbcsClient.Crrs(DefaultNamespace)
	err := client.Delete(name, &metav1.DeleteOptions{})
	return err
}

func (store *managerStore) ListCustomResourceRegister() ([]*commtypes.Crr, error) {
	client := store.BkbcsClient.Crrs(DefaultNamespace)
	v2Crrs, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	crrs := make([]*commtypes.Crr, 0, len(v2Crrs.Items))
	for _, crr := range v2Crrs.Items {
		obj := crr.Spec.Crr
		crrs = append(crrs, &obj)
	}

	return crrs, nil
}

//crd namespace = crd.kind=crd.namespace
func getCrdNamespace(kind, ns string) string {
	return fmt.Sprintf("%s-%s", kind, ns)
}

func (store *managerStore) CheckCustomResourceDefinitionExist(crd *commtypes.Crd) (string, bool) {
	client := store.BkbcsClient.Crds(getCrdNamespace(string(crd.Kind), crd.NameSpace))
	v2Crd, err := client.Get(crd.Name, metav1.GetOptions{})
	if err == nil {
		return v2Crd.ResourceVersion, true
	}

	return "", false
}

func (store *managerStore) SaveCustomResourceDefinition(crd *commtypes.Crd) error {
	//crd namespace = crd.kind=crd.namespace
	realNs := getCrdNamespace(string(crd.Kind), crd.NameSpace)
	err := store.checkNamespace(realNs)
	if err != nil {
		return err
	}

	client := store.BkbcsClient.Crds(realNs)
	v2Crd := &v2.Crd{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdCrd,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      crd.Name,
			Namespace: realNs,
		},
		Spec: v2.CrdSpec{
			Crd: *crd,
		},
	}

	rv, exist := store.CheckCustomResourceDefinitionExist(crd)
	if exist {
		v2Crd.ResourceVersion = rv
		_, err = client.Update(v2Crd)
	} else {
		_, err = client.Create(v2Crd)
	}
	return err
}

func (store *managerStore) DeleteCustomResourceDefinition(kind, ns, name string) error {
	client := store.BkbcsClient.Crds(getCrdNamespace(kind, ns))
	err := client.Delete(name, &metav1.DeleteOptions{})
	return err
}

func (store *managerStore) ListAllCrds(kind string) ([]*commtypes.Crd, error) {
	client := store.BkbcsClient.Crds("")
	v2Crds, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	crds := make([]*commtypes.Crd, 0, len(v2Crds.Items))
	for _, crd := range v2Crds.Items {
		if strings.Contains(crd.Namespace, kind) {
			obj := crd.Spec.Crd
			crds = append(crds, &obj)
		}
	}

	return crds, nil
}

func (store *managerStore) ListCustomResourceDefinition(kind, ns string) ([]*commtypes.Crd, error) {
	client := store.BkbcsClient.Crds(getCrdNamespace(kind, ns))
	v2Crds, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	crds := make([]*commtypes.Crd, 0, len(v2Crds.Items))
	for _, crd := range v2Crds.Items {
		obj := crd.Spec.Crd
		crds = append(crds, &obj)
	}

	return crds, nil
}

func (store *managerStore) FetchCustomResourceDefinition(kind, ns, name string) (*commtypes.Crd, error) {
	client := store.BkbcsClient.Crds(getCrdNamespace(kind, ns))
	v2Crd, err := client.Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &v2Crd.Spec.Crd, nil
}
