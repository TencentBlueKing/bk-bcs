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

package grpclb

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/google/uuid"
)

// Service is service info for etcd discovery register.
type Service struct {
	// service name.
	name string

	// service id.
	id string

	// service keepalive ttl.
	ttl int64

	// stop channel.
	stopCh chan struct{}

	// serice stop flag.
	isStop bool

	// etcd leaseid
	leaseid clientv3.LeaseID

	// etcd KV
	Addr     string `josn:"Addr"`
	Metadata string `json:"Metadata"`
}

// NewService creates a new service context.
func NewService(name, addr, metadata string, ttl int64) *Service {
	service := &Service{
		name:     name,
		ttl:      ttl,
		stopCh:   make(chan struct{}),
		Addr:     addr,
		Metadata: metadata,
	}
	return service
}

func (svi *Service) grantAndKeepAlive(cli *clientv3.Client, key, value string) error {
	resp, err := cli.Grant(context.Background(), svi.ttl)
	if err != nil {
		return err
	}
	svi.leaseid = resp.ID

	if _, err = cli.Put(context.Background(), key, value, clientv3.WithLease(svi.leaseid)); err != nil {
		return err
	}

	if _, err := cli.KeepAlive(context.Background(), svi.leaseid); err != nil {
		return err
	}
	return nil
}

func (svi *Service) ping(cli *clientv3.Client, key, value string) error {
	_, err := cli.Put(context.Background(), key, value, clientv3.WithLease(svi.leaseid))
	if err == nil {
		log.Printf("service register ping success!")
		return nil
	}
	log.Printf("service register ping(grant new lease now), %+v", err)

	if err := svi.grantAndKeepAlive(cli, key, value); err != nil {
		log.Printf("service register ping and grant new lease, %+v", err)
		return err
	}
	log.Printf("service register grant new lease and ping success!")
	return nil
}

func (svi *Service) register(cfg clientv3.Config) error {
	cli, err := clientv3.New(cfg)
	if err != nil {
		return err
	}
	defer cli.Close()

	bytes, err := json.Marshal(svi)
	if err != nil {
		return err
	}

	key := key(DEFAULTSCHEMA, svi.name, svi.id)
	value := string(bytes)

	if err := svi.grantAndKeepAlive(cli, key, value); err != nil {
		return err
	}

	go func() {
		for {
			if svi.isStop {
				break
			}
			time.Sleep(time.Second)

			svi.ping(cli, key, value)
		}
		log.Print("service register stop now!")
	}()

	// wait to exit
	<-svi.stopCh
	cli.Revoke(context.Background(), svi.leaseid)

	log.Println("service register stop keepalive now!")
	return nil
}

func (svi *Service) genServiceid() (string, error) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return uuid.String(), nil
}

// Register registers service to target etcd cluster.
func (svi *Service) Register(cfg clientv3.Config) error {
	serviceid, err := svi.genServiceid()
	if err != nil {
		return err
	}
	svi.id = serviceid

	return svi.register(cfg)
}

// UnRegister unregisters service.
func (svi *Service) UnRegister() {
	select {
	case svi.stopCh <- struct{}{}:
		svi.isStop = true

	case <-time.After(time.Second):
	}
}
