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

package sidecar

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	docker "github.com/fsouza/go-dockerclient"
)

func (s *SidecarController) syncContainerCache() {
	s.syncContainerCacheOnce()
	go func() {
		blog.Infof("Start sync containerInfoCache periodly")
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
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
	s.containerCache = make(map[string]*docker.Container)

	//list all running containers
	apiContainers, err := s.client.ListContainers(docker.ListContainersOptions{All: true})
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

func (s *SidecarController) inspectContainer(ID string) *docker.Container {
	inspectChan := make(chan interface{}, 1)
	timer := time.NewTimer(3 * time.Second)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		defer close(inspectChan)
		c, err := s.client.InspectContainerWithOptions(docker.InspectContainerOptions{ID: ID, Context: ctx})
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
			blog.Errorf("docker InspectContainer %s failed: %s", ID, err.Error())
			return nil
		}
		c, ok := obj.(*docker.Container)
		if !ok {
			blog.Errorf("docker InspectContainer %s failed with type(%T) returned, expected *docker.Container", ID, obj)
			return nil
		}
		return c
	case <-timer.C:
		cancelFunc()
		blog.Errorf("Inspect container %d timeout unexpected, check whether pod is working properly with sharePID mode.", ID)
		return nil
	}
}
