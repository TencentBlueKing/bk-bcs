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

package iregex

import (
	"fmt"
	"regexp"
	"sync"
)

var (
	regexMu = sync.RWMutex{}
	// Cache for regex object.
	// Note that:
	// 1. It uses sync.RWMutex ensuring the concurrent safety.
	// 2. There's no expiring logic for this map.
	regexMap = make(map[string]*regexp.Regexp)
)

// getRegexp returns *regexp.Regexp object with given `pattern`.
// It uses cache to enhance the performance for compiling regular expression pattern,
// which means, it will return the same *regexp.Regexp object with the same regular
// expression pattern.
//
// It is concurrent-safe for multiple goroutines.
func getRegexp(pattern string) (regex *regexp.Regexp, err error) {
	// Retrieve the regular expression object using reading lock.
	regexMu.RLock()
	regex = regexMap[pattern]
	regexMu.RUnlock()
	if regex != nil {
		return
	}
	// If it does not exist in the cache,
	// it compiles the pattern and creates one.
	if regex, err = regexp.Compile(pattern); err != nil {
		err = fmt.Errorf(`regexp.Compile failed for pattern "%s" "%s"`, pattern, err.Error())
		return
	}
	// Cache the result object using writing lock.
	regexMu.Lock()
	regexMap[pattern] = regex
	regexMu.Unlock()
	return
}
