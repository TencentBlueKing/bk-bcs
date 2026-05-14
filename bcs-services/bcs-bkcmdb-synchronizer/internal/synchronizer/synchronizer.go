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
// nolint
package synchronizer

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	_ "net/http/pprof" //nolint
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	bkcmdbkube "configcenter/src/kube/types" // nolint

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/bkuser"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/cache"
	cm "github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/constants"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/mq"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/mq/rabbitmq"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/store/db/sqlite"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/syncer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/tenant"
	pmp "github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/types"
)

const (
	// FullSynchronizationTicker xxx
	FullSynchronizationTicker = 30
)

// Synchronizer the synchronizer
type Synchronizer struct {
	Syncer                   *syncer.Syncer
	Handler                  handler.Handler
	BkcmdbSynchronizerOption *option.BkcmdbSynchronizerOption
	ClientTls                *tls.Config
	MQ                       mq.MQ
	// CMDBClient               client.CMDBClient
}

// NewSynchronizer create a new synchronizer
func NewSynchronizer(bkcmdbSynchronizerOption *option.BkcmdbSynchronizerOption) (*Synchronizer, error) {
	option.SetGlobalConfig(bkcmdbSynchronizerOption)
	cache.InitCache()

	if bkcmdbSynchronizerOption.Synchronizer.EnableMultiTenantMode {
		err := bkuser.SetBkUserClient(bkuser.Options{
			AppCode:   bkcmdbSynchronizerOption.BkUser.AppCode,
			AppSecret: bkcmdbSynchronizerOption.BkUser.AppSecret,
			Server:    bkcmdbSynchronizerOption.BkUser.Server,
			Debug:     bkcmdbSynchronizerOption.BkUser.Debug,
		})
		if err != nil {
			return nil, err
		}
	}

	tlsConfig, err := option.InitTClientTlsConfig()
	if err != nil {
		blog.Errorf("init tls config failed, err: %s", err.Error())
		return nil, err
	}

	return &Synchronizer{
		BkcmdbSynchronizerOption: bkcmdbSynchronizerOption,
		ClientTls:                tlsConfig,
	}, nil
}

// Init init the synchronizer
func (s *Synchronizer) Init() {
	blog.InitLogs(s.BkcmdbSynchronizerOption.Bcslog)

	err := s.initSyncer()
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

	s.initSharedClusterConf()

	// 添加metrics初始化
	err = s.initMetrics()
	if err != nil {
		blog.Errorf("init metrics failed, err: %s", err.Error())
	}
}

// nolint (error) is always nil
func (s *Synchronizer) initSyncer() error {
	s.Syncer = syncer.NewSyncer(
		s.BkcmdbSynchronizerOption,
	)
	s.Syncer.Init()

	return nil
}

// nolint (error) is always nil
func (s *Synchronizer) initHandler() error {
	s.Handler = handler.NewBcsBkcmdbSynchronizerHandler(s.Syncer)
	return nil
}

// nolint (error) is always nil
func (s *Synchronizer) initMQ() error {
	s.MQ = rabbitmq.NewRabbitMQ(&s.BkcmdbSynchronizerOption.RabbitMQ)

	return nil
}

func (s *Synchronizer) initSharedClusterConf() {
	if s.Syncer.BkcmdbSynchronizerOption.SharedCluster.AnnotationKeyProjCode == "" {
		s.Syncer.BkcmdbSynchronizerOption.SharedCluster.AnnotationKeyProjCode = "io.tencent.bcs.projectcode"
	}
}

// initMetrics 初始化metrics相关组件
// nolint
func (s *Synchronizer) initMetrics() error {
	blog.Infof("init metrics...")

	// 设置默认端口
	if s.BkcmdbSynchronizerOption.Metrics.Port == 0 {
		s.BkcmdbSynchronizerOption.Metrics.Port = 8082
	}

	// 启动metrics服务器
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())

		addr := fmt.Sprintf(":%d", s.BkcmdbSynchronizerOption.Metrics.Port)
		blog.Infof("starting metrics server on %s", addr)

		if err := http.ListenAndServe(addr, mux); err != nil {
			blog.Errorf("metrics server failed to start: %v", err)
		}
	}()

	blog.Infof("init metrics success on port %d", s.BkcmdbSynchronizerOption.Metrics.Port)
	blog.Infof("CMDB metrics registered:")
	blog.Infof("  - bkbcs_cmdbsynchronizer_cmdb_requests_total (Counter)")
	blog.Infof("  - bkbcs_cmdbsynchronizer_cmdb_request_duration_seconds (Histogram)")
	blog.Infof("  - bkbcs_cmdbsynchronizer_cmdb_requests_in_flight (Gauge)")

	return nil
}

// Run run the synchronizer
// nolint funlen
func (s *Synchronizer) Run() {
	//rabbit := rabbitmq.NewRabbitMQ(&s.BkcmdbSynchronizerOption.RabbitMQ)
	//s.Rabbit = rabbit
	var podIndex int
	blog.Infof("BkcmdbSynchronizerOption: %v", s.BkcmdbSynchronizerOption)

	// 白名单 & 黑名单集群
	whiteList := make([]string, 0)
	blackList := make([]string, 0)

	if s.BkcmdbSynchronizerOption.Synchronizer.WhiteList != "" {
		whiteList = strings.Split(s.BkcmdbSynchronizerOption.Synchronizer.WhiteList, ",")
	}

	if s.BkcmdbSynchronizerOption.Synchronizer.BlackList != "" {
		blackList = strings.Split(s.BkcmdbSynchronizerOption.Synchronizer.BlackList, ",")
	}

	blog.Infof("whiteList: %v, len: %d", whiteList, len(whiteList))
	blog.Infof("blackList: %v, len: %d", blackList, len(blackList))

	// 获取hostname, 通过hostname获取podIndex
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
			}
			podIndex = num
			fmt.Printf("The number is %d\n", podIndex)
		}
	}

	chn, err := s.MQ.GetChannel()
	defer chn.Close()

	// 注册 headers类型的exchange，并镜像exchange: bcs-storage 的消息
	// headers类型的Exchange不依赖于routing key与binding key的匹配规则来路由消息，是根据发送的消息内容中的headers属性进行匹配
	err = s.MQ.EnsureExchange(chn)
	if err != nil {
		blog.Errorf("ensure exchange failed, err: %s", err.Error())
		return
	}

	//err = chn.Close()
	//if err != nil {
	//	blog.Errorf("close channel failed, err: %s", err.Error())
	//	return
	//}

	blog.Infof("start sync at %s", time.Now().Format("2006-01-02 15:04:05"))

	//err = s.MQ.Close()
	//if err != nil {
	//	blog.Errorf("close rabbitmq failed, err: %s", err.Error())
	//}

	workList, clusterMap, _, err := s.getWorkList(podIndex, whiteList, blackList)
	if err != nil {
		blog.Errorf("get work list failed, err: %s", err.Error())
		return
	}

	blog.Infof("workList: %v", workList)

	gm := common.NewGoroutineManager(s.syncWorker)
	for _, w := range workList {
		blog.Infof("%s started", w)
		gm.Start(w, clusterMap[w])
	}

	// 处理新增删除集群逻辑
	go func() {
		time.Sleep(time.Second * 10)
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		prevWorkList, prevClusterMap := workList, clusterMap
		for ; true; <-ticker.C {
			workListT, clusterMapT, _, err := s.getWorkList(podIndex, whiteList, blackList)
			if err != nil {
				blog.Errorf("get work list failed, err: %s", err.Error())
				continue
			}

			for _, wT := range workListT {
				if exist, _ := common.InArray(wT, prevWorkList); !exist {
					blog.Infof("%s started, performing full sync first", wT)
					// 发现新集群时，先进行一次全量同步
					s.startSyncStorage(clusterMapT[wT])
					gm.Start(wT, clusterMapT[wT])
				}
			}

			for _, wT := range prevWorkList {
				if exist, _ := common.InArray(wT, workListT); !exist {
					blog.Infof("%s stopped", wT)
					gm.Stop(wT, prevClusterMap[wT])
				}
			}

			// 只有主Pod（podIndex=0）负责清理CMDB中多余的集群数据
			if podIndex == 0 {
				err := s.cleanupOrphanedClustersInCMDB(clusterMapT)
				if err != nil {
					blog.Errorf("cleanup orphaned clusters failed: %s", err.Error())
				}
			}
			prevWorkList, prevClusterMap = workListT, clusterMapT
		}
	}()

	// http server url
	go func() {
		http.HandleFunc("/restart", common.HandleRestart(gm))
		http.HandleFunc("/list", common.HandleList(gm))
		http.HandleFunc("/worklist", s.workListHandler(podIndex, whiteList, blackList))

		http.HandleFunc("/syncStorage", s.syncStorageHandler(podIndex, whiteList, blackList))
		http.HandleFunc("/syncStore", s.syncStoreHandler(podIndex, whiteList, blackList))
		http.HandleFunc("/sync", s.syncHandler(podIndex, whiteList, blackList))

		if err := http.ListenAndServe(":8080", nil); err != nil {
			blog.Errorf("Goroutine Manager start error: %v\n", err)
		}
	}()

	// 集群维度数据强制同步
	go func() {
		time.Sleep(time.Second * 10)
		ticker := time.NewTicker(60 * time.Minute)
		defer ticker.Stop()
		for ; true; <-ticker.C {
			blog.Infof("ticker syncStorage")
			workListS, clusterMapS, _, err := s.getWorkList(podIndex, whiteList, blackList)
			if err != nil {
				blog.Errorf("get work list failed, err: %s", err.Error())
				continue
			}

			for _, w := range workListS {
				go s.startSyncStorage(clusterMapS[w])
				time.Sleep(time.Minute)
			}
		}
	}()

	blog.Infof("start consumer success")

	//for _, w := range workList {
	//	blog.Infof("%s started syncStorage", w)
	//	bkCluster, err := s.Syncer.GetBkCluster(clusterMap[w])
	//	if err != nil {
	//		blog.Errorf("get bk cluster failed, err: %s", err.Error())
	//		continue
	//	}
	//	s.syncStorage(clusterMap[w], bkCluster)
	//}

	//for _, w := range workList {
	//	blog.Infof("%s started syncStore", w)
	//	bkCluster, err := s.Syncer.GetBkCluster(clusterMap[w], nil, false)
	//	if err != nil {
	//		blog.Errorf("get bk cluster failed, err: %s", err.Error())
	//		continue
	//	}
	//	s.syncStore(bkCluster)
	//}

	tickerChecker := time.NewTicker(5 * time.Minute)
	defer tickerChecker.Stop()
	for ; true; <-tickerChecker.C {
		blog.Infof("tickerChecker")
		// 重新获取最新的workList和clusterMap
		workListT, clusterMapT, _, err := s.getWorkList(podIndex, whiteList, blackList)
		if err != nil {
			blog.Errorf("get work list failed, err: %s", err.Error())
			continue
		}

		for _, w := range workListT {
			cluster := clusterMapT[w]
			chnQ, _ := s.MQ.GetChannel()
			if qInfo, errQ := chnQ.QueueInspect(w); errQ != nil {
				blog.Errorf("Failed to inspect the queue %s: %v", w, errQ)
			} else {
				blog.Infof("Messages in queue %s: %d", w, qInfo.Messages)
				if qInfo.Messages > 1000000 {
					_, err = chnQ.QueuePurge(w, false)
					if err != nil {
						blog.Errorf("Failed to delete the queue %s: %v", w, err)
					}
					gm.Restart(w, cluster)
					//bkCluster, err := s.Syncer.GetBkCluster(clusterMap[w])
					//if err != nil {
					//	blog.Errorf("get bk cluster failed, err: %s", err.Error())
					//	return
					//}
					//go s.syncStorage(clusterMap[w], bkCluster)
					blog.Infof("Messages in queue %s: %d, is greater than 10000, restarting", w, qInfo.Messages)
					continue
				}
			}
			chnQ.Close()
			if exit, _ := common.InArray(w, gm.List()); !exit {
				gm.Start(w, cluster)
				blog.Infof("%s restarted", w)
			}
		}
	}
}

// startSyncStorage start sync storage for cluster
func (s *Synchronizer) startSyncStorage(cluster *cmp.Cluster) {
	// Check if cluster is nil
	if cluster == nil {
		blog.Errorf("cluster is nil in startSyncStorage")
		return
	}

	ctx, errLocal := tenant.WithTenantIdByResourceForContext(context.Background(), tenant.ResourceMetaData{
		ClusterId: cluster.ClusterID,
	})
	if errLocal != nil {
		blog.Infof("Synchronizer sync cluster %s failed: %v", cluster.ClusterID, errLocal)
		return
	}

	if err := s.Syncer.SyncCluster(ctx, cluster); err != nil {
		blog.Errorf("sync cluster failed, err: %s", err.Error())
		return
	}

	bkCluster, err := s.Syncer.GetBkCluster(ctx, cluster, nil, false)
	if err != nil {
		blog.Errorf("get bk cluster failed for %s, err: %s", cluster.ClusterID, err.Error())
		return
	}

	// 执行全量同步
	s.syncStorage(ctx, cluster, bkCluster, false)
}

// getWorkList get work list for podIndex
func (s *Synchronizer) getWorkList(podIndex int, whiteList, blackList []string) (ClusterList, map[string]*cmp.Cluster,
	ClusterList, error) {
	cmCli, err := s.getClusterManagerGrpcGwClient()
	if err != nil {
		blog.Errorf("get cluster manager grpc gw client failed, err: %s", err.Error())
		return nil, nil, nil, err
	}

	lcReq := cmp.ListClusterReq{}
	resp, err := cmCli.Cli.ListCluster(cmCli.Ctx, &lcReq)
	if err != nil {
		blog.Errorf("list cluster failed, err: %s", err.Error())
		return nil, nil, nil, err
	}

	clusters := resp.Data

	clusterMap := make(map[string]*cmp.Cluster)
	var clusterList ClusterList
	s.runCluster(clusters, whiteList, blackList, clusterMap, &clusterList)

	blog.Infof("clusterList: %v", clusterList)

	sort.Sort(clusterList)

	replicas := s.BkcmdbSynchronizerOption.Synchronizer.Replicas
	workList := clusterList[podIndex*len(clusterList)/replicas : (podIndex+1)*len(clusterList)/replicas]
	blog.Infof("workList: %v", workList)
	return workList, clusterMap, clusterList, nil
}

func (s *Synchronizer) runCluster(clusters []*cmp.Cluster, whiteList, blackList []string,
	clusterMap map[string]*cmp.Cluster, clusterList *ClusterList) {
	for _, cluster := range clusters {
		blog.Infof("Synchronizer.runCluster clusterID: %s", cluster.ClusterID)
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

		// 过滤vcluster虚拟集群
		if cluster.ClusterType == "virtual" {
			continue
		}

		// 去重复共享集群
		if _, ok := clusterMap[cluster.ClusterID]; ok {
			if cluster.IsShared {
				clusterMap[cluster.ClusterID] = cluster
			}
		} else {
			clusterMap[cluster.ClusterID] = cluster
			*clusterList = append(*clusterList, cluster.ClusterID)
		}

	}
}

func (s *Synchronizer) syncWorker(done <-chan bool, input interface{}) {
	cluster := input.(*cmp.Cluster)
	for {
		select {
		case <-done: // 监听停止信号
			blog.Infof("syncWorker goroutine %s stopped", cluster.ClusterID)
			return
		default:
			s.sync(done, cluster) // 执行业务逻辑

			// 可中断休眠
			select {
			case <-done:
				blog.Infof("syncWorker goroutine %s stopped", cluster.ClusterID)
				return
			case <-time.After(5 * time.Second):
			}
		}
	}
}

// Sync sync clusters
func (s *Synchronizer) Sync(cluster *cmp.Cluster) {
	// go s.sync(cluster)
	// go common.Recoverer(1, func() { s.syncMQ(cluster) })
}

func (s *Synchronizer) workListHandler(podIndex int, whiteList, blackList []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 动态获取最新的工作列表
		workList, _, _, err := s.getWorkList(podIndex, whiteList, blackList)
		if err != nil {
			blog.Errorf("get work list failed, err: %s", err.Error())
			http.Error(w, "get work list failed", http.StatusInternalServerError)
			return
		}

		for _, id := range workList {
			fmt.Fprintf(w, "BcsClusterID: %s\n", id)
		}
	}
}

func (s *Synchronizer) syncStorageHandler(podIndex int, whiteList, blackList []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clusterId := r.URL.Query().Get("cluster")
		if clusterId == "" {
			http.Error(w, "缺少cluster", http.StatusBadRequest)
			return
		}

		ctx, errLocal := tenant.WithTenantIdByResourceForContext(context.Background(), tenant.ResourceMetaData{
			ClusterId: clusterId,
		})
		if errLocal != nil {
			http.Error(w, fmt.Sprintf("syncStorageHandler[%s] failed: %v",
				clusterId, errLocal), http.StatusBadRequest)
			return
		}

		// 动态获取最新的集群数据
		_, clusterMap, _, err := s.getWorkList(podIndex, whiteList, blackList)
		if err != nil {
			blog.Errorf("get latest cluster data failed, err: %s", err.Error())
			http.Error(w, "get latest cluster data failed", http.StatusInternalServerError)
			return
		}

		cluster, exists := clusterMap[clusterId]
		if !exists {
			http.Error(w, "cluster not found", http.StatusBadRequest)
			return
		}

		bkCluster, err := s.Syncer.GetBkCluster(ctx, cluster, nil, false)
		if err != nil {
			blog.Errorf("get bk cluster failed, err: %s", err.Error())
			http.Error(w, "get bk cluster failed", http.StatusBadRequest)
			return
		}

		go s.syncStorage(ctx, cluster, bkCluster, true)
		fmt.Fprintf(w, "BcsClusterID: %s\n syncStorage started.", clusterId)
	}
}

func (s *Synchronizer) syncHandler(podIndex int, whiteList, blackList []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clusterId := r.URL.Query().Get("cluster")
		if clusterId == "" {
			http.Error(w, "缺少cluster", http.StatusBadRequest)
			return
		}

		// 动态获取最新的集群数据
		_, _, clusterList, err := s.getWorkList(podIndex, whiteList, blackList)
		if err != nil {
			blog.Errorf("get latest cluster data failed, err: %s", err.Error())
			http.Error(w, "get latest cluster data failed", http.StatusInternalServerError)
			return
		}

		sort.Sort(clusterList)
		replicas := s.BkcmdbSynchronizerOption.Synchronizer.Replicas

		index := -1

		for i := 0; i < replicas; i++ {
			if exist, _ := common.InArray(clusterId,
				clusterList[i*len(clusterList)/replicas:(i+1)*len(clusterList)/replicas]); exist {
				index = i
				break
			}
		}

		if index == -1 {
			http.Error(w, "cluster not found", http.StatusBadRequest)
			return
		}

		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}
		// 使用 LookupCNAME 函数来查找 CNAME
		fqdn, err := net.LookupCNAME(hostname)
		if err != nil {
			blog.Errorf("Error looking up CNAME: %v", err)
		} else {
			blog.Infof("Fully Qualified Domain Name: %s", fqdn)
		}

		re := regexp.MustCompile("[0-9]+")
		replaced := re.ReplaceAllString(fqdn, strconv.Itoa(index))
		forwardUrl := fmt.Sprintf("http://%s:8080/syncStorage?cluster=%s",
			strings.TrimSuffix(replaced, "."), clusterId)

		// 发送 GET 请求
		resp, err := http.Get(forwardUrl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			blog.Errorf("Error: %v", err)
			return
		}
		defer resp.Body.Close() // 确保在函数返回时关闭响应体

		// 检查响应状态码
		if resp.StatusCode != http.StatusOK {
			http.Error(w, fmt.Sprintf("Error: received non-200 response code: %d\n", resp.StatusCode),
				http.StatusBadRequest)
			blog.Errorf("Error: received non-200 response code: %d\n", resp.StatusCode)
			return
		}

		// 读取响应体
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			blog.Errorf("Error reading response body: %v", err)
			return
		}

		// 输出响应体
		blog.Infof("Response Body: %s", string(body))

		blog.Infof("BcsClusterID: %s index: %d, url: %s\n, forward response: %s",
			clusterId, index, forwardUrl, string(body))
		fmt.Fprintf(w, "BcsClusterID: %s sync started.\n", clusterId)
	}
}

func (s *Synchronizer) syncStoreHandler(podIndex int, whiteList, blackList []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clusterId := r.URL.Query().Get("cluster")
		if clusterId == "" {
			http.Error(w, "缺少cluster", http.StatusBadRequest)
			return
		}

		ctx, errLocal := tenant.WithTenantIdByResourceForContext(context.Background(), tenant.ResourceMetaData{
			ClusterId: clusterId,
		})
		if errLocal != nil {
			http.Error(w, fmt.Sprintf("syncStorageHandler[%s] failed: %v",
				clusterId, errLocal), http.StatusBadRequest)
			return
		}

		// 动态获取最新的集群数据
		_, clusterMap, _, err := s.getWorkList(podIndex, whiteList, blackList)
		if err != nil {
			blog.Errorf("get latest cluster data failed, err: %s", err.Error())
			http.Error(w, "get latest cluster data failed", http.StatusInternalServerError)
			return
		}

		cluster, exists := clusterMap[clusterId]
		if !exists {
			http.Error(w, "cluster not found", http.StatusBadRequest)
			return
		}

		bkCluster, err := s.Syncer.GetBkCluster(ctx, cluster, nil, false)
		if err != nil {
			blog.Errorf("get bk cluster failed, err: %s", err.Error())
			http.Error(w, "get bk cluster failed", http.StatusBadRequest)
			return
		}

		go s.syncStore(ctx, bkCluster, true)
		fmt.Fprintf(w, "BcsClusterID: %s\n syncStore started.", clusterId)
	}
}

func (s *Synchronizer) syncStorage(ctx context.Context,
	cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, force bool) {
	path := "/data/bcs/bcs-bkcmdb-synchronizer/db/" + bkCluster.Uid + ".db"

	db := sqlite.Open(path)
	if db == nil {
		blog.Errorf("open db failed, path: %s", path)
	}

	s.syncStore(ctx, bkCluster, force)
	blog.Infof("syncStorage %s started", cluster.ClusterID)
	// err := s.Syncer.SyncPods(cluster, bkCluster, db)
	// if err != nil {
	//	blog.Errorf("sync pod failed, err: %s", err.Error())
	// }

	// err := s.Syncer.SyncWorkloads(cluster, bkCluster, db)
	// if err != nil {
	//	blog.Errorf("sync workload failed, err: %s", err.Error())
	// }

	err := s.Syncer.SyncNamespaces(ctx, cluster, bkCluster, db)
	if err != nil {
		blog.Errorf("sync namespace failed, err: %s", err.Error())
	}

	err = s.Syncer.SyncNodes(ctx, cluster, bkCluster, db)
	if err != nil {
		blog.Errorf("sync node failed, err: %s", err.Error())
	}

	err = s.Syncer.SyncWorkloads(ctx, cluster, bkCluster, db)
	if err != nil {
		blog.Errorf("sync workload failed, err: %s", err.Error())
	}

	err = s.Syncer.SyncPods(ctx, cluster, bkCluster, db)
	if err != nil {
		blog.Errorf("sync pod failed, err: %s", err.Error())
	}
	blog.Infof("syncStorage %s finished", cluster.ClusterID)
}

func (s *Synchronizer) syncStore(ctx context.Context, bkCluster *bkcmdbkube.Cluster, force bool) {
	blog.Infof("syncStore %s started", bkCluster.Uid)
	err := s.Syncer.SyncStore(ctx, bkCluster, force)
	if err != nil {
		blog.Errorf("SyncStore failed, err: %s", err.Error())
	}
}

// Sync sync the cluster
// nolint funlen
func (s *Synchronizer) sync(done <-chan bool, cluster *cmp.Cluster) {
	ctx, err := tenant.WithTenantIdByResourceForContext(context.Background(), tenant.ResourceMetaData{
		ClusterId: cluster.ClusterID,
	})
	if err != nil {
		blog.Infof("Synchronizer sync cluster %s failed: %v", cluster.ClusterID, err)
		return
	}

	if cluster.Status != "RUNNING" || cluster.EngineType != "k8s" {
		blog.Infof("skip sync cluster %s", cluster.ClusterID)

		bkCluster, err := s.Syncer.GetBkCluster(ctx, cluster, nil, false)
		if err != nil {
			blog.Errorf("get bk cluster failed, err: %s", err.Error())
			return
		}

		err = s.Syncer.DeleteAllByCluster(ctx, bkCluster)
		if err != nil {
			blog.Errorf("DeleteAllByCluster err: %s", err.Error())
		}
		return
	}
	blog.Infof("sync cluster: %s", cluster.ClusterID)

	chn, _ := s.MQ.GetChannel()

	err = s.MQ.DeclareQueue(chn, cluster.ClusterID, amqp.Table{})
	if err != nil {
		blog.Errorf("declare queue failed, err: %s", err.Error())
		return
	}

	err = s.Syncer.SyncCluster(ctx, cluster)
	if err != nil {
		blog.Errorf("sync cluster failed, err: %s", err.Error())
		return
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

	var wg sync.WaitGroup
	time.Sleep(time.Second * 10)

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			consumer := fmt.Sprintf("%s.%s.%d", hostname, cluster.ClusterID, i)
			err = s.MQ.StartConsumer(
				chn,
				consumer,
				cluster.ClusterID,
				s.Handler,
				done,
			)

			if err != nil {
				blog.Errorf("start consumer failed, err: %s", err.Error())
				// return
			}

		}(i)
	}

	wg.Wait()
}
func (s *Synchronizer) getClusterManagerGrpcGwClient() (cmCli *client.ClusterManagerClientWithHeader, err error) {
	return cm.GetClusterManagerGrpcGwClient()
}

// cleanupOrphanedClustersInCMDB 清理CMDB中孤立的集群数据
// 只有主Pod（podIndex=0）会执行此操作，避免重复删除
func (s *Synchronizer) cleanupOrphanedClustersInCMDB(clusterMapT map[string]*cmp.Cluster) error {
	blog.Infof("start cleanup orphaned clusters in CMDB")

	// 1. 构建ClusterManager中有效的集群ID列表
	validClusterMap := make(map[string]bool)
	for clusterID := range clusterMapT {
		validClusterMap[clusterID] = true
	}

	pmCli, err := s.Syncer.GetProjectManagerGrpcGwClient()
	if err != nil {
		blog.Errorf("get project manager client failed: %v", err)
		return fmt.Errorf("get project manager client failed: %v", err)
	}

	// 2. 获取所有业务ID - 通过项目获取业务信息
	projectResp, err := pmCli.Cli.ListProjects(pmCli.Ctx, &pmp.ListProjectsRequest{})
	if err != nil {
		blog.Errorf("get all projects from project manager failed: %v", err)
		return fmt.Errorf("get all projects from project manager failed: %v", err)
	}

	if projectResp.Data == nil || len(projectResp.Data.Results) == 0 {
		blog.Infof("no projects found, skipping cleanup")
		return nil
	}

	// 3. 从项目中提取唯一的业务ID，避免重复处理
	uniqueBusinessMap := make(map[string]string)
	uniqueBusinessTenantMap := make(map[string]string)
	for _, project := range projectResp.Data.Results {
		uniqueBusinessMap[project.BusinessID] = project.BusinessName
		uniqueBusinessTenantMap[project.BusinessID] = project.TenantID
	}

	blog.Infof("found %d projects from project manager, extracted %d unique businesses", len(projectResp.Data.Results),
		len(uniqueBusinessMap))

	// 4. 遍历所有唯一的业务ID，获取CMDB中的集群
	totalOrphanedCount := 0
	for bizIDStr, bizName := range uniqueBusinessMap {
		bizID, err := strconv.ParseInt(bizIDStr, 10, 64)
		if err != nil {
			blog.Errorf("parse business ID %s failed: %v", bizIDStr, err)
			continue
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, constants.BkTenantIdHeaderKey, uniqueBusinessTenantMap[bizIDStr])

		blog.Infof("checking orphaned clusters for business %d (%s)", bizID, bizName)

		cmdbClusters, err := s.Syncer.CMDBClient.GetBcsCluster(ctx, &client.GetBcsClusterRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: bizID,
				Fields:  []string{"id", "uid"},
				Page: client.Page{
					Limit: 200,
					Start: 0,
				},
			},
		}, nil, false)
		if err != nil {
			blog.Errorf("get CMDB clusters for biz %d failed: %v", bizID, err)
			continue
		}

		// 4. 找出CMDB中存在但ClusterManager中不存在的集群
		orphanedCount := 0
		for _, cmdbCluster := range *cmdbClusters {
			ctx, err := tenant.WithTenantIdByResourceForContext(context.Background(), tenant.ResourceMetaData{
				ClusterId: cmdbCluster.Uid,
			})
			if err != nil {
				blog.Infof("Synchronizer sync cluster %s failed: %v", cmdbCluster.Uid, err)
				continue
			}
			if !validClusterMap[cmdbCluster.Uid] {
				blog.Infof("found orphaned cluster in CMDB for biz %d: %s, start deletion", bizID, cmdbCluster.Uid)

				// 设置BizID，因为从CMDB查询的集群数据中可能没有包含此字段
				cmdbCluster.BizID = bizID
				err = s.Syncer.DeleteAllByCluster(ctx, &cmdbCluster)
				if err != nil {
					blog.Errorf("delete orphaned cluster %s in biz %d failed: %v", cmdbCluster.Uid, bizID, err)
				} else {
					blog.Infof("successfully deleted orphaned cluster %s from CMDB in biz %d", cmdbCluster.Uid, bizID)
					orphanedCount++
				}
			}
		}

		if orphanedCount > 0 {
			blog.Infof("cleaned up %d orphaned clusters for business %d (%s)", orphanedCount, bizID, bizName)
		}
		totalOrphanedCount += orphanedCount
	}

	blog.Infof("cleanup orphaned clusters completed, deleted %d clusters total across all businesses", totalOrphanedCount)
	return nil
}
