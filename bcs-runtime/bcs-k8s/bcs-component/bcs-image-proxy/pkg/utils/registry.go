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

package utils

import (
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var (
	manifestUriRegexp = regexp.MustCompile(`^/v[1-2]/(.*)/manifests/(.*)`)
	blobUriRegexp     = regexp.MustCompile(`^/v[1-2]/(.*)/blobs/sha256:([a-z0-9A-Z]{64})$`)
)

// IsManifestGet used to check the uri whether is manifest-get
// e.p: /v2/tencentmirrors/centos/manifests/7 => tencentmirrors/centos, 7, nil
func IsManifestGet(r *http.Request) (string, string, bool) {
	if r.Method != http.MethodGet {
		return "", "", false
	}
	if r.URL == nil {
		return "", "", false
	}
	result := manifestUriRegexp.FindStringSubmatch(r.URL.Path)
	if len(result) != 3 {
		return "", "", false
	}
	repo := result[1]
	tag := result[2]
	return repo, tag, true
}

// IsBlobGet used to check the uri whether is blob-download
// e.p: /v2/instantlinux/haproxy-keepalived/blobs/sha256:ec99f8b99825a742d50fb3ce173d291378a46ab54b8ef7dd75e5654e2a296e99
// => instantlinux/haproxy-keepalived, ec99f8b99825a742d50fb3ce173d291378a46ab54b8ef7dd75e5654e2a296e99
func IsBlobGet(url string) (string, string, bool) {
	result := blobUriRegexp.FindStringSubmatch(url)
	if len(result) != 3 {
		return "", "", false
	}
	repo := result[1]
	sha256 := result[2]
	return repo, sha256, true
}

// LayerFileName return layer name
func LayerFileName(digest string) string {
	digest = strings.TrimPrefix(digest, "sha256:")
	return digest + ".tar.gzip"
}

const (
	B  = "B"
	KB = "KB"
	MB = "MB"
	GB = "GB"
	TB = "TB"
	EB = "EB"
)

var (
	// byteUnit 字节单位
	byteUnit = []string{B, KB, MB, GB, TB, EB}
)

// FormatSize 字节的单位转换 保留两位小数
func FormatSize(s int64) string {
	var b int64 = 1
	for i := 0; i < len(byteUnit); i++ {
		b = b << 10
		if s >= b {
			continue
		}
		res, _ := FormatFloat(float64(s)/float64(b>>10), 2)
		return fmt.Sprintf("%.2f%-2s", res, byteUnit[i])
	}
	return fmt.Sprintf("%.2f%-2s", float64(s), B) // 未找到单位
}

// FormatFloat  保留两位小数，舍弃尾数，无进位运算
func FormatFloat(num float64, decimal int) (float64, error) {
	d := float64(1)
	if decimal > 0 {
		d = math.Pow10(decimal)
	}
	res := strconv.FormatFloat(math.Trunc(num*d)/d, 'f', -1, 64)
	return strconv.ParseFloat(res, 64)
}
