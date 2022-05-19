/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/service"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/cmd"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/google/uuid"
	"github.com/micro/go-micro/v2/sync"
	"github.com/micro/go-micro/v2/sync/etcd"
)

func Test_becomeLeader(t *testing.T) {
	opts := &cmd.DataManagerOptions{}
	opts.Etcd.EtcdEndpoints = "127.0.0.1:2379"
	// ready to campaign
	leaderChan := make(chan sync.Leader, 1)
	rootCtx := service.SetupSignalContext()
	ctx := context.Background()
	etcdSync := etcd.NewSync(
		sync.Nodes(strings.Split(opts.Etcd.EtcdEndpoints, ",")...),
		sync.Prefix(common.ServiceDomain),
		sync.WithContext(rootCtx),
	)

	id := uuid.New().String()
	fmt.Printf("id:%s\n", id)
	fmt.Printf("Node %s waiting to become a leader...", id)
	leader, err := etcdSync.Leader(id, sync.LeaderContext(rootCtx))
	if err != nil {
		fmt.Printf("create etcd leader election failed: %s, service exit.\n", err.Error())
		os.Exit(-1)
	}
	leaderChan <- leader
	fmt.Println("I'm leader!")
	// return

	// go func() {
	//	time.Sleep(140 * time.Second)
	//	fmt.Printf("%v: enter ctx done \n", time.Now())
	//	cancel()
	// }()
	// go func() {
	//	time.Sleep(2 * time.Minute)
	//	fmt.Printf("%v: enter new leader \n", time.Now())
	//	newLeader, _ := etcd.NewSync().Leader("test")
	//	leaderChan <- newLeader
	// }()
	for {
		select {
		case <-leaderChan:
			fmt.Printf("become producer\n")
			time.Sleep(5 * time.Second)
			err := workAsTestProducer(ctx, leader, opts)
			if err != nil {
				fmt.Printf("producer err:%v", err)
			}
			fmt.Printf("get out producer\n")
		case <-ctx.Done():
			return
		default:
			fmt.Printf("become consumer\n")
			time.Sleep(5 * time.Second)
			err := workAsTestConsumer(ctx, leaderChan)
			if err != nil {
				fmt.Printf("consumer err:%v", err)
			}
			fmt.Printf("get out consumer\n")
		}
	}
}

func workAsTestConsumer(ctx context.Context, leader chan sync.Leader) error {
	fmt.Println(111)
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("ctx done, leader resign")
			return fmt.Errorf("ctx done")
		default:
			if len(leader) != 0 {
				return nil
			}
			fmt.Println(len(leader))
			fmt.Printf("%v:still consumer\n", time.Now())
			time.Sleep(5 * time.Second)
		}
	}
}

func workAsTestProducer(ctx context.Context, leader sync.Leader, opt *cmd.DataManagerOptions) error {
	fmt.Println(222)
	fmt.Printf("enter lost func\n")
	// lost := make(chan bool)
	lost := leader.Status()
	// go func() {
	//	time.Sleep(30 * time.Second)
	//	lost <- true
	// }()
	fmt.Printf("next\n")
	for {
		select {
		case <-lost:
			fmt.Printf("lost leader\n")
			becomeLeader(ctx, opt)
			return nil
		case <-ctx.Done():
			err := leader.Resign()
			if err != nil {
				fmt.Printf("ctx done, leader resign error :%v", err)
			} else {
				fmt.Printf("ctx done, leader resign")
			}
			return fmt.Errorf("ctx done")
		default:
			fmt.Printf("%v: still producer\n", time.Now())
			time.Sleep(5 * time.Second)
		}
	}
}
