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
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	bresp "github.com/Tencent/bk-bcs/bcs-common/common/http"

	"github.com/coredns/coredns/plugin/etcd/msg"
	etcdcv3 "github.com/coreos/etcd/clientv3"

	restful "github.com/emicklei/go-restful"
	"golang.org/x/net/context"
)

func newHTTPServer(prefix string, cli *etcdcv3.Client) *httpServer {
	return &httpServer{
		RootPrefix: prefix,
		EtcdCli:    cli,
		Ctx:        context.Background(),
	}
}

type httpServer struct {
	// root dir of etcd.
	RootPrefix string
	EtcdCli    *etcdcv3.Client
	Ctx        context.Context
}

func (h httpServer) writeResponse(resp *restful.Response, value interface{}) {
	err := resp.WriteEntity(value)
	if err != nil {
		log.Printf("resp.WriteEntity failed. err: %v", err)
	}
}

func (h httpServer) value2Hash(value []byte) string {
	hash := fnv.New32a()
	_, err := hash.Write(value)
	if err != nil {
		log.Printf("hash.Write failed. err: %v", err)
	}
	return fmt.Sprintf("%x", hash.Sum32())
}

// CreateDomain api for domain creation
func (h httpServer) CreateDomain(req *restful.Request, resp *restful.Response) {
	data, e := ioutil.ReadAll(req.Request.Body)
	if e != nil {
		log.Printf("[ERROR] read create domain request body failed. err: %v", e)
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: e.Error()})
		return
	}
	log.Printf("[INFO] received create domain request, source[%s], req data:%s", req.Request.RemoteAddr, string(data))

	dns := new(DNS)
	if err := json.Unmarshal(data, dns); err != nil {
		log.Printf("[ERROR] create domain, unmarshal request body failed. err: %v", err)
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: err.Error()})
		return
	}
	//fixme: how to make a transaction create
	for _, dmsg := range dns.Messages {
		if len(dmsg.Alias) == 0 {
			// generate a hash key
			dmsgBytes, err := json.Marshal(dmsg)
			if err != nil {
				log.Printf("[ERROR] dns to json marshal err, %s", err.Error())
				h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: err.Error()})
				return
			}
			dmsg.Alias = h.value2Hash(dmsgBytes)
			log.Printf("[INFO] create domain request with empty alias, generate : %v", dmsg.Alias)
		}
		key, err := getDomainKey(dns.DomainName, dmsg.Alias, h.RootPrefix)
		if err != nil {
			log.Printf("[ERROR] create domain[%s] get key failed. err: %v", dns.DomainName, err)
			h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: err.Error()})
			return
		}

		emsg, err := dmsg.toService()
		if err != nil {
			log.Printf("[ERROR] convert create domain request body failed. err: %v", err)
			h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: err.Error()})
			return
		}
		_, err = h.EtcdCli.Put(h.Ctx, key, emsg)
		if err != nil {
			log.Printf("[ERROR] received create domain request ,but set domain[%s] failed. err: %v", dns.DomainName, err)
			h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: fmt.Sprintf("set domain[%s] failed. err: %v", dns.DomainName, err)})
			return
		}
	}

	DnsTotal.Inc()
	h.writeResponse(resp, bresp.APIRespone{Result: true, Code: 0, Message: "create success"})
	log.Printf("create domain[%s] success.", dns.DomainName)
}

func (h httpServer) UpdateDomain(req *restful.Request, resp *restful.Response) {
	var err error
	data, err := ioutil.ReadAll(req.Request.Body)
	if nil != err {
		log.Printf("[ERROR] read update domain request body failed. err: %v", err)
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: fmt.Sprintf("read request body failed. err: %v", err)})
		return
	}
	log.Printf("[INFO] received update domain request, source[%s], req data:%s", req.Request.RemoteAddr, string(data))

	dns := new(DNS)
	if err = json.Unmarshal(data, dns); err != nil {
		log.Printf("[ERROR] unmarshal request body failed. err: %v", err)
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: fmt.Sprintf("unmarshal request body failed. err: %v", err)})
		return
	}

	for _, dmsg := range dns.Messages {
		if len(dmsg.Alias) == 0 {
			// generate a hash key
			dmsgBytes, err := json.Marshal(dmsg)
			if err != nil {
				log.Printf("[ERROR] dns to json marshal err, %s", err.Error())
				h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: err.Error()})
				return
			}
			dmsg.Alias = h.value2Hash(dmsgBytes)
			log.Printf("[INFO] update domain request with empty alias, generate : %v", dmsg.Alias)
		}

		key, err := getDomainKey(dns.DomainName, dmsg.Alias, h.RootPrefix)
		if err != nil {
			log.Printf("[ERROR] update domain[%s] get key failed. err: %v", dns.DomainName, err)
			h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: err.Error()})
			return
		}

		emsg, err := dmsg.toService()
		if err != nil {
			log.Printf("[ERROR] convert update domain request body failed. err: %v", err)
			h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: fmt.Sprintf("invalid request, err: %v", err)})
			return
		}
		_, err = h.EtcdCli.Put(h.Ctx, key, emsg)
		if err != nil {
			log.Printf("[ERROR] received update domain request ,but set domain[%s] failed. err: %v", dns.DomainName, err)
			h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: fmt.Sprintf("set domain[%s] failed. err: %v", dns.DomainName, err)})
			return
		}
		log.Printf("[INFO] update domain[%s] alias[%s] with value: %s success.", dns.DomainName, dmsg.Alias, emsg)
	}

	h.writeResponse(resp, bresp.APIRespone{Result: true, Code: 0, Message: "update success"})
}

func (h httpServer) GetDomain(req *restful.Request, resp *restful.Response) {
	domain := req.Request.FormValue("domain")
	if len(domain) == 0 {
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: "invalid value with empty domain."})
		log.Printf("[ERROR] received get domain reqeust, but get empty domain.")
		return
	}

	// Don't check empty alias here, so that it can support legacy usage.
	alias := req.Request.FormValue("alias")
	log.Printf("[INFO] received get domain request, domain: %s alias: %s", domain, alias)

	key, err := getDomainKey(domain, alias, h.RootPrefix)
	if err != nil {
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: err.Error()})
		log.Printf("[ERROR] get domain[%s] alias[%s] get key failed. err: %v", domain, alias, err)
		return
	}

	r, err := h.EtcdCli.Get(h.Ctx, key, etcdcv3.WithPrefix())
	if err != nil {
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: err.Error()})
		log.Printf("[ERROR] get domain[%s] alias[%s] failed. err: %v", domain, alias, err)
		return
	}

	log.Printf("[INFO] etcd response is: %v", r)
	ds := respToDNS(r, h.RootPrefix)
	if len(ds) == 0 {
		message := fmt.Sprintf("domain [%s] not found", domain)
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: message})
		log.Printf("[ERROR] get domain[%s] alias[%s] failed. err: %v", domain, alias, err)
		return
	}
	if len(ds) > 1 {
		message := fmt.Sprintf("get multiple domains by domain: %s", domain)
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: message})
		log.Printf("[ERROR] get domain[%s] alias[%s] failed. err: %v", domain, alias, err)
		return
	}

	dmsges := &ds[0].Messages
	h.writeResponse(resp, bresp.APIRespone{Result: true, Code: 0, Message: "get success", Data: dmsges})
	js, err := json.Marshal(dmsges)
	if err != nil {
		log.Printf("[ERROR] json.Marshal(dmsg) failed: %s", err.Error())
	}
	log.Printf("[INFO] get domain[%s] alias[%s] success. data: %s", domain, alias, js)
}

func (h httpServer) DeleteDomain(req *restful.Request, resp *restful.Response) {
	domain := req.Request.FormValue("domain")
	if len(domain) == 0 {
		log.Printf("[ERROR] received delete domain reqeust, but get empty domain.")
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: "invalid value with empty domain."})
		return
	}
	h.deleteDomain(req, resp, domain, "")
}

func (h httpServer) DeleteAlias(req *restful.Request, resp *restful.Response) {
	domain := req.Request.FormValue("domain")
	if len(domain) == 0 {
		log.Printf("[ERROR] received delete domain reqeust, but get empty domain.")
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: "invalid value with empty domain."})
		return
	}
	alias := req.Request.FormValue("alias")
	if len(alias) == 0 {
		log.Printf("[ERROR] received delete domain reqeust, but get empty alias.")
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: "invalid value with empty alias."})
		return
	}
	h.deleteDomain(req, resp, domain, alias)
}

func (h httpServer) deleteDomain(req *restful.Request, resp *restful.Response, domain, alias string) {
	log.Printf("[INFO] received delete domain request, domain: %s, alias: %s", domain, alias)

	// prevent from deleting two many domains
	// 1. prevent delete with wildcard
	name := fmt.Sprintf("%s.%s", alias, domain)
	if wildcardDomain.MatchString(name) {
		message := "delete domain with wildcard not support yet."
		log.Printf("[ERROR] delete domain[%s] alias[%s] get key failed. err: %v", domain, alias, message)
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: message})
		return
	}

	key, err := getDomainKey(domain, alias, h.RootPrefix)
	if err != nil {
		log.Printf("[ERROR] delete domain[%s] alias[%s] get key failed. err: %v", domain, alias, err)
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	// 2. prevent delete multiple domains one time
	r, err := h.EtcdCli.Get(h.Ctx, key, etcdcv3.WithPrefix())
	if err != nil {
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: err.Error()})
		log.Printf("[ERROR] get domain[%s] alias[%s] failed. err: %v", domain, alias, err)
		return
	}
	ds := respToDNS(r, h.RootPrefix)
	domains := make([]string, 0)
	for _, dns := range ds {
		domains = append(domains, dns.DomainName)
	}
	log.Printf("[INFO] try to delete %d domains, [%s]", len(ds), strings.Join(domains, ";"))
	if len(ds) > 1 {
		message := fmt.Sprintf("doamin[%s] match multiple domains, only support delete one domain every time.", domain)
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: message})
		log.Printf("[ERROR] delete domain[%s] alias[%s] failed. err: %v", domain, alias, err)
		return
	}

	if len(ds) == 0 {
		var message string
		if alias == "" {
			message = fmt.Sprintf("domain[%s] not found.", domain)
		} else {
			message = fmt.Sprintf("domain[%s] alias[%s] not found.", domain, alias)
		}
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: message})
		log.Printf("[ERROR] delete domain[%s] alias[%s] failed. err: %v", domain, alias, err)
		return
	}

	//todo: when alias is empty, delete option must be WithPrefix()
	_, err = h.EtcdCli.Delete(h.Ctx, key, etcdcv3.WithPrefix())
	if err != nil {
		log.Printf("[ERROR] delete domain[%s] failed. err: %v", domain, err)
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	DnsTotal.Dec()
	h.writeResponse(resp, bresp.APIRespone{Result: true, Code: 0, Message: "delete success"})
	log.Printf("[INFO] delete domain[%s] success.", domain)
}

func (h httpServer) ListDomain(req *restful.Request, resp *restful.Response) {
	zone := req.Request.FormValue("zone")
	if len(zone) == 0 {
		log.Printf("[ERROR] received list domain reqeust, but get empty domain.")
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: "invalid value with empty domain."})
		return
	}

	log.Printf("[INFO] received list domain request, list zone: %s", zone)
	key, err := getDomainKey(zone, "", h.RootPrefix)
	if err != nil {
		log.Printf("[ERROR] list domain[%s] get key failed. err: %v", zone, err)
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	r, err := h.EtcdCli.Get(h.Ctx, key, etcdcv3.WithPrefix())
	if err != nil {
		log.Printf("[ERROR] list domain[%s] failed. err: %v", zone, err)
		h.writeResponse(resp, bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	ds := respToDNS(r, h.RootPrefix)
	h.writeResponse(resp, bresp.APIRespone{Result: true, Code: 0, Message: "list success", Data: ds})
	js, err := json.Marshal(ds)
	if err != nil {
		log.Printf("[ERROR] json.Marshal(ds) failed: %s", err.Error())
	}
	log.Printf("[INFO] list domain[%s] success. data: %s", zone, js)
}

// standard wildcard domain name example: demo.ied.*.bcscustom.com
var wildcardDomain = regexp.MustCompile(`^([a-zA-Z0-9-]+\.)+(\*\.)([a-zA-Z0-9-]+\.)([a-zA-Z0-9]+)$`)

func getDomainKey(name, alias, prefix string) (string, error) {
	name = strings.TrimSuffix(name, ".")
	if len(alias) != 0 {
		name = fmt.Sprintf("%s.%s", alias, name)
	}
	if wildcardDomain.MatchString(name) {
		name = strings.Replace(name, "*", dnsAnyDomain, 1)
	}

	domainPath, isWildcard := msg.PathWithWildcard(name, prefix)
	if isWildcard {
		return "", errors.New("invalid domain name with wildcard")
	}

	// so that it can use etcd v3 api more simple
	// if not ends with `/`, use etcd v3 api prefix to match domain exactly will be much complex
	domainPath += "/"
	return domainPath, nil
}

const dnsAnyDomain string = "dnsany"

// DNS data structure for request
type DNS struct {
	DomainName string       `json:"domain"`
	Messages   []DNSMessage `json:"messages"`
}

// DNSMessage dns detail for one host
type DNSMessage struct {
	Alias    string `json:"alias"` // extend field, alias name for this message
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Weight   int    `json:"weight,omitempty"`
	Text     string `json:"text,omitempty"`
	Mail     bool   `json:"mail,omitempty"`
	TTL      uint32 `json:"ttl,omitempty"`

	// When a SRV record with a "Host: IP-address" is added, we synthesize
	// a srv.Target domain name.  Normally we convert the full Key where
	// the record lives to a DNS name and use this as the srv.Target.  When
	// TargetStrip > 0 we strip the left most TargetStrip labels from the
	// DNS name.
	TargetStrip int `json:"targetstrip,omitempty"`

	// Group is used to group (or *not* to group) different services
	// together. Services with an identical Group are returned in the same
	// answer.
	Group string `json:"group,omitempty"`
}

// ToService make hsoue
func (m DNSMessage) toService() (string, error) {
	if len(m.Alias) == 0 {
		return "", errors.New("alias can not be empty")
	}

	if len(m.Host) == 0 {
		return "", fmt.Errorf("host can not be null")
	}

	if m.Port < 0 || m.Port > 65535 {
		return "", fmt.Errorf("invalid port: %d", m.Port)
	}

	svc := msg.Service{
		Host:        m.Host,
		Port:        m.Port,
		Priority:    m.Priority,
		Weight:      m.Weight,
		Text:        m.Text,
		Mail:        m.Mail,
		TTL:         m.TTL,
		TargetStrip: m.TargetStrip,
		Group:       m.Group,
	}

	js, err := json.Marshal(svc)
	if err != nil {
		return "", err
	}

	return string(js), nil
}

// respToDNS parse etcd response to lcoal DNS slice
func respToDNS(s *etcdcv3.GetResponse, root string) []*DNS {
	domainMesssageMap := make(map[string][]DNSMessage)
	d := make([]*DNS, 0)
	for _, kv := range s.Kvs {
		dmsg := new(DNSMessage)
		if err := json.Unmarshal(kv.Value, dmsg); err != nil {
			log.Printf("[ERROR] unmarshal key[%s] with value: [%s] failed. err :%v", string(kv.Key), string(kv.Value), err)
			return d
		}

		domain, alias := keyTODomainName(string(kv.Key), root)
		dmsg.Alias = alias
		_, ok := domainMesssageMap[domain]
		if !ok {
			domainMesssageMap[domain] = make([]DNSMessage, 0)
		}
		domainMesssageMap[domain] = append(domainMesssageMap[domain], *dmsg)
	}
	for domain, dmsgSlice := range domainMesssageMap {
		d = append(d, &DNS{
			DomainName: domain,
			Messages:   dmsgSlice,
		})
	}
	return d
}

func keyTODomainName(key string, root string) (domain string, alias string) {
	key = strings.Replace(key, "dnsany", "*", 1)
	key = strings.TrimPrefix(key, "/"+root)
	segs := strings.Split(key, "/")
	for i, j := 0, len(segs)-1; i < j; i, j = i+1, j-1 {
		segs[i], segs[j] = segs[j], segs[i]
	}
	if segs[0] == "" {
		segs = segs[1:] // remove empty string at head
	}
	domain = strings.Join(segs[1:], ".")
	alias = segs[0]
	return domain, alias
}
