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
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/admission/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
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

	// 对于可以通过修改ingress以外配置使ingress正常运行的错误，返回warning
	// k8s 1.18之前不支持warning，暂时使用annotation方式
	// ingress controller reconcile过程中会更新该annotation
	var warnings []string
	warnings = append(warnings, s.ingressConverter.CheckIngressServiceAvailable(ingress)...)

	_, err := s.ingressConverter.GetIngressLoadBalancers(ingress)
	if err != nil {
		// 避免lb被删除后导致ingress无法正常更新
		if operation == v1.Update && err == cloud.ErrLoadbalancerNotFound {
			return nil, nil
		}
		warnings = append(warnings, err.Error())
	}

	err = s.conflictHandler.IsIngressConflict(ingress)
	if err != nil {
		return nil, err
	}

	var retPatches []PatchOperation
	annotationWarningPatch := s.generateWarningAnnotationPatch(ingress, warnings)
	retPatches = append(retPatches, annotationWarningPatch)

	return retPatches, nil
}

// generate warning annotations
func (s *Server) generateWarningAnnotationPatch(ingress *networkextensionv1.Ingress, warnings []string) PatchOperation {
	annotations := ingress.Annotations
	op := constant.PatchOperationReplace
	if len(annotations) == 0 {
		op = constant.PatchOperationAdd
		annotations = make(map[string]string)
	}
	annotations[networkextensionv1.AnnotationKeyForWarnings] = strings.Join(warnings, ";")
	return PatchOperation{
		Path:  constant.PatchPathPodAnnotations,
		Op:    op,
		Value: annotations,
	}
}
