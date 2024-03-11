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

// Package sidecar xxx
package sidecar

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	dockertypes "github.com/docker/docker/api/types"
)

func (s *SidecarController) syncContainerCache() {
	s.syncContainerCacheOnce()
	go func() {
		blog.Infof("Start sync containerInfoCache periodly")
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for { // nolint
			select {
			case <-ticker.C:
				s.syncContainerCacheOnce()
			}
		}
	}()
}

func (s *SidecarController) syncContainerCacheOnce() {
	s.containerCacheMutex.Lock()
	defer s.containerCacheMutex.Unlock()
	s.containerCache = make(map[string]*dockertypes.ContainerJSON)

	// list all running containers
	apiContainers, err := s.client.ContainerList(context.Background(), dockertypes.ContainerListOptions{All: true})
	if err != nil {
		blog.Errorf("docker ListContainers failed: %s", err.Error())
		return
	}

	// generate container log config
	for _, apiC := range apiContainers {
		c := s.inspectContainer(apiC.ID)
		if c == nil {
			blog.Errorf("inspect container %s failed", apiC.ID)
			continue
		}
		s.containerCache[apiC.ID] = c
	}
}

func (s *SidecarController) inspectContainer(id string) *dockertypes.ContainerJSON {
	inspectChan := make(chan interface{}, 1)
	timer := time.NewTimer(3 * time.Second)
	// NOCC:vet/vet(工具误报:函数末尾有用到cancelFunc)
	ctx, cancelFunc := context.WithCancel(context.Background()) // nolint cancelFunc is not used on all paths
	go func() {
		defer close(inspectChan)
		c, err := s.client.ContainerInspect(ctx, id)
		if err != nil {
			inspectChan <- err
			return
		}
		inspectChan <- c
	}()

	select {
	case obj := <-inspectChan:
		if !timer.Stop() {
			<-timer.C
		}
		err, ok := obj.(error)
		if ok {
			blog.Errorf("docker InspectContainer %s failed: %s", id, err.Error())
			// NOCC:vet/vet(设计如此:不需要使用cancelFunc)
			return nil // nolint without using the cancelFunc
		}
		c, ok := obj.(dockertypes.ContainerJSON)
		if !ok {
			blog.Errorf("docker InspectContainer %s failed with type(%T) returned, expected dockertypes.ContainerJSON", id, obj)
			return nil // nolint without using the cancelFunc
		}
		return &c
	case <-timer.C:
		cancelFunc()
		blog.Errorf("Inspect container %s timeout unexpected, check whether pod is working properly with sharePID mode.", id)
		return nil
	}
}
