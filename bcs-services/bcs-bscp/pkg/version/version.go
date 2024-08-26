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

// Package version NOTES
package version

import (
	"fmt"
	"regexp"
	"runtime"

	semver "github.com/hashicorp/go-version"
)

func init() {
	// validate if the VERSION is valid
	_, err := parseVersion(VERSION)
	if err != nil {
		msg := fmt.Sprintf("invalid build version, err: %v", err)
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

	// GITTAG xxx
	GITTAG = "1.0.0"

	// DEBUG if enable debug.
	DEBUG = "false"

	// CLIENTTYPE client type (agent、sidecar、sdk、command).
	CLIENTTYPE = "sdk"

	// GoVersion Go 版本号
	GoVersion = runtime.Version()

	// Row print version info by row.
	Row Format = "row"
	// JSON print version info by json.
	JSON Format = "json"
)

// Format defines the format to print version.
type Format string

// Debug show the version if enable debug.
func Debug() bool {
	return DEBUG == "true"
}

// ShowVersion shows the version info.
func ShowVersion(prefix string, format Format) {
	fmt.Println(FormatVersion(prefix, format))
}

// FormatVersion returns service's version.
func FormatVersion(prefix string, format Format) string {
	if prefix != "" {
		prefix += " "
	}
	rawFormat := fmt.Sprintf("%sVersion   :%s\nBuildTime :%s\nGitHash   :%s\nGITTAG    :%s\nGoVersion :%s\nClientType:%s",
		prefix, VERSION, BUILDTIME, GITHASH, GITTAG, GoVersion, CLIENTTYPE)
	jsonFormat := fmt.Sprintf(`%s{"Version": "%s", "BuildTime": "%s", "GitHash": "%s", "GITHASH": "%s", "GoVersion": "%s",
	 "ClientType": "%s"}`,
		prefix, VERSION, BUILDTIME, GITHASH, GITTAG, GoVersion, CLIENTTYPE)

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
		Version:    VERSION,
		Hash:       GITHASH,
		GITTAG:     GITTAG,
		Time:       BUILDTIME,
		GoVersion:  GoVersion,
		ClientType: CLIENTTYPE,
	}
}

// SemanticVersion return the current process's version with semantic version format.
func SemanticVersion() [3]uint32 {
	ver, err := parseVersion(VERSION)
	if err != nil {
		panic(fmt.Sprintf("parse version fail, err: %v", err))
	}
	return ver
}

// versionRegex 限制版本号前缀只能为 v1.x.x 格式
var versionRegex = regexp.MustCompile(`^v1\.\d+\.\d+.*$`)

func parseVersion(v string) ([3]uint32, error) {
	// 语义化版本之上限定 bscp 版本规范
	if !versionRegex.MatchString(v) {
		return [3]uint32{}, fmt.Errorf("the version(%s) format should be like v1.0.0", v)
	}

	// 后面的先行版本号和版本编译信息按语义化版本规则处理
	version, err := semver.NewSemver(v)
	if err != nil {
		return [3]uint32{}, err
	}

	segments := version.Segments()
	// 合法的语义化版本必定有 major / minor / patch 版本号
	return [3]uint32{uint32(segments[0]), uint32(segments[1]), uint32(segments[2])}, nil
}

// SysVersion describe a binary version
type SysVersion struct {
	Version    string `json:"version"`
	Hash       string `json:"hash"`
	GITTAG     string `json:"git_tag"`
	Time       string `json:"time"`
	GoVersion  string `json:"go_version"`
	ClientType string `json:"client_type"`
}
