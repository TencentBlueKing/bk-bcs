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

// validUnixFileSubPathRegexp sub path support character:
// chinese, english, number, '-', '_', '#', '%', ',', '@', '^', '+', '=', '[', ']', '{', '}, '.'
var validUnixFileSubPathRegexp = regexp.MustCompile("^[\u4e00-\u9fa5A-Za-z0-9-_#%,.@^+=\\[\\]{}]+$")

// ValidateUnixFilePath validate unix os file path.
func ValidateUnixFilePath(kit *kit.Kit, path string) error {
	if len(path) < 1 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid path, length should >= 1"))
	}

	if len(path) > 1024 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid path, length should <= 1024"))
	}

	// 1. should start with '/'
	if !strings.HasPrefix(path, "/") {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid path, should start with '/'"))
	}

	// Split the path into parts
	parts := strings.Split(path, "/")[1:] // Ignore the first empty part due to the leading '/'

	if strings.HasSuffix(path, "/") {
		parts = parts[:len(parts)-1] // Ignore the last empty part due to the trailing '/'
	}

	// Iterate over each part to validate
	for _, part := range parts {

		// 2. the verification path cannot all be '{'. '}'
		if dotsRegexp.MatchString(part) {
			return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid path %s, path cannot all be '.' ", part))
		}

		// 3. each sub path support character:
		// chinese, english, number, '-', '_', '#', '%', ',', '@', '^', '+', '=', '[', ']', '{', '}'
		if !validUnixFileSubPathRegexp.MatchString(part) {
			return errf.Errorf(errf.InvalidArgument, i18n.T(kit, fmt.Sprintf(`invalid path, each sub path should only
			 contain chinese, english, number, '-', '_', '#', '%%', ',', '@', '^', '+', '=', '[', ']', '{', '}', '{'. '}`)))
		}

		// 4. each sub path should be separated by '/'
		// (handled by strings.Split above)
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
