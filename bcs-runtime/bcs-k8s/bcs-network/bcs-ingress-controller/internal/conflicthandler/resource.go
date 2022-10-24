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

package conflicthandler

import (
	mapset "github.com/deckarep/golang-set"
)

// portSegment [start,end)
type portSegment struct {
	Start int
	End   int
}

func (p *portSegment) contains(port int) bool {
	if port >= p.Start && port < p.End {
		return true
	}

	return false
}

func (p *portSegment) intersect(ps portSegment) bool {
	if max(p.Start, ps.Start) < min(p.End, ps.End) {
		return true
	}
	return false
}

type resource struct {
	usedPort        mapset.Set // set[int]
	usedPortSegment []portSegment
}

func newResource() *resource {
	return &resource{
		usedPort:        mapset.NewThreadUnsafeSet(),
		usedPortSegment: make([]portSegment, 0),
	}
}

// IsConflict return true if conflict with otherRes
func (r *resource) IsConflict(otherRes *resource) bool {
	if otherRes == nil {
		return false
	}

	for port := range r.usedPort.Iter() {
		if otherRes.usedPort.Contains(port) {
			return true
		}

		for _, otherPortSeg := range otherRes.usedPortSegment {
			if otherPortSeg.contains(port.(int)) {
				return true
			}
		}
	}

	for _, portSeg := range r.usedPortSegment {
		for otherPort := range otherRes.usedPort.Iter() {
			if portSeg.contains(otherPort.(int)) {
				return true
			}
		}

		for _, otherPortSeg := range otherRes.usedPortSegment {
			if portSeg.intersect(otherPortSeg) {
				return true
			}
		}
	}

	return false
}
