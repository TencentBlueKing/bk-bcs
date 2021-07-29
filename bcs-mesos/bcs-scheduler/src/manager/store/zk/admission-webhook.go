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

func getAdmissionWebhookRootPath() string {
	return "/" + bcsRootNode + "/" + AdmissionWebhookNode
}

func (store *managerStore) SaveAdmissionWebhook(admission *commtypes.AdmissionWebhookConfiguration) error {

	data, err := json.Marshal(admission)
	if err != nil {
		return err
	}

	path := getAdmissionWebhookRootPath() + "/" + admission.ObjectMeta.NameSpace + "/" + admission.ObjectMeta.Name
	return store.Db.Insert(path, string(data))
}

func (store *managerStore) FetchAdmissionWebhook(ns, name string) (*commtypes.AdmissionWebhookConfiguration, error) {

	path := getAdmissionWebhookRootPath() + "/" + ns + "/" + name

	data, err := store.Db.Fetch(path)
	if err != nil {
		return nil, err
	}

	admission := &commtypes.AdmissionWebhookConfiguration{}
	if err := json.Unmarshal(data, admission); err != nil {
		blog.Error("fail to unmarshal admission(%s). err:%s", string(data), err.Error())
		return nil, err
	}

	return admission, nil
}

func (store *managerStore) DeleteAdmissionWebhook(ns, name string) error {

	path := getAdmissionWebhookRootPath() + "/" + ns + "/" + name
	if err := store.Db.Delete(path); err != nil {
		blog.Error("fail to delete admission(%s) err:%s", path, err.Error())
		return err
	}

	return nil
}

func (store *managerStore) FetchAllAdmissionWebhooks() ([]*commtypes.AdmissionWebhookConfiguration, error) {
	namespaces, err := store.Db.List(getAdmissionWebhookRootPath())
	if err != nil {
		return nil, err
	}

	admissions := make([]*commtypes.AdmissionWebhookConfiguration, 0)
	for _, ns := range namespaces {
		nsPath := fmt.Sprintf("%s/%s", getAdmissionWebhookRootPath(), ns)
		hookNames, err := store.Db.List(nsPath)
		if err != nil {
			blog.Errorf("zk list %s error %s", nsPath, err.Error())
			continue
		}

		for _, name := range hookNames {
			hookPath := fmt.Sprintf("%s/%s", nsPath, name)
			data, err := store.Db.Fetch(hookPath)
			if err != nil {
				blog.Errorf("zk fetch %s error %s", hookPath, err.Error())
				continue
			}

			var admission *commtypes.AdmissionWebhookConfiguration
			err = json.Unmarshal(data, &admission)
			if err != nil {
				blog.Errorf("unmarshal data %s to commtypes.AdmissionWebhookConfiguration error %s",
					string(data), err.Error())
				continue
			}

			admissions = append(admissions, admission)
		}
	}

	return admissions, nil
}
