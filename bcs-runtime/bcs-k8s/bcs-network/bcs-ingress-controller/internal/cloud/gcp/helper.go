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

package gcp

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"google.golang.org/api/compute/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// do ensure network lb listener, create lb listener by service, return service name
func (e *GCLB) ensureNetworkLBListener(region string, listener *networkextensionv1.Listener) (string, error) {
	// ensure service without selector
	serviceName, err := e.ensureListenerService(region, listener)
	if err != nil {
		return "", err
	}

	// ensure endpoints
	if err := e.ensureListenerEndpoints(listener); err != nil {
		return "", err
	}

	return serviceName, nil
}

// do create application lb listener, support multiple target groups
// forwarding rules --> target http/https proxy --> url maps --> (url 1)backend service(with health check) --> network endpoint group --> pod ip:port
//                                                           --> (url 2)backend service
//                                                           --> (url 3)backend service
func (e *GCLB) ensureApplicationLBListener(region string, listener *networkextensionv1.Listener) (string, error) {
	name, err := e.ensureForwardingRules(listener)
	if err != nil {
		return "", err
	}
	if err := e.ensureL7Rules(listener.Spec.Rules, listener.Spec.Protocol,
		listener.Name, listener.Spec.Port); err != nil {
		return "", err
	}
	return name, nil
}

// ensure service for lb
func (e *GCLB) ensureListenerService(region string, listener *networkextensionv1.Listener) (string, error) {
	// get lb ip
	address, err := e.sdkWrapper.GetAddress(e.project, region, listener.Spec.LoadbalancerID)
	if err != nil {
		blog.Errorf("GetAddress failed, err %s", err.Error())
		return "", fmt.Errorf("GetAddress failed, err %s", err.Error())
	}

	// ensure service without selector
	service := &k8scorev1.Service{}
	objectKey := types.NamespacedName{Namespace: listener.Namespace, Name: listener.Name}
	if err := e.client.Get(context.TODO(), objectKey, service); err != nil {
		if k8serrors.IsNotFound(err) {
			// create service
			service := e.generateListenerService(listener, address.Address)
			if err := e.client.Create(context.TODO(), service); err != nil {
				return "", err
			}
			return service.Name, nil
		}
		return "", err
	}

	if service.DeletionTimestamp != nil {
		blog.Warnf("service %s is being deleted, retry later", service.Name)
		return "", fmt.Errorf("service %s is being deleted, retry later", service.Name)
	}

	// update service
	generateService := e.generateListenerService(listener, address.Address)
	if !reflect.DeepEqual(service.Spec, generateService.Spec) {
		service.Spec = generateService.Spec
		if err := e.client.Update(context.TODO(), service); err != nil {
			return "", err
		}
	}
	return service.Name, nil
}

// generate service for lb listener
func (e *GCLB) generateListenerService(listener *networkextensionv1.Listener, lbIP string) *k8scorev1.Service {
	service := &k8scorev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: listener.Namespace,
			Name:      listener.Name,
		},
		Spec: k8scorev1.ServiceSpec{
			LoadBalancerIP:        lbIP,
			ExternalTrafficPolicy: k8scorev1.ServiceExternalTrafficPolicyTypeLocal,
			Type:                  k8scorev1.ServiceTypeLoadBalancer,
		},
	}
	var ports []k8scorev1.ServicePort
	// get rs port
	targetPort := 0
	if listener.Spec.TargetGroup != nil {
		for _, v := range listener.Spec.TargetGroup.Backends {
			targetPort = v.Port
			break
		}
	}
	if listener.Spec.EndPort == 0 {
		ports = append(ports, k8scorev1.ServicePort{
			Name:       strconv.Itoa(targetPort),
			Protocol:   k8scorev1.Protocol(listener.Spec.Protocol),
			Port:       int32(listener.Spec.Port),
			TargetPort: intstr.FromInt(targetPort),
		})
	}

	// segment listener
	for i, rs := listener.Spec.Port, targetPort; i <= listener.Spec.EndPort; i, rs = i+1, rs+1 {
		ports = append(ports, k8scorev1.ServicePort{
			Name:       strconv.Itoa(rs),
			Protocol:   k8scorev1.Protocol(listener.Spec.Protocol),
			Port:       int32(i),
			TargetPort: intstr.FromInt(rs),
		})
	}
	service.Spec.Ports = ports
	return service
}

func (e *GCLB) ensureListenerEndpoints(listener *networkextensionv1.Listener) error {
	ep := &k8scorev1.Endpoints{}
	objectKey := types.NamespacedName{Namespace: listener.Namespace, Name: listener.Name}
	if err := e.client.Get(context.TODO(), objectKey, ep); err != nil {
		if k8serrors.IsNotFound(err) {
			// create endpoints
			ep := e.generateListenerEndpoints(listener)
			if ep == nil {
				blog.Warnf("endpoints %s is empty", objectKey.String())
				return nil
			}
			return e.client.Create(context.TODO(), ep)
		}
		return err
	}

	if ep.DeletionTimestamp != nil {
		blog.Warnf("endpoints %s is being deleted, retry later", objectKey.String())
		return fmt.Errorf("endpoints %s is being deleted, retry later", objectKey.String())
	}

	// update endpoints
	generateEP := e.generateListenerEndpoints(listener)
	if generateEP.Subsets == nil {
		blog.Warnf("new endpoints %s is empty", objectKey.String())
		e.client.Delete(context.TODO(), generateEP)
		return nil
	}
	if !reflect.DeepEqual(ep.Subsets, generateEP.Subsets) {
		ep.Subsets = generateEP.Subsets
		return e.client.Update(context.TODO(), ep)
	}
	return nil
}

type backend struct {
	IP       string
	Port     int
	NodeName string
}

// generate endpoints for lb listener
func (e *GCLB) generateListenerEndpoints(listener *networkextensionv1.Listener) *k8scorev1.Endpoints {
	ep := &k8scorev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: listener.Namespace,
			Name:      listener.Name,
		},
	}
	if listener.Spec.TargetGroup == nil {
		return ep
	}

	// add target ip
	targetIPs := make(map[string]backend, 0)
	for _, v := range listener.Spec.TargetGroup.Backends {
		targetIPs[v.IP] = backend{IP: v.IP, Port: v.Port}
	}

	pods := &k8scorev1.PodList{}
	if err := e.client.List(context.TODO(), pods, &client.ListOptions{}); err != nil ||
		len(pods.Items) == 0 {
		return ep
	}

	// set target node name
	for _, pod := range pods.Items {
		if _, ok := targetIPs[pod.Status.PodIP]; ok {
			targetIPs[pod.Status.PodIP] = backend{
				IP:       pod.Status.PodIP,
				Port:     targetIPs[pod.Status.PodIP].Port,
				NodeName: pod.Spec.NodeName,
			}
		}
	}

	// generate subsets
	subsets := k8scorev1.EndpointSubset{}
	ports := make([]k8scorev1.EndpointPort, 0)
	portsMap := make(map[int]bool, 0)
	for _, v := range targetIPs {
		nodeName := v.NodeName
		subsets.Addresses = append(subsets.Addresses, k8scorev1.EndpointAddress{IP: v.IP, NodeName: &nodeName})
		if _, ok := portsMap[v.Port]; !ok {
			ports = append(ports, k8scorev1.EndpointPort{Name: strconv.Itoa(v.Port), Port: int32(v.Port),
				Protocol: k8scorev1.Protocol(listener.Spec.Protocol)})
			portsMap[v.Port] = true
		}
	}
	subsets.Ports = ports
	ep.Subsets = append(ep.Subsets, subsets)
	return ep
}

func (e *GCLB) ensureURLMap(listener *networkextensionv1.Listener) (string, error) {
	// ensure default health check
	if err := e.ensureL7HealthCheck(networkextensionv1.ListenerRule{Domain: listener.Name}, listener.Spec.Port); err != nil {
		return "", err
	}

	// ensure default backend-service, default backend-service is used for creating urlmap
	// default backend-service has no endpoints
	bsName, err := e.ensureL7BackendService(networkextensionv1.ListenerRule{Domain: listener.Name},
		listener.Spec.Protocol, listener.Spec.Port)
	if err != nil {
		return "", err
	}

	// ensure url map
	_, err = e.sdkWrapper.GetURLMaps(e.project, listener.Name)
	if err != nil {
		if IsNotFound(err) {
			err := e.sdkWrapper.CreateURLMap(e.project, &compute.UrlMap{Name: listener.Name, DefaultService: "global/backendServices/" + bsName})
			if err != nil {
				blog.Errorf("CreateURLMap %s failed, %s", listener.Name, err.Error())
				return "", err
			}
			return listener.Name, nil
		}
		blog.Errorf("GetURLMaps %s failed, %s", listener.Name, err.Error())
		return "", err
	}
	return listener.Name, nil
}

func (e *GCLB) ensureTargetProxy(listener *networkextensionv1.Listener) (string, error) {
	urlMapName, err := e.ensureURLMap(listener)
	if err != nil {
		return "", err
	}

	// get target proxy
	if listener.Spec.Protocol == "HTTP" {
		_, err = e.sdkWrapper.GetTargetHTTPProxies(e.project, listener.Name)
		if err != nil {
			if IsNotFound(err) {
				err := e.sdkWrapper.CreateTargetHTTPProxy(e.project, &compute.TargetHttpProxy{Name: listener.Name, UrlMap: "global/urlMaps/" + urlMapName})
				if err != nil {
					blog.Errorf("CreateTargetHTTPProxy %s failed, %s", listener.Name, err.Error())
					return "", err
				}
				return listener.Name, nil
			}
			blog.Errorf("GetTargetHTTPProxies %s failed, %s", listener.Name, err.Error())
			return "", err
		}
		return listener.Name, nil
	}
	if listener.Spec.Protocol == "HTTPS" {
		if listener.Spec.Certificate == nil {
			blog.Errorf("listener %s certificate is empty", listener.Name)
			return "", fmt.Errorf("listener %s certificate is empty", listener.Name)
		}
		_, err = e.sdkWrapper.GetTargetHTTPSProxies(e.project, listener.Name)
		if err != nil {
			if IsNotFound(err) {
				err := e.sdkWrapper.CreateTargetHTTPSProxy(e.project, &compute.TargetHttpsProxy{Name: listener.Name,
					UrlMap:          "global/urlMaps/" + urlMapName,
					SslCertificates: []string{"global/sslCertificates/" + listener.Spec.Certificate.CertID}})
				if err != nil {
					blog.Errorf("CreateTargetHTTPSProxy %s failed, %s", listener.Name, err.Error())
					return "", err
				}
				return listener.Name, nil
			}
			blog.Errorf("GetTargetHTTPSProxies %s failed, %s", listener.Name, err.Error())
			return "", err
		}
		return listener.Name, nil
	}
	return "", fmt.Errorf("not support protocol %s", listener.Spec.Protocol)
}

func (e *GCLB) ensureForwardingRules(listener *networkextensionv1.Listener) (string, error) {
	targetProxyName, err := e.ensureTargetProxy(listener)
	if err != nil {
		return "", err
	}

	// get forwarding rules
	fr, err := e.sdkWrapper.GetForwardingRules(e.project, listener.Name)
	if err != nil {
		if IsNotFound(err) {
			// create forwarding rules
			err := e.sdkWrapper.CreateForwardingRules(e.project, listener.Name, targetProxyName,
				listener.Spec.LoadbalancerID, listener.Spec.Port)
			if err != nil {
				return "", err
			}
			return listener.Name, nil
		}
		return "", err
	}
	return fr.Name, nil
}

func (e *GCLB) ensureL7Rules(rules []networkextensionv1.ListenerRule, protocol, urlMapName string, port int) error {
	// ensure rule
	urlMap := &compute.UrlMap{}
	for i, rule := range rules {
		// ensure health check
		if err := e.ensureL7HealthCheck(rule, port); err != nil {
			return err
		}
		// ensure backend-service
		bsName, err := e.ensureL7BackendService(rule, protocol, port)
		if err != nil {
			return err
		}
		// append url map rule
		urlMap.HostRules = append(urlMap.HostRules, &compute.HostRule{
			Hosts: []string{rule.Domain}, PathMatcher: fmt.Sprintf("path-matcher-%d", i),
		})
		urlMap.PathMatchers = append(urlMap.PathMatchers, &compute.PathMatcher{
			Name:           fmt.Sprintf("path-matcher-%d", i),
			DefaultService: "global/backendServices/" + bsName,
			PathRules: []*compute.PathRule{
				{Paths: []string{rule.Path}, Service: "global/backendServices/" + bsName},
			},
		})
		// ensure host and path
		maxRate := 1000
		if rule.ListenerAttribute != nil && rule.ListenerAttribute.MaxRate != 0 {
			maxRate = rule.ListenerAttribute.MaxRate
		}
		if err := e.ensureNEGs(rule.TargetGroup, bsName, maxRate); err != nil {
			return err
		}
	}
	// patch url maps
	err := e.sdkWrapper.PatchURLMaps(e.project, urlMapName, urlMap)
	if err != nil {
		blog.Errorf("PatchURLMaps %s failed, %s", urlMapName, err.Error())
		return err
	}
	return nil
}

// md5(domain+path)
func getRuleName(domain, path string, port int) string {
	return fmt.Sprintf("a%x", md5.Sum([]byte(fmt.Sprintf("%s:%d%s", domain, port, path))))
}

// md5(domain+path)-zoneName
func getNEGName(backendServiceName, zone string) string {
	return fmt.Sprintf("%s-%s", backendServiceName, zone)
}

// ensureL7HealthCheck create health check
func (e *GCLB) ensureL7HealthCheck(rule networkextensionv1.ListenerRule, port int) error {
	_, err := e.sdkWrapper.GetHealthChecks(e.project, getRuleName(rule.Domain, rule.Path, port))
	if err != nil {
		if IsNotFound(err) {
			if err := e.sdkWrapper.CreateHealthChecks(e.project, e.generateHealthCheck(rule, port)); err != nil {
				blog.Errorf("CreateHealthChecks failed, err: %s", err.Error())
				return err
			}
			return nil
		}
		blog.Errorf("GetHealthChecks failed, %s", err.Error())
		return err
	}

	if err := e.sdkWrapper.UpdateHealthChecks(e.project, e.generateHealthCheck(rule, port)); err != nil {
		blog.Errorf("UpdateHealthChecks failed, err: %s", err.Error())
		return err
	}
	return nil
}

// generateHealthCheck generate health check for lb listener rule
func (e *GCLB) generateHealthCheck(rule networkextensionv1.ListenerRule, port int) *compute.HealthCheck {
	hc := &compute.HealthCheck{
		Name: getRuleName(rule.Domain, rule.Path, port),
	}
	if rule.ListenerAttribute == nil || rule.ListenerAttribute.HealthCheck == nil {
		hc.HttpHealthCheck = &compute.HTTPHealthCheck{Port: 80}
		hc.Type = "HTTP"
	}
	if rule.ListenerAttribute != nil && rule.ListenerAttribute.HealthCheck != nil {
		if rule.ListenerAttribute.HealthCheck.IntervalTime != 0 {
			hc.CheckIntervalSec = int64(rule.ListenerAttribute.HealthCheck.IntervalTime)
		}
		if rule.ListenerAttribute.HealthCheck.HealthNum != 0 {
			hc.HealthyThreshold = int64(rule.ListenerAttribute.HealthCheck.HealthNum)
		}
		if rule.ListenerAttribute.HealthCheck.UnHealthNum != 0 {
			hc.UnhealthyThreshold = int64(rule.ListenerAttribute.HealthCheck.UnHealthNum)
		}
		if rule.ListenerAttribute.HealthCheck.Timeout != 0 {
			hc.TimeoutSec = int64(rule.ListenerAttribute.HealthCheck.Timeout)
		}
		if rule.ListenerAttribute.HealthCheck.HealthCheckProtocol != "" {
			hc.Type = rule.ListenerAttribute.HealthCheck.HealthCheckProtocol
		}
		switch hc.Type {
		case "HTTP":
			hc.HttpHealthCheck = &compute.HTTPHealthCheck{
				Port:        int64(rule.ListenerAttribute.HealthCheck.HealthCheckPort),
				RequestPath: rule.ListenerAttribute.HealthCheck.HTTPCheckPath,
			}
		case "HTTPS":
			hc.HttpsHealthCheck = &compute.HTTPSHealthCheck{
				Port:        int64(rule.ListenerAttribute.HealthCheck.HealthCheckPort),
				RequestPath: rule.ListenerAttribute.HealthCheck.HTTPCheckPath,
			}
		case "TCP":
			hc.TcpHealthCheck = &compute.TCPHealthCheck{
				Port: int64(rule.ListenerAttribute.HealthCheck.HealthCheckPort),
			}
		}
	}
	return hc
}

func (e *GCLB) ensureL7BackendService(rule networkextensionv1.ListenerRule, protocol string, port int) (string, error) {
	// ensure backend-service
	name := getRuleName(rule.Domain, rule.Path, port)
	_, err := e.sdkWrapper.GetBackendServices(e.project, name)
	if err != nil {
		if IsNotFound(err) {
			bs := &compute.BackendService{
				Name:                name,
				LoadBalancingScheme: "EXTERNAL",
				Protocol:            protocol,
				HealthChecks:        []string{"global/healthChecks/" + name},
			}
			if rule.ListenerAttribute != nil && rule.ListenerAttribute.BackendInsecure {
				bs.Protocol = "HTTP"
			}
			err := e.sdkWrapper.CreateBackendService(e.project, bs)
			if err != nil {
				blog.Errorf("CreateBackendService failed, err: %s", err.Error())
				return "", err
			}
		} else {
			blog.Errorf("GetBackendServices failed, err: %s", err.Error())
			return "", err
		}
	}
	return name, nil
}

// ensure target group corresponding to neg
func (e *GCLB) ensureNEGs(targetGroup *networkextensionv1.ListenerTargetGroup, backendServiceName string, maxRate int) error {
	// ensure neg in every zone, if not exist, create it
	zones, err := e.sdkWrapper.ListComputeZones(e.project)
	if err != nil {
		blog.Errorf("ListComputeZones failed, err: %s", err.Error())
		return err
	}
	existNegs, err := e.sdkWrapper.ListNetworkEndpointGroups(e.project)
	if err != nil {
		blog.Errorf("ListNetworkEndpointGroups failed, err: %s", err.Error())
		return err
	}
	existNegsMap := make(map[string]*compute.NetworkEndpointGroup, 0)
	for i := range existNegs {
		existNegsMap[existNegs[i].Name] = existNegs[i]
	}
	if targetGroup == nil || targetGroup.Backends == nil {
		err = e.sdkWrapper.PatchBackendService(e.project, backendServiceName, &compute.BackendService{
			Backends: []*compute.Backend{}})
		if err != nil {
			blog.Errorf("PatchBackendService failed, err: %s", err.Error())
			return err
		}
		return nil
	}

	// get instance network and subnetwork
	network, subnetwork, zone, err := e.getNetworkAndSubnetwork(targetGroup.Backends[0].IP)
	if err != nil {
		return err
	}

	// create negs
	newNegs := e.generateNEGs(zones, zone, backendServiceName)
	for _, v := range newNegs {
		if _, ok := existNegsMap[v.Name]; !ok {
			err := e.sdkWrapper.CreateNetworkEndpointGroups(e.project, v.Zone, v.Name, network, subnetwork)
			if err != nil {
				return err
			}
		}
	}

	// ensure backend-service's backend
	backends := make([]*compute.Backend, 0)
	for _, v := range newNegs {
		backends = append(backends, &compute.Backend{
			BalancingMode:      "RATE",
			MaxRatePerEndpoint: float64(maxRate),
			Group:              fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/networkEndpointGroups/%s", e.project, v.Zone, v.Name),
		})
	}
	err = e.sdkWrapper.PatchBackendService(e.project, backendServiceName, &compute.BackendService{Backends: backends})
	if err != nil {
		blog.Errorf("PatchBackendService failed, err: %s", err.Error())
		return err
	}

	// ensure neg endpoint
	if err := e.ensureEndpoints(newNegs, targetGroup.Backends); err != nil {
		return err
	}
	return nil
}

type endpoint struct {
	IP           string
	Port         int
	InstanceName string
	Zone         string
}

func (e *GCLB) ensureEndpoints(negs []*compute.NetworkEndpointGroup, backends []networkextensionv1.ListenerBackend) error {
	endpoints := make([]endpoint, 0)
	for _, backend := range backends {
		endpoints = append(endpoints, endpoint{IP: backend.IP, Port: backend.Port})
	}

	// list instances
	instances, err := e.sdkWrapper.ListInstances(e.project)
	if err != nil {
		blog.Errorf("ListInstances failed, err: %s", err.Error())
		return err
	}

	// list pods
	pods := &k8scorev1.PodList{}
	if err := e.client.List(context.TODO(), pods, &client.ListOptions{}); err != nil {
		return fmt.Errorf("list pods failed, err: %v", err)
	}
	exceptedEndpoints := make(map[string][]*compute.NetworkEndpoint, 0)
	// fill instance name
	for i, endpoint := range endpoints {
		for _, pod := range pods.Items {
			if pod.Status.PodIP == endpoint.IP || pod.Status.HostIP == endpoint.IP {
				endpoints[i].InstanceName = pod.Spec.NodeName
				break
			}
		}
	}
	// fill zone
	for i, endpoint := range endpoints {
		for _, items := range instances.Items {
			for _, in := range items.Instances {
				if in.Name == endpoint.InstanceName {
					zoneStr := strings.Split(in.Zone, "/")
					zone := zoneStr[len(zoneStr)-1]
					endpoints[i].Zone = zone
					break
				}
			}
		}
	}

	// generate excepted endpoints
	for _, endpoint := range endpoints {
		if endpoint.InstanceName == "" || endpoint.Zone == "" {
			blog.Warnf("can't find endpoint %s:%s's instance", endpoint.IP, endpoint.Port)
			continue
		}
		exceptedEndpoints[endpoint.Zone] = append(exceptedEndpoints[endpoint.Zone], &compute.NetworkEndpoint{
			Instance:  endpoint.InstanceName,
			IpAddress: endpoint.IP,
			Port:      int64(endpoint.Port),
		})
	}

	// ensure neg endpoint
	for _, neg := range negs {
		if err := e.ensureZoneEndpoints(neg.Zone, neg.Name, exceptedEndpoints[neg.Zone]); err != nil {
			return err
		}
	}
	return nil
}

// ensureZoneEndpoints ensure neg endpoints in zone
func (e *GCLB) ensureZoneEndpoints(zone, neg string, endpoints []*compute.NetworkEndpoint) error {
	existEndpoints, err := e.sdkWrapper.ListNetworkEndpoints(e.project, zone, neg)
	if err != nil {
		blog.Errorf("ListNetworkEndpoints failed, err: %s", err.Error())
		return err
	}

	// diff
	del := make([]*compute.NetworkEndpoint, 0)
	for _, exist := range existEndpoints.Items {
		found := false
		for _, expected := range endpoints {
			if exist.NetworkEndpoint != nil && exist.NetworkEndpoint.IpAddress == expected.IpAddress &&
				exist.NetworkEndpoint.Port == expected.Port {
				found = true
				break
			}
		}
		if !found && exist.NetworkEndpoint != nil {
			del = append(del, &compute.NetworkEndpoint{Instance: exist.NetworkEndpoint.Instance,
				IpAddress: exist.NetworkEndpoint.IpAddress,
				Port:      exist.NetworkEndpoint.Port})
		}
	}
	add := make([]*compute.NetworkEndpoint, 0)
	for _, expected := range endpoints {
		found := false
		for _, exist := range existEndpoints.Items {
			if exist.NetworkEndpoint != nil && exist.NetworkEndpoint.IpAddress == expected.IpAddress &&
				exist.NetworkEndpoint.Port == expected.Port {
				found = true
				break
			}
		}
		if !found {
			add = append(add, expected)
		}
	}

	// attach endpoints
	if len(add) > 0 {
		if err := e.sdkWrapper.AttachNetworkEndpoints(e.project, zone, neg, add); err != nil {
			blog.Errorf("AttachNetworkEndpoints failed, err: %s", err.Error())
			return err
		}
	}
	// detach endpoints
	if len(del) > 0 {
		if err := e.sdkWrapper.DetachNetworkEndpoints(e.project, zone, neg, del); err != nil {
			blog.Errorf("DetachNetworkEndpoints failed, err: %s", err.Error())
			return err
		}
	}
	return nil
}

func (e *GCLB) generateNEGs(zones []*compute.Zone, fullZone, backendServiceName string) []*compute.NetworkEndpointGroup {
	zoneStrs := strings.Split(fullZone, "/")
	zone := zoneStrs[len(zoneStrs)-1]
	region := zone[:len(zone)-2]
	negs := make([]*compute.NetworkEndpointGroup, 0)
	for _, zone := range zones {
		// zone region is full url, split it
		zoneRegionURL := strings.Split(zone.Region, "/")
		if zoneRegionURL[len(zoneRegionURL)-1] != region {
			continue
		}
		negName := getNEGName(backendServiceName, zone.Name)
		negs = append(negs, &compute.NetworkEndpointGroup{
			Name: negName,
			Zone: zone.Name,
		})
	}
	return negs
}

func (e *GCLB) getNetworkAndSubnetwork(targetIP string) (string, string, string, error) {
	pods := &k8scorev1.PodList{}
	if err := e.client.List(context.TODO(), pods, &client.ListOptions{}); err != nil {
		return "", "", "", fmt.Errorf("list pods failed, err: %v", err)
	}

	nodeName := ""
	nodeIP := ""
	for _, pod := range pods.Items {
		if pod.Status.PodIP == targetIP {
			nodeName = pod.Spec.NodeName
			nodeIP = pod.Status.HostIP
		}
		if pod.Status.HostIP == targetIP {
			nodeName = pod.Spec.NodeName
			nodeIP = pod.Status.HostIP
		}
	}
	if nodeName == "" || nodeIP == "" {
		return "", "", "", errors.New("can't find pod by targetIP")
	}
	instances, err := e.sdkWrapper.ListInstances(e.project)
	if err != nil {
		blog.Errorf("ListInstances failed, err: %s", err.Error())
		return "", "", "", err
	}
	for _, v := range instances.Items {
		for _, in := range v.Instances {
			if len(in.NetworkInterfaces) <= 0 {
				continue
			}
			if in.Name == nodeName || in.NetworkInterfaces[0].NetworkIP == nodeIP {
				return in.NetworkInterfaces[0].Network, in.NetworkInterfaces[0].Subnetwork,
					in.Zone, nil
			}
		}
	}
	return "", "", "", errors.New("can't find instance info")
}

func (e *GCLB) deleteL4Listener(listener *networkextensionv1.Listener) error {
	// delete service
	service := &k8scorev1.Service{}
	objectKey := types.NamespacedName{Namespace: listener.Namespace, Name: listener.Name}
	if err := e.client.Get(context.TODO(), objectKey, service); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		blog.Errorf("get service %s/%s failed, err %s", objectKey.Namespace, objectKey.Name, err.Error())
		return err
	}
	if err := e.client.Delete(context.TODO(), &k8scorev1.Service{ObjectMeta: metav1.ObjectMeta{
		Name: listener.Name, Namespace: listener.Namespace}}); err != nil {
		blog.Errorf("delete listener service %s/%s failed, err %s", objectKey.Namespace, objectKey.Name, err.Error())
		return err
	}

	// delete endpoints
	ep := &k8scorev1.Endpoints{}
	if err := e.client.Get(context.TODO(), objectKey, ep); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		blog.Errorf("get endpoints %s/%s failed, err %s", objectKey.Namespace, objectKey.Name, err.Error())
		return err
	}
	if err := e.client.Delete(context.TODO(), &k8scorev1.Endpoints{ObjectMeta: metav1.ObjectMeta{
		Name: listener.Name, Namespace: listener.Namespace}}); err != nil {
		blog.Errorf("delete endpoints %s/%s failed, err %s", objectKey.Namespace, objectKey.Name, err.Error())
		return err
	}
	return nil
}

func (e *GCLB) deleteL7Listener(listener *networkextensionv1.Listener) error {
	// delete forwarding-rules
	if err := e.sdkWrapper.DeleteForwardingRules(e.project, listener.Name); err != nil && !IsNotFound(err) {
		blog.Errorf("DeleteForwardingRules failed, err: %s", err.Error())
		return err
	}

	// delete target http(s) proxy
	if listener.Spec.Protocol == ProtocolHTTP {
		if err := e.sdkWrapper.DeleteTargetHTTPProxy(e.project, listener.Name); err != nil && !IsNotFound(err) {
			blog.Errorf("DeleteTargetHTTPProxy failed, err: %s", err.Error())
			return err
		}
	}
	if listener.Spec.Protocol == ProtocolHTTPS {
		if err := e.sdkWrapper.DeleteTargetHTTPSProxy(e.project, listener.Name); err != nil && !IsNotFound(err) {
			blog.Errorf("DeleteTargetHTTPSProxy failed, err: %s", err.Error())
			return err
		}
	}

	// delete url-maps
	if err := e.sdkWrapper.DeleteURLMaps(e.project, listener.Name); err != nil && !IsNotFound(err) {
		blog.Errorf("DeleteURLMaps failed, err: %s", err.Error())
		return err
	}

	for _, rule := range listener.Spec.Rules {
		bsName := getRuleName(rule.Domain, rule.Path, listener.Spec.Port)
		defaultBsName := getRuleName(listener.Name, "", listener.Spec.Port)
		// delete default health check
		if err := e.sdkWrapper.DeleteHealthCheck(e.project, defaultBsName); err != nil && !IsNotFound(err) {
			blog.Errorf("DeleteHealthCheck failed, err: %s", err.Error())
			return err
		}
		// delete default backend service
		if err := e.sdkWrapper.DeleteBackendService(e.project, defaultBsName); err != nil && !IsNotFound(err) {
			blog.Errorf("DeleteBackendService failed, err: %s", err.Error())
			return err
		}
		// delete health-checks
		if err := e.sdkWrapper.DeleteHealthCheck(e.project, bsName); err != nil && !IsNotFound(err) {
			blog.Errorf("DeleteHealthCheck failed, err: %s", err.Error())
			return err
		}
		// get backend service
		bs, err := e.sdkWrapper.GetBackendServices(e.project, bsName)
		if err != nil {
			if IsNotFound(err) {
				continue
			}
			blog.Errorf("GetBackendServices failed, err: %s", err.Error())
			return err
		}
		// delete backend-services
		if err := e.sdkWrapper.DeleteBackendService(e.project, bsName); err != nil && !IsNotFound(err) {
			blog.Errorf("DeleteBackendService failed, err: %s", err.Error())
			return err
		}
		// delete negs
		for _, v := range bs.Backends {
			// group fully-qualified URL
			// https://www.googleapis.com/compute/v1/projects/project-demo/zones/us-west4-b/networkEndpointGroups/neg-demo-1
			group := strings.Replace(v.Group, "https://www.googleapis.com/compute/v1/", "", 1)
			groupStrs := strings.Split(group, "/")
			if len(groupStrs) != 6 {
				blog.Warnf("get group failed, group: %s", group)
				continue
			}
			zone := groupStrs[3]
			negName := groupStrs[5]
			e.sdkWrapper.DeleteNetworkEndpointGroups(e.project, zone, negName)
		}
	}
	return nil
}
