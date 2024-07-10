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

// Package util xx
package util

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
)

// DesDecrypt 解密
func DesDecrypt(appCode, appSecret, operator []byte) (apc, aps, opt string, err error) {

	if string(appCode) == "" || string(appSecret) == "" || string(operator) == "" {
		return "", "", "", fmt.Errorf("unable to decrypt appCode appSecret appOperator")
	}

	decryptedAppCode, err := encrypt.DesDecryptFromBase(appCode)
	if err != nil {
		blog.Error("unable to decrypt appCode: %s", err.Error())
		return "", "", "", err
	}
	decryptedAppSecret, err := encrypt.DesDecryptFromBase(appSecret)
	if err != nil {
		blog.Error("unable to decrypt appSecret: %s", err.Error())
		return "", "", "", err
	}
	decryptedAppOperator, err := encrypt.DesDecryptFromBase(operator)
	if err != nil {
		blog.Error("unable to decrypt appOperator: %s", err.Error())
		return "", "", "", err
	}

	return string(decryptedAppCode), string(decryptedAppSecret), string(decryptedAppOperator), nil
}
