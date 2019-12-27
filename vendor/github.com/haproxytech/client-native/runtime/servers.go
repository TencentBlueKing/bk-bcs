// Copyright 2019 HAProxy Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package runtime

import (
	"fmt"
)

//SetServerAddr set ip [port] for server
func (s *SingleRuntime) SetServerAddr(backend, server string, ip string, port int) error {
	var cmd string
	if port > 0 {
		cmd = fmt.Sprintf("set server %s/%s addr %s port %d", backend, server, ip, port)
	} else {
		cmd = fmt.Sprintf("set server %s/%s addr %s", backend, server, ip)
	}
	return s.Execute(cmd)
}

//SetServerState set state for server
func (s *SingleRuntime) SetServerState(backend, server string, state string) error {
	if !ServerStateValid(state) {
		return fmt.Errorf("bad request")
	}
	cmd := fmt.Sprintf("set server %s/%s state %s", backend, server, state)
	return s.Execute(cmd)
}

//SetServerWeight set weight for server
func (s *SingleRuntime) SetServerWeight(backend, server string, weight string) error {
	if !ServerWeightValid(weight) {
		return fmt.Errorf("bad request")
	}
	cmd := fmt.Sprintf("set server %s/%s weight %s", backend, server, weight)
	return s.Execute(cmd)
}

//SetServerHealth set health for server
func (s *SingleRuntime) SetServerHealth(backend, server string, health string) error {
	if !ServerHealthValid(health) {
		return fmt.Errorf("bad request")
	}
	cmd := fmt.Sprintf("set server %s/%s health %s", backend, server, health)
	return s.Execute(cmd)
}
