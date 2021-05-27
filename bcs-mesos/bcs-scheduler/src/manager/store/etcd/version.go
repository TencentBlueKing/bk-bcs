/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http:// opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package etcd

import (
	"context"
	"sort"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	schStore "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const VersionIdKey = "VersionId"

// SaveVersion create version, produce version id
func (store *managerStore) SaveVersion(version *types.Version) error {
	version.Name = strconv.FormatInt(time.Now().UnixNano(), 10)
	runAs := version.RunAs
	if "" == runAs {
		runAs = defaultRunAs
		version.RunAs = defaultRunAs
	}
	err := store.checkNamespace(runAs)
	if err != nil {
		return err
	}

	client := store.BkbcsClient.Versions(runAs)
	v2Version := &v2.Version{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdVersion,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      version.Name,
			Namespace: runAs,
		},
		Spec: v2.VersionSpec{
			Version: *version,
		},
	}
	_, err = client.Create(context.Background(), v2Version, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	saveCacheVersion(version.RunAs, version.ID, version)
	return err
}

// UpdateVersion update version
func (store *managerStore) UpdateVersion(version *types.Version) error {
	runAs := version.RunAs
	if "" == runAs {
		runAs = defaultRunAs
		version.RunAs = defaultRunAs
	}
	err := store.checkNamespace(runAs)
	if err != nil {
		return err
	}

	client := store.BkbcsClient.Versions(runAs)
	v2Version := &v2.Version{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdVersion,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      version.Name,
			Namespace: runAs,
		},
		Spec: v2.VersionSpec{
			Version: *version,
		},
	}
	_, err = client.Create(context.Background(), v2Version, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	saveCacheVersion(version.RunAs, version.ID, version)

	return err
}

// ListVersions list versions
func (store *managerStore) ListVersions(runAs, versionID string) ([]string, error) {
	versions, _ := listCacheVersions(runAs, versionID)
	nodes := make([]string, 0, len(versions))
	for _, version := range versions {
		nodes = append(nodes, version.Name)
	}
	if len(nodes) == 0 {
		blog.Warnf("fetch version(%s.%s) is empty", runAs, versionID)
		return nil, nil
	}
	return nodes, nil
}

func (store *managerStore) listClusterVersions() ([]*types.Version, error) {
	client := store.BkbcsClient.Versions("")
	v2Versions, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodes := make([]*types.Version, 0, len(v2Versions.Items))
	for _, version := range v2Versions.Items {
		obj := version.Spec.Version
		nodes = append(nodes, &obj)
	}
	return nodes, nil
}

// FetchVersion get version
func (store *managerStore) FetchVersion(runAs, versionId, versionNo string) (*types.Version, error) {
	version, _ := getCacheVersion(runAs, versionId, versionNo)
	if version == nil {
		return nil, schStore.ErrNoFound
	}
	return version, nil
}

// DeleteVersion delete version
func (store *managerStore) DeleteVersion(runAs, versionId, versionNo string) error {
	if "" == runAs {
		runAs = defaultRunAs
	}
	client := store.BkbcsClient.Versions(runAs)
	err := client.Delete(context.Background(), versionNo, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

// DeleteVersionNode zk stores this method by removing the version path, and etcd stores in order to be consistent
// with the store interface, this method does not implement anything
func (store *managerStore) DeleteVersionNode(runAs, versionId string) error {
	deleteCacheVersion(runAs, versionId)
	return nil
}

// GetVersion get version
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
