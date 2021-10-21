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

package netservice

import (
	"encoding/json"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcstypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/storage"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	defaultNetSvcPath    = "/bcs/services/netservice"
	defaultHostInfoPath  = "/bcs/services/netservice/hosts"
	defaultPoolInfoPath  = "/bcs/services/netservice/pools"
	defaultLockerPath    = "/bcs/services/netservice/lock"
	defaultDiscoveryPath = "/bcs/services/endpoints/netservice"
)

var (
	logicTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "netservice",
		Subsystem: "logic",
		Name:      "operator_total",
		Help:      "The total number of logic operation.",
	}, []string{"operator", "status"})
	logicLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "netservice",
		Subsystem: "logic",
		Name:      "operator_latency_seconds",
		Help:      "BCS netservice logic operation latency statistic.",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"operator", "status"})
)

const (
	stateSuccess         = "SUCCESS"
	stateNonExistFailure = "NONEXISTFailure"
	stateJSONFailure     = "JSONFailure"
	stateLogicFailure    = "LOGICFailure"
	stateStorageFailure  = "STORAGEFailure"
)

// init all prometheus metrics initialization
func init() {
	prometheus.MustRegister(logicTotal)
	prometheus.MustRegister(logicLatency)
}

func reportMetrics(logic, status string, started time.Time) {
	logicTotal.WithLabelValues(logic, status).Inc()
	logicLatency.WithLabelValues(logic, status).Observe(time.Since(started).Seconds())
}

//NewNetService create netservice logic
func NewNetService(addr string, port, metricPort int, st storage.Storage) *NetService {
	srv := &NetService{
		addr:       addr,
		port:       port,
		metricPort: metricPort,
		store:      st,
	}
	//service init
	if _, err := st.Get(defaultNetSvcPath); err != nil {
		//create node
		st.Add(defaultNetSvcPath, []byte("netservice"))
	}
	if _, err := st.Get(defaultHostInfoPath); err != nil {
		//create node
		st.Add(defaultHostInfoPath, []byte("hosts"))
	}
	if _, err := st.Get(defaultPoolInfoPath); err != nil {
		//create node
		st.Add(defaultPoolInfoPath, []byte("pools"))
	}
	if _, err := st.Get(defaultLockerPath); err != nil {
		//create node
		st.Add(defaultLockerPath, []byte("lock"))
	}
	if _, err := st.Get(defaultDiscoveryPath); err != nil {
		//create node
		st.Add(defaultDiscoveryPath, []byte("server"))
	}
	if err := srv.createSelfNode(); err != nil {
		blog.Errorf("NetServer create self node err, %v", err)
		return nil
	}
	return srv
}

//NetNode node data for
type NetNode struct {
	Addr string `json:"addr"`
	Port int    `json:"port"`
}

//NetService service for all logic, this NetService
//add/delete/update/list all data base on key/value
//todo(DeveloperJim): if we need to store data in SQL system, NetService needs refactor.
type NetService struct {
	addr       string          //local listen addr
	port       int             //local listen port
	metricPort int             //metric port
	store      storage.Storage //storage for host, pool info
}

//createSelfNode create self node in storage for service discovery
func (srv *NetService) createSelfNode() error {
	hostname, _ := os.Hostname()
	node := &bcstypes.NetServiceInfo{
		ServerInfo: bcstypes.ServerInfo{
			IP:         srv.addr,
			Port:       uint(srv.port),
			MetricPort: uint(srv.metricPort),
			Pid:        os.Getpid(),
			HostName:   hostname,
			Scheme:     "https",
			Version:    version.GetVersion(),
		},
	}
	data, _ := json.Marshal(node)
	self := srv.addr + ":" + strconv.Itoa(srv.port)
	key := filepath.Join(defaultDiscoveryPath, self)
	//clean self node first and then create new one
	if exist, _ := srv.store.Exist(key); exist {
		blog.Warn("server node %s exist, clean first", key)
		srv.store.Delete(key)
	}
	err := srv.store.RegisterAndWatch(key, data)
	if err != nil {
		blog.Errorf("create NetService %s temporary node %s err, %v", srv.addr, key, err)
		return err
	}
	blog.Info("NetService %s create temporary node %s success", srv.addr, key)

	return nil
}
