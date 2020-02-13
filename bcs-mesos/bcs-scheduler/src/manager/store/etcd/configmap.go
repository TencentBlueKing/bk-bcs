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
	schStore "bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"bk-bcs/bcs-mesos/pkg/apis/bkbcs/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (store *managerStore) CheckConfigMapExist(configmap *commtypes.BcsConfigMap) (string, bool) {
	v2Cfg, err := store.FetchConfigMap(configmap.NameSpace, configmap.Name)
	if err == nil {
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
			Name:        configmap.Name,
			Namespace:   configmap.NameSpace,
			Labels:      store.filterSpecialLabels(configmap.Labels),
			Annotations: configmap.Annotations,
		},
		Spec: v2.BcsConfigMapSpec{
			BcsConfigMap: *configmap,
		},
	}

	rv, exist := store.CheckConfigMapExist(configmap)
	if exist {
		v2Cfg.ResourceVersion = rv
		v2Cfg, err = client.Update(v2Cfg)
	} else {
		v2Cfg, err = client.Create(v2Cfg)
	}

	configmap.ResourceVersion = v2Cfg.ResourceVersion
	saveCacheConfigmap(configmap)
	return err
}

func (store *managerStore) FetchConfigMap(ns, name string) (*commtypes.BcsConfigMap, error) {
	if cacheMgr.isOK {
		cfg := getCacheConfigmap(ns, name)
		if cfg == nil {
			return nil, schStore.ErrNoFound
		}
		return cfg, nil
	}

	client := store.BkbcsClient.BcsConfigMaps(ns)
	v2Cfg, err := client.Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	obj := v2Cfg.Spec.BcsConfigMap
	obj.ResourceVersion = v2Cfg.ResourceVersion
	return &obj, nil
}

func (store *managerStore) DeleteConfigMap(ns, name string) error {
	client := store.BkbcsClient.BcsConfigMaps(ns)
	err := client.Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	deleteCacheConfigmap(ns, name)
	return nil
}

func (store *managerStore) ListConfigmaps(runAs string) ([]*commtypes.BcsConfigMap, error) {
	client := store.BkbcsClient.BcsConfigMaps(runAs)
	v2Cfg, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	cfgs := make([]*commtypes.BcsConfigMap, 0, len(v2Cfg.Items))
	for _, cfg := range v2Cfg.Items {
		obj := cfg.Spec.BcsConfigMap
		obj.ResourceVersion = cfg.ResourceVersion
		cfgs = append(cfgs, &obj)
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
		obj := cfg.Spec.BcsConfigMap
		obj.ResourceVersion = cfg.ResourceVersion
		cfgs = append(cfgs, &obj)
	}
	return cfgs, nil
}
