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

package check

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

const (
	prefixFlag  = "BCS_CHECK"
	successFlag = "BCS_MODULE_SUCCESS"
	failureFlag = "BCS_MODULE_FAILURE"
)

var (
	program = filepath.Base(os.Args[0])
)

type logFunc func(string, ...interface{})

func do(logger logFunc, flag, message string) {
	logger("%s | %s | %s", prefixFlag, flag, message)
}

// Succeed print success-message to log file
func Succeed() {
	do(blog.Infof, successFlag, fmt.Sprintf("%s is working.", program))
}

// Fail print failure-message to log file
func Fail(message string) {
	do(blog.Errorf, failureFlag, fmt.Sprintf("%s is not working. %s", program, message))
}
