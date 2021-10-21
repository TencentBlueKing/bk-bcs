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
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-dns/plugin/bcsscheduler/metrics"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func (bcs *BcsScheduler) tryReviseRequest(originalName string, req *request.Request) (revised bool) {
	segs := dns.SplitDomainName(originalName)
	if (req.Req.Question[0].Qtype == dns.TypeA || req.Req.Question[0].Qtype == dns.TypeAAAA) && len(segs) == 2 {
		// A record
		segs = append(segs, "svc", bcs.conf.Cluster, bcs.conf.Zones[0])
		req.Clear()
		req.Req.Question[0].Name = strings.Join(segs, ".")
		req.Name()
		req.Req.Question[0].Name = originalName
		revised = true
	} else if (req.Req.Question[0].Qtype == dns.TypeA ||
		req.Req.Question[0].Qtype == dns.TypeAAAA) &&
		// two scenario:
		// svcname.namespace.svc
		// podname.svcname.namespace.svc
		(len(segs) == 3 || len(segs) == 4) &&
		strings.HasSuffix(strings.TrimSuffix(originalName, "."), ".svc") {
		segs = append(segs, bcs.conf.Cluster, bcs.conf.Zones[0])
		req.Clear()
		req.Req.Question[0].Name = strings.Join(segs, ".")
		req.Name()
		req.Req.Question[0].Name = originalName
		revised = true
	} else if req.Req.Question[0].Qtype == dns.TypeSRV &&
		(strings.HasPrefix(originalName, "_") || strings.HasPrefix(originalName, "*.")) &&
		len(segs) == 4 {
		// srv record.
		segs = append(segs, "svc", bcs.conf.Cluster, bcs.conf.Zones[0])
		req.Clear()
		req.Req.Question[0].Name = strings.Join(segs, ".")
		req.Name()
		req.Req.Question[0].Name = originalName
		revised = true
	}
	return
}

func (bcs *BcsScheduler) getQualifiedQuestionName(originalName string, qType uint16) (string, bool) {
	segs := dns.SplitDomainName(originalName)
	qualified := false
	if (qType == dns.TypeA || qType == dns.TypeAAAA) && len(segs) == 2 {
		// A record, svcname.namespace
		segs = append(segs, "svc", bcs.conf.Cluster, bcs.conf.Zones[0])
		qualified = true
	} else if (qType == dns.TypeA || qType == dns.TypeAAAA) && (len(segs) == 3 || len(segs) == 4) &&
		strings.HasSuffix(strings.TrimSuffix(originalName, "."), ".svc") {
		// two scenario:
		// svcname.namespace.svc
		// podname.svcname.namespace.svc
		segs = append(segs, bcs.conf.Cluster, bcs.conf.Zones[0])
		qualified = true
	} else if qType == dns.TypeSRV && (strings.HasPrefix(originalName, "_") || strings.HasPrefix(originalName, "*.")) &&
		len(segs) == 5 && strings.HasSuffix(strings.TrimSuffix(originalName, "."), ".svc") {
		// srv record: _port._protocol.svcname.namespace.svc
		segs = append(segs, bcs.conf.Cluster, bcs.conf.Zones[0])
		qualified = true
	}
	return strings.Join(segs, "."), qualified
}

func (bcs *BcsScheduler) tryWithOutterDns(ctx context.Context, req *request.Request) (yes bool, code int, err error) {
	zone := plugin.Zones(bcs.conf.Zones).Matches(req.Name())
	switch {
	case req.Req.Question[0].Qtype == dns.TypePTR:
		fallthrough
	case req.Req.Question[0].Qtype != dns.TypePTR && zone == "":
		// not belongs to bcs zone.
		if bcs.conf.Fallthrough {
			code, err = plugin.NextOrFailure(bcs.Name(), bcs.Next, ctx, req.W, req.Req)
			if err != nil {
				log.Printf("[ERROR] proxy for %s err: %+v", req.Req.Question[0].Name, err)
				return true, code, err
			}
			return true, code, nil
		}
		code, err = plugin.BackendError(bcs, zone, dns.RcodeNameError, *req, nil, plugin.Options{})
		return true, code, err
	}
	return
}

func (bcs *BcsScheduler) inBcsZone(zoneName string) bool {
	return dns.IsSubDomain("bcs.com.", zoneName)
}

func (bcs *BcsScheduler) inCurrentZone(zoneName string) bool {
	return dns.IsSubDomain(bcs.PrimaryZone(), zoneName)
}

func (bcs *BcsScheduler) dealCommonResolve(zone string, state request.Request) (records, extra []dns.RR, err error) {
	switch state.Type() {
	case "A":
		records, err = plugin.A(bcs, zone, state, nil, plugin.Options{})
	case "AAAA":
		records, err = plugin.AAAA(bcs, zone, state, nil, plugin.Options{})
	case "TXT":
		records, err = plugin.TXT(bcs, zone, state, plugin.Options{})
	case "CNAME":
		records, err = plugin.CNAME(bcs, zone, state, plugin.Options{})
	case "PTR":
		records, err = plugin.PTR(bcs, zone, state, plugin.Options{})
	case "MX":
		records, extra, err = plugin.MX(bcs, zone, state, plugin.Options{})
	case "SRV":
		records, extra, err = plugin.SRV(bcs, zone, state, plugin.Options{})
	case "SOA":
		records, err = plugin.SOA(bcs, zone, state, plugin.Options{})
	case "NS":
		if state.Name() == zone {
			records, extra, err = plugin.NS(bcs, zone, state, plugin.Options{})
			break
		}
		fallthrough
	default:
		// Do a fake A lookup, so we can distinguish between NODATA and NXDOMAIN
		_, err = plugin.A(bcs, zone, state, nil, plugin.Options{})
	}
	return
}

// this function must be called after tryReviseRequest() is called.
func (bcs *BcsScheduler) validateRequest(state request.Request) error {
	segs := dns.SplitDomainName(state.Name())
	segments := segs[:len(segs)-dns.CountLabel(bcs.PrimaryZone())]
	switch state.QType() {
	case dns.TypeSRV:
		if len(segments) != 5 {
			return fmt.Errorf("invalid type SRV request, name: %s", state.Name())
		}
		if segments[0] == "_" || segments[1] == "_" {
			return fmt.Errorf("invalid type SRV request, name: %s", state.Name())
		}
	case dns.TypeA:
		if len(segments) != 3 && len(segments) != 4 {
			return fmt.Errorf("invalid type A request, name: %s", state.Name())
		}
	}
	return nil
}

// ServeDNS implements the plugin.Handler interface.
func (bcs *BcsScheduler) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	start := time.Now()

	originalName := strings.ToLower(dns.Name(r.Question[0].Name).String())
	state := request.Request{W: w, Req: r}
	if state.QClass() != dns.ClassINET {
		metrics.RequestCount.WithLabelValues(metrics.Failure).Inc()
		metrics.RequestLatency.WithLabelValues(metrics.Failure).Observe(time.Since(start).Seconds())
		return dns.RcodeServerFailure, plugin.Error(bcs.Name(), errors.New("can only deal with ClassINET"))
	}

	// handle the special scenes.
	bcs.tryReviseRequest(originalName, &state)
	yes, code, zerr := bcs.tryWithOutterDns(ctx, &state)
	if yes {
		return code, zerr
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, true, true

	writeBackMsg := func(rr []dns.RR, ext []dns.RR) {
		m.Answer = append(m.Answer, rr...)
		m.Extra = append(m.Extra, ext...)
		state.SizeAndDo(m)
		newMsg := state.Scrub(m)
		w.WriteMsg(newMsg)
	}

	zone := plugin.Zones(bcs.conf.Zones).Matches(state.Name())
	//log.Printf("[DEBUG] zones: %v, zone: %s, name: %s", bcs.conf.Zones, zone, state.Name())
	if zone != "" && bcs.inBcsZone(zone) {
		if bcs.inCurrentZone(state.Name()) {
			// request is in current zone.
			if e := bcs.validateRequest(state); e != nil {
				log.Printf("[ERROR] validate request[%s] failed, err: %v", state.Name(), e)
				metrics.RequestCount.WithLabelValues(metrics.Failure).Inc()
				metrics.RequestLatency.WithLabelValues(metrics.Failure).Observe(time.Since(start).Seconds())
				return plugin.BackendError(bcs, zone, dns.RcodeNameError, state, e, plugin.Options{})
			}
			var records, extra []dns.RR
			var err error
			records, extra, err = bcs.dealCommonResolve(zone, state)
			recursiveRecords, yes := bcs.dealRecursiveDNSWithClusterIP(state)

			if yes && len(recursiveRecords) != 0 {
				// add additional recursive records
				records = append(records, recursiveRecords...)
				if err != nil {
					// do response with these records
					writeBackMsg(records, extra)
					metrics.RequestCount.WithLabelValues(metrics.Success).Inc()
					metrics.RequestLatency.WithLabelValues(metrics.Success).Observe(time.Since(start).Seconds())
					return dns.RcodeSuccess, nil
				}
			}

			if err != nil {
				log.Printf("[ERROR] scheduler get current request failed err: %s", err.Error())
				metrics.RequestCount.WithLabelValues(metrics.Failure).Inc()
				metrics.RequestLatency.WithLabelValues(metrics.Failure).Observe(time.Since(start).Seconds())
				return dns.RcodeServerFailure, err
			}

			if len(records) == 0 {
				if state.QType() != dns.TypeAAAA {
					log.Printf("[ERROR] scheduler get no endpoint for %s, type: %s", state.Name(), state.Type())
				}
				metrics.RequestCount.WithLabelValues(metrics.Failure).Inc()
				metrics.RequestLatency.WithLabelValues(metrics.Failure).Observe(time.Since(start).Seconds())
				return plugin.BackendError(bcs, zone, dns.RcodeServerFailure, state, fmt.Errorf("got no endpoints"), plugin.Options{})
			}

			writeBackMsg(records, extra)
			metrics.RequestCount.WithLabelValues(metrics.Success).Inc()
			metrics.RequestLatency.WithLabelValues(metrics.Success).Observe(time.Since(start).Seconds())
			return dns.RcodeSuccess, nil

		}
		// request is in other bcs zone. reoute to upper bcs-dns server.
		msg, err := bcs.Lookup(state, state.Name(), state.QType())
		if err != nil {
			log.Printf("[ERROR] request %s post to bcs upper cluster failed, err: %s", state.Name(), err.Error())
			metrics.RequestOutProxyCount.WithLabelValues(metrics.Failure).Inc()
			metrics.RequestLatency.WithLabelValues(metrics.Failure).Observe(time.Since(start).Seconds())
			return dns.RcodeServerFailure, err
		}
		writeBackMsg(msg.Answer, msg.Extra)
		metrics.RequestOutProxyCount.WithLabelValues(metrics.Success).Inc()
		metrics.RequestLatency.WithLabelValues(metrics.Success).Observe(time.Since(start).Seconds())
		return dns.RcodeSuccess, nil
	}
	// this request is not belong to bcs zone.
	return bcs.routeToNextOrFailure(ctx, w, state)
}

func (bcs *BcsScheduler) routeToNextOrFailure(ctx context.Context, w dns.ResponseWriter, req request.Request) (int, error) {
	// clear request name first, in case the question name is not same with name.
	req.Clear()

	if bcs.conf.Fallthrough {
		return plugin.NextOrFailure(bcs.Name(), bcs.Next, ctx, w, req.Req)
	}
	log.Printf("[WARN] scheduler do not forward request to next plugin. request name:%v ", req.Name())

	// Make err nil when returning here, so we don't log spam for NXDOMAIN.
	return plugin.BackendError(bcs, req.Name(), dns.RcodeNameError, req, nil, plugin.Options{})
}

// Name implements the Handler interface.
func (bcs *BcsScheduler) Name() string { return "bcsscheduler" }
