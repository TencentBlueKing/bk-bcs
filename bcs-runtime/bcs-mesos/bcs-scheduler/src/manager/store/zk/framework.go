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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

func (store *managerStore) SaveFrameworkID(frameworkId string) error {

	framework := &types.Framework{ID: frameworkId}
	data, err := json.Marshal(framework)
	if err != nil {
		blog.Error("fail to encode object framework by json. err:%s", err.Error())
		return err
	}

	path := "/" + bcsRootNode + "/" + frameWorkNode

	return store.Db.Insert(path, string(data))
}

func (store *managerStore) FetchFrameworkID() (string, error) {

	path := "/" + bcsRootNode + "/" + frameWorkNode

	data, err := store.Db.Fetch(path)
	if err != nil {
		blog.Error("fail to get framework id, err(%s)", err.Error())
		return "", err
	}

	framework := &types.Framework{}
	if err := json.Unmarshal(data, framework); err != nil {
		blog.Error("fail to unmarshal framework(%s), err(%s)", string(data), err.Error())
		return "", err
	}

	return framework.ID, nil
}

func (store *managerStore) HasFrameworkID() (bool, error) {
	_, err := store.FetchFrameworkID()
	if err != nil {
		return false, err
	}

	return true, nil
}
