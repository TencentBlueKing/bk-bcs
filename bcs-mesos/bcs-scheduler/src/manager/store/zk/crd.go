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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
)

func getCrrRootPath() string {
	return "/" + bcsRootNode + "/" + crrNode
}

func getCrdRootPath() string {
	return "/" + bcsRootNode + "/" + crdNode
}

//save custom resource register
func (store *managerStore) SaveCustomResourceRegister(crr *commtypes.Crr) error {
	by, err := json.Marshal(crr)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s/%s", getCrrRootPath(), crr.Spec.Names.Kind)
	return store.Db.Insert(key, string(by))
}

func (store *managerStore) DeleteCustomResourceRegister(name string) error {
	key := fmt.Sprintf("%s/%s", getCrrRootPath(), name)

	return store.Db.Delete(key)
}

func (store *managerStore) ListCustomResourceRegister() ([]*commtypes.Crr, error) {
	list, err := store.Db.List(getCrrRootPath())
	if err != nil {
		return nil, err
	}

	crrs := make([]*commtypes.Crr, 0)
	if list == nil {
		return crrs, nil
	}

	for _, name := range list {
		key := fmt.Sprintf("%s/%s", getCrrRootPath(), name)

		by, err := store.Db.Fetch(key)
		if err != nil {
			return nil, err
		}

		var crr *commtypes.Crr
		err = json.Unmarshal(by, &crr)
		if err != nil {
			return nil, err
		}

		crrs = append(crrs, crr)
	}

	return crrs, nil
}

func (store *managerStore) SaveCustomResourceDefinition(crd *commtypes.Crd) error {
	by, err := json.Marshal(crd)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s/%s/%s/%s", getCrdRootPath(), crd.Kind, crd.NameSpace, crd.Name)
	return store.Db.Insert(key, string(by))
}

func (store *managerStore) DeleteCustomResourceDefinition(kind, ns, name string) error {
	key := fmt.Sprintf("%s/%s/%s/%s", getCrdRootPath(), kind, ns, name)

	return store.Db.Delete(key)
}

func (store *managerStore) ListAllCrds(kind string) ([]*commtypes.Crd, error) {
	rootPath := fmt.Sprintf("/%s/crd/%s", bcsRootNode, kind)
	runAses, err := store.Db.List(rootPath)
	if err != nil {
		return nil, err
	}
	if nil == runAses {
		blog.Error("no runAs in (%s)", rootPath)
		return nil, nil
	}

	var crds []*commtypes.Crd
	for _, runAs := range runAses {
		crd, err := store.ListCustomResourceDefinition(kind, runAs)
		if err != nil {
			blog.Errorf("list crd(%s:%s) error %s", kind, runAs, err.Error())
			continue
		}

		crds = append(crds, crd...)
	}

	return crds, nil
}

func (store *managerStore) ListCustomResourceDefinition(kind, ns string) ([]*commtypes.Crd, error) {
	nsKey := fmt.Sprintf("%s/%s/%s", getCrdRootPath(), kind, ns)

	list, err := store.Db.List(nsKey)
	if err != nil {
		return nil, err
	}

	crds := make([]*commtypes.Crd, 0)
	if list == nil {
		return crds, nil
	}

	for _, name := range list {
		key := fmt.Sprintf("%s/%s", nsKey, name)

		by, err := store.Db.Fetch(key)
		if err != nil {
			return nil, err
		}

		var crd *commtypes.Crd
		err = json.Unmarshal(by, &crd)
		if err != nil {
			return nil, err
		}

		crds = append(crds, crd)
	}

	return crds, nil
}

func (store *managerStore) FetchCustomResourceDefinition(kind, ns, name string) (*commtypes.Crd, error) {
	key := fmt.Sprintf("%s/%s/%s/%s", getCrdRootPath(), kind, ns, name)
	by, err := store.Db.Fetch(key)
	if err != nil {
		return nil, err
	}

	var crd *commtypes.Crd
	err = json.Unmarshal(by, &crd)
	if err != nil {
		return nil, err
	}

	return crd, nil
}
