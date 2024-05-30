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
 *
 */

package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"

	"github.com/pkg/errors"
)

// GzipAndBase64String will gzip the string, and base64 encode it
func GzipAndBase64String(str string) ([]byte, error) {
	var buf bytes.Buffer
	gWriter := gzip.NewWriter(&buf)
	if _, err := gWriter.Write([]byte(str)); err != nil {
		return nil, errors.Wrapf(err, "gzip failed")
	}
	_ = gWriter.Close()

	bufBS := buf.Bytes()
	base64Enc := base64.StdEncoding
	dst := make([]byte, base64Enc.EncodedLen(len(bufBS)))
	base64Enc.Encode(dst, bufBS)
	return dst, nil
}

// GzipAndBase64Bytes will gzip the bytes and base64 encode it
func GzipAndBase64Bytes(bs []byte) ([]byte, error) {
	var buf bytes.Buffer
	gWriter := gzip.NewWriter(&buf)
	if _, err := gWriter.Write(bs); err != nil {
		return nil, errors.Wrapf(err, "gzip failed")
	}
	_ = gWriter.Close()

	bufBS := buf.Bytes()
	base64Enc := base64.StdEncoding
	dst := make([]byte, base64Enc.EncodedLen(len(bufBS)))
	base64Enc.Encode(dst, bufBS)
	return dst, nil
}

// UnGzipBase64String will base64 decode the string and ungzip it
func UnGzipBase64String(str string) ([]byte, error) {
	bs, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, errors.Wrapf(err, "base64 decode failed, str: %s", str)
	}

	reader := bytes.NewReader(bs)
	gReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "gzip read failed")
	}
	// defer gReader.Close()

	newBS, err := ioutil.ReadAll(gReader)
	if err != nil {
		return nil, errors.Wrapf(err, "read gzip reader failed")
	}
	return newBS, nil
}
