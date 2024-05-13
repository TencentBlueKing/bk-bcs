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

package dbm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHideAuth(t *testing.T) {
	testCases := []struct {
		origin string
		expect string
	}{
		{
			origin: `X-Bkapi-Authorization: {"bk_app_code": "test-app-code", "bk_app_secret": "test--app--secret(test%^&)", "bk_username": "test-operator"}`,
			expect: `X-Bkapi-Authorization: {"bk_app_code":"***", "bk_app_secret":"***", "bk_username":"***"}`,
		},
		{
			origin: `X-Bkapi-Authorization: {"bk_app_code":    "test-app-code", "bk_app_secret":  "test--app--secret(test%^&)", "bk_username":  "test-operator"}`,
			expect: `X-Bkapi-Authorization: {"bk_app_code":"***", "bk_app_secret":"***", "bk_username":"***"}`,
		},
	}

	for _, test := range testCases {
		actual := hideAuth(test.origin)
		assert.Equal(t, actual, test.expect, "they should be equal")
	}
}
