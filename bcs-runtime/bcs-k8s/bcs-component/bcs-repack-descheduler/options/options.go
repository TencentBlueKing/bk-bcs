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

// Package options defines the config
package options

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// DeSchedulerOption defines the config options for de-scheduler server
type DeSchedulerOption struct {
	conf.FileConfig
	conf.LogConfig

	Debug bool `json:"debug" value:"false" usage:"enable pprof with debug"`

	Address      string `json:"address" value:"0.0.0.0" usage:"the address of the server"`
	HttpPort     int32  `json:"httpPort" value:"8080" usage:"the port of webhook server"`
	MetricPort   int32  `json:"metricPort" value:"8081" usage:"the port of controller's metric server'"`
	ExtenderPort int32  `json:"extenderPort" value:"8088" usage:"the port of http extender"`

	ServerCert string `json:"servercert"`
	ServerKey  string `json:"serverkey"`
	ServerCa   string `json:"serverca"`

	WebhookHost     string `json:"webhookHost"`
	WebhookPort     int32  `json:"webhookPort"`
	WebhookCertDir  string `json:"webhookCertDir"`
	WebhookCertName string `json:"webhookCertName"`
	WebhookKeyName  string `json:"webhookKeyName"`

	MaxEvictionParallel int32 `json:"maxEvictionParallel" value:"100" usage:"parallel num of eviction goroutines"`
	MaxEvictionNodes    int32 `json:"maxEvictionNodes" value:"1" usage:"parallel num of eviction nodes"`

	// 远程模型计算地址
	BKDataUrl       string `json:"bkDataUrl" value:"" usage:"url of bkdata model"`
	BKDataAppCode   string `json:"bkDataAppCode" value:"" usage:"appcode of bkdata"`
	BKDataAppSecret string `json:"bkDataAppSecret" value:"" usage:"appsecret of bkdata"`
	BKDataToken     string `json:"bkDataToken" value:"" usage:"token of bkdata"`
}

func (o DeSchedulerOption) deepCopy() (result DeSchedulerOption, err error) {
	bs, err := json.Marshal(o)
	if err != nil {
		return result, errors.Wrapf(err, "deepcopy marshal failed")
	}
	prev := new(DeSchedulerOption)
	if err := json.Unmarshal(bs, prev); err != nil {
		return result, errors.Wrapf(err, "deepcopy unmarshal failed")
	}
	return *prev, nil
}

// ConfigInterface defines the interface of config. It can start watcher of config, and
// return the config changed message.
type ConfigInterface interface {
	Parse() error
	GetOptions() *DeSchedulerOption
	GetPrevOptions() *DeSchedulerOption
	Watch(ctx context.Context) error
	ChangedChan() <-chan struct{}
}

type handler struct {
	prevOp  *DeSchedulerOption
	op      *DeSchedulerOption
	watcher *fsnotify.Watcher

	changed chan struct{}
}

// newHandler create the instance of config handler
func newHandler() *handler {
	return &handler{
		op:      new(DeSchedulerOption),
		changed: make(chan struct{}, 1000),
	}
}

var (
	globalHandler *handler
)

// GlobalConfigHandler return the global config handler
func GlobalConfigHandler() ConfigInterface {
	if globalHandler == nil {
		globalHandler = newHandler()
	}
	return globalHandler
}

// Parse will parse the config file to struct
func (h *handler) Parse() error {
	if h.op != nil && h.op.ConfigFile != "" {
		prev, err := h.op.deepCopy()
		if err != nil {
			return errors.Wrapf(err, "parse failed when deepcopy")
		}
		h.prevOp = &prev

		bs, err := os.ReadFile(h.op.ConfigFile)
		if err != nil {
			return errors.Wrapf(err, "read config '%s' failed", h.op.ConfigFile)
		}
		if err := json.Unmarshal(bs, h.op); err != nil {
			return errors.Wrapf(err, "unmarshal failed")
		}
		blog.Infof("Using config file: %s, data: %s", h.op.ConfigFile, string(bs))
		return nil
	}

	conf.Parse(h.op)
	bs, err := json.Marshal(h.op)
	if err != nil {
		return errors.Wrapf(err, "marshal options failed")
	}
	blog.InitLogs(h.op.LogConfig)
	blog.Infof("Using config file: %s, data: %s", h.op.ConfigFile, string(bs))
	return nil
}

// GetOptions return current config options
func (h *handler) GetOptions() *DeSchedulerOption {
	return h.op
}

// GetPrevOptions return prev config options
func (h *handler) GetPrevOptions() *DeSchedulerOption {
	return h.prevOp
}

// Watch will watch the config file changed, and stopped with context done.
func (h *handler) Watch(ctx context.Context) (err error) {
	if err := h.createFileWatcher(); err != nil {
		return errors.Wrapf(err, "init filewatcher failed")
	}
	defer func() {
		blog.Infof("ConfigFile watcher is stopped.")
		_ = h.watcher.Close()
	}()
	for {
		select {
		case event, ok := <-h.watcher.Events:
			if !ok {
				return errors.Errorf("filewatcher channel closed with unknown")
			}
			if err := h.handleConfigChanged(&event); err != nil {
				blog.Errorf("FileWatcher config parse when changed failed: %s", err.Error())
			}
		case err, ok := <-h.watcher.Errors:
			if ok {
				return errors.Wrapf(err, "filewatcher stopped with error")
			}
		case <-ctx.Done():
			return nil
		}
	}
}

// ChangedChan return the channel that config file changed
func (h *handler) ChangedChan() <-chan struct{} {
	return h.changed
}

func (h *handler) createFileWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrapf(err, "new file watcher failed")
	}
	configFileDir := filepath.Dir(h.op.ConfigFile)
	err = watcher.Add(configFileDir)
	if err != nil {
		return errors.Wrapf(err, "add config filepath %s failed", h.op.ConfigFile)
	}
	blog.Infof("ConfigFile directory '%s' is added to watch.", configFileDir)
	h.watcher = watcher
	return nil
}

func (h *handler) handleConfigChanged(event *fsnotify.Event) error {
	blog.V(4).Infof("FileWatcher received event: %s", event.String())
	if event.Op&fsnotify.Write == fsnotify.Write && event.Name == h.op.ConfigFile {
		blog.Infof("FileWatcher watched file changed: %s", event.Name)
		if err := h.Parse(); err != nil {
			return errors.Wrapf(err, "parse options failed")
		}
		blog.Infof("FileWatcher config file changed success.")
		h.changed <- struct{}{}
	}
	return nil
}
