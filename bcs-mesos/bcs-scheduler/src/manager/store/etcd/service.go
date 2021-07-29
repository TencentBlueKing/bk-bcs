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
	"context"

	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	schStore "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckServiceExist check if service exists
func (store *managerStore) CheckServiceExist(service *commtypes.BcsService) (string, bool) {
	svc, _ := store.fetchServiceInDB(service.NameSpace, service.Name)
	if svc != nil {
		return svc.ResourceVersion, true
	}

	return "", false
}

// SaveService save service to db
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
		v2Svc, err = client.Update(context.Background(), v2Svc, metav1.UpdateOptions{})
	} else {
		v2Svc, err = client.Create(context.Background(), v2Svc, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	service.ResourceVersion = v2Svc.ResourceVersion
	saveCacheService(service)
	return err
}

// FetchService get service by name and namespace
func (store *managerStore) FetchService(ns, name string) (*commtypes.BcsService, error) {
	svc := getCacheService(ns, name)
	if svc == nil {
		return svc, schStore.ErrNoFound
	}
	return svc, nil
}

func (store *managerStore) fetchServiceInDB(ns, name string) (*commtypes.BcsService, error) {
	client := store.BkbcsClient.BcsServices(ns)
	v2Svc, err := client.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	obj := v2Svc.Spec.BcsService
	obj.ResourceVersion = v2Svc.ResourceVersion
	return &obj, nil
}

// DeleteService delete service by name and namespace
func (store *managerStore) DeleteService(ns, name string) error {
	client := store.BkbcsClient.BcsServices(ns)
	err := client.Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	deleteCacheService(ns, name)
	return nil
}

// ListAllServices list all services
func (store *managerStore) ListAllServices() ([]*commtypes.BcsService, error) {
	if cacheMgr.isOK {
		return listCacheServices()
	}

	client := store.BkbcsClient.BcsServices("")
	v2Svcs, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	svcs := make([]*commtypes.BcsService, 0, len(v2Svcs.Items))
	for _, svc := range v2Svcs.Items {
		obj := svc.Spec.BcsService
		obj.ResourceVersion = svc.ResourceVersion
		svcs = append(svcs, &obj)
	}
	return svcs, nil
}
