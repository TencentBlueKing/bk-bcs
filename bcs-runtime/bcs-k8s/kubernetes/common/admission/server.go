/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package admission

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"k8s.io/klog"
)

// ServerRunOptions Server Run Options
type ServerRunOptions struct {
	WebhookAddress string
	WebhookPort    int
	TLSCA          string
	TLSCert        string
	TLSKey         string
}

// NewServerRunOptions New Server Run Options
func NewServerRunOptions() *ServerRunOptions {
	options := &ServerRunOptions{}
	return options
}

// Validate validate
func (s *ServerRunOptions) Validate() error {
	address := net.ParseIP(s.WebhookAddress)
	if address.To4() == nil {
		return fmt.Errorf("%v is not a valid IP address", s.WebhookAddress)
	}
	return nil
}

func getTLSConfig(s *ServerRunOptions) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		NextProtos: []string{"http/1.1"},
		//		Certificates: []tls.Certificate{cert},
		// Avoid fallback on insecure SSL protocols
		MinVersion: tls.VersionTLS10,
	}
	if s.TLSCA != "" {
		certPool := x509.NewCertPool()
		file, err := ioutil.ReadFile(s.TLSCA)
		if err != nil {
			return nil, fmt.Errorf("Could not read CA certificate: %v", err)
		}
		certPool.AppendCertsFromPEM(file)
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = certPool
	}

	return tlsConfig, nil
}

// Run run
func Run(s *ServerRunOptions, serverType string, client interface{}, stopCh <-chan struct{}) error {

	webHook, err := NewWebhookServer(client, serverType)
	if err != nil {
		return err
	}

	// Start debug monitor.
	mux := http.NewServeMux()
	mux.HandleFunc("/validate-crd", webHook.ServeCRD)
	mux.HandleFunc("/validate-workload", webHook.ServeWorkload)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", "ok")
	})

	server := &http.Server{
		Addr:         net.JoinHostPort(s.WebhookAddress, strconv.Itoa(s.WebhookPort)),
		Handler:      mux,
		ReadTimeout:  300 * time.Second,
		WriteTimeout: 300 * time.Second,
	}

	klog.V(1).Infof("listening on %v", server.Addr)
	if s.TLSCert != "" && s.TLSKey != "" {
		klog.V(1).Infof("using HTTPS service")
		tlsConfig, err := getTLSConfig(s)
		if err != nil {
			return err
		}
		server.TLSConfig = tlsConfig
		go func() {
			klog.Fatal(server.ListenAndServeTLS(s.TLSCert, s.TLSKey))
		}()
	} else {
		go func() {
			klog.V(1).Infof("using HTTP service")
			klog.Fatal(server.ListenAndServe())
		}()
	}

	select {
	case <-stopCh:
		klog.Info("http server received stop signal, waiting for all requests to finish")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			klog.Error(err)
		}
	}
	return nil
}
