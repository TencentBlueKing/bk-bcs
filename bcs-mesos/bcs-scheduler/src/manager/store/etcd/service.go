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
	commtypes "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-mesos/pkg/apis/bkbcs/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (store *managerStore) CheckServiceExist(service *commtypes.BcsService) (string, bool) {
	client := store.BkbcsClient.BcsServices(service.NameSpace)
	v2Svc, err := client.Get(service.Name, metav1.GetOptions{})
	if err == nil {
		return v2Svc.ResourceVersion, true
	}

	return "", false
}

func (store *managerStore) SaveService(service *commtypes.BcsService) error {
	err := store.checkNamespace(service.NameSpace)
	if err != nil {
		return err
	}

	client := store.BkbcsClient.BcsServices(service.NameSpace)
	v2Svc := &v2.BcsService{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdBcsService,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        service.Name,
			Namespace:   service.NameSpace,
			Labels:      store.filterSpecialLabels(service.Labels),
			Annotations: service.Annotations,
		},
		Spec: v2.BcsServiceSpec{
			BcsService: *service,
		},
	}

	rv, exist := store.CheckServiceExist(service)
	if exist {
		v2Svc.ResourceVersion = rv
		_, err = client.Update(v2Svc)
	} else {
		_, err = client.Create(v2Svc)
	}
	return err
}

func (store *managerStore) FetchService(ns, name string) (*commtypes.BcsService, error) {
	client := store.BkbcsClient.BcsServices(ns)
	v2Svc, err := client.Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &v2Svc.Spec.BcsService, nil
}

func (store *managerStore) DeleteService(ns, name string) error {
	client := store.BkbcsClient.BcsServices(ns)
	err := client.Delete(name, &metav1.DeleteOptions{})
	return err
}

func (store *managerStore) ListServices(runAs string) ([]*commtypes.BcsService, error) {
	client := store.BkbcsClient.BcsServices(runAs)
	v2Svcs, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	svcs := make([]*commtypes.BcsService, 0, len(v2Svcs.Items))
	for _, svc := range v2Svcs.Items {
		obj := svc.Spec.BcsService
		svcs = append(svcs, &obj)
	}
	return svcs, nil
}

func (store *managerStore) ListAllServices() ([]*commtypes.BcsService, error) {
	client := store.BkbcsClient.BcsServices("")
	v2Svcs, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	svcs := make([]*commtypes.BcsService, 0, len(v2Svcs.Items))
	for _, svc := range v2Svcs.Items {
		obj := svc.Spec.BcsService
		svcs = append(svcs, &obj)
	}
	return svcs, nil
}
