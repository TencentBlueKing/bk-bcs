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

package options

import (
	yaml "github.com/asim/go-micro/plugins/config/encoder/yaml/v4"
	"github.com/pkg/errors"
	microConf "go-micro.dev/v4/config"
	"go-micro.dev/v4/config/reader"
	"go-micro.dev/v4/config/reader/json"
	"go-micro.dev/v4/config/source"
	"go-micro.dev/v4/config/source/file"
	"golang.org/x/sync/errgroup"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
)

// WebConsoleManagerOption is option in flags
type WebConsoleManagerOption struct {
	BCSConfig string
}

// MultiCredConf 凭证
type MultiCredConf struct {
	watcherMap map[string]microConf.Watcher
	confMap    map[string]microConf.Config
}

// makeMicroConf 配置文件
func makeMicroConf(filePath string) (microConf.Config, error) {
	conf, err := microConf.NewConfig(
		microConf.WithReader(json.NewReader(reader.WithEncoder(yaml.NewEncoder()))),
	)
	if err != nil {
		return nil, err
	}

	if err := conf.Load(file.NewSource(file.WithPath(filePath))); err != nil {
		return nil, err
	}
	return conf, nil
}

// MultiCredConf 新增
func NewMultiCredConf(filePaths []string) (*MultiCredConf, error) {
	multiCredConf := &MultiCredConf{
		confMap:    make(map[string]microConf.Config),
		watcherMap: make(map[string]microConf.Watcher),
	}

	for _, filePath := range filePaths {
		_, ok := multiCredConf.confMap[filePath]
		if ok {
			return nil, errors.New("credential config is duplicated")
		}

		conf, err := makeMicroConf(filePath)
		if err != nil {
			return nil, err
		}

		if err := config.G.ReadCred(filePath, conf.Get("credentials").Bytes()); err != nil {
			return nil, err
		}
		logger.Infof("load credential conf from %s, len=%d", filePath, len(config.G.Credentials[filePath]))

		multiCredConf.confMap[filePath] = conf

	}

	return multiCredConf, nil
}

// Watch 监听多个文件变化
func (m *MultiCredConf) Watch() error {
	var eg errgroup.Group

	for name, conf := range m.confMap {
		w, err := m.watch(name, conf, &eg)
		if err != nil {
			return err
		}
		m.watcherMap[name] = w
	}

	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

// watch 监听单个文件变化
func (m *MultiCredConf) watch(name string, conf microConf.Config, eg *errgroup.Group) (microConf.Watcher, error) {
	w, err := conf.Watch("credentials")
	if err != nil {
		return nil, err
	}

	eg.Go(func() error {
		for {
			value, err := w.Next()
			if err != nil {
				if err.Error() == source.ErrWatcherStopped.Error() {
					return nil
				}
				return err
			}
			// watch 会传入 null 空值
			if string(value.Bytes()) == "null" {
				continue
			}
			if err := config.G.ReadCred(name, value.Bytes()); err != nil {
				logger.Errorf("reload credential error, %s", err)
			}
			logger.Infof("reload credential conf from %s, len=%d", name, len(config.G.Credentials[name]))
		}
	})

	return w, nil
}

// Stop 停止所有监听
func (m *MultiCredConf) Stop() {
	for name, w := range m.watcherMap {
		logger.Infof("receive interput, stop watch %s", name)
		w.Stop()
	}
}
