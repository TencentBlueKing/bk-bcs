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
	commtypes "bk-bcs/bcs-common/common/types"
	"fmt"
)

//save admission webhook
func (b *backend) SaveAdmissionWebhook(admission *commtypes.AdmissionWebhookConfiguration) error {
	admissions, err := b.FetchAllAdmissionWebhooks()
	if err != nil {
		return err
	}
	for _, currAdmission := range admissions {
		if currAdmission.ResourcesRef.Operation == admission.ResourcesRef.Operation &&
			currAdmission.ResourcesRef.Kind == admission.ResourcesRef.Kind {
			return fmt.Errorf("AdmissionWebhookConfiguration Operation %s Kind %s conflict with "+
				"AdmissionWebhook(%s:%s)", admission.ResourcesRef.Operation, admission.ResourcesRef.Kind,
				currAdmission.NameSpace, currAdmission.Name)
		}
	}

	return b.store.SaveAdmissionWebhook(admission)
}

func (b *backend) UpdateAdmissionWebhook(admission *commtypes.AdmissionWebhookConfiguration) error {
	return b.store.SaveAdmissionWebhook(admission)
}

func (b *backend) FetchAdmissionWebhook(ns, name string) (*commtypes.AdmissionWebhookConfiguration, error) {
	return b.store.FetchAdmissionWebhook(ns, name)
}

func (b *backend) DeleteAdmissionWebhook(ns string, name string) error {
	return b.store.DeleteAdmissionWebhook(ns, name)
}

func (b *backend) FetchAllAdmissionWebhooks() ([]*commtypes.AdmissionWebhookConfiguration, error) {
	return b.store.FetchAllAdmissionWebhooks()
}
