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
	"errors"
	"fmt"
	"regexp"
	"strings"

	"bscp.io/pkg/criteria/errf"
)

// reservedResNamePrefix internal reserved string prefix, case-insensitive.
var reservedResNamePrefix = []string{"_bk"}

// validResNamePrefix verify whether the resource naming takes up the reserved resource name prefix of bscp.
func validResNamePrefix(name string) error {
	lowerName := strings.ToLower(name)
	for _, prefix := range reservedResNamePrefix {
		if strings.HasPrefix(lowerName, prefix) {
			return fmt.Errorf("resource name '%s' is prefixed with '%s' is reserved name, "+
				"which is not allows to use", lowerName, prefix)
		}
	}

	return nil
}

// qualifiedAppNameRegexp bscp app's name regexp.
var qualifiedAppNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9][\w\-]*[a-zA-Z0-9]$`)

// ValidateAppName validate bscp app name's length and format.
func ValidateAppName(name string) error {
	if len(name) < 1 {
		return errors.New("invalid name, length should >= 1")
	}

	if len(name) > 128 {
		return errors.New("invalid name, length should <= 128")
	}

	if err := validResNamePrefix(name); err != nil {
		return err
	}

	if !qualifiedAppNameRegexp.MatchString(name) {
		return fmt.Errorf("invalid name: %s, only allows to include english、numbers、underscore (_)"+
			"、hyphen (-), and must start and end with an english、numbers", name)
	}

	return nil
}

// qualifiedVariableNameRegexp bscp variable's name regexp.
var qualifiedVariableNameRegexp = regexp.MustCompile(`^(?i)(bk_bscp_)\w*$`)

// ValidateVariableName validate bscp variable's length and format.
func ValidateVariableName(name string) error {
	if len(name) < 9 {
		return errf.Errorf(errf.InvalidArgument, "invalid name, "+
			"length should >= 9 and must start with prefix bk_bscp_ (ignore case)")
	}

	if len(name) > 128 {
		return errf.Errorf(errf.InvalidArgument, "invalid name, length should <= 128")
	}

	if !qualifiedVariableNameRegexp.MatchString(name) {
		return errf.Errorf(errf.InvalidArgument,
			"invalid name: %s, only allows to include english、numbers、underscore (_)"+
				", and must start with prefix bk_bscp_ (ignore case)", name)
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
func ValidateName(name string) error {
	if len(name) < 1 {
		return errors.New("invalid name, length should >= 1")
	}

	if len(name) > 128 {
		return errors.New("invalid name, length should <= 128")
	}

	if err := validResNamePrefix(name); err != nil {
		return err
	}

	if !qualifiedNameRegexp.MatchString(name) {
		return fmt.Errorf("invalid name: %s, only allows to include chinese、english、numbers、underscore (_)"+
			"、hyphen (-), and must start and end with an chinese、english、numbers", name)
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
func ValidateReleaseName(name string) error {
	if len(name) < 1 {
		return errors.New("invalid name, length should >= 1")
	}

	if len(name) > 128 {
		return errors.New("invalid name, length should <= 128")
	}

	if err := validResNamePrefix(name); err != nil {
		return err
	}

	if !qualifiedRNRegexp.MatchString(name) {
		return fmt.Errorf("invalid name: %s, only allows to include chinese、english、numbers、underscore (_)"+
			"、hyphen (-)、point (.), and must start and end with an chinese、english、numbers", name)
	}

	return nil
}

const (
	qualifiedCfgItemNameFmt string = "^[A-Za-z0-9-_.]+$"
)

// qualifiedCfgItemNameRegexp config item name regexp.
var qualifiedCfgItemNameRegexp = regexp.MustCompile(qualifiedCfgItemNameFmt)

// ValidateCfgItemName validate config item's name.
func ValidateCfgItemName(name string) error {
	if len(name) < 1 {
		return errors.New("invalid name, length should >= 1")
	}

	if len(name) > 64 {
		return errors.New("invalid name, length should <= 64")
	}

	if err := validResNamePrefix(name); err != nil {
		return err
	}

	if name == "." || name == ".." {
		return fmt.Errorf("invalid name: %s, not allows to be '.' or '..'", name)
	}

	if !qualifiedCfgItemNameRegexp.MatchString(name) {
		return fmt.Errorf("invalid name: %s, only allows to include english、numbers、underscore (_)"+
			"、hyphen (-)、point (.)", name)
	}

	return nil
}

// ValidateNamespace validate namespace is valid or not.
func ValidateNamespace(namespace string) error {
	if len(namespace) < 1 {
		return errors.New("invalid namespace, length should >= 1")
	}

	if len(namespace) > 128 {
		return errors.New("invalid namespace, length should <= 128")
	}

	if err := validResNamePrefix(namespace); err != nil {
		return err
	}

	if !qualifiedNameRegexp.MatchString(namespace) {
		return fmt.Errorf("invalid namespace: %s, only allows to include chinese、english、numbers、"+
			"underscore (_)、hyphen (-), and must start and end with an chinese、english、numbers", namespace)
	}
	return nil
}

// ValidateUserName validate username.
func ValidateUserName(username string) error {
	if len(username) < 1 {
		return errors.New("invalid username, length should >= 1")
	}

	if len(username) > 32 {
		return errors.New("invalid username, length should <= 32")
	}

	return nil
}
