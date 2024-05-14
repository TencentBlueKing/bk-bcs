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
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	k8scorev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
)

// inject port pool item info into pod annotations and envs
func (s *Server) mutatingNode(node *k8scorev1.Node) ([]PatchOperation, error) {
	if node.Annotations == nil {
		return nil, fmt.Errorf("node '%s/%s' annotation is nil", node.GetNamespace(), node.GetName())
	}
	// get port info that should be injected
	annotationValue, ok := node.Annotations[constant.AnnotationForPortPoolPorts]
	if !ok {
		return nil, fmt.Errorf("node '%s/%s' lack annotation key %s", node.GetNamespace(), node.GetName(),
			constant.AnnotationForPortPoolPorts)
	}
	// Node webhook会处理Creat&Update两种事件， 所以当已有端口信息时需要忽略
	// injected node should not be injected again
	if _, pok := node.Annotations[constant.AnnotationForPortPoolBindings]; pok {
		return nil, nil
	}
	annotationPorts, err := parserAnnotation(annotationValue)
	if err != nil {
		return nil, fmt.Errorf("node '%s/%s' parse annotation failed, err %s", node.GetNamespace(),
			node.GetName(), err.Error())
	}
	portEntryList, err := s.getPortEntryListFromNode(annotationPorts)
	if err != nil {
		return nil, fmt.Errorf("get port entry list from node %s/%s failed, err %s",
			node.GetNamespace(), node.GetName(), err.Error())
	}

	// check for existed port binding
	portBinding, err := s.checkExistedPortBinding(node.GetName(), node.GetNamespace(), portEntryList, node.GetAnnotations())
	if err != nil {
		return nil, errors.Wrapf(err, "check portbinding existed failed")
	}
	if portBinding != nil {
		blog.Infof("node '%s/%s' reuse portbinding", node.GetNamespace(), node.GetName())
		return nil, errors.Errorf("node '%s/%s' cannot reuse portBinding", node.GetNamespace(), node.GetName())
	}

	blog.Infof("node '%s/%s' do port inject", node.GetNamespace(), node.GetName())

	portPoolItemStatusList, portItemListArr, err := s.portAllocate(portEntryList)
	if err != nil {
		return nil, err
	}

	// patch annotations
	var retPatches []PatchOperation
	annotationPortsPatch, err := s.generatePortsAnnotationPatch(
		node.Annotations, portPoolItemStatusList, portItemListArr, portEntryList)
	if err != nil {
		s.cleanAllocatedResource(portItemListArr)
		return nil, fmt.Errorf("generate ports of port pool annotations failed, err %s", err.Error())
	}
	retPatches = append(retPatches, annotationPortsPatch)

	for _, itemList := range portItemListArr {
		for _, item := range itemList {
			blog.Infof("node '%s/%s' allocate port from cache: %v", node.GetNamespace(), node.GetName(), item)
		}
	}

	return retPatches, nil
}

// combine container port info into port entry format
func (s *Server) getPortEntryListFromNode(annotationPorts []*annotationPort) ([]*portEntry, error) {
	var retPorts []*portEntry
	for _, port := range annotationPorts {
		tmpEntry := &portEntry{
			PoolNamespace: port.poolNamespace,
			PoolName:      port.poolName,
			Protocol:      port.protocol,
			HostPort:      port.hostPort,
			ItemName:      port.itemName,
		}
		if len(port.poolNamespace) == 0 {
			// node's namespace is empty, use 'default'
			// tmpEntry.poolNamespace = node.GetNamespace()
			tmpEntry.PoolNamespace = s.nodePortBindingNs
		}
		if len(port.protocol) == 0 {
			return nil, errors.New("empty protocol is invalid for node")
		}
		portNumber, err := strconv.Atoi(port.portIntOrStr)
		if err != nil {
			return nil, errors.New("only numeric port is valid for node")
		}
		tmpEntry.Port = portNumber
		tmpEntry.HostPort = true // node port always host port

		retPorts = append(retPorts, tmpEntry)
	}
	return retPorts, nil
}
