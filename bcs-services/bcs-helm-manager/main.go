/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/app"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/options"

	microCfg "github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/encoder/yaml"
	"github.com/micro/go-micro/v2/config/reader"
	"github.com/micro/go-micro/v2/config/reader/json"
	"github.com/micro/go-micro/v2/config/source/env"
	microFile "github.com/micro/go-micro/v2/config/source/file"
	microFlg "github.com/micro/go-micro/v2/config/source/flag"
)

func parseFlags() {
	// config file path
	flag.String("conf", "", "config file path")
	flag.Parse()
}

func main() {
	parseFlags()

	opt := &options.HelmManagerOptions{}
	config, err := microCfg.NewConfig(microCfg.WithReader(json.NewReader(
		reader.WithEncoder(yaml.NewEncoder()),
	)))
	if err != nil {
		blog.Fatalf("create config failed, %s", err.Error())
	}

	envSource := env.NewSource(
		env.WithStrippedPrefix("HELM"),
	)

	if err = config.Load(
		microFlg.NewSource(
			microFlg.IncludeUnset(true),
		),
	); err != nil {
		blog.Fatalf("load config from flag failed, %s", err.Error())
	}

	if len(config.Get("conf").String("")) > 0 {
		err = config.Load(microFile.NewSource(microFile.WithPath(config.Get("conf").String(""))), envSource)
		if err != nil {
			blog.Fatalf("load config from file failed, err %s", err.Error())
		}
	}

	if err = config.Scan(opt); err != nil {
		blog.Fatalf("scan config failed, %s", err.Error())
	}

	blog.InitLogs(conf.LogConfig{
		LogDir:          opt.BcsLog.LogDir,
		LogMaxSize:      opt.BcsLog.LogMaxSize,
		LogMaxNum:       opt.BcsLog.LogMaxNum,
		ToStdErr:        opt.BcsLog.ToStdErr,
		AlsoToStdErr:    opt.BcsLog.AlsoToStdErr,
		Verbosity:       opt.BcsLog.Verbosity,
		StdErrThreshold: opt.BcsLog.StdErrThreshold,
		VModule:         opt.BcsLog.VModule,
		TraceLocation:   opt.BcsLog.TraceLocation,
	})

	blog.Info(string(config.Bytes()))
	options.GlobalOptions = opt
	helmManager := app.NewHelmManager(opt)
	if err := helmManager.Init(); err != nil {
		blog.Fatalf("init helm manager failed, %s", err.Error())
	}
	helmManager.RegistryStop()

	if err := helmManager.Run(); err != nil {
		blog.Fatalf("run helm manager failed, %s", err.Error())
	}
}
