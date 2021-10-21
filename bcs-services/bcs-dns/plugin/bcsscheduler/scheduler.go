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

package bcsscheduler

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/signals"
	bcstypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/cache"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/master"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-dns/plugin/bcsscheduler/controller"
	bcsSchedulerUtil "github.com/Tencent/bk-bcs/bcs-services/bcs-dns/plugin/bcsscheduler/util"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-dns/storage"
	etcdstorage "github.com/Tencent/bk-bcs/bcs-services/bcs-dns/storage/etcd"
	clientGoCache "k8s.io/client-go/tools/cache"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/etcd/msg"
	"github.com/coredns/coredns/plugin/pkg/dnsutil"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

//error definition
var errNoItems = errors.New("no items found")
var errInvalidRequest = errors.New("invalid query name")

//var errZoneMisMatch = errors.New("zone group dismatch")

// ForwardClusterIpDomainError err
type ForwardClusterIpDomainError struct {
	ClusterIP string
}

func (fc ForwardClusterIpDomainError) Error() string {
	return fmt.Sprintf("need to forward cluster ip domain: %s", fc.ClusterIP)
}

const (
	defaultDNSPath = "/bcs/services/endpoints/dns"
)

//NewScheduler create bcs scheduler backend
func NewScheduler(config *ConfigItem) *BcsScheduler {
	scheduler := &BcsScheduler{
		conf: config,
	}
	//create storage
	var err error
	if len(config.Storage) != 0 {
		scheduler.storage, err = etcdstorage.NewStorage(config.StoragePath, config.Storage, config.StorageCA, config.StorageCert, config.StorageKey)
	}
	if err != nil {
		log.Printf("[ERROR] scheduler create storage failed, %s", err.Error())
		return nil
	}
	if len(config.Register) == 0 {
		scheduler.registery = &master.Empty{}
	} else {
		hostname, _ := os.Hostname()
		node := &bcstypes.ServerInfo{
			IP:         util.GetIPAddress(),
			Port:       uint(53),
			Pid:        os.Getpid(),
			HostName:   hostname,
			Scheme:     "dns",
			Version:    version.BcsVersion,
			MetricPort: config.MetricPort,
		}
		path := defaultDNSPath
		scheduler.registery, err = master.NewZookeeperMaster(config.Register, path, node)
		if err != nil {
			log.Printf("[ERROR] scheduler create Register failed, %s", err.Error())
			return nil
		}
	}
	return scheduler
}

//recordRequest parse dns request to local info
type recordRequest struct {
	port        string //port for endpoint
	protocol    string //protocol for endpoint
	podname     string // pod's name
	serviceName string //service name
	namespace   string //namespace from domain
	typeName    string //type name, like svc/pod, pod is not supported
	zone        string //zone info
}

func (r recordRequest) IsPodDNS() bool {
	return len(r.podname) != 0
}

//BcsScheduler plugin for reading service/endpoints info from bcs-scheduler
type BcsScheduler struct {
	conf               *ConfigItem                     //all config item from configuration file
	svcController      controller.Controller           //service controller for cache
	svcCache           *bcsSchedulerUtil.ServiceCache  //service cache
	endpointController controller.Controller           //endpoint controller for cache
	endpointCache      *bcsSchedulerUtil.EndpointCache //endpoint cache
	storage            storage.Storage                 //remote etcd storage for all cluster
	registery          master.Master                   //Master interface
	Next               plugin.Handler                  //next plugin
}

// Services communicates with the backend to retrieve the service definition. if func Services is called,
// it means domain must belong to this cluster, Exact indicates on exact much are that we are allowed to recurs.
func (bcs *BcsScheduler) Services(state request.Request, exact bool, opt plugin.Options) (svcs []msg.Service, err error) {
	r := bcs.parseRequest(state.Name(), state.QType())

	switch state.QType() {
	case dns.TypeA:
		s, e := bcs.Records(state, true)
		return s, e // Haven't implemented debug queries yet.
	case dns.TypeSRV:
		s, e := bcs.Records(state, true)
		// SRV for external services is not yet implemented, so remove those records
		noext := []msg.Service{}
		for _, svc := range s {
			if t, _ := svc.HostType(); t != dns.TypeCNAME {
				noext = append(noext, svc)
			}
		}
		return noext, e
	case dns.TypeTXT:
		err = bcs.recordsForTXT(r, &svcs)
		return svcs, err
	case dns.TypeNS:
		err = bcs.recordsForNS(r, &svcs)
		return svcs, err
	}
	return nil, nil
}

// Reverse communicates with the backend to retrieve service definition based on a IP address
// instead of a name. I.e. a reverse DNS lookup.
func (bcs *BcsScheduler) Reverse(state request.Request, exact bool, opt plugin.Options) (svcList []msg.Service, err error) {
	ip := dnsutil.ExtractAddressFromReverse(state.Name())
	if ip == "" {
		return nil, nil
	}
	//search ip in all endpoints
	bcsEndpoints := bcs.endpointCache.ListEndpoints()
	if len(bcsEndpoints) == 0 {
		return svcList, errNoItems
	}
	for _, endpoint := range bcsEndpoints {
		//check sub item from
		if len(endpoint.Endpoints) == 0 {
			continue
		}
		for _, pod := range endpoint.Endpoints {
			if ip == pod.ContainerIP || ip == pod.NodeIP {
				//construct domain for this request ip, format:
				//serviceName.Namespace.svc.$cluster.$zone
				domain := endpoint.GetName() + "." + endpoint.GetNamespace() + ".svc." + bcs.PrimaryZone()
				svc := msg.Service{Host: domain}
				svcList = append(svcList, svc)
			}
		}
	}
	return svcList, nil
}

// Lookup is used to find records else where.
func (bcs *BcsScheduler) Lookup(state request.Request, name string, typ uint16) (*dns.Msg, error) {
	return bcs.conf.Proxy.Lookup(state, name, typ)
}

// IsNameError return true if err indicated a record not found condition
func (bcs *BcsScheduler) IsNameError(err error) bool {
	return err == errNoItems || err == errInvalidRequest
}

// Serial returns a SOA serial number to construct a SOA record.
func (bcs *BcsScheduler) Serial(state request.Request) uint32 {
	return uint32(time.Now().Unix())
}

// MinTTL returns the minimum TTL to be used in the SOA record.
func (bcs *BcsScheduler) MinTTL(state request.Request) uint32 {
	return 30
}

// Transfer handles a zone transfer it writes to the client just
// like any other handler.
func (bcs *BcsScheduler) Transfer(ctx context.Context, state request.Request) (int, error) {
	return dns.RcodeServerFailure, nil
}

// Debug returns a string used when returning debug services.
func (bcs *BcsScheduler) Debug() string {
	return "debug"
}

/**
 * inner method for data
 */

//InitSchedulerCache bcs scheduler backend init it
func (bcs *BcsScheduler) InitSchedulerCache() error {
	//Register for master
	if err := bcs.registery.Init(); err != nil {
		return err
	}
	if err := bcs.registery.Register(); err != nil {
		return err
	}
	//create bcs-service controller
	store := cache.CreateCache(bcsSchedulerUtil.DNSDataKeyFunc)
	bcs.svcCache = &bcsSchedulerUtil.ServiceCache{
		Store: store,
	}
	svc := filepath.Join(bcs.conf.EndpointPath, "service")
	svcEvents := &clientGoCache.ResourceEventHandlerFuncs{
		AddFunc:    bcs.svcOnAdd,
		UpdateFunc: bcs.svcOnUpdate,
		DeleteFunc: bcs.svcOnDelete,
	}

	//create bcs endpoint controller
	estore := cache.CreateCache(bcsSchedulerUtil.DNSDataKeyFunc)
	bcs.endpointCache = &bcsSchedulerUtil.EndpointCache{
		Store: estore,
	}
	endpoint := filepath.Join(bcs.conf.EndpointPath, "endpoint")
	endpointEvents := &clientGoCache.ResourceEventHandlerFuncs{
		AddFunc:    bcs.endpointOnAdd,
		UpdateFunc: bcs.endpointOnUpdate,
		DeleteFunc: bcs.endpointOnDelete,
	}

	var svcErr error
	var epErr error
	if len(bcs.conf.Endpoints) != 0 {
		bcs.svcController, svcErr = controller.NewZkController(bcs.conf.Endpoints, svc, bcs.conf.ResyncPeriod, bcs.svcCache.Store, &bcsSchedulerUtil.SvcDecoder{}, svcEvents)
		if svcErr != nil {
			log.Printf("[ERROR] Scheduler create BcsService Controller failed, %s", svcErr.Error())
			return svcErr
		}

		bcs.endpointController, epErr = controller.NewZkController(bcs.conf.Endpoints, endpoint, bcs.conf.ResyncPeriod, bcs.endpointCache.Store, &bcsSchedulerUtil.EndpointDecoder{}, endpointEvents)
		if epErr != nil {
			log.Printf("[ERROR] Scheduler create BcsEndpoint Controller failed, %s", epErr.Error())
			return epErr
		}
	} else if bcs.conf.KubeConfig != "" {
		bcs.svcController, svcErr = controller.NewEtcdController(bcs.conf.KubeConfig, "service", bcs.conf.ResyncPeriod, bcs.svcCache.Store, svcEvents)
		if svcErr != nil {
			log.Printf("[ERROR] Scheduler create BcsService Controller failed, %s", svcErr.Error())
			return svcErr
		}

		bcs.endpointController, svcErr = controller.NewEtcdController(bcs.conf.KubeConfig, "endpoint", bcs.conf.ResyncPeriod, bcs.endpointCache.Store, endpointEvents)
		if epErr != nil {
			log.Printf("[ERROR] Scheduler create BcsEndpoint Controller failed, %s", epErr.Error())
			return epErr
		}
	} else {
		return fmt.Errorf("scheduler create controller failed, no endpoints and kubeconfig provided")
	}

	//todo(developer): create etcdStorage for data persistence

	return nil
}

//Start start all go event with context
func (bcs *BcsScheduler) Start() error {
	log.Printf("%s", version.GetVersion())
	time.Sleep(time.Second * 1)
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	go bcs.svcController.RunController(stopCh)
	go bcs.endpointController.RunController(stopCh)
	//todo(developer): starting etcdStorage event goroutine
	return nil
}

//Stop stop all event, clean cache and exit
func (bcs *BcsScheduler) Stop() error {
	bcs.registery.Clean()
	bcs.registery.Finit()
	bcs.svcController.StopController()
	bcs.endpointController.StopController()
	time.Sleep(time.Second * 5)
	return nil
}

//PrimaryZone Get zone info
func (bcs *BcsScheduler) PrimaryZone() string {
	return bcs.conf.Cluster + "." + bcs.conf.Zones[0]
}

func (bcs *BcsScheduler) parseRequest(name string, dnsType uint16) (req recordRequest) {
	//only support two request:
	// * SRV Request: _port._protocol.service.namespace.svc.zone
	// * A Request (service): service.namespace.svc.zone
	// * A Request (pod): podname.service.namespace.svc.zone

	//defer log.Printf("[DEBUG] -> parse [%s] request: %+v", name, req)
	var segments []string
	segments = dns.SplitDomainName(name)
	segments = segments[:len(segments)-dns.CountLabel(bcs.PrimaryZone())]
	length := len(segments)

	if len(segments) == 1 && dnsType == dns.TypeTXT {
		req.typeName = segments[0]
		return req
	}

	if length < 3 {
		// this can not be happen.
		return recordRequest{}
	}

	req.zone = bcs.PrimaryZone()

	//now, segments contains all part without zone info
	//svcIndex := 0
	if dnsType == dns.TypeSRV {
		//construct port & protocol
		if strings.HasPrefix(segments[0], "_") {
			req.port = strings.Replace(segments[0], "_", "", 1)
		} else if isWildchard(segments[0]) {
			req.port = segments[0]
		}
		if strings.HasPrefix(segments[1], "_") {
			req.protocol = strings.Replace(segments[1], "_", "", 1)
		} else if isWildchard(segments[1]) {
			req.protocol = segments[1]
		}

	}

	if dnsType == dns.TypeA && length == 4 {
		// this is a pod record
		req.podname = segments[0]
	}

	req.typeName = segments[length-1]
	req.namespace = segments[length-2]
	req.serviceName = segments[length-3]

	return req
}

func isWildchard(s string) bool {
	return s == "*"
}
