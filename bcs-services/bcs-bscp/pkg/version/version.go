/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

// Package version NOTES
package version

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func init() {
	// validate if the VERSION is valid
	_, err := parseVersion()
	if err != nil {
		msg := fmt.Sprintf("invalid build version, the version(%s) format should be like like v1.0.0 or "+
			"v1.0.0-alpha1, err: %v", VERSION, err)
		fmt.Fprintf(os.Stderr, msg)
		panic(msg)
	}
}

const (
	// LOGO is bk bscp inner logo.
	LOGO = `
===================================================================================
oooooooooo   oooo    oooo         oooooooooo     oooooooo     oooooo    oooooooooo
 888     Y8b  888   8P             888     Y8b d8P      Y8  d8P    Y8b   888    Y88
 888     888  888  d8              888     888 Y88bo       888           888    d88
 888oooo888   88888[      8888888  888oooo888     Y8888o   888           888ooo88P
 888     88b  888 88b              888     88b        Y88b 888           888
 888     88P  888   88b            888     88P oo      d8P  88b    ooo   888
o888bood8P   o888o  o888o         o888bood8P   88888888P     Y8bood8P   o888o
===================================================================================`
)

var (
	// VERSION is version info.
	VERSION = "v1.0.0"

	// BUILDTIME  build time.
	BUILDTIME = "unknown"

	// GITHASH git hash for release.
	GITHASH = "unknown"

	// DEBUG if enable debug.
	DEBUG = "false"

	// GoVersion Go 版本号
	GoVersion = runtime.Version()

	// Row print version info by row.
	Row VersionFormat = "row"
	// JSON print version info by json.
	JSON VersionFormat = "json"
)

// VersionFormat defines the format to print version.
type VersionFormat string

// Debug show the version if enable debug.
func Debug() bool {
	if DEBUG == "true" {
		return true
	}

	return false
}

// ShowVersion shows the version info.
func ShowVersion(prefix string, format VersionFormat) {
	fmt.Println(FormatVersion(prefix, format))
}

// FormatVersion returns service's version.
func FormatVersion(prefix string, format VersionFormat) string {
	if prefix != "" {
		prefix = prefix + " "
	}
	rawFormat := fmt.Sprintf("%sVersion  : %s\nBuildTime: %s\nGitHash  : %s\nGoVersion: %s", prefix, VERSION, BUILDTIME, GITHASH, GoVersion)
	jsonFormat := fmt.Sprintf(`%s{"Version": "%s", "BuildTime": "%s", "GitHash": "%s", "GoVersion": "%s"}`, prefix, VERSION, BUILDTIME, GITHASH, GoVersion)

	switch format {
	case Row:
		return rawFormat
	case JSON:
		return jsonFormat
	default:
		return rawFormat
	}
}

// GetStartInfo returns start info that includes version and logo.
func GetStartInfo() string {
	startInfo := fmt.Sprintf("%s\n\n%s\n", LOGO, FormatVersion("", Row))
	return startInfo
}

// Version NOTES
func Version() *SysVersion {
	return &SysVersion{
		Version:   VERSION,
		Hash:      GITHASH,
		Time:      BUILDTIME,
		GoVersion: GoVersion,
	}
}

// SemanticVersion return the current process's version with semantic version format.
func SemanticVersion() [3]uint32 {
	ver, err := parseVersion()
	if err != nil {
		panic(fmt.Sprintf("parse version fail, error: %v", err))
	}
	return ver
}

var versionRegex = regexp.MustCompile(`^v1\.\d?(\.\d?){1,2}(-[a-z]+\d+)?$`)

func parseVersion() ([3]uint32, error) {
	if !versionRegex.MatchString(VERSION) {
		return [3]uint32{}, errors.New("the version should be suffixed with format like v1.0.0")
	}

	ver := strings.Split(VERSION, "-")[0]
	ver = strings.Trim(ver, " ")
	ver = strings.TrimPrefix(ver, "v")
	ele := strings.Split(ver, ".")
	if len(ele) < 3 {
		return [3]uint32{}, errors.New("version should be like v1.0.0")
	}

	major, err := strconv.Atoi(ele[0])
	if err != nil {
		return [3]uint32{}, fmt.Errorf("invalid major version: %s", ele[0])
	}

	minor, err := strconv.Atoi(ele[1])
	if err != nil {
		return [3]uint32{}, fmt.Errorf("invalid minor version: %s", ele[0])
	}

	patch, err := strconv.Atoi(ele[2])
	if err != nil {
		return [3]uint32{}, fmt.Errorf("invalid patch version: %s", ele[0])
	}

	return [3]uint32{uint32(major), uint32(minor), uint32(patch)}, nil
}

// SysVersion describe a binary version
type SysVersion struct {
	Version   string `json:"version"`
	Hash      string `json:"hash"`
	Time      string `json:"time"`
	GoVersion string `json:"go_version"`
}
