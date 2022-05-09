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

package template

import (
	"errors"
	"fmt"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResolveArgs(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		args           []hookv1alpha1.Argument
		expectedString string
		expectedError  error
	}{
		{
			name:     "template error",
			template: "{{123",
			expectedError: fmt.Errorf("Cannot find end tag=%q in the template=%q starting from %q",
				"}}", "{{123", "123"),
		},
		{
			name: "args value error",
			args: []hookv1alpha1.Argument{
				{Name: "name"},
			},
			expectedError: fmt.Errorf("argument \"%s\" was not supplied", "name"),
		},
		{
			name:     "resolve failed",
			template: "{{}}",
			args: []hookv1alpha1.Argument{
				{Name: "name", Value: func() *string { s := "n"; return &s }()},
			},
			expectedError: errors.New("failed to resolve {{}}"),
		},
		{
			name:     "resolve args",
			template: "{{args.name}}",
			args: []hookv1alpha1.Argument{
				{Name: "name", Value: func() *string { s := "n"; return &s }()},
			},
			expectedString: "n",
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			actualString, actualError := ResolveArgs(s.template, s.args)
			assert.Equal(t, s.expectedError, actualError)
			assert.Equal(t, s.expectedString, actualString)
		})
	}
}
