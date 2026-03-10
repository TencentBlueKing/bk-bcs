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
	"github.com/pkg/errors"
	v1 "k8s.io/api/admission/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func (s *Server) mutateIngress(ingress *networkextensionv1.Ingress, operation v1.Operation) ([]PatchOperation, error) {
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
		if operation == v1.Update && err == cloud.ErrLoadbalancerNotFound {
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
