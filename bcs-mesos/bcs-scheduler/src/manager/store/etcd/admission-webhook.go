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

	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckAdmissionWebhookExist check if the admission webhook exists
func (store *managerStore) CheckAdmissionWebhookExist(
	admission *commtypes.AdmissionWebhookConfiguration) (string, bool) {
	admission.NameSpace = commtypes.DefaultAdmissionNamespace
	client := store.BkbcsClient.AdmissionWebhookConfigurations(admission.NameSpace)
	obj, err := client.Get(context.Background(), admission.Name, metav1.GetOptions{})
	if err == nil {
		return obj.ResourceVersion, true
	}

	return "", false
}

// SaveAdmissionWebhook save admission webhook into db
func (store *managerStore) SaveAdmissionWebhook(admission *commtypes.AdmissionWebhookConfiguration) error {
	admission.NameSpace = commtypes.DefaultAdmissionNamespace
	err := store.checkNamespace(admission.NameSpace)
	if err != nil {
		return err
	}

	client := store.BkbcsClient.AdmissionWebhookConfigurations(admission.NameSpace)
	v2Admission := &v2.AdmissionWebhookConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdAdmissionWebhookConfiguration,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        admission.Name,
			Namespace:   admission.NameSpace,
			Labels:      admission.Labels,
			Annotations: admission.Annotations,
		},
		Spec: v2.AdmissionWebhookConfigurationSpec{
			AdmissionWebhookConfiguration: *admission,
		},
	}

	rv, exist := store.CheckAdmissionWebhookExist(admission)
	if exist {
		v2Admission.ResourceVersion = rv
		_, err = client.Update(context.Background(), v2Admission, metav1.UpdateOptions{})
	} else {
		_, err = client.Create(context.Background(), v2Admission, metav1.CreateOptions{})
	}
	return err
}

// FetchAdmissionWebhook get admission webhook
func (store *managerStore) FetchAdmissionWebhook(ns, name string) (*commtypes.AdmissionWebhookConfiguration, error) {
	ns = commtypes.DefaultAdmissionNamespace
	client := store.BkbcsClient.AdmissionWebhookConfigurations(ns)
	v2Admission, err := client.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &v2Admission.Spec.AdmissionWebhookConfiguration, nil
}

// DeleteAdmissionWebhook delete admission webhook
func (store *managerStore) DeleteAdmissionWebhook(ns, name string) error {
	ns = commtypes.DefaultAdmissionNamespace
	client := store.BkbcsClient.AdmissionWebhookConfigurations(ns)
	err := client.Delete(context.Background(), name, metav1.DeleteOptions{})
	return err
}

// FetchAllAdmissionWebhooks fetch all admission webhooks
func (store *managerStore) FetchAllAdmissionWebhooks() ([]*commtypes.AdmissionWebhookConfiguration, error) {
	client := store.BkbcsClient.AdmissionWebhookConfigurations("")
	v2Admissions, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	admissions := make([]*commtypes.AdmissionWebhookConfiguration, 0, len(v2Admissions.Items))
	for _, v2 := range v2Admissions.Items {
		obj := v2.Spec.AdmissionWebhookConfiguration
		admissions = append(admissions, &obj)
	}
	return admissions, nil
}
