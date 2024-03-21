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

package command

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/options"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/cachemanager"
)

var (
	ConfigFile string
)

func InitCacheManager(ctx context.Context, kubeConfig string) (cachemanager.CacheInterface, error) {
	cacheManager := cachemanager.NewCacheManager()
	if kubeConfig != "" {
		if err := cacheManager.InitWithKubeConfig(kubeConfig); err != nil {
			return nil, errors.Wrapf(err, "cache manager init failed")
		}
	} else {
		if err := cacheManager.Init(); err != nil {
			return nil, errors.Wrapf(err, "cache manager init failed")
		}
	}
	go func() {
		if err := cacheManager.Start(ctx); err != nil {
			panic(err)
		}
	}()
	return cacheManager, nil
}

func InitConfig() error {
	cfgHandler := options.GlobalConfigHandler()
	var bs []byte
	bs, err := os.ReadFile(ConfigFile)
	if err != nil {
		return errors.Wrapf(err, "config '%s' read failed", ConfigFile)
	}
	if err = json.Unmarshal(bs, cfgHandler.GetOptions()); err != nil {
		return errors.Wrapf(err, "unmarshal config '%s' failed", ConfigFile)
	}
	return nil
}

// Exit the command
func Exit(template string, args ...string) {
	blog.Errorf(template, args)
	blog.CloseLogs()
	os.Exit(1)
}
