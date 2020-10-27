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

package zookeeper

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/storage"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	operatorTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "netservice",
		Subsystem: "storage",
		Name:      "operator_total",
		Help:      "The total number of operator to storage.",
	}, []string{"operator", "status"})
	operatorLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "netservice",
		Subsystem: "storage",
		Name:      "operator_latency_seconds",
		Help:      "BCS netservice storage operation latency statistic.",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"operator", "status"})
)

const (
	statusSuccess = "SUCCESS"
	statusFailure = "FAILURE"
)

// init prometheus metrics for zookeeper
func init() {
	prometheus.MustRegister(operatorTotal)
	prometheus.MustRegister(operatorLatency)
}

func reportMetrics(operator, status string, started time.Time) {
	operatorTotal.WithLabelValues(operator, status).Inc()
	operatorLatency.WithLabelValues(operator, status).Observe(time.Since(started).Seconds())
}

//ZkClient interface to define zk operation
//interface is only use for dependency injection
type ZkClient interface {
	Get(path string) ([]byte, *zkclient.Stat, error)
	GetW(path string) ([]byte, *zkclient.Stat, <-chan zkclient.Event, error)
	Children(path string) ([]string, *zkclient.Stat, error)
	ChildrenW(path string) ([]string, *zkclient.Stat, <-chan zkclient.Event, error)
	Exists(path string) (bool, error)
	ExistsW(path string) (bool, *zkclient.Stat, <-chan zkclient.Event, error)
	Set(path string, data []byte, version int32) error
	Create(path string, data []byte) (string, error)
	CreateEphAndSeq(path string, data []byte) (string, error)
	CreateEphAndSeqEx(path string, data []byte) (string, error)
	Delete(path string, version int32) error
	Close()
}

//wrapperClient wrapper for zk client in bcs-common.zkclient
type wrapperClient struct {
	client *zkclient.ZkClient
}

func (wrapper *wrapperClient) Close() {
	wrapper.client.Close()
}

func (wrapper *wrapperClient) Get(path string) ([]byte, *zkclient.Stat, error) {
	return wrapper.client.GetEx(path)
}

func (wrapper *wrapperClient) GetW(path string) ([]byte, *zkclient.Stat, <-chan zkclient.Event, error) {
	return wrapper.client.GetW(path)
}

func (wrapper *wrapperClient) Children(path string) ([]string, *zkclient.Stat, error) {
	return wrapper.client.GetChildrenEx(path)
}

func (wrapper *wrapperClient) ChildrenW(path string) ([]string, *zkclient.Stat, <-chan zkclient.Event, error) {
	return wrapper.client.ChildrenW(path)
}

func (wrapper *wrapperClient) Exists(path string) (bool, error) {
	return wrapper.client.Exist(path)
}

func (wrapper *wrapperClient) ExistsW(path string) (bool, *zkclient.Stat, <-chan zkclient.Event, error) {
	return wrapper.client.ExistW(path)
}

func (wrapper *wrapperClient) Set(path string, data []byte, version int32) error {
	return wrapper.client.Set(path, string(data), version)
}

func (wrapper *wrapperClient) Create(path string, data []byte) (string, error) {
	return path, wrapper.client.Update(path, string(data))
}

func (wrapper *wrapperClient) CreateEphAndSeq(path string, data []byte) (string, error) {
	return path, wrapper.client.CreateEphAndSeq(path, data)
}

func (wrapper *wrapperClient) CreateEphAndSeqEx(path string, data []byte) (string, error) {
	return wrapper.client.CreateEphAndSeqEx(path, data)
}

func (wrapper *wrapperClient) Delete(path string, version int32) error {
	return wrapper.client.Del(path, version)
}

//ConLocker wrapper lock for release connection
type ConLocker struct {
	path string           //path for lock
	lock *zkclient.ZkLock //bcs common lock
}

//Lock try to Lock
func (cl *ConLocker) Lock() error {
	started := time.Now()
	err := cl.lock.LockEx(cl.path, time.Second*3)
	if err != nil {
		reportMetrics("lock", statusFailure, started)
	} else {
		reportMetrics("lock", statusSuccess, started)
	}
	return err
}

//Unlock release lock and connection
func (cl *ConLocker) Unlock() error {
	started := time.Now()
	if err := cl.lock.UnLock(); err != nil {
		reportMetrics("unlock", statusFailure, started)
	} else {
		reportMetrics("unlock", statusSuccess, started)
	}
	return nil
}

//NewStorage create etcd storage
func NewStorage(hosts string) storage.Storage {
	//create zookeeper connection
	host := strings.Split(hosts, ",")
	blog.Info("Storage create zookeeper connection with %s", hosts)
	// conn, _, conErr := zk.Connect(host, time.Second*5)
	// if conErr != nil {
	// 	blog.Error("Storage create zookeeper connection failed: %v", conErr)
	// 	return nil
	// }
	bcsClient := zkclient.NewZkClient(host)
	if conErr := bcsClient.ConnectEx(time.Second * 5); conErr != nil {
		blog.Errorf("Storage create zookeeper connection failed: %v", conErr)
		return nil
	}

	blog.Infof("Storage connect to zookeeper %s success", hosts)
	s := &zkStorage{
		zkHost: host,
		zkClient: &wrapperClient{
			client: bcsClient,
		},
	}
	return s
}

//eStorage storage data in etcd
type zkStorage struct {
	zkHost   []string //zookeeper host info, for reconnection
	zkClient ZkClient //zookeeper client for operation
}

func (zks *zkStorage) Stop() {
	zks.zkClient.Close()
}

func (zks *zkStorage) GetLocker(key string) (storage.Locker, error) {
	started := time.Now()
	defer func() {
		reportMetrics("getLocker", statusSuccess, started)
	}()
	blog.Infof("zkStorage create %s locker", key)
	bcsLock := zkclient.NewZkLock(zks.zkHost)
	wrap := &ConLocker{
		path: key,
		lock: bcsLock,
	}
	return wrap, nil
}

//Register register self node
func (zks *zkStorage) Register(path string, data []byte) error {
	started := time.Now()
	_, err := zks.zkClient.CreateEphAndSeq(path, data)
	if err != nil {
		blog.Errorf("zkStorage register %s failed, %v", path, err)
		reportMetrics("createEphAndSeq", statusFailure, started)
		return err
	}
	reportMetrics("createEphAndSeq", statusSuccess, started)
	return nil
}

//RegisterAndWatch register and watch self node
func (zks *zkStorage) RegisterAndWatch(path string, data []byte) error {
	newPath, err := zks.zkClient.CreateEphAndSeqEx(path, data)
	if err != nil {
		blog.Errorf("fail to register server node(%s), err %s", path, err.Error())
		return fmt.Errorf("fail to register server node(%s), err %s", path, err.Error())
	}

	go func() {
		tick := time.NewTicker(20 * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				started := time.Now()
				isExisted, err := zks.zkClient.Exists(newPath)
				if err != nil {
					reportMetrics("exist", statusFailure, started)
					blog.Warnf("zkClient.exists failed, wait next tick, err %s", err.Error())
				} else {
					reportMetrics("exist", statusFailure, started)
				}
				if !isExisted {
					started = time.Now()
					blog.Warnf("server node(%s) does not exist, try to create new server node", path)
					newPath, err = zks.zkClient.CreateEphAndSeqEx(path, data)
					if err != nil {
						blog.Warnf("failed to create node(%s), wait next tick", path)
						reportMetrics("createEphAndSeq", statusFailure, started)
					} else {
						blog.Warnf("create server node(%s) successfully", newPath)
						reportMetrics("createEphAndSeq", statusSuccess, started)
					}
				}
			}
		}

	}()

	fmt.Printf("finish register server node(%s) and watch it\n", path)
	return nil
}

//Add add data
func (zks *zkStorage) Add(key string, value []byte) error {
	started := time.Now()
	_, err := zks.zkClient.Create(key, value)
	if err != nil {
		blog.Errorf("zkStorage add %s with value %s err, %v", key, string(value), err)
		reportMetrics("create", statusFailure, started)
		return err
	}
	reportMetrics("create", statusFailure, started)
	return nil
}

//Delete delete node by key
func (zks *zkStorage) Delete(key string) ([]byte, error) {
	//done(DeveloperJim): get data before delete
	started := time.Now()
	data, err := zks.Get(key)
	if err != nil {
		reportMetrics("get", statusFailure, started)
		return []byte(""), err
	}
	reportMetrics("get", statusSuccess, started)
	started = time.Now()
	err = zks.zkClient.Delete(key, -1)
	if err != nil {
		reportMetrics("delete", statusFailure, started)
		return data, err
	}
	reportMetrics("delete", statusSuccess, started)
	return data, nil
}

//Update update node by value
func (zks *zkStorage) Update(key string, value []byte) error {
	started := time.Now()
	err := zks.zkClient.Set(key, value, -1)
	if err != nil {
		reportMetrics("set", statusFailure, started)
	} else {
		reportMetrics("set", statusSuccess, started)
	}
	return err
}

//Get get data of path
func (zks *zkStorage) Get(key string) ([]byte, error) {
	started := time.Now()
	data, stat, err := zks.zkClient.Get(key)
	if err != nil {
		reportMetrics("get", statusFailure, started)
		return nil, err
	} else if stat == nil {
		reportMetrics("get", statusFailure, started)
		return nil, fmt.Errorf("lost %s stat", key)
	}
	reportMetrics("get", statusSuccess, started)
	return data, nil
}

//List all children nodes
func (zks *zkStorage) List(key string) ([]string, error) {
	started := time.Now()
	list, stat, err := zks.zkClient.Children(key)
	if err != nil {
		reportMetrics("children", statusFailure, started)
		return nil, err
	}
	if stat == nil {
		reportMetrics("children", statusFailure, started)
		return nil, fmt.Errorf("path %s status lost", key)
	}
	reportMetrics("children", statusSuccess, started)
	return list, nil
}

//Exist check path exist
func (zks *zkStorage) Exist(key string) (bool, error) {
	started := time.Now()
	e, err := zks.zkClient.Exists(key)
	if err != nil {
		reportMetrics("exist", statusFailure, started)
		return false, err
	}
	reportMetrics("exist", statusFailure, started)
	return e, nil
}
