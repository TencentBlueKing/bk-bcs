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
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

func md5V(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// CalIAMNamespaceID trans namespace to permID, permSystem limit 32 length
func CalIAMNamespaceID(clusterID string, name string) (string, error) {
	clusterStrs := strings.Split(clusterID, "-")
	if len(clusterStrs) != 3 {
		return "", fmt.Errorf("CalIAMNamespaceID err: %v", "length not equal 3")
	}
	clusterIDx := clusterStrs[len(clusterStrs)-1]

	iamNsID := clusterIDx + ":" + md5V(name)[8:16] + name[:2]
	if len(iamNsID) > 32 {
		return "", fmt.Errorf("CalIAMNamespaceID iamNamespaceID more than 32characters")
	}

	return iamNsID, nil
}
