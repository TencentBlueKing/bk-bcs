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

package register

import "fmt"

const (
	//ProtocolHTTP for http
	ProtocolHTTP = "http"
	//ProtocolHTTPS for http
	ProtocolHTTPS = "https"
	//ProtocolTCP for tcp
	ProtocolTCP = "tcp"
	//ProtocolUDP for udp
	ProtocolUDP = "udp"
	//ProtocolGrpc for grpc
	ProtocolGrpc = "grpc"
)

//Route inner data structure for traffics transffer.
// this model is used for frontend listenning or register
type Route struct {
	//Name for route
	Name string
	//Prototol for frontend listenning, such as tcp, udp, http(s)
	Protocol string
	//Port for listen, if port is 0, use specified default tcp/udp/http port
	Port uint
	//Paths filter when protocol is http(s)
	Paths []string
	//PathRewrite rewrite path for http traffic
	PathRewrite bool
	//Header filter when using http(s)
	Header map[string]string
	//plugin Option for http modification
	Plugin *Plugins
	//Service relative svc name
	Service string
	Labels  map[string]string
}

//Service inner data structure for backend service
type Service struct {
	Name      string
	Protocol  string
	Host      string
	Port      uint
	Retries   int
	Path      string
	Algorithm string
	//Option for plugin
	Plugin *Plugins
	//Routes several route can redirect traffics to same service
	Routes   []Route
	Backends []Backend
	Labels   map[string]string
}

//Valid check service information
func (s *Service) Valid() error {
	if len(s.Name) == 0 {
		return fmt.Errorf("service name required")
	}
	if len(s.Protocol) == 0 {
		s.Protocol = ProtocolHTTP
	}
	if s.Retries == 0 {
		s.Retries = 1
	}
	if len(s.Routes) == 0 {
		return fmt.Errorf("route required")
	}
	if len(s.Backends) == 0 {
		return fmt.Errorf("backend required")
	}
	return nil
}

//Plugins holder for all gateway plugins
type Plugins struct {
	HeadOption *HeaderOption
	AuthOption *BCSAuthOption
}

//HeaderOption for proxy rules that change http header
type HeaderOption struct {
	//clean specified header
	Clean []string
	//Add more values
	Add map[string]string
	//Replace specified header
	Replace map[string]string
}

//BCSAuthOption for bkbcs-auth plugin
type BCSAuthOption struct {
	Name          string
	AuthEndpoints string
	AuthToken     string
	Module        string
}

//Backend inner data structure for application instance
type Backend struct {
	Target string
	Weight int
}
