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

package processor

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"bk-bcs/bcs-common/common/blog"

	ingressType "bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/clb/v1"
	cloudListenerType "bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/network/v1"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/clbingress"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/cloudlb"
	qcloudlb "bk-bcs/bcs-services/bcs-clb-controller/pkg/cloudlb/qcloud"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/common"
	listenerclient "bk-bcs/bcs-services/bcs-clb-controller/pkg/listenerclient"
	svcclient "bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// update duration histogram for updater.Update function
	updateTotalDurationMetric = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "clb",
		Subsystem: "updater",
		Name:      "update_duration_total_seconds",
		Help:      "update duration a update loop",
		Buckets:   prometheus.LinearBuckets(0, 10, 10),
	})
	// listener operation duration histogram vector
	listenerDurationMetric = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "clb",
		Subsystem: "updater",
		Name:      "listener_op_duration_seconds",
		Help:      "clb listener operation duration",
		Buckets:   prometheus.LinearBuckets(0, 3, 10),
	}, []string{"action"})
	// listener operation error counter to cloud api
	listenerErrorsMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "clb",
		Subsystem: "updater",
		Name:      "listener_op_cloud_errors",
		Help:      "clb listener operation errors",
	}, []string{"action"})
	// listener operation error counter to kube apiserver
	listenerApiserverErrorsMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "clb",
		Subsystem: "updater",
		Name:      "listener_op_apiserver_errors",
		Help:      "clb listener apiserver operation errors",
	}, []string{"action"})
	// listener add operation counter
	clbListenersAddMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "clb",
		Subsystem: "updater",
		Name:      "add_listeners",
		Help:      "added listener number",
	}, []string{"name"})
	// listener update operation counter
	clbListenersUpdateMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "clb",
		Subsystem: "updater",
		Name:      "update_listeners",
		Help:      "updated listener number",
	}, []string{"name"})
	// listener delete operation counter
	clbListenersDeleteMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "clb",
		Subsystem: "updater",
		Name:      "delete_listeners",
		Help:      "deleted listener number",
	}, []string{"name"})
)

func init() {
	prometheus.Register(updateTotalDurationMetric)
	prometheus.Register(listenerDurationMetric)
	prometheus.Register(listenerErrorsMetric)
	prometheus.Register(listenerApiserverErrorsMetric)
	prometheus.Register(clbListenersAddMetric)
	prometheus.Register(clbListenersUpdateMetric)
	prometheus.Register(clbListenersDeleteMetric)
}

// Updater generate listeners from ingress and service discovery
type Updater struct {
	// Processor options
	opt    *Option
	lbInfo *cloudListenerType.CloudLoadBalancer //loadbalance info for cloud-lb
	// interface to operate cloud lb instance
	cloudlbCtl cloudlb.Interface
	// service client for service discovery
	serviceClient svcclient.Client
	// ingress client for ingress discovery
	ingressRegistry clbingress.Registry
	// client for save Intermediate cloud listener info
	listenerClient listenerclient.Interface
}

// NewUpdater create updater
func NewUpdater(opt *Option, svcClient svcclient.Client, ingressRegistry clbingress.Registry, listenerClient listenerclient.Interface) (*Updater, error) {
	lbInfo := &cloudListenerType.CloudLoadBalancer{
		ID:          "",
		Name:        opt.ClbName,
		NetworkType: opt.NetType,
	}
	cloudlbCtl, err := qcloudlb.NewClient(lbInfo)
	if err != nil {
		return nil, fmt.Errorf("create cloudlb control client with lbInfo %v failed, err %s", lbInfo, err.Error())
	}
	// load cloud lb config from env, these configs are particular
	err = cloudlbCtl.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("cloudlb control client load config failed, err %s", err.Error())
	}
	return &Updater{
		opt:             opt,
		lbInfo:          lbInfo,
		cloudlbCtl:      cloudlbCtl,
		serviceClient:   svcClient,
		ingressRegistry: ingressRegistry,
		listenerClient:  listenerClient,
	}, nil
}

// EnsureLoadBalancer create new loadbalance cloud lb instance or take over existed cloud lb instance
func (updater *Updater) EnsureLoadBalancer() error {
	lbInfoDesc, err := updater.cloudlbCtl.DescribeLoadbalance(updater.lbInfo.Name)
	if err != nil {
		blog.Errorf("describe loadbalance with name %s failed, err %s", updater.lbInfo.Name, err.Error())
		return fmt.Errorf("describe loadbalance with name %s failed, err %s", updater.lbInfo.Name, err.Error())
	}
	if lbInfoDesc == nil {
		blog.Infof("clb with name %s does not exists, try to create one", updater.lbInfo.Name)
		lbInfoCreated, err := updater.cloudlbCtl.CreateLoadbalance()
		if err != nil {
			blog.Errorf("create loadbalance with info %v failed, err %s", updater.lbInfo, err.Error())
			return fmt.Errorf("create loadbalance with info %v failed, err %s", updater.lbInfo, err.Error())
		}
		updater.lbInfo.ID = lbInfoCreated.ID
		return nil
	}
	// exit when there is existed loadbalance with same name but different net type
	if updater.lbInfo.NetworkType != lbInfoDesc.NetworkType {
		blog.Errorf("clb with name %s is already exist, want %s but with invalid network type %s", updater.lbInfo.Name, updater.lbInfo.NetworkType, lbInfoDesc.NetworkType)
		return fmt.Errorf("clb with name %s is already exist, want %s but with invalid network type %s", updater.lbInfo.Name, updater.lbInfo.NetworkType, lbInfoDesc.NetworkType)
	}
	updater.lbInfo.ID = lbInfoDesc.ID
	updater.lbInfo.PublicIPs = lbInfoDesc.PublicIPs
	updater.lbInfo.VIPS = lbInfoDesc.VIPS
	return nil
}

// ListRemoteListener list remote listener
// currently it calls cloud api directly
// TODO: cloud lb interface should has events informer
// this function should get remote listeners from local cache
func (updater *Updater) ListRemoteListener() ([]*cloudListenerType.CloudListener, error) {
	return updater.cloudlbCtl.ListListeners()
}

// Update sync update to clb
func (updater *Updater) Update() error {
	// get all associated ingresses
	updateTotalTimer := prometheus.NewTimer(updateTotalDurationMetric)
	ingressList, err := updater.ingressRegistry.ListIngresses()
	if err != nil {
		blog.Errorf("list all ingress failed, err %s", err.Error())
		return fmt.Errorf("list all ingress failed, err %s", err.Error())
	}
	// validate all ingresses
	isValid := updater.validateClbIngress(ingressList)
	if !isValid {
		blog.Errorf("validate clb ingress failed")
		return fmt.Errorf("validate clb ingress failed")
	}
	// generate cloud listener objects from ingresses
	listenerList, err := updater.generateCloudListeners(ingressList)
	if err != nil {
		blog.Errorf("generate listeners failed, err %s", err.Error())
		return fmt.Errorf("generate listeners failed, err %s", err.Error())
	}
	// get cloud listener objects from local cache
	oldList, err := updater.getCloudListenerFromCache()
	if err != nil {
		blog.Errorf("get listeners from cache failed, err %s", err.Error())
		return fmt.Errorf("get listeners from cache failed, err %s", err.Error())
	}
	// get different objects between new objects and old objects
	toDeleteListeners, toAddListeners := updater.getDiffCloudListeners(oldList, listenerList)

	// delete listeners
	for _, d := range toDeleteListeners {
		dtimer := prometheus.NewTimer(listenerDurationMetric.With(prometheus.Labels{"action": "delete"}))
		err := updater.cloudlbCtl.Delete(d)
		if err != nil {
			listenerErrorsMetric.With(prometheus.Labels{"action": "delete"}).Inc()
			blog.Errorf("failed to delete listener %v from cloud, err %s", d, err.Error())
			continue
		}
		err = updater.listenerClient.Delete(d)
		if err != nil {
			listenerApiserverErrorsMetric.With(prometheus.Labels{"action": "delete"}).Inc()
			blog.Errorf("failed to delete listener %v to k8s apiserver, err %s", d, err.Error())
			continue
		}
		dtimer.ObserveDuration()
		clbListenersDeleteMetric.WithLabelValues(d.GetName()).Inc()
	}
	// add listeners
	for _, a := range toAddListeners {
		atimer := prometheus.NewTimer(listenerDurationMetric.With(prometheus.Labels{"action": "add"}))
		err := updater.cloudlbCtl.Add(a)
		if err != nil {
			listenerErrorsMetric.With(prometheus.Labels{"action": "add"}).Inc()
			blog.Errorf("failed add listener %v to cloud, err %s", a, err.Error())
			continue
		}
		err = updater.listenerClient.Create(a)
		if err != nil {
			listenerApiserverErrorsMetric.With(prometheus.Labels{"action": "add"}).Inc()
			blog.Errorf("failed to add listener %v to k8s apiserver, err %s", a, err.Error())
			continue
		}
		atimer.ObserveDuration()
		clbListenersAddMetric.WithLabelValues(a.GetName()).Inc()
	}
	// get all listeners that shoud be updated
	updatesOld, updatesNew := updater.getUpdateCloudListeners(oldList, listenerList)

	// do update
	for index, u := range updatesNew {
		utimer := prometheus.NewTimer(listenerDurationMetric.With(prometheus.Labels{"action": "update"}))
		err := updater.cloudlbCtl.Update(updatesOld[index], u)
		if err != nil {
			listenerErrorsMetric.With(prometheus.Labels{"action": "update"}).Inc()
			blog.Errorf("failed to update listener %v to cloud, err %s", u, err.Error())
			continue
		}
		err = updater.listenerClient.Update(u)
		if err != nil {
			listenerApiserverErrorsMetric.With(prometheus.Labels{"action": "update"}).Inc()
			blog.Errorf("failed to update listener %v to k8s apiserver, err %s", u, err.Error())
			continue
		}
		utimer.ObserveDuration()
		clbListenersUpdateMetric.WithLabelValues(u.GetName()).Inc()
	}
	if len(toDeleteListeners) != 0 || len(toAddListeners) != 0 || len(updatesNew) != 0 {
		updateTotalTimer.ObserveDuration()
	}
	return nil
}

// if ok, return nil
// check port conflicts
func (updater *Updater) validateFourLayerRuleConflict(
	rule *ingressType.ClbRule, fourLayerMap map[int]*ingressType.ClbRule,
	sevenLayerMap map[int]map[string]*ingressType.ClbHttpRule) error {
	if conflictRule, ok := fourLayerMap[rule.ClbPort]; ok {
		blog.Errorf("rule %s has conflict port %d conflict with rule %s", rule.ToString(), rule.ClbPort, conflictRule.ToString())
		return fmt.Errorf("rule %s has conflict port %d conflict with rule %s", rule.ToString(), rule.ClbPort, conflictRule.ToString())
	}
	if conflictRuleMap, ok := sevenLayerMap[rule.ClbPort]; ok && len(conflictRuleMap) != 0 {
		blog.Errorf("rule %s has conflict port %d with http rule", rule.ToString(), rule.ClbPort)
		return fmt.Errorf("rule %s has conflict port %d with http rule", rule.ToString(), rule.ClbPort)
	}
	fourLayerMap[rule.ClbPort] = rule
	return nil
}

// if ok, return nil
// check ports conflicts and (domain, url) conflicts
func (updater *Updater) validateSevenLayerRuleConflict(
	rule *ingressType.ClbHttpRule, fourLayerMap map[int]*ingressType.ClbRule,
	sevenLayerMap map[int]map[string]*ingressType.ClbHttpRule) error {
	if conflictRule, ok := fourLayerMap[rule.ClbPort]; ok {
		blog.Errorf("rule %s has conflict port %d conflict with rule %s", rule.ToString(), rule.ClbPort, conflictRule.ToString())
		return fmt.Errorf("rule %s has conflict port %d conflict with rule %s", rule.ToString(), rule.ClbPort, conflictRule.ToString())
	}
	httpRuleMap, ok := sevenLayerMap[rule.ClbPort]
	if ok {
		if httpRule, isExisted := httpRuleMap[rule.Host+rule.Path]; isExisted {
			blog.Errorf("rule %s has conflict host %s and url %s with rule %s", rule.ToString(), rule.Host, rule.Path, httpRule.ToString())
			return fmt.Errorf("rule %s has conflict host %s and url %s with rule %s", rule.ToString(), rule.Host, rule.Path, httpRule.ToString())
		}
		sevenLayerMap[rule.ClbPort][rule.Host+rule.Path] = rule
		return nil
	}
	newMap := make(map[string]*ingressType.ClbHttpRule)
	newMap[rule.Host+rule.Path] = rule
	sevenLayerMap[rule.ClbPort] = newMap
	return nil
}

// validate all ingresses
func (updater *Updater) validateClbIngress(ingressList []*ingressType.ClbIngress) bool {
	fourLayerMap := make(map[int]*ingressType.ClbRule)
	sevenLayerMap := make(map[int]map[string]*ingressType.ClbHttpRule)
	// for each ingress rule
	// 1. validate all fields
	// 2. check if there is conflicts, once a rule is checked, it will be added into fourLayerMap or sevenLayerMap
	for _, tmpIngress := range ingressList {
		// tcp rules
		for _, tmpTcpRule := range tmpIngress.Spec.TCP {
			err := tmpTcpRule.Validate()
			if err != nil {
				tmpIngress.SetStatusMessage(ingressType.ClbIngressStatusAbnormal,
					fmt.Sprintf("rule %s validate failed, err %s", tmpTcpRule.ToString(), err.Error()))
				err = updater.ingressRegistry.SetIngress(tmpIngress)
				if err != nil {
					blog.Warnf("set ingress %s/%s failed, err %s", tmpIngress.GetNamespace(), tmpIngress.GetName(), err.Error())
				}
				blog.Errorf("rule %s validate failed, err %s", tmpTcpRule.ToString(), err.Error())
				return false
			}
			err = updater.validateFourLayerRuleConflict(tmpTcpRule, fourLayerMap, sevenLayerMap)
			if err != nil {
				tmpIngress.SetStatusMessage(ingressType.ClbIngressStatusAbnormal, err.Error())
				err = updater.ingressRegistry.SetIngress(tmpIngress)
				if err != nil {
					blog.Warnf("set ingress %s/%s failed, err %s", tmpIngress.GetNamespace(), tmpIngress.GetName(), err.Error())
				}
				return false
			}
		}
		// udp rules
		for _, tmpUdpRule := range tmpIngress.Spec.UDP {
			err := tmpUdpRule.Validate()
			if err != nil {
				tmpIngress.SetStatusMessage(ingressType.ClbIngressStatusAbnormal,
					fmt.Sprintf("rule %s validate failed, err %s", tmpUdpRule.ToString(), err.Error()))
				err = updater.ingressRegistry.SetIngress(tmpIngress)
				if err != nil {
					blog.Warnf("set ingress %s/%s failed, err %s", tmpIngress.GetNamespace(), tmpIngress.GetName(), err.Error())
				}
				blog.Errorf("rule %s validate failed, err %s", tmpUdpRule.ToString(), err.Error())
				return false
			}
			err = updater.validateFourLayerRuleConflict(tmpUdpRule, fourLayerMap, sevenLayerMap)
			if err != nil {
				tmpIngress.SetStatusMessage(ingressType.ClbIngressStatusAbnormal, err.Error())
				err = updater.ingressRegistry.SetIngress(tmpIngress)
				if err != nil {
					blog.Warnf("set ingress %s/%s failed, err %s", tmpIngress.GetNamespace(), tmpIngress.GetName(), err.Error())
				}
				return false
			}
		}
		// http rules
		for _, tmpHttpRule := range tmpIngress.Spec.HTTP {
			err := tmpHttpRule.ValidateHTTP()
			if err != nil {
				tmpIngress.SetStatusMessage(ingressType.ClbIngressStatusAbnormal,
					fmt.Sprintf("rule %s validate failed, err %s", tmpHttpRule.ToString(), err.Error()))
				err = updater.ingressRegistry.SetIngress(tmpIngress)
				if err != nil {
					blog.Warnf("set ingress %s/%s failed, err %s", tmpIngress.GetNamespace(), tmpIngress.GetName(), err.Error())
				}
				blog.Errorf("rule %s validate failed, err %s", tmpHttpRule.ToString(), err.Error())
				return false
			}
			err = updater.validateSevenLayerRuleConflict(tmpHttpRule, fourLayerMap, sevenLayerMap)
			if err != nil {
				tmpIngress.SetStatusMessage(ingressType.ClbIngressStatusAbnormal, err.Error())
				err = updater.ingressRegistry.SetIngress(tmpIngress)
				if err != nil {
					blog.Warnf("set ingress %s/%s failed, err %s", tmpIngress.GetNamespace(), tmpIngress.GetName(), err.Error())
				}
				return false
			}
		}
		// https rules
		for _, tmpHttpsRule := range tmpIngress.Spec.HTTPS {
			err := tmpHttpsRule.ValidateHTTPS()
			if err != nil {
				tmpIngress.SetStatusMessage(ingressType.ClbIngressStatusAbnormal,
					fmt.Sprintf("rule %s validate failed, err %s", tmpHttpsRule.ToString(), err.Error()))
				err = updater.ingressRegistry.SetIngress(tmpIngress)
				if err != nil {
					blog.Warnf("set ingress %s/%s failed, err %s", tmpIngress.GetNamespace(), tmpIngress.GetName(), err.Error())
				}
				blog.Errorf("rule %s validate failed, err %s", tmpHttpsRule.ToString(), err.Error())
				return false
			}
			err = updater.validateSevenLayerRuleConflict(tmpHttpsRule, fourLayerMap, sevenLayerMap)
			if err != nil {
				tmpIngress.SetStatusMessage(ingressType.ClbIngressStatusAbnormal, err.Error())
				err = updater.ingressRegistry.SetIngress(tmpIngress)
				if err != nil {
					blog.Warnf("set ingress %s/%s failed, err %s", tmpIngress.GetNamespace(), tmpIngress.GetName(), err.Error())
				}
				return false
			}
		}
	}
	return true
}

// getBackendListFromIngressRule get backends ip and port from ingress rules
func (updater *Updater) getBackendListFromIngressRule(rule *ingressType.ClbRule) (cloudListenerType.BackendList, error) {
	// get AppService according to service name and namespace
	appSvc, err := updater.serviceClient.GetAppService(rule.Namespace, rule.ServiceName)
	if err != nil {
		blog.Errorf("get AppService by ns %s name %s failed, err %s", rule.Namespace, rule.ServiceName, err.Error())
		return nil, fmt.Errorf("get AppService by ns %s name %s failed, err %s", rule.Namespace, rule.ServiceName, err.Error())
	}
	// find port according to port and clb rule
	var ruledSvcPort svcclient.ServicePort
	foundPort := false
	for _, svcPort := range appSvc.ServicePorts {
		if svcPort.ServicePort == rule.ServicePort {
			ruledSvcPort = svcPort
			foundPort = true
		}
	}
	if !foundPort {
		blog.Errorf("no found port %d of service %s.%s", rule.ServicePort, rule.ServiceName, rule.Namespace)
		return nil, fmt.Errorf("no found port %d of service %s.%s", rule.ServicePort, rule.ServiceName, rule.Namespace)
	}
	// check if there are no backends
	if len(appSvc.Nodes) == 0 {
		blog.Errorf("service %s.%s has no endpoint", rule.ServiceName, rule.Namespace)
		return nil, fmt.Errorf("service %s.%s has no endpoint", rule.ServiceName, rule.Namespace)
	}
	// get backends ip and port
	var backendList cloudListenerType.BackendList
	backendMap := make(map[string]*cloudListenerType.Backend)
	for _, node := range appSvc.Nodes {
		for _, port := range node.Ports {
			// svc port and node port may be associated by name port or port number
			if port.NodePort == ruledSvcPort.TargetPort || port.Name == ruledSvcPort.Name {
				var newBackend *cloudListenerType.Backend
				// for overlay ip, we use service NodeIP and service NodePort as backend ip and port
				if updater.opt.BackendIPType == common.BackendIPTypeOverlay {
					newBackend = &cloudListenerType.Backend{
						IP:     node.ProxyIP,
						Port:   ruledSvcPort.ProxyPort,
						Weight: 10,
					}
					// for underlay ip
					// use pod ip and port directly
				} else {
					newBackend = &cloudListenerType.Backend{
						IP:     node.NodeIP,
						Port:   port.NodePort,
						Weight: 10,
					}
					// support pod with mesos bridge network
					if port.ProxyPort > 0 {
						newBackend.IP = node.ProxyIP
						newBackend.Port = port.ProxyPort
					}
				}
				// to filter same ip and port, cloud lb cannot bind the same backend twice
				if _, ok := backendMap[newBackend.IP+strconv.Itoa(newBackend.Port)]; ok {
					continue
				}
				backendList = append(backendList, newBackend)
				backendMap[newBackend.IP+strconv.Itoa(newBackend.Port)] = newBackend
				break
			}
		}
	}
	return backendList, nil
}

func (updater *Updater) generate4LayerListener(layer4Rule *ingressType.ClbRule, protocol string) (*cloudListenerType.CloudListener, error) {
	// [Design]
	// creating a listener for a clb rule, even if there is neither service nor backend for the clb rule.
	// it may too much time to create a clb listener, so we can create it previously
	listener := &cloudListenerType.CloudListener{
		TypeMeta: metav1.TypeMeta{
			Kind:       "cloudlistener",
			APIVersion: cloudListenerType.SchemeGroupVersion.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			// there may be multiple clb instance for the same service,
			// add clbname into listener name to avoid listener name conflicting between multiple clb instance.
			Name:      generateCloudListenerName(updater.opt.ClbName, protocol, layer4Rule.ClbPort),
			Namespace: layer4Rule.Namespace,
			Labels: map[string]string{
				"bmsf.tencent.com/clbname": updater.opt.ClbName,
			},
		},
		Spec: cloudListenerType.CloudListenerSpec{
			LoadBalancerID: updater.lbInfo.ID,
			Protocol:       protocol,
			ListenPort:     layer4Rule.ClbPort,
			TargetGroup:    cloudListenerType.NewTargetGroup("", "", "", 0),
		},
		Status: cloudListenerType.CloudListenerStatus{
			LastUpdateTime: metav1.NewTime(time.Now()),
		},
	}
	listener.Spec.TargetGroup.SessionExpire = layer4Rule.SessionTime
	if layer4Rule.LbPolicy != nil {
		// listener.Spec.TargetGroup.LBPolicy WRR is already set
		// TODO: wrr weight can be define by pod annotations
		if layer4Rule.LbPolicy.Strategy == cloudListenerType.ClbLBPolicyLeastConn ||
			layer4Rule.LbPolicy.Strategy == cloudListenerType.ClbLBPolicyIPHash {
			listener.Spec.TargetGroup.LBPolicy = layer4Rule.LbPolicy.Strategy
		}
	}
	if layer4Rule.HealthCheck != nil {
		if layer4Rule.HealthCheck.Enabled == false {
			listener.Spec.TargetGroup.HealthCheck.Enabled = 0
		} else {
			if layer4Rule.HealthCheck.Timeout != 0 {
				listener.Spec.TargetGroup.HealthCheck.Timeout = layer4Rule.HealthCheck.Timeout
			}
			if layer4Rule.HealthCheck.IntervalTime != 0 {
				listener.Spec.TargetGroup.HealthCheck.IntervalTime = layer4Rule.HealthCheck.IntervalTime
			}
			if layer4Rule.HealthCheck.HealthNum != 0 {
				listener.Spec.TargetGroup.HealthCheck.HealthNum = layer4Rule.HealthCheck.HealthNum
			}
			if layer4Rule.HealthCheck.UnHealthNum != 0 {
				listener.Spec.TargetGroup.HealthCheck.UnHealthNum = layer4Rule.HealthCheck.UnHealthNum
			}
		}
	}
	// when get backends error, we still generate listener,
	// because creating listener on clb instance is slow.
	backendList, err := updater.getBackendListFromIngressRule(layer4Rule)
	if err != nil {
		blog.Warnf("get backend list from ingress rule %s failed, err %s", layer4Rule.ToString(), err.Error())
	}
	listener.Spec.TargetGroup.Backends = backendList
	return listener, nil
}

func (updater *Updater) generate7LayerListener(layer7Rule *ingressType.ClbHttpRule, protocol string) (*cloudListenerType.CloudListener, error) {
	// [Design]
	// creating a listener for a clb rule, even if there is neither service nor backend for the clb rule.
	// it may too much time to create a clb listener, so we can create it previously
	listener := &cloudListenerType.CloudListener{
		TypeMeta: metav1.TypeMeta{
			Kind:       "cloudlistener",
			APIVersion: cloudListenerType.SchemeGroupVersion.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      generateCloudListenerName(updater.opt.ClbName, protocol, layer7Rule.ClbRule.ClbPort),
			Namespace: layer7Rule.ClbRule.Namespace,
			Labels: map[string]string{
				"bmsf.tencent.com/clbname": updater.opt.ClbName,
			},
		},
		Spec: cloudListenerType.CloudListenerSpec{
			LoadBalancerID: updater.lbInfo.ID,
			Protocol:       protocol,
			ListenPort:     layer7Rule.ClbRule.ClbPort,
		},
		Status: cloudListenerType.CloudListenerStatus{
			LastUpdateTime: metav1.NewTime(time.Now()),
		},
	}

	rule := cloudListenerType.NewRule(layer7Rule.Host, layer7Rule.Path)
	rule.TargetGroup.SessionExpire = layer7Rule.SessionTime
	if layer7Rule.ClbRule.LbPolicy != nil {
		// listener.Spec.TargetGroup.LBPolicy WRR is already set
		// TODO: wrr weight can be define by pod annotations
		if layer7Rule.ClbRule.LbPolicy.Strategy == cloudListenerType.ClbLBPolicyLeastConn ||
			layer7Rule.ClbRule.LbPolicy.Strategy == cloudListenerType.ClbLBPolicyIPHash {
			rule.TargetGroup.LBPolicy = layer7Rule.ClbRule.LbPolicy.Strategy
		}
	}
	if layer7Rule.ClbRule.HealthCheck != nil {
		if layer7Rule.ClbRule.HealthCheck.Enabled == false {
			rule.TargetGroup.HealthCheck.Enabled = 0
		} else {
			if layer7Rule.ClbRule.HealthCheck.Timeout != 0 {
				rule.TargetGroup.HealthCheck.Timeout = layer7Rule.ClbRule.HealthCheck.Timeout
			}
			if layer7Rule.ClbRule.HealthCheck.IntervalTime != 0 {
				rule.TargetGroup.HealthCheck.IntervalTime = layer7Rule.ClbRule.HealthCheck.IntervalTime
			}
			if layer7Rule.ClbRule.HealthCheck.HealthNum != 0 {
				rule.TargetGroup.HealthCheck.HealthNum = layer7Rule.ClbRule.HealthCheck.HealthNum
			}
			if layer7Rule.ClbRule.HealthCheck.UnHealthNum != 0 {
				rule.TargetGroup.HealthCheck.UnHealthNum = layer7Rule.ClbRule.HealthCheck.UnHealthNum
			}
			if len(layer7Rule.ClbRule.HealthCheck.HTTPCheckPath) != 0 && strings.HasPrefix(layer7Rule.ClbRule.HealthCheck.HTTPCheckPath, "/") {
				rule.TargetGroup.HealthCheck.HTTPCheckPath = layer7Rule.ClbRule.HealthCheck.HTTPCheckPath
			}
			if layer7Rule.ClbRule.HealthCheck.HTTPCode != 0 {
				rule.TargetGroup.HealthCheck.HTTPCode = layer7Rule.ClbRule.HealthCheck.HTTPCode
			}
		}
	}
	// set tls config for https listener
	if protocol == cloudListenerType.ClbListenerProtocolHTTPS {
		listener.Spec.TLS = &cloudListenerType.CloudListenerTls{}
		listener.Spec.TLS.Mode = layer7Rule.TLS.Mode
		if len(layer7Rule.TLS.CertID) != 0 {
			listener.Spec.TLS.CertID = layer7Rule.TLS.CertID
		}
		if len(layer7Rule.TLS.CertServerName) != 0 {
			listener.Spec.TLS.CertServerName = layer7Rule.TLS.CertServerName
		}
		if len(layer7Rule.TLS.CertServerKey) != 0 {
			listener.Spec.TLS.CertServerKey = layer7Rule.TLS.CertServerKey
		}
		if len(layer7Rule.TLS.CertServerContent) != 0 {
			listener.Spec.TLS.CertServerContent = layer7Rule.TLS.CertServerContent
		}
		if len(layer7Rule.TLS.CertCaID) != 0 {
			listener.Spec.TLS.CertCaID = layer7Rule.TLS.CertCaID
		}
		if len(layer7Rule.TLS.CertClientCaName) != 0 {
			listener.Spec.TLS.CertClientCaName = layer7Rule.TLS.CertClientCaName
		}
		if len(layer7Rule.TLS.CertClientCaContent) != 0 {
			listener.Spec.TLS.CertClientCaContent = layer7Rule.TLS.CertClientCaContent
		}
	}
	// when get backends error, we still generate listener,
	// because creating listener on clb instance is slow.
	backendList, err := updater.getBackendListFromIngressRule(&layer7Rule.ClbRule)
	if err != nil {
		blog.Warnf("get backend List from rule %s failed, err %s", layer7Rule.ToString(), err.Error())
	}
	rule.TargetGroup.Backends = backendList
	listener.Spec.Rules = append(listener.Spec.Rules, rule)
	return listener, nil
}

// for http listener and https listener, multiple rules may listen the same port
// merge listener with same port here
func (updater *Updater) combineHttpListener(mainListener *cloudListenerType.CloudListener, secondListener *cloudListenerType.CloudListener) {
	mainListener.Spec.Rules = append(mainListener.Spec.Rules, secondListener.Spec.Rules...)
}

// [Design]
// for statfulset rule, we expose a port for each pod,
// ports number depends on the StartIndex and EndIndex in statefulSetRule
func (updater *Updater) convertStatefulSetRuleToListener(statefulSetRule *ingressType.ClbStatefulSetRule, protocol string) ([]*cloudListenerType.CloudListener, error) {
	var retListenerList []*cloudListenerType.CloudListener
	for portIndex := statefulSetRule.StartIndex; portIndex <= statefulSetRule.EndIndex; portIndex++ {
		listener := &cloudListenerType.CloudListener{
			TypeMeta: metav1.TypeMeta{
				Kind:       "cloudlistener",
				APIVersion: cloudListenerType.SchemeGroupVersion.Version,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      generateCloudListenerName(updater.opt.ClbName, protocol, statefulSetRule.StartPort+portIndex),
				Namespace: statefulSetRule.Namespace,
				Labels: map[string]string{
					"bmsf.tencent.com/clbname": updater.opt.ClbName,
				},
			},
			Spec: cloudListenerType.CloudListenerSpec{
				LoadBalancerID: updater.lbInfo.ID,
				Protocol:       protocol,
				ListenPort:     statefulSetRule.StartPort + portIndex,
				TargetGroup:    cloudListenerType.NewTargetGroup("", "", "", 0),
			},
			Status: cloudListenerType.CloudListenerStatus{
				LastUpdateTime: metav1.NewTime(time.Now()),
			},
		}
		listener.Spec.TargetGroup.SessionExpire = statefulSetRule.SessionTime
		if statefulSetRule.LbPolicy != nil {
			// listener.Spec.TargetGroup.LBPolicy WRR is already set
			// TODO: wrr weight can be define by pod annotations
			if statefulSetRule.LbPolicy.Strategy == cloudListenerType.ClbLBPolicyLeastConn ||
				statefulSetRule.LbPolicy.Strategy == cloudListenerType.ClbLBPolicyIPHash {
				listener.Spec.TargetGroup.LBPolicy = statefulSetRule.LbPolicy.Strategy
			}
		}
		if statefulSetRule.HealthCheck != nil {
			if statefulSetRule.HealthCheck.Enabled == false {
				listener.Spec.TargetGroup.HealthCheck.Enabled = 0
			} else {
				if statefulSetRule.HealthCheck.Timeout != 0 {
					listener.Spec.TargetGroup.HealthCheck.Timeout = statefulSetRule.HealthCheck.Timeout
				}
				if statefulSetRule.HealthCheck.IntervalTime != 0 {
					listener.Spec.TargetGroup.HealthCheck.IntervalTime = statefulSetRule.HealthCheck.IntervalTime
				}
				if statefulSetRule.HealthCheck.HealthNum != 0 {
					listener.Spec.TargetGroup.HealthCheck.HealthNum = statefulSetRule.HealthCheck.HealthNum
				}
				if statefulSetRule.HealthCheck.UnHealthNum != 0 {
					listener.Spec.TargetGroup.HealthCheck.UnHealthNum = statefulSetRule.HealthCheck.UnHealthNum
				}
			}
		}
		retListenerList = append(retListenerList, listener)
	}
	// get AppService from statefulset by service client according to service name and namespace in stateful set rule
	appServices, err := updater.serviceClient.ListAppServiceFromStatefulSet(statefulSetRule.Namespace, statefulSetRule.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("get app services from statefulSetRule %v, err %s", statefulSetRule, err.Error())
	}
	if len(appServices) == 0 {
		return retListenerList, nil
	}
	for portIndex, appService := range appServices {
		if len(appService.ServicePorts) != 1 {
			return nil, fmt.Errorf("service port length of stateful set appService %v is not equal to 1", appService)
		}
		// check if the service port is in certain range
		if appService.ServicePorts[0].ServicePort < statefulSetRule.StartIndex || appService.ServicePorts[0].ServicePort > statefulSetRule.EndIndex {
			blog.Infof("index %d is not in [%d, %d], skip appService %s",
				appService.ServicePorts[0].ServicePort,
				statefulSetRule.StartIndex, statefulSetRule.EndIndex, appService.GetName()+"/"+appService.GetNamespace())
			continue
		}
		listener := retListenerList[portIndex]
		if len(appService.Nodes) != 0 {
			node := appService.Nodes[0]
			newBackend := &cloudListenerType.Backend{
				IP:     node.NodeIP,
				Port:   statefulSetRule.StartPort + node.Ports[0].NodePort,
				Weight: 10,
			}
			listener.Spec.TargetGroup.Backends = append(listener.Spec.TargetGroup.Backends, newBackend)
		}
	}
	return retListenerList, nil
}

// like convertStatefulSetRuleToListener, but set extra config (health check, tls)
func (updater *Updater) convertStatefulSetHttpToListener(statefulSetRule *ingressType.ClbStatefulSetHttpRule, protocol string) ([]*cloudListenerType.CloudListener, error) {
	var retListenerList []*cloudListenerType.CloudListener
	for portIndex := statefulSetRule.StartIndex; portIndex <= statefulSetRule.EndIndex; portIndex++ {
		listener := &cloudListenerType.CloudListener{
			TypeMeta: metav1.TypeMeta{
				Kind:       "cloudlistener",
				APIVersion: cloudListenerType.SchemeGroupVersion.Version,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      generateCloudListenerName(updater.opt.ClbName, protocol, statefulSetRule.StartPort+portIndex),
				Namespace: statefulSetRule.Namespace,
				Labels: map[string]string{
					"bmsf.tencent.com/clbname": updater.opt.ClbName,
				},
			},
			Spec: cloudListenerType.CloudListenerSpec{
				LoadBalancerID: updater.lbInfo.ID,
				Protocol:       protocol,
				ListenPort:     statefulSetRule.StartPort + portIndex,
			},
			Status: cloudListenerType.CloudListenerStatus{
				LastUpdateTime: metav1.NewTime(time.Now()),
			},
		}
		rule := cloudListenerType.NewRule(statefulSetRule.Host, statefulSetRule.Path)
		rule.TargetGroup.SessionExpire = statefulSetRule.SessionTime
		if statefulSetRule.ClbRule.LbPolicy != nil {
			// listener.Spec.TargetGroup.LBPolicy WRR is already set
			// TODO: wrr weight can be define by pod annotations
			if statefulSetRule.ClbRule.LbPolicy.Strategy == cloudListenerType.ClbLBPolicyLeastConn ||
				statefulSetRule.ClbRule.LbPolicy.Strategy == cloudListenerType.ClbLBPolicyIPHash {
				rule.TargetGroup.LBPolicy = statefulSetRule.ClbRule.LbPolicy.Strategy
			}
		}
		// health check config
		if statefulSetRule.ClbRule.HealthCheck != nil {
			if statefulSetRule.ClbRule.HealthCheck.Enabled == false {
				rule.TargetGroup.HealthCheck.Enabled = 0
			} else {
				if statefulSetRule.ClbRule.HealthCheck.Timeout != 0 {
					rule.TargetGroup.HealthCheck.Timeout = statefulSetRule.ClbRule.HealthCheck.Timeout
				}
				if statefulSetRule.ClbRule.HealthCheck.IntervalTime != 0 {
					rule.TargetGroup.HealthCheck.IntervalTime = statefulSetRule.ClbRule.HealthCheck.IntervalTime
				}
				if statefulSetRule.ClbRule.HealthCheck.HealthNum != 0 {
					rule.TargetGroup.HealthCheck.HealthNum = statefulSetRule.ClbRule.HealthCheck.HealthNum
				}
				if statefulSetRule.ClbRule.HealthCheck.UnHealthNum != 0 {
					rule.TargetGroup.HealthCheck.UnHealthNum = statefulSetRule.ClbRule.HealthCheck.UnHealthNum
				}
				if len(statefulSetRule.ClbRule.HealthCheck.HTTPCheckPath) != 0 && strings.HasPrefix(statefulSetRule.ClbRule.HealthCheck.HTTPCheckPath, "/") {
					rule.TargetGroup.HealthCheck.HTTPCheckPath = statefulSetRule.ClbRule.HealthCheck.HTTPCheckPath
				}
				if statefulSetRule.ClbRule.HealthCheck.HTTPCode != 0 {
					rule.TargetGroup.HealthCheck.HTTPCode = statefulSetRule.ClbRule.HealthCheck.HTTPCode
				}
			}
		}
		// tls config
		listener.Spec.Rules = append(listener.Spec.Rules, rule)
		if protocol == cloudListenerType.ClbListenerProtocolHTTPS {
			listener.Spec.TLS = &cloudListenerType.CloudListenerTls{}
			listener.Spec.TLS.Mode = statefulSetRule.TLS.Mode
			if len(statefulSetRule.TLS.CertID) != 0 {
				listener.Spec.TLS.CertID = statefulSetRule.TLS.CertID
			}
			if len(statefulSetRule.TLS.CertServerName) != 0 {
				listener.Spec.TLS.CertServerName = statefulSetRule.TLS.CertServerName
			}
			if len(statefulSetRule.TLS.CertServerKey) != 0 {
				listener.Spec.TLS.CertServerKey = statefulSetRule.TLS.CertServerKey
			}
			if len(statefulSetRule.TLS.CertServerContent) != 0 {
				listener.Spec.TLS.CertServerContent = statefulSetRule.TLS.CertServerContent
			}
			if len(statefulSetRule.TLS.CertCaID) != 0 {
				listener.Spec.TLS.CertCaID = statefulSetRule.TLS.CertCaID
			}
			if len(statefulSetRule.TLS.CertClientCaName) != 0 {
				listener.Spec.TLS.CertClientCaName = statefulSetRule.TLS.CertClientCaName
			}
			if len(statefulSetRule.TLS.CertClientCaContent) != 0 {
				listener.Spec.TLS.CertClientCaContent = statefulSetRule.TLS.CertClientCaContent
			}
		}
		retListenerList = append(retListenerList, listener)
	}
	appServices, err := updater.serviceClient.ListAppServiceFromStatefulSet(statefulSetRule.Namespace, statefulSetRule.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("get app services from statefulSetRule %v, err %s", statefulSetRule, err.Error())
	}

	if len(appServices) == 0 {
		return retListenerList, nil
	}
	for portIndex, appService := range appServices {
		if len(appService.ServicePorts) != 1 {
			return nil, fmt.Errorf("service port length of stateful set appService %v is not equal to 1", appService)
		}
		if appService.ServicePorts[0].ServicePort < statefulSetRule.StartIndex || appService.ServicePorts[0].ServicePort > statefulSetRule.EndIndex {
			blog.Infof("index %d is not in [%d, %d], skip appService %s", appService.ServicePorts[0].ServicePort,
				statefulSetRule.StartIndex, statefulSetRule.EndIndex, appService.GetName()+"/"+appService.GetNamespace())
			continue
		}
		rule := retListenerList[portIndex].Spec.Rules[0]
		if len(appService.Nodes) != 0 {
			node := appService.Nodes[0]
			newBackend := &cloudListenerType.Backend{
				IP:     node.NodeIP,
				Port:   statefulSetRule.StartPort + node.Ports[0].NodePort,
				Weight: 10,
			}
			rule.TargetGroup.Backends = append(rule.TargetGroup.Backends, newBackend)
		}
	}
	return retListenerList, nil
}

func (updater *Updater) generateCloudListeners(ingressList []*ingressType.ClbIngress) ([]*cloudListenerType.CloudListener, error) {
	// valide ingresses
	// TODO: validation does not include statefulset ingresses
	ok := updater.validateClbIngress(ingressList)
	if !ok {
		return nil, fmt.Errorf("validate clb ingress list falied")
	}
	listenerMap := make(map[int]*cloudListenerType.CloudListener)
	for _, tmpIngress := range ingressList {
		for _, tmpTcpRule := range tmpIngress.Spec.TCP {
			listener, err := updater.generate4LayerListener(tmpTcpRule, cloudListenerType.ClbListenerProtocolTCP)
			if err != nil {
				blog.Warnf("generate tcp listener for rule %v failed, err %v, skip", tmpTcpRule, err.Error())
				continue
			}
			blog.V(5).Infof("get tcp listener: %v", listener)
			listenerMap[listener.Spec.ListenPort] = listener
		}
		for _, tmpUdpRule := range tmpIngress.Spec.UDP {
			listener, err := updater.generate4LayerListener(tmpUdpRule, cloudListenerType.ClbListenerProtocolUDP)
			if err != nil {
				blog.Warnf("generate udp listener for rule %v failed, err %v, skip", tmpUdpRule, err.Error())
				continue
			}
			blog.V(5).Infof("get udp listener: %v", listener)
			listenerMap[listener.Spec.ListenPort] = listener
		}
		for _, tmpHttpRule := range tmpIngress.Spec.HTTP {
			listener, err := updater.generate7LayerListener(tmpHttpRule, cloudListenerType.ClbListenerProtocolHTTP)
			if err != nil {
				blog.Warnf("generate http listener for rule %v failed, err %s, skip", tmpHttpRule, err.Error())
				continue
			}
			blog.V(5).Infof("get http listener: %v", listener)
			if existedListener, ok := listenerMap[listener.Spec.ListenPort]; ok {
				updater.combineHttpListener(existedListener, listener)
			} else {
				listenerMap[listener.Spec.ListenPort] = listener
			}
		}
		for _, tmpHttpsRule := range tmpIngress.Spec.HTTPS {
			listener, err := updater.generate7LayerListener(tmpHttpsRule, cloudListenerType.ClbListenerProtocolHTTPS)
			if err != nil {
				blog.Warnf("generate https listener for rule %v failed, err %s, skip", tmpHttpsRule, err.Error())
				continue
			}
			blog.V(5).Infof("get https listener: %v", listener)
			if existedListener, ok := listenerMap[listener.Spec.ListenPort]; ok {
				updater.combineHttpListener(existedListener, listener)
			} else {
				listenerMap[listener.Spec.ListenPort] = listener
			}
		}

		// generate listener for stateful set
		if tmpIngress.Spec.StatefulSet != nil {
			for _, tmpTcpRule := range tmpIngress.Spec.StatefulSet.TCP {
				listeners, err := updater.convertStatefulSetRuleToListener(tmpTcpRule, cloudListenerType.ClbListenerProtocolTCP)
				if err != nil {
					blog.Warnf("convert stateful set tcp rule failed, err %s", err.Error())
					continue
				}
				for _, listener := range listeners {
					listenerMap[listener.Spec.ListenPort] = listener
				}
			}
			for _, tmpUdpRule := range tmpIngress.Spec.StatefulSet.UDP {
				listeners, err := updater.convertStatefulSetRuleToListener(tmpUdpRule, cloudListenerType.ClbListenerProtocolUDP)
				if err != nil {
					blog.Warnf("convert stateful set udp rule failed, err %s", err.Error())
					continue
				}
				for _, listener := range listeners {
					listenerMap[listener.Spec.ListenPort] = listener
				}
			}
			for _, tmpHttpRule := range tmpIngress.Spec.StatefulSet.HTTP {
				listeners, err := updater.convertStatefulSetHttpToListener(tmpHttpRule, cloudListenerType.ClbListenerProtocolHTTP)
				if err != nil {
					blog.Warnf("convert stateful set http rule failed, err %s", err.Error())
					continue
				}
				for _, listener := range listeners {
					listenerMap[listener.Spec.ListenPort] = listener
				}
			}
			for _, tmpHttpsRule := range tmpIngress.Spec.StatefulSet.HTTPS {
				listeners, err := updater.convertStatefulSetHttpToListener(tmpHttpsRule, cloudListenerType.ClbListenerProtocolHTTPS)
				if err != nil {
					blog.Warnf("convert stateful set http rule failed, err %s", err.Error())
					continue
				}
				for _, listener := range listeners {
					listenerMap[listener.Spec.ListenPort] = listener
				}
			}
		}

	}

	var retList []*cloudListenerType.CloudListener
	for _, e := range listenerMap {
		retList = append(retList, e)
	}
	return retList, nil
}

func (updater *Updater) getCloudListenerFromCache() ([]*cloudListenerType.CloudListener, error) {
	return updater.listenerClient.ListListeners()
}

// get diff cloud listener
func (updater *Updater) getDiffCloudListeners(olds, curs []*cloudListenerType.CloudListener) (
	[]*cloudListenerType.CloudListener, []*cloudListenerType.CloudListener) {
	var dels, adds []*cloudListenerType.CloudListener
	tmpMapAdd := make(map[string]*cloudListenerType.CloudListener)
	for _, old := range olds {
		tmpMapAdd[old.Key()] = old
	}
	for _, cur := range curs {
		_, ok := tmpMapAdd[cur.Key()]
		if !ok {
			adds = append(adds, cur)
		}
	}
	tmpMapDel := make(map[string]*cloudListenerType.CloudListener)
	for _, cur := range curs {
		tmpMapDel[cur.Key()] = cur
	}
	for _, old := range olds {
		_, ok := tmpMapDel[old.Key()]
		if !ok {
			dels = append(dels, old)
		}
	}
	blog.Infof("getDiffCloudListeners adds-%d, dels-%d", len(adds), len(dels))
	for index, add := range adds {
		blog.V(3).Infof("[no.%d] add %s", index, add.ToString())
	}
	for index, del := range dels {
		blog.V(3).Infof("[no.%d] del %s", index, del.ToString())
	}
	return dels, adds
}

// get cloud listener to be updated
func (updater *Updater) getUpdateCloudListeners(olds, curs []*cloudListenerType.CloudListener) (
	[]*cloudListenerType.CloudListener, []*cloudListenerType.CloudListener) {
	var toUpdates, updates []*cloudListenerType.CloudListener
	tmpMap := make(map[string]*cloudListenerType.CloudListener)
	for _, old := range olds {
		tmpMap[old.Key()] = old
	}
	for _, cur := range curs {
		old, ok := tmpMap[cur.Key()]
		if ok {
			if !old.IsEqual(cur) {
				toUpdates = append(toUpdates, old)
				updates = append(updates, cur)
			}
		}
	}
	blog.Infof("getUpdateCloudListeners updates-%d", len(updates))
	for index, u := range updates {
		blog.V(3).Infof("[no.%d] new: %s, old: %s", index, u.ToString(), tmpMap[u.Key()].ToString())
	}
	return toUpdates, updates
}
