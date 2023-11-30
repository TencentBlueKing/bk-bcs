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

package keys

import "fmt"

// ResKind is instance of the resource kind generator
var ResKind = resKind("resource-kind")

type resKind string

// AppID return the app id's resource kind
func (rk resKind) AppID(bizID uint32, appName string) string {
	return fmt.Sprintf("app-id-%d-%s", bizID, appName)
}

// AppMeta return the app meta's resource kind
func (rk resKind) AppMeta(appID uint32) string {
	return fmt.Sprintf("apm-%d", appID)
}

// ReleasedCI return the released CI's resource kind
func (rk resKind) ReleasedCI(releaseID uint32) string {
	return fmt.Sprintf("rci-%d", releaseID)
}

// ReleasedKV return the released kv resource kind
func (rk resKind) ReleasedKV(releaseID uint32) string {
	return fmt.Sprintf("rkv-%d", releaseID)
}

// RKvValue return the released kv resource kind
func (rk resKind) RKvValue(releaseID uint32, key string) string {
	return fmt.Sprintf("rkv-value-%d-%s", releaseID, key)
}

// ReleasedHook return the released hook's resource kind
func (rk resKind) ReleasedHook(releaseID uint32) string {
	return fmt.Sprintf("rhook-%d", releaseID)
}

// CheckAppHasRI return the check app has released instance resource kind.
func (rk resKind) CheckAppHasRI(appID uint32) string {
	return fmt.Sprintf("app-cri-%d", appID)
}

// AppStrategy return the check app has released instance resource kind.
func (rk resKind) AppStrategy(appID uint32) string {
	return fmt.Sprintf("app-stg-%d", appID)
}

// ReleasedGroup return the released group resource kind.
func (rk resKind) ReleasedGroup(appID uint32) string {
	return fmt.Sprintf("released-group-%d", appID)
}

// ReleasedInstance return the released instance resource kind.
func (rk resKind) ReleasedInstance(appID uint32) string {
	return fmt.Sprintf("released-inst-%d", appID)
}

// CredentialMatchedCI return the credential matched ci resource kind.
func (rk resKind) CredentialMatchedCI(bizID uint32) string {
	return fmt.Sprintf("credential-matched-ci-%d", bizID)
}
