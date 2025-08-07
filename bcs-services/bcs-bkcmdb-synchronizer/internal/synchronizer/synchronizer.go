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
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
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
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/bkuser"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/cache"
	cm "github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/mq"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/mq/rabbitmq"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/store/db/sqlite"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/syncer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/tenant"
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

	// 集群列表
	cmCli, err := s.getClusterManagerGrpcGwClient()
	if err != nil {
		blog.Errorf("get cluster manager grpc gw client failed, err: %s", err.Error())
		return
	}
	blog.Infof("start sync at %s", time.Now().Format("2006-01-02 15:04:05"))

	lcReq := cmp.ListClusterReq{}
	resp, err := cmCli.Cli.ListCluster(cmCli.Ctx, &lcReq)
	if err != nil {
		blog.Errorf("list cluster failed, err: %s", err.Error())
		return
	}
	clusters := resp.Data

	// 获取需要同步的集群列表
	var (
		clusterMap  = make(map[string]*cmp.Cluster)
		clusterList ClusterList
	)
	s.runCluster(clusters, whiteList, blackList, clusterMap, &clusterList)
	sort.Sort(clusterList)

	blog.Infof("clusterList: %v", clusterList)

	// 有状态服务的副本数
	replicas := s.BkcmdbSynchronizerOption.Synchronizer.Replicas

	// 分而治之: 副本启动时根据副本数平均分配每个副本处理的集群数目, 获取当前副本处理的集群列表
	workList := clusterList[podIndex*len(clusterList)/replicas : (podIndex+1)*len(clusterList)/replicas]
	blog.Infof("podWork handle ClusterList: %v", workList)

	// 启动 goroutineManager，每个集群启动一个 goroutine
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
		for ; true; <-ticker.C {
			respT, errT := cmCli.Cli.ListCluster(cmCli.Ctx, &lcReq)
			if errT != nil {
				blog.Errorf("list cluster failed, err: %s", errT.Error())
				continue
			}

			clustersT := respT.Data

			clusterMapT := make(map[string]*cmp.Cluster)
			var clusterListT ClusterList

			s.runCluster(clustersT, whiteList, blackList, clusterMapT, &clusterListT)

			blog.Infof("clusterListT: %v", clusterListT)

			sort.Sort(clusterListT)

			workListT := clusterListT[podIndex*len(clusterListT)/replicas : (podIndex+1)*len(clusterListT)/replicas]

			for _, wT := range workListT {
				if exist, _ := common.InArray(wT, workList); !exist {
					blog.Infof("%s started", wT)
					gm.Start(wT, clusterMapT[wT])
				}
			}

			for _, wT := range workList {
				if exist, _ := common.InArray(wT, workListT); !exist {
					blog.Infof("%s stopped", wT)
					gm.Stop(wT, clusterMapT[wT])
				}
			}
		}
	}()

	// http server url
	go func() {
		http.HandleFunc("/restart", common.HandleRestart(gm))
		http.HandleFunc("/list", common.HandleList(gm))
		http.HandleFunc("/worklist", common.HandleWorkList(gm, workList))
		http.HandleFunc("/syncStorage", s.syncStorageHandler(clusterMap))
		http.HandleFunc("/syncStore", s.syncStoreHandler(clusterMap))
		http.HandleFunc("/sync", s.syncHandler(clusterList))

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
			for _, w := range workList {

				ctx, errLocal := tenant.WithTenantIdByResourceForContext(context.Background(), tenant.ResourceMetaData{
					ClusterId: w,
				})
				if errLocal != nil {
					blog.Infof("Synchronizer sync cluster %s failed: %v", w, errLocal)
					continue
				}

				bkCluster, err := s.Syncer.GetBkCluster(ctx, clusterMap[w], nil, false)
				if err != nil {
					blog.Errorf("get bk cluster failed, err: %s", err.Error())
					continue
				}
				go s.syncStorage(ctx, clusterMap[w], bkCluster, false)
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
		for _, w := range workList {
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
					gm.Restart(w, clusterMap[w])
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
				gm.Start(w, clusterMap[w])
				blog.Infof("%s restarted", w)
			}
		}
	}
}

func (s *Synchronizer) runCluster(clusters []*cmp.Cluster, whiteList, blackList []string,
	clusterMap map[string]*cmp.Cluster, clusterList *ClusterList) {
	for _, cluster := range clusters {
		blog.Infof("clusterID: %s", cluster.ClusterID)
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

func (s *Synchronizer) syncStorageHandler(clusterMap map[string]*cmp.Cluster) http.HandlerFunc {
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

		bkCluster, err := s.Syncer.GetBkCluster(ctx, clusterMap[clusterId], nil, false)
		if err != nil {
			blog.Errorf("get bk cluster failed, err: %s", err.Error())
			http.Error(w, "get bk cluster failed", http.StatusBadRequest)
			return
		}

		go s.syncStorage(ctx, clusterMap[clusterId], bkCluster, true)
		fmt.Fprintf(w, "BcsClusterID: %s\n syncStorage started.", clusterId)
	}
}

func (s *Synchronizer) syncHandler(clusterList ClusterList) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clusterId := r.URL.Query().Get("cluster")
		if clusterId == "" {
			http.Error(w, "缺少cluster", http.StatusBadRequest)
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
		body, err := ioutil.ReadAll(resp.Body)
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

func (s *Synchronizer) syncStoreHandler(clusterMap map[string]*cmp.Cluster) http.HandlerFunc {
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

		bkCluster, err := s.Syncer.GetBkCluster(ctx, clusterMap[clusterId], nil, false)
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
