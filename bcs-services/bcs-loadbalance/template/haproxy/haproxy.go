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

package haproxy

import (
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/metric"
	conf "bk-bcs/bcs-services/bcs-loadbalance/template"
	"bk-bcs/bcs-services/bcs-loadbalance/types"
	"bk-bcs/bcs-services/bcs-loadbalance/util"
	"fmt"
	"html/template"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

//NewManager create haproxy config file manager
func NewManager(binPath, cfgPath, generatePath, backupPath, templatePath string, statusFetchPeriod int) conf.Manager {
	return &Manager{
		haproxyBin:        binPath,
		cfgFile:           cfgPath,
		tmpDir:            generatePath,
		backupDir:         backupPath,
		templateFile:      filepath.Join(templatePath, "haproxy.cfg.template"),
		statusFetchPeriod: statusFetchPeriod,
		stopCh:            make(chan struct{}),
		healthInfo: metric.HealthMeta{
			IsHealthy:   conf.HealthStatusOK,
			Message:     conf.HealthStatusOKMsg,
			CurrentRole: metric.SlaveRole,
		},
	}
}

//Manager implements TemplateManager interface, control
//haproxy config file generating, validation, backup and reloading
type Manager struct {
	haproxyBin        string            //absolute path for haproxy executable binary
	cfgFile           string            //absolute path for haproxy cfg file
	backupDir         string            //absolute path for cfg file backup storage
	tmpDir            string            //temporary file for create new file
	templateFile      string            //template file
	statusFetchPeriod int               // period for fetch haproxy stats data
	stopCh            chan struct{}     // chan for stop fetching haproxy stats data
	stats             *Status           //stats for haproxy
	statsMutex        sync.Mutex        // lock for stats haproxy
	healthInfo        metric.HealthMeta //Health information
	healthLock        sync.RWMutex
}

//Start point, do not block
func (m *Manager) Start() error {
	//check template exist
	if !conf.IsFileExist(m.haproxyBin) {
		blog.Error("haproxy executable file lost")
		return fmt.Errorf("haproxy executable file lost")
	}
	if !conf.IsFileExist(m.templateFile) {
		blog.Error("haproxy.cfg.template do not exist")
		return fmt.Errorf("haproxy.cfg.template do not exist")
	}
	//create other file directory
	err := os.MkdirAll(m.backupDir, os.ModePerm)
	if err != nil {
		blog.Warnf("mkdir %s failed, err %s", m.backupDir, err.Error())
	}
	err = os.MkdirAll(m.tmpDir, os.ModePerm)
	if err != nil {
		blog.Warnf("mkdir %s failed, err %s", m.tmpDir, err.Error())
	}
	// run haproxy status fetcher
	go m.runStatusFetch()
	return nil
}

//Stop stop
func (m *Manager) Stop() {
	close(m.stopCh)
}

//Create config file with tmpData,
func (m *Manager) Create(tmpData *types.TemplateData) (string, error) {
	var err error
	var t *template.Template
	var writer *os.File
	//loading template file
	t, err = template.ParseFiles(m.templateFile)
	if err != nil {
		blog.Errorf("Parse template file %s failed: %s", m.templateFile, err.Error())
		return "", err
	}
	//create new config file
	fileName := "haproxy." + strconv.Itoa(rand.Int()) + ".cfg"
	absName := filepath.Join(m.tmpDir, fileName)
	writer, err = os.Create(absName)
	if err != nil {
		blog.Errorf("Create tempory new config file %s failed: %s", absName, err.Error())
		return "", err
	}
	err = t.Execute(writer, tmpData)
	if err != nil {
		blog.Errorf("Template Execute Err: %s", err.Error())
		return "", err
	}
	blog.Infof("Create new haproxy.cfg %s success", absName)
	return absName, nil
}

//CheckDifference two file are difference, true is difference
func (m *Manager) CheckDifference(oldFile, curFile string) bool {
	if !conf.IsFileExist(oldFile) {
		blog.Errorf("Old haproxy.cfg %s Do not exist", oldFile)
		return false
	}
	if !conf.IsFileExist(curFile) {
		blog.Errorf("Current haproxy.cfg %s Do not exist", oldFile)
		return false
	}
	//calculate oldFile md5
	oldMd5, err := util.Md5SumForFile(oldFile)
	if err != nil {
		blog.Errorf("calculate old haproxy file %s md5sum failed, err %s", oldFile, err.Error())
		return false
	}
	//calculate curFile md5
	newMd5, err := util.Md5SumForFile(curFile)
	if err != nil {
		blog.Errorf("calculate cur haproxy file %s md5sum failed, err %s", curFile, err.Error())
		return false
	}
	//compare
	if oldMd5 != newMd5 {
		blog.Info("New and old haproxy.cfg MD5 is difference")
		return true
	}
	return false
}

//Validate new cfg file grammar is OK
func (m *Manager) Validate(newFile string) bool {
	command := m.haproxyBin + " -c -f " + newFile
	output, ok := util.ExeCommand(command)
	if !ok {
		blog.Errorf("Validate with command [%s] failed", command)
		return false
	}
	blog.Infof("Validate with command %s, output: %s", command, output)
	return true
}

//Replace old cfg file with cur one, return old file backup
func (m *Manager) Replace(oldFile, curFile string) error {
	return util.ReplaceFile(oldFile, curFile)
}

//Reload haproxy with new config file
func (m *Manager) Reload(cfgFile string) error {
	command := m.haproxyBin + " -f " + cfgFile + " -sf $(cat /var/run/haproxy.pid)"
	output, ok := util.ExeCommand(command)
	if !ok {
		blog.Errorf("Reload with command [%s] failed: %s", command, output)
		return fmt.Errorf("Reload config err")
	}
	blog.Infof("Reload with command %s, output: %s", command, output)
	return nil
}
