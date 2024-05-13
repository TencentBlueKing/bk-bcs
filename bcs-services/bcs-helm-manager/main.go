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

// Package main xxx
package main

import (
	"flag"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commonConf "github.com/Tencent/bk-bcs/bcs-common/common/conf"
	microYaml "github.com/go-micro/plugins/v4/config/encoder/yaml"
	microCfg "go-micro.dev/v4/config"
	"go-micro.dev/v4/config/reader"
	microJson "go-micro.dev/v4/config/reader/json"
	"go-micro.dev/v4/config/source/env"
	microFile "go-micro.dev/v4/config/source/file"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/app"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/options"
)

var (
	conf        string
	credentials string
)

func parseFlags() {
	// config file path
	flag.StringVar(&conf, "conf", "", "config file path")
	flag.StringVar(&credentials, "credentials", "", "credential config file path")
	flag.Parse()
}
func main() {
	parseFlags()

	opt := &options.HelmManagerOptions{}
	config, err := microCfg.NewConfig(microCfg.WithReader(microJson.NewReader(
		reader.WithEncoder(microYaml.NewEncoder()),
	)))
	if err != nil {
		blog.Fatalf("create config failed, %s", err.Error())
	}

	envSource := env.NewSource(
		env.WithStrippedPrefix("HELM"),
	)

	// 加载主配置
	if len(conf) > 0 {
		err = config.Load(microFile.NewSource(microFile.WithPath(conf)), envSource)
		if err != nil {
			blog.Fatalf("load config from file failed, err %s", err.Error())
		}
	}

	credConf, _ := microCfg.NewConfig()
	if len(credentials) > 0 {
		credConf, err = makeMicroCredConf(credentials)
		if err != nil {
			blog.Fatalf("load credentials from file failed, err %s", err.Error())
		}
		// 设置白名单配置
		config.Set(credConf.Get("credentials"), "credentials")

	}

	if err = config.Scan(opt); err != nil {
		blog.Fatalf("scan config failed, %s", err.Error())
	}

	// addons 动态配置
	addonsConf, _ := microCfg.NewConfig()
	if len(opt.Release.AddonsConfigFile) > 0 {
		addonsConf, err = makeMicroCredConf(opt.Release.AddonsConfigFile)
		if err != nil {
			blog.Fatalf("load addons from file failed, err %s", err.Error())
		}
	}

	// 初始化 I18N 相关配置
	if err = i18n.InitMsgMap(); err != nil {
		blog.Fatalf("init i18n message map failed %s", err.Error())
	}

	blog.InitLogs(commonConf.LogConfig{
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
	helmManager := app.NewHelmManager(opt, credConf, addonsConf)
	if err := helmManager.Init(); err != nil {
		blog.Fatalf("init helm manager failed, %s", err.Error())
	}
	helmManager.RegistryStop()

	if err := helmManager.Run(); err != nil {
		blog.Fatalf("run helm manager failed, %s", err.Error())
	}
}

func makeMicroCredConf(filePath string) (microCfg.Config, error) {
	config, err := microCfg.NewConfig(
		microCfg.WithReader(microJson.NewReader(reader.WithEncoder(microYaml.NewEncoder()))),
	)
	if err != nil {
		return nil, err
	}

	if err := config.Load(microFile.NewSource(microFile.WithPath(filePath))); err != nil {
		return nil, err
	}
	return config, nil
}
