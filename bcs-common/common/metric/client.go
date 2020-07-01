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

package metric

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func newMetricHandler(conf Config, healthFunc HealthFunc, metrics ...*MetricContructor) error {
	controller, err := newMetricController(conf, metrics...)
	if err != nil {
		return fmt.Errorf("new metric controller failed. err: %v", err)
	}

	registry := prometheus.NewRegistry()
	if err := registry.Register(controller); err != nil {
		return fmt.Errorf("register user defined metrics failed. err: %v", err)
	}
	if !conf.DisableGolangMetric {
		if err := registry.Register(prometheus.NewGoCollector()); err != nil {
			return fmt.Errorf("register golang metrics failed. err: %v", err)
		}
	}
	metricHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	mux := http.NewServeMux()
	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		h := healthFunc()
		info := HealthInfo{
			RunMode:    conf.RunMode,
			Module:     conf.ModuleName,
			ClusterID:  conf.ClusterID,
			IP:         conf.IP,
			HealthMeta: h,
			AtTime:     time.Now().Unix(),
		}
		js, err := json.MarshalIndent(info, "", "    ")
		if nil != err {
			w.WriteHeader(http.StatusInternalServerError)
			info := fmt.Sprintf("get health info failed. err: %v", err)
			w.Write([]byte(info))
			return
		}
		w.Write(js)
	})

	rootHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(`
<html>
<head><title>` + conf.ModuleName + ` Details </title></head>
<body>
<h1> ` + conf.ModuleName + ` Details </h1>
<p> Version: ` + version.BcsVersion + `</p>
<p> Tag: ` + version.BcsTag + `</p>
<p> Git Hash: ` + version.BcsGitHash + `</p>
<p> Build Hash: ` + version.BcsBuildTime + `</p>
<h2> Metrics</h2>
<p> <a href='/metrics'>/metrics</a> </p>
<h2> Healthz</h2>
<p> <a href='/healthz'>/healthz</a> </p>
</body>
</html>
						`))
	})

	mux.Handle("/", rootHandler)
	mux.Handle("/metrics", metricHandler)
	mux.Handle("/healthz", healthHandler)

	if err := listenAndServe(conf, mux); err != nil {
		return fmt.Errorf("listen and serve failed, err: %v", err)
	}
	return nil
}

func listenAndServe(c Config, mux http.Handler) error {
	addr := fmt.Sprintf("%s:%d", c.IP, c.MetricPort)

	if c.SvrCertFile == "" && c.SvrKeyFile == "" {
		go func() {
			blog.Infof("started metric and listen insecure server on %s", addr)
			blog.Fatal(http.ListenAndServe(addr, mux))
		}()
		return nil
	}

	// user https
	ca, err := ioutil.ReadFile(c.SvrCaFile)
	if nil != err {
		return err
	}
	capool := x509.NewCertPool()
	capool.AppendCertsFromPEM(ca)
	tlsconfig, err := ssl.ServerTslConfVerityClient(c.SvrCaFile,
		c.SvrCertFile,
		c.SvrKeyFile,
		c.SvrKeyPwd)
	if err != nil {
		return err
	}
	tlsconfig.BuildNameToCertificate()

	blog.Info("start metric secure serve on %s", addr)

	ln, err := net.Listen("tcp", net.JoinHostPort(c.IP, strconv.FormatUint(uint64(c.MetricPort), 10)))
	if err != nil {
		return err
	}
	listener := tls.NewListener(ln, tlsconfig)
	go func() {
		if err := http.Serve(listener, mux); nil != err {
			blog.Fatalf("server https failed. err: %v", err)
		}
	}()
	return nil
}
