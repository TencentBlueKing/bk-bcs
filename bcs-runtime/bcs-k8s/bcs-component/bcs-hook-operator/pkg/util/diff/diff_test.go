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

package diff

import (
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateTwoWayMergePatch(t *testing.T) {
	tests := []struct {
		name          string
		origin        interface{}
		new           interface{}
		dataStruct    interface{}
		expectedBytes []byte
		expectedDiff  bool
		expectedError bool
	}{
		{
			name:          "empty",
			origin:        hookv1alpha1.HookRun{},
			new:           hookv1alpha1.HookRun{},
			dataStruct:    hookv1alpha1.HookRun{},
			expectedBytes: []byte{'{', '}'},
		},
		{
			name:          "origin error",
			origin:        make(chan int),
			new:           hookv1alpha1.HookRun{},
			dataStruct:    hookv1alpha1.HookRun{},
			expectedError: true,
		},
		{
			name:          "new error",
			origin:        hookv1alpha1.HookRun{},
			new:           make(chan int),
			dataStruct:    hookv1alpha1.HookRun{},
			expectedError: true,
		},
		{
			name:          "data struct error",
			origin:        hookv1alpha1.HookRun{},
			new:           hookv1alpha1.HookRun{},
			dataStruct:    "1",
			expectedError: true,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			b, empty, err := CreateTwoWayMergePatch(s.origin, s.new, s.dataStruct)
			gotError := err != nil
			assert.Equal(t, gotError, s.expectedError)
			assert.Equal(t, s.expectedBytes, b)
			assert.Equal(t, s.expectedDiff, empty)
		})
	}
}
