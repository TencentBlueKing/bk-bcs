// +build linux
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
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	docker "github.com/fsouza/go-dockerclient"
)

func (s *SidecarController) reloadLogbeat() error {
	return nil
}

func (s *SidecarController) getActualPath(logPath string, container *docker.Container) (string, error) {
	if !filepath.IsAbs(logPath) {
		blog.Errorf("log path specified as \"%s\" is not an absolute path", logPath)
		return "", fmt.Errorf("log path specified as \"%s\" is not an absolute path", logPath)
	}
	var mounted = false
	var retpath string
	for _, mountinfo := range container.Mounts {
		if strings.HasPrefix(logPath, mountinfo.Destination+string(filepath.Separator)) {
			mounted = true
			retpath = mountinfo.Source + strings.TrimPrefix(logPath, mountinfo.Destination)
			break
		}
	}
	if !mounted {
		retpath = s.getContainerRootPath(container) + logPath
	}
	blog.V(3).Infof("origin path: %s, mounted path: %s", logPath, retpath)
	retpath = s.getCleanPath(retpath)
	blog.V(3).Infof("origin path: %s, clean path: %s", logPath, retpath)
	return retpath, nil
}
