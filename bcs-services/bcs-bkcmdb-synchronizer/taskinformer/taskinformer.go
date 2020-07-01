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

package taskinformer

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/common"
)

// TaskHandler handler for task
type TaskHandler interface {
	OnAdd(add common.Cluster)
	OnUpdate(old, new common.Cluster)
	OnDelete(del common.Cluster)
}

// Informer is component which find cluster and informer cluster
type Informer struct {
	// self info
	serverInfo *types.ServerInfo

	// task change handlers
	handlers []TaskHandler

	// zkclient
	zkCli *zkclient.ZkClient

	// cached clusters
	clusters map[string]common.Cluster
}

// NewInformer create new informer
func NewInformer(info *types.ServerInfo, zkCli *zkclient.ZkClient) *Informer {
	return &Informer{
		serverInfo: info,
		zkCli:      zkCli,
		clusters:   make(map[string]common.Cluster),
	}
}

// RegisterHandler register task handler
func (i *Informer) RegisterHandler(h TaskHandler) {
	i.handlers = append(i.handlers, h)
}

func (i *Informer) inform(data []byte) error {
	newClusters := make([]common.Cluster, 0)
	err := json.Unmarshal(data, &newClusters)
	if err != nil {
		blog.Errorf("[task informer] decode task clusters %s failed, err %s", string(data), err.Error())
		return err
	}

	newMap := make(map[string]common.Cluster)
	for _, cluster := range newClusters {
		newMap[cluster.ClusterID] = cluster
		if oldCluster, ok := i.clusters[cluster.ClusterID]; !ok {
			for _, h := range i.handlers {
				h.OnAdd(cluster)
			}
		} else {
			if !reflect.DeepEqual(cluster, oldCluster) {
				for _, h := range i.handlers {
					h.OnUpdate(oldCluster, cluster)
				}
			}
		}
	}

	for _, old := range i.clusters {
		if _, ok := newMap[old.ClusterID]; !ok {
			for _, h := range i.handlers {
				h.OnDelete(old)
			}
		}
	}
	return nil
}

func (i *Informer) watch(ctx context.Context, path string) {
	datas, _, ch, err := i.zkCli.GetW(path)
	if err != nil {
		blog.Warnf("[task informer] watch path %s failed, err %s", path, err.Error())
		time.Sleep(3 * time.Second)
		go i.watch(ctx, path)
		return
	}

	err = i.inform(datas)
	if err != nil {
		blog.Warnf("[task informer] do inform error %s", err.Error())
	}

	for {
		select {
		case zkEvent := <-ch:
			switch zkEvent.Type {
			case zkclient.EventNodeDataChanged:
				datas, err := i.zkCli.Get(path)
				if err != nil {
					blog.Warnf("[task informer] get path %s failed, err %s", path, err.Error())
					continue
				}
				err = i.inform([]byte(datas))
				if err != nil {
					blog.Warnf("[task informer] do inform event data error %s", err.Error())
				}
			case zkclient.EventNodeDeleted:
				blog.Warnf("[task informer] node %s deleted", path)
				time.Sleep(3 * time.Second)
				go i.watch(ctx, path)
				return
			}
		case <-ctx.Done():
			blog.Infof("[task informer] informer context done")
			return
		}
	}

}

// Run run the informer
func (i *Informer) Run(ctx context.Context) {
	path := common.BCS_BKCMDB_SYNC_DIR_WORKER + "/" + i.serverInfo.IP
	go i.watch(ctx, path)
}
