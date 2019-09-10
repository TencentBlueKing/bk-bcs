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

package app

import (
	"bk-bcs/bcs-common/common/bcs-health/api"
	"bk-bcs/bcs-common/common/blog"
	loadbalance "bk-bcs/bcs-common/pkg/loadbalance/v2"
	"bk-bcs/bcs-services/bcs-loadbalance/clear"
	"bk-bcs/bcs-services/bcs-loadbalance/monitor"
	bcsprometheus "bk-bcs/bcs-services/bcs-loadbalance/monitor/prometheus"
	"bk-bcs/bcs-services/bcs-loadbalance/monitor/status"
	"bk-bcs/bcs-services/bcs-loadbalance/option"
	"bk-bcs/bcs-services/bcs-loadbalance/rdiscover"
	"bk-bcs/bcs-services/bcs-loadbalance/template"
	"bk-bcs/bcs-services/bcs-loadbalance/template/haproxy"
	"bk-bcs/bcs-services/bcs-loadbalance/template/nginx"
	"bk-bcs/bcs-services/bcs-loadbalance/types"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// EventHandler is event interface when
type EventHandler interface {
	OnAdd(obj interface{})
	OnDelete(obj interface{})
	OnUpdate(oldObj, newObj interface{})
}

//InitLogger init app logger
func InitLogger(config *option.LBConfig) {
	blog.InitLogs(config.LogConfig)
}

// CloseLogger close logger
func CloseLogger() {
	blog.CloseLogs()
}

//NewEventProcessor create EventProcessor with LBConfig
func NewEventProcessor(config *option.LBConfig) *LBEventProcessor {
	processor := &LBEventProcessor{
		update:       false,
		generate:     false,
		reload:       false,
		signals:      make(chan os.Signal, 10),
		exit:         make(chan struct{}),
		config:       config,
		clearManager: clear.NewClearManager(),
	}

	// register both service zookeeper and cluster zookeeper
	// service zookeeper for health check, service register
	// cluster zookeeper for prometheus metrics collector
	zkSubRegPath := config.ClusterID + "/" + config.Group
	processor.rd = rdiscover.NewRDiscover(config.BcsZkAddr, zkSubRegPath, config.ClusterID, config.Proxy, config.Address, uint(config.MetricPort))
	if len(config.ClusterZk) != 0 {
		processor.clusterRd = rdiscover.NewRDiscover(config.ClusterZk, zkSubRegPath, config.ClusterID, config.Proxy, config.Address, uint(config.MetricPort))
	}

	processor.reflector = NewReflector(config, processor)
	// new Alarming interface
	blog.Infof("new bcs health with ca %s, cert %s, key %s", config.CAFile, config.ClientCertFile, config.ClientKeyFile)
	tls := api.TLSConfig{
		CaFile:   config.CAFile,
		CertFile: config.ClientCertFile,
		KeyFile:  config.ClientKeyFile,
	}
	if err := api.NewBcsHealth(config.BcsZkAddr, tls); nil != err {
		blog.Errorf("new bcs health instance failed. err: %s", err.Error())
	}

	lbMonitor := monitor.NewMonitor(config.Address, int(config.MetricPort))
	newMetricResource := bcsprometheus.NewPromMetric()
	if config.Proxy == option.ProxyHaproxy {
		blog.Infof("use haproxy transmit")
		processor.cfgManager = haproxy.NewManager(
			config.BinPath,
			config.CfgPath,
			config.GeneratingDir,
			config.CfgBackupDir,
			config.TemplateDir,
			config.StatusFetchPeriod,
		)
	} else {
		blog.Infof("use nginx transmit")
		processor.cfgManager = nginx.NewManager(
			config.BinPath,
			config.CfgPath,
			config.GeneratingDir,
			config.CfgBackupDir,
			config.TemplateDir)
	}
	prometheus.MustRegister(processor.cfgManager)
	newStatusResource := status.NewStatus(processor.cfgManager.GetStatusFunction())
	lbMonitor.RegisterResource(newMetricResource)
	lbMonitor.RegisterResource(newStatusResource)
	processor.monitor = lbMonitor
	return processor
}

//LBEventProcessor event loop for handling data change event.
type LBEventProcessor struct {
	update       bool                 //update flag
	generate     bool                 //flag for resetting HAProxy configuration
	reload       bool                 //flag for reloading HAProxy
	signals      chan os.Signal       //handle all signal we need, reserved
	exit         chan struct{}        //flag for processor exit
	config       *option.LBConfig     //config item from config file
	reflector    DataReflector        //data cache holder
	cfgManager   template.Manager     //template manager
	rd           *rdiscover.RDiscover //bcs zookeeper register
	clusterRd    *rdiscover.RDiscover //cluster zookeeper register
	clearManager *clear.Manager       //timer to clear template file
	monitor      *monitor.Monitor     // monitor to support metric and status api
}

//Start starting point for event processing
//1. start reflector to cache data from storage
//2. start template manager for Create/Reload config for haproxy.cfg
//3. start local logic loop for check data changed
func (lp *LBEventProcessor) Start() error {

	go func() {
		if err := lp.monitor.Run(); err != nil {
			blog.Errorf("run lb monitor failed, err %s", err.Error())
		}
	}()
	blog.Infof("run lb monitor")

	go func() {
		if err := lp.rd.Start(); err != nil {
			blog.Errorf("start register zookeeper error: %s", err.Error())
			//should go ahead to work event if register zookeeper failed
		}
	}()
	blog.Infof("start register success")

	if len(lp.config.ClusterZk) != 0 {
		go func() {
			if err := lp.clusterRd.Start(); err != nil {
				blog.Errorf("start register cluster zookeeper error: %s", err.Error())
			}
		}()
		blog.Infof("start cluster register success")
	}

	if err := lp.reflector.Start(); err != nil {
		blog.Errorf("start Reflector error: %s", err.Error())
		return err
	}
	blog.Infof("start reflector success")

	if err := lp.cfgManager.Start(); err != nil {
		blog.Errorf("start ConfigManager error: %s", err.Error())
		return err
	}
	blog.Infof("start config manager successfully")

	lp.clearManager.Start()
	go lp.run()
	return nil
}

//run main loop
func (lp *LBEventProcessor) run() {
	updateTick := time.NewTicker(time.Second * time.Duration(int64(lp.config.CfgCheckPeriod)))
	syncTick := time.NewTicker(time.Second * time.Duration(int64(lp.config.SyncPeriod)))
	for {
		select {
		case <-lp.exit:
			blog.Infof("EeventProcessor Get close event, return")
			return
		case <-updateTick.C:
			//ready to check update event
			if !lp.update {
				continue
			}
			if lp.reload {
				blog.Infof("configuration is under reloading, skip operation.")
				continue
			}
			lp.update = false
			lp.configHandle()
		case <-syncTick.C:
			if lp.reload {
				blog.Infof("configuration is under reloading, skip operation.")
				continue
			}
			lp.update = false
			lp.configHandle()
		}
	}
}

//configHandle Get all data from reflector, export to template
//to generating haproxy.cfg
func (lp *LBEventProcessor) configHandle() {
	lp.reload = true
	//Get all data from ServiceReflector
	tData := new(types.TemplateData)
	tData.HTTP, tData.HTTPS, tData.TCP, tData.UDP = lp.reflector.Lister()
	if len(tData.HTTP) == 0 && len(tData.HTTPS) == 0 && len(tData.TCP) == 0 && len(tData.UDP) == 0 {
		blog.Warnf("No any service in Reflector, try reload if changed")
	}
	tData.LogFlag = true
	tData.SSLCert = ""
	//haproxy reload
	if !lp.doReload(tData) {
		blog.Errorf("Do proxy reloading failed, wait for next tick")
	}
	lp.reload = false
}

//doReload reset HAproy configuration
func (lp *LBEventProcessor) doReload(data *types.TemplateData) bool {
	//create configuration
	newFile, creatErr := lp.cfgManager.Create(data)
	if creatErr != nil {
		blog.Errorf("Create proxy with template data faield: %s", creatErr.Error())
		return false
	}
	//check difference between new file and old file
	if !lp.cfgManager.CheckDifference(lp.config.CfgPath, newFile) {
		blog.Warnf("No difference in new configuration file")
		return true
	}
	//use check command validate correct of configuration
	if !lp.cfgManager.Validate(newFile) {
		template.LoadbalanceConfigRenderTotal.WithLabelValues("fail").Inc()
		blog.Errorf("Validate %s with proxy command failed", newFile)
		return false
	}
	template.LoadbalanceConfigRenderTotal.WithLabelValues("success").Inc()
	blog.Infof("Generation config file %s success", newFile)
	//replace new file, backup old one
	err := lp.cfgManager.Replace(lp.config.CfgPath, newFile)
	if err != nil {
		template.LoadbalanceConfigRefreshTotal.WithLabelValues("fail").Inc()
		blog.Errorf("Replace config with %s and backup failed", newFile)
		return false
	}
	template.LoadbalanceConfigRefreshTotal.WithLabelValues("success").Inc()
	//reload with haproxy command
	if err := lp.cfgManager.Reload(lp.config.CfgPath); err != nil {
		template.LoadbalanceProxyReloadTotal.WithLabelValues("fail").Inc()
		return false
	}
	template.LoadbalanceProxyReloadTotal.WithLabelValues("success").Inc()
	blog.Infof("Reload proxy config %s success.", lp.config.CfgPath)
	return true
}

//Stop stop processor all worker gracefully
func (lp *LBEventProcessor) Stop() {
	lp.reflector.Stop()
	lp.cfgManager.Stop()
	if err := lp.rd.Stop(); err != nil {
		blog.Warnf("register stop failed, err %s", err.Error())
	}
	if len(lp.config.ClusterZk) != 0 {
		if err := lp.clusterRd.Stop(); err != nil {
			blog.Warnf("cluster zk register stop failed, err %s", err.Error())
		}
	}
	lp.clearManager.Stop()
	close(lp.exit)
}

//HandleSignal interface for handle signal from system/User
func (lp *LBEventProcessor) HandleSignal(signalChan <-chan os.Signal) {
	for {
		select {
		case <-lp.exit:
			blog.Info("EventProcessor Signal Handler exit")
			return
		case <-signalChan:
			blog.Infof("Get signal from system. Exit")
			lp.Stop()
			return
		}
	}
}

//OnAdd receive data Add event
func (lp *LBEventProcessor) OnAdd(obj interface{}) {
	svr, ok := obj.(*loadbalance.ExportService)
	if !ok {
		blog.Errorf("%v is not type ExportService", obj)
		return
	}
	blog.Infof("Service %s added, ready to refresh", svr.ServiceName)
	LoadbalanceZookeeperEventMetric.WithLabelValues("add", svr.ServiceName, svr.Namespace).Inc()
	lp.update = true
}

//OnDelete receive data Delete event
func (lp *LBEventProcessor) OnDelete(obj interface{}) {
	svr, ok := obj.(*loadbalance.ExportService)
	if !ok {
		blog.Errorf("%v is not type ExportService", obj)
		return
	}
	blog.Infof("Service %s deleted, ready to refresh", svr.ServiceName)
	LoadbalanceZookeeperEventMetric.WithLabelValues("delete", svr.ServiceName, svr.Namespace).Inc()
	lp.update = true
}

//OnUpdate receive data Update event
func (lp *LBEventProcessor) OnUpdate(oldObj, newObj interface{}) {
	newSvr, ok := newObj.(*loadbalance.ExportService)
	if !ok {
		blog.Errorf("new obj %v is not type ExportService", newObj)
		return
	}
	oldSvc, ok := oldObj.(*loadbalance.ExportService)
	if !ok {
		blog.Errorf("old obj %v is not type ExportService", oldObj)
	}
	if reflect.DeepEqual(oldSvc, newSvr) {
		blog.Infof(fmt.Sprintf("Service %s No changed, skip update event", newSvr.ServiceName))
		return
	}
	blog.Infof("Service %s update, ready to refresh", newSvr.ServiceName)
	LoadbalanceZookeeperEventMetric.WithLabelValues("update", newSvr.ServiceName, newSvr.Namespace).Inc()
	lp.update = true
}
