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

package conflicthandler

// portSegment [start,end)
type portSegment struct {
	Start     int
	End       int
	Protocols []string
}

// return true if conflict
func (ps *portSegment) isConflictWithPort(isTCPUDPReuse bool, p portStruct) bool {
	if p.val >= ps.Start && p.val < ps.End {
		if isProtocolConflict(isTCPUDPReuse, ps.Protocols, p.Protocols) {
			return true
		}
	}

	return false
}

// return true if conflict
func (ps *portSegment) isConflictWithPortSeg(isTCPUDPReuse bool, newPs portSegment) bool {
	if max(ps.Start, newPs.Start) < min(ps.End, newPs.End) {
		if isProtocolConflict(isTCPUDPReuse, ps.Protocols, newPs.Protocols) {
			return true
		}
	}

	return false
}

type portStruct struct {
	val       int
	Protocols []string
}

// return true if conflict
func (p *portStruct) isConflictWithPort(isTCPUDPReuse bool, newPort portStruct) bool {
	if p.val == newPort.val {
		if isProtocolConflict(isTCPUDPReuse, p.Protocols, newPort.Protocols) {
			return true
		}
	}

	return false
}

// return true if conflict
func (p *portStruct) isConflictWithPortSeg(isTCPUDPReuse bool, ps portSegment) bool {
	return ps.isConflictWithPort(isTCPUDPReuse, *p)
}

type resource struct {
	usedPort        map[int]portStruct
	usedPortSegment []portSegment
}

func newResource() *resource {
	return &resource{
		usedPort:        make(map[int]portStruct),
		usedPortSegment: make([]portSegment, 0),
	}
}

// IsConflict return true if conflict with otherRes
func (r *resource) IsConflict(isTCPUDPReuse bool, otherRes *resource) bool {
	if otherRes == nil {
		return false
	}

	for portVal, port := range r.usedPort {
		if otherPort, ok := otherRes.usedPort[portVal]; ok {
			if port.isConflictWithPort(isTCPUDPReuse, otherPort) {
				return true
			}
		}

		for _, otherPortSeg := range otherRes.usedPortSegment {
			if port.isConflictWithPortSeg(isTCPUDPReuse, otherPortSeg) {
				return true
			}
		}
	}

	for _, portSeg := range r.usedPortSegment {
		for _, port := range otherRes.usedPort {
			if portSeg.isConflictWithPort(isTCPUDPReuse, port) {
				return true
			}
		}

		for _, otherPortSeg := range otherRes.usedPortSegment {
			if portSeg.isConflictWithPortSeg(isTCPUDPReuse, otherPortSeg) {
				return true
			}
		}
	}

	return false
}
