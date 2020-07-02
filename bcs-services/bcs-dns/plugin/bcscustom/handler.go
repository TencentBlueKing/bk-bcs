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

package bcscustom

import (
	"log"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-dns/plugin/util"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
	"time"
)

// ServeDNS implements the plugin.Handler interface.
func (bc *BcsCustom) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	start := time.Now()

	opt := plugin.Options{}
	state := request.Request{W: w, Req: r, Context: ctx}
	zone := plugin.Zones(bc.Zones).Matches(state.Name())
	if zone == "" {
		if bc.FallThrough {
			return plugin.NextOrFailure(bc.Name(), bc.Next, ctx, w, r)
		}
		return plugin.BackendError(bc, zone, dns.RcodeNameError, state, nil, opt)
	}

	interceptor := util.NewResponseInterceptor(state.W)
	m := new(dns.Msg)
	m.SetQuestion(state.Name(), state.QType())
	if code, err := bc.EtcdPlugin.ServeDNS(ctx, interceptor, m); err != nil {
		log.Printf("[ERROR] get request[%s] from etcd plugin failed, err:%v", state.Name(), err)
		RequestCount.WithLabelValues(Failure).Inc()
		RequestLatency.WithLabelValues(Failure).Observe(time.Since(start).Seconds())
		return plugin.BackendError(bc, zone, code, state, err, opt)
	}
	m.SetReply(r)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, true, true
	if interceptor.Msg != nil && len(interceptor.Msg.Answer) != 0 {
		m.Answer = append(m.Answer, interceptor.Msg.Answer...)
		m.Extra = append(m.Extra, interceptor.Msg.Extra...)
	} else {
		result, err := bc.Lookup(state, state.Name(), state.QType())
		if err != nil {
			log.Printf("[ERROR] get request[%s] from *upstream* failed, err:%v", state.Name(), err)
			RequestCount.WithLabelValues(Failure).Inc()
			RequestLatency.WithLabelValues(Failure).Observe(time.Since(start).Seconds())
			return plugin.BackendError(bc, zone, dns.RcodeNameError, state, err, opt)
		}
		m.Answer = append(m.Answer, result.Answer...)
	}

	m.Answer = dns.Dedup(m.Answer, nil)
	m.Ns = dns.Dedup(m.Ns, nil)
	m.Extra = dns.Dedup(m.Extra, nil)
	state.SizeAndDo(m)
	m = state.Scrub(m)
	if err := w.WriteMsg(m); err != nil {
		log.Printf("[ERROR] response to client failed, %s", err.Error())
		RequestCount.WithLabelValues(Failure).Inc()
		RequestLatency.WithLabelValues(Failure).Observe(time.Since(start).Seconds())
		return dns.RcodeServerFailure, err
	}
	RequestCount.WithLabelValues(Success).Inc()
	RequestLatency.WithLabelValues(Success).Observe(time.Since(start).Seconds())
	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (bc *BcsCustom) Name() string { return "bcscustom" }
