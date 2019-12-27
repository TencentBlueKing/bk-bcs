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
	"strconv"
)

//SetFrontendMaxConn set maxconn for frontend
func (s *SingleRuntime) SetFrontendMaxConn(frontend string, maxconn int) error {
	cmd := fmt.Sprintf("set maxconn frontend %s %s", frontend, strconv.FormatInt(int64(maxconn), 10))
	return s.Execute(cmd)
}
