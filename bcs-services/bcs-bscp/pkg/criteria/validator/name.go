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

package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// reservedResNamePrefix internal reserved string prefix, case-insensitive.
var reservedResNamePrefix = []string{"_bk"}

// validResNamePrefix verify whether the resource naming takes up the reserved resource name prefix of bscp.
func validResNamePrefix(kit *kit.Kit, name string) error {
	lowerName := strings.ToLower(name)
	for _, prefix := range reservedResNamePrefix {
		if strings.HasPrefix(lowerName, prefix) {
			return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "resource name '%s' is prefixed with '%s' is reserved name, "+
				"which is not allows to use", lowerName, prefix))
		}
	}

	return nil
}

// qualifiedAppNameRegexp bscp app's name regexp.
var qualifiedAppNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9][\w\-]*[a-zA-Z0-9]$`)

// ValidateAppName validate bscp app name's length and format.
func ValidateAppName(kit *kit.Kit, name string) error {
	if len(name) < 1 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid name, length should >= 1"))
	}

	if len(name) > 128 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid name, length should <= 128"))
	}

	if err := validResNamePrefix(kit, name); err != nil {
		return err
	}

	if !qualifiedAppNameRegexp.MatchString(name) {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid name: %s, only allows to include english、"+
			"numbers、underscore (_)、hyphen (-), and must start and end with an english、numbers", name))
	}

	return nil
}

// qualifiedAppAliasRegexp bscp app's alias regexp.
var qualifiedAppAliasRegexp = regexp.MustCompile(`^[a-zA-Z0-9\p{Han}][\w\p{Han}\-]*[a-zA-Z0-9\p{Han}]$`)

// ValidateAppAlias validate bscp app Alias length and format.
func ValidateAppAlias(kit *kit.Kit, alias string) error {
	if len(alias) < 1 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid name, length should >= 1"))
	}

	if len(alias) > 128 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid name, length should <= 128"))
	}

	if err := validResNamePrefix(kit, alias); err != nil {
		return err
	}

	if !qualifiedAppAliasRegexp.MatchString(alias) {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, fmt.Sprintf(`invalid name: %s, only allows to include Chinese,
		 English, numbers, underscore (_), hyphen (-), and must start and end with Chinese, English, or a number`, alias)))
	}

	return nil
}

// qualifiedVariableNameRegexp bscp variable's name regexp.
var qualifiedVariableNameRegexp = regexp.MustCompile(`^(?i)(bk_bscp_)\w*$`)

// ValidateVariableName validate bscp variable's length and format.
func ValidateVariableName(kit *kit.Kit, name string) error {
	if len(name) < 9 {
		return errf.Errorf(errf.InvalidArgument,
			i18n.T(kit, "invalid name, "+
				"length should >= 9 and must start with prefix bk_bscp_ (ignore case)"))
	}

	if len(name) > 128 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid name, length should <= 128"))
	}

	if !qualifiedVariableNameRegexp.MatchString(name) {
		return errf.Errorf(errf.InvalidArgument,
			i18n.T(kit, "invalid name: %s, only allows to include english、numbers、underscore (_)"+
				", and must start with prefix bk_bscp_ (ignore case)", name))
	}

	return nil
}

const (
	// qualifiedNameFmt bscp resource's name format.
	// '.' And '/' as reserved characters, users are absolutely not allowed to create
	qualifiedNameFmt string = "(" + qnameNameFmt + qnameExtNameFmt + "*)?" + qnameNameFmt
	qnameNameFmt     string = "[\u4E00-\u9FA5A-Za-z0-9]"
	qnameExtNameFmt  string = "[\u4E00-\u9FA5A-Za-z0-9-_]"
)

// qualifiedNameRegexp bscp resource's name regexp.
var qualifiedNameRegexp = regexp.MustCompile("^" + qualifiedNameFmt + "$")

// ValidateName validate bscp resource name's length and format.
func ValidateName(kit *kit.Kit, name string) error {
	if len(name) < 1 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid name, length should >= 1"))
	}

	if len(name) > 128 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid name, length should <= 128"))
	}

	if err := validResNamePrefix(kit, name); err != nil {
		return err
	}

	if !qualifiedNameRegexp.MatchString(name) {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, fmt.Sprintf(`invalid name: %s, only allows to include chinese、
		english、numbers、underscore (_)、hyphen (-), and must start and end with an chinese、english、numbers`, name)))
	}

	return nil
}

const (
	qualifiedReleaseNameFmt string = "(" + qReleaseNameFmt + qExtReleaseNameFmt + "*)?" + qReleaseNameFmt
	qReleaseNameFmt         string = "[\u4E00-\u9FA5A-Za-z0-9]"
	qExtReleaseNameFmt      string = "[\u4E00-\u9FA5A-Za-z0-9-_.]"
)

// qualifiedRNRegexp release name regexp.
var qualifiedRNRegexp = regexp.MustCompile("^" + qualifiedReleaseNameFmt + "$")

// ValidateReleaseName validate release name's length and format.
func ValidateReleaseName(kit *kit.Kit, name string) error {
	if len(name) < 1 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid name, length should >= 1"))
	}

	if len(name) > 128 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid name, length should <= 128"))
	}

	if err := validResNamePrefix(kit, name); err != nil {
		return err
	}

	if !qualifiedRNRegexp.MatchString(name) {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid name: %s, only allows to include Chinese, English,"+
			"numbers, underscore (_),hyphen (-), and must start and end with Chinese, English, or a number", name))
	}

	return nil
}

// qualifiedFileNameRegexp file name regexp.
// support character: chinese, english, number,
// '-', '_', '#', '%', ',', '@', '^', '+', '=', '[', ']', '{', '}, '.'
var qualifiedFileNameRegexp = regexp.MustCompile("^[\u4e00-\u9fa5A-Za-z0-9-_#%,.@^+=\\[\\]\\{\\}]+$")

var dotsRegexp = regexp.MustCompile(`^\.+$`)

// ValidateFileName validate config item's name.
func ValidateFileName(kit *kit.Kit, name string) error {
	if len(name) < 1 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid name, length should >= 1"))
	}

	if len(name) > 64 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid name, length should <= 64"))
	}

	if err := validResNamePrefix(kit, name); err != nil {
		return err
	}

	if dotsRegexp.MatchString(name) {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid name %s, name cannot all be '.'", name))
	}

	if !qualifiedFileNameRegexp.MatchString(name) {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, fmt.Sprintf(`invalid name %s, should only contains chinese,
		 english, number, '-', '_', '#', '%%', ',', '@', '^', '+', '=', '[', ']', '{', '}', '.'`, name)))
	}

	return nil
}

// ValidateNamespace validate namespace is valid or not.
func ValidateNamespace(kit *kit.Kit, namespace string) error {
	if len(namespace) < 1 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid namespace, length should >= 1"))
	}

	if len(namespace) > 128 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid namespace, length should <= 128"))
	}

	if err := validResNamePrefix(kit, namespace); err != nil {
		return err
	}

	if !qualifiedNameRegexp.MatchString(namespace) {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid name: %s, only allows to include Chinese, English,"+
			"numbers, underscore (_),hyphen (-), and must start and end with Chinese, English, or a number", namespace))
	}
	return nil
}

// ValidateUserName validate username.
func ValidateUserName(kit *kit.Kit, username string) error {
	if len(username) < 1 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid username, length should >= 1"))
	}

	if len(username) > 32 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid username, length should <= 32"))
	}

	return nil
}
