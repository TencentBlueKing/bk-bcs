/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1

import (
	"context"
	"errors"
	"fmt"
	"net"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/internal/constant"
)

type bcsNetPoolClient struct {
	client client.Client
}

// SetupWebhookWithManager setup webhook for BCSPool with manager
func (r *BCSNetPool) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(&bcsNetPoolClient{client: mgr.GetClient()}).
		WithValidator(&bcsNetPoolClient{client: mgr.GetClient()}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-netservice-bkbcs-tencent-com-v1-bcsnetpool,mutating=true,failurePolicy=fail,sideEffects=None,groups=netservice.bkbcs.tencent.com,resources=bcsnetpools,verbs=create;update,versions=v1,name=mbcsnetpool.kb.io,admissionReviewVersions=v1

var _ admission.CustomDefaulter = &bcsNetPoolClient{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (c *bcsNetPoolClient) Default(ctx context.Context, obj runtime.Object) error {
	return nil
}

//+kubebuilder:webhook:path=/validate-netservice-bkbcs-tencent-com-v1-bcsnetpool,mutating=false,failurePolicy=fail,sideEffects=None,groups=netservice.bkbcs.tencent.com,resources=bcsnetpools,verbs=create;update;delete,versions=v1,name=vbcsnetpool.kb.io,admissionReviewVersions=v1

var _ admission.CustomValidator = &bcsNetPoolClient{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (c *bcsNetPoolClient) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	pool, ok := obj.(*BCSNetPool)
	if !ok {
		return errors.New("object is not BCSNetPool")
	}

	blog.Infof("validate create pool %s", pool.Name)
	if net.ParseIP(pool.Spec.Net) == nil {
		return fmt.Errorf("spec.net %s is not valid when creating bcsnetpool %s", pool.Spec.Net, pool.Name)
	}
	if net.ParseIP(pool.Spec.Gateway) == nil {
		return fmt.Errorf("spec.gateway is not valid %s when creating bcsnetpool %s", pool.Spec.Gateway, pool.Name)
	}
	for _, ip := range pool.Spec.AvailableIPs {
		if net.ParseIP(ip) == nil {
			return fmt.Errorf("%s in spec.availableIPs is not valid when creating bcsnetpool %s", ip, pool.Name)
		}
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (c *bcsNetPoolClient) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) error {
	pool, ok := newObj.(*BCSNetPool)
	if !ok {
		return errors.New("object is not BCSNetPool")
	}
	oldPool, ok2 := oldObj.(*BCSNetPool)
	if !ok2 {
		return errors.New("object is not BCSNetPool")
	}

	blog.Infof("validate update pool %s", pool.Name)
	if net.ParseIP(pool.Spec.Net) == nil {
		return fmt.Errorf("spec.net %s is not valid when updating bcsnetpool %s", pool.Spec.Net, pool.Name)
	}
	if net.ParseIP(pool.Spec.Gateway) == nil {
		return fmt.Errorf("spec.gateway is not valid %s when updating bcsnetpool %s", pool.Spec.Gateway, pool.Name)
	}
	for _, ip := range pool.Spec.AvailableIPs {
		if net.ParseIP(ip) == nil {
			return fmt.Errorf("%s in spec.availableIPs is not valid when updating bcsnetpool %s", ip, pool.Name)
		}
	}

	// 找出更新操作中要删除的已存在的IP
	var delIPList []string
	newIPMap := make(map[string]bool)
	for _, v := range pool.Spec.AvailableIPs {
		newIPMap[v] = true
	}

	for _, v := range oldPool.Spec.AvailableIPs {
		if _, exists := newIPMap[v]; !exists {
			delIPList = append(delIPList, v)
		}
	}

	if err := c.checkActiveIP(ctx, delIPList, pool); err != nil {
		return err
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (c *bcsNetPoolClient) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	return nil
}

func (c *bcsNetPoolClient) checkActiveIP(ctx context.Context, s []string, pool *BCSNetPool) error {
	for _, ip := range s {
		netIP := &BCSNetIP{}
		if err := c.client.Get(ctx, types.NamespacedName{Name: ip}, netIP); err != nil {
			if k8serrors.IsNotFound(err) {
				blog.Warnf("BCSNetIP %s missing in pool %s", ip, pool.Name)
				continue
			}
			return err
		}
		if netIP.Status.Phase == constant.BCSNetIPActiveStatus {
			return fmt.Errorf("can not perform operation for pool %s, active IP %s exists", pool.Name, ip)
		}
	}
	return nil
}
