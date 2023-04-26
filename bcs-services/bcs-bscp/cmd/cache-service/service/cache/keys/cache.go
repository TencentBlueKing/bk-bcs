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
	cpStrategyTTLRange:          [2]int{30 * 60, 60 * 60},
	releasedGroupTTLRange:       [2]int{30 * 60, 60 * 60},
	credentialMatchedCITTLRange: [2]int{30 * 60, 60 * 60},
	releasedCITTLRange:          [2]int{6 * oneDaySeconds, 7 * oneDaySeconds},
	releasedInstTTLRange:        [2]int{15 * 60, 30 * 60},
	appMetaTTLRange:             [2]int{6 * oneDaySeconds, 7 * oneDaySeconds},
	appHasRITTLRange:            [2]int{5 * 60, 10 * 60},
}

type namespace string

const (
	cacheHead string = "bscp"

	cpStrategy          namespace = "cp-strategy"
	releasedConfigItem  namespace = "released-ci"
	releasedGroup       namespace = "released-group"
	credentialMatchedCI namespace = "credential-matched-ci"
	appMeta             namespace = "app-meta"
	appID               namespace = "app-id"
)

type keyGenerator struct {
	nullKeyTTLRange             [2]int
	cpStrategyTTLRange          [2]int
	releasedGroupTTLRange       [2]int
	credentialMatchedCITTLRange [2]int
	releasedCITTLRange          [2]int
	releasedInstTTLRange        [2]int
	appMetaTTLRange             [2]int
	appHasRITTLRange            [2]int
}

// CPStrategy generate current published strategy's cache key
func (k keyGenerator) CPStrategy(bizID uint32, cpsID uint32) string {
	return element{
		biz: bizID,
		ns:  cpStrategy,
		key: strconv.FormatUint(uint64(cpsID), 10),
	}.String()
}

// CPStrategyTtlSec return the current published strategy's TTL seconds
func (k keyGenerator) CPStrategyTtlSec(withRange bool) int {

	if withRange {
		rand.Seed(time.Now().UnixNano())
		seconds := rand.Intn(k.cpStrategyTTLRange[1]-k.cpStrategyTTLRange[0]) + k.cpStrategyTTLRange[0]
		return seconds
	}

	return k.cpStrategyTTLRange[1]
}

// ReleasedGroup generate a release's released group cache key to save all the released groups under this release
func (k keyGenerator) ReleasedGroup(bizID uint32, appID uint32) string {
	return element{
		biz: bizID,
		ns:  releasedGroup,
		key: strconv.FormatUint(uint64(appID), 10),
	}.String()
}

// ReleasedCITtlSec generate the current released config item's TTL seconds
func (k keyGenerator) ReleasedGroupTtlSec(withRange bool) int {

	if withRange {
		rand.Seed(time.Now().UnixNano())
		seconds := rand.Intn(k.releasedGroupTTLRange[1]-k.releasedGroupTTLRange[0]) + k.releasedGroupTTLRange[0]
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
		rand.Seed(time.Now().UnixNano())
		seconds := rand.Intn(k.credentialMatchedCITTLRange[1]-
			k.credentialMatchedCITTLRange[0]) + k.credentialMatchedCITTLRange[0]
		return seconds
	}

	return k.credentialMatchedCITTLRange[1]
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

// ReleasedCITtlSec generate the current released config item's TTL seconds
func (k keyGenerator) ReleasedCITtlSec(withRange bool) int {

	if withRange {
		rand.Seed(time.Now().UnixNano())
		seconds := rand.Intn(k.releasedCITTLRange[1]-k.releasedCITTLRange[0]) + k.releasedCITTLRange[0]
		return seconds
	}

	return k.releasedCITTLRange[1]
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

// AppMetaTtlSec generate the app meta's TTL seconds
func (k keyGenerator) AppMetaTtlSec(withRange bool) int {

	if withRange {
		rand.Seed(time.Now().UnixNano())
		seconds := rand.Intn(k.appMetaTTLRange[1]-k.appMetaTTLRange[0]) + k.appMetaTTLRange[0]
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
	rand.Seed(time.Now().UnixNano())
	seconds := rand.Intn(k.nullKeyTTLRange[1]-k.nullKeyTTLRange[0]) + k.nullKeyTTLRange[0]
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
	// FalseVal TODO
	FalseVal = "0"
	// TrueVal TODO
	TrueVal = "1"
)
