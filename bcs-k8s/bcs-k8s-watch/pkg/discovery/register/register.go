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

package register

import (
	"fmt"
	"time"

	regd "github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// NodeRegister is the low-level register for any type of nodes
type NodeRegister struct {
	ZkServerAddresses string
	StateChan         chan ServiceState

	zkService               *regd.RegDiscover
	zkServiceConnectTimeout time.Duration

	baseKey string
	node    Node
}

// ServiceState is event for StartDiscover
type ServiceState struct {
	Nodes        int
	NodePayloads []string
	MyPostion    int
}

func NewNodeRegister(zkServerAddresses string, baseKey string, node Node) *NodeRegister {
	return &NodeRegister{
		ZkServerAddresses: zkServerAddresses,
		StateChan:         make(chan ServiceState, 1024),

		baseKey:                 baseKey,
		zkServiceConnectTimeout: time.Duration(5 * time.Second),
		node:                    node,
	}
}

// EnsureConn connects to zookeeper server if needed
func (r *NodeRegister) EnsureConn() error {
	if r.zkService != nil {
		return nil
	}

	if r.ZkServerAddresses == "" {
		return errors.New("must provide zookeeper server address")
	}

	r.zkService = regd.NewRegDiscoverEx(r.ZkServerAddresses, r.zkServiceConnectTimeout)
	if err := r.zkService.Start(); err != nil {
		return fmt.Errorf("new register discover instance failed: %s", err)
	}
	return nil
}

// DoRegister register current node info to ZK Servers
func (r *NodeRegister) DoRegister() error {
	if err := r.EnsureConn(); err != nil {
		return fmt.Errorf("no connection: %s", err)
	}

	if err := r.zkService.RegisterAndWatchService(r.GetRegisterKey(), r.node.Payload()); nil != err {
		return fmt.Errorf("register service failed: %v", err)
	}
	blog.Infof("registered node, node info: %s", r.node.Payload())
	return nil
}

// GetRegisterkey accepts a node and returns the key of zookeeper
func (r *NodeRegister) GetRegisterKey() string {
	return fmt.Sprintf("%s/%s", r.baseKey, r.node.PrimaryKey())
}

// StartDiscover watch for new events on current cluster node, will retry on connection errors
func (r *NodeRegister) StartDiscover(timeoutSeconds int) error {
	if err := r.EnsureConn(); err != nil {
		return fmt.Errorf("no connection: %s", err)
	}

	// Retry discover on every error
	for {
		err := r.startSingleDiscover(timeoutSeconds)
		if err == nil {
			break
		}
	}
	return nil
}

// startSingleDiscover starts a single discover process, will return error if error returns from eventChan
func (r *NodeRegister) startSingleDiscover(timeoutSeconds int) error {
	var timeoutChan <-chan time.Time
	if timeoutSeconds > 0 {
		timeoutChan = time.After(time.Duration(timeoutSeconds) * time.Second)
	} else {
		timeoutChan = make(chan time.Time)
	}

	// Discover other children under the same clusterID
	eventChan, err := r.zkService.DiscoverService(r.baseKey)
	if err != nil {
		return fmt.Errorf("start running discover service failed: %v", err)
	}

	// CHANGE: 2018-06-15 comment, let the RegisterAndWatchService do the re-connection
	//debounceDoRegister := debounceCallable(3*time.Second, r.DoRegister)
	blog.Infof("register watch path: %s", r.baseKey)
Outer:
	for {
		select {
		case <-timeoutChan:
			blog.Info("waiting for timeout seconds exceeds, quit discovering")
			break Outer
		case e := <-eventChan:
			// If error happens, the original eventChan will have no more events because the underlay loopDiscover
			// function call will return, to continue discover process, we will should restart a new discover process
			if e.Err != nil {
				blog.Error("received error from event chan: %s, will register self and call StartDiscover again.", e.Err)
				r.SendEmptyEvent()

				// CHANGE: 2018-06-15 comment, let the RegisterAndWatchService do the re-connection
				//debounceDoRegister()

				time.Sleep(5 * time.Second)
				return fmt.Errorf("eventChan error: %s", e.Err)
			}

			// Re-regiter to zookeeper if found any node data was written by ifself
			myPosition := -1
			for i, nodeStr := range e.Server {
				if r.node.OwnsPayload([]byte(nodeStr)) {
					myPosition = i
					break
				}
			}

			// This operation may block if no one is consuming things from stateChan
			// and length of events has exceeds capacity, so we will try to clean up some
			// space if channel is half full.
			// TODO: Why not start a goroutine for data send to make sure it never blocks?
			for len(r.StateChan) > cap(r.StateChan)/2 {
				select {
				case <-r.StateChan:
				default:
				}
			}

			// NOTE: this is the data for leader election, without this, election will fail
			// This operation should never blocks
			r.StateChan <- ServiceState{
				Nodes:        len(e.Server),
				NodePayloads: e.Server,
				MyPostion:    myPosition,
			}

			// CHANGE: 2018-06-15
			// NOTE: change the RegisterService to RegisterAndWatchService, it's a goroutine,
			//       and at that time, register to zk maybe not finish yet, the position=-1
			//       result: the node reigster twice to zk

			// with this change, we should let the RegisterAndWatchService care about error
			//                   this function should running for leader election only

			// Auto-reregister if it can not found myself on nodes.
			// WARN: If there are more than 1 sequential events with no self can be found, it might cause
			// DoRegister being called for more than 1 times, this may results that more than 2 nodes belonged to
			// this serviceNode could be found on zookeeper server.
			// To avoid this, we will use a technical call "debouce" to compact sequential calls in to a single call
			// which is delayed by max 3 seconds.
			//if myPosition == -1 {
			//	blog.Warnf("couldn't find myself [key: %s] from zk, start register again.", r.node.PrimaryKey())
			//	debounceDoRegister()
			//}
		}
	}
	return nil
}

func (r *NodeRegister) SendEmptyEvent() {
	r.StateChan <- ServiceState{
		Nodes:        0,
		NodePayloads: nil,
		MyPostion:    -1,
	}
}

func (r *NodeRegister) Stop() error {
	err := r.zkService.Stop()
	r.SendEmptyEvent()
	return err
}
