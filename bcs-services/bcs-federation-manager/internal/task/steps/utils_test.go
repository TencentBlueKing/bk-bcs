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

// Package steps include all steps for federation manager
package steps

import (
	"testing"
)

func TestGenerateRandomToken(t *testing.T) {
	token, err := GenerateRandomStr(16)
	if err != nil {
		t.Errorf("Failed to generate token: %v", err)
	}
	if len(token) != 16 {
		t.Errorf("Expected token length of 16, but got %d", len(token))
	}
	if token == "" {
		t.Error("Generated token should not be empty")
	}

	token2, err2 := GenerateRandomStr(6)
	if err2 != nil {
		t.Errorf("Failed to generate token2: %v", err2)
	}
	if len(token2) != 6 {
		t.Errorf("Expected token2 length of 6, but got %d", len(token2))
	}
	if token2 == "" {
		t.Error("Generated token2 should not be empty")
	}

	t.Logf("Generated token: %s", token)
	t.Logf("Generated token2: %s", token2)
}
