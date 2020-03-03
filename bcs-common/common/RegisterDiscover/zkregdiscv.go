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

package RegisterDiscover

import (
	"bk-bcs/bcs-common/common/blog"
	//"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/zkclient"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
)

//ZkRegDiscv do register and discover by zookeeper
type ZkRegDiscv struct {
	zkcli          *zkclient.ZkClient
	cancel         context.CancelFunc
	rootCxt        context.Context
	sessionTimeOut time.Duration
}

//NewZkRegDiscv create a object of ZkRegDiscv
func NewZkRegDiscv(serv string, timeOut time.Duration) *ZkRegDiscv {
	zkservs := strings.Split(serv, ",")
	return &ZkRegDiscv{
		zkcli:          zkclient.NewZkClient(zkservs),
		sessionTimeOut: timeOut,
	}
}

//Start used to run register and discover server
func (zkRD *ZkRegDiscv) Start() error {
	//connect zookeeper
	if err := zkRD.zkcli.ConnectEx(zkRD.sessionTimeOut); err != nil {

		return fmt.Errorf("fail to connect zookeeper. err:%s", err.Error())
	}

	// create root context
	zkRD.rootCxt, zkRD.cancel = context.WithCancel(context.Background())

	return nil
}

//Stop used to stop register and discover server
func (zkRD *ZkRegDiscv) Stop() error {
	//close the connection of zookeeper
	zkRD.zkcli.Close()

	//cancel
	zkRD.cancel()

	return nil
}

//Register create ephemeral node for the service
func (zkRD *ZkRegDiscv) Register(path string, data []byte) error {
	//blog.Info("register server. path(%s), data(%s)", path, string(data))
	return zkRD.zkcli.CreateEphAndSeq(path, data)
}

//RegisterAndWatch create ephemeral node for the service and watch it. if it exit, register again
func (zkRD *ZkRegDiscv) RegisterAndWatch(path string, data []byte) error {
	//blog.Info("register server and watch it. path(%s), data(%s)", path, string(data))
	go func() {
		newPath, err := zkRD.zkcli.CreateEphAndSeqEx(path, data)
		if err != nil {
			blog.V(3).Infof("fail to register server node(%s). err:%s", path, err.Error())
			time.Sleep(5 * time.Second)
			zkRD.RegisterAndWatch(path, data)
			return
		}

		watchCtx, _ := context.WithCancel(zkRD.rootCxt)

		_, _, watchEvn, err := zkRD.zkcli.ExistW(newPath)
		if err != nil {
			blog.V(3).Infof("fail to watch register node(%s), err:%s", newPath, err.Error())
			time.Sleep(5 * time.Second)
			zkRD.zkcli.Del(newPath, -1)
			zkRD.RegisterAndWatch(path, data)
			return
		}

		select {
		case <-watchCtx.Done():
			blog.V(3).Infof("watch register node(%s) done", path)
			return
		case <-watchEvn:
			blog.V(3).Infof("watch register node(%s) exist changed, event(%v)", path, watchEvn)
			zkRD.RegisterAndWatch(path, data)
		}
	}()

	blog.V(3).Infof("finish register server node(%s) and watch it", path)
	return nil
}

//Discover watch the children
func (zkRD *ZkRegDiscv) Discover(path string) (<-chan *DiscoverEvent, error) {
	blog.V(3).Infof("begin to discover by watch children of path(%s)", path)
	discvCtx, _ := context.WithCancel(zkRD.rootCxt)

	env := make(chan *DiscoverEvent, 1)

	go zkRD.loopDiscover(path, discvCtx, env)

	return env, nil
}

func (zkRD *ZkRegDiscv) loopDiscover(path string, discvCtx context.Context, env chan *DiscoverEvent) {
	for {
		discvEnv := &DiscoverEvent{
			Err: nil,
			Key: path,
		}

		servNodes, watchEnv, err := zkRD.zkcli.WatchChildren(path)
		if err != nil {
			blog.V(3).Infof("fail to watch children for path(%s), err:%s", path, err.Error())
			if zkclient.ErrNoNode == err {
				blog.V(3).Infof("children node(%s) is not exist, will watch after 5s", path)
				time.Sleep(5 * time.Second)
				continue
			}

			discvEnv.Err = err
			env <- discvEnv
			return
		}

		discvEnv.Nodes = append(discvEnv.Nodes, servNodes...)
		//sort server node
		servNodes = zkRD.sortNode(servNodes)

		//get server info
		for _, node := range servNodes {
			servPath := path + "/" + node
			servInfo, err := zkRD.zkcli.Get(servPath)
			if err != nil {
				blog.V(3).Infof("fail to get server info from zookeeper by path(%s), err:%s", servPath, err.Error())
				continue
			}

			discvEnv.Server = append(discvEnv.Server, servInfo)
		}

		//write into discoverEvent channel
		env <- discvEnv

		select {
		case <-discvCtx.Done():
			blog.V(3).Infof("discover path(%s) done", path)
			return
		case <-watchEnv:
			blog.V(3).Infof("watch found the children of path(%s) change", path)
		}
	}
}

// DiscoverNodes discover by specified path
func (zkRD *ZkRegDiscv) DiscoverNodes(path string) (*DiscoverEvent, error) {
	discvEnv := &DiscoverEvent{
		Err: nil,
		Key: path,
	}

	servNodes, _, err := zkRD.zkcli.WatchChildren(path)
	if err != nil {
		discvEnv.Err = err
		return discvEnv, nil
	}

	discvEnv.Nodes = append(discvEnv.Nodes, servNodes...)
	//sort server node
	servNodes = zkRD.sortNode(servNodes)

	//get server info
	for _, node := range servNodes {
		servPath := path + "/" + node
		servInfo, err := zkRD.zkcli.Get(servPath)
		if err != nil {
			blog.V(3).Infof("fail to get server info from zookeeper by path(%s), err:%s", servPath, err.Error())
			continue
		}

		discvEnv.Server = append(discvEnv.Server, servInfo)
	}

	return discvEnv, nil
}

func (zkRD *ZkRegDiscv) sortNode(nodes []string) []string {
	var sortPart []int
	mapSortNode := make(map[int]string)
	for _, chNode := range nodes {
		if len(chNode) <= 10 {
			blog.V(3).Infof("node(%s) is less then 10, there is not the seq number", chNode)
			continue
		}

		p, err := strconv.Atoi(chNode[len(chNode)-10 : len(chNode)])
		if err != nil {
			blog.V(3).Infof("fail to conv string to seq number for node(%s), err:%s", chNode, err.Error())
			continue
		}

		sortPart = append(sortPart, p)
		mapSortNode[p] = chNode
	}

	sort.Ints(sortPart)

	var children []string
	for _, part := range sortPart {
		children = append(children, mapSortNode[part])
	}

	return children
}

// DiscoverNodes discover by specified path
func (zkRD *ZkRegDiscv) DiscoverNodesV2(path string) (*DiscoverEvent, error) {
	discvEnv := &DiscoverEvent{
		Err: nil,
		Key: path,
	}

	servNodes, _, err := zkRD.zkcli.GetChildrenEx(path)
	if err != nil {
		discvEnv.Err = err
		return discvEnv, nil
	}

	discvEnv.Nodes = append(discvEnv.Nodes, servNodes...)
	//sort server node
	servNodes = zkRD.sortNode(servNodes)

	//get server info
	for _, node := range servNodes {
		servPath := path + "/" + node
		servInfo, err := zkRD.zkcli.Get(servPath)
		if err != nil {
			blog.V(3).Infof("fail to get server info from zookeeper by path(%s), err:%s", servPath, err.Error())
			continue
		}

		discvEnv.Server = append(discvEnv.Server, servInfo)
	}

	return discvEnv, nil
}
