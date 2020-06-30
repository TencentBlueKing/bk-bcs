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

package store

import (
	"encoding/json"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-executor/process-executor/types"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type store struct {
	dataDir        string
	processInfoDir string
}

func NewStore(dir string) Store {
	return &store{
		dataDir:        dir,
		processInfoDir: path.Join(dir, "processinfos"),
	}
}

func (s *store) StoreProcessInfo(processInfo *types.ProcessInfo) error {
	file, err := os.OpenFile(path.Join(s.processInfoDir, fmt.Sprintf("%s.info", processInfo.Id)), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	blog.V(3).Infof("store process %s", processInfo.Id)
	err = file.Truncate(0)
	if err != nil {
		return err
	}
	by, _ := json.Marshal(processInfo)
	_, err = file.Write(by)
	return err
}

func (s *store) DeleteProcessInfo(processInfo *types.ProcessInfo) error {
	err := os.Remove(path.Join(s.processInfoDir, fmt.Sprintf("%s.info", processInfo.Id)))
	if err != nil {
		blog.Errorf("store delete processinfo %s error %s", processInfo.Id, err.Error())
	}

	return nil
}

func (s *store) GetAllProcessInfos() ([]*types.ProcessInfo, error) {
	files, err := ioutil.ReadDir(s.processInfoDir)
	if err != nil {
		return nil, err
	}

	processInfos := make([]*types.ProcessInfo, 0)
	for _, file := range files {
		if !strings.Contains(file.Name(), ".info") {
			continue
		}

		p := path.Join(s.processInfoDir, file.Name())
		blog.Infof("read processinfo file %s", p)
		by, err := ioutil.ReadFile(p)
		if err != nil {
			blog.Errorf("read file %s error %s", p, err.Error())
			continue
		}

		var process *types.ProcessInfo
		err = json.Unmarshal(by, &process)
		if err != nil {
			return nil, fmt.Errorf("Unmarshal data %s to types.ProcessInfo error %s", string(by), err.Error())
		}

		processInfos = append(processInfos, process)
	}

	return processInfos, nil
}
