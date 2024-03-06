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

// Package utils implement simple utils
package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

// ToJsonString transfer any to json string, ignore error
func ToJsonString(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

// RemoveString remove string from slice
func RemoveString(strs []string, str string) []string {
	var newSlice []string
	for _, s := range strs {
		if s != str {
			newSlice = append(newSlice, s)
		}
	}
	return newSlice
}

// StringToInt str to int
// note: 可以使用 cast.ToInt() 替代
func StringToInt(s string) int {
	if s == "" {
		return 0
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}

	return i
}

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

// FormatRevision 格式化Revision
func FormatRevision(branch, commitId string) string {
	return fmt.Sprintf("%s@%s", branch, commitId)
}
