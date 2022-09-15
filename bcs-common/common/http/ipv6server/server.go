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

package ipv6server

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

var ErrListenNull = errors.New(
	"in the ListenAndServe method of IPv6Server, the Listen() method returns []net.Listener cannot be empty",
)

// IPv6Server ipv6 server
// 兼容ipv6 和 ipv4地址
type IPv6Server struct {
	// http.Server
	*http.Server
	// address ip地址
	address []string
	// port 端口号
	port string
	// network 指定监听网络类型，如：“tcp”、“tcp4”、“tcp6”，默认为：”tcp“
	network string
}

// verifyIp 验证IP
func (s *IPv6Server) verifyIp() (ips []string) {
	ips = make([]string, 0)
	for _, v := range s.address {
		if net.ParseIP(v) != nil {
			ips = append(ips, v)
		}
	}
	return ips
}

// joinHostPort 拼接ip和port
func (s *IPv6Server) joinHostPort() (ips []string) {
	ips = s.verifyIp()
	for i, v := range ips {
		ips[i] = net.JoinHostPort(v, s.port)
	}
	return ips
}

// Listen 监听一组IP
func (s *IPv6Server) Listen() (listeners []net.Listener, err error) {
	for _, v := range s.joinHostPort() {
		listen, err := net.Listen(s.network, v)
		if err != nil {
			if strings.Contains(err.Error(), "bind: cannot assign requested address") {
				// 单栈环境，会出现该错误
				blog.Warn("unable to listen %s, err: %s", v, err.Error())
				continue
			}
			return nil, err
		}
		listeners = append(listeners, listen)
		blog.Info("listen %s", v)
	}
	return listeners, nil
}

// SetHttpServerTLSConfig http.Server.TLSConfig
func (s *IPv6Server) SetHttpServerTLSConfig(tlsConfig *tls.Config) {
	s.Server.TLSConfig = tlsConfig
}

// SetHttpServerHandler http.Server.Handler
func (s *IPv6Server) SetHttpServerHandler(handler http.Handler) {
	s.Server.Handler = handler
}

// SetHttpServerAddr http.Server.addr
func (s *IPv6Server) SetHttpServerAddr(addr string) {
	s.Server.Addr = addr
}

// SetAddress address
func (s *IPv6Server) SetAddress(address []string) {
	s.address = address
}

// SetPort port
func (s *IPv6Server) SetPort(port string) {
	s.port = port
}

// SetNetwork network
func (s *IPv6Server) SetNetwork(network string) {
	s.network = network
}

// ListenAndServe 监听 address，并启动web server
// 与StartWebServer作用一样
func (s *IPv6Server) ListenAndServe() error {
	listeners, err := s.Listen()
	if err != nil {
		return err
	}
	if len(listeners) == 0 {
		return ErrListenNull
	}
	errs := make(chan error, 1)
	defer close(errs)
	for _, v := range listeners {
		go func(listen net.Listener) {
			errs <- s.Server.Serve(listen)
		}(v)
	}
	return <-errs
}

// StartWebServer 监听 address，并启动web server。
// 与ListenAndServe作用一样
func (s *IPv6Server) StartWebServer() error {
	return s.ListenAndServe()
}

// Close immediately closes all active net.Listeners and any
// connections in state StateNew, StateActive, or StateIdle. For a
// graceful shutdown, use Shutdown.
//
// Close does not attempt to close (and does not even know about)
// any hijacked connections, such as WebSockets.
//
// Close returns any error returned from closing the Server's
// underlying Listener(s).
func (s *IPv6Server) Close() error {
	return s.Server.Close()
}

// RegisterOnShutdown registers a function to call on Shutdown.
// This can be used to gracefully shutdown connections that have
// undergone ALPN protocol upgrade or that have been hijacked.
// This function should start protocol-specific graceful shutdown,
// but should not wait for shutdown to complete.
func (s *IPv6Server) RegisterOnShutdown(f func()) {
	s.Server.RegisterOnShutdown(f)
}

// Shutdown gracefully shuts down the server without interrupting any
// active connections. Shutdown works by first closing all open
// listeners, then closing all idle connections, and then waiting
// indefinitely for connections to return to idle and then shut down.
// If the provided context expires before the shutdown is complete,
// Shutdown returns the context's error, otherwise it returns any
// error returned from closing the Server's underlying Listener(s).
//
// When Shutdown is called, Serve, ListenAndServe, and
// ListenAndServeTLS immediately return ErrServerClosed. Make sure the
// program doesn't exit and waits instead for Shutdown to return.
//
// Shutdown does not attempt to close nor wait for hijacked
// connections such as WebSockets. The caller of Shutdown should
// separately notify such long-lived connections of shutdown and wait
// for them to close, if desired. See RegisterOnShutdown for a way to
// register shutdown notification functions.
//
// Once Shutdown has been called on a server, it may not be reused;
// future calls to methods such as Serve will return ErrServerClosed.
func (s *IPv6Server) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}

// Serve accepts incoming connections on the Listener l, creating a
// new service goroutine for each. The service goroutines read requests and
// then call srv.Handler to reply to them.
//
// HTTP/2 support is only enabled if the Listener returns *tls.Conn
// connections and they were configured with "h2" in the TLS
// Config.NextProtos.
//
// Serve always returns a non-nil error and closes l.
// After Shutdown or Close, the returned error is ErrServerClosed.
func (s *IPv6Server) Serve(l net.Listener) error {
	return s.Server.Serve(l)
}

// ServeTLS accepts incoming connections on the Listener l, creating a
// new service goroutine for each. The service goroutines perform TLS
// setup and then read requests, calling srv.Handler to reply to them.
//
// Files containing a certificate and matching private key for the
// server must be provided if neither the Server's
// TLSConfig.Certificates nor TLSConfig.GetCertificate are populated.
// If the certificate is signed by a certificate authority, the
// certFile should be the concatenation of the server's certificate,
// any intermediates, and the CA's certificate.
//
// ServeTLS always returns a non-nil error. After Shutdown or Close, the
// returned error is ErrServerClosed.
func (s *IPv6Server) ServeTLS(l net.Listener, certFile, keyFile string) error {
	return s.Server.ServeTLS(l, certFile, keyFile)
}

// SetKeepAlivesEnabled controls whether HTTP keep-alives are enabled.
// By default, keep-alives are always enabled. Only very
// resource-constrained environments or servers in the process of
// shutting down should disable them.
func (s *IPv6Server) SetKeepAlivesEnabled(v bool) {
	s.Server.SetKeepAlivesEnabled(v)
}

// ListenAndServeTLS listens on the TCP network address srv.Addr and
// then calls ServeTLS to handle requests on incoming TLS connections.
// Accepted connections are configured to enable TCP keep-alives.
//
// Filenames containing a certificate and matching private key for the
// server must be provided if neither the Server's TLSConfig.Certificates
// nor TLSConfig.GetCertificate are populated. If the certificate is
// signed by a certificate authority, the certFile should be the
// concatenation of the server's certificate, any intermediates, and
// the CA's certificate.
//
// If srv.Addr is blank, ":https" is used.
//
// ListenAndServeTLS always returns a non-nil error. After Shutdown or
// Close, the returned error is ErrServerClosed.
func (s *IPv6Server) ListenAndServeTLS(certFile, keyFile string) error {
	listeners, err := s.Listen()
	if err != nil {
		return err
	}
	if len(listeners) == 0 {
		return ErrListenNull
	}
	errs := make(chan error, 1)
	defer close(errs)
	for _, v := range listeners {
		go func(listen net.Listener) {
			defer listen.Close()
			errs <- s.Server.ServeTLS(listen, certFile, keyFile)
		}(v)
	}
	return <-errs
}

// Serve accepts incoming HTTP connections on the listener l,
// creating a new service goroutine for each. The service goroutines
// read requests and then call handler to reply to them.
//
// The handler is typically nil, in which case the DefaultServeMux is used.
//
// HTTP/2 support is only enabled if the Listener returns *tls.Conn
// connections and they were configured with "h2" in the TLS
// Config.NextProtos.
//
// Serve always returns a non-nil error.
func Serve(l net.Listener, handler http.Handler) error {
	srv := &IPv6Server{
		Server: &http.Server{
			Handler: handler,
		},
	}
	return srv.Serve(l)
}

// ServeTLS accepts incoming HTTPS connections on the listener l,
// creating a new service goroutine for each. The service goroutines
// read requests and then call handler to reply to them.
//
// The handler is typically nil, in which case the DefaultServeMux is used.
//
// Additionally, files containing a certificate and matching private key
// for the server must be provided. If the certificate is signed by a
// certificate authority, the certFile should be the concatenation
// of the server's certificate, any intermediates, and the CA's certificate.
//
// ServeTLS always returns a non-nil error.
func ServeTLS(l net.Listener, handler http.Handler, certFile, keyFile string) error {
	srv := &IPv6Server{
		Server: &http.Server{
			Handler: handler,
		},
	}
	return srv.ServeTLS(l, certFile, keyFile)
}

// ListenAndServe listens on the TCP network address addr and then calls
// Serve with handler to handle requests on incoming connections.
// Accepted connections are configured to enable TCP keep-alives.
//
// The handler is typically nil, in which case the DefaultServeMux is used.
//
// ListenAndServe always returns a non-nil error.
func ListenAndServe(addr string, handler http.Handler) error {
	server := &IPv6Server{
		Server: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
	return server.Server.ListenAndServe()
}

// ListenAndServeTLS acts identically to ListenAndServe, except that it
// expects HTTPS connections. Additionally, files containing a certificate and
// matching private key for the server must be provided. If the certificate
// is signed by a certificate authority, the certFile should be the concatenation
// of the server's certificate, any intermediates, and the CA's certificate.
func ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler) error {
	server := &IPv6Server{
		Server: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
	return server.Server.ListenAndServeTLS(certFile, keyFile)
}

// NewParameterlessIPv6Server 创建无参 IPv6Server
func NewParameterlessIPv6Server() *IPv6Server {
	return newIPv6Server(nil, "", "", nil, nil)
}

// NewNoneHandlerIPv6Server 创建 无Handler的IPv6Server
func NewNoneHandlerIPv6Server(address []string, port, network string) *IPv6Server {
	return newIPv6Server(address, port, network, nil, nil)
}

// NewIPv6Server 创建 普通 IPv6Server
func NewIPv6Server(address []string, port, network string, handler http.Handler) *IPv6Server {
	return newIPv6Server(address, port, network, nil, handler)
}

// NewTlsIPv6Server 创建 TLS IPv6Server
func NewTlsIPv6Server(address []string, port, network string, tlsConfig *tls.Config, handler http.Handler) *IPv6Server {
	return newIPv6Server(address, port, network, tlsConfig, handler)
}

// newIPv6Server 创建IPv6Server
func newIPv6Server(address []string, port, network string, tlsConfig *tls.Config, handler http.Handler) *IPv6Server {
	if len(address) == 0 {
		address = append(address, types.IPv4Loopback)
	}
	if network == "" {
		network = types.TCP
	}
	if port == "" {
		port = types.DefaultPort
	}
	return &IPv6Server{
		Server: &http.Server{
			Handler:   handler,
			TLSConfig: tlsConfig,
		},
		address: address,
		port:    port,
		network: network,
	}
}
