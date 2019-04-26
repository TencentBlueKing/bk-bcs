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
	"bk-bcs/bcs-common/common/conf"
	loadbalance "bk-bcs/bcs-common/pkg/loadbalance/v2"
	"bk-bcs/bcs-services/bcs-loadbalance/clear"
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
)

// EventHandler is event interface when
type EventHandler interface {
	OnAdd(obj interface{})
	OnDelete(obj interface{})
	OnUpdate(oldObj, newObj interface{})
}

// NewDefaultBlogCfg construct default blog config
func NewDefaultBlogCfg() conf.LogConfig {
	return conf.LogConfig{
		LogDir:          "./logs",
		LogMaxSize:      500,
		LogMaxNum:       10,
		StdErrThreshold: "2",
	}
}

//InitLogger init app logger
func InitLogger(config *option.LBConfig) {
	blog.InitLogs(NewDefaultBlogCfg())
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
	zkSubRegPath := config.ClusterID + "/" + config.Group
	processor.rd = rdiscover.NewRDiscover(config.BcsZkAddr, zkSubRegPath, config.ClusterID, config.Proxy, config.MetricPort)
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

	if config.Proxy == option.ProxyHaproxy {
		blog.Infof("use haproxy transmit")
		processor.cfgManager = haproxy.NewManager(
			config.BinPath,
			config.CfgPath,
			config.GeneratingDir,
			config.CfgBackupDir,
			config.TemplateDir)
	} else {
		blog.Infof("use nginx transmit")
		processor.cfgManager = nginx.NewManager(
			config.BinPath,
			config.CfgPath,
			config.GeneratingDir,
			config.CfgBackupDir,
			config.TemplateDir)
	}

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
	clearManager *clear.Manager       //timer to clear template file
}

//Start starting point for event processing
//1. start reflector to cache data from storage
//2. start template manager for Create/Reload config for haproxy.cfg
//3. start local logic loop for check data changed
func (lp *LBEventProcessor) Start() error {
	//step 0 (step 4 before,change on 2018/2/26 by developerJim)
	go func() {
		if err := lp.rd.Start(); err != nil {
			blog.Errorf("start register zookeeper error: %s", err.Error())
			//should go ahead to work event if register zookeeper failed
		}
	}()
	blog.Infof("start register success")
	//step 1
	if err := lp.reflector.Start(); err != nil {
		blog.Errorf("start Reflector error: %s", err.Error())
		return err
	}
	blog.Infof("start reflector success")
	//step 2, whether is master depend on step 0
	if err := lp.cfgManager.Start(); err != nil {
		blog.Errorf("start ConfigManager error: %s", err.Error())
		return err
	}
	//step 3
	lp.clearManager.Start()

	//register metric and healthz check
	if err := lp.metricRegister(); err != nil {
		blog.Warnf("register metric failed, err %s", err.Error())
	}

	//step 5
	lp.run()
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
	} else {
		blog.Infof("Reload proxy config %s success.", lp.config.CfgPath)
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
		return false
	}
	//use check command validate correct of configuration
	if !lp.cfgManager.Validate(newFile) {
		blog.Errorf("Validate %s with proxy command failed", newFile)
		return false
	}
	blog.Infof("Generation config file %s success", newFile)
	//replace new file, backup old one
	err := lp.cfgManager.Replace(lp.config.CfgPath, newFile)
	if err != nil {
		blog.Errorf("Replace config with %s and backup failed", newFile)
		return false
	}
	//reload with haproxy command
	if err := lp.cfgManager.Reload(lp.config.CfgPath); err != nil {
		return false
	}
	return true
}

//Stop stop processor all worker gracefully
func (lp *LBEventProcessor) Stop() {
	lp.reflector.Stop()
	lp.cfgManager.Stop()
	if err := lp.rd.Stop(); err != nil {
		blog.Warnf("register stop failed, err %s", err.Error())
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
	lp.update = true
}
