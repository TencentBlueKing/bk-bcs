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

// Package nspolicy xxx
package nspolicy

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-clusternet-controller/pkg/constant"
)

var (
	// ErrNoAvailableCluster error of no available cluster
	ErrNoAvailableCluster = errors.New("no available cluster")
	// ErrAnnotationNotFound error of annotation key not found
	ErrAnnotationNotFound = errors.New("annotation key not found")
)

// NamespacePolicy object represents policy for certain namespace
type NamespacePolicy struct {
	NsObject    *corev1.Namespace
	ClusterIDRe *regexp.Regexp
}

// NewNamespacePolicy create new namespace policy object
func NewNamespacePolicy(ns *corev1.Namespace) *NamespacePolicy {
	return &NamespacePolicy{
		NsObject:    ns,
		ClusterIDRe: regexp.MustCompile(constant.RegexClusterIDFormat),
	}
}

// GetAvailableClusterIDs get available clusterids from namespace annotation
func (np *NamespacePolicy) GetAvailableClusterIDs() ([]string, error) {
	clusterIDStr, ok := np.NsObject.Annotations[constant.NamespaceAnnotationKeyForClusterRange]
	if !ok {
		return nil, ErrNoAvailableCluster
	}
	clusterIDs := strings.Split(clusterIDStr, ",")
	for _, clusterID := range clusterIDs {
		if !np.ClusterIDRe.Match([]byte(clusterID)) {
			return nil, fmt.Errorf("annotation %s for ns %s is invalid, invalid part %s",
				constant.NamespaceAnnotationKeyForClusterRange, np.NsObject.GetName(), clusterID)
		}
	}
	return clusterIDs, nil
}

// GetClusterPriority get cluster priority from namespace annotation
func (np *NamespacePolicy) GetClusterPriority() (map[string]int64, error) {
	priorityStr, ok := np.NsObject.Annotations[constant.NamespaceAnnotationKeyForClusterPriority]
	if !ok {
		return nil, ErrAnnotationNotFound
	}
	priorityMap := make(map[string]int64)
	if err := json.Unmarshal([]byte(priorityStr), &priorityMap); err != nil {
		return nil, fmt.Errorf("failed to decode cluster priority, err %s", err.Error())
	}
	return priorityMap, nil
}
