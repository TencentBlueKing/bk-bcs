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

func (store *managerStore) CheckConfigMapExist(configmap *commtypes.BcsConfigMap) (string, bool) {
	client := store.BkbcsClient.BcsConfigMaps(configmap.NameSpace)
	v2Cfg, _ := client.Get(configmap.Name, metav1.GetOptions{})
	if v2Cfg != nil {
		return v2Cfg.ResourceVersion, true
	}

	return "", false
}

func (store *managerStore) SaveConfigMap(configmap *commtypes.BcsConfigMap) error {
	err := store.checkNamespace(configmap.NameSpace)
	if err != nil {
		return err
	}

	client := store.BkbcsClient.BcsConfigMaps(configmap.NameSpace)
	v2Cfg := &v2.BcsConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdBcsConfigMap,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmap.Name,
			Namespace: configmap.NameSpace,
		},
		Spec: v2.BcsConfigMapSpec{
			BcsConfigMap: *configmap,
		},
	}

	rv, exist := store.CheckConfigMapExist(configmap)
	if exist {
		v2Cfg.ResourceVersion = rv
		_, err = client.Update(v2Cfg)
	} else {
		_, err = client.Create(v2Cfg)
	}
	return err
}

func (store *managerStore) FetchConfigMap(ns, name string) (*commtypes.BcsConfigMap, error) {
	client := store.BkbcsClient.BcsConfigMaps(ns)
	v2Cfg, err := client.Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &v2Cfg.Spec.BcsConfigMap, nil
}

func (store *managerStore) DeleteConfigMap(ns, name string) error {
	client := store.BkbcsClient.BcsConfigMaps(ns)
	err := client.Delete(name, &metav1.DeleteOptions{})
	return err
}

func (store *managerStore) ListConfigmaps(runAs string) ([]*commtypes.BcsConfigMap, error) {
	client := store.BkbcsClient.BcsConfigMaps(runAs)
	v2Cfg, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	cfgs := make([]*commtypes.BcsConfigMap, 0, len(v2Cfg.Items))
	for _, cfg := range v2Cfg.Items {
		cfgs = append(cfgs, &cfg.Spec.BcsConfigMap)
	}
	return cfgs, nil
}

func (store *managerStore) ListAllConfigmaps() ([]*commtypes.BcsConfigMap, error) {
	client := store.BkbcsClient.BcsConfigMaps("")
	v2Cfg, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	cfgs := make([]*commtypes.BcsConfigMap, 0, len(v2Cfg.Items))
	for _, cfg := range v2Cfg.Items {
		cfgs = append(cfgs, &cfg.Spec.BcsConfigMap)
	}
	return cfgs, nil
}
