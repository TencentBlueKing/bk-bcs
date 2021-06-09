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
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/google/uuid"
)

const (
	// defaultPingInterval is default ping action interval.
	defaultPingInterval = 3 * time.Second

	// defaultSyncMasterStateInterval is default master state synchronize interval.
	defaultSyncMasterStateInterval = time.Second
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

	// service register flag.
	isRegistered bool

	// service instance master state.
	isMaster bool

	// etcd client.
	etcdCli *clientv3.Client

	// etcd leaseid
	leaseid clientv3.LeaseID

	// etcd kv pair for service instance.
	key   string
	value string

	// Addr is service address for json marshal.
	Addr string `josn:"Addr"`

	// Metadata is service metadata for json marshal.
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

// IsMaster return if the service instance is the master node.
func (svi *Service) IsMaster() (bool, error) {
	if !svi.isRegistered {
		return false, errors.New("service has not yet registered")
	}
	return svi.isMaster, nil
}

func (svi *Service) syncMasterState() error {
	// TODO synchronize state base on watch.

	// get current instance version info.
	resp, err := svi.etcdCli.Get(context.Background(), svi.key, clientv3.WithPrefix(), clientv3.WithSerializable())
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		return errors.New("empty service instance create revision")
	}
	createRevision := resp.Kvs[0].CreateRevision

	// get first service instance version info.
	opts := clientv3.WithFirstCreate()
	opts = append(opts, clientv3.WithSerializable())
	resp, err = svi.etcdCli.Get(context.Background(), key(DEFAULTSCHEMA, svi.name, ""), opts...)
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		return errors.New("empty first service instance create revision")
	}
	firstCreateRevision := resp.Kvs[0].CreateRevision

	svi.isMaster = (firstCreateRevision == createRevision)

	return nil
}

func (svi *Service) synchronizing() {
	for {
		if svi.isStop {
			break
		}
		time.Sleep(defaultSyncMasterStateInterval)

		if err := svi.syncMasterState(); err != nil {
			log.Printf("sync master state %+v", err)
		}
	}
}

func (svi *Service) grantAndKeepAlive() error {
	resp, err := svi.etcdCli.Grant(context.Background(), svi.ttl)
	if err != nil {
		return err
	}
	svi.leaseid = resp.ID

	if _, err = svi.etcdCli.Put(context.Background(), svi.key, svi.value, clientv3.WithLease(svi.leaseid)); err != nil {
		return err
	}
	if _, err = svi.etcdCli.KeepAlive(context.Background(), svi.leaseid); err != nil {
		return err
	}
	return nil
}

func (svi *Service) ping() error {
	_, err := svi.etcdCli.Put(context.Background(), svi.key, svi.value, clientv3.WithLease(svi.leaseid))
	if err == nil {
		log.Printf("service register ping success!")
		return nil
	}
	log.Printf("service register ping(grant new lease now), %+v", err)

	if err := svi.grantAndKeepAlive(); err != nil {
		return fmt.Errorf("grant new lease, %+v", err)
	}
	log.Printf("service register grant new lease and ping success!")

	return nil
}

func (svi *Service) pinging() {
	for {
		if svi.isStop {
			break
		}
		time.Sleep(defaultPingInterval)

		if err := svi.ping(); err != nil {
			log.Printf("ping and grant, %+v", err)
		}
	}
}

func (svi *Service) register() error {
	bytes, err := json.Marshal(svi)
	if err != nil {
		return err
	}
	svi.key = key(DEFAULTSCHEMA, svi.name, svi.id)
	svi.value = string(bytes)

	if err := svi.grantAndKeepAlive(); err != nil {
		return err
	}

	// keep pinging.
	go svi.pinging()

	// keep synchronizing include master state.
	go svi.synchronizing()

	// register success.
	svi.isRegistered = true

	// wait to revoke.
	<-svi.stopCh
	svi.etcdCli.Revoke(context.Background(), svi.leaseid)

	svi.isStop = true
	svi.isRegistered = false
	log.Println("service register stop keepalive now!")

	return nil
}

func (svi *Service) genServiceID() (string, error) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return uuid.String(), nil
}

// Register registers service to target etcd cluster.
func (svi *Service) Register(cfg clientv3.Config) error {
	// gen new service id.
	serviceID, err := svi.genServiceID()
	if err != nil {
		return err
	}
	svi.id = serviceID

	// create etcd client.
	cli, err := clientv3.New(cfg)
	if err != nil {
		return err
	}
	defer cli.Close()
	svi.etcdCli = cli

	// register service.
	return svi.register()
}

// UnRegister unregisters service.
func (svi *Service) UnRegister() {
	select {
	case svi.stopCh <- struct{}{}:
	case <-time.After(time.Second):
	}
}
