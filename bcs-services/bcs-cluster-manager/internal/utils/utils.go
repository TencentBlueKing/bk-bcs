/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/kirito41dd/xslice"
)

// SplitAddrString split address string
func SplitAddrString(addrs string) []string {
	addrs = strings.Replace(addrs, ";", ",", -1)
	addrArray := strings.Split(addrs, ",")
	return addrArray
}

// GetXRequestIDFromHTTPRequest get X-Request-Id from http request
func GetXRequestIDFromHTTPRequest(req *http.Request) string {
	if req == nil {
		return ""
	}
	return req.Header.Get("X-Request-Id")
}

// RecoverPrintStack capture panic and print stack
func RecoverPrintStack(proc string) {
	if r := recover(); r != nil {
		blog.Errorf("[%s][recover] panic: %v, stack %v\n", proc, r, string(debug.Stack()))
		return
	}

	return
}

// StringInSlice returns true if given string in slice
func StringInSlice(s string, l []string) bool {
	for _, objStr := range l {
		if s == objStr {
			return true
		}
	}
	return false
}

// StringContainInSlice returns true if given string contain in slice
func StringContainInSlice(s string, l []string) bool {
	for _, objStr := range l {
		if strings.Contains(s, objStr) {
			return true
		}
	}
	return false
}


// IntInSlice return true if i in l
func IntInSlice(i int, l []int) bool {
	for _, obj := range l {
		if i == obj {
			return true
		}
	}
	return false
}

// SplitStringsChunks split strings chunk
func SplitStringsChunks(strList []string, limit int) [][]string {
	if limit <= 0 || len(strList) == 0 {
		return nil
	}
	i := xslice.SplitToChunks(strList, limit)
	ss, ok := i.([][]string)
	if !ok {
		return nil
	}

	return ss
}
