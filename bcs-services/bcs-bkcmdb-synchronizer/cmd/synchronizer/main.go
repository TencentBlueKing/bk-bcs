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
	microCfg "go-micro.dev/v4/config"
	microFile "go-micro.dev/v4/config/source/file"
	microFlg "go-micro.dev/v4/config/source/flag"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/option"
)

var (
	// BkcmdbSynchronizerOption xxx
	BkcmdbSynchronizerOption = &option.BkcmdbSynchronizerOption{}
)

func init() {
	flag.String("conf", "", "config file")
	flag.Parse()

	config, err := microCfg.NewConfig()
	if err != nil {
		blog.Fatalf("create config failed, err: %s", err.Error())
	}

	if err = config.Load(
		microFlg.NewSource(
			microFlg.IncludeUnset(true),
		),
	); err != nil {
		blog.Fatalf("load config failed, err: %s", err.Error())
	}

	if len(config.Get("conf").String("")) > 0 {
		err = config.Load(microFile.NewSource(microFile.WithPath(config.Get("conf").String(""))))
		if err != nil {
			blog.Fatalf("load config failed, err: %s", err.Error())
		}
	}

	if err = config.Scan(BkcmdbSynchronizerOption); err != nil {
		blog.Fatalf("load config failed, err: %s", err.Error())
	}

	if err := common.DecryptCMOption(BkcmdbSynchronizerOption); err != nil {
		blog.Fatalf("load config failed, err: %s", err.Error())
	}
}

func main() {
	Synchronizerd()
}
