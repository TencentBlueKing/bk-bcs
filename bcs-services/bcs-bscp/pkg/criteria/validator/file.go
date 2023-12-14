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

// Package validator is for validating data
package validator

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
)

// qualifiedUnixFilePathRegexp unix file path validate regexp.
var qualifiedUnixFilePathRegexp = regexp.MustCompile(`^\/([A-Za-z0-9_]+[A-Za-z0-9-_.]*\/?)*$`)

// ValidateUnixFilePath validate unix os file path.
func ValidateUnixFilePath(path string) error {
	if len(path) < 1 {
		return errors.New("invalid path, length should >= 1")
	}

	if len(path) > 256 {
		return errors.New("invalid path, length should <= 256")
	}

	if !qualifiedUnixFilePathRegexp.MatchString(path) {
		return fmt.Errorf("invalid path, path does not conform to the unix file path format specification")
	}

	return nil
}

// qualifiedWinFilePathRegexp win file path validate regexp.
var qualifiedWinFilePathRegexp = regexp.MustCompile("^[a-zA-Z]:(\\\\[\\w\u4e00-\u9fa5\\s]+)+")

// ValidateWinFilePath validate win file path.
func ValidateWinFilePath(path string) error {
	if len(path) < 1 {
		return errors.New("invalid path, length should >= 1")
	}

	if len(path) > 256 {
		return errors.New("invalid path, length should <= 256")
	}

	if !qualifiedWinFilePathRegexp.MatchString(path) {
		return fmt.Errorf("invalid path, path does not conform to the win file path format specification")
	}

	return nil
}

// qualifiedReloadFilePathRegexp reload file path validate regexp.
var qualifiedReloadFilePathRegexp = regexp.MustCompile(fmt.Sprintf("^.*[\\\\/]*%s[\\\\/]*.*$",
	constant.SideWorkspaceDir))

// ValidateReloadFilePath validate reload file path.
func ValidateReloadFilePath(path string) error {
	if len(path) == 0 {
		return errors.New("reload file path is required")
	}

	if len(path) > 128 {
		return errors.New("invalid reload file path, should <= 128")
	}

	if yes := filepath.IsAbs(path); !yes {
		return errors.New("reload file path is not the absolute path")
	}

	if qualifiedReloadFilePathRegexp.MatchString(path) {
		return fmt.Errorf("%s sub path is system reserved path, do not allow to use", constant.SideWorkspaceDir)
	}

	return nil
}
