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

package tools

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// SHA256 returns a sha256 string of the data string.
func SHA256(data string) string {
	hash := sha256.New()
	if _, err := io.WriteString(hash, data); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// ByteSHA256 returns a sha256 string of the data byte.
func ByteSHA256(data []byte) string {
	hash := sha256.New()
	hash.Write(data) // nolint
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// FileSHA256 returns sha256 string of the file.
func FileSHA256(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// MD5 returns a md5 string of the data string.
func MD5(data string) string {
	hash := md5.New() // nolint
	if _, err := io.WriteString(hash, data); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// ByteMD5 returns a md5 string of the data byte.
func ByteMD5(data []byte) string {
	hash := md5.New() // nolint
	hash.Write(data)  // nolint
	return fmt.Sprintf("%x", hash.Sum(nil))
}
