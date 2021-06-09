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
	"fmt"
	"html/template"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/metric"
	conf "github.com/Tencent/bk-bcs/bcs-services/bcs-loadbalance/template"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-loadbalance/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-loadbalance/util"
)

// NewManager create haproxy config file manager
func NewManager(lbName, binPath, cfgPath, generatePath, backupPath, templatePath string, statusFetchPeriod int) (conf.Manager, error) {
	envConfig := loadEnvConfig()
	haproxyClient, err := NewRuntimeClient(envConfig.SockPath)
	if err != nil {
		blog.Infof("create haproxy runtime client with sockpath %s failed, err %s", envConfig.SockPath, err.Error())
		return nil, fmt.Errorf("create haproxy runtime client with sockpath %s failed, err %s", envConfig.SockPath, err.Error())
	}
	manager := &Manager{
		LoadbalanceName:   lbName,
		haproxyBin:        binPath,
		cfgFile:           cfgPath,
		tmpDir:            generatePath,
		backupDir:         backupPath,
		templateFile:      filepath.Join(templatePath, "haproxy.cfg.template"),
		statusFetchPeriod: statusFetchPeriod,
		stopCh:            make(chan struct{}),
		healthInfo: metric.HealthMeta{
			IsHealthy:   conf.HealthStatusOK,
			Message:     conf.HealthStatusOKMsg,
			CurrentRole: metric.SlaveRole,
		},
		envConfig:     envConfig,
		haproxyClient: haproxyClient,
	}
	manager.initMetric()
	return manager, nil
}

// Manager implements TemplateManager interface, control
// haproxy config file generating, validation, backup and reloading
type Manager struct {
	Cluster           string // cluster id
	LoadbalanceName   string // loadbalance instance id
	haproxyBin        string //absolute path for haproxy executable binary
	cfgFile           string //absolute path for haproxy cfg file
	backupDir         string //absolute path for cfg file backup storage
	tmpDir            string //temporary file for create new file
	templateFile      string //template file
	envConfig         *EnvConfig
	haproxyClient     *RuntimeClient
	currentConfig     *Config
	ConfigCache       *Config
	statusFetchPeriod int           // period for fetch haproxy stats data
	stopCh            chan struct{} // chan for stop fetching haproxy stats data
	stats             *Status       //stats for haproxy
	statsMutex        sync.Mutex    // lock for stats haproxy
	sockMutex         sync.Mutex
	healthInfo        metric.HealthMeta //Health information
	healthLock        sync.RWMutex
}

// Start point, do not block
func (m *Manager) Start() error {
	// check template exist
	if !conf.IsFileExist(m.haproxyBin) {
		blog.Error("haproxy executable file lost")
		return fmt.Errorf("haproxy executable file lost")
	}
	if !conf.IsFileExist(m.templateFile) {
		blog.Error("haproxy.cfg.template do not exist")
		return fmt.Errorf("haproxy.cfg.template do not exist")
	}
	// create other file directory
	err := os.MkdirAll(m.backupDir, os.ModePerm)
	if err != nil {
		blog.Warnf("mkdir %s failed, err %s", m.backupDir, err.Error())
	}
	err = os.MkdirAll(m.tmpDir, os.ModePerm)
	if err != nil {
		blog.Warnf("mkdir %s failed, err %s", m.tmpDir, err.Error())
	}
	// run haproxy status fetcher
	go m.runStatusFetch()
	return nil
}

// Stop stop
func (m *Manager) Stop() {
	close(m.stopCh)
}

// Create config file with tmpData,
func (m *Manager) Create(tmpData *types.TemplateData) (string, error) {
	var err error
	var t *template.Template
	var writer *os.File
	// loading template file
	t, err = template.ParseFiles(m.templateFile)
	if err != nil {
		blog.Errorf("Parse template file %s failed: %s", m.templateFile, err.Error())
		return "", err
	}
	// create new config file
	fileName := "haproxy." + strconv.Itoa(rand.Int()) + ".cfg"
	absName := filepath.Join(m.tmpDir, fileName)
	writer, err = os.Create(absName)
	if err != nil {
		blog.Errorf("Create tempory new config file %s failed: %s", absName, err.Error())
		return "", err
	}
	m.currentConfig = m.convertData(tmpData)
	m.currentConfig.generateServerName()
	m.currentConfig.generateRenderData()
	err = t.Execute(writer, m.currentConfig)
	if err != nil {
		blog.Errorf("Template Execute Err: %s", err.Error())
		return "", err
	}
	blog.Infof("Create new haproxy.cfg %s success", absName)
	return absName, nil
}

func getServerName(backendName string, index int) string {
	return backendName + "_" + strconv.Itoa(index)
}

// TryUpdateWithoutReload update haproxy config without reloading
// needReload: true for reload
func (m *Manager) TryUpdateWithoutReload(tmpData *types.TemplateData) (needReload bool) {
	m.currentConfig = m.convertData(tmpData)
	if m.ConfigCache == nil {
		return true
	}
	haproxyNeedReload, needUpdate, haproxyCommands := m.checkConfigDifference(m.currentConfig)
	if haproxyNeedReload {
		return true
	}
	if needUpdate {
		blog.Infof("there is %d commands to execute", len(haproxyCommands))
		for _, command := range haproxyCommands {
			m.sockMutex.Lock()
			commandStr, err := m.haproxyClient.ExecuteRaw(command)
			m.sockMutex.Unlock()
			if err != nil {
				blog.Infof("execute haproxy command %s failed, err %s, need reload", command, err.Error())
				return true
			}
			blog.Infof("executed haproxy command: %s, ret: %s", command, commandStr)
		}
	} else {
		blog.Infof("no updates")
	}

	return false
}

// CheckDifference two file are difference, true is difference
func (m *Manager) CheckDifference(oldFile, curFile string) bool {
	if !conf.IsFileExist(oldFile) {
		blog.Errorf("Old haproxy.cfg %s Do not exist", oldFile)
		return false
	}
	if !conf.IsFileExist(curFile) {
		blog.Errorf("Current haproxy.cfg %s Do not exist", oldFile)
		return false
	}
	// calculate oldFile md5
	oldMd5, err := util.Md5SumForFile(oldFile)
	if err != nil {
		blog.Errorf("calculate old haproxy file %s md5sum failed, err %s", oldFile, err.Error())
		return false
	}
	// calculate curFile md5
	newMd5, err := util.Md5SumForFile(curFile)
	if err != nil {
		blog.Errorf("calculate cur haproxy file %s md5sum failed, err %s", curFile, err.Error())
		return false
	}
	// compare
	if oldMd5 != newMd5 {
		blog.Info("New and old haproxy.cfg MD5 is difference")
		return true
	}
	m.ConfigCache = m.currentConfig
	return false
}

// Validate new cfg file grammar is OK
func (m *Manager) Validate(newFile string) bool {
	command := m.haproxyBin + " -c -f " + newFile
	m.sockMutex.Lock()
	output, ok := util.ExeCommand(command)
	m.sockMutex.Unlock()
	if !ok {
		blog.Errorf("Validate with command [%s] failed", command)
		return false
	}
	blog.Infof("Validate with command %s, output: %s", command, output)
	return true
}

// Replace old cfg file with cur one, return old file backup
func (m *Manager) Replace(oldFile, curFile string) error {
	return util.ReplaceFile(oldFile, curFile)
}

// Reload haproxy with new config file
func (m *Manager) Reload(cfgFile string) error {
	command := m.haproxyBin + " -f " + cfgFile + " -sf $(cat /var/run/haproxy.pid)"
	m.sockMutex.Lock()
	output, ok := util.ExeCommand(command)
	m.sockMutex.Unlock()
	if !ok {
		blog.Errorf("Reload with command [%s] failed: %s", command, output)
		return fmt.Errorf("Reload config err")
	}
	blog.Infof("Reload with command %s, output: %s", command, output)
	m.ConfigCache = m.currentConfig
	return nil
}

// convertHTTPData
func (m *Manager) convertHTTPData(HTTP types.HTTPServiceInfoList, protocol string) map[int]*HTTPFrontend {
	httpMap := make(map[int]*HTTPFrontend)
	for _, http := range HTTP {
		tmpHTTPFrontend, ok := httpMap[http.ServicePort]
		if !ok {
			tmpHTTPFrontend = &HTTPFrontend{
				ServicePort: http.ServicePort,
				Name:        http.Name,
				Backends:    make(map[string]*HTTPBackend),
			}
			httpMap[http.ServicePort] = tmpHTTPFrontend
		}
		for _, back := range http.Backends {
			tmpHTTPBackend := &HTTPBackend{
				Name:                protocol + "_" + back.UpstreamName,
				Domain:              http.BCSVHost,
				URL:                 back.Path,
				Balance:             http.Balance,
				HealthCheckInterval: m.envConfig.ServerHealthCheckInterval,
				RiseHealthCheckNum:  m.envConfig.ServerRiseHealthCheckNum,
				FallHealthCheckNum:  m.envConfig.ServerFallHealthCheckNum,
				Servers:             make(map[string]*RealServer),
			}
			for _, server := range back.BackendList {
				tmpServer := &RealServer{
					Name:   "",
					IP:     server.IP,
					Port:   server.Port,
					Weight: server.Weight,
				}
				tmpHTTPBackend.Servers[tmpServer.Key()] = tmpServer
			}
			tmpHTTPFrontend.Backends[http.BCSVHost+"_"+back.UpstreamName] = tmpHTTPBackend
			break
		}
		httpMap[http.ServicePort] = tmpHTTPFrontend
	}
	return httpMap
}

func (m *Manager) convertTCPData(TCP types.FourLayerServiceInfoList) map[int]*TCPListener {
	tcpMap := make(map[int]*TCPListener)
	for _, tcp := range TCP {
		tmpTCPListener, ok := tcpMap[tcp.ServicePort]
		if !ok {
			tmpTCPListener = &TCPListener{
				ServicePort:         tcp.ServicePort,
				Name:                "tcp_" + tcp.Name,
				Balance:             tcp.Balance,
				HealthCheckInterval: m.envConfig.ServerHealthCheckInterval,
				RiseHealthCheckNum:  m.envConfig.ServerRiseHealthCheckNum,
				FallHealthCheckNum:  m.envConfig.ServerFallHealthCheckNum,
				Servers:             make(map[string]*RealServer),
			}
		}
		for _, server := range tcp.Backends {
			tmpServer := &RealServer{
				Name:   "",
				IP:     server.IP,
				Port:   server.Port,
				Weight: server.Weight,
			}
			tmpTCPListener.Servers[tmpServer.Key()] = tmpServer
		}
		tcpMap[tcp.ServicePort] = tmpTCPListener
	}
	return tcpMap
}

// convertData convert template data to haproxy config data
// TODO: to deal with port conflict
func (m *Manager) convertData(tmpData *types.TemplateData) *Config {
	newConfig := &Config{
		LogEnabled: m.envConfig.LogEnabled,
		LogLevel:   m.envConfig.LogLevel,
		SockPath:   m.envConfig.SockPath,
		ThreadNum:  m.envConfig.ThreadNum,
		MaxConn:    m.envConfig.MaxConn,
		PidPath:    m.envConfig.PidPath,
		SSLCert:    m.envConfig.SSLCert,
		DefaultProxyConfig: &ProxyDefault{
			Retries:              m.envConfig.Retries,
			Backlog:              m.envConfig.Backlog,
			MaxConn:              m.envConfig.ProxyMaxConn,
			TimeoutConnection:    m.envConfig.ProxyTimeoutConnection,
			TimeoutClient:        m.envConfig.ProxyTimeoutClient,
			TimeoutServer:        m.envConfig.ProxyTimeoutServer,
			TimeoutTunnel:        m.envConfig.ProxyTimeoutTunnel,
			TimeoutHTTPKeepAlive: m.envConfig.ProxyTimeoutHTTPKeepAlive,
			TimeoutHTTPRequest:   m.envConfig.ProxyTimeoutHTTPRequest,
			TimeoutQueue:         m.envConfig.ProxyTimeoutQueue,
			TimeoutTarpit:        m.envConfig.ProxyTimeoutTarpit,
			Options:              m.envConfig.ProxyOptions,
			HTTPOptions:          m.envConfig.HTTPProxyOptions,
		},
		Stats: &StatsFrontend{
			Port:         m.envConfig.StatsFrontendPort,
			URI:          m.envConfig.StatsFrontendURI,
			AuthUser:     m.envConfig.StatsFrontendAuthUser,
			AuthPassword: m.envConfig.StatsFrontendAuthPassword,
		},
		HTTPMap:  m.convertHTTPData(tmpData.HTTP, "http"),
		HTTPSMap: m.convertHTTPData(tmpData.HTTPS, "https"),
		TCPMap:   m.convertTCPData(tmpData.TCP),
	}
	return newConfig
}

// checkConfigDifference check the difference between the incoming template data and cache data
// needReload: true means should reload haproxy to do this change
// needUpdate: true means should call haproxy runtime command
// haproxyCommand: haproxy runtime commands that should be executed
func (m *Manager) checkConfigDifference(newConfig *Config) (needReload bool, needUpdate bool, haproxyCommands []string) {
	configNeedReload := false
	configNeedUpdate := false
	var configCommands []string
	if m.ConfigCache == nil {
		blog.Infof("haproxy manager config cache is nil")
		return true, false, nil
	}
	// check global config
	if newConfig.SockPath != m.ConfigCache.SockPath ||
		newConfig.ThreadNum != m.ConfigCache.ThreadNum ||
		newConfig.MaxConn != m.ConfigCache.MaxConn ||
		newConfig.PidPath != m.ConfigCache.PidPath ||
		newConfig.SSLCert != m.ConfigCache.SSLCert {
		blog.Infof("haproxy global config is different, new %v, old %v", newConfig, m.ConfigCache)
		return true, false, nil
	}
	// check default proxy config
	if newConfig.DefaultProxyConfig != nil && m.ConfigCache.DefaultProxyConfig != nil {
		if newConfig.DefaultProxyConfig.Retries != m.ConfigCache.DefaultProxyConfig.Retries ||
			newConfig.DefaultProxyConfig.Backlog != m.ConfigCache.DefaultProxyConfig.Backlog ||
			newConfig.DefaultProxyConfig.MaxConn != m.ConfigCache.DefaultProxyConfig.MaxConn ||
			newConfig.DefaultProxyConfig.TimeoutConnection != m.ConfigCache.DefaultProxyConfig.TimeoutConnection ||
			newConfig.DefaultProxyConfig.TimeoutClient != m.ConfigCache.DefaultProxyConfig.TimeoutClient ||
			newConfig.DefaultProxyConfig.TimeoutServer != m.ConfigCache.DefaultProxyConfig.TimeoutServer ||
			newConfig.DefaultProxyConfig.TimeoutTunnel != m.ConfigCache.DefaultProxyConfig.TimeoutTunnel ||
			newConfig.DefaultProxyConfig.TimeoutHTTPKeepAlive != m.ConfigCache.DefaultProxyConfig.TimeoutHTTPKeepAlive ||
			newConfig.DefaultProxyConfig.TimeoutHTTPRequest != m.ConfigCache.DefaultProxyConfig.TimeoutHTTPRequest ||
			newConfig.DefaultProxyConfig.TimeoutQueue != m.ConfigCache.DefaultProxyConfig.TimeoutQueue ||
			newConfig.DefaultProxyConfig.TimeoutTarpit != m.ConfigCache.DefaultProxyConfig.TimeoutTarpit {
			if len(newConfig.DefaultProxyConfig.Options) != len(m.ConfigCache.DefaultProxyConfig.Options) {
				blog.Infof("different length, new default options length %d, old default options length %d",
					len(newConfig.DefaultProxyConfig.Options), len(m.ConfigCache.DefaultProxyConfig.Options))
				return true, false, nil
			}
			if len(newConfig.DefaultProxyConfig.Options) != 0 {
				for index, value := range newConfig.DefaultProxyConfig.Options {
					if value != m.ConfigCache.DefaultProxyConfig.Options[index] {
						blog.Infof("different options, new options %v, old options %v",
							newConfig.DefaultProxyConfig.Options,
							m.ConfigCache.DefaultProxyConfig.Options)
						return true, false, nil
					}
				}
			}
		}
	}
	// http frontend
	if len(newConfig.HTTPMap) != len(m.ConfigCache.HTTPMap) {
		blog.Infof("new HTTP frontend list is different from cached HTTP frontend list, new length %d, old length %d", len(newConfig.HTTPMap), len(m.ConfigCache.HTTPMap))
		return true, false, nil
	}
	if len(newConfig.HTTPMap) != 0 {
		for port, httpFrontend := range newConfig.HTTPMap {
			frontendNeedReload, frontendNeedUpdate, frontendCommands := checkConfigDiffBetweenHTTPFrontend(httpFrontend, m.ConfigCache.HTTPMap[port])
			if frontendNeedReload {
				return true, false, nil
			}
			if frontendNeedUpdate {
				configNeedUpdate = true
				configCommands = append(configCommands, frontendCommands...)
			}
		}
	}
	// https frontend
	if len(newConfig.HTTPSMap) != len(m.ConfigCache.HTTPSMap) {
		blog.Infof("new HTTPS frontend list is different from cached HTTPS frontend list, new length %d, old length %d", len(newConfig.HTTPSMap), len(m.ConfigCache.HTTPSMap))
		return true, false, nil
	}
	if len(newConfig.HTTPSMap) != 0 {
		for port, httpFrontend := range newConfig.HTTPSMap {
			frontendNeedReload, frontendNeedUpdate, frontendCommands := checkConfigDiffBetweenHTTPFrontend(httpFrontend, m.ConfigCache.HTTPSMap[port])
			if frontendNeedReload {
				return true, false, nil
			}
			if frontendNeedUpdate {
				configNeedUpdate = true
				configCommands = append(configCommands, frontendCommands...)
			}
		}
	}
	// tcp listener
	if len(newConfig.TCPMap) != len(m.ConfigCache.TCPMap) {
		blog.Infof("new TCP frontend list is different from cache TCP frontend list , new length %d, old length %d", len(newConfig.TCPMap), len(m.ConfigCache.TCPMap))
		return true, false, nil
	}
	if len(newConfig.TCPMap) != 0 {
		for port, tcpListener := range newConfig.TCPMap {
			listenerNeedReload, listenerNeedUpdate, listenerCommands := checkConfigDiffBetweenTCPListener(tcpListener, m.ConfigCache.TCPMap[port])
			if listenerNeedReload {
				return true, false, nil
			}
			if listenerNeedUpdate {
				configNeedUpdate = true
				configCommands = append(configCommands, listenerCommands...)
			}
		}
	}

	return configNeedReload, configNeedUpdate, configCommands
}

func checkConfigDiffBetweenTCPListener(newListener *TCPListener, oldListener *TCPListener) (needReload bool, needUpdate bool, haproxyCommands []string) {
	if newListener == nil && oldListener == nil {
		return false, false, nil
	}
	if (newListener == nil && oldListener != nil) || (newListener != nil && oldListener == nil) {
		blog.Infof("find empty tcp listener")
		return true, false, nil
	}
	if newListener.ServicePort != oldListener.ServicePort {
		blog.Infof("%v has different port from %v", newListener, oldListener)
		return true, false, nil
	}
	if newListener.Name != oldListener.Name {
		blog.Infof("%v has different name from %v", newListener, oldListener)
		return true, false, nil
	}
	if len(newListener.Servers) > len(oldListener.Servers) {
		blog.Infof("new listener %s has %d servers, the old only has %d servers", newListener.Name, len(newListener.Servers), len(oldListener.Servers))
		return true, false, nil
	}

	return checkConfigDiffBetweenRealServer(newListener.Name, newListener.Servers, oldListener.Servers)

}

func checkConfigDiffBetweenHTTPFrontend(newFront *HTTPFrontend, oldFront *HTTPFrontend) (needReload bool, needUpdate bool, haproxyCommands []string) {
	frontendNeedReload := false
	frontendNeedUpdate := false
	var frontendCommands []string
	if newFront == nil && oldFront == nil {
		return false, false, nil
	}
	if (newFront == nil && oldFront != nil) || (newFront != nil && oldFront == nil) {
		blog.Infof("find empty http frontend")
		return true, false, nil
	}
	if newFront.ServicePort != oldFront.ServicePort {
		blog.Infof("%v has different port from %v", newFront, oldFront)
		return true, false, nil
	}
	if newFront.Name != oldFront.Name {
		blog.Infof("%v has different name from %v", newFront, oldFront)
		return true, false, nil
	}
	if len(newFront.Backends) != len(oldFront.Backends) {
		blog.Infof("different backends length for http front, new frontend %v, old frontend %v", newFront, oldFront)
		return true, false, nil
	}

	for backendKey, newFrontBackend := range newFront.Backends {
		oldFrontBackend, ok := oldFront.Backends[backendKey]
		if !ok {
			blog.Infof("backend %v is newly added", newFrontBackend)
			return true, false, nil
		}
		backendNeedReload, backendNeedUpdate, backendCommands := checkConfigDiffBetweenHTTPBackend(newFrontBackend, oldFrontBackend)
		if backendNeedReload {
			return true, false, nil
		}
		if backendNeedUpdate {
			frontendNeedUpdate = true
			frontendCommands = append(frontendCommands, backendCommands...)
		}
	}
	return frontendNeedReload, frontendNeedUpdate, frontendCommands
}

func checkConfigDiffBetweenHTTPBackend(newBack *HTTPBackend, oldBack *HTTPBackend) (needReload bool, needUpdate bool, haproxyCommands []string) {
	if newBack == nil && oldBack == nil {
		return false, false, nil
	}
	if (newBack == nil && oldBack != nil) || (newBack != nil && oldBack == nil) {
		blog.Infof("find empty http backend")
		return true, false, nil
	}
	if newBack.Name != oldBack.Name || newBack.Domain != oldBack.Domain || newBack.URL != oldBack.URL {
		blog.Infof("find differences in http backend, new backend %v, old backend %v", newBack, oldBack)
		return true, false, nil
	}
	if len(newBack.Servers) > len(oldBack.Servers) {
		blog.Infof("newBackend %s has %d servers, the old only has %d servers", newBack.Name, len(newBack.Servers), len(oldBack.Servers))
		return true, false, nil
	}
	return checkConfigDiffBetweenRealServer(newBack.Name, newBack.Servers, oldBack.Servers)
}

func checkConfigDiffBetweenRealServer(backendName string, newServers map[string]*RealServer, oldServers map[string]*RealServer) (needReload bool, needUpdate bool, haproxyCommands []string) {
	backendNeedReload := false
	backendNeedUpdate := false
	var backendCommands []string
	var newServerList []*RealServer
	for key, newRealServer := range newServers {
		oldRealServer, ok := oldServers[key]
		if !ok {
			newServerList = append(newServerList, newRealServer)
			continue
		} else {
			// enable server
			if oldRealServer.Disabled {
				backendNeedUpdate = true
				backendCommands = append(backendCommands,
					newEnableServerCommand(backendName, oldRealServer.Name))
				oldRealServer.Disabled = false
			}
			// update server weight
			if newRealServer.Weight != oldRealServer.Weight {
				backendNeedUpdate = true
				backendCommands = append(backendCommands,
					newSetServerWeightCommand(backendName, oldRealServer.Name, newRealServer.Weight))
				oldRealServer.Weight = newRealServer.Weight
			}
		}
	}
	var oldServerList []*RealServer
	for key, oldRealServer := range oldServers {
		_, ok := newServers[key]
		if !ok {
			oldServerList = append(oldServerList, oldRealServer)
			continue
		}
	}
	if len(newServerList) > 0 || len(oldServerList) > 0 {
		backendNeedUpdate = true
		i := 0
		occupiedServerMap := make(map[string]*RealServer)
		// use new server IP port to occupy the position of the server to be deleted
		for ; i < len(newServerList); i++ {
			newServer := newServerList[i]
			oldServer := oldServerList[i]
			if oldServer.Disabled {
				backendCommands = append(backendCommands,
					newEnableServerCommand(backendName, oldServer.Name))
				oldServer.Disabled = false
			}
			backendCommands = append(backendCommands, newSetServerAddrCommand(backendName, oldServer.Name, newServer.IP, newServer.Port))
			newServer.Name = oldServer.Name
			occupiedServerMap[oldServer.Key()] = oldServer
			delete(oldServers, oldServer.Key())
			oldServers[newServer.Key()] = newServer
		}
		// in case that the deleted servers is more than the news, we just set weight to 0 to avoid haproxy reload
		for _, oldServer := range oldServerList {
			_, ok := occupiedServerMap[oldServer.Key()]
			if !ok && !oldServer.Disabled {
				backendCommands = append(backendCommands, newDisableServerCommand(backendName, oldServer.Name))
				oldServer.Disabled = true
			}
		}
	}
	return backendNeedReload, backendNeedUpdate, backendCommands
}

func newSetServerWeightCommand(backend, server string, weight int) string {
	return fmt.Sprintf("set server %s/%s weight %d", backend, server, weight)
}

func newSetServerAddrCommand(backend, server, addr string, port int) string {
	return fmt.Sprintf("set server %s/%s addr %s port %d", backend, server, addr, port)
}

func newEnableServerCommand(backend, server string) string {
	return fmt.Sprintf("enable server %s/%s", backend, server)
}

func newDisableServerCommand(backend, server string) string {
	return fmt.Sprintf("disable server %s/%s", backend, server)
}
