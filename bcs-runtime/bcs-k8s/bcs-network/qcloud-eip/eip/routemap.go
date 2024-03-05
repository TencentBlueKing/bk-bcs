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

package eip

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// RouteMap store (route_table_id, ips[]) mapping
type RouteMap struct {
	FilePath string
	Data     map[string]int
	DataLock sync.Mutex
}

func newRouteMap(path string) *RouteMap {
	return &RouteMap{
		FilePath: path,
		Data:     make(map[string]int),
	}
}

func (rm *RouteMap) addRecord(ip string, routeTableID int) {
	rm.DataLock.Lock()
	defer rm.DataLock.Unlock()
	rm.Data[ip] = routeTableID
}

func (rm *RouteMap) loadFromFile() error {
	f, err := os.Open(rm.FilePath)
	if err != nil {
		blog.Errorf("open file %s failed, %s", rm.FilePath, err.Error())
		return fmt.Errorf("open file %s failed, %s", rm.FilePath, err.Error())
	}
	defer f.Close()
	allBytes, err := ioutil.ReadAll(f)
	if err != nil {
		blog.Errorf("read file %s failed, err %s", rm.FilePath, err.Error())
		return fmt.Errorf("read file %s failed, err %s", rm.FilePath, err.Error())
	}
	err = json.Unmarshal(allBytes, &rm.Data)
	if err != nil {
		blog.Errorf("json unmarshal %s failed, err %s", string(allBytes), err.Error())
		return fmt.Errorf("json unmarshal %s failed, err %s", string(allBytes), err.Error())
	}
	return nil
}

func (rm *RouteMap) saveToFile() error {
	bytes, err := json.Marshal(&rm.Data)
	if err != nil {
		blog.Errorf("json marshal %v failed, err %s", rm.Data, err.Error())
		return fmt.Errorf("json marshal %v failed, err %s", rm.Data, err.Error())
	}
	f, err := os.Create(rm.FilePath)
	if err != nil {
		blog.Errorf("open file %s failed, %s", rm.FilePath, err.Error())
		return fmt.Errorf("open file %s failed, %s", rm.FilePath, err.Error())
	}
	defer f.Close()

	_, err = f.Write(bytes)
	if err != nil {
		blog.Errorf("write %s to file %s failed, err %s", string(bytes), rm.FilePath, err.Error())
		return fmt.Errorf("write %s to file %s failed, err %s", string(bytes), rm.FilePath, err.Error())
	}
	return nil
}

func (rm *RouteMap) removeFile() error {
	err := os.Remove(rm.FilePath)
	if err != nil {
		blog.Errorf("failed to remove %s, err %s", rm.FilePath, err.Error())
		return fmt.Errorf("failed to remove %s, err %s", rm.FilePath, err.Error())
	}
	return nil
}
