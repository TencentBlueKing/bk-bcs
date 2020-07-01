/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package taskmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"sync"
	"time"

	jump "github.com/lithammer/go-jump-consistent-hash"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/esb/apigateway/paascc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/discovery"
)

const (
	DEFAULT_CHECK_INTERVAL = 60
)

// Manager task manager for discover task and dispatch task
type Manager struct {
	// paas environment [test, prod, debug, uat]
	paasEnv string
	// cluster environment [prod, stag]
	clusterEnv    string
	paasccClient  paascc.ClientInterface
	zk            *zkclient.ZkClient
	checkInterval int

	// discovery
	disc        *discovery.Client
	servers     []string
	serversLock sync.Mutex

	runFlag    bool
	cancelFunc context.CancelFunc

	curClusters     map[string]*common.Cluster
	newClusters     map[string]*common.Cluster
	clusterChanged  bool
	newClustersLock sync.Mutex
}

// NewManager create new task manager
func NewManager(paasEnv, clusterEnv string,
	checkInterval int,
	disc *discovery.Client, zk *zkclient.ZkClient, paasccClient paascc.ClientInterface) (*Manager, error) {

	interval := checkInterval
	if checkInterval <= 0 {
		interval = DEFAULT_CHECK_INTERVAL
	}

	// ensure /bcs/bcs-services/bkcmdb-sychroniezer/cluster
	isExisted, err := zk.Exist(common.BCS_BKCMDB_SYNC_DIR_CLUSTER)
	if err != nil {
		return nil, fmt.Errorf("[task manager] zk judge path %s failed, err %s", common.BCS_BKCMDB_SYNC_DIR_CLUSTER, err.Error())
	}
	if !isExisted {
		blog.Infof("[task manager] path %s is not existed, try to ensure", common.BCS_BKCMDB_SYNC_DIR_CLUSTER)
		err = zk.CreateDeepNode(common.BCS_BKCMDB_SYNC_DIR_CLUSTER, []byte("bkcmdb-synchronizer-clusters"))
		if err != nil {
			return nil, fmt.Errorf("[task manager] zk create path %s failed, err %s", common.BCS_BKCMDB_SYNC_DIR_CLUSTER, err.Error())
		}
	}
	// ensure /bcs/bcs-services/bkcmdb-sychroniezer/worker
	isExisted, err = zk.Exist(common.BCS_BKCMDB_SYNC_DIR_WORKER)
	if err != nil {
		return nil, fmt.Errorf("[task manager] zk judge path %s failed, err %s", common.BCS_BKCMDB_SYNC_DIR_WORKER, err.Error())
	}
	if !isExisted {
		blog.Infof("[task manager] path %s is not existed, try to ensure", common.BCS_BKCMDB_SYNC_DIR_WORKER)
		err = zk.CreateDeepNode(common.BCS_BKCMDB_SYNC_DIR_WORKER, []byte("bkcmdb-synchronizer-workers"))
		if err != nil {
			return nil, fmt.Errorf("[task manager] zk create path %s failed, err %s", common.BCS_BKCMDB_SYNC_DIR_WORKER, err.Error())
		}
	}

	return &Manager{
		paasccClient:  paasccClient,
		zk:            zk,
		checkInterval: interval,
		paasEnv:       paasEnv,
		clusterEnv:    clusterEnv,
		disc:          disc,
		runFlag:       true,
	}, nil
}

// Run run manager
func (m *Manager) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelFunc = cancel
	m.runFlag = true

	clusterTicker := time.NewTicker(time.Second * time.Duration(m.checkInterval))
	taskTicker := time.NewTicker(5 * time.Second)

	if err := m.syncClusters(); err != nil {
		blog.Warnf("[task manager] sync clusters failed, err %s", err.Error())
	}

	if err := m.syncTasks(); err != nil {
		blog.Warnf("[task manager] sync tasks failed, err %s", err.Error())
	}

	for {
		select {
		case <-clusterTicker.C:
			if err := m.syncClusters(); err != nil {
				blog.Warnf("[task manager] sync clusters failed, err %s", err.Error())
			}
		case <-taskTicker.C:
			if err := m.syncTasks(); err != nil {
				blog.Warnf("[task manager] sync tasks failed, err %s", err.Error())
			}
		case <-ctx.Done():
			blog.Infof("[task manager] manager context done")
			return
		}
	}
}

// Stop manager
func (m *Manager) Stop() {
	m.runFlag = false
	m.cancelFunc()
}

func (m *Manager) syncClusters() error {

	if err := m.getCurrentClusters(); err != nil {
		return err
	}

	if err := m.getNewClusters(); err != nil {
		return err
	}

	var addClusters []*common.Cluster
	for _, cluster := range m.newClusters {
		if _, ok := m.curClusters[cluster.ClusterID]; !ok {
			addClusters = append(addClusters, cluster)
		}
	}

	var delClusters []*common.Cluster
	for _, cluster := range m.curClusters {
		if _, ok := m.newClusters[cluster.ClusterID]; !ok {
			delClusters = append(delClusters, cluster)
		}
	}

	if len(addClusters) != 0 || len(delClusters) != 0 {
		blog.Infof("[task manager] clusters changeed")
		m.clusterChanged = true
	}

	for _, cluster := range addClusters {
		if !m.runFlag {
			return fmt.Errorf("[task manager] not running")
		}
		data, err := json.Marshal(cluster)
		if err != nil {
			blog.Warnf("[task manager] mashal cluster %#v failed, err %s", cluster, err.Error())
			continue
		}
		path := filepath.Join(common.BCS_BKCMDB_SYNC_DIR_CLUSTER, cluster.ClusterID)
		err = m.zk.Create(path, data)
		if err != nil {
			blog.Warnf("[task manager] create zk path %s failed, err %s", path, err.Error())
			continue
		}
		blog.Infof("[task manager] add new cluster %#v", cluster)
	}
	for _, cluster := range delClusters {
		if !m.runFlag {
			return fmt.Errorf("[task manager] not running")
		}
		path := filepath.Join(common.BCS_BKCMDB_SYNC_DIR_CLUSTER, cluster.ClusterID)
		err := m.zk.Del(path, -1)
		if err != nil {
			blog.Warnf("[task manager] delete zk path %s failed, err %s", path, err.Error())
			continue
		}
		blog.Infof("[task manager] delete old cluster %#v", cluster)
	}
	return nil
}

func (m *Manager) getCurrentClusters() error {
	children, err := m.zk.GetChildren(common.BCS_BKCMDB_SYNC_DIR_CLUSTER)
	if err != nil {
		return err
	}

	clueterMap := make(map[string]*common.Cluster)
	for _, cluster := range children {
		if !m.runFlag {
			return fmt.Errorf("[task manager] not running")
		}
		clusterInfo, err := m.zk.Get(filepath.Join(common.BCS_BKCMDB_SYNC_DIR_CLUSTER, cluster))
		if err != nil {
			return err
		}
		tmpCluster := new(common.Cluster)
		err = json.Unmarshal([]byte(clusterInfo), tmpCluster)
		if err != nil {
			return err
		}
		clueterMap[cluster] = tmpCluster
	}
	m.curClusters = clueterMap
	return nil
}

func (m *Manager) getNewClusters() error {

	newClusters := make(map[string]*common.Cluster)
	listProjectRes, err := m.paasccClient.ListProjects(m.paasEnv)
	if err != nil {
		return err
	}
	if len(listProjectRes.Data) == 0 {
		return fmt.Errorf("[task manager] get no project")
	}
	for _, proj := range listProjectRes.Data {
		if proj.CcAppID == 0 {
			continue
		}
		if !m.runFlag {
			return fmt.Errorf("[task manager] not running")
		}
		listClusterRes, err := m.paasccClient.ListProjectClusters(m.paasEnv, proj.ProjectID)
		if err != nil {
			return err
		}
		if listClusterRes.Data.Count == 0 {
			continue
		}
		for _, cluster := range listClusterRes.Data.Results {
			if len(cluster.ClusterID) != 0 && cluster.Environment == m.clusterEnv {
				newClusters[cluster.ClusterID] = &common.Cluster{
					ClusterID: cluster.ClusterID,
					ProjectID: proj.ProjectID,
					BizID:     proj.CcAppID,
				}
			}
		}
	}
	m.newClustersLock.Lock()
	m.newClusters = newClusters
	m.newClustersLock.Unlock()
	return nil
}

func (m *Manager) syncTasks() error {
	servers := m.disc.GetServers()
	if len(servers) == 0 {
		blog.Errorf("[task manager] no worker discovered")
		return fmt.Errorf("[task manager] no worker discovered")
	}

	var serverArr []string
	serverMap := make(map[string]string)
	for _, s := range servers {
		serverArr = append(serverArr, s.IP)
		serverMap[s.IP] = s.IP
	}
	sort.Strings(serverArr)

	if reflect.DeepEqual(m.servers, serverArr) && !m.clusterChanged {
		blog.Infof("servers and clusters no changs")
		return nil
	}
	m.clusterChanged = false
	m.servers = serverArr
	blog.Infof("[task manager] begin to sync tasks")

	var tmpArr [][]*common.Cluster
	for j := 0; j < len(servers); j++ {
		tmpArr = append(tmpArr, make([]*common.Cluster, 0))
	}
	hasher := jump.New(len(servers), jump.NewCRC64())
	m.newClustersLock.Lock()
	for _, cluster := range m.newClusters {
		index := hasher.Hash(cluster.ClusterID)
		tmpArr[index] = append(tmpArr[index], cluster)
	}
	m.newClustersLock.Unlock()

	children, err := m.zk.GetChildren(common.BCS_BKCMDB_SYNC_DIR_WORKER)
	if err != nil {
		blog.Infof("[task manager] get zk path %s children failed, err %s", common.BCS_BKCMDB_SYNC_DIR_WORKER, err.Error())
		return err
	}
	// update data
	for index, clusterArr := range tmpArr {
		serverPath := filepath.Join(common.BCS_BKCMDB_SYNC_DIR_WORKER, serverArr[index])
		isExisted, err := m.zk.Exist(serverPath)
		if err != nil {
			blog.Warnf("[task manager] judge zk path %s exists failed, err %s", serverPath, err.Error())
			continue
		}
		newBytes, err := json.Marshal(clusterArr)
		if err != nil {
			blog.Warnf("[task manager] json marshal cluster array %v failed, err %s", clusterArr, err.Error())
			continue
		}
		blog.Infof("[task manager] sync %s tasks %s", serverArr[index], string(newBytes))
		if isExisted {
			err := m.zk.Update(serverPath, string(newBytes))
			if err != nil {
				blog.Warnf("[task manager] update zk path %s failed, err %s", serverPath, err.Error())
				continue
			}
		} else {
			err := m.zk.Create(serverPath, newBytes)
			if err != nil {
				blog.Warnf("[task manager] create zk path %s failed, err %s", serverPath, err.Error())
				continue
			}
		}
	}
	// clean old data
	for _, child := range children {
		if _, ok := serverMap[child]; !ok {
			blog.Infof("[task manager] clean %s worker", child)
			delPath := filepath.Join(common.BCS_BKCMDB_SYNC_DIR_WORKER, child)
			err := m.zk.Del(delPath, -1)
			if err != nil {
				blog.Warnf("[task manager] delete zk path %s failed, err %s", delPath, err.Error())
			}
		}
	}

	return nil
}
