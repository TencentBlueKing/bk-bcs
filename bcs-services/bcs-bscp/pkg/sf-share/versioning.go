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

package sfs

import (
	pbbase "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
)

// CurrentAPIVersion is the current api version used between sidecar and feed server.
var CurrentAPIVersion = &pbbase.Versioning{
	Major: 1,
	Minor: 0,
	Patch: 0,
}

// leastAPIVersion is the least sidecar's api version that this feed server can work for.
var leastAPIVersion = &pbbase.Versioning{
	Major: 1,
	Minor: 0,
	Patch: 0,
}

// IsAPIVersionMatch test if the sidecar's version match the
// feed server's version request.
func IsAPIVersionMatch(ver *pbbase.Versioning) bool {

	if ver == nil {
		return false
	}

	if ver.Major < leastAPIVersion.Major {
		return false
	}

	if ver.Major == leastAPIVersion.Major {
		if ver.Minor < leastAPIVersion.Minor {
			return false
		}

		if ver.Minor == leastAPIVersion.Minor {
			if ver.Patch < leastAPIVersion.Patch {
				return false
			}
		}

		return true
	}

	return true

}

// leastSidecarVersion is the least sidecar's version that this feed server can work for.
var leastSidecarVersion = &pbbase.Versioning{
	Major: 1,
	Minor: 0,
	Patch: 0,
}

// IsSidecarVersionMatch test if the sidecar's version match the
// feed server's version request.
func IsSidecarVersionMatch(ver *pbbase.Versioning) bool {

	if ver.Major < leastSidecarVersion.Major {
		return false
	}

	if ver.Major == leastSidecarVersion.Major {
		if ver.Minor < leastSidecarVersion.Minor {
			return false
		}

		if ver.Minor == leastSidecarVersion.Minor {
			if ver.Patch < leastSidecarVersion.Patch {
				return false
			}
		}

		return true
	}

	return true

}
