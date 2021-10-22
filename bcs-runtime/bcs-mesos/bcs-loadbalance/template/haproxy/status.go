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
	"bytes"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// LoadbalanceHaproxyStatsFetchStateMetric loadbalance metric for zookeeper connection
	LoadbalanceHaproxyStatsFetchStateMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "loadbalance",
			Subsystem: "haproxy",
			Name:      "fetchstats_state",
			Help:      "the state for haproxy stats fetching state, 0 for abnormal, 1 for normal",
		},
	)
)

func init() {
	prometheus.Register(LoadbalanceHaproxyStatsFetchStateMetric)
	LoadbalanceHaproxyStatsFetchStateMetric.Set(1)
}

// https://cbonte.github.io/haproxy-dconv/1.7/management.html#9.1
const (
	proxyNameHaproxy int = iota
	serviceNameHaproxy
	queueCurrentHaproxy
	queueMaxHaproxy
	sessionCurrentHaproxy
	sessionMaxHaproxy
	sessionLimitHaproxy
	sessionTotalHaproxy
	bytesInHaproxy
	bytesOutHaproxy
	deniedRequestHaproxy  // requests denied because of security concerns
	deniedResponseHaproxy // responses denied because of security concerns
	// - early termination from the client, before the request has been sent.
	// - read error from the client
	// - client timeout
	// - client closed connection
	// - various bad requests from the client.
	// - request was tarpitted.
	errorRequestHaproxy // request errors. Some of the possible causes are:
	errorConnectHaproxy
	errorResponseHaproxy
	wConnectRetryHaproxy
	wConnectRedispatchHaproxy
	statusHaproxy // (UP/DOWN/NOLB/MAINT/MAINT(via)/MAINT(resolution)...)
	weightHaproxy
	activeServerNumHaproxy
	backupServerNumHaproxy
	checkFailHaproxy
	checkDownHaproxy
	secondSinceLastCheckDownHaproxy // number of seconds since the last UP<->DOWN transition
	downTimeTotalHaproxy
	queueMaxLimitHaproxy
	pidHaproxy
	uniqueProxyIDdHaproxy
	serverIDHaproxy
	throttlePercentageCurrentHaproxy
	lbTotalTimesSelectedHaproxy // total number of times a server was selected, either for new sessions, or when re-dispatching. The server counter is the number of times that server was selected.
	trackedHaproxy              // id of proxy/server if tracking is enabled.
	typeHaproxy                 // (0=frontend, 1=backend, 2=server, 3=socket/listener)
	sessionRateHaproxy          // number of sessions per second over last elapsed second
	sessionRateLimHaproxy       // configured limit on new sessions per second
	sessionRateMaxHaproxy       // max number of new sessions per second
	checkStatusHaproxy
	checkCodeHaproxy
	checkDurationHaproxy
	httpResponse1xxHaproxy
	httpResponse2xxHaproxy
	httpResponse3xxHaproxy
	httpResponse4xxHaproxy
	httpResponse5xxHaproxy
	httpResponseOtherHaproxy
	hanaFailHaproxy // failed health checks details
	requestRateHaproxy
	requestRateMaxHaproxy
	requestTotalHaproxy
	cliAbortHaproxy       // number of data transfers aborted by the client
	srvAbortHaproxy       // number of data transfers aborted by the server
	compByteInHaproxy     // number of HTTP response bytes fed to the compressor
	compByteOutHaproxy    // number of HTTP response bytes fed to the compressor
	compByteBypassHaproxy // number of bytes that bypassed the HTTP compressor
	compByteRsponseHaproxy
	lastSesseionSecondHaproxy
	lastCheckContentHaproxy
	lastAgentCheckContentHaproxy
	queueTimeOver1024RequestHaproxy
	connectTimeOver1024RequestHaproxy
	responseTimeOver1024RequestHaproxy
	totalTimeOver1024RequestHaproxy
	agentStatusHaproxy
	agentCodeHaproxy
	agentDurationHaproxy
	checkDescHaproxy
	agentDescHaproxy
	checkRiseHaproxy        // server's "rise" parameter used by checks
	checkFallHaproxy        // server's "fall" parameter used by checks
	checkHealthHaproxy      // server's health check value between 0 and rise+fall-1
	agentRiseHaproxy        // agent's "rise" parameter, normally 1
	agentFallHaproxy        // agent's "fall" parameter, normally 1
	agentHealthHaproxy      // agent's health parameter, between 0 and rise+fall-1
	addrHaproxy             // address:port
	cookieHaproxy           // server's cookie value or backend's cookie name
	modeHaproxy             // proxy mode (tcp, http, health, unknown)
	algorithmLBHaproxy      // load balancing algorithm
	connRateHaproxy         // number of connections over the last elapsed second
	connRateMaxHaproxy      // highest known connRateHaproxy
	connTotalHaproxy        // cumulative number of connections
	interceptedHaproxy      // cum. number of intercepted requests (monitor, stats)
	deniedConnectionHaproxy // requests denied by "tcp-request connection" rules
	deniedSessionHaproxy    // requests denied by "tcp-request session" rules
)

// Server server info for haproxy
type Server struct {
	Name               string `json:"name"`
	ServerName         string `json:"server_name"`
	CurrentQueue       int64  `json:"current_queue"`
	MaxQueue           int64  `json:"max_queue"`
	CurrentSession     int64  `json:"current_session"`
	MaxSession         int64  `json:"max_session"`
	SessionLimit       int64  `json:"session_limit"`
	SessionTotal       int64  `json:"session_total"`
	BytesIn            int64  `json:"bytes_in"`
	BytesOut           int64  `json:"bytes_out"`
	RequestDeny        int64  `json:"request_deny"`
	ResponseDeny       int64  `json:"response_deny"`
	RequestError       int64  `json:"request_error"`
	ConnectError       int64  `json:"connect_error"`
	ResponseError      int64  `json:"response_error"`
	ConnectRetry       int64  `json:"connect_retry"`
	ConnectRedispatch  int64  `json:"connect_redispatch"`
	Status             string `json:"status"`
	Weight             int64  `json:"weight"`
	Active             int64  `json:"active_server_num"`
	Backup             int64  `json:"backup_server_num"`
	CheckFail          int64  `json:"check_fail_num"`
	CheckDown          int64  `json:"check_down_num"`
	DownTime           int64  `json:"down_time_second"`
	DownTimeTotal      int64  `json:"down_time_total"`
	QueueMaxLimit      int64  `json:"queue_limit"`
	CurrentSessionRate int64  `json:"current_session_rate"`
	MaxSessionRate     int64  `json:"max_session_rate"`
	SessionRateLimit   int64  `json:"session_rate_limit"`
	CheckStatus        string `json:"check_status"`
	RequestRate        int64  `json:"request_rate"`
	RequestMaxRate     int64  `json:"request_max_rate"`
	RequestTotal       int64  `json:"request_total"`
	LastSessionSecond  int64  `json:"last_session_second"`
	LastCheckContent   string `json:"last_check_content"`
	Address            string `json:"address"`
	Mode               string `json:"mode"`
	ConnectRate        int64  `json:"connect_rate"`
	ConnectMaxRate     int64  `json:"connect_max_rate"`
}

// Service service info for haproxy
type Service struct {
	Frontend *Server   `json:"frontend"`
	Backend  *Server   `json:"backend"`
	Servers  []*Server `json:"servers"`
}

// Status status info for haproxy
type Status struct {
	HaproxyPid   int64      `json:"haproxy_pid"`
	UpTime       string     `json:"up_time"`
	UpTimeSecond int64      `json:"up_time_second"`
	ULimitN      int64      `json:"ulimit_n"`
	SystemLimit  string     `json:"system_limits"`
	MaxSock      int64      `json:"max_sock"`
	MaxConn      int64      `json:"max_conn"`
	MaxPipes     int64      `json:"max_pipes"`
	CurrentConn  int64      `json:"current_conn"`
	CurrentPipes int64      `json:"current_pipes"`
	ConnRate     int64      `json:"conn_rate"`
	ConnMaxRate  int64      `json:"conn_max_rate"`
	Services     []*Service `json:"services"`
}

// StatusResponse response for status api
type StatusResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    *Status `json:"data"`
}

func convertNumber(str string) int64 {
	if len(str) == 0 {
		return -1
	}
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return -2
	}
	return num
}

func convertServer(str string) (*Server, error) {
	fields := strings.Split(str, ",")
	if len(fields) != 84 {
		return nil, fmt.Errorf("fileds in one line must be 84, but get %d", len(fields))
	}
	return &Server{
		Name:               fields[proxyNameHaproxy],
		ServerName:         fields[serviceNameHaproxy],
		CurrentQueue:       convertNumber(fields[queueCurrentHaproxy]),
		MaxQueue:           convertNumber(fields[queueMaxHaproxy]),
		CurrentSession:     convertNumber(fields[sessionCurrentHaproxy]),
		MaxSession:         convertNumber(fields[sessionMaxHaproxy]),
		SessionLimit:       convertNumber(fields[sessionLimitHaproxy]),
		SessionTotal:       convertNumber(fields[sessionTotalHaproxy]),
		BytesIn:            convertNumber(fields[bytesInHaproxy]),
		BytesOut:           convertNumber(fields[bytesOutHaproxy]),
		RequestDeny:        convertNumber(fields[deniedRequestHaproxy]),
		ResponseDeny:       convertNumber(fields[deniedResponseHaproxy]),
		RequestError:       convertNumber(fields[errorRequestHaproxy]),
		ConnectError:       convertNumber(fields[errorConnectHaproxy]),
		ResponseError:      convertNumber(fields[errorResponseHaproxy]),
		ConnectRetry:       convertNumber(fields[wConnectRetryHaproxy]),
		ConnectRedispatch:  convertNumber(fields[wConnectRedispatchHaproxy]),
		Status:             fields[statusHaproxy],
		Weight:             convertNumber(fields[weightHaproxy]),
		Active:             convertNumber(fields[activeServerNumHaproxy]),
		Backup:             convertNumber(fields[backupServerNumHaproxy]),
		CheckFail:          convertNumber(fields[checkFailHaproxy]),
		CheckDown:          convertNumber(fields[checkDownHaproxy]),
		DownTime:           convertNumber(fields[secondSinceLastCheckDownHaproxy]),
		DownTimeTotal:      convertNumber(fields[downTimeTotalHaproxy]),
		QueueMaxLimit:      convertNumber(fields[queueMaxLimitHaproxy]),
		CurrentSessionRate: convertNumber(fields[sessionRateHaproxy]),
		MaxSessionRate:     convertNumber(fields[sessionRateMaxHaproxy]),
		SessionRateLimit:   convertNumber(fields[sessionRateLimHaproxy]),
		CheckStatus:        fields[checkStatusHaproxy],
		RequestRate:        convertNumber(fields[requestRateHaproxy]),
		RequestMaxRate:     convertNumber(fields[requestRateMaxHaproxy]),
		RequestTotal:       convertNumber(fields[requestTotalHaproxy]),
		LastSessionSecond:  convertNumber(fields[lastSesseionSecondHaproxy]),
		LastCheckContent:   fields[lastCheckContentHaproxy],
		Address:            fields[addrHaproxy],
		Mode:               fields[modeHaproxy],
		ConnectRate:        convertNumber(fields[connRateHaproxy]),
		ConnectMaxRate:     convertNumber(fields[connRateMaxHaproxy]),
	}, nil
}

func (m *Manager) convertStat(str string) ([]*Service, error) {
	serviceMap := make(map[string]*Service)
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		// skip annotation in response
		if strings.HasPrefix(line, "#") {
			continue
		}
		tmpServer, err := convertServer(line)
		if err != nil {
			blog.Infof("convert line %s to server failed, err %s", line, err.Error())
			return nil, fmt.Errorf("convert line %s to server failed, err %s", line, err.Error())
		}
		if _, ok := serviceMap[tmpServer.Name]; !ok {
			serviceMap[tmpServer.Name] = &Service{}
		}
		switch tmpServer.ServerName {
		case "FRONTEND":
			serviceMap[tmpServer.Name].Frontend = tmpServer
		case "BACKEND":
			serviceMap[tmpServer.Name].Backend = tmpServer
		default:
			serviceMap[tmpServer.Name].Servers = append(serviceMap[tmpServer.Name].Servers, tmpServer)
		}
	}
	retServices := make([]*Service, 0)
	for _, svc := range serviceMap {
		retServices = append(retServices, svc)
	}
	return retServices, nil
}

func (m *Manager) convertInfo(str string) (*Status, error) {
	status := &Status{}
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		fields := strings.Split(line, ": ")
		if len(fields) != 2 {
			blog.Warnf("info line %s invalid", line)
			continue
		}
		switch fields[0] {
		case "Pid":
			status.HaproxyPid = convertNumber(fields[1])
		case "Uptime":
			status.UpTime = fields[1]
		case "Uptime_sec":
			status.UpTimeSecond = convertNumber(fields[1])
		case "Ulimit-n":
			status.ULimitN = convertNumber(fields[1])
		case "Maxsock":
			status.MaxSock = convertNumber(fields[1])
		case "Maxconn":
			status.MaxConn = convertNumber(fields[1])
		case "Maxpipes":
			status.MaxPipes = convertNumber(fields[1])
		case "CurrConns":
			status.CurrentConn = convertNumber(fields[1])
		case "ConnRate":
			status.ConnRate = convertNumber(fields[1])
		case "MaxConnRate":
			status.ConnMaxRate = convertNumber(fields[1])
		}
	}
	return status, nil
}

func sendCommandToHaproxy(sockAddr, command string) (string, error) {
	unixAddr, err := net.ResolveUnixAddr("unix", sockAddr)
	if err != nil {
		return "", fmt.Errorf("resolve unix socket addr %s failed, err %s", sockAddr, err.Error())
	}
	conn, err := net.DialUnix("unix", nil, unixAddr)
	if err != nil {
		return "", fmt.Errorf("dial unix socket failed, err %s", err.Error())
	}
	defer conn.Close()
	_, err = conn.Write([]byte(command))
	if err != nil {
		return "", fmt.Errorf("send command %s failed, err %s", command, err.Error())
	}

	bytesBuffer := new(bytes.Buffer)
	buffer := make([]byte, 1024)
	for {
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			blog.Errorf("read err from tcp connection, err %s", err.Error())
			return "", fmt.Errorf("read err from tcp connection, err %s", err.Error())
		}
		bytesWriteToBuffer, err := bytesBuffer.Write(buffer[:bytesRead])
		if err != nil {
			blog.Errorf("write to buffer failed, err %s", err.Error())
			return "", fmt.Errorf("write to buffer failed, err %s", err.Error())
		}
		if bytesWriteToBuffer != bytesRead {
			blog.Errorf("write bytes %d to buffer, but %d bytes succeed", bytesWriteToBuffer, bytesRead)
			return "", fmt.Errorf("write bytes %d to buffer, but %d bytes succeed", bytesWriteToBuffer, bytesRead)
		}
	}
	return bytesBuffer.String(), nil
}

func (m *Manager) fetch() error {
	showInfoCommand := "show info\nquit\n"
	showStatCommand := "show stat\nquit\n"
	m.sockMutex.Lock()
	infoStr, err := m.haproxyClient.ExecuteRaw(showInfoCommand)
	m.sockMutex.Unlock()
	if err != nil {
		blog.Errorf("send command %s to haproxy failed, err %s", showInfoCommand, err.Error())
		return err
	}
	status, err := m.convertInfo(infoStr)
	if err != nil {
		blog.Errorf("convert str %s to haproxy info failed, err %s", infoStr, err.Error())
		return err
	}
	m.sockMutex.Lock()
	str, err := m.haproxyClient.ExecuteRaw(showStatCommand)
	m.sockMutex.Unlock()
	if err != nil {
		blog.Errorf("send command %s to haproxy failed, err %s", showStatCommand, err.Error())
		return err
	}
	svcs, err := m.convertStat(str)
	if err != nil {
		blog.Errorf("convert str %s to haproxy service array failed, err %s", str, err.Error())
		return err
	}
	status.Services = svcs
	m.statsMutex.Lock()
	m.stats = status
	m.statsMutex.Unlock()
	return nil
}

func (m *Manager) status(req *restful.Request, resp *restful.Response) {
	var status *Status
	m.statsMutex.Lock()
	status = m.stats
	m.statsMutex.Unlock()
	resp.WriteEntity(&StatusResponse{Code: 0, Message: "success", Data: status})
}

func (m *Manager) runStatusFetch() {
	tick := time.NewTicker(time.Second * time.Duration(m.statusFetchPeriod))
	for {
		select {
		case <-tick.C:
			err := m.fetch()
			if err != nil {
				LoadbalanceHaproxyStatsFetchStateMetric.Set(0)
			} else {
				LoadbalanceHaproxyStatsFetchStateMetric.Set(1)
			}
		case <-m.stopCh:
			return
		}
	}
}

// GetStatusFunction get status function
func (m *Manager) GetStatusFunction() restful.RouteFunction {
	return m.status
}
