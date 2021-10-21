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

package nginx

import (
	"fmt"
	"html/template"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/metric"
	conf "github.com/Tencent/bk-bcs/bcs-services/bcs-loadbalance/template"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-loadbalance/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-loadbalance/util"
)

//NewManager create haproxy config file manager
func NewManager(binPath, cfgPath, generatePath, backupPath, templatePath string) conf.Manager {
	return &Manager{
		nginxBin:     binPath,
		cfgFile:      cfgPath,
		tmpDir:       generatePath,
		backupDir:    backupPath,
		templateFile: filepath.Join(templatePath, "nginx.conf.template"),
		healthInfo: metric.HealthMeta{
			IsHealthy:   conf.HealthStatusOK,
			Message:     conf.HealthStatusOKMsg,
			CurrentRole: metric.SlaveRole,
		},
	}
}

//Manager implements TemplateManager interface, control
//nginx config file generating, validation, backup and reloading
type Manager struct {
	nginxBin     string            //absolute path for haproxy executable binary
	cfgFile      string            //absolute path for haproxy cfg file
	backupDir    string            //absolute path for cfg file backup storage
	tmpDir       string            //temperary file for create new file
	templateFile string            //template file
	healthInfo   metric.HealthMeta //Health information
	healthLock   sync.RWMutex
}

//Start point, do not block
func (m *Manager) Start() error {
	//check template exist
	if !conf.IsFileExist(m.nginxBin) {
		blog.Error("nginx executable file lost")
		return fmt.Errorf("nginx executable file lost")
	}
	if !conf.IsFileExist(m.templateFile) {
		blog.Error("nginx.conf.template do not exist")
		return fmt.Errorf("nginx.conf.template do not exist")
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
	return nil
}

//Stop stop
func (m *Manager) Stop() {

}

//Create config file with tmpData,
func (m *Manager) Create(tmpData *types.TemplateData) (string, error) {
	//loading template file
	t, err := template.ParseFiles(m.templateFile)
	if err != nil {
		blog.Errorf("Parse template file %s failed: %s", m.templateFile, err.Error())
		return "", err
	}
	//create new config file
	fileName := "nginx." + strconv.Itoa(rand.Int()) + ".conf"
	absName := filepath.Join(m.tmpDir, fileName)
	writer, wErr := os.Create(absName)
	if wErr != nil {
		blog.Errorf("Create tempory new config file %s failed: %s", absName, wErr.Error())
		return "", wErr
	}
	//fix nginx vhost bug, 2018-09-26 12:12:41
	for i := range tmpData.HTTP {
		if len(tmpData.HTTP[i].BCSVHost) == 0 {
			blog.Warnf("nginx got empty http vhost info, %s", tmpData.HTTP[i].Name)
			continue
		}
		domains := strings.Split(tmpData.HTTP[i].BCSVHost, ":")
		tmpData.HTTP[i].BCSVHost = domains[0]
	}
	exErr := t.Execute(writer, tmpData)
	if exErr != nil {
		blog.Errorf("Template Execute Err: %s", exErr.Error())
		return "", exErr
	}
	blog.Infof("Create new nginx.conf %s success", absName)
	return absName, nil
}

//CheckDifference two file are difference, true is difference
func (m *Manager) CheckDifference(oldFile, curFile string) bool {
	var err error
	if !conf.IsFileExist(oldFile) {
		blog.Errorf("Old nginx.conf %s Do not exist", oldFile)
		return false
	}
	if !conf.IsFileExist(curFile) {
		blog.Errorf("Current nginx.conf %s Do not exist", oldFile)
		return false
	}
	//calculate oldFile md5
	oldMd5, err := util.Md5SumForFile(oldFile)
	if err != nil {
		blog.Errorf("calculate old nginx file %s md5sum failed, err %s", oldFile, err.Error())
		return false
	}
	//calculate curFile md5
	newMd5, err := util.Md5SumForFile(curFile)
	if err != nil {
		blog.Errorf("calculate cur nginx file %s md5sum failed, err %s", curFile, err.Error())
		return false
	}
	//compare
	if oldMd5 != newMd5 {
		blog.Info("New and old nginx.conf MD5 is difference")
		return true
	}
	return false
}

//Validate new cfg file grammar is OK
func (m *Manager) Validate(newFile string) bool {
	command := m.nginxBin + " -t -c " + newFile
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
	command := m.nginxBin + " -s reload"
	output, ok := util.ExeCommand(command)
	if !ok {
		blog.Errorf("Reload with command [%s] failed: %s", command, output)
		return fmt.Errorf("Reload config err")
	}
	blog.Infof("Reload with command %s, output: %s", command, output)
	return nil
}

// TryUpdateWithoutReload update haproxy config without reloading
// needReload: true for reload
func (m *Manager) TryUpdateWithoutReload(tmpData *types.TemplateData) (needReload bool) {
	// always reload
	return true
}
