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

package deviceplugin

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/fsnotify/fsnotify"
)

// DevicePluginOp operate device plugin lifecycle, listen changes of kubelet socket file
type DevicePluginOp struct {
	kubeletSockPath string
	plugin          *EniDevicePlugin
	stopCh          chan struct{}
}

// NewDevicePluginOp create device plugin operator
func NewDevicePluginOp(kubeletSocketPath, pluginSocketPath, resourceName string) *DevicePluginOp {
	return &DevicePluginOp{
		kubeletSockPath: kubeletSocketPath,
		plugin:          NewEniDevicePlugin(kubeletSocketPath, pluginSocketPath, resourceName),
		stopCh:          make(chan struct{}),
	}
}

func (op *DevicePluginOp) startPlugin() error {
	if op.plugin == nil {
		return fmt.Errorf("plugin is empty")
	}
	if err := op.plugin.Start(); err != nil {
		return fmt.Errorf("failed to start device plugin, err %s", err.Error())
	}
	return nil
}

func (op *DevicePluginOp) stopPlugin() {
	if op.plugin == nil {
		return
	}
	if err := op.plugin.Stop(); err != nil {
		blog.Warnf("failed stop plugin, err %s", err.Error())
	}
	blog.Infof("stop plugin successfully")
}

// GetPlugin get device plugin object
func (op *DevicePluginOp) GetPlugin() *EniDevicePlugin {
	return op.plugin
}

// Start start device plugin operator
func (op *DevicePluginOp) Start() {
	blog.Infof("create file watcher")
	fileWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		blog.Errorf("create file watcher failed, err %s, will retry after few second", err.Error())
		time.Sleep(5 * time.Second)
		go op.Start()
		return
	}
	defer fileWatcher.Close()
	err = fileWatcher.Add(op.kubeletSockPath)
	if err != nil {
		blog.Errorf("add file %s to watcher failed, err %s, will retry after few second",
			op.kubeletSockPath, err.Error())
		time.Sleep(5 * time.Second)
		go op.Start()
		return
	}

	// restart plugin
	op.stopPlugin()
	if err := op.startPlugin(); err != nil {
		blog.Errorf("start plugin failed, will retry after few second, err %s", err.Error())
		time.Sleep(3 * time.Second)
		go op.Start()
		return
	}

	for {
		select {
		case we := <-fileWatcher.Events:
			// kubelet socket path event and event type is create
			if we.Name == op.kubeletSockPath && (we.Op&fsnotify.Remove) == fsnotify.Remove {
				blog.Infof("file watcher event: kubelet socket file %s removed, try to restart device-plugin",
					op.kubeletSockPath)
				time.Sleep(2 * time.Second)
				go op.Start()
				return
			}
		case err := <-fileWatcher.Errors:
			blog.Warnf("file watcher errors: %s", err.Error())
			go op.Start()
			return

		case <-op.stopCh:
			blog.Infof("device operator receive stop signal")
			op.stopPlugin()
			return
		}
	}
}

// Stop stop
func (op *DevicePluginOp) Stop() {
	close(op.stopCh)
}
