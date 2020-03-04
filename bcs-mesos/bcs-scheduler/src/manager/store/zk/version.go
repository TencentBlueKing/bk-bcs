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
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"encoding/json"
	"sort"
	"strconv"
	"time"
)

func getVersionRootPath() string {
	return "/" + bcsRootNode + "/" + versionNode + "/"
}

//create version, produce version id
func (store *managerStore) SaveVersion(version *types.Version) error {
	version.Name = strconv.FormatInt(time.Now().UnixNano(), 10)
	data, err := json.Marshal(version)
	if err != nil {
		blog.Error("fail to encode object version(ID:%s) by json. err:%s", version.ID, err.Error())
		return err
	}

	runAs := version.RunAs
	if "" == runAs {
		runAs = defaultRunAs
	}

	path := getVersionRootPath() + runAs + "/" + version.ID + "/" + version.Name
	return store.Db.Insert(path, string(data))
}

func (store *managerStore) UpdateVersion(version *types.Version) error {
	data, err := json.Marshal(version)
	if err != nil {
		blog.Error("fail to encode object version(ID:%s) by json. err:%s", version.ID, err.Error())
		return err
	}

	runAs := version.RunAs
	if "" == runAs {
		runAs = defaultRunAs
	}

	path := getVersionRootPath() + runAs + "/" + version.ID + "/" + version.Name

	return store.Db.Insert(path, string(data))
}

func (store *managerStore) ListVersions(runAs, versionID string) ([]string, error) {

	if "" == runAs {
		runAs = defaultRunAs
	}

	path := getVersionRootPath() + runAs + "/" + versionID
	blog.V(3).Infof("list versions from (%s)", path)

	return store.Db.List(path)
}

func (store *managerStore) FetchVersion(runAs, versionId, versionNo string) (*types.Version, error) {

	if "" == runAs {
		runAs = defaultRunAs
	}

	path := getVersionRootPath() + runAs + "/" + versionId + "/" + versionNo
	blog.V(3).Infof("get version from (%s)", path)

	data, err := store.Db.Fetch(path)
	if err != nil {
		blog.Error("fail to get version(ID:%s, versionNo(%s)). err:%s", versionId, versionNo, err.Error())
		return nil, err
	}

	var version types.Version
	if err := json.Unmarshal(data, &version); err != nil {
		blog.Error("fail to unmarshal version. version(%s), err:%s", path, err.Error())
		return nil, err
	}

	return &version, nil
}

func (store *managerStore) DeleteVersion(runAs, versionId, versionNo string) error {

	if "" == runAs {
		runAs = defaultRunAs
	}

	path := getVersionRootPath() + runAs + "/" + versionId + "/" + versionNo

	if err := store.Db.Delete(path); err != nil {
		blog.Error("fail to delete version, version id(%s) No(%s), err:%s", versionId, versionNo, err.Error())
		return err
	}

	return nil
}

func (store *managerStore) DeleteVersionNode(runAs, versionId string) error {

	if "" == runAs {
		runAs = defaultRunAs
	}

	path := getVersionRootPath() + runAs + "/" + versionId

	if err := store.Db.Delete(path); err != nil {
		blog.Error("fail to delete version node, version id(%s), err:%s", versionId, err.Error())
		return err
	}

	return nil
}

func (store *managerStore) GetVersion(runAs, appId string) (*types.Version, error) {

	versions, err := store.ListVersions(runAs, appId)
	if err != nil {
		blog.Error("fail to list versions. err:%s", err.Error())
		return nil, err
	}

	if len(versions) != 0 {
		sort.Strings(versions)

		newestVersion, err := store.FetchVersion(runAs, appId, versions[len(versions)-1])
		if err != nil {
			blog.Error("fail to fetch version by runAs(%s), appId(%s), versionNo(%s). err:%s", runAs, appId, versions[len(versions)-1], err.Error())
			return nil, err
		}
		return newestVersion, nil
	}

	return nil, nil
}
