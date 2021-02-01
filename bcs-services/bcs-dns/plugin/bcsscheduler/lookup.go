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
	"context"
	"fmt"
	"log"
	"net"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	bcstypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	dnsUtil "github.com/Tencent/bk-bcs/bcs-services/bcs-dns/plugin/util"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/etcd/msg"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

const (
	defaultNSName          = "ns.scheduler."
	defaultStoragePrefix   = "bcsdns"
	defaultDNSVersion      = "1.0.0"
	serviceTypeClusterIP   = "ClusterIP"
	serviceTypeIntegration = "Integration"
)

type service struct {
	name        string     //service name
	namespace   string     //namespace
	clusterIP   []string   //cluter ip for loadbalance proxy, reserved
	serviceport []endport  //port for cluster ip. reserved
	endpoints   []endpoint //ip & port
}

type endpoint struct {
	addr     string    //endpoint address
	endports []endport //port for exposed in endpoint
}

type endport struct {
	name     string //port name
	port     string //container port
	hport    string //host port
	protocol string //protocol
}

// Records looks up services in bcs. If exact is true, it will lookup
// just this name. This is used when find matches when completing SRV lookups
// for instance.
func (bcs *BcsScheduler) Records(state request.Request, exact bool) ([]msg.Service, error) {
	r := bcs.parseRequest(state.Name(), state.QType())
	return bcs.records(r)
}

func (bcs *BcsScheduler) records(req recordRequest) ([]msg.Service, error) {
	var svcList []service
	if isWildchard(req.serviceName) || isWildchard(req.namespace) {
		if req.IsPodDNS() {
			return nil, fmt.Errorf("unsupported wirdchard match in pod dns")
		}
		//we get wildchard in domain request, so we need
		//iterator all endpoints to find name matched
		list, err := bcs.wildChardRecords(req)
		if err != nil {
			return nil, err
		}
		svcList = append(svcList, list...)
	} else {
		//get domain by namespace/service directly
		localSvc, err := bcs.svcRecords(req)
		if err != nil {
			return nil, err
		}
		svcList = append(svcList, *localSvc)
	}
	if len(svcList) == 0 {
		//not endpoints info found in cache
		log.Printf("[WARN] DNSlookup found no service record for request %v", req)
		return nil, errNoItems
	}
	records := bcs.formatToMessage(svcList, req)
	return records, nil
}

func (bcs *BcsScheduler) svcRecords(req recordRequest) (*service, error) {
	key := filepath.Join(req.namespace, req.serviceName)
	epItem, ok, gerr := bcs.endpointCache.Store.GetByKey(key)
	if !ok || gerr != nil {
		log.Printf("[ERROR] DNSlookup Get no item with key %s in Cache", key)
		return nil, errNoItems
	}
	bcsEndpoint := epItem.(*bcstypes.BcsEndpoint)
	if len(bcsEndpoint.Endpoints) == 0 {
		log.Printf("[ERROR] DNSlookup Get no endpoint with key %s in Cache", key)
		return nil, errNoItems
	}
	localSvc := &service{
		name:      req.serviceName,
		namespace: req.namespace,
	}
	//checking if service is in ClusterMode
	bcsSvc := bcs.svcCache.GetServiceByEndpoint(bcsEndpoint)
	if bcsSvc != nil && (bcsSvc.Spec.Type == serviceTypeClusterIP || bcsSvc.Spec.Type == serviceTypeIntegration) {
		if req.IsPodDNS() {
			return nil, fmt.Errorf("unsupported cluster ip svc usage in pod dns")
		}
		endpoints, err := bcs.formatClusterIPSvc(bcsSvc, req)
		if nil != err {
			return nil, err
		}
		localSvc.endpoints = append(localSvc.endpoints, endpoints...)
	} else {
		//get service ip address list
		for _, end := range bcsEndpoint.Endpoints {
			if req.IsPodDNS() {
				if req.podname != end.Target.Name {
					continue
				}
			}
			ep := bcs.formatBcsEndpoint(end, req)
			localSvc.endpoints = append(localSvc.endpoints, ep)
		}

	}
	return localSvc, nil
}

func (bcs *BcsScheduler) wildChardRecords(req recordRequest) ([]service, error) {
	var svcList []service
	for _, epItem := range bcs.endpointCache.ListEndpoints() {
		if !isWildchard(req.namespace) && epItem.GetNamespace() != req.namespace {
			continue
		}
		if !isWildchard(req.serviceName) && epItem.GetName() != req.serviceName {
			continue
		}
		if len(epItem.Endpoints) == 0 {
			continue
		}
		localSvc := service{
			name:      req.serviceName,
			namespace: req.namespace,
		}
		//checking if service is in ClusterIP Mode
		bcsSvc := bcs.svcCache.GetServiceByEndpoint(epItem)
		if bcsSvc != nil && bcsSvc.Spec.Type == serviceTypeClusterIP {
			endpoints, err := bcs.formatClusterIPSvc(bcsSvc, req)
			if nil != err {
				return nil, err
			}
			localSvc.endpoints = append(localSvc.endpoints, endpoints...)

		} else {
			//get service ip address list
			for _, end := range epItem.Endpoints {
				ep := bcs.formatBcsEndpoint(end, req)
				localSvc.endpoints = append(localSvc.endpoints, ep)
			}
		}
		svcList = append(svcList, localSvc)
	}
	return svcList, nil
}

func (bcs *BcsScheduler) recordsForTXT(r recordRequest, svcs *[]msg.Service) (err error) {
	switch r.typeName {
	case "dns-version":
		s := msg.Service{
			Text: defaultDNSVersion,
			TTL:  28800,
			Key:  msg.Path(strings.Join([]string{r.typeName, r.zone}, "."), defaultStoragePrefix),
		}
		*svcs = append(*svcs, s)
		return nil
	}
	return nil
}

func (bcs *BcsScheduler) recordsForNS(r recordRequest, svcs *[]msg.Service) error {
	ns := bcs.selfDNSRecord()
	s := msg.Service{
		Host: ns.A.String(),
		Key:  msg.Path(strings.Join([]string{ns.Hdr.Name, r.zone}, "."), defaultStoragePrefix),
	}
	*svcs = append(*svcs, s)
	return nil
}

func (bcs *BcsScheduler) selfDNSRecord() dns.A {
	//get local ip address & default dns name
	addrList := util.GetIPAddress()
	var self dns.A
	if len(addrList) == 0 {
		return self
	}
	self.Hdr.Name = defaultNSName
	self.A = net.ParseIP(addrList)
	return self
}

//formatToMessage format inner service struct to etcd service message
func (bcs *BcsScheduler) formatToMessage(svcs []service, req recordRequest) (records []msg.Service) {
	for _, svc := range svcs {
		domainpart := svc.name + "." + svc.namespace + ".svc." + req.zone
		for _, end := range svc.endpoints {
			if len(end.endports) > 0 {
				//we get port info, for SRV
				for _, port := range end.endports {
					iport, err := strconv.Atoi(port.port)
					if err != nil {
						continue
					}
					s := msg.Service{
						Key:  msg.Path(domainpart, defaultStoragePrefix),
						Host: end.addr,
						Port: iport,
					}
					records = append(records, s)
				}
			} else {
				//only for typeA ?!
				s := msg.Service{
					Key:  msg.Path(domainpart, defaultStoragePrefix),
					Host: end.addr,
				}
				records = append(records, s)
			}
		}

	}
	return records
}

//formatBcsEndpoint format bcs endpoint to local endpoint
func (bcs *BcsScheduler) formatBcsEndpoint(end bcstypes.Endpoint, req recordRequest) endpoint {
	endpt := endpoint{}
	if "bridge" == strings.ToLower(end.NetworkMode) || "host" == strings.ToLower(end.NetworkMode) {
		endpt.addr = end.NodeIP
	} else {
		endpt.addr = end.ContainerIP
	}
	//check port & protocol info for SRV
	for _, port := range end.Ports {
		if !isWildchard(port.Protocol) && strings.ToLower(port.Protocol) != req.protocol {
			continue
		}
		if !isWildchard(port.Name) && strings.ToLower(port.Name) != req.port {
			continue
		}
		eport := endport{
			name:     port.Name,
			protocol: strings.ToLower(port.Protocol),
			port:     strconv.Itoa(port.ContainerPort),
		}
		endpt.endports = append(endpt.endports, eport)
	}
	return endpt
}

//formatBcsService format bcs service to local endpoint when BcsService under ClusterIP mode
func (bcs *BcsScheduler) formatClusterIPSvc(svc *bcstypes.BcsService, req recordRequest) ([]endpoint, error) {
	eps := make([]endpoint, 0)

	if svc.Spec.ClusterIP == nil {
		return eps, fmt.Errorf("invalid cluster ip value in service: %s, ns: %s", svc.Name, svc.NameSpace)
	} else if len(svc.Spec.ClusterIP) == 0 {
		return eps, fmt.Errorf("invalid cluster ip value in service: %s, ns: %s", svc.Name, svc.NameSpace)
	}

	for _, ip := range svc.Spec.ClusterIP {
		endpt := endpoint{}
		if isIP := net.ParseIP(ip); nil == isIP {
			// this is a domain. look up upper first
			continue
		} else {
			endpt.addr = ip
		}
		for _, port := range svc.Spec.Ports {
			if !isWildchard(port.Protocol) && strings.ToLower(port.Protocol) != req.protocol {
				continue
			}
			if !isWildchard(port.Name) && strings.ToLower(port.Name) != req.port {
				continue
			}
			eport := endport{
				name:     port.Name,
				protocol: strings.ToLower(port.Protocol),
				port:     strconv.Itoa(port.Port),
			}
			endpt.endports = append(endpt.endports, eport)
		}
		eps = append(eps, endpt)
	}
	return eps, nil
}

var svcRegexp = regexp.MustCompile(`^([a-zA-Z0-9-]+\.){2}(svc\.).*`)

func (bcs *BcsScheduler) dealRecursiveDNSWithClusterIP(state request.Request) ([]dns.RR, bool) {
	if state.QType() != dns.TypeA {
		return nil, false
	}

	r := bcs.parseRequest(state.Name(), state.QType())

	// only deal with typeA records.
	if state.QType() != dns.TypeA {
		return nil, false
	}

	// do not support pod dns recursive rule.
	if r.IsPodDNS() {
		return nil, false
	}

	svc := bcs.svcCache.GetService(r.namespace, r.serviceName)
	if svc == nil || (svc != nil && (svc.Spec.Type != "ClusterIP" && svc.Spec.Type != "Integration")) {
		return nil, false
	}

	originName := state.Name()
	originQuestion := (*state.Req).Question[0].Name
	allRR := make([]dns.RR, 0)
	for _, ip := range svc.Spec.ClusterIP {
		if isIP := net.ParseIP(ip); isIP != nil {
			// this is a ip, skip.
			continue
		}
		// this is a domain, need to recursive dns records
		domain := strings.TrimSuffix(strings.Trim(ip, " "), ".") + "."

		// TODO: this judge need to consider the self defined dns record in bcs-dns.
		// which will be needed in the future.
		if svcRegexp.MatchString(domain) {
			switch svc.Spec.Type {
			case "ClusterIP":
				// this is a service domain, belongs to upper bcs-dns
				(*state.Req).Question[0].Name = domain
				state.Clear()
				state.Name()
				(*state.Req).Question[0].Name = originName
				dnsMsg, err := bcs.Lookup(state, state.Name(), state.QType())
				if nil != err {
					log.Printf("[ERROR] mode: ClusterIP, recursive lookup service, name: %s domain[%s] recored faild, err :%v", state.Name(), domain, err)
					continue
				}
				//log.Printf("[DEBUG] recursive lookup service, name: %s, domain: %s success. msg: %+v", state.Name(), domain, dnsMsg)
				for _, answer := range dnsMsg.Answer {
					A, yes := answer.(*dns.A)
					if !yes {
						log.Printf("[WARN] can not convert recursive response to dns.A type, source name: %s, domain name: %s", state.Name(), domain)
						continue
					}
					// revise the domain to the original request name.
					A.Hdr.Name = originQuestion
					allRR = append(allRR, A)
				}
				continue

			case "Integration":
				//log.Printf("[DEBUG] -> run svc with Integration mode, request:%s.", domain)
				qualifiedName, _ := bcs.getQualifiedQuestionName(domain, dns.TypeA)
				if !bcs.inCurrentZone(qualifiedName) {
					log.Printf("[WARN] unsupported request[%s] with svc Integration mode.", domain)
					continue
				}
				zone := plugin.Zones(bcs.conf.Zones).Matches(state.Name())
				(*state.Req).Question[0].Name = qualifiedName
				state.Clear()
				state.Name()
				records, _, err := bcs.dealCommonResolve(zone, state)
				if err != nil {
					log.Printf("[ERROR] mode:Integration, get current zone resolve dns:[%s] failed, err:%v", qualifiedName, err)
					continue
				}
				for _, r := range records {
					A, yes := r.(*dns.A)
					if !yes {
						log.Printf("[WARN] can not convert recursive response to dns.A type, qualifed name: %s, domain name: %s", qualifiedName, domain)
						continue
					}
					// revise the domain to the original request name.
					A.Hdr.Name = originQuestion
					allRR = append(allRR, A)
				}
				continue

			}
		}

		// this is a out of bcs dns record.
		next := bcs.Next
		interceptor := dnsUtil.NewResponseInterceptor(state.W)
		r := new(dns.Msg)
		r.SetQuestion(domain, dns.TypeA)
		_, err := plugin.NextOrFailure(bcs.Name(), next, context.Background(), interceptor, r)
		if err != nil {
			log.Printf("[ERROR] recursive lookup domain[%s] with next failed. err: %v", domain, err)
			continue
		}
		if interceptor.Msg == nil {
			log.Printf("[WARN] recursive proxy domain: %s, got a nil reply.", domain)
			continue
		}
		for _, a := range interceptor.Msg.Answer {
			A, yes := a.(*dns.A)
			if !yes {
				log.Printf("[WARN] can not convert recursive response to dns.A type, source name: %s, domain name: %s", state.Name(), domain)
				continue
			}
			// revise the domain to the original request name.
			A.Hdr.Name = originQuestion
			allRR = append(allRR, A)
		}
	}
	return allRR, true
}
