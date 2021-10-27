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

package clear

import (
	"io/ioutil"
	"os"
	"sort"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

type fileInfo []os.FileInfo

//Less sort Less interface
func (fi fileInfo) Less(i, j int) bool {
	return fi[i].ModTime().Unix() < fi[j].ModTime().Unix()
}

//Len sort Len interface
func (fi fileInfo) Len() int {
	return len(fi)
}

//Swap sort swap interface
func (fi fileInfo) Swap(i, j int) {
	fi[i], fi[j] = fi[j], fi[i]
}

//NewClearManager new a ClearManager
func NewClearManager() *Manager {
	return &Manager{
		exit: make(chan struct{}),
	}
}

//Manager timer clear go runtinue
type Manager struct {
	exit chan struct{} //flag for processor exit
}

//Start start the runtinue
func (cm *Manager) Start() {
	go cm.run()
}

//main goruntine to run timer to clear older files
//and cacth the signal to exit
func (cm *Manager) run() {
	tick := time.NewTicker(time.Second * time.Duration(int64(120)))
	defer tick.Stop()
	for {
		select {
		case <-cm.exit:
			blog.Infof("Manager Get close event, return")
			return
		case <-tick.C:
			blog.Infof("begin to clearFiles")
			//clear oldest template file when files more than 100
			cm.clearFiles()
			blog.Infof("finish clearFiles")
		}
	}
}

//clearFiles clear files more than 100,sort then by time,oldest in the list head
func (cm *Manager) clearFiles() {
	files, e := ioutil.ReadDir("/bcs-lb/generate")
	if e != nil {
		blog.Errorf("ReadDir failed:%s", e.Error())
		return
	}
	fileNum := len(files)
	blog.Infof("files num : %d", fileNum)
	if fileNum > 100 {
		clearNum := fileNum - 100
		blog.Infof("clear num : %d", clearNum)
		sort.Sort(fileInfo(files))
		for i, f := range files {
			if i < (clearNum - 1) {
				err := os.Remove("/bcs-lb/generate/" + f.Name())
				if err != nil {
					blog.Warnf("remove %s failed:%s", f.Name(), err.Error())
				} else {
					blog.Infof("remove %s successfully", f.Name())
				}
			} else {
				blog.Infof("remove %d files this timer", clearNum)
				break
			}
		}
	}
}

//Stop stop the runtinue
func (cm *Manager) Stop() {
	close(cm.exit)
}
