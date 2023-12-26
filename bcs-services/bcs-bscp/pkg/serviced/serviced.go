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
 */

// Package serviced NOTES
package serviced

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	etcd3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// State defines the service's state related operations.
type State interface {
	// IsMaster test if this service instance is
	// master or not.
	IsMaster() bool
	// DisableMasterSlave disable/enable this service instance's master-slave check.
	// if disabled, treat this service as a slave instead of checking if it is master from service discovery.
	DisableMasterSlave(disable bool)
	// Healthz etcd health check.
	Healthz() error
}

// Service defines all the service and discovery
// related operations.
type Service interface {
	State
	// Register the service
	Register() error
	// Deregister the service
	Deregister() error
}

// Discover defines service discovery related operations.
type Discover interface {
	LBRoundRobin() grpc.DialOption
}

// ServiceDiscover defines all the service and discovery
// related operations.
type ServiceDiscover interface {
	Service
	Discover
}

// NewService create a service instance.
func NewService(cfg etcd3.Config, opt ServiceOption) (Service, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	cli, err := etcd3.New(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "init etcd client")
	}

	httpClient := &http.Client{Transport: &http.Transport{TLSClientConfig: cfg.TLS}}
	ctx, cancel := context.WithCancel(context.Background())
	s := &serviced{
		cli:        cli,
		cfg:        cfg,
		svcOpt:     opt,
		ctx:        ctx,
		cancel:     cancel,
		httpClient: httpClient,
	}

	// keep synchronizing current node's master state.
	s.syncMasterState()
	return s, nil
}

// NewServiceD create a service and discovery instance.
func NewServiceD(cfg etcd3.Config, opt ServiceOption) (ServiceDiscover, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	cli, err := etcd3.New(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "init etcd client")
	}

	httpClient := &http.Client{Transport: &http.Transport{TLSClientConfig: cfg.TLS}}
	ctx, cancel := context.WithCancel(context.Background())
	s := &serviced{
		cli:        cli,
		cfg:        cfg,
		svcOpt:     opt,
		ctx:        ctx,
		cancel:     cancel,
		httpClient: httpClient,
	}

	resolver.Register(newEtcdBuilder(cli))
	// keep synchronizing current node's master state.
	s.syncMasterState()
	return s, nil
}

// NewDiscovery create a service discovery instance.
func NewDiscovery(cfg etcd3.Config) (Discover, error) {
	cli, err := etcd3.New(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "init etcd client")
	}

	httpClient := &http.Client{Transport: &http.Transport{TLSClientConfig: cfg.TLS}}
	resolver.Register(newEtcdBuilder(cli))
	return &serviced{
		cli:        cli,
		cfg:        cfg,
		httpClient: httpClient,
	}, nil
}

type serviced struct {
	cli    *etcd3.Client
	cfg    etcd3.Config
	svcOpt ServiceOption

	// isRegisteredFlag service register flag.
	isRegisteredFlag  bool
	isRegisteredRWMux sync.RWMutex

	// leaseID is grant lease's id that used to put kv.
	leaseID      etcd3.LeaseID
	leaseIDRWMux sync.RWMutex

	// isMasterFlag service instance master state.
	isMasterFlag  bool
	isMasterRwMux sync.RWMutex

	// disableMasterSlaveFlag defines if the service instance's master-slave check is disabled and treated as slave.
	disableMasterSlaveFlag bool

	// watchChan is watch etcd service path's watch channel.
	watchChan etcd3.WatchChan

	ctx        context.Context
	cancel     context.CancelFunc
	httpClient *http.Client
}

// LBRoundRobin returns a load balance based on all the
// service's instance.
func (s *serviced) LBRoundRobin() grpc.DialOption {
	return grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name))
}

// Register the service
func (s *serviced) Register() error {
	if s.isRegister() {
		return errors.New("only one is allowed to register for the current service")
	}

	// get service key and value.
	key := key(ServiceDiscoveryName(s.svcOpt.Name), s.svcOpt.Uid)
	addr := resolver.Address{
		Addr:       net.JoinHostPort(s.svcOpt.IP, strconv.Itoa(int(s.svcOpt.Port))),
		ServerName: string(s.svcOpt.Name),
	}
	bytes, err := json.Marshal(addr)
	if err != nil {
		return err
	}
	value := string(bytes)

	// grant lease, and put kv with lease.
	lease := etcd3.NewLease(s.cli)
	leaseResp, err := lease.Grant(s.ctx, defaultGrantLeaseTTL)
	if err != nil {
		logs.Errorf("grant lease failed, err: %v", err)
		return err
	}
	_, err = s.cli.Put(s.ctx, key, value, etcd3.WithLease(leaseResp.ID))
	if err != nil {
		logs.Errorf("put kv with lease failed, key: %s, value: %s, err: %v", key, value, err)
		return err
	}
	s.updateLeaseID(leaseResp.ID)
	s.updateRegisterFlag(true)

	// start to keep alive lease.
	s.keepAlive(key, value)

	return nil
}

func (s *serviced) keepAlive(key string, value string) {
	go func() {
		lease := etcd3.NewLease(s.cli)
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				curLeaseID := s.getLeaseID()
				// if the current lease is 0, you need to lease the lease and use this put kv (bind lease).
				// if the lease is not 0, the put has been completed and the lease needs to be renewed.
				if curLeaseID == 0 {
					leaseResp, err := lease.Grant(s.ctx, defaultGrantLeaseTTL)
					if err != nil {
						logs.Errorf("grant lease failed, key: %s, err: %v", key, err)
						time.Sleep(defaultErrSleepTime)
						continue
					}
					_, err = s.cli.Put(s.ctx, key, value, etcd3.WithLease(leaseResp.ID))
					if err != nil {
						logs.Errorf("put kv failed, key: %s, lease: %d, err: %v", key, leaseResp.ID, err)
						time.Sleep(defaultErrSleepTime)
						continue
					}

					s.updateRegisterFlag(true)
					s.updateLeaseID(leaseResp.ID)
				} else {
					// before keep alive, need to judge service key if exist.
					// if not exist, need to re-register.
					resp, err := s.cli.Get(s.ctx, key)
					if err != nil {
						logs.Errorf("get key failed, lease: %d, err: %v", curLeaseID, err)
						s.keepAliveFailed()
						continue
					}
					if len(resp.Kvs) == 0 {
						logs.Warnf("current service key [%s, %s] is not exist, need to re-register", key, value)
						s.keepAliveFailed()
						continue
					}

					if _, err := lease.KeepAliveOnce(s.ctx, curLeaseID); err != nil {
						logs.Errorf("keep alive lease failed, lease: %d, err: %v", curLeaseID, err)
						s.keepAliveFailed()
						continue
					}
				}
				time.Sleep(defaultKeepAliveInterval)
			}
		}
	}()
}

// keepAliveFailed keep alive lease failed, need to exec action.
func (s *serviced) keepAliveFailed() {
	s.updateRegisterFlag(false)
	s.updateLeaseID(0)
	time.Sleep(defaultErrSleepTime)
}

// Deregister the service
func (s *serviced) Deregister() error {
	s.cancel()

	if _, err := s.cli.Delete(context.Background(), key(ServiceDiscoveryName(s.svcOpt.Name),
		s.svcOpt.Uid)); err != nil {
		return err
	}

	s.updateRegisterFlag(false)
	return nil
}

// IsMaster test if this service instance is
// master or not.
func (s *serviced) IsMaster() bool {
	s.isMasterRwMux.RLock()
	defer s.isMasterRwMux.RUnlock()

	if s.disableMasterSlaveFlag {
		logs.Infof("master-slave is disabled, returns this service instance master state as slave")
		return false
	}
	return s.isMasterFlag
}

// DisableMasterSlave disable/enable this service instance's master-slave check.
// if disabled, treat this service as a slave instead of checking if it is master from service discovery.
func (s *serviced) DisableMasterSlave(disable bool) {
	s.isMasterRwMux.RLock()
	s.disableMasterSlaveFlag = disable
	s.isMasterRwMux.RUnlock()

	logs.Infof("master-slave disabled status: %v", disable)
}

// HealthInfo is etcd health info, e.g. '{"health":"true"}'.
type HealthInfo struct {
	// Health is state flag, it's string not boolean.
	Health string `json:"health"`
}

// Healthz checks the etcd health state.
func (s *serviced) Healthz() error {
	if len(s.cfg.Endpoints) == 0 {
		return errors.New("has no etcd endpoints")
	}

	scheme := "http"
	if s.cfg.TLS != nil {
		scheme = "https"
	}

	for _, endpoint := range s.cfg.Endpoints {
		resp, err := s.httpClient.Get(fmt.Sprintf("%s://%s/health", scheme, endpoint))
		if err != nil {
			return fmt.Errorf("get etcd health failed, err: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("response status: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read etcd healthz body failed, err: %v", err)
		}

		info := &HealthInfo{}
		if err := json.Unmarshal(body, info); err != nil {
			return fmt.Errorf("unmarshal etcd healthz info failed, err: %v", err)
		}

		if info.Health != "true" {
			return fmt.Errorf("endpoint %s etcd not healthy", endpoint)
		}
	}

	return nil
}

// syncMasterState determine whether the current node is the primary node.
func (s *serviced) syncMasterState() {
	svrPath := ServiceDiscoveryName(s.svcOpt.Name)
	svrKey := key(svrPath, s.svcOpt.Uid)

	// watch service register path change event. if receive event, need to sync master state.
	go func() {
		s.watchChan = s.cli.Watch(context.Background(), svrPath, etcd3.WithPrefix(), etcd3.WithPrevKV())

		for {
			resp, ok := <-s.watchChan
			// if the watchChan send is abnormally closed, you also need to finally
			// determine whether it is the master node.
			isMaster, err := s.isMaster(svrPath, svrKey)
			if err != nil {
				if logs.V(2) {
					logs.Errorf("sync service: %s master state failed, err: %v", s.svcOpt.Name, err)
				}
				time.Sleep(defaultErrSleepTime)
				continue
			}
			s.updateMasterFlag(isMaster)

			// if the abnormal pipe is closed, you need to retry watch
			if !ok || resp.Err() != nil {
				s.watchChan = s.cli.Watch(context.Background(), svrPath, etcd3.WithPrefix(),
					etcd3.WithPrevKV())
			}
		}
	}()

	// the bottom of the plan, sync master state regularly.
	go func() {
		for {
			isMaster, err := s.isMaster(svrPath, svrKey)
			if err != nil {
				if logs.V(2) {
					logs.Errorf("sync service: %s master state failed, err: %v", s.svcOpt.Name, err)
				}
				time.Sleep(defaultErrSleepTime)
				continue
			}
			s.updateMasterFlag(isMaster)

			time.Sleep(defaultSyncMasterInterval)
		}
	}()
}

// isMaster judge current service is master node.
func (s *serviced) isMaster(srvPath, srvKey string) (bool, error) {
	// get current instance version info.
	resp, err := s.cli.Get(context.Background(), srvKey, etcd3.WithPrefix(), etcd3.WithSerializable())
	if err != nil {
		return false, err
	}
	if len(resp.Kvs) == 0 {
		return false, errors.New("current service not register, key: " + srvKey)
	}
	cr := resp.Kvs[0].CreateRevision

	// get first service instance version info.
	opts := etcd3.WithFirstCreate()
	opts = append(opts, etcd3.WithSerializable())
	resp, err = s.cli.Get(context.Background(), srvPath, opts...)
	if err != nil {
		return false, err
	}
	if len(resp.Kvs) == 0 {
		return false, errors.New("current service not register, service path: " + srvPath)
	}
	firstCR := resp.Kvs[0].CreateRevision

	logs.V(6).Infof("current service(%s) master state: %v", srvKey, cr == firstCR)
	return cr == firstCR, nil
}

// updateMasterFlag update isMasterFlag by rw mux.
func (s *serviced) updateMasterFlag(isMaster bool) {
	s.isMasterRwMux.Lock()
	s.isMasterFlag = isMaster
	s.isMasterRwMux.Unlock()
}

// updateRegisterFlag update isMasterFlag by rw mux.
func (s *serviced) updateRegisterFlag(isRegister bool) {
	s.isRegisteredRWMux.Lock()
	s.isRegisteredFlag = isRegister
	s.isRegisteredRWMux.Unlock()
}

// updateLeaseID update leaseID by rw mux.
func (s *serviced) updateLeaseID(id etcd3.LeaseID) {
	s.leaseIDRWMux.Lock()
	s.leaseID = id
	s.leaseIDRWMux.Unlock()
}

// isRegister return is register flag by rw mux.
func (s *serviced) isRegister() bool {
	s.isRegisteredRWMux.RLock()
	defer s.isRegisteredRWMux.RUnlock()
	return s.isRegisteredFlag
}

// getLeaseID return leaseID by rw mux.
func (s *serviced) getLeaseID() etcd3.LeaseID {
	s.leaseIDRWMux.RLock()
	defer s.leaseIDRWMux.RUnlock()
	return s.leaseID
}
