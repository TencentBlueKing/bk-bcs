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
	"context"
	"fmt"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/hostnetportpoolcache"
)

// validateHostNetPortPool validates a HostNetPortPool's own fields and checks whether its
// port range overlaps any other HostNetPortPool in the cluster. Overlapping ranges are
// rejected because HostNetPortPool has no nodeSelector: any node may host pods referencing
// either pool, so overlapping ranges could hand the same physical port to different pods.
func (s *Server) validateHostNetPortPool(pool *networkextensionv1.HostNetPortPool) error {
	if err := checkHostNetPortPoolFields(pool); err != nil {
		return err
	}
	return s.checkHostNetPortPoolConflict(pool)
}

// checkHostNetPortPoolFields validates the intrinsic fields of a single pool.
func checkHostNetPortPoolFields(pool *networkextensionv1.HostNetPortPool) error {
	if pool == nil {
		return fmt.Errorf("HostNetPortPool must not be nil")
	}
	start := int(pool.Spec.StartPort)
	end := int(pool.Spec.EndPort)
	segLen := int(pool.Spec.SegmentLength)

	if start >= end {
		return fmt.Errorf("startPort %d must be less than endPort %d", start, end)
	}
	if segLen < 1 {
		return fmt.Errorf("segmentLength %d must be at least 1", segLen)
	}
	if end-start < segLen {
		return fmt.Errorf("port range [%d, %d) is smaller than segmentLength %d, no segment can be allocated",
			start, end, segLen)
	}
	return nil
}

// checkHostNetPortPoolConflict lists all HostNetPortPools cluster-wide and rejects the
// incoming pool if its [startPort, endPort) range overlaps with any other pool.
func (s *Server) checkHostNetPortPoolConflict(pool *networkextensionv1.HostNetPortPool) error {
	poolList := &networkextensionv1.HostNetPortPoolList{}
	if err := s.k8sClient.List(context.Background(), poolList); err != nil {
		return fmt.Errorf("list HostNetPortPool for conflict check failed: %s", err.Error())
	}

	newStart := int(pool.Spec.StartPort)
	newEnd := int(pool.Spec.EndPort)
	for i := range poolList.Items {
		other := &poolList.Items[i]
		// skip self
		if other.GetName() == pool.GetName() && other.GetNamespace() == pool.GetNamespace() {
			continue
		}
		if hostnetportpoolcache.RangesOverlap(
			newStart, newEnd, int(other.Spec.StartPort), int(other.Spec.EndPort)) {
			return fmt.Errorf(
				"port range [%d, %d) overlaps with HostNetPortPool %s/%s range [%d, %d)",
				newStart, newEnd, other.GetNamespace(), other.GetName(),
				other.Spec.StartPort, other.Spec.EndPort)
		}
	}
	return nil
}
