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

// Package clusterset xxx
package clusterset

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/pkg/errors"
)

// Setter set the cluster to file
type Setter struct {
	rootDir              string
	bcsGlobalClusterFile string
}

const (
	defaultDirPath    = "./.bcs"
	globalClusterFile = "bcs_cluster_global"

	EnvBcsCluster = "CLUSTER"
)

func (s *Setter) preCheck() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "get user home dir failed")
	}
	dfp := path.Join(homeDir, defaultDirPath)
	fi, err := os.Stat(dfp)
	if err != nil {
		if os.IsNotExist(err) {
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

	gcfp := path.Join(s.rootDir, globalClusterFile)
	fi, err = os.Stat(gcfp)
	if err != nil {
		if os.IsNotExist(err) {
			var gcfi *os.File
			gcfi, err = os.Create(gcfp)
			if err != nil {
				return errors.Wrapf(err, "create file '%s' failed", gcfp)
			}
			defer gcfi.Close()
		} else {
			return errors.Wrapf(err, "os.stat file '%s' failed", gcfp)
		}
	} else {
		if fi.IsDir() {
			return errors.Errorf("%s should be a file", gcfp)
		}
	}
	s.bcsGlobalClusterFile = gcfp
	return nil
}

// SetCluster set the cluster to global cluster file
func (s *Setter) SetCluster(cluster string) error {
	if err := s.preCheck(); err != nil {
		return errors.Wrapf(err, "pre-check failed")
	}
	c := exec.Command("bash", "-c", fmt.Sprintf(`echo "%s" > %s`, cluster, s.bcsGlobalClusterFile))
	if _, err := c.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "set global cluster failed")
	}
	return nil
}

// GetCurrentCluster get the cluster from env or global cluster file
func (s *Setter) GetCurrentCluster() (string, error) {
	if err := s.preCheck(); err != nil {
		return "", errors.Wrapf(err, "pre-check failed")
	}
	clusterID := strings.TrimSpace(os.Getenv(EnvBcsCluster))
	if clusterID != "" {
		return clusterID, nil
	}
	bs, err := os.ReadFile(s.bcsGlobalClusterFile)
	if err != nil {
		return "", errors.Wrapf(err, "read global cluster file '%s' failed", s.bcsGlobalClusterFile)
	}
	return strings.TrimSpace(string(bs)), nil
}
