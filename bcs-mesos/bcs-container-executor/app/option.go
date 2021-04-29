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

package app

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/logs"

	"github.com/spf13/pflag"
)

const (
	//DefaultCNIDirectory default cni directory
	DefaultCNIDirectory = "/data/bcs/bcs-cni"
)

// CommandFlags hold all command line flags from mesos-slave
type CommandFlags struct {
	User                string // user for authentication
	Passwd              string // password for authentication
	DockerSocket        string // docker socket path
	MappedDirectory     string // The sandbox directory path that is mapped in the docker container.
	NetworkMode         string // mode for cni/cnm
	CNIPluginDir        string // cni plugin directory, $CNIPluginDir/bin for binary, $CNIPluginDir/conf for configuration
	NetworkImage        string // cni network images
	ExtendedResourceDir string // dir for extended resources
}

// NewCommandFlags return new DockerFalgs with default value
func NewCommandFlags() *CommandFlags {
	return &CommandFlags{
		User:                "",
		Passwd:              "",
		DockerSocket:        "unix:///var/run/docker.sock",
		MappedDirectory:     "/etc/mnt/bcs",
		NetworkMode:         "",
		CNIPluginDir:        DefaultCNIDirectory,
		ExtendedResourceDir: "/data/bcs/extended-resources",
	}
}

// ParseCmdFlags from command line input
func ParseCmdFlags() *CommandFlags {
	cmdFlag := NewCommandFlags()
	flag := pflag.CommandLine
	flag.StringVar(&cmdFlag.User, "user", cmdFlag.User, "user for executor")
	flag.StringVar(&cmdFlag.Passwd, "uuid", cmdFlag.Passwd, "uuid for executor")
	flag.StringVar(&cmdFlag.DockerSocket, "docker-socket", cmdFlag.DockerSocket,
		"container name for running docker container")
	flag.StringVar(&cmdFlag.MappedDirectory, "mapped-directory", cmdFlag.MappedDirectory,
		"The sandbox directory path that is mapped in the docker container.")
	flag.StringVar(&cmdFlag.CNIPluginDir, "cni-plugin", cmdFlag.CNIPluginDir,
		"cni interface plugin directory, $cni_plugin/bin for binary, $cni_plugin/conf for configuration")
	flag.StringVar(&cmdFlag.NetworkMode, "network-mode", cmdFlag.NetworkMode,
		"container network mode: cni or cnm. default empty")
	flag.StringVar(&cmdFlag.NetworkImage, "network-image", cmdFlag.NetworkImage, "container network image")
	flag.StringVar(&cmdFlag.ExtendedResourceDir, "extended-resource-directory", cmdFlag.ExtendedResourceDir,
		"the directory where all executor records extended resource allocation")
	util.InitFlags()
	// parse base64 uuid to password, skip if uuid empty
	if len(cmdFlag.Passwd) != 0 {
		defer func() {
			// recover when Descrypt panic
			if err := recover(); err != nil {
				logs.Errorf("%+v\n", err)
			}
		}()
		uuidBytes := []byte(cmdFlag.Passwd)
		// Warning: DesDecryptFromBase failed will panic
		passwdBytes, _ := encrypt.DesDecryptFromBase(uuidBytes)
		cmdFlag.Passwd = string(passwdBytes)
	}
	return cmdFlag
}
