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

package zk

import (
	"encoding/json"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
)

func getEndpointRootPath() string {
	return "/" + bcsRootNode + "/" + endpointNode
}

func (store *managerStore) SaveEndpoint(endpoint *commtypes.BcsEndpoint) error {

	data, err := json.Marshal(endpoint)
	if err != nil {
		return err
	}

	path := getEndpointRootPath() + "/" + endpoint.ObjectMeta.NameSpace + "/" + endpoint.ObjectMeta.Name

	return store.Db.Insert(path, string(data))
}

func (store *managerStore) FetchEndpoint(ns, name string) (*commtypes.BcsEndpoint, error) {

	path := getEndpointRootPath() + "/" + ns + "/" + name

	data, err := store.Db.Fetch(path)
	if err != nil {
		return nil, err
	}

	endpoint := &commtypes.BcsEndpoint{}
	if err := json.Unmarshal(data, endpoint); err != nil {
		blog.Error("fail to unmarshal endpoint(%s). err:%s", string(data), err.Error())
		return nil, err
	}

	return endpoint, nil
}

func (store *managerStore) DeleteEndpoint(ns, name string) error {

	path := getEndpointRootPath() + "/" + ns + "/" + name
	if err := store.Db.Delete(path); err != nil {
		blog.Error("fail to delete endpoint(%s) err:%s", path, err.Error())
		return err
	}

	return nil
}
