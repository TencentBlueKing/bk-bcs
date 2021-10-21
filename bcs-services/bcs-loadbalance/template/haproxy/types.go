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

package haproxy

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	envHaproxyLogEnabled                = "LB_HAPROXY_ENABLE_LOG"
	envHaproxyLogLevel                  = "LB_HAPROXY_LOG_LEVEL"
	envHaproxySockPathName              = "LB_HAPROXY_SOCK_PATH"
	envHaproxyThreadNum                 = "LB_HAPROXY_THREADNUM"
	envHaproxyMaxConn                   = "LB_HAPROXY_MAX_CONN"
	envHaproxyPidPath                   = "LB_HAPROXY_PID_PATH"
	envHaproxySSLCert                   = "LB_HAPROXY_SSLCERT"
	envHaproxyRetries                   = "LB_HAPROXY_RETRY"
	envHaproxyBacklog                   = "LB_HAPROXY_BACKLOG"
	envHaproxyProxyMaxConn              = "LB_HAPROXY_PROXY_MAX_CONN"
	envHaproxyProxyTimeoutConnection    = "LB_HAPROXY_TIMEOUT_CONNECTION"
	envHaproxyProxyTimeoutClient        = "LB_HARPOXY_TIMEOUT_CLIENT"
	envHaproxyProxyTimeoutServer        = "LB_HAPROXY_TIMEOUT_SERVER"
	envHaproxyProxyTimeoutTunnel        = "LB_HAPROXY_TIMEOUT_TUNNEL"
	envHaproxyProxyTimeoutHTTPKeepAlive = "LB_HAPROXY_TIMEOUT_HTTP_KEEP_ALIVE"
	envHaproxyProxyTimeoutHTTPRequest   = "LB_HAPROXY_TIMEOUT_HTTP_REQUEST"
	envHaproxyProxyTimeoutQueue         = "LB_HAPROXY_TIMEOUT_QUEUE"
	envHaproxyProxyTimeoutTarpit        = "LB_HAPROXY_TIMEOUT_TARPIT"
	// options for both tcp backend and http backend
	// split by comma
	envHaproxyProxyOptions = "LB_HAPROXY_OPTIONS"
	// options only for http backend
	// split by comma
	envHaproxyHTTPProxyOptions = "LB_HAPROXY_HTTP_OPTIONS"

	// config for server health check
	envHaproxyServerHealthCheckInterval = "LB_HARPOXY_SERVER_HEALTH_CHECK_INTERVAL"
	envHaproxyServerRiseHealthCheckNum  = "LB_HAPROXY_SERVER_RISE_HEALTH_CHECK_NUM"
	envHaproxyServerFallHealthCheckNum  = "LB_HAPROXY_SERVER_FALL_HEALTH_CHECK_NUM"
	// config for lua stats page
	envHaproxyStatsFrontendPort         = "LB_HAPROXY_STATS_FRONTEND_PORT"
	envHaproxyStatsFrontendURI          = "LB_HAPROXY_STATS_FRONTEND_URI"
	envHaproxyStatsFrontendAuthUser     = "LB_HAPROXY_STATS_FRONTEND_AUTH_USER"
	envHaproxyStatsFrontendAuthPassword = "LB_HAPROXY_STATS_FRONTEND_AUTH_PASSWORD"

	// default config for configs
	defaultHaproxyLogEnabled                = 0
	defaultHaproxyLogLevel                  = "err"
	defaultHaproxySockPath                  = "/var/run/haproxy.sock"
	defaultHaproxyThreadNum                 = 4
	defaultHaproxyMaxConn                   = 302400
	defaultHaproxyPidPath                   = "/var/run/haproxy.pid"
	defaultHaproxyRetries                   = 1
	defaultHaproxyBacklog                   = 10000
	defaultHaproxyProxyMaxConn              = 202400
	defaultHaproxyProxyTimeoutConnection    = 3
	defaultHaproxyProxyTimeoutClient        = 15
	defaultHaproxyProxyTimeoutServer        = 15
	defaultHaproxyProxyTimeoutTunnel        = 3600
	defaultHaproxyProxyTimeoutHTTPKeepAlive = 60
	defaultProxyTimeoutHTTPRequest          = 15
	defaultHaproxyProxyTimeoutQueue         = 30
	defaultHaproxyProxyTimeoutTarpit        = 60
	defaultHaproxyProxyOptions              = "dontlognull,http-server-close,redispatch,srvtcpka,clitcpka"
	defaultHaproxyHTTPProxyOptions          = "httplog"
	defaultHaproxyServerHealthCheckInterval = 2000
	defaultHaproxyServerRiseHealthCheckNum  = 2
	defaultHaproxyServerFallHealthCheckNum  = 2
	defaultHaproxyStatsFrontendPort         = 8080
	defaultHaproxyStatsFrontendURI          = "/bcsadm?token=bcsteam"
	defaultHaproxyStatsFrontendAuthUser     = "bcsadmin"
	defaultHaproxyStatsFrontendAuthPassword = "Bcs1qaz2wsx"
)

// EnvConfig config for haproxy from env
type EnvConfig struct {
	LogEnabled                bool
	LogLevel                  string
	SockPath                  string
	ThreadNum                 int64
	MaxConn                   int64
	PidPath                   string
	SSLCert                   string
	Retries                   int64
	Backlog                   int64
	ProxyMaxConn              int64
	ProxyTimeoutConnection    int64
	ProxyTimeoutClient        int64
	ProxyTimeoutServer        int64
	ProxyTimeoutTunnel        int64
	ProxyTimeoutHTTPKeepAlive int64
	ProxyTimeoutHTTPRequest   int64
	ProxyTimeoutQueue         int64
	ProxyTimeoutTarpit        int64
	ProxyOptions              []string
	HTTPProxyOptions          []string
	ServerHealthCheckInterval int64
	ServerRiseHealthCheckNum  int64
	ServerFallHealthCheckNum  int64
	StatsFrontendPort         int64
	StatsFrontendURI          string
	StatsFrontendAuthUser     string
	StatsFrontendAuthPassword string
}

var defaultValueMap = map[string]int64{
	envHaproxyLogEnabled:                defaultHaproxyLogEnabled,
	envHaproxyThreadNum:                 defaultHaproxyThreadNum,
	envHaproxyMaxConn:                   defaultHaproxyMaxConn,
	envHaproxyRetries:                   defaultHaproxyRetries,
	envHaproxyBacklog:                   defaultHaproxyBacklog,
	envHaproxyProxyMaxConn:              defaultHaproxyProxyMaxConn,
	envHaproxyProxyTimeoutConnection:    defaultHaproxyProxyTimeoutConnection,
	envHaproxyProxyTimeoutClient:        defaultHaproxyProxyTimeoutClient,
	envHaproxyProxyTimeoutServer:        defaultHaproxyProxyTimeoutServer,
	envHaproxyProxyTimeoutTunnel:        defaultHaproxyProxyTimeoutTunnel,
	envHaproxyProxyTimeoutHTTPKeepAlive: defaultHaproxyProxyTimeoutHTTPKeepAlive,
	envHaproxyProxyTimeoutHTTPRequest:   defaultProxyTimeoutHTTPRequest,
	envHaproxyProxyTimeoutQueue:         defaultHaproxyProxyTimeoutQueue,
	envHaproxyProxyTimeoutTarpit:        defaultHaproxyProxyTimeoutTarpit,
	envHaproxyServerHealthCheckInterval: defaultHaproxyServerHealthCheckInterval,
	envHaproxyServerRiseHealthCheckNum:  defaultHaproxyServerRiseHealthCheckNum,
	envHaproxyServerFallHealthCheckNum:  defaultHaproxyServerFallHealthCheckNum,
	envHaproxyStatsFrontendPort:         defaultHaproxyStatsFrontendPort,
}

// loadNumEnv load number type config from env
func loadNumEnv(envName string) int64 {
	envValue := os.Getenv(envName)
	if len(envValue) != 0 {
		parsedValue, err := strconv.ParseInt(envValue, 10, 64)
		if err == nil {
			return parsedValue
		}
		blog.Warnf("parse %s failed, err %s", envName, err.Error())
	}
	return defaultValueMap[envName]
}

func loadEnvConfig() *EnvConfig {
	logEnabled := false
	logEnabledNum := loadNumEnv(envHaproxyLogEnabled)
	if logEnabledNum != 0 {
		logEnabled = true
	}
	logLevel := os.Getenv(envHaproxyLogLevel)
	if len(logLevel) == 0 {
		logLevel = defaultHaproxyLogLevel
	}
	sockPath := os.Getenv(envHaproxySockPathName)
	if len(sockPath) == 0 {
		sockPath = defaultHaproxySockPath
	}
	threadNum := loadNumEnv(envHaproxyThreadNum)
	maxConn := loadNumEnv(envHaproxyMaxConn)
	pidPath := os.Getenv(envHaproxyPidPath)
	if len(pidPath) == 0 {
		pidPath = defaultHaproxyPidPath
	}
	sslCert := os.Getenv(envHaproxySSLCert)
	retries := loadNumEnv(envHaproxyRetries)
	backlog := loadNumEnv(envHaproxyBacklog)
	proxyMaxConn := loadNumEnv(envHaproxyProxyMaxConn)
	proxyTimeoutConnection := loadNumEnv(envHaproxyProxyTimeoutConnection)
	proxyTimeoutClient := loadNumEnv(envHaproxyProxyTimeoutClient)
	proxyTimeoutServer := loadNumEnv(envHaproxyProxyTimeoutServer)
	proxyTimeoutTunnel := loadNumEnv(envHaproxyProxyTimeoutTunnel)
	proxyTimeoutHTTPKeepAlive := loadNumEnv(envHaproxyProxyTimeoutHTTPKeepAlive)
	proxyTimeoutHTTPRequest := loadNumEnv(envHaproxyProxyTimeoutHTTPRequest)
	proxyTimeoutQueue := loadNumEnv(envHaproxyProxyTimeoutQueue)
	proxyTimeoutTarpit := loadNumEnv(envHaproxyProxyTimeoutTarpit)
	var proxyOptions []string
	proxyOptionsStr := os.Getenv(envHaproxyProxyOptions)
	if len(proxyOptionsStr) == 0 {
		proxyOptionsStr = defaultHaproxyProxyOptions
	}
	proxyOptions = strings.Split(proxyOptionsStr, ",")
	httpProxyOptionsStr := os.Getenv(envHaproxyHTTPProxyOptions)
	if len(httpProxyOptionsStr) == 0 {
		httpProxyOptionsStr = defaultHaproxyHTTPProxyOptions
	}
	httpProxyOptions := strings.Split(httpProxyOptionsStr, ",")
	serverHealthCheckInterval := loadNumEnv(envHaproxyServerHealthCheckInterval)
	serverRiseHealthCheckNum := loadNumEnv(envHaproxyServerRiseHealthCheckNum)
	serverFallHealthCheckNum := loadNumEnv(envHaproxyServerFallHealthCheckNum)
	statsFrontendPort := loadNumEnv(envHaproxyStatsFrontendPort)
	statsFrontendURI := os.Getenv(envHaproxyStatsFrontendURI)
	if len(statsFrontendURI) == 0 {
		statsFrontendURI = defaultHaproxyStatsFrontendURI
	}
	statsFrontendAuthUser := os.Getenv(envHaproxyStatsFrontendAuthUser)
	if len(statsFrontendAuthUser) == 0 {
		statsFrontendAuthUser = defaultHaproxyStatsFrontendAuthUser
	}
	statsFrontendAuthPassword := os.Getenv(envHaproxyStatsFrontendAuthPassword)
	if len(statsFrontendAuthPassword) == 0 {
		statsFrontendAuthPassword = defaultHaproxyStatsFrontendAuthPassword
	}

	return &EnvConfig{
		LogEnabled:                logEnabled,
		LogLevel:                  logLevel,
		SockPath:                  sockPath,
		ThreadNum:                 threadNum,
		MaxConn:                   maxConn,
		PidPath:                   pidPath,
		SSLCert:                   sslCert,
		Retries:                   retries,
		Backlog:                   backlog,
		ProxyMaxConn:              proxyMaxConn,
		ProxyTimeoutConnection:    proxyTimeoutConnection,
		ProxyTimeoutClient:        proxyTimeoutClient,
		ProxyTimeoutServer:        proxyTimeoutServer,
		ProxyTimeoutTunnel:        proxyTimeoutTunnel,
		ProxyTimeoutHTTPKeepAlive: proxyTimeoutHTTPKeepAlive,
		ProxyTimeoutHTTPRequest:   proxyTimeoutHTTPRequest,
		ProxyTimeoutQueue:         proxyTimeoutQueue,
		ProxyTimeoutTarpit:        proxyTimeoutTarpit,
		ProxyOptions:              proxyOptions,
		HTTPProxyOptions:          httpProxyOptions,
		ServerHealthCheckInterval: serverHealthCheckInterval,
		ServerRiseHealthCheckNum:  serverRiseHealthCheckNum,
		ServerFallHealthCheckNum:  serverFallHealthCheckNum,
		StatsFrontendPort:         statsFrontendPort,
		StatsFrontendURI:          statsFrontendURI,
		StatsFrontendAuthUser:     statsFrontendAuthUser,
		StatsFrontendAuthPassword: statsFrontendAuthPassword,
	}
}

// Config data structure for haproxy config file
type Config struct {
	LogEnabled         bool
	LogLevel           string
	SockPath           string
	ThreadNum          int64
	MaxConn            int64
	PidPath            string
	SSLCert            string
	DefaultProxyConfig *ProxyDefault
	Stats              *StatsFrontend
	// map[ServicePort]
	HTTPMap  map[int]*HTTPFrontend
	HTTPSMap map[int]*HTTPFrontend
	TCPMap   map[int]*TCPListener
	// data for render template in order
	HTTPList  HTTPFrontendList
	HTTPSList HTTPFrontendList
	TCPList   TCPListenerList
}

// generateRenderData generate frontend list from frontend map, ensure the rendered data is always in order
func (c *Config) generateRenderData() {
	c.HTTPList = nil
	for _, frontend := range c.HTTPMap {
		frontend.generateRenderData()
		c.HTTPList = append(c.HTTPList, frontend)
	}
	sort.Sort(c.HTTPList)

	c.HTTPSList = nil
	for _, frontend := range c.HTTPSMap {
		frontend.generateRenderData()
		c.HTTPSList = append(c.HTTPSList, frontend)
	}
	sort.Sort(c.HTTPSList)

	c.TCPList = nil
	for _, listener := range c.TCPMap {
		listener.generateRenderData()
		c.TCPList = append(c.TCPList, listener)
	}
	sort.Sort(c.TCPList)
}

// generateServerName generate real server name from ordered ips of servers
func (c *Config) generateServerName() {
	// create server name for each server
	for _, frontend := range c.HTTPMap {
		for _, backend := range frontend.Backends {
			var tmpList IPRealServerList
			for _, server := range backend.Servers {
				tmpList = append(tmpList, server)
			}
			// sort by ip
			sort.Sort(tmpList)
			for index, server := range tmpList {
				server.Name = getServerName(backend.Name, index)
			}
		}
	}
	for _, frontend := range c.HTTPSMap {
		for _, backend := range frontend.Backends {
			var tmpList IPRealServerList
			for _, server := range backend.Servers {
				tmpList = append(tmpList, server)
			}
			// sort by ip
			sort.Sort(tmpList)
			for index, server := range tmpList {
				server.Name = getServerName(backend.Name, index)
			}
		}
	}
	for _, listener := range c.TCPMap {
		var tmpList IPRealServerList
		for _, server := range listener.Servers {
			tmpList = append(tmpList, server)
		}
		// sort by ip
		sort.Sort(tmpList)
		for index, server := range tmpList {
			server.Name = getServerName(listener.Name, index)
		}
	}
}

// ProxyDefault default config for proxy
type ProxyDefault struct {
	Retries              int64
	Backlog              int64
	MaxConn              int64
	TimeoutConnection    int64
	TimeoutClient        int64
	TimeoutServer        int64
	TimeoutTunnel        int64
	TimeoutHTTPKeepAlive int64
	TimeoutHTTPRequest   int64
	TimeoutQueue         int64
	TimeoutTarpit        int64
	Options              []string
	HTTPOptions          []string
}

// StatsFrontend stats for frontend
type StatsFrontend struct {
	Port         int64
	URI          string
	AuthUser     string
	AuthPassword string
}

// HTTPFrontend http frontend
type HTTPFrontend struct {
	ServicePort int
	Name        string
	// map[ServiceName]
	Backends map[string]*HTTPBackend
	// data for render template in order
	BackendList HTTPBackendList
}

// HTTPFrontendList http frontend list, sorted by service port
type HTTPFrontendList []*HTTPFrontend

func (hl HTTPFrontendList) Len() int {
	return len(hl)
}

func (hl HTTPFrontendList) Less(i, j int) bool {
	return hl[i].ServicePort < hl[j].ServicePort
}

func (hl HTTPFrontendList) Swap(i, j int) {
	hl[i], hl[j] = hl[j], hl[i]
}

func (hf *HTTPFrontend) generateRenderData() {
	hf.BackendList = nil
	for _, backend := range hf.Backends {
		backend.generateRenderData()
		hf.BackendList = append(hf.BackendList, backend)
	}
	sort.Sort(hf.BackendList)
}

// HTTPBackend http backend
type HTTPBackend struct {
	Name                string
	Domain              string
	URL                 string
	Balance             string
	HealthCheckInterval int64
	RiseHealthCheckNum  int64
	FallHealthCheckNum  int64
	// map[IP]
	Servers map[string]*RealServer
	// data for render template in order
	ServerList RealServerList
}

// HTTPBackendList http backend list
type HTTPBackendList []*HTTPBackend

func (hl HTTPBackendList) Len() int {
	return len(hl)
}

func (hl HTTPBackendList) Less(i, j int) bool {
	return hl[i].Name < hl[j].Name
}

func (hl HTTPBackendList) Swap(i, j int) {
	hl[i], hl[j] = hl[j], hl[i]
}

func (hb *HTTPBackend) generateRenderData() {
	hb.ServerList = nil
	for _, rs := range hb.Servers {
		hb.ServerList = append(hb.ServerList, rs)
	}
	sort.Sort(hb.ServerList)
}

// TCPListener tcp frontend
type TCPListener struct {
	ServicePort         int
	Name                string
	Balance             string
	HealthCheckInterval int64
	RiseHealthCheckNum  int64
	FallHealthCheckNum  int64
	// map[IP]
	Servers map[string]*RealServer
	// data for render template in order
	ServerList RealServerList
}

// TCPListenerList tcp listener list for sort
type TCPListenerList []*TCPListener

func (hl TCPListenerList) Len() int {
	return len(hl)
}

func (hl TCPListenerList) Less(i, j int) bool {
	return hl[i].Name < hl[j].Name
}

func (hl TCPListenerList) Swap(i, j int) {
	hl[i], hl[j] = hl[j], hl[i]
}

func (tl *TCPListener) generateRenderData() {
	tl.ServerList = nil
	for _, rs := range tl.Servers {
		tl.ServerList = append(tl.ServerList, rs)
	}
	sort.Sort(tl.ServerList)
}

// RealServer backend config
type RealServer struct {
	Name     string
	IP       string
	Port     int
	Weight   int
	Disabled bool
}

// Key generate the map key for real server in backend
func (rs *RealServer) Key() string {
	return rs.IP + ":" + strconv.Itoa(rs.Port)
}

// RealServerList real server list, sorted by name
type RealServerList []*RealServer

func (hl RealServerList) Len() int {
	return len(hl)
}

func (hl RealServerList) Less(i, j int) bool {
	return hl[i].Name < hl[j].Name
}

func (hl RealServerList) Swap(i, j int) {
	hl[i], hl[j] = hl[j], hl[i]
}

// IPRealServerList real server list, sorted by ip
type IPRealServerList []*RealServer

func (hl IPRealServerList) Len() int {
	return len(hl)
}

func (hl IPRealServerList) Less(i, j int) bool {
	return hl[i].IP < hl[j].IP
}

func (hl IPRealServerList) Swap(i, j int) {
	hl[i], hl[j] = hl[j], hl[i]
}
