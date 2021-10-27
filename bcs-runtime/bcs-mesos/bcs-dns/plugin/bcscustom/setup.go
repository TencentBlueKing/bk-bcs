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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/etcd"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/coredns/coredns/plugin/pkg/fall"
	mwtls "github.com/coredns/coredns/plugin/pkg/tls"
	"github.com/coredns/coredns/plugin/pkg/upstream"
	"github.com/coredns/coredns/plugin/proxy"
	etcdcv3 "github.com/coreos/etcd/clientv3"
	restful "github.com/emicklei/go-restful"
	"github.com/mholt/caddy"
	"golang.org/x/net/context"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	caddy.RegisterPlugin("bcscustom", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	e, err := parseConfig(c)
	if err != nil {
		return plugin.Error("bcscustom", err)
	}

	if err := common.SavePid(conf.ProcessConfig{PidDir: "./pid"}); err != nil {
		return plugin.Error("bcsscheduler", err)
	}

	metrics.MustRegister(c, RequestCount, RequestLatency, DnsTotal)
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		e.Next = next
		return e
	})

	return nil
}

// #lizard forgives
// nolint
func parseConfig(c *caddy.Controller) (*BcsCustom, error) {
	bc := BcsCustom{
		// set the default value.
		RootPrefix: "bcscustom",
		Ctx:        context.Background(),
	}
	var (
		tlsConfig *tls.Config
		errOutter error
		endpoints = []string{defaultEndpoint}
	)
	for c.Next() {
		bc.Zones = c.RemainingArgs()
		if len(bc.Zones) == 0 {
			bc.Zones = make([]string, len(c.ServerBlockKeys))
			copy(bc.Zones, c.ServerBlockKeys)
		}
		for i, str := range bc.Zones {
			bc.Zones[i] = plugin.Host(str).Normalize()
		}

		if c.NextBlock() {
			for {
				switch c.Val() {
				case "fallthrough":
					bc.FallThrough = true
				case "etcd-endpoints":
					args := c.RemainingArgs()
					if len(args) == 0 {
						return nil, c.ArgErr()
					}
					endpoints = args
				case "upstream":
					args := c.RemainingArgs()
					if len(args) == 0 {
						return nil, c.ArgErr()
					}
					u, err := upstream.New(args)
					if err != nil {
						return nil, err
					}
					bc.Upstream = u
				case "etcd-tls": // cert key cacertfile
					args := c.RemainingArgs()
					tlsConfig, errOutter = mwtls.NewTLSConfigFromArgs(args...)
					if errOutter != nil {
						return nil, errOutter
					}
				case "listen":
					if !c.NextArg() {
						return nil, c.ArgErr()
					}
					bc.Listen = c.Val()
				case "ca-file":
					if !c.NextArg() {
						return nil, c.ArgErr()
					}
					if len(c.Val()) != 0 {
						bc.SvrTLS.CaFile = c.Val()
					}
				case "key-file":
					if !c.NextArg() {
						return nil, c.ArgErr()
					}
					if len(c.Val()) != 0 {
						bc.SvrTLS.KeyFile = c.Val()
					}
				case "cert-file":
					if !c.NextArg() {
						return nil, c.ArgErr()
					}
					if len(c.Val()) != 0 {
						bc.SvrTLS.CertFile = c.Val()
					}
				case "root-prefix":
					if !c.NextArg() {
						return nil, c.ArgErr()
					}
					if len(c.Val()) != 0 {
						bc.RootPrefix = c.Val()
					}
				default:
					if c.Val() != "}" {
						return nil, c.Errf("unknown property '%s'", c.Val())
					}
				}

				if !c.Next() {
					break
				}
			}

		}
		client, err := newEtcdClient(endpoints, tlsConfig)
		if err != nil {
			log.Printf("ERROR newEtcdClient failed, err: %v", err)
			return nil, err
		}
		bc.EtcdCli = client
		bc.endpoints = endpoints
		stub := make(map[string]proxy.Proxy)
		bc.EtcdPlugin = &etcd.Etcd{
			Next:       nil,
			Fall:       fall.F{},
			Zones:      bc.Zones,
			PathPrefix: bc.RootPrefix,
			Upstream:   bc.Upstream,
			Client:     bc.EtcdCli,
			Stubmap:    &stub,
			//Endpoints:  bc.endpoints,
			Ctx: context.Background(),
		}

		log.Printf("[Info] bcscustom config info: %+#v", bc)

		if err := startHTTPServer(bc.RootPrefix, client, bc.SvrTLS, bc.Listen); err != nil {
			log.Printf("[ERROR] start http server failed. err: %v", err)
			return nil, err
		}

		return &bc, nil
	}
	return nil, nil
}

func newEtcdClient(endpoints []string, cc *tls.Config) (*etcdcv3.Client, error) {
	etcdCfg := etcdcv3.Config{
		Endpoints: endpoints,
		TLS:       cc,
	}
	cli, err := etcdcv3.New(etcdCfg)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

const defaultEndpoint = "http://localhost:2379"

func startHTTPServer(rootPrefix string, etcdCli *etcdcv3.Client, svrTLS TLS, listen string) error {
	svr := newHTTPServer(rootPrefix, etcdCli)
	api := new(restful.WebService).Path("/bcsdns/v1/").Produces(restful.MIME_JSON)
	container := restful.NewContainer()
	container.Add(api)
	api.Route(api.POST("domains").To(svr.CreateDomain))
	api.Route(api.DELETE("domains").To(svr.DeleteDomain))
	api.Route(api.DELETE("domains/alias").To(svr.DeleteAlias))
	api.Route(api.PUT("domains").To(svr.UpdateDomain))
	api.Route(api.GET("domains").To(svr.GetDomain))
	api.Route(api.GET("domains/subdomains").To(svr.ListDomain))

	if len(svrTLS.KeyFile) == 0 &&
		len(svrTLS.CertFile) == 0 &&
		len(svrTLS.CaFile) == 0 {
		log.Printf("[INFO] start insecure server on:%s", listen)
		go func() {
			s := &http.Server{
				Addr:    listen,
				Handler: container,
			}
			if err := s.ListenAndServe(); err != nil {
				log.Fatalf("start http server failed. err: %v", err)
			}
		}()
		return nil
	}

	// use https server
	ca, err := ioutil.ReadFile(svrTLS.CaFile)
	if nil != err {
		return fmt.Errorf("read server tls file failed. err:%v", err)
	}
	capool := x509.NewCertPool()
	capool.AppendCertsFromPEM(ca)
	tlsconfig, err := ssl.ServerTslConfVerityClient(svrTLS.CaFile,
		svrTLS.CertFile,
		svrTLS.KeyFile,
		"Q5PNRjEZ7ri9vFGo")
	if err != nil {
		return fmt.Errorf("generate tls config failed. err: %v", err)
	}
	tlsconfig.BuildNameToCertificate()

	log.Printf("start secure serve on %s", listen)

	ln, err := net.Listen("tcp", listen)
	if err != nil {
		return fmt.Errorf("listen secure server failed. err: %v", err)
	}
	listener := tls.NewListener(ln, tlsconfig)
	go func() {
		if err := http.Serve(listener, container); nil != err {
			log.Fatalf("server https failed. err: %v", err)
		}
	}()
	return nil
}
