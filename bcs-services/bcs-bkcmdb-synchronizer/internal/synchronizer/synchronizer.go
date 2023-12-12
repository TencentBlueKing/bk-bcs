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
 */

// Package synchronizer define methods for synchronizer
package synchronizer

import (
	"crypto/tls"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	cmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client"
	cm "github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/clustermanager"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/projectmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/mq"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/mq/rabbitmq"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/syncer"
)

const (
	FullSynchronizationTicker = 100000
)

// Synchronizer the synchronizer
type Synchronizer struct {
	Syncer                   *syncer.Syncer
	Handler                  handler.Handler
	BkcmdbSynchronizerOption *option.BkcmdbSynchronizerOption
	ClientTls                *tls.Config
	MQ                       mq.MQ
	//CMDBClient               client.CMDBClient
}

// ClusterList the cluster list
type ClusterList []string

// Len is a method that returns the length of the ClusterList.
func (s ClusterList) Len() int {
	return len(s)
}

// Swap is a method that swaps two elements in the ClusterList.
func (s ClusterList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less is a method that compares two elements in the ClusterList and returns true if the element at index i is less than the element at index j.
func (s ClusterList) Less(i, j int) bool {
	// Split the elements at index i and j by "-".
	is := strings.Split(s[i], "-")
	js := strings.Split(s[j], "-")

	// Convert the last part of the split elements to integers.
	idi, _ := strconv.Atoi(is[len(is)-1])
	idj, _ := strconv.Atoi(js[len(js)-1])

	// Return true if the element at index i is less than the element at index j.
	return idi < idj
}

// NewSynchronizer create a new synchronizer
func NewSynchronizer(bkcmdbSynchronizerOption *option.BkcmdbSynchronizerOption) *Synchronizer {
	return &Synchronizer{
		BkcmdbSynchronizerOption: bkcmdbSynchronizerOption,
	}
}

// Init init the synchronizer
func (s *Synchronizer) Init() {
	blog.InitLogs(s.BkcmdbSynchronizerOption.Bcslog)
	err := s.initTlsConfig()
	if err != nil {
		blog.Errorf("init tls config failed, err: %s", err.Error())
	}
	//
	//err = s.initCMDBClient()
	//if err != nil {
	//	blog.Errorf("init cmdb client failed, err: %s", err.Error())
	//}

	err = s.initSyncer()
	if err != nil {
		blog.Errorf("init syncer failed, err: %s", err.Error())
	}

	err = s.initHandler()
	if err != nil {
		blog.Errorf("init handler failed, err: %s", err.Error())
	}

	err = s.initMQ()
	if err != nil {
		blog.Errorf("init mq failed, err: %s", err.Error())
	}

}

func (s *Synchronizer) initTlsConfig() error {
	if len(s.BkcmdbSynchronizerOption.Client.ClientCrt) != 0 &&
		len(s.BkcmdbSynchronizerOption.Client.ClientKey) != 0 &&
		len(s.BkcmdbSynchronizerOption.Client.ClientCa) != 0 {
		tlsConfig, err := ssl.ClientTslConfVerity(
			s.BkcmdbSynchronizerOption.Client.ClientCa,
			s.BkcmdbSynchronizerOption.Client.ClientCrt,
			s.BkcmdbSynchronizerOption.Client.ClientKey,
			s.BkcmdbSynchronizerOption.Client.ClientCrtPwd,
		)
		//static.ClientCertPwd)
		if err != nil {
			blog.Errorf("init tls config failed, err: %s", err.Error())
			return err
		}
		s.ClientTls = tlsConfig
		blog.Infof("init tls config success")

	}
	return nil
}

func (s *Synchronizer) initSyncer() error {
	s.Syncer = syncer.NewSyncer(
		s.BkcmdbSynchronizerOption,
	)
	s.Syncer.Init()

	return nil
}

func (s *Synchronizer) initHandler() error {
	s.Handler = handler.NewBcsBkcmdbSynchronizerHandler(s.Syncer)
	return nil
}

func (s *Synchronizer) initMQ() error {
	s.MQ = rabbitmq.NewRabbitMQ(&s.BkcmdbSynchronizerOption.RabbitMQ)

	return nil
}

// Run run the synchronizer
func (s *Synchronizer) Run() {
	//rabbit := rabbitmq.NewRabbitMQ(&s.BkcmdbSynchronizerOption.RabbitMQ)
	//s.Rabbit = rabbit
	var podIndex int
	blog.Infof("BkcmdbSynchronizerOption: %v", s.BkcmdbSynchronizerOption)

	whiteList := make([]string, 0)
	blackList := make([]string, 0)

	if s.BkcmdbSynchronizerOption.Synchronizer.WhiteList != "" {
		whiteList = strings.Split(s.BkcmdbSynchronizerOption.Synchronizer.WhiteList, ",")
	}

	if s.BkcmdbSynchronizerOption.Synchronizer.BlackList != "" {
		blackList = strings.Split(s.BkcmdbSynchronizerOption.Synchronizer.BlackList, ",")
	}

	blog.Infof("whiteList: %v, len: ", whiteList, len(whiteList))
	blog.Infof("blackList: %v, len: ", blackList, len(blackList))

	hostname, err := os.Hostname()
	if err != nil {
		blog.Errorf("error: %v", err)
	} else {
		blog.Infof("Hostname : %s", hostname)
		h := strings.Split(hostname, "-")
		if len(h) > 0 {
			num, errr := strconv.Atoi(h[len(h)-1])
			if errr != nil {
				blog.Errorf("Error: %v\n", errr)
				return
			} else {
				podIndex = num
				fmt.Printf("The number is %d\n", podIndex)
			}
		}
	}

	chn, err := s.MQ.GetChannel()

	err = s.MQ.EnsureExchange(chn)
	if err != nil {
		blog.Errorf("ensure exchange failed, err: %s", err.Error())
		return
	}

	err = chn.Close()
	if err != nil {
		blog.Errorf("close channel failed, err: %s", err.Error())
		return
	}

	cmCli, err := s.getClusterManagerGrpcGwClient()
	if err != nil {
		blog.Errorf("get cluster manager grpc gw client failed, err: %s", err.Error())
		return
	}

	ticker := time.NewTicker(FullSynchronizationTicker * time.Hour)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		blog.Infof("start sync at %s", time.Now().Format("2006-01-02 15:04:05"))

		err = s.MQ.Close()
		if err != nil {
			blog.Errorf("close rabbitmq failed, err: %s", err.Error())
		}

		lcReq := cmp.ListClusterReq{}
		//if s.BkcmdbSynchronizerOption.Synchronizer.Env != "stag" {
		//	lcReq.Environment = s.BkcmdbSynchronizerOption.Synchronizer.Env
		//}

		resp, err := cmCli.Cli.ListCluster(cmCli.Ctx, &lcReq)
		if err != nil {
			blog.Errorf("list cluster failed, err: %s", err.Error())
			return
		}

		clusters := resp.Data
		clusterMap := make(map[string]*cmp.Cluster)
		var clusterList ClusterList

		for _, cluster := range clusters {
			if len(whiteList) > 0 {
				if exit, _ := common.InArray(cluster.ClusterID, whiteList); !exit {
					continue
				}
			}

			if len(blackList) > 0 {
				if exit, _ := common.InArray(cluster.ClusterID, blackList); exit {
					continue
				}
			}

			if cluster.ClusterType == "virtual" {
				continue
			}

			if _, ok := clusterMap[cluster.ClusterID]; ok {
				if cluster.IsShared {
					clusterMap[cluster.ClusterID] = cluster
				}
			} else {
				clusterMap[cluster.ClusterID] = cluster
				clusterList = append(clusterList, cluster.ClusterID)
			}

		}

		blog.Infof("clusterList: %v", clusterList)

		sort.Sort(clusterList)

		replicas := s.BkcmdbSynchronizerOption.Synchronizer.Replicas

		workList := clusterList[podIndex*len(clusterList)/replicas : (podIndex+1)*len(clusterList)/replicas]

		blog.Infof("workList: %v", workList)

		for _, w := range workList {
			go s.sync(clusterMap[w])
		}

		//for _, cluster := range clusterMap {
		//	if exist, _ := common.InArray(cluster.ClusterID, clusterList); exist {
		//		go s.sync(cluster)
		//	}
		//	////s.Sync(cluster)
		//	//go s.sync(cluster)
		//}

		blog.Infof("start consumer success")
	}
}

// Sync sync clusters
func (s *Synchronizer) Sync(cluster *cmp.Cluster) {
	go s.sync(cluster)
	//go common.Recoverer(1, func() { s.syncMQ(cluster) })
}

// Sync sync the cluster
func (s *Synchronizer) sync(cluster *cmp.Cluster) {
	if cluster.Status != "RUNNING" || cluster.EngineType != "k8s" {
		blog.Infof("skip sync cluster %s", cluster.ClusterID)
		return
	}
	blog.Infof("sync cluster: %s", cluster.ClusterID)

	chn, err := s.MQ.GetChannel()

	err = s.MQ.DeclareQueue(chn, cluster.ClusterID, amqp.Table{})
	if err != nil {
		blog.Errorf("declare queue failed, err: %s", err.Error())
		return
	}

	err = s.Syncer.SyncCluster(cluster)
	if err != nil {
		blog.Errorf("sync cluster failed, err: %s", err.Error())
		return
	}

	bkCluster, err := s.Syncer.GetBkCluster(cluster)
	if err != nil {
		blog.Errorf("get bk cluster failed, err: %s", err.Error())
		return
	}

	err = s.Syncer.SyncPods(cluster, bkCluster)
	if err != nil {
		blog.Errorf("sync pod failed, err: %s", err.Error())
	}

	err = s.Syncer.SyncWorkloads(cluster, bkCluster)
	if err != nil {
		blog.Errorf("sync workload failed, err: %s", err.Error())
	}

	err = s.Syncer.SyncNamespaces(cluster, bkCluster)
	if err != nil {
		blog.Errorf("sync namespace failed, err: %s", err.Error())
	}

	err = s.Syncer.SyncNodes(cluster, bkCluster)
	if err != nil {
		blog.Errorf("sync node failed, err: %s", err.Error())
	}

	err = s.Syncer.SyncWorkloads(cluster, bkCluster)
	if err != nil {
		blog.Errorf("sync workload failed, err: %s", err.Error())
	}

	err = s.Syncer.SyncPods(cluster, bkCluster)
	if err != nil {
		blog.Errorf("sync pod failed, err: %s", err.Error())
	}

	err = s.MQ.BindQueue(
		chn,
		cluster.ClusterID,
		fmt.Sprintf("%s.headers", s.BkcmdbSynchronizerOption.RabbitMQ.SourceExchange),
		amqp.Table{"clusterId": cluster.ClusterID},
	)
	if err != nil {
		blog.Errorf("bind queue failed, err: %s", err.Error())
		return
	}

	//h := handler.NewBcsBkcmdbSynchronizerHandler(s.Syncer)
	err = s.MQ.StartConsumer(
		chn,
		cluster.ClusterID,
		s.Handler,
	)
	if err != nil {
		blog.Errorf("start consumer failed, err: %s", err.Error())
		return
	}

}

func (s *Synchronizer) getClusterManagerGrpcGwClient() (cmCli *client.ClusterManagerClientWithHeader, err error) {
	opts := &cm.Options{
		Module:          cm.ModuleClusterManager,
		Address:         s.BkcmdbSynchronizerOption.Bcsapi.GrpcAddr,
		EtcdRegistry:    nil,
		ClientTLSConfig: s.ClientTls,
		AuthToken:       s.BkcmdbSynchronizerOption.Bcsapi.BearerToken,
	}
	cmCli, err = cm.NewClusterManagerGrpcGwClient(opts)
	return cmCli, err
}

func (s *Synchronizer) getProjectManagerGrpcGwClient() (pmCli *client.ProjectManagerClientWithHeader, err error) {
	opts := &pm.Options{
		Module:          pm.ModuleProjectManager,
		Address:         s.BkcmdbSynchronizerOption.Bcsapi.GrpcAddr,
		EtcdRegistry:    nil,
		ClientTLSConfig: s.ClientTls,
		AuthToken:       s.BkcmdbSynchronizerOption.Bcsapi.BearerToken,
	}
	pmCli, err = pm.NewProjectManagerGrpcGwClient(opts)
	return pmCli, err
}
