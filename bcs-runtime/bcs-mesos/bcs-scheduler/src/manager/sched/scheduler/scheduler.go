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

package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	rd "github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	alarm "github.com/Tencent/bk-bcs/bcs-common/common/bcs-health/api"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/metric"
	typesplugin "github.com/Tencent/bk-bcs/bcs-common/common/plugin"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	commtype "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	master "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos/master"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/sched"
	types "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-scheduler/src/manager/remote/alertmanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-scheduler/src/manager/sched/client"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-scheduler/src/manager/sched/misc"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-scheduler/src/manager/sched/offer"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-scheduler/src/manager/sched/operator"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-scheduler/src/manager/sched/servermetric"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-scheduler/src/manager/sched/task"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-scheduler/src/manager/store"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-scheduler/src/pluginManager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-scheduler/src/util"

	"github.com/andygrunwald/megos"
	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
)

// MAX_DATA_UPDATE_INTERVAL Interval for update task, taskgroup, application in ZK
const MAX_DATA_UPDATE_INTERVAL = 180

// DATA_CHECK_INTERVAL Interval for checking ZK data
const DATA_CHECK_INTERVAL = 1200

// MESOS_HEARTBEAT_TIMEOUT HeartBeat timeout between scheduler and mesos master
const MESOS_HEARTBEAT_TIMEOUT = 120

// MAX_STAGING_UPDATE_INTERVAL max interval for staging
const MAX_STAGING_UPDATE_INTERVAL = 180

const (
	SchedulerRoleMaster = "master"
	SchedulerRoleSlave  = "slave"
)

// Scheduler represents a Mesos scheduler
type Scheduler struct {
	master    string
	framework *mesos.FrameworkInfo
	store     store.Store

	client         *client.Client
	operatorClient *client.Client

	// Scheduler Listen IP
	IP string
	// Scheduler Listen Port
	Port int

	// Current Schedulers in the cluster
	Schedulers []*commtype.SchedulerServInfo
	// Current Mesos Masters in the cluster
	Memsoses []*commtype.MesosServInfo

	mesosHeartBeatTime int64
	mesosMasterID      string
	currMesosMaster    string
	currMesosResp      *http.Response
	// Current Role: master, slave, none
	Role        string
	serviceLock sync.Mutex

	slaveLock sync.RWMutex

	agentSchedInofLock sync.RWMutex

	lostSlave map[string]int64

	// Cluster ID from mesos master
	ClusterId  string
	config     util.Scheduler
	clientCert *commtype.CertConfig

	// BCS Cluster ID
	BcsClusterId string

	eventManager *bcsEventManager

	oprMgr      *operator.OperatorMgr
	dataChecker *DataCheckMgr

	// Service Manager
	ServiceMgr *ServiceMgr

	offerPool offer.OfferPool

	pluginManager *pluginManager.PluginManager

	// alert interface
	alertManager alertmanager.AlertManageInterface
	//stop daemonset signal
	stopDaemonset chan struct{}

	// queue for scheduler transaction
	transactionQueue workqueue.RateLimitingInterface
}

// NewScheduler returns a pointer to new Scheduler
func NewScheduler(config util.Scheduler, store store.Store, alert alertmanager.AlertManageInterface) *Scheduler {
	s := &Scheduler{
		config:       config,
		store:        store,
		alertManager: alert,
		eventManager: newBcsEventManager(config),
		lostSlave:    make(map[string]int64),
	}

	para := &offer.OfferPara{Sched: s, Store: store}
	s.offerPool = offer.NewOfferPool(para)
	s.clientCert = &commtype.CertConfig{
		CertFile:   config.ClientCertFile,
		KeyFile:    config.ClientKeyFile,
		CAFile:     config.ClientCAFile,
		CertPasswd: static.ClientCertPwd,
	}

	//init executor info
	task.InitExecutorInfo(s.config.ContainerExecutor, s.config.ProcessExecutor, s.config.CniDir, s.config.NetImage)

	s.eventManager.Run()

	s.store.InitLockPool()
	s.store.InitDeploymentLockPool()
	s.store.InitCmdLockPool()

	// TODO, the follow statements are only used for passing test,
	// should resovled to make sure test pass
	s.client = client.New("foobar", "make test pass")
	s.operatorClient = client.New("foobar", "make test pass")

	var err error

	if s.config.Plugins != "" {
		blog.Infof("start init plugin manager")
		plugins := strings.Split(s.config.Plugins, ",")

		s.pluginManager, err = pluginManager.NewPluginManager(plugins, s.config.PluginDir)
		if err != nil {
			blog.Errorf("NewPluginManager error %s", err.Error())
		}
	}

	return s
}

func (s *Scheduler) runPrometheusMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	addr := s.IP + ":" + strconv.Itoa(int(s.config.MetricPort))
	blog.Infof("scheduler listen metrics %s", addr)
	go http.ListenAndServe(addr, nil)
}

func (s *Scheduler) runMetric() {

	blog.Infof("run metric: port(%d)", s.config.MetricPort)

	conf := metric.Config{
		RunMode:     metric.Master_Slave_Mode,
		ModuleName:  commtype.BCS_MODULE_SCHEDULER,
		MetricPort:  s.config.MetricPort,
		IP:          s.IP,
		ClusterID:   s.BcsClusterId,
		SvrCaFile:   s.config.ServerCAFile,
		SvrCertFile: s.config.ServerCertFile,
		SvrKeyFile:  s.config.ServerKeyFile,
		SvrKeyPwd:   static.ServerCertPwd,
	}

	healthFunc := func() metric.HealthMeta {
		ok, msg := servermetric.IsHealthy()
		role := servermetric.GetRole()
		return metric.HealthMeta{
			CurrentRole: role,
			IsHealthy:   ok,
			Message:     msg,
		}
	}

	if err := metric.NewMetricController(
		conf,
		healthFunc); nil != err {
		blog.Errorf("run metric fail: %s", err.Error())
	}

	blog.Infof("run metric ok")
}

// for mesos master and scheduler changing
func (s *Scheduler) lockService() {
	s.serviceLock.Lock()
}
func (s *Scheduler) unlockService() {
	s.serviceLock.Unlock()
}

// Start start starts the scheduler and subscribes to event stream
func (s *Scheduler) Start() error {
	if s.config.Cluster == "" {
		blog.Errorf("scheduler cluster unknown")
		return fmt.Errorf("scheduler cluster unknown")
	}
	blog.Info("scheduler run for cluster %s", s.config.Cluster)
	s.BcsClusterId = s.config.Cluster

	s.ServiceMgr = NewServiceMgr(s)
	if s.ServiceMgr == nil {
		return fmt.Errorf("new serviceMgr(%s:/blueking) error", s.config.ZK)
	}
	go s.ServiceMgr.Worker()

	// get Host and Port
	splitID := strings.Split(s.config.Address, ":")
	if len(splitID) < 2 {
		return fmt.Errorf("listen address %s format error", s.config.Address)
	}
	s.IP = splitID[0]
	port, err := strconv.Atoi(splitID[1])
	if err != nil {
		return fmt.Errorf("can not get port from %s", s.config.Address)
	}
	s.Port = port
	blog.Info("scheduler run address(%s:%d)", s.IP, s.Port)

	s.runPrometheusMetrics()
	//s.runMetric()

	s.Role = "unknown"
	s.currMesosMaster = ""
	go s.discvMesos()

	time.Sleep(3 * time.Second)
	go s.regDiscove()

	// register to BCS
	go s.registerBCS()

	blog.Info("scheduler Start end")

	return nil
}

// create frameworkInfo on initial start
// OR load preexisting frameworkId make mesos believe it's a RESTART of framework
func createOrLoadFrameworkInfo(config util.Scheduler, store store.Store) (*mesos.FrameworkInfo, error) {
	fw := &mesos.FrameworkInfo{
		//User:            proto.String(config.MesosFrameworkUser),
		User:            proto.String(""),
		Name:            proto.String("bcs"),
		Hostname:        proto.String(config.Hostname),
		FailoverTimeout: proto.Float64(60 * 60 * 24 * 7),
		Checkpoint:      proto.Bool(true),
		Capabilities: []*mesos.FrameworkInfo_Capability{
			{
				Type: mesos.FrameworkInfo_Capability_PARTITION_AWARE.Enum(),
			},
		},
	}

	frameworkID, err := store.FetchFrameworkID()
	if err != nil {
		if strings.ToLower(err.Error()) != "zk: node does not exist" && !strings.Contains(err.Error(), "not found") {
			blog.Error("Fetch framework id failed: %s", err.Error())
			return nil, err
		}

		blog.Warn("Fetch framework id failed: %s, will create a new framework", err.Error())
		frameworkID = ""
	}
	blog.Info("fetch frameworkId %s from DB", frameworkID)
	if frameworkID != "" {
		fw.Id = &mesos.FrameworkID{
			Value: proto.String(frameworkID),
		}
	}

	return fw, nil
}

func (s *Scheduler) discvMesos() {
	blog.Info("scheduler begin to discover mesos master from (%s)", s.config.MesosMasterZK)
	servermetric.SetMesosMaster("")

	MesosDiscv := s.config.MesosMasterZK
	regDiscv := rd.NewRegDiscover(MesosDiscv)
	if regDiscv == nil {
		blog.Error("new mesos discover(%s) return nil", MesosDiscv)
		time.Sleep(3 * time.Second)
		go s.discvMesos()
		return
	}
	blog.Info("new mesos discover(%s) succ", MesosDiscv)

	err := regDiscv.Start()
	if err != nil {
		blog.Error("mesos discover start error(%s)", err.Error())
		time.Sleep(3 * time.Second)
		go s.discvMesos()
		return
	}
	blog.Info("mesos discover start succ")
	defer regDiscv.Stop()

	discvPath := "/mesos"
	discvMesosEvent, err := regDiscv.DiscoverService(discvPath)
	if err != nil {
		blog.Error("watch mesos master under (%s: %s) error(%s)", MesosDiscv, discvPath, err.Error())
		time.Sleep(3 * time.Second)
		go s.discvMesos()
		return
	}
	blog.Info("watch mesos master under (%s: %s)", MesosDiscv, discvPath)

	tick := time.NewTicker(120 * time.Second)
	defer tick.Stop()
	for {
		select {
		//case <-rdCxt.Done():
		//blog.Warn("scheduler worker to exit")
		//regDiscv.Stop()
		//return	nil
		case <-tick.C:
			blog.Info("mesos discove(%s:%s), curr mesos master:%s", MesosDiscv, discvPath, s.currMesosMaster)
			// add mesos heartbeat check
			if s.Role == SchedulerRoleMaster {
				s.lockService()
				heartbeat := s.mesosHeartBeatTime
				now := time.Now().Unix()
				s.unlockService()
				if now-heartbeat > MESOS_HEARTBEAT_TIMEOUT {
					s.SendHealthMsg(alarm.WarnKind, "", "mesos heartbeat timeout, master:"+s.currMesosMaster, "", nil)
					blog.Warn("mesos master(%s) heartbeat timeout, redo discovering ", s.currMesosMaster)
					s.lockService()
					s.currMesosMaster = ""
					s.mesosMasterID = ""
					s.unlockService()
					go s.discvMesos()
					return
				}
			}
		case event := <-discvMesosEvent:
			blog.Info("discover event for mesos master")
			if event.Err != nil {
				blog.Error("get mesos discover event err:%s", event.Err.Error())
				time.Sleep(3 * time.Second)
				go s.discvMesos()
				return
			}
			currMaster := ""
			MasterID := ""
			blog.Info("get mesos master node num(%d)", len(event.Server))
			s.Memsoses = make([]*commtype.MesosServInfo, 0)
			for i, server := range event.Server {
				blog.Info("get mesos master: server[%d]: %s %s", i, event.Key, server)

				var serverInfo types.ReschedMesosInfo
				if err = json.Unmarshal([]byte(server), &serverInfo); err != nil {
					blog.Error("fail to unmarshal mesos master(%s), err:%s", string(server), err.Error())
					continue
				}

				masterInfo := new(commtype.MesosServInfo)
				masterInfo.ServerInfo.IP = serverInfo.Address.IP
				masterInfo.ServerInfo.Port = uint(serverInfo.Address.Port)
				masterInfo.ServerInfo.HostName = serverInfo.Hostname
				//masterInfo.ServerInfo.Pid = serverInfo.Pid
				masterInfo.ServerInfo.Scheme = "http"
				masterInfo.ServerInfo.Version = serverInfo.Version
				s.Memsoses = append(s.Memsoses, masterInfo)
				if i == 0 {
					currMaster = serverInfo.Address.IP + ":" + strconv.Itoa(serverInfo.Address.Port)
					MasterID = serverInfo.Id
				}
			}
			if currMaster == "" {
				blog.Error("get mesos master list empty")
				s.lockService()
				s.currMesosMaster = ""
				s.mesosMasterID = ""
				s.unlockService()
				servermetric.SetMesosMaster("")
				continue
			}

			blog.Info("get mesos master leader: %s:%s", currMaster, MasterID)
			err = s.checkMesosChange(currMaster, MasterID)
			if err != nil {
				blog.Errorf("check mesos change err: %s, to rediscover after 3 second ... ", err.Error())
				time.Sleep(3 * time.Second)
				go s.discvMesos()
				return
			}
		} // select
	} // for
}

func (s *Scheduler) checkMesosChange(currMaster, MasterID string) error {
	blog.Info("scheduler check mesos change begin")
	s.lockService()
	defer func() {
		blog.Info("scheduler check mesos change end")
		s.unlockService()
	}()

	if currMaster == s.currMesosMaster && s.mesosMasterID == MasterID {
		blog.Info("mesos master leader: %s not changed", currMaster)
		return nil
	}

	blog.Info("mesos master leader change: %s --> %s", s.currMesosMaster, currMaster)
	s.currMesosMaster = currMaster
	s.mesosMasterID = MasterID
	servermetric.SetMesosMaster(s.currMesosMaster)

	if s.Role != SchedulerRoleMaster {
		blog.Info("mesos master leader changed to %s, but scheduler's role is %s, do nothing", currMaster, s.Role)
		return nil
	}

	blog.Info("mesos master leader changed to %s, scheduler is master, do change", currMaster)

	// we should use better end way, ex. use context , to do
	if s.currMesosResp != nil {
		blog.Info("close current http ...")
		s.currMesosResp.Body.Close()
		s.currMesosResp = nil
	}
	time.Sleep(3 * time.Second)

	// start new
	state, err := stateFromMasters([]string{s.currMesosMaster})
	if err != nil {
		return fmt.Errorf("get state from mesos master(%s) err:%s ", s.currMesosMaster, err.Error())
	}

	blog.Info("get mesos master state: Leader(%s) Cluster(%s)", state.Leader, state.Cluster)
	s.master = state.Leader
	cluster := state.Cluster
	if cluster == "" {
		cluster = "Unnamed"
	}
	s.ClusterId = cluster
	s.client = client.New(state.Leader, "/api/v1/scheduler")

	s.framework, err = createOrLoadFrameworkInfo(s.config, s.store)
	if err != nil {
		return fmt.Errorf("load framworkinfo err:%s ", err.Error())
	}
	if err = s.subscribe(); err != nil {
		return fmt.Errorf("subscribe mesos master(%s) err:%s ", state.Leader, err.Error())
	}
	blog.Info("subscribe to mesos master(%s) succ", state.Leader)
	if s.dataChecker == nil {
		blog.Info("to create data checker")
		s.dataChecker, _ = CreateDataCheckMgr(s.store, s)
		go func() {
			DataCheckManage(s.dataChecker, s.config.DoRecover)
		}()
	}
	if s.dataChecker != nil {
		var msg DataCheckMsg
		msg.MsgType = "opencheck"
		s.dataChecker.SendMsg(&msg)
		blog.Info("after open data checker")
	}

	return nil
}

func (s *Scheduler) regDiscove() {

	blog.Info("scheduler to do registe and discove...")

	servermetric.SetRole(metric.UnknownRole)

	// register service
	regDiscv := rd.NewRegDiscoverEx(s.config.RegDiscvSvr, time.Second*10)
	if regDiscv == nil {
		blog.Error("new scheduler regDiscv(%s) return nil, redo after 3 second ...", s.config.RegDiscvSvr)
		time.Sleep(3 * time.Second)
		go s.regDiscove()
		return
	}
	blog.Info("new scheduler regDiscv(%s) succ", s.config.RegDiscvSvr)

	err := regDiscv.Start()
	if err != nil {
		blog.Error("scheduler regDiscv(%s) start error(%s), redo after 3 second ...", s.config.RegDiscvSvr, err.Error())
		time.Sleep(3 * time.Second)
		go s.regDiscove()
		return
	}
	blog.Info("scheduler regDiscv(%s) start succ", s.config.RegDiscvSvr)

	defer regDiscv.Stop()

	host, err := os.Hostname()
	if err != nil {
		blog.Error("mesos scheduler get hostname err: %s", err.Error())
		host = "UNKOWN"
	}

	var regInfo commtype.SchedulerServInfo
	regInfo.ServerInfo.Cluster = s.ClusterId
	regInfo.ServerInfo.Pid = os.Getpid()
	regInfo.ServerInfo.Version = version.GetVersion()
	regInfo.ServerInfo.IP = s.IP
	regInfo.ServerInfo.Port = uint(s.Port)
	regInfo.ServerInfo.MetricPort = s.config.MetricPort
	regInfo.ServerInfo.HostName = host
	regInfo.ServerInfo.Scheme = s.config.Scheme
	key := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_SCHEDULER + "/" + s.IP + ":" + strconv.Itoa(s.Port)
	data, err := json.Marshal(regInfo)
	if err != nil {
		blog.Error("scheduler regDiscv json Marshal error(%s)", err.Error())
		time.Sleep(3 * time.Second)
		go s.regDiscove()
		return
	}

	err = regDiscv.RegisterService(key, []byte(data))
	if err != nil {
		blog.Error("scheduler RegisterService(%s) error(%s), redo after 3 second ...", key, err.Error())
		time.Sleep(3 * time.Second)
		go s.regDiscove()
		return
	}

	blog.Info("scheduler RegisterService(%s:%s) succ", key, data)

	discvPath := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_SCHEDULER
	discvEvent, err := regDiscv.DiscoverService(discvPath)
	if err != nil {
		blog.Error("scheduler DiscoverService(%s:%s) error(%s), redo after 3 second ...",
			s.config.RegDiscvSvr, discvPath, err.Error())
		time.Sleep(3 * time.Second)
		go s.regDiscove()
		return
	}
	blog.Info("scheduler DiscoverService(%s:%s) succ", s.config.RegDiscvSvr, discvPath)

	tick := time.NewTicker(180 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			blog.Info("scheduler(%s:%d-%s) is running, discove(%s-%s)",
				s.IP, s.Port, s.Role, s.config.RegDiscvSvr, discvPath)
		case event := <-discvEvent:
			blog.Info("get scheduler discover event")
			if event.Err != nil {
				blog.Error("get scheduler discover event err:%s,  redo after 3 second ...", event.Err.Error())
				time.Sleep(3 * time.Second)
				go s.regDiscove()
				return
			}
			isMaster := false
			isRegstered := false
			s.Schedulers = make([]*commtype.SchedulerServInfo, 0)
			for i, server := range event.Server {
				blog.Info("discoved : server[%d]: %s %s", i, event.Key, server)
				if server == string(data) {
					blog.Info("discoved : server[%d] is myself", i)
					isRegstered = true
				}
				if i == 0 && server == string(data) {
					isMaster = true
					blog.Info("discoved : I am master")
				}

				serverInfo := new(commtype.SchedulerServInfo)
				if err = json.Unmarshal([]byte(server), serverInfo); err != nil {
					blog.Error("fail to unmarshal scheduler(%s), err:%s", string(server), err.Error())
					continue
				}
				s.Schedulers = append(s.Schedulers, serverInfo)
			}
			if isRegstered == false {
				blog.Warn("scheduler is not regestered in zk, do register after 3 second ...")
				time.Sleep(3 * time.Second)
				go s.regDiscove()
				return
			}
			if isMaster {
				err = s.checkRoleChange(SchedulerRoleMaster)
				servermetric.SetRole(metric.MasterRole)
			} else {
				err = s.checkRoleChange(SchedulerRoleSlave)
				servermetric.SetRole(metric.SlaveRole)
			}
			if err != nil {
				blog.Error("scheduler check role change err:%s", err.Error())
				time.Sleep(3 * time.Second)
				go s.regDiscove()
				return
			}
		} // end select
	} // end for
}

func (s *Scheduler) registerBCS() {

	blog.Info("BCS register run ...")

	if s.BcsClusterId == "" {
		blog.Error("no cluster information, BCS register redo after 3 second ...")
		time.Sleep(3 * time.Second)
		go s.registerBCS()
		return
	}

	// register service
	regDiscv := rd.NewRegDiscoverEx(s.config.BcsZK, time.Second*10)
	if regDiscv == nil {
		blog.Error("prepare register schedulr to BCS(%s) return nil, redo after 3 second ...", s.config.BcsZK)
		time.Sleep(3 * time.Second)
		go s.registerBCS()
		return
	}
	blog.Info("prepare register scheduler to BCS(%s) succ", s.config.BcsZK)

	err := regDiscv.Start()
	if err != nil {
		blog.Error("BCS register start error(%s), redo after 3 second ...", err.Error())
		time.Sleep(3 * time.Second)
		go s.registerBCS()
		return
	}
	blog.Info("BCS register start succ")

	defer regDiscv.Stop()

	host, err := os.Hostname()
	if err != nil {
		blog.Error("mesos scheduler get hostname err: %s", err.Error())
		host = "UNKOWN"
	}
	var regInfo commtype.SchedulerServInfo
	regInfo.ServerInfo.Cluster = s.BcsClusterId
	regInfo.ServerInfo.Pid = os.Getpid()
	regInfo.ServerInfo.Version = version.GetVersion()
	regInfo.ServerInfo.IP = s.IP
	regInfo.ServerInfo.Port = uint(s.Port)
	regInfo.ServerInfo.MetricPort = s.config.MetricPort
	regInfo.ServerInfo.HostName = host
	regInfo.ServerInfo.Scheme = s.config.Scheme

	key := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_SCHEDULER + "/" + s.BcsClusterId + "/" + s.IP
	data, err := json.Marshal(regInfo)
	if err != nil {
		blog.Error("json Marshal error(%s)", err.Error())
		return
	}
	err = regDiscv.RegisterService(key, []byte(data))
	if err != nil {
		blog.Error("BCS register(%s) error(%s), redo after 3 second ...", key, err.Error())
		time.Sleep(3 * time.Second)
		go s.registerBCS()
		return
	}
	blog.Info("BCS register(%s:%s) succ", key, data)

	discvPath := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_SCHEDULER + "/" + s.BcsClusterId
	discvEvent, err := regDiscv.DiscoverService(discvPath)
	if err != nil {
		blog.Error("BCS register discove path(%s) error(%s), redo after 3 second ...", discvPath, err.Error())
		time.Sleep(3 * time.Second)
		go s.registerBCS()
		return
	}
	blog.Info("BCS register discove path(%s) succ", discvPath)

	tick := time.NewTicker(180 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			blog.Info("BCS register: scheduler(%s:%d-%s) is running, discove path(%s-%s)",
				s.IP, s.Port, s.Role, s.config.BcsZK, discvPath)

		case event := <-discvEvent:
			blog.Info("BCS register get event")
			if event.Err != nil {
				blog.Error("BCS register get event err:%s,  redo after 3 second ...", event.Err.Error())
				time.Sleep(3 * time.Second)
				go s.registerBCS()
				return
			}

			isRegstered := false
			for i, server := range event.Server {
				blog.Info("BCS register discove : server[%d]: %s %s", i, event.Key, server)
				if server == string(data) {
					blog.Info("BCS register discove: server[%d] is myself", i)
					isRegstered = true
				}
			}

			if isRegstered == false {
				blog.Warn("scheduler is not regestered in BCS, do register after 3 second ...")
				time.Sleep(3 * time.Second)
				go s.registerBCS()
				return
			}
		} // end select
	} // end for
}

func (s *Scheduler) checkRoleChange(currRole string) error {
	blog.Info("scheduler check role change begin")
	s.lockService()
	defer func() {
		blog.Info("scheduler check role change end")
		s.unlockService()
	}()

	if currRole == s.Role {
		blog.Info("scheduler role: %s not changed", s.Role)
		return nil
	}

	blog.Info("scheduler role change: %s --> %s", s.Role, currRole)
	if currRole != SchedulerRoleMaster {
		if s.currMesosResp != nil {
			blog.Info("close current http ...")
			s.currMesosResp.Body.Close()
			s.currMesosResp = nil
		}
		if s.oprMgr != nil {
			blog.Info("close current operator manager ...")
			var msgOp operator.OperatorMsg
			msgOp.MsgType = "stop"
			s.oprMgr.SendMsg(&msgOp)
			s.oprMgr = nil
		}

		if s.ServiceMgr != nil {
			var msgOpen ServiceMgrMsg
			msgOpen.MsgType = "close"
			s.ServiceMgr.SendMsg(&msgOpen)
			blog.Info("after close service manager")
		}

		if s.dataChecker != nil {
			var msg DataCheckMsg
			msg.MsgType = "closecheck"
			s.dataChecker.SendMsg(&msg)
			blog.Info("after close data check goroutine")
		}
		//s.store.StopStoreMetrics()
		s.store.UnInitCacheMgr()
		//stop check and build daemonset
		s.stopBuildDaemonset()
		// stop transaction loop
		s.stopTransactionLoop()
		return nil
	}
	//init cache
	err := s.store.InitCacheMgr(s.config.UseCache)
	if err != nil {
		blog.Errorf("InitCacheMgr failed: %s, and exit", err.Error())
		os.Exit(1)
	}
	//sync agent pods index
	err = s.syncAgentsettingPods()
	if err != nil {
		blog.Errorf("syncAgentsettingPods failed: %s, and exit", err.Error())
		os.Exit(1)
	}
	//current role is master
	s.Role = currRole
	go s.store.StartStoreObjectMetrics()
	go s.startCheckDeployments()
	if s.ServiceMgr != nil {
		var msgOpen ServiceMgrMsg
		msgOpen.MsgType = "open"
		s.ServiceMgr.SendMsg(&msgOpen)
		blog.Info("after open service manager")
	}

	if s.currMesosMaster == "" {
		blog.Info("scheduler changed to %s, but curr mesos is nil, do nothing", s.Role)
		return nil
	}

	// start new
	state, err := stateFromMasters([]string{s.currMesosMaster})
	if err != nil {
		return fmt.Errorf("get state from mesos master(%s) err:%s ", s.currMesosMaster, err.Error())
	}

	blog.Info("get mesos master state: Leader(%s) Cluster(%s)", state.Leader, state.Cluster)
	s.master = state.Leader
	cluster := state.Cluster
	if cluster == "" {
		cluster = "Unnamed"
	}
	s.ClusterId = cluster
	s.client = client.New(state.Leader, "/api/v1/scheduler")

	s.framework, err = createOrLoadFrameworkInfo(s.config, s.store)
	if err != nil {
		return fmt.Errorf("load framworkinfo err:%s ", err.Error())
	}
	if err := s.subscribe(); err != nil {
		return fmt.Errorf("subscribe mesos master(%s) err:%s ", state.Leader, err.Error())
	}
	blog.Info("subscribe to mesos master(%s) succ", state.Leader)

	if s.oprMgr == nil {
		// create operator manager
		blog.Info("to create operator manager")
		s.operatorClient = client.New(state.Leader, "/api/v1")
		s.oprMgr, _ = operator.CreateOperatorMgr(s.store, s.operatorClient)
		go operator.OperatorManage(s.oprMgr)
		s.oprMgr.SendMsg(&operator.OperatorMsg{
			MsgType: "opencheck",
		})
	}

	if s.dataChecker == nil {
		// create data checker
		blog.Info("to create data checker")
		s.dataChecker, _ = CreateDataCheckMgr(s.store, s)
		go func() {
			DataCheckManage(s.dataChecker, s.config.DoRecover)
		}()
	}
	if s.dataChecker != nil {
		var msg DataCheckMsg
		msg.MsgType = "opencheck"
		s.dataChecker.SendMsg(&msg)
		blog.Info("after open data checker")
	}
	//start check and build daemonset
	go s.startBuildDaemonsets()
	// start transaction loop
	s.startTransactionLoop()

	return nil

}

func stateFromMasters(masters []string) (*megos.State, error) {
	masterUrls := make([]*url.URL, 0)
	for _, master := range masters {
		masterURL, _ := url.Parse(fmt.Sprintf("http://%s", master))
		blog.Info("mesos master Url: %s", masterURL)
		masterUrls = append(masterUrls, masterURL)
	}

	mesosClient := megos.NewClient(masterUrls, nil)
	return mesosClient.GetStateFromCluster()
}

// UpdateMesosAgents update mesos info
func (s *Scheduler) UpdateMesosAgents() {
	s.oprMgr.UpdateMesosAgents()
}

//for build pod index in agent
func (s *Scheduler) syncAgentsettingPods() error {
	taskg, err := s.store.ListClusterTaskgroups()
	if err != nil {
		blog.Infof("ListClusterTaskgroups failed: %s", err.Error())
		return err
	}
	//empty agentsetting pods
	settings, err := s.store.ListAgentsettings()
	if err != nil {
		blog.Errorf("ListAgentsettings failed: %s", err.Error())
		return err
	}
	for _, setting := range settings {
		setting.Pods = make([]string, 0)
		err = s.store.SaveAgentSetting(setting)
		if err != nil {
			blog.Errorf("SaveAgentSetting %s failed: %s", setting.InnerIP, err.Error())
			return err
		}
	}

	//save agentsetting pods
	for _, taskgroup := range taskg {
		nodeIP := taskgroup.GetAgentIp()
		if nodeIP == "" {
			blog.Errorf("taskgroup %s GetAgentIp failed.", taskgroup.ID)
			continue
		}

		setting, err := s.store.FetchAgentSetting(nodeIP)
		if err != nil && !errors.Is(err, store.ErrNoFound) {
			blog.Errorf("FetchAgentSetting %s failed: %s", nodeIP, err.Error())
			return err
		}
		if setting == nil {
			setting = &commtype.BcsClusterAgentSetting{
				InnerIP: nodeIP,
				Pods:    make([]string, 0),
			}
		}
		setting.Pods = append(setting.Pods, taskgroup.ID)
		err = s.store.SaveAgentSetting(setting)
		if err != nil {
			blog.Errorf("SaveAgentSetting %s failed: %s", setting.InnerIP, err.Error())
			return err
		}
	}

	return nil
}

// Stop stop whole scheduler
func (s *Scheduler) Stop() {
	blog.Info("scheduler Stop ...")
}

func (s *Scheduler) send(call *sched.Call) (*http.Response, error) {
	payload, err := proto.Marshal(call)
	if err != nil {
		return nil, err
	}
	//blog.V(3).Infof("send pkg to master: %s", string(payload))
	return s.client.Send(payload)
}

// Subscribe subscribes the scheduler to the Mesos cluster.
// It keeps the http connection opens with the Master to stream
// subsequent events.
func (s *Scheduler) subscribe() error {
	blog.Info("Subscribe with mesos master %s", s.master)
	call := &sched.Call{
		Type: sched.Call_SUBSCRIBE.Enum(),
		Subscribe: &sched.Call_Subscribe{
			FrameworkInfo: s.framework,
		},
	}
	if s.framework.Id != nil {
		call.FrameworkId = &mesos.FrameworkID{
			Value: proto.String(s.framework.Id.GetValue()),
		}
	}

	resp, err := s.send(call)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Subscribe with unexpected response status: %d", resp.StatusCode)
	}

	blog.Info("client for mesos master streamID:%s", s.client.StreamID)
	s.currMesosResp = resp
	go s.handleEvents(resp)

	return nil
}

// main loop of a scheduler module
// if error, maybe need resubsribe in scheduler master state
func (s *Scheduler) handleEvents(resp *http.Response) {
	r := misc.NewReader(resp.Body)
	dec := json.NewDecoder(r)
	for {
		event := new(sched.Event)
		if err := dec.Decode(event); err != nil {
			blog.Error("Decode mesos event failed: %s", err)
			return
		}

		switch event.GetType() {
		case sched.Event_SUBSCRIBED:
			sub := event.GetSubscribed()
			blog.Info("subscribe mesos successful with frameworkId %s", sub.FrameworkId.GetValue())
			if registered, _ := s.store.HasFrameworkID(); !registered {
				if err := s.store.SaveFrameworkID(sub.FrameworkId.GetValue()); err != nil {
					blog.Error("save frameworkId to DB failed: %s", sub.FrameworkId.GetValue(), err)
					return
				}
				blog.Info("save frameworkId %s to DB succeed", sub.FrameworkId.GetValue())
			}

			if s.framework.Id == nil {
				s.framework.Id = sub.FrameworkId
			}
			s.lockService()
			s.mesosHeartBeatTime = time.Now().Unix()
			s.unlockService()

		case sched.Event_OFFERS:
			for _, offer := range event.Offers.Offers {
				by, _ := json.Marshal(offer)
				blog.V(3).Infof("mesos report offer %s", string(by))

				cpus, mem, disk := s.OfferedResources(offer)
				blog.Infof("mesos report offer %s||%s: cpu(%f) mem(%f) disk(%f)",
					offer.GetHostname(), *(offer.Id.Value), cpus, mem, disk)
			}
			s.offerPool.AddOffers(event.Offers.Offers)

		case sched.Event_RESCIND:
			blog.Info("mesos report rescind offers event")

		case sched.Event_UPDATE:
			blog.V(3).Infof("mesos report update event")
			status := event.GetUpdate().GetStatus()
			go func() {
				s.StatusReport(status)
			}()

		case sched.Event_MESSAGE:
			message := event.GetMessage()
			blog.V(3).Infof("receive message(%s)", message.String())
			data := message.GetData()
			var bcsMsg *types.BcsMessage
			err := json.Unmarshal(data, &bcsMsg)
			if err != nil {
				blog.Error("unmarshal bcsmessage(%s) err:%s", data, err.Error())
				continue
			}
			switch *bcsMsg.Type {
			case types.Msg_Res_COMMAND_TASK:
				go s.ProcessCommandMessage(bcsMsg)
			case types.Msg_TASK_STATUS_UPDATE:
				go s.UpdateTaskStatus(message.GetAgentId().GetValue(), message.GetExecutorId().GetValue(), bcsMsg)
			default:
				blog.Error("unknown message type(%s)", *bcsMsg.Type)
			}

		case sched.Event_FAILURE:
			//blog.Warn("Received failure event")
			fail := event.GetFailure()
			if fail.ExecutorId != nil {
				blog.Info("Executor(%s) terminated with status(%d) on agent(%s)",
					fail.ExecutorId.GetValue(), fail.GetStatus(), fail.GetAgentId().GetValue())
			} else {
				if fail.GetAgentId() != nil {
					blog.Info("Agent " + fail.GetAgentId().GetValue() + " failed ")
				}
			}

		case sched.Event_ERROR:
			err := event.GetError().GetMessage()
			blog.Error("mesos report error event. err:%s", err)

		case sched.Event_HEARTBEAT:
			blog.V(3).Infof("mesos report heartbeat event")
			s.lockService()
			s.mesosHeartBeatTime = time.Now().Unix()
			s.unlockService()
		default:
			blog.Warn("unkown mesos event type(%d)", event.GetType())
		}

	}
}

// SendHealthMsg Send health message
func (s *Scheduler) SendHealthMsg(
	kind alarm.MessageKind, runAs, message string, alarmID string, convergenceSeconds *uint16) {
	if convergenceSeconds == nil {
		blog.Warn("send health message(%s): ns(%s), alarmID(%s) ", message, runAs, alarmID)
	} else {
		blog.Warn("send health message(%s): ns(%s), alarmID(%s), convergenceSeconds(%d)",
			message, runAs, alarmID, *convergenceSeconds)
	}

	currentTime := time.Now().Local()

	health := alarm.HealthInfo{
		Module:             commtype.BCS_MODULE_SCHEDULER,
		AlarmName:          "scheduler",
		Kind:               kind,
		AlarmID:            alarmID,
		ConvergenceSeconds: convergenceSeconds,

		IP:         s.IP,
		ClusterID:  s.BcsClusterId,
		Namespace:  runAs,
		Message:    message,
		Version:    version.GetVersion(),
		ReportTime: currentTime.Format("2006-01-02 15:04:05.000"),
	}

	if s.alertManager != nil {
		err := s.alertManager.CreateAlertInfoToAlertManager(&alertmanager.CreateBusinessAlertInfoReq{
			Starttime:    time.Now().Unix(),
			Generatorurl: "",
			AlarmType:    "module",
			ClusterID:    s.ClusterId,
			AlertAnnotation: &alertmanager.AlertAnnotation{
				Message: message,
				Comment: "",
			},
			ModuleAlertLabel: &alertmanager.ModuleAlertLabel{
				ModuleName: health.Module,
				ModuleIP:   s.IP,
				AlarmName:  health.AlarmName,
				AlarmLevel: string(kind),
			},
		}, time.Second*10)
		if err != nil {
			blog.Warn("CreateBusinessAlertInfo send health message(%s) failed: err[%v]", message, err)
		}
	}

	return
}

func (s *Scheduler) produceEvent(object interface{}) error {
	btype := reflect.TypeOf(object)

	if btype.Kind() != reflect.Struct {
		return fmt.Errorf("object type must be struct")
	}

	var event *commtype.BcsStorageEventIf

	switch btype.Name() {
	case reflect.TypeOf(types.Task{}).Name():
		task := object.(types.Task)
		event = s.newTaskEvent(&task)

	default:
		return fmt.Errorf("object type %s is invalid", btype.Name())
	}

	go s.eventManager.syncEvent(event)
	return nil
}

func (s *Scheduler) newTaskEvent(task *types.Task) *commtype.BcsStorageEventIf {
	event := &commtype.BcsStorageEventIf{
		ID:        task.ID,
		Env:       commtype.Event_Env_Mesos,
		Kind:      commtype.TaskEventKind,
		Component: commtype.Event_Component_Scheduler,
		Type:      task.Status,
		EventTime: task.UpdateTime,
		Describe:  task.Message,
		ClusterId: s.BcsClusterId,
		ExtraInfo: commtype.EventExtraInfo{
			Namespace: task.RunAs,
			Name:      task.AppId,
			Kind:      commtype.ApplicationExtraKind,
		},
		Data: &v1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      task.AppId,
				Namespace: task.RunAs,
			},
			InvolvedObject: v1.ObjectReference{
				Kind:      string(commtype.ApplicationExtraKind),
				Namespace: task.RunAs,
				Name:      task.AppId,
			},
			Message: task.Message,
		},
	}

	if task.Status == types.TASK_STATUS_ERROR || task.Status == types.TASK_STATUS_FAIL ||
		task.Status == types.TASK_STATUS_LOST {
		event.Level = commtype.Event_Level_Warning
	} else {
		event.Level = commtype.Event_Level_Normal
	}

	return event
}

// PushTransaction push transaction into queue
func (s *Scheduler) PushEventQueue(transaction *types.Transaction) {
	if s.transactionQueue != nil {
		nsName := k8stypes.NamespacedName{
			Namespace: transaction.Namespace,
			Name:      transaction.TransactionID,
		}
		s.transactionQueue.Forget(nsName)
		s.transactionQueue.Add(nsName)
	}
}

// DeclineResource is used to send DECLINE request to mesos to release offer. This
// is very important, otherwise resource will be taked until framework exited.
func (s *Scheduler) DeclineResource(offerId *string) (*http.Response, error) {
	call := &sched.Call{
		FrameworkId: s.framework.GetId(),
		Type:        sched.Call_DECLINE.Enum(),
		Decline: &sched.Call_Decline{
			OfferIds: []*mesos.OfferID{
				{
					Value: offerId,
				},
			},
			Filters: &mesos.Filters{
				RefuseSeconds: proto.Float64(1),
			},
		},
	}

	return s.send(call)
}

// DeclineOffers Decline offer from mesos master
func (s *Scheduler) DeclineOffers(offers []*mesos.Offer) error {
	for _, offer := range offers {
		_, err := s.DeclineResource(offer.Id.Value)
		if err != nil {
			blog.Error("fail to decline offer(%s), err:%s", *(offer.Id.Value), err.Error())
			return err
		}
	}

	return nil
}

// OfferedResources Get offered resource from mesos master
func (s *Scheduler) OfferedResources(offer *mesos.Offer) (cpus, mem, disk float64) {
	for _, res := range offer.GetResources() {
		if res.GetName() == "cpus" {
			cpus += *res.GetScalar().Value
		}
		if res.GetName() == "mem" {
			mem += *res.GetScalar().Value
		}
		if res.GetName() == "disk" {
			disk += *res.GetScalar().Value
		}
	}

	return
}

// GetHostAttributes Get agent attributes
func (s *Scheduler) GetHostAttributes(para *typesplugin.HostPluginParameter) (map[string]*typesplugin.HostAttributes, error) {
	if s.pluginManager == nil {
		return nil, fmt.Errorf("pluginManager is nil")
	}

	return s.pluginManager.GetHostAttributes(para)
}

// FetchAgentSetting Get agent setting by IP
func (s *Scheduler) FetchAgentSetting(ip string) (*commtype.BcsClusterAgentSetting, error) {
	return s.store.FetchAgentSetting(ip)
}

// FetchAgentSchedInfo Get agent schedInfo by hostname
func (s *Scheduler) FetchAgentSchedInfo(hostname string) (*types.AgentSchedInfo, error) {
	s.agentSchedInofLock.RLock()
	defer s.agentSchedInofLock.RUnlock()

	return s.store.FetchAgentSchedInfo(hostname)
}

// UpdateAgentSchedInfo Update agent schedinfo by hostname
func (s *Scheduler) UpdateAgentSchedInfo(hostname, taskGroupID string, deltaResource *types.Resource) error {
	s.agentSchedInofLock.Lock()
	defer s.agentSchedInofLock.Unlock()

	agent, err := s.store.FetchAgentSchedInfo(hostname)
	if err != nil && !errors.Is(err, store.ErrNoFound) {
		blog.Errorf("get host(%s) schedinfo err(%s)", hostname, err.Error())
		return err
	}

	if agent == nil {
		if deltaResource == nil {
			blog.V(3).Infof("get host(%s) schedinfo return empty when delete taskgroup(%s) delta resource",
				hostname, taskGroupID)
			return nil
		}

		blog.Infof("get host(%s) schedinfo return empty, create it", hostname)
		agent = &types.AgentSchedInfo{
			HostName:   hostname,
			DeltaCPU:   0,
			DeltaMem:   0,
			DeltaDisk:  0,
			Taskgroups: make(map[string]*types.Resource),
		}
	}
	if agent.Taskgroups == nil {
		blog.Warnf("get host(%s) schedinfo ,create taskroup map", hostname)
		agent.Taskgroups = make(map[string]*types.Resource)
	}

	// delete taskgroup delta info
	if deltaResource == nil {
		blog.Infof("delete taskgroup(%s) from host(%s) schedinfo", taskGroupID, hostname)
		delete(agent.Taskgroups, taskGroupID)
	} else {
		// add or update taskgroup delta info
		blog.Infof("set taskgroup(%s)(delta: %f | %f | %f) in host(%s) schedinfo list",
			taskGroupID, deltaResource.Cpus, deltaResource.Mem, deltaResource.Disk, hostname)
		agent.Taskgroups[taskGroupID] = deltaResource
	}

	// computer total delta resource for agent
	agent.DeltaCPU = 0
	agent.DeltaMem = 0
	agent.DeltaDisk = 0
	for id, data := range agent.Taskgroups {
		blog.V(3).Infof("delta resource for host(%s) taskgroup(%s) : %f | %f  | %f ",
			hostname, id, data.Cpus, data.Mem, data.Disk)
		agent.DeltaCPU += data.Cpus
		agent.DeltaMem += data.Mem
		agent.DeltaDisk += data.Disk
	}
	blog.Info("delta resource for host(%s): %f | %f | %f ",
		hostname, agent.DeltaCPU, agent.DeltaMem, agent.DeltaDisk)

	err = s.store.SaveAgentSchedInfo(agent)
	if err != nil {
		blog.Info("save host(%s) schedinfo err(%s)", hostname, err.Error())
		return err
	}

	return nil
}

// GetClusterId Get Cluster ID
func (s *Scheduler) GetClusterId() string {
	return s.BcsClusterId
}

// GetFirstOffer Get current first offer from pool
func (s *Scheduler) GetFirstOffer() *offer.Offer {
	return s.offerPool.GetFirstOffer()
}

// GetNextOffer Get next offer from pool
func (s *Scheduler) GetNextOffer(offer *offer.Offer) *offer.Offer {
	return s.offerPool.GetNextOffer(offer)
}

// GetAllOffers Get current all offers
func (s *Scheduler) GetAllOffers() []*offer.Offer {
	return s.offerPool.GetAllOffers()
}

// UseOffer Use offer
func (s *Scheduler) UseOffer(o *offer.Offer) bool {
	return s.offerPool.UseOffer(o)
}

// GetClusterResource Get cluster resources
func (s *Scheduler) GetClusterResource() (*commtype.BcsClusterResource, error) {

	blog.Info("get cluster resource from mesos master")
	if s.currMesosMaster == "" {
		blog.Error("get cluster resource error: no mesos master")
		return nil, fmt.Errorf("system error: no mesos master")
	}

	return s.GetMesosResourceIn(s.operatorClient)
}

// GetMesosResourceIn Get cluster current resource information from mesos master
func (s *Scheduler) GetMesosResourceIn(mesosClient *client.Client) (*commtype.BcsClusterResource, error) {

	if mesosClient == nil {
		blog.Error("get cluster resource error: mesos Client is nil")
		return nil, fmt.Errorf("system error: mesos client is nil")
	}

	call := &master.Call{
		Type: master.Call_GET_AGENTS.Enum(),
	}
	req, err := proto.Marshal(call)
	if err != nil {
		blog.Error("get cluster resource: query agentInfo proto.Marshal err: %s", err.Error())
		return nil, fmt.Errorf("system error: proto marshal error")
	}
	resp, err := mesosClient.Send(req)
	if err != nil {
		blog.Error("get cluster resource: query agentInfo Send err: %s", err.Error())
		return nil, fmt.Errorf("send request to mesos error: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		blog.Error("get cluster resource: query agentInfo unexpected response statusCode: %d", resp.StatusCode)
		return nil, fmt.Errorf("mesos response statuscode: %d", resp.StatusCode)
	}

	var response master.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		blog.Error("get cluster resource: Decode response failed: %s", err.Error())
		return nil, fmt.Errorf("mesos response decode err: %s", err.Error())
	}
	blog.V(3).Infof("get cluster resource: response msg type(%d)", response.GetType())
	agentInfo := response.GetGetAgents()
	if agentInfo == nil {
		blog.Warn("get cluster resource: response Agents == nil")
	}

	clusterRes := new(commtype.BcsClusterResource)
	cpuTotal := 0.0
	cpuUsed := 0.0
	memTotal := 0.0
	memUsed := 0.0
	diskTotal := 0.0
	diskUsed := 0.0
	for _, oneAgent := range agentInfo.Agents {
		agent := new(commtype.BcsClusterAgentInfo)
		//blog.V(3).Infof("get agents: ===>agent[%d]: %+v", index, oneAgent)
		agent.HostName = oneAgent.GetAgentInfo().GetHostname()

		szSplit := strings.Split(oneAgent.GetPid(), "@")
		if len(szSplit) == 2 {
			agent.IP = szSplit[1]
		} else {
			agent.IP = oneAgent.GetPid()
		}
		if strings.Contains(agent.IP, ":") {
			agent.IP = strings.Split(agent.IP, ":")[0]
		}

		totalRes := oneAgent.GetTotalResources()
		for _, resource := range totalRes {
			if resource.GetName() == "cpus" {
				agent.CpuTotal = resource.GetScalar().GetValue()
				cpuTotal += agent.CpuTotal
			}
			if resource.GetName() == "mem" {
				agent.MemTotal = resource.GetScalar().GetValue()
				memTotal += agent.MemTotal
			}
			if resource.GetName() == "disk" {
				agent.DiskTotal = resource.GetScalar().GetValue()
				diskTotal += agent.DiskTotal
			}
		}

		//get delta resources for this agent
		agentDeltaCPU := 0.0
		agentDeltaMem := 0.0
		agentDeltaDisk := 0.0
		agentSchedInfo, _ := s.FetchAgentSchedInfo(agent.HostName)
		if agentSchedInfo != nil {
			if agentSchedInfo.DeltaCPU > 0 {
				agentDeltaCPU = agentSchedInfo.DeltaCPU
			}
			if agentSchedInfo.DeltaMem > 0 {
				agentDeltaMem = agentSchedInfo.DeltaMem
			}
			if agentSchedInfo.DeltaDisk > 0 {
				agentDeltaDisk = agentSchedInfo.DeltaDisk
			}
		}

		usedRes := oneAgent.GetAllocatedResources()
		for _, resource := range usedRes {
			if resource.GetName() == "cpus" {
				agent.CpuUsed = resource.GetScalar().GetValue() + agentDeltaCPU
				cpuUsed += agent.CpuUsed
			}
			if resource.GetName() == "mem" {
				agent.MemUsed = resource.GetScalar().GetValue() + agentDeltaMem
				memUsed += agent.MemUsed
			}
			if resource.GetName() == "disk" {
				agent.DiskUsed = resource.GetScalar().GetValue() + agentDeltaDisk
				diskUsed += agent.DiskUsed
			}
		}

		// added  20180929, add attributes into agent info
		// HostAttributes is set in mesos slave
		// Attributes is HostAttributes + AgentSettings, which are the exact attributes that matters while scheduling.
		agent.HostAttributes = mesosAttribute2commonAttribute(oneAgent.AgentInfo.Attributes)
		agent.Attributes = agent.HostAttributes
		settings, err := s.FetchAgentSetting(agent.IP)
		if err != nil && err != store.ErrNoFound {
			blog.Errorf("get cluster resource: query ageng settings failed IP(%s): %v", agent.IP, err)
			return nil, err
		}

		if settings != nil {
			agent.Disabled = settings.Disabled
			for key, value := range settings.AttrStrings {
				agent.Attributes = append(agent.Attributes, &commtype.BcsAgentAttribute{
					Name: key,
					Type: commtype.MesosValueType_Text,
					Text: &commtype.MesosValue_Text{
						Value: value.Value,
					},
				})
			}
			for key, value := range settings.AttrScalars {
				agent.Attributes = append(agent.Attributes, &commtype.BcsAgentAttribute{
					Name: key,
					Type: commtype.MesosValueType_Scalar,
					Scalar: &commtype.MesosValue_Scalar{
						Value: value.Value,
					},
				})
			}
		}

		if oneAgent.RegisteredTime != nil && oneAgent.RegisteredTime.Nanoseconds != nil {
			agent.RegisteredTime = *oneAgent.RegisteredTime.Nanoseconds
		}
		if oneAgent.ReregisteredTime != nil && oneAgent.ReregisteredTime.Nanoseconds != nil {
			agent.ReRegisteredTime = *oneAgent.ReregisteredTime.Nanoseconds
		}

		clusterRes.Agents = append(clusterRes.Agents, *agent)
	}

	clusterRes.CpuTotal = cpuTotal
	clusterRes.MemTotal = memTotal
	clusterRes.DiskTotal = diskTotal
	clusterRes.CpuUsed = cpuUsed
	clusterRes.MemUsed = memUsed
	clusterRes.DiskUsed = diskUsed

	blog.Info("get cluster resource: cpu %f/%f  || mem %f/%f || disk %f/%f",
		cpuUsed, cpuTotal, memUsed, memTotal, diskUsed, diskTotal)

	blog.V(3).Infof("get cluster resource: %+v", clusterRes)

	return clusterRes, nil
}

// GetCurrentOffers get all offers from offer pool
func (s *Scheduler) GetCurrentOffers() []*types.OfferWithDelta {
	offers := s.offerPool.GetAllOffers()

	inOffers := make([]*types.OfferWithDelta, 0)
	for _, o := range offers {
		inOffers = append(inOffers, &types.OfferWithDelta{
			Offer: o.Offer,
			DeltaResource: &types.Resource{
				Cpus: o.DeltaCPU,
				Mem:  o.DeltaMem,
				Disk: o.DeltaDisk,
			},
		})
	}
	return inOffers
}

//convert mesos.Attribute to commtype.BcsAgentAttribute
func mesosAttribute2commonAttribute(oldAttributeList []*mesos.Attribute) []*commtype.BcsAgentAttribute {
	if oldAttributeList == nil {
		return nil
	}

	attributeList := make([]*commtype.BcsAgentAttribute, 0)

	for _, oldAttribute := range oldAttributeList {
		if oldAttribute == nil {
			continue
		}

		attribute := new(commtype.BcsAgentAttribute)
		if oldAttribute.Name != nil {
			attribute.Name = *oldAttribute.Name
		}
		if oldAttribute.Type != nil {
			switch *oldAttribute.Type {
			case mesos.Value_SCALAR:
				attribute.Type = commtype.MesosValueType_Scalar
				if oldAttribute.Scalar != nil && oldAttribute.Scalar.Value != nil {
					attribute.Scalar = &commtype.MesosValue_Scalar{
						Value: *oldAttribute.Scalar.Value,
					}
				}
			case mesos.Value_RANGES:
				attribute.Type = commtype.MesosValueType_Ranges
				if oldAttribute.Ranges != nil {
					rangeList := make([]*commtype.MesosValue_Ranges, 0)
					for _, oldRange := range oldAttribute.Ranges.Range {
						newRange := &commtype.MesosValue_Ranges{}
						if oldRange.Begin != nil {
							newRange.Begin = *oldRange.Begin
						}
						if oldRange.End != nil {
							newRange.End = *oldRange.End
						}
						rangeList = append(rangeList, newRange)
					}
				}
			case mesos.Value_SET:
				attribute.Type = commtype.MesosValueType_Set
				if oldAttribute.Set != nil {
					attribute.Set = &commtype.MesosValue_Set{
						Item: oldAttribute.Set.Item,
					}
				}
			case mesos.Value_TEXT:
				attribute.Type = commtype.MesosValueType_Text
				if oldAttribute.Text != nil && oldAttribute.Text.Value != nil {
					attribute.Text = &commtype.MesosValue_Text{
						Value: *oldAttribute.Text.Value,
					}
				}
			}
		}
		attributeList = append(attributeList, attribute)
	}
	return attributeList
}

// FetchTaskGroup get taskgroup from taskID
func (s *Scheduler) FetchTaskGroup(taskGroupID string) (*types.TaskGroup, error) {
	return s.store.FetchTaskGroup(taskGroupID)
}

//CheckPodBelongDaemonset check taskgroup whether belongs to daemonset
func (s *Scheduler) CheckPodBelongDaemonset(taskgroupID string) bool {
	namespace, name := types.GetRunAsAndAppIDbyTaskGroupID(taskgroupID)
	version, err := s.store.GetVersion(namespace, name)
	if err != nil {
		blog.Errorf("Fetch taskgroup(%s) version(%s.%s) error %s", taskgroupID, namespace, name, err.Error())
		return false
	}
	if version == nil {
		blog.Errorf("Fetch taskgroup(%s) version(%s.%s) is empty", taskgroupID, namespace, name)
		return false
	}

	if version.Kind == commtype.BcsDataType_Daemonset {
		return true
	}
	return false
}
