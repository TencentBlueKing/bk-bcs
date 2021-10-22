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

package log

import (
	"io"
	"log"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/response"
	"github.com/mholt/caddy"
	"github.com/miekg/dns"
)

func init() {
	caddy.RegisterPlugin("log", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	rules, err := logParse(c)
	if err != nil {
		return plugin.Error("log", err)
	}

	blog.InitLogs(conf.LogConfig{
		ToStdErr:        false,
		AlsoToStdErr:    false,
		Verbosity:       3,
		StdErrThreshold: "2",
		VModule:         "",
		TraceLocation:   "",
		LogDir:          rules[0].OutputFile,
		LogMaxSize:      500,
		LogMaxNum:       10,
	})

	// Open the log files for writing when the server starts
	c.OnStartup(func() error {
		for i := 0; i < len(rules); i++ {
			var writer io.Writer
			if rules[i].OutputFile == "stdout" {
				writer = os.Stdout
			} else if rules[i].OutputFile == "stderr" {
				writer = os.Stderr
			} else {
				writer = &blog.GlogWriter{}
			}
			rules[i].Log = log.New(writer, "", 0)
		}
		return nil

	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return Logger{Next: next, Rules: rules, ErrorFunc: dnsserver.DefaultErrorFunc}
	})

	return nil
}

func logParse(c *caddy.Controller) ([]Rule, error) {
	var rules []Rule

	for c.Next() {
		args := c.RemainingArgs()

		if len(args) == 0 {
			// Nothing specified; use defaults
			rules = append(rules, Rule{
				NameScope:  ".",
				OutputFile: DefaultLogFilename,
				Format:     DefaultLogFormat,
			})
		} else if len(args) == 1 {
			rules = append(rules, Rule{
				NameScope:  dns.Fqdn(args[0]),
				OutputFile: DefaultLogFilename,
				Format:     DefaultLogFormat,
			})
		} else {
			// Name scope, and maybe a format specified
			var format string

			switch args[1] {
			case "{common}":
				format = CommonLogFormat
			case "{combined}":
				format = CombinedLogFormat
			default:
				format = args[1]
			}

			rules = append(rules, Rule{
				NameScope:  dns.Fqdn(args[0]),
				OutputFile: DefaultLogFilename,
				Format:     format,
			})
		}

		// Class refinements in an extra block.
		for c.NextBlock() {
			switch c.Val() {
			// class followed by all, denial, error or success.
			case "class":
				classes := c.RemainingArgs()
				if len(classes) == 0 {
					return nil, c.ArgErr()
				}
				cls, err := response.ClassFromString(classes[0])
				if err != nil {
					return nil, err
				}
				// update class and the last added Rule (bit icky)
				rules[len(rules)-1].Class = cls
			case "log_dir":
				logDirs := c.RemainingArgs()
				if len(logDirs) == 0 {
					return nil, c.ArgErr()
				}
				rules[len(rules)-1].OutputFile = logDirs[0]
			default:
				return nil, c.ArgErr()
			}
		}
	}

	return rules, nil
}
