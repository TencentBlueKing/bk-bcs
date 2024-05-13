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

package controller

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis/tkex/v1alpha1"
)

// RegisterValidateCreate register the webhook of DeschedulePolicy with Create event. It will check
// the namespace uniq and some params illegal.
func (m *ControllerManager) RegisterValidateCreate(policy *v1alpha1.DeschedulePolicy) error {
	blog.Infof("[Webhook] ValidatingCreate received: %s", policy.Name)
	// 当前设计中，集群仅允许一个全局 Policy，固定名称和命名空间
	if policy.Name != apis.DefaultPolicyName {
		return errors.Errorf("policy name must be '%s'", apis.DefaultPolicyName)
	}
	return m.validate(policy)
}

// RegisterValidateUpdate will check the params illegal of DeschedulePolicy
func (m *ControllerManager) RegisterValidateUpdate(policy *v1alpha1.DeschedulePolicy) error {
	blog.Infof("[Webhook] ValidateUpdate received: %s", policy.Name)
	return m.validate(policy)
}

func (m *ControllerManager) validate(policy *v1alpha1.DeschedulePolicy) error {
	strategy := &policy.Spec.Converge
	if strategy.TimeRange == "" {
		return errors.Errorf("converge timeRange cannot be empty")
	}
	schedule, err := cron.ParseStandard(strategy.TimeRange)
	if err != nil {
		return errors.Wrapf(err, "converge spec.converge.timeRange parse failed, timeRange=%s",
			strategy.TimeRange)
	}
	blog.Infof("[Webhook] %s/%s next time is '%s'", policy.Namespace, policy.Name,
		schedule.Next(time.Now()).Format("2006-01-02 15:04:05"))
	return nil
}
