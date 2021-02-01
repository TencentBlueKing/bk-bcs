/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package rssche

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/coreos/etcd/clientv3"
)

// Operation defines the op type.
type Operation uint8

const (
	// Put resource.
	Put Operation = iota

	// Delete resource.
	Delete
)

// Update is update event data
type Update struct {
	// Op operation type.
	Op Operation

	// Metadata resource metadata, such as loadinfo of server instance.
	Metadata string `json:"Metadata"`
}

// Scheduler is a resource scheduler.
type Scheduler struct {
	target   string
	resource Resource
	isStop   bool
	etcdCli  *clientv3.Client
	wch      clientv3.WatchChan
}

// NewScheduler create a new Scheduler for resource with name target.
func NewScheduler(target string, resource Resource) *Scheduler {
	return &Scheduler{
		target:   fmt.Sprintf("%s/%s", DEFAULTSCHEMA, target),
		resource: resource,
	}
}

// Init initialize a new Scheduler.
func (s *Scheduler) Init(cfg clientv3.Config) error {
	cli, err := clientv3.New(cfg)
	if err != nil {
		return err
	}
	s.etcdCli = cli
	return nil
}

// watch watchs the updates from etcd cluster.
func (s *Scheduler) watch() error {
	wr, ok := <-s.wch
	if !ok {
		return errors.New("watch channel closed")
	}
	if err := wr.Err(); err != nil {
		return err
	}

	// updates.
	updates := make([]*Update, 0, len(wr.Events))

	for _, e := range wr.Events {
		var update Update
		var err error

		switch e.Type {
		case clientv3.EventTypePut:
			err = json.Unmarshal(e.Kv.Value, &update)
			update.Op = Put

		case clientv3.EventTypeDelete:
			err = json.Unmarshal(e.PrevKv.Value, &update)
			update.Op = Delete

		default:
			continue
		}

		if err == nil {
			updates = append(updates, &update)
		}
	}

	s.resource.Update(updates)

	return nil
}

// Start starts a Scheduler. gets resource list and watchs the updates until stop.
func (s *Scheduler) Start() error {
	resp, err := s.etcdCli.Get(context.Background(), s.target, clientv3.WithPrefix(), clientv3.WithSerializable())
	if err != nil {
		return err
	}

	// resource.
	updates := make([]*Update, 0, len(resp.Kvs))

	for _, kv := range resp.Kvs {
		var update Update
		if err := json.Unmarshal(kv.Value, &update); err != nil {
			continue
		}
		updates = append(updates, &update)
	}
	s.resource.Update(updates)

	opts := []clientv3.OpOption{
		clientv3.WithRev(resp.Header.Revision + 1),
		clientv3.WithPrefix(),
		clientv3.WithPrevKV(),
	}
	s.wch = s.etcdCli.Watch(context.Background(), s.target, opts...)

	go func() {
		for {
			if s.isStop {
				log.Print("stop the scheduler now!")
				break
			}

			if err := s.watch(); err != nil {
				log.Printf("scheduler watch %+v", err)
			}
		}
	}()

	return nil
}

// Stop stops the scheduler.
func (s *Scheduler) Stop() {
	s.isStop = true
	s.etcdCli.Close()
}
