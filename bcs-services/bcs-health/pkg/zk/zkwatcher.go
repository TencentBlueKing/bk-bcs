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
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
)

type ZkEventHandlerFunc struct {
	AddFunc    func(branch, leaf, value string)
	UpdateFunc func(branch, leaf, oldvalue, newvalue string)
	DeleteFunc func(branch, leaf, value string)
}

func (f ZkEventHandlerFunc) OnAddLeaf(branch, leaf, value string) {
	f.AddFunc(branch, leaf, value)
}

func (f ZkEventHandlerFunc) OnUpdateLeaf(branch, leaf, oldvalue, newvalue string) {
	f.UpdateFunc(branch, leaf, oldvalue, newvalue)
}

func (f ZkEventHandlerFunc) OnDeleteLeaf(branch, leaf, value string) {
	f.DeleteFunc(branch, leaf, value)
}

type ZkWatcherEventHandler interface {
	OnAddLeaf(branch, leaf, value string)
	OnUpdateLeaf(branch, leaf, oldvalue, newvalue string)
	OnDeleteLeaf(branch, leaf, value string)
}

func NewZkWatcher(zkAddrs []string, baseBranch string, callBacker ZkWatcherEventHandler) (*ZkWatcher, error) {
	cli := zkclient.NewZkClient(zkAddrs)
	if err := cli.ConnectEx(3 * time.Second); nil != err {
		return nil, fmt.Errorf("new zk client failed. err: %v", err)
	}
	return &ZkWatcher{
		zkCli:      cli,
		BaseBranch: baseBranch,
		CallBacker: callBacker,
		resourceBox: &Box{
			Resource: make(map[string]map[string]string),
		},
	}, nil
}

type Box struct {
	Lock     sync.Mutex
	Resource map[string]map[string]string
}

type ZkWatcher struct {
	zkCli       *zkclient.ZkClient
	BaseBranch  string
	CallBacker  ZkWatcherEventHandler
	resourceBox *Box
}

func (z *ZkWatcher) Run() error {

	if len(z.BaseBranch) == 0 {
		return errors.New("base branch can not be null")
	}

	if nil == z.CallBacker {
		return errors.New("event call backer can not be nil")
	}

	if err := z.TryWatchBranch(z.BaseBranch); nil != err {
		return err
	}

	return nil
}

func (z *ZkWatcher) TryWatchBranch(zkBranch string) error {
	if true == z.IsExist(zkBranch) {
		return nil
	}
	// a new branch, add this branch first.
	if zkBranch != z.BaseBranch {
		z.AddNewBranch(zkBranch)
	}

	blog.Infof("start to watch new branch %s", zkBranch)

	go func() {
		for {
			childrens, eventChan, err := z.zkCli.WatchChildren(zkBranch)
			if err != nil {
				blog.Errorf("watch zk children %s failed. err: %v", zkBranch, err)
				if strings.Contains(err.Error(), "node does not exist") == true {
					z.DeleteBranch(zkBranch)
					blog.Errorf("zk node:%s does not exist. stop watch and close connection now.", zkBranch)
					return
				}
				time.Sleep(10 * time.Second)
				continue
			}
			leafMapper := make(map[string]struct{})
			for _, child := range childrens {
				if true == z.IsLeaf(child) {
					leafMapper[child] = struct{}{}
					path := fmt.Sprintf("%s/%s", zkBranch, child)
					if err := z.UpdateLeaf(z.zkCli, zkBranch, child); nil != err {
						blog.Errorf("update leaf failed. path: %v, err: %v", path, err)
						continue
					}
					continue
				}

				// this is a new branch
				newBranch := fmt.Sprintf("%s/%s", zkBranch, child)
				if err := z.TryWatchBranch(newBranch); nil != err {
					blog.Errorf("watch new branch: %s failed. err: %v", newBranch, err)
					continue
				}
			}
			// check leave child
			z.CheckRedundancyLeaf(zkBranch, leafMapper)

			// wait for another change.
			event := <-eventChan
			blog.Infof("path(%s) in zk received an event, reason :%s.", zkBranch, event.Type)
			if event.Type == zkclient.EventNodeDeleted {
				z.DeleteBranch(zkBranch)
				blog.Infof("zk node: %s is deleted, stop watch and close the connection.", zkBranch)
				return
			}
		}
	}()
	return nil
}

func (z *ZkWatcher) IsLeaf(child string) bool {
	return strings.HasPrefix(child, "_c_")
}

func (r *ZkWatcher) IsExist(key string) bool {
	r.resourceBox.Lock.Lock()
	defer r.resourceBox.Lock.Unlock()
	_, exist := r.resourceBox.Resource[key]
	return exist
}

func (z *ZkWatcher) CheckRedundancyLeaf(branch string, children map[string]struct{}) {
	z.resourceBox.Lock.Lock()
	defer z.resourceBox.Lock.Unlock()
	branchRes, exist := z.resourceBox.Resource[branch]
	if false == exist {
		blog.Warnf("check the redundancy leaf, but the branch: %s do not exist in the resource", branch)
		return
	}

	for subKey, value := range branchRes {
		if _, exist := children[subKey]; false == exist {
			// find a redundancy leaf, need to alarm.
			blog.Warnf("find a redundancy leaf: %s in branch: %s, old value: %#v", subKey, branch, value)
			delete(z.resourceBox.Resource[branch], subKey)
			z.CallBacker.OnDeleteLeaf(branch, subKey, value)
		}
	}
}

func (z *ZkWatcher) Set(key, subkey, value string) {
	z.resourceBox.Lock.Lock()
	defer z.resourceBox.Lock.Unlock()
	if _, exist := z.resourceBox.Resource[key]; false == exist {
		z.resourceBox.Resource[key] = make(map[string]string)
	}
	old, exist := z.resourceBox.Resource[key][subkey]
	if false == exist {
		// a new key
		z.resourceBox.Resource[key][subkey] = value
		blog.Infof("add new key: %s, value: %s", fmt.Sprintf("%s/%s", key, subkey), value)
		z.CallBacker.OnAddLeaf(key, subkey, value)
		return
	}

	if 0 == strings.Compare(old, value) {
		return
	}
	// need to update this key.
	z.resourceBox.Resource[key][subkey] = value
	blog.Infof("update changed key:%s, old: %s, new: %s", fmt.Sprintf("%s/%s", key, subkey), old, value)
	z.CallBacker.OnUpdateLeaf(key, subkey, old, value)
	return
}

func (z *ZkWatcher) AddNewBranch(branch string) {
	z.resourceBox.Lock.Lock()
	defer z.resourceBox.Lock.Unlock()
	if _, exist := z.resourceBox.Resource[branch]; false == exist {
		z.resourceBox.Resource[branch] = make(map[string]string)
		return
	}
	return
}

func (z *ZkWatcher) DeleteBranch(branch string) {
	z.resourceBox.Lock.Lock()
	leafs, exist := z.resourceBox.Resource[branch]
	if false == exist {
		return
	}
	delete(z.resourceBox.Resource, branch)
	z.resourceBox.Lock.Unlock()

	blog.Warnf("delete branch: %s.", branch)

	for leaf, srvinfo := range leafs {
		z.CallBacker.OnDeleteLeaf(branch, leaf, srvinfo)
	}
}

func (z *ZkWatcher) UpdateLeaf(zkCli *zkclient.ZkClient, fatherPath, child string) error {
	path := fmt.Sprintf("%s/%s", fatherPath, child)
	value, err := zkCli.Get(path)
	if nil != err {
		return fmt.Errorf("get path: %s value failed. err: %v", path, err)
	}
	z.Set(fatherPath, child, string(value))
	return nil
}
