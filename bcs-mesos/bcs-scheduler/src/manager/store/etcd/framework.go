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
	"context"

	"github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// FrameworkNode node name of mesos framework
	FrameworkNode = "frameworknode"
)

// CheckFrameworkIDExist check if framework id exists
func (store *managerStore) CheckFrameworkIDExist() (string, bool) {
	client := store.BkbcsClient.Frameworks(DefaultNamespace)
	v2Fw, err := client.Get(context.Background(), FrameworkNode, metav1.GetOptions{})
	if err == nil {
		return v2Fw.ResourceVersion, true
	}

	return "", false
}

// SaveFrameworkID save framework id
func (store *managerStore) SaveFrameworkID(frameworkId string) error {
	client := store.BkbcsClient.Frameworks(DefaultNamespace)
	v2Fw := &v2.Framework{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdFramework,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      FrameworkNode,
			Namespace: DefaultNamespace,
		},
		Spec: v2.FrameworkSpec{
			FrameworkId: frameworkId,
		},
	}

	var err error
	rv, exist := store.CheckFrameworkIDExist()
	if exist {
		v2Fw.ResourceVersion = rv
		_, err = client.Update(context.Background(), v2Fw, metav1.UpdateOptions{})
	} else {
		_, err = client.Create(context.Background(), v2Fw, metav1.CreateOptions{})
	}
	return err
}

// FetchFrameworkID get framework id
func (store *managerStore) FetchFrameworkID() (string, error) {
	client := store.BkbcsClient.Frameworks(DefaultNamespace)
	v2Fw, err := client.Get(context.Background(), FrameworkNode, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return v2Fw.Spec.FrameworkId, nil
}

// HasFrameworkID check if there is framework id
func (store *managerStore) HasFrameworkID() (bool, error) {
	_, err := store.FetchFrameworkID()
	if err != nil {
		return false, err
	}

	return true, nil
}
