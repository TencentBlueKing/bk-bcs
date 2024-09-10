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
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// ValidateUnixFilePath validate unix os file path.
func ValidateUnixFilePath(kit *kit.Kit, path string) error {
	if len(path) < 1 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid path %s, length should >= 1", path))
	}

	if len(path) > 1024 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid path %s, length should <= 1024", path))
	}

	// 1. 检查是否以 '/' 开头
	if !strings.HasPrefix(path, "/") {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid path %s, the path must start with '/'", path))
	}

	// 2. 检查是否包含连续的 '/'
	if strings.Contains(path, "//") {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid path %s, the path"+
			"cannot contain consecutive '/'", path))
	}

	// 3. 检查是否以 '/' 结尾（除非是根路径）
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid path %s, the path cannot end with '/'", path))
	}

	return nil
}

// qualifiedWinFilePathRegexp win file path validate regexp.
var qualifiedWinFilePathRegexp = regexp.MustCompile("^[a-zA-Z]:(\\\\[\\w\u4e00-\u9fa5\\s]+)+")

// ValidateWinFilePath validate win file path.
func ValidateWinFilePath(kit *kit.Kit, path string) error {
	if len(path) < 1 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid path, length should >= 1"))
	}

	if len(path) > 256 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid path, length should <= 256"))
	}

	if !qualifiedWinFilePathRegexp.MatchString(path) {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid path,"+
			"path does not conform to the win file path format specification"))
	}

	return nil
}

// qualifiedReloadFilePathRegexp reload file path validate regexp.
var qualifiedReloadFilePathRegexp = regexp.MustCompile(fmt.Sprintf("^.*[\\\\/]*%s[\\\\/]*.*$",
	constant.SideWorkspaceDir))

// ValidateReloadFilePath validate reload file path.
func ValidateReloadFilePath(kit *kit.Kit, path string) error {
	if len(path) == 0 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "reload file path is required"))
	}

	if len(path) > 128 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid reload file path, should <= 128"))
	}

	if yes := filepath.IsAbs(path); !yes {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "reload file path is not the absolute path"))
	}

	if qualifiedReloadFilePathRegexp.MatchString(path) {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "%s sub path is system reserved path, do not allow to use",
			constant.SideWorkspaceDir))
	}

	return nil
}
