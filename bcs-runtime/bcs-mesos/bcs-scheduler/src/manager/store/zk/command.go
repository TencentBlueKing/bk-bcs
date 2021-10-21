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

package zk

import (
	"encoding/json"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
)

var cmdLocks map[string]*sync.Mutex
var cmdRWlock sync.RWMutex

func (store *managerStore) InitCmdLockPool() {
	if cmdLocks == nil {
		blog.Info("init command lock pool")
		cmdLocks = make(map[string]*sync.Mutex)
	}
}

func (store *managerStore) LockCommand(cmdId string) {

	cmdRWlock.RLock()
	myLock, ok := cmdLocks[cmdId]
	cmdRWlock.RUnlock()
	if ok {
		myLock.Lock()
		return
	}

	cmdRWlock.Lock()
	myLock, ok = cmdLocks[cmdId]
	if !ok {
		blog.Info("create command lock(%s), current locknum(%d)", cmdId, len(cmdLocks))
		cmdLocks[cmdId] = new(sync.Mutex)
		myLock, _ = cmdLocks[cmdId]
	}
	cmdRWlock.Unlock()

	myLock.Lock()
	return
}

func (store *managerStore) UnLockCommand(cmdId string) {
	cmdRWlock.RLock()
	myLock, ok := cmdLocks[cmdId]
	cmdRWlock.RUnlock()

	if !ok {
		blog.Error("command lock(%s) not exist when do unlock", cmdId)
		return
	}
	myLock.Unlock()
}

func getCommandRootPath() string {
	return "/" + bcsRootNode + "/" + commandNode
}

func (store *managerStore) SaveCommand(command *commtypes.BcsCommandInfo) error {

	data, err := json.Marshal(command)
	if err != nil {
		return err
	}

	path := getCommandRootPath() + "/" + command.Id

	return store.Db.Insert(path, string(data))
}

func (store *managerStore) FetchCommand(ID string) (*commtypes.BcsCommandInfo, error) {

	path := getCommandRootPath() + "/" + ID

	data, err := store.Db.Fetch(path)
	if err != nil {
		blog.Info("get path(%s) err:%s", path, err.Error())
		return nil, err
	}

	command := &commtypes.BcsCommandInfo{}
	if err := json.Unmarshal(data, command); err != nil {
		blog.Error("fail to unmarshal command(%s). err:%s", string(data), err.Error())
		return nil, err
	}

	return command, nil
}

func (store *managerStore) DeleteCommand(ID string) error {

	path := getCommandRootPath() + "/" + ID
	if err := store.Db.Delete(path); err != nil {
		blog.Error("fail to delete command(%s) err:%s", path, err.Error())
		return err
	}

	return nil
}
