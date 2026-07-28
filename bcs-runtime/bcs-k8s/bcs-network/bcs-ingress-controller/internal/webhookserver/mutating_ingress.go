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

package webhookserver

import (
	stderrors "errors"
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/admission/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func (s *Server) mutateIngress(ingress, oldIngress *networkextensionv1.Ingress, operation v1.Operation) (
	[]PatchOperation, error) {
	// SNI 一旦开启无法在线关闭（腾讯云约束）。更新时若把某端口的 SNI 从开启改为关闭，
	// 直接拒绝并提示用户删除该规则/监听器后重建，避免产生无法生效的静默配置。
	if operation == v1.Update {
		if ok, msg := checkSNINotDisabledOnUpdate(oldIngress, ingress); !ok {
			return nil, errors.New(msg)
		}
	}
	// 对于必须修改ingress配置的错误，返回errResponse
	isValid, msg := s.ingressValidater.IsIngressValid(ingress)
	if !isValid {
		return nil, errors.New(msg)
	}
	isValid, msg = s.ingressValidater.CheckNoConflictsInIngress(ingress)
	if !isValid {
		return nil, errors.New(msg)
	}

	_, err := s.ingressConverter.GetIngressLoadBalancers(ingress)
	if err != nil {
		// 避免lb被删除后导致ingress无法正常更新
		if operation == v1.Update && stderrors.Is(err, cloud.ErrLoadbalancerNotFound) {
			return nil, nil
		}
		return nil, err
	}

	err = s.conflictHandler.IsIngressConflict(ingress)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// checkSNINotDisabledOnUpdate rejects disabling SNI (1->0) on an existing HTTPS listener.
// Tencent Cloud CLB does not support disabling SNI once enabled; the listener must be
// deleted and recreated. This only compares the Ingress spec (old vs new) and cannot
// detect SNI enabled out-of-band on the CLB console.
func checkSNINotDisabledOnUpdate(oldIngress, newIngress *networkextensionv1.Ingress) (bool, string) {
	if oldIngress == nil {
		return true, ""
	}
	oldSNIOn := make(map[int]bool)
	for _, rule := range oldIngress.Spec.Rules {
		if rule.ListenerAttribute != nil && rule.ListenerAttribute.SniSwitch != 0 {
			oldSNIOn[rule.Port] = true
		}
	}
	for _, rule := range newIngress.Spec.Rules {
		if !oldSNIOn[rule.Port] {
			continue
		}
		newSNIOn := rule.ListenerAttribute != nil && rule.ListenerAttribute.SniSwitch != 0
		if !newSNIOn {
			return false, fmt.Sprintf("cannot disable SNI on port %d: Tencent Cloud CLB does not support disabling "+
				"SNI once enabled; delete the rule/listener and recreate it with sniSwitch:0", rule.Port)
		}
	}
	return true, ""
}
