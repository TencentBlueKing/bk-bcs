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
	"context"
	"encoding/json"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	schStore "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	v2 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var cmdLocks map[string]*sync.Mutex
var cmdRWlock sync.RWMutex

// InitCmdLockPool init command lock pool
func (store *managerStore) InitCmdLockPool() {
	if cmdLocks == nil {
		blog.Info("init command lock pool")
		cmdLocks = make(map[string]*sync.Mutex)
	}
}

// LockCommand lock command
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

// UnLockCommand unlock command
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

// CheckCommandExist check if a command exists
func (store *managerStore) CheckCommandExist(command *commtypes.BcsCommandInfo) (string, bool) {
	v2Cmd, _ := store.FetchCommand(command.Id)
	if v2Cmd != nil {
		return v2Cmd.ResourceVersion, true
	}

	return "", false
}

// SaveCommand save command to db
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
		v2Cmd, err = client.Update(context.Background(), v2Cmd, metav1.UpdateOptions{})
	} else {
		v2Cmd, err = client.Create(context.Background(), v2Cmd, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	command.ResourceVersion = v2Cmd.ResourceVersion
	saveCacheCommand(command)
	return err
}

// FetchCommand fetch command from cache
func (store *managerStore) FetchCommand(ID string) (*commtypes.BcsCommandInfo, error) {
	cmd := getCacheCommand(ID)
	if cmd == nil {
		return nil, schStore.ErrNoFound
	}

	return cmd, nil
}

// DeleteCommand delete command
func (store *managerStore) DeleteCommand(ID string) error {
	client := store.BkbcsClient.BcsCommandInfos(DefaultNamespace)
	err := client.Delete(context.Background(), ID, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	deleteCacheCommand(ID)
	return nil
}

// list all commands from etcd
func (store *managerStore) listAllCommands() ([]*commtypes.BcsCommandInfo, error) {
	client := store.BkbcsClient.BcsCommandInfos("")
	v2cmd, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	cmds := make([]*commtypes.BcsCommandInfo, 0, len(v2cmd.Items))
	for _, cmd := range v2cmd.Items {
		obj := cmd.Spec.BcsCommandInfo
		obj.ResourceVersion = cmd.ResourceVersion
		cmds = append(cmds, &obj)
	}
	return cmds, nil
}
