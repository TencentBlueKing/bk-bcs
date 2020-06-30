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

package etcd

import (
	"encoding/json"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/pkg/apis/bkbcs/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (store *managerStore) CheckCommandExist(command *commtypes.BcsCommandInfo) (string, bool) {
	client := store.BkbcsClient.BcsCommandInfos(DefaultNamespace)
	v2Cmd, err := client.Get(command.Id, metav1.GetOptions{})
	if err == nil {
		return v2Cmd.ResourceVersion, true
	}

	return "", false
}

func (store *managerStore) SaveCommand(command *commtypes.BcsCommandInfo) error {
	client := store.BkbcsClient.BcsCommandInfos(DefaultNamespace)
	v2Cmd := &v2.BcsCommandInfo{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdBcsCommandInfo,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      command.Id,
			Namespace: DefaultNamespace,
		},
		Spec: v2.BcsCommandInfoSpec{
			BcsCommandInfo: *command,
		},
	}

	by, _ := json.Marshal(v2Cmd)
	blog.Infof("BcsCommandInfo %s", string(by))
	var err error
	rv, exist := store.CheckCommandExist(command)
	if exist {
		v2Cmd.ResourceVersion = rv
		_, err = client.Update(v2Cmd)
	} else {
		_, err = client.Create(v2Cmd)
	}
	return err
}

func (store *managerStore) FetchCommand(ID string) (*commtypes.BcsCommandInfo, error) {
	client := store.BkbcsClient.BcsCommandInfos(DefaultNamespace)
	v2Cmd, err := client.Get(ID, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &v2Cmd.Spec.BcsCommandInfo, nil
}

func (store *managerStore) DeleteCommand(ID string) error {
	client := store.BkbcsClient.BcsCommandInfos(DefaultNamespace)
	err := client.Delete(ID, &metav1.DeleteOptions{})
	return err
}
