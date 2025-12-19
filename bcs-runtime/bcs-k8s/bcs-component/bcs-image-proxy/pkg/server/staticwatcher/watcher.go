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

package staticwatcher

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/options"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/store"
)

// StaticFilesWatcher defines the static file watcher
type StaticFilesWatcher struct {
	op         *options.ImageProxyOption
	cacheStore store.CacheStore
}

// NewStaticFileWatcher create the static file watcher instance
func NewStaticFileWatcher() *StaticFilesWatcher {
	return &StaticFilesWatcher{
		op:         options.GlobalOptions(),
		cacheStore: store.GlobalRedisStore(),
	}
}

// Init the static file watcher
func (w *StaticFilesWatcher) Init(ctx context.Context) error {
	staticPaths := []string{w.op.TransferPath, w.op.SmallFilePath, w.op.OCIPath}
	for _, sp := range staticPaths {
		if err := w.initLayerFiles(ctx, sp); err != nil {
			return errors.Wrapf(err, "init static-layers for '%s' failed", sp)
		}
	}
	return nil
}

func (w *StaticFilesWatcher) initLayerFiles(ctx context.Context, filePath string) error {
	if err := filepath.Walk(filePath, func(fp string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".tar.gzip") {
			return nil
		}
		digest := strings.TrimSuffix(info.Name(), ".tar.gzip")
		if err = w.cacheStore.SaveStaticLayer(ctx, digest, fp, false); err != nil {
			logctx.Errorf(ctx, "cache save static '%s' failed: %s", fp, err.Error())
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// Watch the static file
func (w *StaticFilesWatcher) Watch(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(err, "create file watcher failed")
	}
	defer watcher.Close()
	done := make(chan struct{})
	staticPaths := []string{w.op.TransferPath, w.op.SmallFilePath, w.op.OCIPath}
	ticker := time.NewTicker(120 * time.Second)
	defer ticker.Stop()
	go func() {
		defer close(done)
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if !strings.HasSuffix(event.Name, ".tar.gzip") {
					continue
				}
				if event.Op != fsnotify.Create && event.Op != fsnotify.Remove {
					continue
				}
				digest := strings.TrimSuffix(path.Base(event.Name), ".tar.gzip")
				switch event.Op {
				case fsnotify.Create:
					if err = w.cacheStore.SaveStaticLayer(ctx, digest, event.Name, true); err != nil {
						logctx.Errorf(ctx, "cache save static '%s' failed: %s", event.Name, err.Error())
					}
				case fsnotify.Remove:
					if err = w.cacheStore.DeleteStaticLayer(ctx, digest); err != nil {
						logctx.Errorf(ctx, "cache delete static '%s' failed: %s", event.Name, err.Error())
					}
				default:
				}
			case <-ticker.C:
				for _, sp := range staticPaths {
					if err = w.initLayerFiles(ctx, sp); err != nil {
						logctx.Errorf(ctx, "save static-layers for '%s' failed: %s", sp, err.Error())
					}
				}
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	for _, sp := range staticPaths {
		if err = watcher.Add(sp); err != nil {
			return errors.Wrapf(err, "add file watcher '%s' failed", sp)
		}
	}
	<-done
	return nil
}
