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

// Package keys is for the cache keys
package keys

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

var oneHourSeconds = 60 * 60
var oneDaySeconds = 24 * oneHourSeconds

// Key is an instance of the keyFactory
var Key = &keyGenerator{
	nullKeyTTLRange:             [2]int{60, 120},
	releasedGroupTTLRange:       [2]int{30 * 60, 60 * 60},
	credentialMatchedCITTLRange: [2]int{30 * 60, 60 * 60},
	credentialTTLRange:          [2]int{30 * 60, 60 * 60},
	releasedCITTLRange:          [2]int{6 * oneDaySeconds, 7 * oneDaySeconds},
	releasedHookTTLRange:        [2]int{6 * oneDaySeconds, 7 * oneDaySeconds},
	appMetaTTLRange:             [2]int{6 * oneDaySeconds, 7 * oneDaySeconds},
	appHasRITTLRange:            [2]int{5 * 60, 10 * 60},
}

type namespace string

const (
	cacheHead string = "bscp"

	releasedConfigItem  namespace = "released-ci"
	releasedHook        namespace = "released-hook"
	releasedGroup       namespace = "released-group"
	credentialMatchedCI namespace = "credential-matched-ci" //nolint:gosec
	credential          namespace = "credential"
	appMeta             namespace = "app-meta"
	appID               namespace = "app-id"
	releasedKv          namespace = "released-kv"
	clientMetric        namespace = "client-metric"
	publish             namespace = "publish"
)

type keyGenerator struct {
	nullKeyTTLRange             [2]int
	releasedGroupTTLRange       [2]int
	credentialMatchedCITTLRange [2]int
	credentialTTLRange          [2]int
	releasedCITTLRange          [2]int
	releasedKvTTLRange          [2]int
	releasedHookTTLRange        [2]int
	appMetaTTLRange             [2]int
	appHasRITTLRange            [2]int
}

// ClientMetricKey generate the client metric cache key.
func (k keyGenerator) ClientMetricKey(bizID uint32, appID uint32) string {
	return element{
		biz: bizID,
		ns:  clientMetric,
		key: strconv.FormatUint(uint64(appID), 10),
	}.String()
}

// ReleasedGroup generate a release's released group cache key to save all the released groups under this release
func (k keyGenerator) ReleasedGroup(bizID uint32, appID uint32) string {
	return element{
		biz: bizID,
		ns:  releasedGroup,
		key: strconv.FormatUint(uint64(appID), 10),
	}.String()
}

// ReleasedGroupTtlSec ReleasedCITtlSec generate the current released config item's TTL seconds
func (k keyGenerator) ReleasedGroupTtlSec(withRange bool) int {

	if withRange {
		//nolint:gosec
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		seconds := r.Intn(k.releasedGroupTTLRange[1]-k.releasedGroupTTLRange[0]) + k.releasedGroupTTLRange[0]
		return seconds
	}

	return k.releasedGroupTTLRange[1]
}

// CredentialMatchedCI generate a biz's credential matched ci key to save all the ci ids that matched by credential
func (k keyGenerator) CredentialMatchedCI(bizID uint32, credential string) string {
	return element{
		biz: bizID,
		ns:  credentialMatchedCI,
		key: credential,
	}.String()
}

// CredentialMatchedCITtlSec generate the credential matched ci's TTL seconds
func (k keyGenerator) CredentialMatchedCITtlSec(withRange bool) int {

	if withRange {
		//nolint:gosec
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		seconds := r.Intn(k.credentialMatchedCITTLRange[1]-
			k.credentialMatchedCITTLRange[0]) + k.credentialMatchedCITTLRange[0]
		return seconds
	}

	return k.credentialMatchedCITTLRange[1]
}

// Credential generate a biz's credential key to save the credential
func (k keyGenerator) Credential(bizID uint32, str string) string {
	return element{
		biz: bizID,
		ns:  credential,
		key: str,
	}.String()
}

// CredentialTtlSec generate the credential's TTL seconds
func (k keyGenerator) CredentialTtlSec(withRange bool) int {
	if withRange {
		//nolint:gosec
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		seconds := r.Intn(k.credentialTTLRange[1]-k.credentialTTLRange[0]) + k.credentialTTLRange[0]
		return seconds
	}
	return k.credentialTTLRange[0]
}

// ReleasedCI generate a release's CI cache key to save all the CIs under
// this release
func (k keyGenerator) ReleasedCI(bizID uint32, releaseID uint32) string {
	return element{
		biz: bizID,
		ns:  releasedConfigItem,
		key: strconv.FormatUint(uint64(releaseID), 10),
	}.String()
}

// ReleasedKv generate a release's Kv cache key to save all the Kvs under
// this release
func (k keyGenerator) ReleasedKv(bizID uint32, releaseID uint32) string {
	return element{
		biz: bizID,
		ns:  releasedKv,
		key: strconv.FormatUint(uint64(releaseID), 10),
	}.String()
}

// ReleasedCITtlSec generate the current released config item's TTL seconds
func (k keyGenerator) ReleasedCITtlSec(withRange bool) int {

	if withRange {
		//nolint:gosec
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		seconds := r.Intn(k.releasedCITTLRange[1]-k.releasedCITTLRange[0]) + k.releasedCITTLRange[0]
		return seconds
	}

	return k.releasedCITTLRange[1]
}

// ReleasedKvTtlSec generate the current released kv TTL seconds
func (k keyGenerator) ReleasedKvTtlSec(withRange bool) int {

	if withRange {
		//nolint:gosec
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		seconds := r.Intn(k.releasedKvTTLRange[1]-k.releasedKvTTLRange[0]) + k.releasedKvTTLRange[0]
		return seconds
	}

	return k.releasedKvTTLRange[1]
}

// ReleasedHook generate a release's hook cache key to save pre and post hook undert his release
func (k keyGenerator) ReleasedHook(bizID uint32, releaseID uint32) string {
	return element{
		biz: bizID,
		ns:  releasedHook,
		key: strconv.FormatUint(uint64(releaseID), 10),
	}.String()
}

// ReleasedHookTtlSec generate the current released hook's TTL seconds
func (k keyGenerator) ReleasedHookTtlSec(withRange bool) int {

	if withRange {
		//nolint:gosec
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		seconds := r.Intn(k.releasedHookTTLRange[1]-k.releasedHookTTLRange[0]) + k.releasedHookTTLRange[0]
		return seconds
	}

	return k.releasedHookTTLRange[1]
}

// AppMeta generate the app id cache key.
func (k keyGenerator) AppID(bizID uint32, appName string) string {
	return element{
		biz: bizID,
		ns:  appID,
		key: appName,
	}.String()
}

// AppMeta generate the app meta cache key.
func (k keyGenerator) AppMeta(bizID uint32, appID uint32) string {
	return element{
		biz: bizID,
		ns:  appMeta,
		key: strconv.FormatUint(uint64(appID), 10),
	}.String()
}

// PublishString generate the publish cache key.
func (k keyGenerator) PublishString(bizID, appID uint32) string {
	return element{
		biz: bizID,
		ns:  publish,
		key: strconv.FormatUint(uint64(appID), 10),
	}.PublishString()
}

// PublishPattern generate the publish pattern cache key.
func (k keyGenerator) PublishPattern() string {
	return element{
		ns: publish,
	}.PublishPattern()
}

// AppMetaTtlSec generate the app meta's TTL seconds
func (k keyGenerator) AppMetaTtlSec(withRange bool) int {

	if withRange {
		//nolint:gosec
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		seconds := r.Intn(k.appMetaTTLRange[1]-k.appMetaTTLRange[0]) + k.appMetaTTLRange[0]
		return seconds
	}

	return k.appMetaTTLRange[1]
}

// NullValue returns a value which means an empty cache value.
func (k keyGenerator) NullValue() string {
	return "NULL"
}

// NullKeyTtlSec return the null key's ttl seconds
func (k keyGenerator) NullKeyTtlSec() int {
	//nolint:gosec
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	seconds := r.Intn(k.nullKeyTTLRange[1]-k.nullKeyTTLRange[0]) + k.nullKeyTTLRange[0]
	return seconds
}

type element struct {
	// all the cache key is formatted with hashtag.
	biz uint32
	ns  namespace
	key string
}

// String format the element to a string
func (ele element) String() string {
	return fmt.Sprintf("{%d}%s:%s:%s", ele.biz, cacheHead, ele.ns, ele.key)
}

const (
	// FalseVal ..
	FalseVal = "0"
	// TrueVal ..
	TrueVal = "1"
)

// PublishString format the element publish to a string
func (ele element) PublishString() string {
	return fmt.Sprintf("%s:%s:%d:%s", cacheHead, ele.ns, ele.biz, ele.key)
}

// PublishPattern format the element publish pattern to a string
func (ele element) PublishPattern() string {
	return fmt.Sprintf("%s:%s*", cacheHead, ele.ns)
}
