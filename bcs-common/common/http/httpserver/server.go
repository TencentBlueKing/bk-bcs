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

package httpserver

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"strconv"
)

// HttpServer is data struct of http server
type HttpServer struct {
	addr         string
	port         uint
	insecureAddr string
	insecurePort uint
	sock         string
	isSSL        bool
	caFile       string
	certFile     string
	keyFile      string
	certPasswd   string
	webContainer *restful.Container
	router       *mux.Router
}

// NewHttpServer init httpServer
func NewHttpServer(port uint, addr, sock string) *HttpServer {
	return &HttpServer{
		addr:         addr,
		port:         port,
		sock:         sock,
		webContainer: restful.NewContainer(),
		router:       mux.NewRouter(),
		isSSL:        false,
	}
}

// SetInsecureServer set insecureAddr & insecurePort
func (s *HttpServer) SetInsecureServer(insecureAddr string, insecurePort uint) {
	s.insecureAddr = insecureAddr
	s.insecurePort = insecurePort
}

// GetWebContainer get httpServer webContainer
func (s *HttpServer) GetWebContainer() *restful.Container {
	return s.webContainer
}

// GetRouter get httpServer router
func (s *HttpServer) GetRouter() *mux.Router {
	return s.router
}

// SetSsl set http ssl
func (s *HttpServer) SetSsl(cafile, certfile, keyfile, certPasswd string) {
	s.caFile = cafile
	s.certFile = certfile
	s.keyFile = keyfile
	s.certPasswd = certPasswd
	s.isSSL = true
}

// RegisterWebServer register http webserver
func (s *HttpServer) RegisterWebServer(rootPath string, filters []restful.FilterFunction, actions []*Action) error {
	//new a web service
	ws := s.NewWebService(rootPath, filters)

	//register action
	s.RegisterActions(ws, actions)

	return nil
}

// RegisterApiDocs register api docs
func (s *HttpServer) RegisterApiDocs(apidocsPath string) error {
	if apidocsPath == "" {
		apidocsPath = "/apidocs.json"
	}
	config := restfulspec.Config{
		WebServices: s.webContainer.RegisteredWebServices(),
		APIPath:     apidocsPath,
	}
	s.webContainer.Add(restfulspec.NewOpenAPIService(config))
	return nil
}

// NewWebService set http webService
func (s *HttpServer) NewWebService(rootPath string, filters []restful.FilterFunction) *restful.WebService {
	ws := new(restful.WebService)
	if "" != rootPath {
		ws.Path(rootPath)
	}

	ws.Produces(restful.MIME_JSON, restful.MIME_XML, restful.MIME_OCTET)

	if len(filters) != 0 {
		for i := range filters {
			ws.Filter(filters[i])
		}
	}

	s.webContainer.Add(ws)

	return ws
}

// RegisterActions register actions
func (s *HttpServer) RegisterActions(ws *restful.WebService, actions []*Action) {
	for _, action := range actions {
		switch action.Verb {
		case "POST":
			route := ws.POST(action.Path).To(action.Handler)
			ws.Route(route)
			blog.Infof("register post api, url(%s)", action.Path)
		case "GET":
			route := ws.GET(action.Path).To(action.Handler)
			ws.Route(route)
			blog.Infof("register get api, url(%s)", action.Path)
		case "PUT":
			route := ws.PUT(action.Path).To(action.Handler)
			ws.Route(route)
			blog.Infof("register put api, url(%s)", action.Path)
		case "DELETE":
			route := ws.DELETE(action.Path).To(action.Handler)
			ws.Route(route)
			blog.Infof("register delete api, url(%s)", action.Path)
		case "PATCH":
			route := ws.PATCH(action.Path).To(action.Handler)
			ws.Route(route)
			blog.Infof("register patch api, url(%s)", action.Path)
		default:
			blog.Error("unrecognized action verb: %s", action.Verb)
		}
	}
}

// ListenAndServe listen httpServer
func (s *HttpServer) ListenAndServe() error {

	var chError = make(chan error)
	//list and serve by addrport
	go func() {
		addrport := net.JoinHostPort(s.addr, strconv.FormatUint(uint64(s.port), 10))
		httpserver := &http.Server{Addr: addrport, Handler: s.webContainer}
		if s.isSSL {
			tlsConf, err := ssl.ServerTslConf(s.caFile, s.certFile, s.keyFile, s.certPasswd)
			if err != nil {
				blog.Error("fail to load certfile, err:%s", err.Error())
				chError <- fmt.Errorf("fail to load certfile")
				return
			}
			httpserver.TLSConfig = tlsConf
			blog.Info("Start https service on(%s:%d)", s.addr, s.port)
			chError <- httpserver.ListenAndServeTLS("", "")
		} else {
			blog.Info("Start http service on(%s:%d)", s.addr, s.port)
			chError <- httpserver.ListenAndServe()
		}
	}()

	return <-chError
}

// ListenAndServeMux  listen httpServer by serverMux
func (s *HttpServer) ListenAndServeMux(verifyClientTLS bool) error {

	//list and serve by addrport
	if s.isSSL {
		addrport := net.JoinHostPort(s.addr, strconv.FormatUint(uint64(s.port), 10))
		httpserver := &http.Server{Addr: addrport, Handler: s.router}

		// listen to https single certification
		tlsConf, err := ssl.ServerTslConfVerity(s.certFile, s.keyFile, s.certPasswd)
		if verifyClientTLS {
			tlsConf, err = ssl.ServerTslConfVerityClient(s.caFile, s.certFile, s.keyFile, s.certPasswd)
		}
		if err != nil {
			return fmt.Errorf("fail to load certfile, err:%s", err.Error())
		}
		httpserver.TLSConfig = tlsConf
		blog.Info("Start https service on(%s:%d)", s.addr, s.port)
		go func() {
			err := httpserver.ListenAndServeTLS("", "")
			fmt.Printf("tls server failed: %v\n", err)
		}()
	}
	if s.insecureAddr != "" && s.insecurePort != 0 {
		addrport := net.JoinHostPort(s.insecureAddr, strconv.FormatUint(uint64(s.insecurePort), 10))
		httpserver := &http.Server{Addr: addrport, Handler: s.router}

		blog.Info("Start http service on(%s:%d)", s.insecureAddr, s.insecurePort)
		go func() {
			err := httpserver.ListenAndServe()
			fmt.Printf("insecure server failed: %v\n", err)
		}()
	}

	return nil
}

// Serve serve httpServer
func (s *HttpServer) Serve(l net.Listener) error {

	var chError = make(chan error)
	//list and serve by addrport
	go func() {
		httpserver := &http.Server{Handler: s.webContainer}
		if s.isSSL {
			tlsConf, err := ssl.ServerTslConf(s.caFile, s.certFile, s.keyFile, s.certPasswd)
			if err != nil {
				blog.Error("fail to load certfile, err:%s", err.Error())
				chError <- fmt.Errorf("fail to load certfile")
				return
			}
			httpserver.TLSConfig = tlsConf
			blog.Info("Start https service on(%s:%d)", s.addr, s.port)
			chError <- httpserver.ServeTLS(l, "", "")
		} else {
			blog.Info("Start http service on(%s:%d)", s.addr, s.port)
			chError <- httpserver.Serve(l)
		}
	}()

	return <-chError
}
