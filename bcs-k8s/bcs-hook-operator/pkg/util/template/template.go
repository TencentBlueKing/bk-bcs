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

package template

import (
	"fmt"
	"io"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	"github.com/valyala/fasttemplate"
)

const (
	openBracket  = "{{"
	closeBracket = "}}"
)

// ResolveArgs substitute the supplied arguments in the given template
func ResolveArgs(template string, args []v1alpha1.Argument) (string, error) {
	t, err := fasttemplate.NewTemplate(template, openBracket, closeBracket)
	if err != nil {
		return "", err
	}
	argsMap := make(map[string]string)
	for i := range args {
		arg := args[i]
		if arg.Value == nil {
			return "", fmt.Errorf("argument \"%s\" was not supplied", arg.Name)
		}
		argsMap[fmt.Sprintf("args.%s", arg.Name)] = *arg.Value
	}
	var unresolvedErr error
	s := t.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		cleanedTag := strings.TrimSpace(tag)
		if value, ok := argsMap[cleanedTag]; ok {
			return w.Write([]byte(value))
		}
		unresolvedErr = fmt.Errorf("failed to resolve {{%s}}", tag)

		return w.Write([]byte(""))
	})
	return s, unresolvedErr
}
