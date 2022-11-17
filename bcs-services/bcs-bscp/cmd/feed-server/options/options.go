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
	"errors"
	"fmt"
	"regexp"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/runtime/flags"
	"bscp.io/pkg/tools"

	"github.com/spf13/pflag"
)

// Option defines the app's runtime flag options.
type Option struct {
	Sys *cc.SysOption
	// PublicKey used to api gateway jwt token.
	PublicKey string
	// Name current feed server instance name，is the only one of all feed server.
	Name string
	// DisableJWT whether to enable blueking api-gateway jwt parser.if disable-tgw = false, api-service
	// parse request will api-gateway parser, requests from other parties will not be parsed.if
	// disable-tgw = true, api-service parse requests for direct access that have not been processed by
	// the gateway. Parse rule details：pkg/runtime/parser/parser.go
	DisableJWT bool
}

// InitOptions init config server's options from command flags.
func InitOptions() (*Option, error) {
	fs := pflag.CommandLine
	sysOpt := flags.SysFlags(fs)
	opt := &Option{Sys: sysOpt}

	fs.StringVarP(&opt.PublicKey, "public-key", "", "", "the api gateway public key path")
	fs.BoolVarP(&opt.DisableJWT, "disable-jwt", "", false, "to disable jwt authorize for "+
		"all the incoming request. Note: disable jwt authorize may cause security problems.")
	fs.StringVarP(&opt.Name, "name", "n", "", "feed server instance name, that is the only one of all feed server. "+
		"And only allows to include english、numbers, and must start and end with an english")

	// parses the command-line flags from os.Args[1:]. must be called after all flags are defined
	// and before flags are accessed by the program.
	pflag.Parse()

	// check if the command-line flag is show current version info cmd.
	sysOpt.CheckV()

	if len(opt.Name) == 0 {
		opt.Name = minor()
	}

	if err := ValidateSvcInstName(opt.Name); err != nil {
		return nil, err
	}

	return opt, nil
}

// minor used to generate the default feed server instance name, which generates an 8-bit string
func minor() string {
	return tools.RandString(8)
}

// ValidateSvcInstName validate service instance's name.
func ValidateSvcInstName(name string) error {
	if len(name) < 1 {
		return errors.New("invalid name, length should >= 1")
	}

	if len(name) > 32 {
		return errors.New("invalid name, length should <= 32")
	}

	if !regexp.MustCompile("^[A-Za-z]+[A-Za-z0-9]+$").MatchString(name) {
		return fmt.Errorf("invalid name: %s, only allows to include english、numbers, and must start and "+
			"end with an english", name)
	}

	return nil
}
