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
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	ingress "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/clb/v1"
	loadbalance "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/network/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient"

	"github.com/prometheus/client_golang/prometheus"
)

// metric desc for normal clb ingress rule
func newClbIngressRuleMetricDesc(clbname string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("clb", "processor", "ingressrule"),
		"clb ingress rule metric info for clb controller",
		[]string{"service", "namespace", "clbport", "protocol", "domain", "path"},
		prometheus.Labels{
			"clbname": clbname,
		},
	)
}

// metric desc for statefulset ingress rule
func newClbStatefulSetRuleMetricDesc(clbname string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("clb", "processor", "ingressrule_statefulset"),
		"stateful clb ingress rule metric info for clb controller",
		[]string{"service", "namespace", "clbport", "protocol", "domain", "path", "startIndex", "endIndex"},
		prometheus.Labels{
			"clbname": clbname,
		},
	)
}

// metric desc for appnode
func newAppNodeMetricDesc(clbname string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("clb", "processor", "app_node"),
		"app node metric info for clb controller",
		[]string{"service", "namespace", "service_port", "node_ip", "pod_ip", "port"},
		prometheus.Labels{
			"clbname": clbname,
		},
	)
}

// metric desc for listener
func newClbListenerMetricDesc(clbname string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("clb", "processor", "listener"),
		"clb listener metric info for clb controller",
		[]string{"id", "clbport", "protocol", "domain", "path"},
		prometheus.Labels{
			"clbname": clbname,
		},
	)
}

// metric desc for clb backends
func newClbBackendMetricDesc(clbname string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("clb", "processor", "backend"),
		"clb backend metric info for clb controller",
		[]string{"id", "clbport", "protocol", "ip", "port"},
		prometheus.Labels{
			"clbname": clbname,
		},
	)
}

// metric desc for remote backend
func newRemoteBackendMetricDesc(clbname string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("clb", "processor", "remote_backend"),
		"actual backend of clb listener on tencent cloud",
		[]string{"clbport", "domain", "url", "ip", "port", "status"},
		prometheus.Labels{
			"clbname": clbname,
		},
	)
}

// Describe implements promethues exporter Describe interface
func (p *Processor) Describe(ch chan<- *prometheus.Desc) {
	clbname := p.opt.ClbName
	ch <- newClbIngressRuleMetricDesc(clbname)
	ch <- newClbStatefulSetRuleMetricDesc(clbname)
	ch <- newClbListenerMetricDesc(clbname)
	ch <- newClbBackendMetricDesc(clbname)
	ch <- newAppNodeMetricDesc(clbname)
	ch <- newRemoteBackendMetricDesc(clbname)
}

// collect app node info from AppService cache
// TODO: for mesos bridge network, the metric is not suitable
func (p *Processor) collectAppService(ch chan<- prometheus.Metric, appService *serviceclient.AppService) {
	if len(appService.Nodes) != 0 {
		for _, servicePort := range appService.ServicePorts {
			for _, node := range appService.Nodes {
				for _, port := range node.Ports {
					// match port
					if servicePort.TargetPort == port.NodePort || servicePort.Name == port.Name {
						ch <- prometheus.MustNewConstMetric(
							newAppNodeMetricDesc(p.opt.ClbName),
							prometheus.GaugeValue,
							float64(1),
							[]string{
								appService.GetName(), appService.GetNamespace(),
								strconv.Itoa(servicePort.ServicePort),
								node.ProxyIP, node.NodeIP, strconv.Itoa(port.NodePort),
							}...,
						)
					}
				}
			}
		}
	}
}

// collect ingress info from clb cache
func (p *Processor) collectIngress(
	ch chan<- prometheus.Metric,
	httpArr, httpsArr []*ingress.ClbHttpRule, tcpArr, udpArr []*ingress.ClbRule) {
	// http rule
	for _, http := range httpArr {
		ch <- prometheus.MustNewConstMetric(
			newClbIngressRuleMetricDesc(p.opt.ClbName),
			prometheus.GaugeValue,
			float64(http.ClbPort),
			[]string{http.ServiceName, http.Namespace,
				strconv.Itoa(http.ClbPort), "http", http.Host, http.Path}...,
		)
		appService, err := p.serviceClient.GetAppService(http.Namespace, http.ServiceName)
		if err != nil {
			blog.Warnf("get AppService by (%s, %s) failed, err %s", http.Namespace, http.ServiceName, err.Error())
			continue
		}
		p.collectAppService(ch, appService)
	}
	// https rules
	for _, https := range httpsArr {
		ch <- prometheus.MustNewConstMetric(
			newClbIngressRuleMetricDesc(p.opt.ClbName),
			prometheus.GaugeValue,
			float64(https.ClbPort),
			[]string{https.ServiceName, https.Namespace,
				strconv.Itoa(https.ClbPort), "https", https.Host, https.Path}...,
		)
		appService, err := p.serviceClient.GetAppService(https.Namespace, https.ServiceName)
		if err != nil {
			blog.Warnf("get AppService by (%s, %s) failed, err %s", https.Namespace, https.ServiceName, err.Error())
			continue
		}
		p.collectAppService(ch, appService)
	}
	// tcp rules
	for _, tcp := range tcpArr {
		ch <- prometheus.MustNewConstMetric(
			newClbIngressRuleMetricDesc(p.opt.ClbName),
			prometheus.GaugeValue,
			float64(tcp.ClbPort),
			[]string{tcp.ServiceName, tcp.Namespace,
				strconv.Itoa(tcp.ClbPort), "tcp", "", ""}...,
		)
		appService, err := p.serviceClient.GetAppService(tcp.Namespace, tcp.ServiceName)
		if err != nil {
			blog.Warnf("get AppService by (%s, %s) failed, err %s", tcp.Namespace, tcp.ServiceName, err.Error())
			continue
		}
		p.collectAppService(ch, appService)
	}
	// udp rules
	for _, udp := range udpArr {
		ch <- prometheus.MustNewConstMetric(
			newClbIngressRuleMetricDesc(p.opt.ClbName),
			prometheus.GaugeValue,
			float64(udp.ClbPort),
			[]string{udp.ServiceName, udp.Namespace,
				strconv.Itoa(udp.ClbPort), "udp", "", ""}...,
		)
		appService, err := p.serviceClient.GetAppService(udp.Namespace, udp.ServiceName)
		if err != nil {
			blog.Warnf("get AppService by (%s, %s) failed, err %s", udp.Namespace, udp.ServiceName, err.Error())
			continue
		}
		p.collectAppService(ch, appService)
	}
}

// collectStatefulSetIngress collect ingress rule for statefulset rule
func (p *Processor) collectStatefulSetIngress(
	ch chan<- prometheus.Metric,
	httpArr, httpsArr []*ingress.ClbStatefulSetHttpRule, tcpArr, udpArr []*ingress.ClbStatefulSetRule) {
	// http
	for _, http := range httpArr {
		ch <- prometheus.MustNewConstMetric(
			newClbStatefulSetRuleMetricDesc(p.opt.ClbName),
			prometheus.GaugeValue,
			float64(http.ClbPort),
			[]string{http.ServiceName, http.Namespace,
				strconv.Itoa(http.ClbPort), "http", http.Host, http.Path,
				strconv.Itoa(http.StartIndex), strconv.Itoa(http.EndIndex)}...,
		)
	}
	// https
	for _, https := range httpsArr {
		ch <- prometheus.MustNewConstMetric(
			newClbStatefulSetRuleMetricDesc(p.opt.ClbName),
			prometheus.GaugeValue,
			float64(https.ClbPort),
			[]string{https.ServiceName, https.Namespace,
				strconv.Itoa(https.ClbPort), "https", https.Host, https.Path,
				strconv.Itoa(https.StartIndex), strconv.Itoa(https.EndIndex)}...,
		)
	}
	// tcp
	for _, tcp := range tcpArr {
		ch <- prometheus.MustNewConstMetric(
			newClbStatefulSetRuleMetricDesc(p.opt.ClbName),
			prometheus.GaugeValue,
			float64(tcp.ClbPort),
			[]string{tcp.ServiceName, tcp.Namespace,
				strconv.Itoa(tcp.ClbPort), "tcp", "", "",
				strconv.Itoa(tcp.StartIndex), strconv.Itoa(tcp.EndIndex)}...,
		)
	}
	// udp
	for _, udp := range udpArr {
		ch <- prometheus.MustNewConstMetric(
			newClbStatefulSetRuleMetricDesc(p.opt.ClbName),
			prometheus.GaugeValue,
			float64(udp.ClbPort),
			[]string{udp.ServiceName, udp.Namespace,
				strconv.Itoa(udp.ClbPort), "udp", "", "",
				strconv.Itoa(udp.StartIndex), strconv.Itoa(udp.EndIndex)}...,
		)
	}
}

// collect cloud listener info from cloud api
func (p *Processor) collectRemoteListener(ch chan<- prometheus.Metric, listeners []*loadbalance.CloudListener) {
	if len(listeners) == 0 {
		return
	}
	for _, listener := range listeners {
		if listener.Status.HealthStatus == nil {
			continue
		}
		if len(listener.Status.HealthStatus.RulesHealth) == 0 {
			continue
		}
		for _, ruleHealth := range listener.Status.HealthStatus.RulesHealth {
			for _, backendHealth := range ruleHealth.Backends {
				healthCode := 0
				if backendHealth.HealthStatus {
					healthCode = 1
				}
				ch <- prometheus.MustNewConstMetric(
					newRemoteBackendMetricDesc(p.opt.ClbName),
					prometheus.GaugeValue,
					float64(healthCode),
					[]string{
						strconv.Itoa(listener.Spec.ListenPort),
						ruleHealth.Domain,
						ruleHealth.URL,
						backendHealth.IP,
						strconv.Itoa(backendHealth.Port),
						backendHealth.HealthStatusDetail,
					}...,
				)
			}
		}
	}
}

// Collect implements prometheus exporter Collect interface
func (p *Processor) Collect(ch chan<- prometheus.Metric) {
	// ingress
	ingresses, err := p.ingressRegistry.ListIngresses()
	if err != nil {
		blog.Warnf("failed to list ingress in exporter, err %s", err.Error())
	} else {
		for _, ingress := range ingresses {
			p.collectIngress(ch, ingress.Spec.HTTP, ingress.Spec.HTTPS, ingress.Spec.TCP, ingress.Spec.UDP)
			if ingress.Spec.StatefulSet != nil {
				p.collectStatefulSetIngress(ch, ingress.Spec.StatefulSet.HTTP, ingress.Spec.StatefulSet.HTTPS,
					ingress.Spec.StatefulSet.TCP, ingress.Spec.StatefulSet.UDP)
			}
		}
	}
	// local listener
	listeners, err := p.updater.listenerClient.ListListeners()
	if err != nil {
		blog.Warnf("failed to list ingress in exporter, err %s", err.Error())
	} else {
		for _, listener := range listeners {
			switch listener.Spec.Protocol {
			// http or https listener
			case loadbalance.ClbListenerProtocolHTTP, loadbalance.ClbListenerProtocolHTTPS:
				for _, rule := range listener.Spec.Rules {
					ch <- prometheus.MustNewConstMetric(
						newClbListenerMetricDesc(p.opt.ClbName),
						prometheus.GaugeValue,
						float64(listener.Spec.ListenPort),
						[]string{
							listener.Spec.ListenerID,
							strconv.Itoa(listener.Spec.ListenPort),
							listener.Spec.Protocol,
							rule.Domain,
							rule.URL,
						}...,
					)
					if len(rule.TargetGroup.Backends) == 0 {
						continue
					}
					for _, backend := range rule.TargetGroup.Backends {
						ch <- prometheus.MustNewConstMetric(
							newClbBackendMetricDesc(p.opt.ClbName),
							prometheus.GaugeValue,
							float64(listener.Spec.ListenPort),
							[]string{
								listener.Spec.ListenerID,
								strconv.Itoa(listener.Spec.ListenPort),
								listener.Spec.Protocol,
								backend.IP,
								strconv.Itoa(backend.Port),
							}...,
						)
					}
				}
			// tcp or udp listener
			case loadbalance.ClbListenerProtocolTCP, loadbalance.ClbListenerProtocolUDP:
				ch <- prometheus.MustNewConstMetric(
					newClbListenerMetricDesc(p.opt.ClbName),
					prometheus.GaugeValue,
					float64(listener.Spec.ListenPort),
					[]string{
						listener.Spec.ListenerID,
						strconv.Itoa(listener.Spec.ListenPort),
						listener.Spec.Protocol, "", "",
					}...,
				)
				if len(listener.Spec.TargetGroup.Backends) == 0 {
					continue
				}
				for _, backend := range listener.Spec.TargetGroup.Backends {
					ch <- prometheus.MustNewConstMetric(
						newClbBackendMetricDesc(p.opt.ClbName),
						prometheus.GaugeValue,
						float64(listener.Spec.ListenPort),
						[]string{
							listener.Spec.ListenerID,
							strconv.Itoa(listener.Spec.ListenPort),
							listener.Spec.Protocol,
							backend.IP,
							strconv.Itoa(backend.Port),
						}...,
					)
				}
			default:
				blog.Warnf("invalid protocol %s", listener.Spec.Protocol)
				continue
			}
		}
	}
	// remote listener
	remoteListeners, err := p.updater.ListRemoteListener()
	if err != nil {
		blog.Warnf("failed to list remote listeners in exporter, err %s", err.Error())
	}
	if len(remoteListeners) != 0 {
		p.collectRemoteListener(ch, remoteListeners)
	}
}
