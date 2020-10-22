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

package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/google/uuid"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/micro/go-micro/v2/util/backoff"
)

//NewEtcdRegistry create etcd registry instance
func NewEtcdRegistry(option *Options) Registry {
	//setting default options
	if option.TTL == time.Duration(0) {
		option.TTL = time.Duration(time.Second * 40)
	}
	if option.Interval == time.Duration(0) {
		option.Interval = time.Duration(time.Second * 30)
	}
	//create etcd registry
	r := etcd.NewRegistry(
		registry.Addrs(option.RegistryAddr...),
		registry.TLSConfig(option.Config),
	)
	//creat local service
	option.id = uuid.New().String()
	svc := &registry.Service{
		Name:    option.Name,
		Version: option.Version,
		Nodes: []*registry.Node{
			&registry.Node{
				Id:       fmt.Sprintf("%s-%s", option.Name, option.id),
				Address:  option.RegAddr,
				Metadata: option.Meta,
			},
		},
	}
	ctx, stop := context.WithCancel(context.Background())
	e := &etcdRegister{
		option:       option,
		ctx:          ctx,
		stop:         stop,
		etcdregistry: r,
		localService: svc,
		registered:   false,
	}
	return e
}

// etcd registry implementation
type etcdRegister struct {
	option       *Options
	ctx          context.Context
	stop         context.CancelFunc
	etcdregistry registry.Registry
	localService *registry.Service
	registered   bool
}

//Register service information to registry
// register do not block, if module want to
// clean registe information, call Deregister
func (e *etcdRegister) Register() error {
	if e.registered {
		return fmt.Errorf("already registered")
	}
	if err := e.innerRegister(); err != nil {
		return err
	}
	e.registered = true
	//start backgroud goroutine for interval keep alive
	// because we setting ttl for register
	go func() {
		tick := time.NewTicker(e.option.Interval)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				//ready to keepAlive registered node information
				if err := e.innerRegister(); err != nil {
					blog.Errorf("register %s information %++v failed, %s", e.localService.Name, e.localService, err.Error())
					blog.Warnf("try register next tick...")
				}
			case <-e.ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (e *etcdRegister) innerRegister() error {
	var rerr error
	for i := 0; i < 3; i++ {
		if err := e.etcdregistry.Register(
			e.localService,
			registry.RegisterTTL(e.option.TTL),
		); err != nil {
			//try again until max failed
			rerr = err
			roption := e.etcdregistry.Options()
			blog.Errorf("etcd registry register err, %s, options: %+v\n", err.Error(), roption)
			time.Sleep(backoff.Do(i + 1))
			continue
		}
		//register success, clean error
		rerr = nil
		break
	}
	return rerr
}

//Deregister clean service information from registry
func (e *etcdRegister) Deregister() error {
	//stop backgroud keepalive goroutine
	e.stop()
	//clean registered node information
	if err := e.etcdregistry.Deregister(e.localService); err != nil {
		blog.Warnf("Deregister %s information %++v failed, %s", e.localService.Name, e.localService, err.Error())
	}
	e.registered = false
	return nil
}
