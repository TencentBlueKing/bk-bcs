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

package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// StopFunc define subsystem graceful stop interface
type StopFunc func()

// StartSignalHandler trap system signal for exit
func StartSignalHandler(stop context.CancelFunc, gracefulExit int) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-ch
	blog.Infof("server received stop signal.")
	// trap system signal, stop
	stop()
	tick := time.NewTicker(time.Second * time.Duration(gracefulExit))
	select {
	case <-ch:
		// double kill, just terminate immediately
		os.Exit(-1)
	case <-tick.C:
		// timeout
		return
	}
}

// DeepCopyMap will deepcopy the map
func DeepCopyMap(m map[string]string) map[string]string {
	r := make(map[string]string)
	if len(m) == 0 {
		return r
	}
	for k, v := range m {
		r[k] = v
	}
	return r
}
