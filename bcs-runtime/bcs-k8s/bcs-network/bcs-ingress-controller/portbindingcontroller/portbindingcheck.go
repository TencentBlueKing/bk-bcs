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

package portbindingcontroller

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

// checkPortBindingCreate check if related portbinding create successfully
func checkPortBindingCreate(cli client.Client, namespace, name string) {
	blog.Infof("starts to check related portbinding %s/%s status", namespace, name)
	timeout := time.After(time.Minute)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			blog.Warnf("portbinding '%s/%s' is not ready, inc fail metric", namespace, name)
			metrics.IncreaseFailMetric(metrics.ObjectPortbinding, metrics.EventTypeAdd)
			return
		case <-ticker.C:
			portBinding := &networkextensionv1.PortBinding{}
			err := cli.Get(context.TODO(), types.NamespacedName{
				Namespace: namespace,
				Name:      name,
			}, portBinding)
			if err != nil {
				if k8serrors.IsNotFound(err) {
					blog.V(5).Infof("not found portbinding '%s/%s' related to created pod", namespace, name)
					continue
				}
				blog.Warnf("failed to get portbinding '%s/%s' related to created pod: %s", namespace, name,
					err.Error())
				continue
			}

			if portBinding.Status.Status == constant.PortBindingStatusReady {
				blog.Infof("portbinding '%s/%s' is ready", namespace, name)
				return
			}
		}
	}
}

// checkPortBindingDelete check if related portbinding delete successfully
func checkPortBindingDelete(cli client.Client, namespace, name string) {
	blog.Infof("starts to check portbinding %s/%s clean", namespace, name)
	metrics.CleanPortAllocateMetric(name, namespace)
	timeout := time.After(time.Minute)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			blog.Warnf("portbinding '%s/%s' clean not finished, inc fail metric", namespace, name)
			metrics.IncreaseFailMetric(metrics.ObjectPortbinding, metrics.EventTypeDelete)
			return
		case <-ticker.C:
			portBinding := &networkextensionv1.PortBinding{}
			err := cli.Get(context.TODO(), types.NamespacedName{
				Namespace: namespace,
				Name:      name,
			}, portBinding)
			if err != nil {
				if k8serrors.IsNotFound(err) {
					blog.Infof("portbinding '%s/%s' clean finish", namespace, name)
					return
				}
				blog.Warnf("failed to get portbinding '%s/%s' related to created pod: %s", namespace, name,
					err.Error())
				continue
			}
		}
	}
}
