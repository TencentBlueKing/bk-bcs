/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

// Package options NOTES
package options

import (
	"bscp.io/pkg/cc"
	"bscp.io/pkg/runtime/flags"

	"github.com/spf13/pflag"
)

// Option defines the app's runtime flag options.
type Option struct {
	Sys *cc.SysOption
}

// InitOptions init config server's options from command flags.
func InitOptions() *Option {
	fs := pflag.CommandLine
	sysOpt := flags.SysFlags(fs)

	fs.IntVar(&sysOpt.GRPCPort, "grpc-port", 9515, "grpc service port")
	fs.IntVar(&sysOpt.Port, "port", 9615, "http/metrics port")

	// parses the command-line flags from os.Args[1:]. must be called after all flags are defined
	// and before flags are accessed by the program.
	pflag.Parse()

	// check if the command-line flag is show current version info cmd.
	sysOpt.CheckV()

	return &Option{Sys: sysOpt}
}
