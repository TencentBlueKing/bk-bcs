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
 */

package clusterops

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteValidatingWebhookConfig delete validating webhook config
func (ko *K8SOperator) DeleteValidatingWebhookConfig(ctx context.Context, clusterId, name string) error {
	if ko == nil {
		return ErrServerNotInit
	}

	if len(name) == 0 {
		return fmt.Errorf("webhook name empty")
	}

	clientInterface, err := ko.GetClusterClient(clusterId)
	if err != nil {
		blog.Errorf("DeleteValidatingWebhookConfig[%s] GetClusterClient failed: %v", clusterId, err)
		return err
	}

	_, err = clientInterface.AdmissionregistrationV1().ValidatingWebhookConfigurations().
		Get(ctx, name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		blog.Errorf("DeleteValidatingWebhookConfig[%s:%s] delete validate webhook failed: %v",
			clusterId, name, err)
		return err
	}
	if errors.IsNotFound(err) {
		blog.Infof("DeleteValidatingWebhookConfig[%s:%s] notfound", clusterId, name)
		return nil
	}

	err = clientInterface.AdmissionregistrationV1().ValidatingWebhookConfigurations().Delete(ctx,
		name, metav1.DeleteOptions{})
	if err != nil {
		blog.Errorf("DeleteValidatingWebhookConfig[%s:%s] failed: %v", clusterId, name, err)
		return err
	}
	blog.Infof("DeleteValidatingWebhookConfig[%s:%s] success", clusterId, name)

	return nil
}
