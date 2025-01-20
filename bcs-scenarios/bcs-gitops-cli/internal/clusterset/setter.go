// nolint
// NOCC:tosa/license(设计如此)
//go:build linux
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
 */

package clusterset

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// ClusterSetter set the cluster to file
type ClusterSetter struct {
	rootDir string
}

func (s *ClusterSetter) preCheck() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "get user home dir failed")
	}
	dfp := path.Join(homeDir, defaultDirPath)
	fi, err := os.Stat(dfp)
	if err != nil {
		if os.IsNotExist(err) {
			// NOCC:gas/permission(设计如此)
			if err = os.Mkdir(dfp, 0755); err != nil {
				return errors.Wrapf(err, "create dir '%s' failed", dfp)
			}
		} else {
			return errors.Wrapf(err, "os.stat dir '%s' failed", dfp)
		}
	} else {
		if !fi.IsDir() {
			return errors.Errorf("%s should be a directory", dfp)
		}
	}
	s.rootDir = dfp
	return nil
}

func (s *ClusterSetter) getParentSessionID() (int, error) {
	pid := os.Getpid()
	statFile := fmt.Sprintf("/proc/%d/stat", pid)
	data, err := os.ReadFile(statFile)
	if err != nil {
		return 0, errors.Wrapf(err, "read file '%s failed'", statFile)
	}

	fields := strings.Fields(string(data))
	if len(fields) < 4 {
		return 0, fmt.Errorf("unexpected stat file format: %s", string(data))
	}
	ppid, err := strconv.Atoi(fields[3])
	if err != nil {
		return 0, errors.Wrapf(err, "convert pid from %s to int failed", fields[3])
	}
	parentStatFile := fmt.Sprintf("/proc/%d/stat", ppid)
	parentData, err := os.ReadFile(parentStatFile)
	if err != nil {
		return 0, errors.Wrapf(err, "read file '%s' failed", parentStatFile)
	}
	parentFields := strings.Fields(string(parentData))
	if len(parentFields) < 5 {
		return 0, fmt.Errorf("unexpected parent stat file format: %s", string(parentData))
	}

	// 获取父进程的会话 ID (SID)
	sid, err := strconv.Atoi(parentFields[4])
	if err != nil {
		return 0, errors.Wrapf(err, "convert pid from %s to int failed", parentFields[4])
	}
	return sid, nil
}

func (s *ClusterSetter) getSessionFile(fileName string) string {
	return path.Join(s.rootDir, fileName)
}

func (s *ClusterSetter) createSessionFile(fileName string) (string, error) {
	sessionFile := path.Join(s.rootDir, fileName)
	_ = os.RemoveAll(sessionFile) // nolint
	fi, err := os.Create(sessionFile)
	if err != nil {
		return "", errors.Wrapf(err, "create file '%s' failed", sessionFile)
	}
	defer fi.Close()
	return sessionFile, nil
}

// SetCluster set the cluster to global cluster file
func (s *ClusterSetter) SetCluster(cluster *ClusterInfo) error {
	if err := s.preCheck(); err != nil {
		return errors.Wrapf(err, "pre-check failed")
	}
	parentSessionID, err := s.getParentSessionID()
	if err != nil {
		return errors.Wrapf(err, "get parent session id failed")
	}
	bs, err := json.Marshal(cluster)
	if err != nil {
		return errors.Wrapf(err, "marshal cluster failed")
	}
	sessionFile := fmt.Sprintf(tmpClusterFile, parentSessionID)
	if sessionFile, err = s.createSessionFile(sessionFile); err != nil {
		return errors.Wrapf(err, "create session file failed")
	}
	if err = os.WriteFile(sessionFile, bs, 0644); err != nil {
		return errors.Wrapf(err, "set global cluster failed")
	}
	return nil
}

// SetClusterGlobal set the cluster to global cluster file
func (s *ClusterSetter) SetClusterGlobal(cluster *ClusterInfo) error {
	if err := s.preCheck(); err != nil {
		return errors.Wrapf(err, "pre-check failed")
	}
	bs, err := json.Marshal(cluster)
	if err != nil {
		return errors.Wrapf(err, "marshal cluster failed")
	}
	var globalSessionFile string
	if globalSessionFile, err = s.createSessionFile(globalClusterFile); err != nil {
		return errors.Wrapf(err, "create session file failed")
	}
	if err = os.WriteFile(globalSessionFile, bs, 0644); err != nil {
		return errors.Wrapf(err, "set global cluster failed")
	}
	return nil
}

func (s *ClusterSetter) readSessionFile(sessionFile string) (*ClusterInfo, error) {
	bs, err := os.ReadFile(sessionFile)
	if err == nil {
		if strings.TrimSpace(string(bs)) == "" {
			return nil, nil
		}
		info := new(ClusterInfo)
		if err = json.Unmarshal(bs, info); err != nil {
			return info, errors.Wrapf(err, "unmarshal cluster info failed")
		}
		return info, nil
	} else {
		if !os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "read file '%s' failed", sessionFile)
		}
	}
	return nil, nil
}

// GetCurrentCluster get the cluster from env or global cluster file
func (s *ClusterSetter) GetCurrentCluster() (string, error) {
	if err := s.preCheck(); err != nil {
		return "", errors.Wrapf(err, "pre-check failed")
	}
	parentSessionID, err := s.getParentSessionID()
	if err != nil {
		return "", errors.Wrapf(err, "get parent session id failed")
	}
	tmpSessionFile := s.getSessionFile(fmt.Sprintf(tmpClusterFile, parentSessionID))
	clsInfo, err := s.readSessionFile(tmpSessionFile)
	if err != nil {
		return "", errors.Wrapf(err, "read session file '%s' failed", tmpSessionFile)
	}
	if clsInfo != nil {
		return clsInfo.ClusterID, nil
	}

	globalSessionFile := s.getSessionFile(globalClusterFile)
	clsInfo, err = s.readSessionFile(globalSessionFile)
	if err != nil {
		return "", errors.Wrapf(err, "read session file '%s' failed", globalSessionFile)
	}
	if clsInfo != nil {
		return clsInfo.ClusterID, nil
	}
	return "", nil
}

// ReturnClusterInfo return cluster info
func (s *ClusterSetter) ReturnClusterInfo() ([]*ClusterInfo, error) {
	if err := s.preCheck(); err != nil {
		return nil, errors.Wrapf(err, "pre-check failed")
	}
	result := make([]*ClusterInfo, 0)

	parentSessionID, err := s.getParentSessionID()
	if err != nil {
		return nil, errors.Wrapf(err, "get parent session id failed")
	}
	tmpSessionFile := s.getSessionFile(fmt.Sprintf(tmpClusterFile, parentSessionID))
	clsInfo, err := s.readSessionFile(tmpSessionFile)
	if err != nil {
		return nil, errors.Wrapf(err, "read session file '%s' failed", tmpSessionFile)
	}
	if clsInfo != nil {
		clsInfo.Status = "(current-session)"
		result = append(result, clsInfo)
	}

	globalSessionFile := s.getSessionFile(globalClusterFile)
	clsInfo, err = s.readSessionFile(globalSessionFile)
	if err != nil {
		return nil, errors.Wrapf(err, "read session file '%s' failed", globalSessionFile)
	}
	if clsInfo != nil {
		clsInfo.Status = "(global-session)"
		result = append(result, clsInfo)
	}
	return result, nil
}
