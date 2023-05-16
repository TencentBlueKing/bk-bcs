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

package backend

import (
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
)

// SaveConfigMap xxx
func (b *backend) SaveConfigMap(configmap *commtypes.BcsConfigMap) error {

	return b.store.SaveConfigMap(configmap)
}

// FetchConfigMap xxx
func (b *backend) FetchConfigMap(ns, name string) (*commtypes.BcsConfigMap, error) {
	return b.store.FetchConfigMap(ns, name)
}

// DeleteConfigMap xxx
func (b *backend) DeleteConfigMap(ns string, name string) error {
	return b.store.DeleteConfigMap(ns, name)
}

// SaveSecret xxx
func (b *backend) SaveSecret(secret *commtypes.BcsSecret) error {

	return b.store.SaveSecret(secret)
}

// FetchSecret xxx
func (b *backend) FetchSecret(ns, name string) (*commtypes.BcsSecret, error) {
	return b.store.FetchSecret(ns, name)
}

// DeleteSecret xxx
func (b *backend) DeleteSecret(ns string, name string) error {
	return b.store.DeleteSecret(ns, name)
}

// SaveService xxx
func (b *backend) SaveService(service *commtypes.BcsService) error {

	b.sched.ServiceMgr.ServiceUpdate(service)

	return b.store.SaveService(service)
}

// FetchService xxx
func (b *backend) FetchService(ns, name string) (*commtypes.BcsService, error) {
	return b.store.FetchService(ns, name)
}

// DeleteService xxx
func (b *backend) DeleteService(ns string, name string) error {

	service := new(commtypes.BcsService)
	service.ObjectMeta.NameSpace = ns
	service.ObjectMeta.Name = name
	b.sched.ServiceMgr.ServiceDelete(service)

	return b.store.DeleteService(ns, name)
}
