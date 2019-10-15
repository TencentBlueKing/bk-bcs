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

func (store *managerStore) CheckSecretExist(secret *commtypes.BcsSecret) (string, bool) {
	client := store.BkbcsClient.BcsSecrets(secret.NameSpace)
	v2Sec, _ := client.Get(secret.Name, metav1.GetOptions{})
	if v2Sec != nil {
		return v2Sec.ResourceVersion, true
	}

	return "", false
}

func (store *managerStore) SaveSecret(secret *commtypes.BcsSecret) error {
	err := store.checkNamespace(secret.NameSpace)
	if err != nil {
		return err
	}

	client := store.BkbcsClient.BcsSecrets(secret.NameSpace)
	v2Sec := &v2.BcsSecret{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdBcsSecret,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name,
			Namespace: secret.NameSpace,
		},
		Spec: v2.BcsSecretSpec{
			BcsSecret: *secret,
		},
	}

	rv, exist := store.CheckSecretExist(secret)
	if exist {
		v2Sec.ResourceVersion = rv
		_, err = client.Update(v2Sec)
	} else {
		_, err = client.Create(v2Sec)
	}
	return err
}

func (store *managerStore) FetchSecret(ns, name string) (*commtypes.BcsSecret, error) {
	client := store.BkbcsClient.BcsSecrets(ns)
	v2Sec, err := client.Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &v2Sec.Spec.BcsSecret, nil
}

func (store *managerStore) DeleteSecret(ns, name string) error {
	client := store.BkbcsClient.BcsSecrets(ns)
	err := client.Delete(name, &metav1.DeleteOptions{})
	return err
}

func (store *managerStore) ListSecrets(runAs string) ([]*commtypes.BcsSecret, error) {
	client := store.BkbcsClient.BcsSecrets(runAs)
	v2Secs, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	secrets := make([]*commtypes.BcsSecret, 0, len(v2Secs.Items))
	for _, sec := range v2Secs.Items {
		secrets = append(secrets, &sec.Spec.BcsSecret)
	}

	return secrets, nil
}

func (store *managerStore) ListAllSecrets() ([]*commtypes.BcsSecret, error) {
	client := store.BkbcsClient.BcsSecrets("")
	v2Secs, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	secrets := make([]*commtypes.BcsSecret, 0, len(v2Secs.Items))
	for _, sec := range v2Secs.Items {
		secrets = append(secrets, &sec.Spec.BcsSecret)
	}

	return secrets, nil
}
