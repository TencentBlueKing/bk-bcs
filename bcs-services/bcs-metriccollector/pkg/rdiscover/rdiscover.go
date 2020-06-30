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

package rdiscover

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"

	"golang.org/x/net/context"

	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

//MetricServers define struct include all routes server information
type MetricServers struct {
	metricServInfo   []*types.MetricServiceInfo
	exporterServInfo []*types.DataExporterInfo
}

//RDiscover route register and discover
type RDiscover struct {
	rd      *RegisterDiscover.RegDiscover
	cancel  context.CancelFunc
	dsStore *MetricServers
	dsLock  sync.RWMutex
}

//NewRDiscover create a object of RDiscover
func NewRDiscover(zkserv string) *RDiscover {
	return &RDiscover{

		rd:      RegisterDiscover.NewRegDiscoverEx(zkserv, 10*time.Second),
		cancel:  nil,
		dsStore: &MetricServers{},
	}
}

//Start the rdiscover
func (r *RDiscover) Start() error {
	//create root context
	rootctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	//start regdiscover
	if err := r.rd.Start(); err != nil {
		blog.Error("fail to start register and discover serv. err:%s", err.Error())
		return err
	}

	// discover other bcs service
	// metric service
	metriceServicePath := types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_METRICSERVICE
	metricEvent, err := r.rd.DiscoverService(metriceServicePath)
	if err != nil {
		blog.Error("fail to register discover for metric service. err:%s", err.Error())
		return err
	}

	// exporter service
	exporterServicePath := types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_EXPORTER
	exporterEvent, err := r.rd.DiscoverService(exporterServicePath)
	if err != nil {
		blog.Error("fail to register discover for metric service. err:%s", err.Error())
		return err
	}

	go func() {
		for {
			select {
			case metricEnv := <-metricEvent:
				r.dealMetricServer(rootctx, metricEnv.Nodes)
			case exporterEnv := <-exporterEvent:
				r.dealExporterServer(rootctx, exporterEnv.Nodes)
			case <-rootctx.Done():
				blog.Warn("route register and discover done")
				return
			}
		}

	}()
	return nil
}

//Stop the rdiscover
func (r *RDiscover) Stop() error {
	r.cancel()

	r.rd.Stop()

	return nil
}

// GetMetricServer fetch metric server
func (r *RDiscover) GetMetricServer() (string, error) {
	r.dsLock.RLock()
	defer r.dsLock.RUnlock()

	if len(r.dsStore.metricServInfo) <= 0 {
		err := fmt.Errorf("there is no metric server to use")
		blog.Error("%s", err.Error())
		return "", err
	}

	//rand
	rand.Seed(int64(time.Now().Nanosecond()))
	lServ := len(r.dsStore.metricServInfo)
	servInfo := r.dsStore.metricServInfo[rand.Intn(lServ)]

	host := servInfo.Scheme + "://" + servInfo.IP + ":" + strconv.Itoa(int(servInfo.Port))
	return host, nil
}

// GetExporterServer fetch exporter server
func (r *RDiscover) GetExporterServer() (string, error) {
	r.dsLock.RLock()
	defer r.dsLock.RUnlock()

	if len(r.dsStore.exporterServInfo) <= 0 {
		err := fmt.Errorf("there is no exporter server to use")
		blog.Error("%s", err.Error())
		return "", err
	}

	//rand
	rand.Seed(int64(time.Now().Nanosecond()))
	lServ := len(r.dsStore.exporterServInfo)
	servInfo := r.dsStore.exporterServInfo[rand.Intn(lServ)]
	host := servInfo.Scheme + "://" + servInfo.IP + ":" + strconv.Itoa(int(servInfo.Port))
	return host, nil
}

func (r *RDiscover) dealMetricServer(ctx context.Context, nodes []string) error {
	blog.Info("deal metric server node:%v", nodes)

	mdCxt, _ := context.WithCancel(ctx)

	for _, node := range nodes {
		go r.discoverMetricServer(mdCxt, node)
	}

	return nil
}

func (r *RDiscover) dealExporterServer(ctx context.Context, nodes []string) error {

	blog.Info("deal exporter server node:%v", nodes)
	mdCxt, _ := context.WithCancel(ctx)
	for _, node := range nodes {
		go r.discoverExporterServer(mdCxt, node)
	}

	return nil
}

func (r *RDiscover) discoverMetricServer(mdCxt context.Context, node string) {

	path := types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_METRICSERVICE

	mdEvent, err := r.rd.DiscoverService(path)
	if err != nil {
		blog.Error("fail to discover mesosdriver(%s), err:%s", path, err.Error())
		return
	}

	for {
		select {
		case mds := <-mdEvent:
			r.saveMetricServer(mds.Server)
		case <-mdCxt.Done():
			blog.Warn("discover metric server(%s) done", path)
			return
		}
	}
}

func (r *RDiscover) saveMetricServer(servs []string) {

	blog.Info("discover metric server(%v) ", servs)
	if len(servs) <= 0 {
		blog.Warn("thers is no metric server")
		return
	}

	dsServs := &MetricServers{
		metricServInfo: []*types.MetricServiceInfo{},
	}

	for _, serv := range servs {

		servInfo := new(types.MetricServiceInfo)
		if err := json.Unmarshal([]byte(serv), servInfo); err != nil {
			blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
			continue
		}

		dsServs.metricServInfo = append(dsServs.metricServInfo, servInfo)
	}

	r.dsLock.Lock()
	defer r.dsLock.Unlock()
	r.dsStore.metricServInfo = dsServs.metricServInfo
}

func (r *RDiscover) discoverExporterServer(mdCxt context.Context, node string) {

	path := types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_EXPORTER

	mdEvent, err := r.rd.DiscoverService(path)
	if err != nil {
		blog.Error("fail to discover mesosdriver(%s), err:%s", path, err.Error())
		return
	}

	for {
		select {
		case mds := <-mdEvent:
			r.saveExporterServer(mds.Server)
		case <-mdCxt.Done():
			blog.Warn("discover exporter server(%s) done", path)
			return
		}
	}
}
func (r *RDiscover) saveExporterServer(servs []string) {

	blog.Info("discover metric server(%v) ", servs)
	if len(servs) <= 0 {
		blog.Warn("thers is no metric server")
		return
	}

	dsServs := &MetricServers{
		exporterServInfo: []*types.DataExporterInfo{},
	}

	for _, serv := range servs {

		servInfo := new(types.DataExporterInfo)
		if err := json.Unmarshal([]byte(serv), servInfo); err != nil {
			blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
			continue
		}

		dsServs.exporterServInfo = append(dsServs.exporterServInfo, servInfo)
	}

	r.dsLock.Lock()
	defer r.dsLock.Unlock()
	r.dsStore.exporterServInfo = dsServs.exporterServInfo
}
