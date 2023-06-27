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

package constant

// Note:
// This scope is used to define all the constant keys which is used inside and outside
// the BSCP system except sidecar.
const (
	// KitKey
	KitKey = "X-BSCP-KIT"

	// RidKey is request id header key.
	RidKey = "X-Bkapi-Request-Id"
	// RidKeyGeneric for generic header key
	RidKeyGeneric = "X-Request-Id"

	// UserKey is operator name header key.
	UserKey = "X-Bkapi-User-Name"

	// AppCodeKey is blueking application code header key.
	AppCodeKey = "X-Bkapi-App-Code"

	// Space
	SpaceIDKey     = "X-Bkapi-Space-Id"
	SpaceTypeIDKey = "X-Bkapi-Space-Type-Id"
	BizIDKey       = "X-Bkapi-Biz-Id"

	// LanguageKey the language key word.
	LanguageKey = "HTTP_BLUEKING_LANGUAGE"

	// BKGWJWTTokenKey is blueking api gateway jwt header key.
	BKGWJWTTokenKey = "X-Bkapi-JWT" // nolint

	// BKTokenForTest is a token for test
	BKTokenForTest = "bk-token-for-test"

	// BKUserForTestPrefix is a user prefix for test
	BKUserForTestPrefix = "bk-user-for-test-"

	// BKSystemUser can be saved for user field in db when some operations come from bscp system itself
	BKSystemUser = "system"

	// ContentIDHeaderKey is common content sha256 id.
	ContentIDHeaderKey = "X-Bkapi-File-Content-Id"
)

// Note:
// 1. This scope defines keys which is used only by sidecar and feed server.
// 2. All the defined key should be prefixed with 'Side'.
const (
	// SidecarMetaKey defines the key to store the sidecar's metadata info.
	SidecarMetaKey = "sidecar-meta"
	// SideRidKey defines the incoming request id between sidecar and feed server.
	SideRidKey = "side-rid"
	// SideUserKey defines the sidecar's user key.
	SideUserKey = "side-user"
	// SideWorkspaceDir sidecar workspace dir name.
	SideWorkspaceDir = "bk-bscp"
)

const (
	AuthLoginProviderKey = "auth-login-provider"
	AuthLoginUID         = "auth-login-uid"
	AuthLoginToken       = "auth-login-token" // nolint
)

var (
	// RidKeys support request_id keys
	RidKeys = []string{
		RidKey,
		RidKeyGeneric,
	}
)
