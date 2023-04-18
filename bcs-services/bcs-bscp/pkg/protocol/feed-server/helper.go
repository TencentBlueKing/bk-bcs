/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package pbfs

import (
	"fmt"
)

// Validate the handshake message is valid or not.
func (x *HandshakeMessage) Validate() error {

	if err := x.ApiVersion.Validate(); err != nil {
		return fmt.Errorf("invalid api version, %v", err)
	}

	if x.Spec.BizId <= 0 {
		return fmt.Errorf("invalid biz id: %d", x.Spec.BizId)
	}

	if err := x.Spec.Version.Validate(); err != nil {
		return fmt.Errorf("invalid sidecar version, %v", err)
	}

	return nil
}
