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
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"strings"
	"unsafe"
)

// GzipEncode gzip 编码
func GzipEncode(bs []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)

	_, err := w.Write(bs)
	if err != nil {
		return nil, err
	}

	if err = w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GzipDecode gzip解码
func GzipDecode(encodedPlan []byte) ([]byte, error) {
	re := bytes.NewReader(encodedPlan)
	gr, err := gzip.NewReader(re)
	if err != nil {
		return nil, err
	}

	o, err := io.ReadAll(gr)
	if err != nil {
		return nil, err
	}

	if err = gr.Close(); err != nil {
		return nil, err
	}
	return o, nil
}

// SliceByteToString 字节转字符串
func SliceByteToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	// NOCC:gas/calls(设计如此)
	return *(*string)(unsafe.Pointer(&b))
}

// StringToSliceByte 字符串转字节
func StringToSliceByte(s string) []byte {
	// NOCC:gas/calls(设计如此)
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	// NOCC:gas/calls(设计如此)
	return *(*[]byte)(unsafe.Pointer(&h))
}

// StringsContainsOr string contains slice
func StringsContainsOr(s string, fields ...string) bool {
	for i := 0; i < len(fields); i++ {
		if strings.Contains(s, fields[i]) {
			return true
		}
	}
	return false
}

// CheckGitRepoSimilar check the git repo similar
func CheckGitRepoSimilar(original, compare string) bool {
	t1 := strings.Split(original, "/")
	t2 := strings.Split(compare, "/")
	originalRepo := t1[len(t1)-1]
	compareRepo := t2[len(t2)-1]
	return strings.TrimSuffix(originalRepo, ".git") == strings.TrimSuffix(compareRepo, ".git")
}

// MarshalObject marshal object
func MarshalObject(obj interface{}) string {
	bs, err := json.Marshal(obj)
	if err == nil {
		return string(bs)
	}
	return ""
}
