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

// Package handler define methods for handling mq event
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	bkcmdbkube "configcenter/src/kube/types" // nolint
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	cmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage"
	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage/tkex/gamedeployment/v1alpha1"
	gsv1alpha1 "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage/tkex/gamestatefulset/v1alpha1" // nolint
	"github.com/avast/retry-go"
	"github.com/mitchellh/mapstructure"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client"
	cm "github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/clustermanager"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/projectmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/store/db/sqlite"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/syncer"
)

// Deployment 常量表示一种Kubernetes资源类型，用于部署应用程序。
const (
	// Deployment 表示Kubernetes中的部署资源类型
	Deployment = "Deployment"
)

var workloadKindList = []string{"GameDeployment", "GameStatefulSet", "StatefulSet", "DaemonSet", "Deployment"}

// ClusterList the cluster list
type ClusterList []string

// BcsBkcmdbSynchronizerHandler is the handler of bcs-bkcmdb-synchronizer
type BcsBkcmdbSynchronizerHandler struct {
	// BkcmdbSynchronizerOption *option.BkcmdbSynchronizerOption
	Syncer *syncer.Syncer
	// BkCluster *bkcmdbkube.Cluster
	Chn   *amqp.Channel
	CmCli *client.ClusterManagerClientWithHeader
	PmCli *client.ProjectManagerClientWithHeader
}

// Handler is the handler of bcs-bkcmdb-synchronizer
type Handler interface {
	HandleMsg(chn *amqp.Channel, clusterId string, messages <-chan amqp.Delivery, done <-chan bool)
}

// MsgHeader is the message header
type MsgHeader struct {
	ClusterId    string `mapstructure:"clusterId"`
	Event        string `mapstructure:"event"`
	Namespace    string `mapstructure:"namespace"`
	ResourceName string `mapstructure:"resourceName"`
	ResourceType string `mapstructure:"resourceType"`
}

// NewBcsBkcmdbSynchronizerHandler create a new handler
func NewBcsBkcmdbSynchronizerHandler(sync *syncer.Syncer) *BcsBkcmdbSynchronizerHandler {
	optsCm := &cm.Options{
		Module:          cm.ModuleClusterManager,
		Address:         sync.BkcmdbSynchronizerOption.Bcsapi.GrpcAddr,
		EtcdRegistry:    nil,
		ClientTLSConfig: sync.ClientTls,
		AuthToken:       sync.BkcmdbSynchronizerOption.Bcsapi.BearerToken,
	}
	cmCli, _ := cm.NewClusterManagerGrpcGwClient(optsCm)

	optsPm := &pm.Options{
		Module:          pm.ModuleProjectManager,
		Address:         sync.BkcmdbSynchronizerOption.Bcsapi.GrpcAddr,
		EtcdRegistry:    nil,
		ClientTLSConfig: sync.ClientTls,
		AuthToken:       sync.BkcmdbSynchronizerOption.Bcsapi.BearerToken,
	}

	// Create a new project manager gRPC gateway client with the configuration.
	pmCli, _ := pm.NewProjectManagerGrpcGwClient(optsPm)
	return &BcsBkcmdbSynchronizerHandler{
		//BkcmdbSynchronizerOption: option,
		Syncer: sync,
		CmCli:  cmCli,
		PmCli:  pmCli,
	}
}

type msgBuffer struct {
	T time.Time
	M []amqp.Delivery
}

// HandleMsg handle the message from rabbitmq
// nolint funlen
func (b *BcsBkcmdbSynchronizerHandler) HandleMsg(
	chn *amqp.Channel, clusterId string, messages <-chan amqp.Delivery, done <-chan bool) {
	b.Chn = chn

	path := "/data/bcs/bcs-bkcmdb-synchronizer/db/" + clusterId + ".db"

	db := sqlite.Open(path)
	if db == nil {
		blog.Errorf("open db failed, path: %s", path)
		return
	}

	bkCluster, err := b.handleCluster(clusterId, db)
	if err != nil {
		blog.Errorf("handleCluster err: %v", err)
		return
	}

	t := time.Now()

	podMsg := msgBuffer{
		t,
		make([]amqp.Delivery, 0),
	}
	//
	// deployMsg := msgBuffer{
	//	t,
	//	make([]amqp.Delivery, 100),
	// }
	//
	// stsMsg := msgBuffer{
	//	t,
	//	make([]amqp.Delivery, 100),
	// }
	//
	// dsMsg := msgBuffer{
	//	t,
	//	make([]amqp.Delivery, 100),
	// }
	//
	// gDeployMsg := msgBuffer{
	//	t,
	//	make([]amqp.Delivery, 100),
	// }
	//
	// gStsMsg := msgBuffer{
	//	t,
	//	make([]amqp.Delivery, 100),
	// }
	//
	// nsMsg := msgBuffer{
	//	t,
	//	make([]amqp.Delivery, 100),
	// }
	//
	nodeMsg := msgBuffer{
		t,
		make([]amqp.Delivery, 0),
	}

	for msg := range messages {
		select {
		case <-done:
			blog.Infof("goroutine stop, stop handleMsg.")
			return
		default:

		}
		// blog.Infof("Received a message")
		// blog.Infof("Message: %v", msg)

		header := msg.Headers

		if v, ok := header["resourceType"]; ok {
			var errH error
			blog.Infof("resourceType: %v", v)
			switch v.(string) {
			case "Pod":
				m := podMsg.M
				m = append(m, msg)
				podMsg.M = m
				// errH = b.handlePod(msg, bkCluster)
				errH = b.handlePods(&podMsg, bkCluster, db)
			case Deployment:
				errH = b.handleDeployment(msg, bkCluster, db)
			case "StatefulSet":
				errH = b.handleStatefulSet(msg, bkCluster, db)
			case "DaemonSet":
				errH = b.handleDaemonSet(msg, bkCluster, db)
			case "GameDeployment":
				errH = b.handleGameDeployment(msg, bkCluster, db)
			case "GameStatefulSet":
				errH = b.handleGameStatefulSet(msg, bkCluster, db)
			case "Namespace":
				errH = b.handleNamespace(msg, bkCluster, db)
			case "Node":
				// errH = b.handleNode(msg, bkCluster)
				m := nodeMsg.M
				m = append(m, msg)
				nodeMsg.M = m
				errH = b.handleNodes(&nodeMsg, bkCluster, db)
			case "Event":
				errH = b.handlePods(&podMsg, bkCluster, db)
				if errH != nil {
					blog.Errorf("errH: %s", errH.Error())
				}
				errH = b.handleNodes(&nodeMsg, bkCluster, db)
				// errH = b.handleEvent(msg, bkCluster)
			}

			if errH != nil {
				blog.Errorf("errH: %s", errH.Error())
				// if err := b.PublishMsg(msg, 3); err != nil {
				//	blog.Errorf("republish err: %s", err.Error())
				// }
			}
		}

		// ack
		// if err := msg.Ack(true); err != nil {
		//	blog.Infof("Unable to acknowledge the message, err: %s", err.Error())
		// }
	}
}

// handle cluster
// nolint funlen
func (b *BcsBkcmdbSynchronizerHandler) handleCluster(
	clusterId string, db *gorm.DB) (bkCluster *bkcmdbkube.Cluster, err error) {

	lcReq := cmp.ListClusterReq{
		ClusterID: clusterId,
	}

	// list cluster
	resp, err := b.CmCli.Cli.ListCluster(b.CmCli.Ctx, &lcReq)
	if err != nil {
		blog.Errorf("list cluster failed, err: %s", err.Error())
		return nil, err
	}

	clusters := resp.Data
	clusterMap := make(map[string]*cmp.Cluster)
	var clusterList ClusterList

	whiteList := make([]string, 0)
	blackList := make([]string, 0)

	if b.Syncer.BkcmdbSynchronizerOption.Synchronizer.WhiteList != "" {
		whiteList = strings.Split(b.Syncer.BkcmdbSynchronizerOption.Synchronizer.WhiteList, ",")
	}

	if b.Syncer.BkcmdbSynchronizerOption.Synchronizer.BlackList != "" {
		blackList = strings.Split(b.Syncer.BkcmdbSynchronizerOption.Synchronizer.BlackList, ",")
	}

	// white list
	blog.Infof("whiteList: %v, len: %d", whiteList, len(whiteList))
	blog.Infof("blackList: %v, len: %d", blackList, len(blackList))

	// loop clusters
	for _, cluster := range clusters {
		blog.Infof("1cluster: %s", cluster.ClusterID)
		if len(whiteList) > 0 {
			if exit, _ := common.InArray(cluster.ClusterID, whiteList); !exit {
				continue
			}
			blog.Infof("2cluster: %s", cluster.ClusterID)
		}

		if len(blackList) > 0 {
			if exit, _ := common.InArray(cluster.ClusterID, blackList); exit {
				continue
			}
		}

		blog.Infof("3cluster: %s", cluster.ClusterID)

		// virtual cluster
		if cluster.ClusterType == "virtual" {
			continue
		}
		blog.Infof("4cluster: %s", cluster.ClusterID)
		if _, ok := clusterMap[cluster.ClusterID]; ok {
			if cluster.IsShared {
				clusterMap[cluster.ClusterID] = cluster
			}
		} else {
			clusterMap[cluster.ClusterID] = cluster
			clusterList = append(clusterList, cluster.ClusterID)
			blog.Infof("5cluster: %s", cluster.ClusterID)
		}

	}

	// get bk cluster
	bkCluster, err = b.Syncer.GetBkCluster(clusterMap[clusterId], db, true)
	if err != nil {
		blog.Errorf("handleCluster: Unable to get bkcluster, err: %s", err.Error())
		return nil, err
	}

	return bkCluster, err
}

// handle pod
func (b *BcsBkcmdbSynchronizerHandler) handlePod(msg amqp.Delivery, bkCluster *bkcmdbkube.Cluster) error { // nolint
	blog.Infof("handlePod Message: %v", msg.Headers)
	msgHeader, err := getMsgHeader(&msg.Headers)
	if err != nil {
		blog.Errorf("handlePod unable to get headers, err: %s", err.Error())
		return fmt.Errorf("handlePod unable to get headers, err: %s", err.Error())
	}

	blog.Infof("Headers: %s", msgHeader.ClusterId)
	pod := &corev1.Pod{}
	err = json.Unmarshal(msg.Body, pod)
	if err != nil {
		blog.Errorf("handlePod: Unable to unmarshal")
		return fmt.Errorf("handlePod: Unable to unmarshal")
	}

	switch msgHeader.Event {
	case "update": // nolint
		err = b.handlePodUpdate(pod, bkCluster)
		if err != nil {
			blog.Errorf("handlePodUpdate err: %s", err.Error())
			return fmt.Errorf("handlePodUpdate err: %s", err.Error())
		}
	case "delete": // nolint
		err = b.handlePodDelete(pod, bkCluster)
		if err != nil {
			blog.Errorf("handlePodDelete err: %s", err.Error())
			return fmt.Errorf("handlePodDelete err: %s", err.Error())
		}
	default:
		blog.Errorf("handlePod: Unknown event: %s", msgHeader.Event)
	}
	return nil
}

// handle pod
func (b *BcsBkcmdbSynchronizerHandler) handlePods(podMsg *msgBuffer, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	// blog.Infof("handlePod Message: %v", msg.Headers)
	// msgHeader, err := getMsgHeader(&msg.Headers)
	// if err != nil {
	//	blog.Errorf("handlePod unable to get headers, err: %s", err.Error())
	//	return fmt.Errorf("handlePod unable to get headers, err: %s", err.Error())
	// }
	//
	// blog.Infof("Headers: %s", msgHeader.ClusterId)
	// pod := &corev1.Pod{}
	// err = json.Unmarshal(msg.Body, pod)
	// if err != nil {
	//	blog.Errorf("handlePod: Unable to unmarshal")
	//	return fmt.Errorf("handlePod: Unable to unmarshal")
	// }
	blog.Infof("podMsg: %d", len(podMsg.M))
	if time.Since(podMsg.T) < 10*time.Second {
		// blog.Infof("podMsg.T: %s, %s", podMsg.T, time.Now().Sub(podMsg.T))
		if len(podMsg.M) < 100 {
			return nil
		}
	}

	podsUpdate := make(map[string]*corev1.Pod)
	podsDelete := make(map[string]*corev1.Pod)

	for _, msg := range podMsg.M {
		blog.Infof("handlePod Message: %v", msg.Headers)
		msgHeader, err := getMsgHeader(&msg.Headers)
		if err != nil {
			blog.Errorf("handlePod unable to get headers, err: %s", err.Error())
			return fmt.Errorf("handlePod unable to get headers, err: %s", err.Error())
		}
		blog.Infof("Headers: %s", msgHeader.ClusterId)

		pod := &corev1.Pod{}
		err = json.Unmarshal(msg.Body, pod)
		if err != nil {
			blog.Errorf("handlePod: Unable to unmarshal")
			return fmt.Errorf("handlePod: Unable to unmarshal")
		}
		switch msgHeader.Event {
		case "update":
			podsUpdate[string(pod.UID)] = pod
		case "delete":
			podsDelete[string(pod.UID)] = pod
			blog.Infof("podToDelete: %s+%s+%s", msgHeader.ClusterId, pod.Namespace, pod.Name)
		default:
			blog.Errorf("handlePod: Unknown event: %s", msgHeader.Event)
		}
	}

	err := b.handlePodsUpdate(podsUpdate, bkCluster, db)
	if err != nil {
		blog.Errorf("handlePodsUpdate err: %s", err.Error())
		// return fmt.Errorf("handlePodsUpdate err: %s", err.Error())
	}

	err = b.handlePodsDelete(podsDelete, bkCluster, db)
	if err != nil {
		blog.Errorf("handlePodsDelete err: %s", err.Error())
		// return fmt.Errorf("handlePodsDelete err: %s", err.Error())
	}

	podMsg.M = make([]amqp.Delivery, 0)
	podMsg.T = time.Now()

	return nil
}

// handle pod update
// nolint funlen
func (b *BcsBkcmdbSynchronizerHandler) handlePodUpdate(pod *corev1.Pod, bkCluster *bkcmdbkube.Cluster) error {
	bkPods, err := b.Syncer.GetBkPods(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{pod.Name},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
			{
				Field:    "namespace",
				Operator: "in",
				Value:    []string{pod.Namespace},
			},
		},
	}, false, nil)
	if err != nil {
		return err
	}

	if len(*bkPods) > 1 {
		return fmt.Errorf("len(bkPods) = %d", len(*bkPods))
	}

	storageCli, err := b.Syncer.GetBcsStorageClient()
	if err != nil {
		return err
	}
	pods, err := storageCli.QueryK8SPod(bkCluster.Uid, pod.Namespace, pod.Name)
	if err != nil {
		blog.Errorf("query k8s pod failed, err: %s", err.Error())
		return err
	}
	pod = pods[0].Data

	// handle pod create
	if len(*bkPods) == 0 {
		err = b.handlePodCreate(pod, bkCluster)
		if err != nil {
			blog.Errorf("handlePodCreate failed for pod %s: %v", pod.Name, err)
			return err
		}
	}

	if len(*bkPods) == 1 {
		if pod.Status.Phase != corev1.PodRunning {
			err = b.handlePodDelete(pod, bkCluster)
			if err != nil {
				blog.Errorf("handlePodDelete err: %s", err.Error())
				return fmt.Errorf("handlePodDelete err: %s", err.Error())
			}
		}

		// bkContainers, err := b.Syncer.CMDBClient.GetBcsContainer(&client.GetBcsContainerRequest{
		//	CommonRequest: client.CommonRequest{
		//		BKBizID: (*bkPods)[0].BizID,
		//		Page: client.Page{
		//			Limit: 200,
		//			Start: 0,
		//		},
		//	},
		//	BkPodID: (*bkPods)[0].ID,
		// }, nil, false)

		bkContainers, err := b.Syncer.GetBkContainers((*bkPods)[0].BizID, &client.PropertyFilter{
			Condition: "AND",
			Rules: []client.Rule{
				{
					Field:    "bk_pod_id",
					Operator: "in",
					Value:    []int64{(*bkPods)[0].ID},
				},
			},
		}, false, nil)

		if err != nil {
			blog.Errorf("handlePodUpdate GetBcsContainer err: %v", err)
			return fmt.Errorf("handlePodUpdate GetBcsContainer err: %v", err)
		}
		for i, c := range *bkContainers {
			if *c.ContainerID != pod.Status.ContainerStatuses[i].ContainerID {
				err = b.handlePodDelete(pod, bkCluster)
				if err != nil {
					blog.Errorf("handlePodDelete err: %s", err.Error())
					return fmt.Errorf("handlePodDelete err: %s", err.Error())
				}

				err := b.handlePodCreate(pod, bkCluster)
				if err != nil {
					blog.Errorf("handlePodCreate failed for pod %s: %v", pod.Name, err)
					return err
				}
				break
			}
		}
	}

	return nil
}

// nolint funlen
func (b *BcsBkcmdbSynchronizerHandler) handlePodsUpdate(
	podsUpdate map[string]*corev1.Pod, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	if len(podsUpdate) == 0 {
		return nil
	}
	nsPod := make(map[string][]string)
	for _, v := range podsUpdate {
		nsPods := nsPod[v.Namespace]
		nsPods = append(nsPods, v.Name)
		nsPod[v.Namespace] = nsPods
	}

	bkPodsMap := make(map[string]bkcmdbkube.Pod)

	for ns, pNames := range nsPod {
		bkPods, err := b.Syncer.GetBkPods(bkCluster.BizID, &client.PropertyFilter{
			Condition: "AND",
			Rules: []client.Rule{
				{
					Field:    "name",
					Operator: "in",
					Value:    pNames,
				},
				{
					Field:    "cluster_uid",
					Operator: "in",
					Value:    []string{bkCluster.Uid},
				},
				{
					Field:    "namespace",
					Operator: "in",
					Value:    []string{ns},
				},
			},
		}, true, db)
		if err != nil {
			blog.Errorf("GetBkPods error: %v", err)
			continue
		}
		for _, bkPod := range *bkPods {
			bkPodsMap[bkPod.NameSpace+*bkPod.Name] = bkPod
		}
	}

	storageCli, err := b.Syncer.GetBcsStorageClient()
	if err != nil {
		return err
	}

	k8sPods := make([]storage.Pod, 0)

	for k, v := range nsPod {
		for _, p := range v {
			pods, errP := storageCli.QueryK8SPod(bkCluster.Uid, k, p)

			if errP != nil {
				blog.Errorf("query k8s pod %s %s %s failed, err: %s", bkCluster.Uid, k, p, errP.Error())
				continue
			}
			k8sPods = append(k8sPods, *pods[0])
		}
	}

	k8sPodsMap := make(map[string]storage.Pod)

	for _, pod := range k8sPods {
		k8sPodsMap[pod.Data.Namespace+pod.Data.Name] = pod
	}

	podsDelete := make(map[string]*corev1.Pod)
	podsCreate := make(map[string]*corev1.Pod)

	for k, k8sPod := range k8sPodsMap {
		if bkPod, exist := bkPodsMap[k]; exist {
			if k8sPod.Data.Status.Phase != corev1.PodRunning {
				podsDelete[string(k8sPod.Data.UID)] = k8sPod.Data
				blog.Infof("podToDelete: %s+%s+%s", bkCluster.Uid, k8sPod.Data.Namespace, k8sPod.Data.Name)
				continue
			}

			blog.Infof("bkpod: %v", bkPod)
			blog.Infof("bkPod.BizID: %d, bkPod.ID: %d", bkPod.BizID, bkPod.ID)

			// bkContainers, err := b.Syncer.CMDBClient.GetBcsContainer(&client.GetBcsContainerRequest{
			//	CommonRequest: client.CommonRequest{
			//		BKBizID: bkPod.BizID,
			//		Page: client.Page{
			//			Limit: 200,
			//			Start: 0,
			//		},
			//	},
			//	BkPodID: bkPod.ID,
			// }, nil, false)

			bkContainers, errC := b.Syncer.GetBkContainers(bkPod.BizID, &client.PropertyFilter{
				Condition: "AND",
				Rules: []client.Rule{
					{
						Field:    "bk_pod_id",
						Operator: "in",
						Value:    []int64{bkPod.ID},
					},
				},
			}, true, db)

			if errC != nil {
				blog.Errorf("handlePodUpdate GetBcsContainer err: %v", errC)
				continue
			}

			bkContainerIds := make([]string, 0)

			for _, c := range *bkContainers {
				bkContainerIds = append(bkContainerIds, *c.ContainerID)
			}

			for _, cs := range k8sPod.Data.Status.ContainerStatuses {
				if ok, _ := common.InArray(cs.ContainerID, bkContainerIds); !ok {
					blog.Infof("pod: %s needs to recreate.", k8sPod.Data.Name)
					podsDelete[string(k8sPod.Data.UID)] = k8sPod.Data
					blog.Infof("podToDelete: %s+%s+%s", bkCluster.Uid, k8sPod.Data.Namespace, k8sPod.Data.Name)
					podsCreate[string(k8sPod.Data.UID)] = k8sPod.Data
					break
				}
			}
			blog.Infof("bkContainerIds: %s, ContainerStatuses: %v",
				bkContainerIds, k8sPod.Data.Status.ContainerStatuses)

		} else {
			podsCreate[string(k8sPod.Data.UID)] = k8sPod.Data
		}
	}

	bkPodIDs := make([]int64, 0)

	for k, bkPod := range bkPodsMap {
		if _, exist := k8sPodsMap[k]; !exist {
			bkPodIDs = append(bkPodIDs, bkPod.ID)
			blog.Infof("podToDelete: %s+%s", bkPod.NameSpace, *bkPod.Name)
		}
	}

	err = retry.Do(
		func() error {
			return b.Syncer.DeleteBkPods(bkCluster, &bkPodIDs, db)
		},
		retry.Delay(time.Second*2),
		retry.Attempts(3),
		retry.DelayType(retry.FixedDelay),
	)

	if err != nil {
		blog.Errorf("handlePodsDelete err: %s", err.Error())
	}

	err = b.handlePodsDelete(podsDelete, bkCluster, db)
	if err != nil {
		blog.Errorf("handlePodsDelete err: %s", err.Error())
		// return fmt.Errorf("handlePodsDelete err: %s", err.Error())
	}

	err = b.handlePodsCreate(podsCreate, bkCluster, db)
	if err != nil {
		blog.Errorf("handlePodsCreate err: %s", err.Error())
		// return fmt.Errorf("handlePodsDelete err: %s", err.Error())
	}

	return err
}

// handle pod delete
func (b *BcsBkcmdbSynchronizerHandler) handlePodDelete(pod *corev1.Pod, bkCluster *bkcmdbkube.Cluster) error { // nolint
	bkPods, err := b.Syncer.GetBkPods(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{pod.Name},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
			{
				Field:    "namespace",
				Operator: "in",
				Value:    []string{pod.Namespace},
			},
		},
	}, false, nil)
	if err != nil {
		return err
	}

	if len(*bkPods) > 1 {
		return fmt.Errorf("len(bkPods) = %d", len(*bkPods))
	}

	if len(*bkPods) == 0 {
		return fmt.Errorf("pod %s not found", pod.Name)
	}

	bkPod := (*bkPods)[0]

	// b.Syncer.DeleteBkPods(b.BkCluster.BizID, &[]int64{bkPod.ID})
	err = retry.Do(
		func() error {
			return b.Syncer.DeleteBkPods(bkCluster, &[]int64{bkPod.ID}, nil)
		},
		retry.Delay(time.Second*2),
		retry.Attempts(3),
		retry.DelayType(retry.FixedDelay),
	)

	return err
}

func (b *BcsBkcmdbSynchronizerHandler) handlePodsDelete(
	podsDelete map[string]*corev1.Pod, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	if len(podsDelete) == 0 {
		return nil
	}

	nsPod := make(map[string][]string)
	for _, v := range podsDelete {
		nsPods := nsPod[v.Namespace]
		nsPods = append(nsPods, v.Name)
		nsPod[v.Namespace] = nsPods
	}

	blog.Infof("handlePodsDelete podNames: %v", nsPod)

	bkPodIDs := make([]int64, 0)

	for ns, pNames := range nsPod {
		bkPods, err := b.Syncer.GetBkPods(bkCluster.BizID, &client.PropertyFilter{
			Condition: "AND",
			Rules: []client.Rule{
				{
					Field:    "name",
					Operator: "in",
					Value:    pNames,
				},
				{
					Field:    "cluster_uid",
					Operator: "in",
					Value:    []string{bkCluster.Uid},
				},
				{
					Field:    "namespace",
					Operator: "in",
					Value:    []string{ns},
				},
			},
		}, true, db)
		if err != nil {
			blog.Errorf("GetBkPods error: %v", err)
			continue
		}
		if len(*bkPods) == 0 {
			blog.Errorf("pods %s not found", pNames)
			continue
		}
		for _, bkPod := range *bkPods {
			bkPodIDs = append(bkPodIDs, bkPod.ID)
		}
	}

	// b.Syncer.DeleteBkPods(b.BkCluster.BizID, &[]int64{bkPod.ID})
	err := retry.Do(
		func() error {
			return b.Syncer.DeleteBkPods(bkCluster, &bkPodIDs, db)
		},
		retry.Delay(time.Second*2),
		retry.Attempts(3),
		retry.DelayType(retry.FixedDelay),
	)

	return err
}

// handle pod create
// nolint funlen
func (b *BcsBkcmdbSynchronizerHandler) handlePodCreate(pod *corev1.Pod, bkCluster *bkcmdbkube.Cluster) error {
	var operator []string
	lcReq := cmp.ListClusterReq{
		ClusterID: bkCluster.Uid,
	}

	resp, err := b.CmCli.Cli.ListCluster(b.CmCli.Ctx, &lcReq)
	if err != nil {
		blog.Errorf("list cluster failed, err: %s", err.Error())
		return err
	}

	clusters := resp.Data

	bkNamespaces, err := b.Syncer.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{pod.Namespace},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, false, nil)
	if err != nil {
		return err
	}

	if len(*bkNamespaces) != 1 {
		return errors.New(fmt.Sprintf("len(bkNamespaces) = %d", len(*bkNamespaces)))
	}

	bkNamespace := (*bkNamespaces)[0]

	bkWorkloadPods, err := b.Syncer.GetBkWorkloads(bkCluster.BizID, "pods", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "namespace",
				Operator: "in",
				Value:    []string{pod.Namespace},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, false, nil)
	if err != nil {
		blog.Errorf("get bk workload pods failed, err: %s", err.Error())
		return err
	}

	if len(*bkWorkloadPods) != 1 {
		blog.Errorf("get bk workload pods len is %d", len(*bkWorkloadPods))
		return errors.New(fmt.Sprintf("len(*bkWorkloadPods) = %d", len(*bkWorkloadPods)))
	}

	p := bkcmdbkube.PodsWorkload{}
	err = common.InterfaceToStruct((*bkWorkloadPods)[0], &p)
	if err != nil {
		blog.Errorf("convert bk workload pods failed, err: %s", err.Error())
		return err
	}

	workloadKind := "pods"
	workloadName := "pods"
	workloadID := p.ID

	if len(pod.OwnerReferences) == 1 {
		ownerRef := pod.OwnerReferences[0]
		if ownerRef.Kind == "ReplicaSet" {
			storageCli, err := b.Syncer.GetBcsStorageClient()
			if err != nil {
				return err
			}
			rsList, err := storageCli.QueryK8sReplicaSet(bkCluster.Uid, pod.Namespace, ownerRef.Name)
			if err != nil {
				return errors.New(fmt.Sprintf("query replicaSet %s failed, err: %s",
					ownerRef.Name, err.Error()))
			}
			if len(rsList) != 1 {
				for _, rs := range rsList {
					blog.Infof("rs: %v", rs.Data)
				}
				return errors.New(fmt.Sprintf("replicaSet %s not found", ownerRef.Name))
			}
			rs := rsList[0]

			if len(rs.Data.OwnerReferences) == 0 {
				return errors.New("no owner references")
			}
			rsOwnerRef := rs.Data.OwnerReferences[0]
			switch rsOwnerRef.Kind {
			case Deployment:
				workloadKind = "deployment"
				workloadName = rsOwnerRef.Name
				bkWorkloads, err := b.Syncer.GetBkWorkloads(bkCluster.BizID, workloadKind, &client.PropertyFilter{
					Condition: "AND",
					Rules: []client.Rule{
						{
							Field:    "cluster_uid",
							Operator: "in",
							Value:    []string{bkCluster.Uid},
						},
						{
							Field:    "namespace",
							Operator: "in",
							Value:    []string{bkNamespace.Name},
						},
						{
							Field:    "name",
							Operator: "in",
							Value:    []string{workloadName},
						},
					},
				}, false, nil)

				if err != nil {
					return err
				}

				if len(*bkWorkloads) == 0 {
					return errors.New(fmt.Sprintf("no workload %s in %s", workloadName, bkNamespace.Name))
				}

				if len(*bkWorkloads) > 1 {
					return errors.New(fmt.Sprintf("len(bkWorkloads) = %d", len(*bkWorkloads)))
				}

				workloadID = (int64)((*bkWorkloads)[0].(map[string]interface{})["id"].(float64))
				if labels := (*bkWorkloads)[0].(map[string]interface{})["labels"]; labels != nil {
					if creator, creatorOk :=
						labels.(map[string]interface{})["io.tencent.paas.creator"]; creatorOk && (creator != "") {
						operator = append(operator, creator.(string))
					} else if creator, creatorOk =
						labels.(map[string]interface{})["io．tencent．paas．creator"]; creatorOk && (creator != "") {
						operator = append(operator, creator.(string))
					} else if updater, updaterOk :=
						labels.(map[string]interface{})["io.tencent.paas.updater"]; updaterOk && (updater != "") {
						operator = append(operator, updater.(string))
					} else if updater, updaterOk =
						labels.(map[string]interface{})["io．tencent．paas．updator"]; updaterOk && (updater != "") {
						operator = append(operator, updater.(string))
					}
				}
			default:
				return errors.New(fmt.Sprintf("kind %s is not supported", rsOwnerRef.Kind))
			}

		} else if exist, _ := common.InArray(ownerRef.Kind, workloadKindList); exist {
			workloadKind = common.FirstLower(ownerRef.Kind)
			workloadName = ownerRef.Name
			bkWorkloads, err := b.Syncer.GetBkWorkloads(bkCluster.BizID, workloadKind, &client.PropertyFilter{
				Condition: "AND",
				Rules: []client.Rule{
					{
						Field:    "cluster_uid",
						Operator: "in",
						Value:    []string{bkCluster.Uid},
					},
					{
						Field:    "namespace",
						Operator: "in",
						Value:    []string{bkNamespace.Name},
					},
					{
						Field:    "name",
						Operator: "in",
						Value:    []string{workloadName},
					},
				},
			}, false, nil)

			if err != nil {
				return err
			}

			if len(*bkWorkloads) == 0 {
				return errors.New(fmt.Sprintf("no workload %s in %s", workloadName, bkNamespace.Name))
			}

			if len(*bkWorkloads) > 1 {
				return errors.New(fmt.Sprintf("len(bkWorkloads) = %d", len(*bkWorkloads)))
			}

			workloadID = (int64)((*bkWorkloads)[0].(map[string]interface{})["id"].(float64))
			if labels := (*bkWorkloads)[0].(map[string]interface{})["labels"]; labels != nil {
				if creator, creatorOk :=
					labels.(map[string]interface{})["io.tencent.paas.creator"]; creatorOk && (creator != "") {
					operator = append(operator, creator.(string))
				} else if creator, creatorOk =
					labels.(map[string]interface{})["io．tencent．paas．creator"]; creatorOk && (creator != "") {
					operator = append(operator, creator.(string))
				} else if updater, updaterOk :=
					labels.(map[string]interface{})["io.tencent.paas.updater"]; updaterOk && (updater != "") {
					operator = append(operator, updater.(string))
				} else if updater, updaterOk =
					labels.(map[string]interface{})["io．tencent．paas．updator"]; updaterOk && (updater != "") {
					operator = append(operator, updater.(string))
				}
			}
		} else {
			return errors.New(fmt.Sprintf("kind %s is not supported", ownerRef.Kind))

		}
	}

	var nodeID, hostID int64

	bkNodes, err := b.Syncer.GetBkNodes(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{pod.Spec.NodeName},
			},
		},
	}, false, nil)

	if err != nil {
		return err
	}

	if len(*bkNodes) != 1 {
		return errors.New(fmt.Sprintf("len(bkNodes) = %d", len(*bkNodes)))
	}

	bkNode := (*bkNodes)[0]

	nodeID = bkNode.ID
	hostID = bkNode.HostID

	podIPs := make([]bkcmdbkube.PodIP, 0)
	for _, ip := range pod.Status.PodIPs {
		podIPs = append(podIPs, bkcmdbkube.PodIP{
			IP: ip.IP,
		})
	}

	containerStatusMap := make(map[string]corev1.ContainerStatus)

	for _, containerStatus := range pod.Status.ContainerStatuses {
		containerStatusMap[containerStatus.Name] = containerStatus
	}

	containers := make([]bkcmdbkube.ContainerBaseFields, 0)
	for _, container := range pod.Spec.Containers {

		ports := make([]bkcmdbkube.ContainerPort, 0)

		for _, port := range container.Ports {
			ports = append(ports, bkcmdbkube.ContainerPort{
				Name:          port.Name,
				HostPort:      port.HostPort,
				ContainerPort: port.ContainerPort,
				Protocol:      bkcmdbkube.Protocol(port.Protocol),
				HostIP:        port.HostIP,
			})
		}

		env := make([]bkcmdbkube.EnvVar, 0)

		for _, envVar := range container.Env {
			env = append(env, bkcmdbkube.EnvVar{
				Name:  envVar.Name,
				Value: envVar.Value,
			})
		}

		mounts := make([]bkcmdbkube.VolumeMount, 0)

		for _, mount := range container.VolumeMounts {
			mounts = append(mounts, bkcmdbkube.VolumeMount{
				Name:        mount.Name,
				MountPath:   mount.MountPath,
				SubPath:     mount.SubPath,
				ReadOnly:    mount.ReadOnly,
				SubPathExpr: mount.SubPathExpr,
			})
		}

		containerID := containerStatusMap[container.Name].ContainerID

		if containerID == "" {
			return errors.New("container not found")
		}

		cName := container.Name
		cImage := container.Image
		cArgs := container.Args

		containers = append(containers, bkcmdbkube.ContainerBaseFields{
			Name:        &cName,
			Image:       &cImage,
			ContainerID: &containerID,
			Ports:       &ports,
			Args:        &cArgs,
			Environment: &env,
			Mounts:      &mounts,
		})
	}

	if len(operator) == 0 && (bkNamespace.Labels != nil) {
		if creator, creatorOk := (*bkNamespace.Labels)["io.tencent.paas.creator"]; creatorOk && (creator != "") {
			operator = append(operator, creator)
		} else if creator, creatorOk =
			(*bkNamespace.Labels)["io．tencent．paas．creator"]; creatorOk && (creator != "") {
			operator = append(operator, creator)
		} else if updater, updaterOk :=
			(*bkNamespace.Labels)["io.tencent.paas.updater"]; updaterOk && (updater != "") {
			operator = append(operator, updater)
		} else if updater, updaterOk =
			(*bkNamespace.Labels)["io．tencent．paas．updator"]; updaterOk && (updater != "") {
			operator = append(operator, updater)
		}
	}

	if len(operator) == 0 {
		if clusters[0].Creator != "" {
			operator = append(operator, clusters[0].Creator)
		} else if clusters[0].Updater != "" {
			operator = append(operator, clusters[0].Updater)
		}
	}

	if len(operator) == 0 {
		operator = append(operator, "")
	}

	b.Syncer.CreateBkPods(bkCluster, map[int64][]client.CreateBcsPodRequestDataPod{
		bkNamespace.BizID: []client.CreateBcsPodRequestDataPod{
			{
				Spec: &client.CreateBcsPodRequestPodSpec{
					ClusterID:    &bkCluster.ID,
					NameSpaceID:  &bkNamespace.ID,
					WorkloadKind: &workloadKind,
					WorkloadID:   &workloadID,
					NodeID:       &nodeID,
					Ref: &bkcmdbkube.Reference{
						Kind: bkcmdbkube.WorkloadType(workloadKind),
						Name: workloadName,
						ID:   workloadID,
					},
				},

				Name:       &pod.Name,
				HostID:     &hostID,
				Priority:   pod.Spec.Priority,
				Labels:     &pod.Labels,
				IP:         &pod.Status.PodIP,
				IPs:        &podIPs,
				Containers: &containers,
				Operator:   &operator,
			},
		},
	}, nil)

	blog.Infof("podToAdd: %s+%s+%s", bkCluster.Uid, &pod.Namespace, &pod.Name)

	return nil
}

// nolint funlen
func (b *BcsBkcmdbSynchronizerHandler) handlePodsCreate(podsCreate map[string]*corev1.Pod, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	if len(podsCreate) == 0 {
		return nil
	}

	var podNames []string
	for _, v := range podsCreate {
		podNames = append(podNames, v.Name)
	}
	blog.Infof("handlePodsCreate podNames: %v", podNames)

	lcReq := cmp.ListClusterReq{
		ClusterID: bkCluster.Uid,
	}

	resp, err := b.CmCli.Cli.ListCluster(b.CmCli.Ctx, &lcReq)
	if err != nil {
		blog.Errorf("list cluster failed, err: %s", err.Error())
		return err
	}

	clusters := resp.Data

	for _, pod := range podsCreate {
		var operator []string
		bkNamespaces, err := b.Syncer.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
			Condition: "AND",
			Rules: []client.Rule{
				{
					Field:    "name",
					Operator: "in",
					Value:    []string{pod.Namespace},
				},
				{
					Field:    "cluster_uid",
					Operator: "in",
					Value:    []string{bkCluster.Uid},
				},
			},
		}, true, db)
		if err != nil {
			blog.Errorf("GetBkNamespaces error: %v", err)
			continue
		}

		if len(*bkNamespaces) != 1 {
			blog.Errorf("len(bkNamespaces) = %d", len(*bkNamespaces))
			continue
		}

		bkNamespace := (*bkNamespaces)[0]

		bkWorkloadPods, err := b.Syncer.GetBkWorkloads(bkCluster.BizID, "pods", &client.PropertyFilter{
			Condition: "AND",
			Rules: []client.Rule{
				{
					Field:    "namespace",
					Operator: "in",
					Value:    []string{pod.Namespace},
				},
				{
					Field:    "cluster_uid",
					Operator: "in",
					Value:    []string{bkCluster.Uid},
				},
			},
		}, true, db)
		if err != nil {
			blog.Errorf("get bk workload pods failed, err: %s", err.Error())
			continue
		}

		if len(*bkWorkloadPods) != 1 {
			blog.Errorf("get bk workload pods len is %d", len(*bkWorkloadPods))
			continue
		}

		p := bkcmdbkube.PodsWorkload{}
		err = common.InterfaceToStruct((*bkWorkloadPods)[0], &p)
		if err != nil {
			blog.Errorf("convert bk workload pods failed, err: %s", err.Error())
			continue
		}

		workloadKind := "pods"
		workloadName := "pods"
		workloadID := p.ID

		if len(pod.OwnerReferences) == 1 {
			ownerRef := pod.OwnerReferences[0]
			if ownerRef.Kind == "ReplicaSet" {
				storageCli, errS := b.Syncer.GetBcsStorageClient()
				if errS != nil {
					continue
				}
				rsList, errRS := storageCli.QueryK8sReplicaSet(bkCluster.Uid, pod.Namespace, ownerRef.Name)
				if errRS != nil {
					blog.Errorf("query replicaSet %s failed, err: %s", ownerRef.Name, errRS.Error())
					continue
				}
				if len(rsList) != 1 {
					for _, rs := range rsList {
						blog.Infof("rs: %v", rs.Data)
					}
					blog.Errorf("replicaSet %s not found", ownerRef.Name)
					continue
				}
				rs := rsList[0]

				if len(rs.Data.OwnerReferences) == 0 {
					blog.Errorf("no owner references")
					continue
				}
				rsOwnerRef := rs.Data.OwnerReferences[0]
				switch rsOwnerRef.Kind {
				case Deployment:
					workloadKind = "deployment"
					workloadName = rsOwnerRef.Name
					bkWorkloads, errW := b.Syncer.GetBkWorkloads(bkCluster.BizID, workloadKind, &client.PropertyFilter{
						Condition: "AND",
						Rules: []client.Rule{
							{
								Field:    "cluster_uid",
								Operator: "in",
								Value:    []string{bkCluster.Uid},
							},
							{
								Field:    "namespace",
								Operator: "in",
								Value:    []string{bkNamespace.Name},
							},
							{
								Field:    "name",
								Operator: "in",
								Value:    []string{workloadName},
							},
						},
					}, true, db)

					if errW != nil {
						continue
					}

					if len(*bkWorkloads) == 0 {
						blog.Errorf("no workload %s in %s", workloadName, bkNamespace.Name)
						continue
					}

					if len(*bkWorkloads) > 1 {
						blog.Errorf("len(bkWorkloads) = %d", len(*bkWorkloads))
						continue
					}

					workloadID = (int64)((*bkWorkloads)[0].(map[string]interface{})["id"].(float64))
					if labels := (*bkWorkloads)[0].(map[string]interface{})["labels"]; labels != nil {
						if creator, creatorOk :=
							labels.(map[string]interface{})["io.tencent.paas.creator"]; creatorOk && (creator != "") {
							operator = append(operator, creator.(string))
						} else if creator, creatorOk =
							labels.(map[string]interface{})["io．tencent．paas．creator"]; creatorOk && (creator != "") {
							operator = append(operator, creator.(string))
						} else if updater, updaterOk :=
							labels.(map[string]interface{})["io.tencent.paas.updater"]; updaterOk && (updater != "") {
							operator = append(operator, updater.(string))
						} else if updater, updaterOk =
							labels.(map[string]interface{})["io．tencent．paas．updator"]; updaterOk && (updater != "") {
							operator = append(operator, updater.(string))
						}
					}
				default:
					blog.Errorf("kind %s is not supported", rsOwnerRef.Kind)
					continue
				}

			} else if exist, _ := common.InArray(ownerRef.Kind, workloadKindList); exist {
				workloadKind = common.FirstLower(ownerRef.Kind)
				workloadName = ownerRef.Name
				bkWorkloads, errW := b.Syncer.GetBkWorkloads(bkCluster.BizID, workloadKind, &client.PropertyFilter{
					Condition: "AND",
					Rules: []client.Rule{
						{
							Field:    "cluster_uid",
							Operator: "in",
							Value:    []string{bkCluster.Uid},
						},
						{
							Field:    "namespace",
							Operator: "in",
							Value:    []string{bkNamespace.Name},
						},
						{
							Field:    "name",
							Operator: "in",
							Value:    []string{workloadName},
						},
					},
				}, true, db)

				if errW != nil {
					continue
				}

				if len(*bkWorkloads) == 0 {
					blog.Errorf("no workload %s in %s", workloadName, bkNamespace.Name)
					continue
				}

				if len(*bkWorkloads) > 1 {
					blog.Errorf("len(bkWorkloads) = %d", len(*bkWorkloads))
					continue
				}

				workloadID = (int64)((*bkWorkloads)[0].(map[string]interface{})["id"].(float64))
				if labels := (*bkWorkloads)[0].(map[string]interface{})["labels"]; labels != nil {
					if creator, creatorOk :=
						labels.(map[string]interface{})["io.tencent.paas.creator"]; creatorOk && (creator != "") {
						operator = append(operator, creator.(string))
					} else if creator, creatorOk =
						labels.(map[string]interface{})["io．tencent．paas．creator"]; creatorOk && (creator != "") {
						operator = append(operator, creator.(string))
					} else if updater, updaterOk :=
						labels.(map[string]interface{})["io.tencent.paas.updater"]; updaterOk && (updater != "") {
						operator = append(operator, updater.(string))
					} else if updater, updaterOk =
						labels.(map[string]interface{})["io．tencent．paas．updator"]; updaterOk && (updater != "") {
						operator = append(operator, updater.(string))
					}
				}
			} else {
				blog.Errorf("kind %s is not supported", ownerRef.Kind)
				continue
			}
		}

		var nodeID, hostID int64

		bkNodes, err := b.Syncer.GetBkNodes(bkCluster.BizID, &client.PropertyFilter{
			Condition: "AND",
			Rules: []client.Rule{
				{
					Field:    "cluster_uid",
					Operator: "in",
					Value:    []string{bkCluster.Uid},
				},
				{
					Field:    "name",
					Operator: "in",
					Value:    []string{pod.Spec.NodeName},
				},
			},
		}, true, db)

		if err != nil {
			continue
		}

		if len(*bkNodes) != 1 {
			blog.Errorf("len(bkNodes) = %d", len(*bkNodes))
			continue
		}

		bkNode := (*bkNodes)[0]

		nodeID = bkNode.ID
		hostID = bkNode.HostID

		podIPs := make([]bkcmdbkube.PodIP, 0)
		for _, ip := range pod.Status.PodIPs {
			podIPs = append(podIPs, bkcmdbkube.PodIP{
				IP: ip.IP,
			})
		}

		containerStatusMap := make(map[string]corev1.ContainerStatus)

		for _, containerStatus := range pod.Status.ContainerStatuses {
			containerStatusMap[containerStatus.Name] = containerStatus
		}

		containers := make([]bkcmdbkube.ContainerBaseFields, 0)
		for _, container := range pod.Spec.Containers {

			ports := make([]bkcmdbkube.ContainerPort, 0)

			for _, port := range container.Ports {
				ports = append(ports, bkcmdbkube.ContainerPort{
					Name:          port.Name,
					HostPort:      port.HostPort,
					ContainerPort: port.ContainerPort,
					Protocol:      bkcmdbkube.Protocol(port.Protocol),
					HostIP:        port.HostIP,
				})
			}

			env := make([]bkcmdbkube.EnvVar, 0)

			for _, envVar := range container.Env {
				env = append(env, bkcmdbkube.EnvVar{
					Name:  envVar.Name,
					Value: envVar.Value,
				})
			}

			mounts := make([]bkcmdbkube.VolumeMount, 0)

			for _, mount := range container.VolumeMounts {
				mounts = append(mounts, bkcmdbkube.VolumeMount{
					Name:        mount.Name,
					MountPath:   mount.MountPath,
					SubPath:     mount.SubPath,
					ReadOnly:    mount.ReadOnly,
					SubPathExpr: mount.SubPathExpr,
				})
			}

			containerID := containerStatusMap[container.Name].ContainerID

			if containerID == "" {
				blog.Errorf("container not found")
				continue
			}

			cName := container.Name
			cImage := container.Image
			cArgs := container.Args

			containers = append(containers, bkcmdbkube.ContainerBaseFields{
				Name:        &cName,
				Image:       &cImage,
				ContainerID: &containerID,
				Ports:       &ports,
				Args:        &cArgs,
				Environment: &env,
				Mounts:      &mounts,
			})
		}

		if len(operator) == 0 && (bkNamespace.Labels != nil) {
			if creator, creatorOk :=
				(*bkNamespace.Labels)["io.tencent.paas.creator"]; creatorOk && (creator != "") {
				operator = append(operator, creator)
			} else if creator, creatorOk =
				(*bkNamespace.Labels)["io．tencent．paas．creator"]; creatorOk && (creator != "") {
				operator = append(operator, creator)
			} else if updater, updaterOk :=
				(*bkNamespace.Labels)["io.tencent.paas.updater"]; updaterOk && (updater != "") {
				operator = append(operator, updater)
			} else if updater, updaterOk =
				(*bkNamespace.Labels)["io．tencent．paas．updator"]; updaterOk && (updater != "") {
				operator = append(operator, updater)
			}
		}

		if len(operator) == 0 {
			if clusters[0].Creator != "" {
				operator = append(operator, clusters[0].Creator)
			} else if clusters[0].Updater != "" {
				operator = append(operator, clusters[0].Updater)
			}
		}

		if len(operator) == 0 {
			operator = append(operator, "")
		}

		b.Syncer.CreateBkPods(bkCluster, map[int64][]client.CreateBcsPodRequestDataPod{
			bkNamespace.BizID: {
				{
					Spec: &client.CreateBcsPodRequestPodSpec{
						ClusterID:    &bkCluster.ID,
						NameSpaceID:  &bkNamespace.ID,
						WorkloadKind: &workloadKind,
						WorkloadID:   &workloadID,
						NodeID:       &nodeID,
						Ref: &bkcmdbkube.Reference{
							Kind: bkcmdbkube.WorkloadType(workloadKind),
							Name: workloadName,
							ID:   workloadID,
						},
					},

					Name:       &pod.Name,
					HostID:     &hostID,
					Priority:   pod.Spec.Priority,
					Labels:     &pod.Labels,
					IP:         &pod.Status.PodIP,
					IPs:        &podIPs,
					Containers: &containers,
					Operator:   &operator,
				},
			},
		}, db)
		blog.Infof("podToAdd: %s+%s+%s", bkCluster.Uid, pod.Namespace, pod.Name)
	}

	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleDeployment(
	msg amqp.Delivery, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	blog.Infof("handleDeployment Message: %v", msg.Headers)
	msgHeader, err := getMsgHeader(&msg.Headers)
	if err != nil {
		blog.Errorf("handleDeployment unable to get headers, err: %s", err.Error())
		return fmt.Errorf("handleDeployment unable to get headers, err: %s", err.Error())
	}

	blog.Infof("Headers: %s", msgHeader.ClusterId)
	deployment := &appv1.Deployment{}
	err = json.Unmarshal(msg.Body, deployment)
	if err != nil {
		blog.Errorf("handleDeployment: Unable to unmarshal")
		return fmt.Errorf("handleDeployment: Unable to unmarshal")
	}

	switch msgHeader.Event {
	case "update": // nolint
		err = b.handleDeploymentUpdate(deployment, bkCluster, db)
		if err != nil {
			blog.Errorf("handleDeploymentUpdate err: %s", err.Error())
			return fmt.Errorf("handleDeploymentUpdate err: %s", err.Error())
		}
	case "delete": // nolint
		err = b.handleDeploymentDelete(deployment, bkCluster, db)
		if err != nil {
			blog.Errorf("handleDeploymentDelete err: %s", err.Error())
			return fmt.Errorf("handleDeploymentDelete err: %s", err.Error())
		}
	default:
		blog.Errorf("handleDeployment: Unknown event: %s", msgHeader.Event)
	}
	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleDeploymentUpdate(
	deployment *appv1.Deployment, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	bkDeployments, err := b.Syncer.GetBkWorkloads(bkCluster.BizID, "deployment", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{deployment.Name},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
			{
				Field:    "namespace",
				Operator: "in",
				Value:    []string{deployment.Namespace},
			},
		},
	}, true, db)

	if err != nil {
		return err
	}

	if len(*bkDeployments) == 0 {
		err := b.handleDeploymentCreate(deployment, bkCluster, db)
		if err != nil {
			blog.Errorf(fmt.Sprintf("handleDeploymentCreate err: %s", err.Error()))
			return err
		}
	}

	if len(*bkDeployments) == 1 {
		bd := (*bkDeployments)[0]
		bkDeployment := bkcmdbkube.Deployment{}
		err := common.InterfaceToStruct(bd, &bkDeployment)
		if err != nil {
			blog.Errorf("convert bk deployment failed, err: %s", err.Error())
			return err
		}

		deploymentToUpdate := make(map[int64]*client.UpdateBcsWorkloadRequestData, 0)
		needToUpdate, updateData := b.Syncer.CompareDeployment(&bkDeployment, &storage.Deployment{Data: deployment})
		if needToUpdate {
			deploymentToUpdate[bkDeployment.ID] = updateData
			blog.Infof("deploymentToUpdate: %s+%s+%s", bkCluster.Uid, bkDeployment.Namespace, bkDeployment.Name)
			b.Syncer.UpdateBkWorkloads(bkCluster, "deployment", &deploymentToUpdate, db)
		}
	}

	if len(*bkDeployments) > 1 {
		blog.Errorf("handleDeploymentUpdate: More than one deployment found")
		return fmt.Errorf("handleDeploymentUpdate: More than one deployment found")
	}

	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleDeploymentDelete(
	deployment *appv1.Deployment, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	bkDeployments, err := b.Syncer.GetBkWorkloads(bkCluster.BizID, "deployment", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{deployment.Name},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
			{
				Field:    "namespace",
				Operator: "in",
				Value:    []string{deployment.Namespace},
			},
		},
	}, true, db)

	if err != nil {
		return err
	}

	if len(*bkDeployments) > 1 {
		return fmt.Errorf("len(bkDeployments) = %d", len(*bkDeployments))
	}

	if len(*bkDeployments) == 0 {
		return fmt.Errorf("deployment %s not found", deployment.Name)
	}

	bd := (*bkDeployments)[0]
	bkDeployment := bkcmdbkube.Deployment{}
	err = common.InterfaceToStruct(bd, &bkDeployment)
	if err != nil {
		blog.Errorf("convert bk deployment failed, err: %s", err.Error())
		return err
	}

	blog.Infof("deploymentToDelete: %s+%s+%s", bkCluster.Uid, bkDeployment.Namespace, bkDeployment.Name)

	// err = b.Syncer.DeleteBkWorkloads(b.BkCluster.BizID, "deployment", &[]int64{bkDeployment.ID})
	err = retry.Do(
		func() error {
			return b.Syncer.DeleteBkWorkloads(bkCluster, "deployment", &[]int64{bkDeployment.ID}, db)
		},
		retry.Delay(time.Second*1),
		retry.Attempts(2),
		retry.DelayType(retry.FixedDelay),
	)

	return err
}

func (b *BcsBkcmdbSynchronizerHandler) handleDeploymentCreate(
	deployment *appv1.Deployment, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	bkNamespaces, err := b.Syncer.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{deployment.Namespace},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)
	if err != nil {
		return err
	}

	if len(*bkNamespaces) != 1 {
		return fmt.Errorf("len(bkNamespaces) = %d", len(*bkNamespaces))
	}

	bkNamespace := (*bkNamespaces)[0]

	deploymentToAdd := make(map[int64][]client.CreateBcsWorkloadRequestData, 0)
	toAddData := b.Syncer.GenerateBkDeployment(&bkNamespace, &storage.Deployment{Data: deployment})
	deploymentToAdd[bkNamespace.BizID] = []client.CreateBcsWorkloadRequestData{*toAddData}
	blog.Infof("deploymentToAdd: %s+%s+%s", bkCluster.Uid, deployment.Namespace, deployment.Name)

	b.Syncer.CreateBkWorkloads(bkCluster, "deployment", deploymentToAdd, db)
	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleStatefulSet(
	msg amqp.Delivery, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	blog.Infof("handleStatefulSet Message: %v", msg.Headers)
	msgHeader, err := getMsgHeader(&msg.Headers)
	if err != nil {
		blog.Errorf("handleStatefulSet unable to get headers, err: %s", err.Error())
		return fmt.Errorf("handleStatefulSet unable to get headers, err: %s", err.Error())
	}

	blog.Infof("Headers: %s", msgHeader.ClusterId)
	statefulSet := &appv1.StatefulSet{}
	err = json.Unmarshal(msg.Body, statefulSet)
	if err != nil {
		blog.Errorf("handleStatefulSet: Unable to unmarshal")
		return fmt.Errorf("handleStatefulSet: Unable to unmarshal")
	}

	switch msgHeader.Event {
	case "update": // nolint
		err = b.handleStatefulSetUpdate(statefulSet, bkCluster, db)
		if err != nil {
			blog.Errorf("handleStatefulSetUpdate err: %s", err.Error())
			return fmt.Errorf("handleStatefulSetUpdate err: %s", err.Error())
		}
	case "delete": // nolint
		err = b.handleStatefulSetDelete(statefulSet, bkCluster, db)
		if err != nil {
			blog.Errorf("handleStatefulSetDelete err: %s", err.Error())
			return fmt.Errorf("handleStatefulSetDelete err: %s", err.Error())
		}
	default:
		blog.Errorf("handleStatefulSet: Unknown event: %s", msgHeader.Event)
	}
	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleStatefulSetUpdate(
	statefulSet *appv1.StatefulSet, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	bkStatefulSets, err := b.Syncer.GetBkWorkloads(bkCluster.BizID, "statefulSet", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{statefulSet.Name},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
			{
				Field:    "namespace",
				Operator: "in",
				Value:    []string{statefulSet.Namespace},
			},
		},
	}, true, db)

	if err != nil {
		return err
	}

	if len(*bkStatefulSets) == 0 {
		err := b.handleStatefulSetCreate(statefulSet, bkCluster, db)
		if err != nil {
			blog.Errorf(fmt.Sprintf("handleStatefulSetCreate err: %s", err.Error()))
			return err
		}
	}

	if len(*bkStatefulSets) == 1 {
		bs := (*bkStatefulSets)[0]
		bkStatefulSet := bkcmdbkube.StatefulSet{}
		err := common.InterfaceToStruct(bs, &bkStatefulSet)
		if err != nil {
			blog.Errorf("convert bk statefulSet failed, err: %s", err.Error())
			return err
		}

		statefulSetToUpdate := make(map[int64]*client.UpdateBcsWorkloadRequestData, 0)
		needToUpdate, updateData := b.Syncer.CompareStatefulSet(&bkStatefulSet, &storage.StatefulSet{Data: statefulSet})
		if needToUpdate {
			statefulSetToUpdate[bkStatefulSet.ID] = updateData
			blog.Infof("statefulSetToUpdate: %s+%s+%s",
				bkCluster.Uid, bkStatefulSet.Namespace, bkStatefulSet.Name)
			b.Syncer.UpdateBkWorkloads(bkCluster, "statefulSet", &statefulSetToUpdate, db)
		}
	}

	if len(*bkStatefulSets) > 1 {
		blog.Errorf("handleStatefulSetUpdate: More than one statefulSet found")
		return fmt.Errorf("handleStatefulSetUpdate: More than one statefulSet found")
	}

	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleStatefulSetDelete(
	statefulSet *appv1.StatefulSet, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	bkStatefulSets, err := b.Syncer.GetBkWorkloads(bkCluster.BizID, "statefulSet", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{statefulSet.Name},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
			{
				Field:    "namespace",
				Operator: "in",
				Value:    []string{statefulSet.Namespace},
			},
		},
	}, true, db)

	if err != nil {
		return err
	}

	if len(*bkStatefulSets) > 1 {
		return fmt.Errorf("len(bkStatefulSets) = %d", len(*bkStatefulSets))
	}

	if len(*bkStatefulSets) == 0 {
		return fmt.Errorf("statefulSet %s not found", statefulSet.Name)
	}

	bs := (*bkStatefulSets)[0]
	bkStatefulSet := bkcmdbkube.StatefulSet{}
	err = common.InterfaceToStruct(bs, &bkStatefulSet)
	if err != nil {
		blog.Errorf("convert bk statefulSet failed, err: %s", err.Error())
		return err
	}

	blog.Infof("statefulSetToDelete: %s+%s+%s", bkCluster.Uid, bkStatefulSet.Namespace, bkStatefulSet.Name)

	// err = b.Syncer.DeleteBkWorkloads(b.BkCluster.BizID, "statefulSet", &[]int64{bkStatefulSet.ID})
	err = retry.Do(
		func() error {
			return b.Syncer.DeleteBkWorkloads(bkCluster, "statefulSet", &[]int64{bkStatefulSet.ID}, db)
		},
		retry.Delay(time.Second*1),
		retry.Attempts(2),
		retry.DelayType(retry.FixedDelay),
	)

	return err
}

func (b *BcsBkcmdbSynchronizerHandler) handleStatefulSetCreate(
	statefulSet *appv1.StatefulSet, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	bkNamespaces, err := b.Syncer.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{statefulSet.Namespace},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)
	if err != nil {
		return err
	}

	if len(*bkNamespaces) != 1 {
		return fmt.Errorf("len(bkNamespaces) = %d", len(*bkNamespaces))
	}

	bkNamespace := (*bkNamespaces)[0]

	statefulSetToAdd := make(map[int64][]client.CreateBcsWorkloadRequestData, 0)
	toAddData := b.Syncer.GenerateBkStatefulSet(&bkNamespace, &storage.StatefulSet{Data: statefulSet})
	statefulSetToAdd[bkNamespace.BizID] = []client.CreateBcsWorkloadRequestData{*toAddData}
	blog.Infof("statefulSetToAdd: %s+%s+%s", bkCluster.Uid, statefulSet.Namespace, statefulSet.Name)

	b.Syncer.CreateBkWorkloads(bkCluster, "statefulSet", statefulSetToAdd, db)
	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleDaemonSet(
	msg amqp.Delivery, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	blog.Infof("handleDaemonSet Message: %v", msg.Headers)
	msgHeader, err := getMsgHeader(&msg.Headers)
	if err != nil {
		blog.Errorf("handleDaemonSet unable to get headers, err: %s", err.Error())
		return fmt.Errorf("handleDaemonSet unable to get headers, err: %s", err.Error())
	}

	blog.Infof("Headers: %s", msgHeader.ClusterId)
	daemonSet := &appv1.DaemonSet{}
	err = json.Unmarshal(msg.Body, daemonSet)
	if err != nil {
		blog.Errorf("handleDaemonSet: Unable to unmarshal")
		return fmt.Errorf("handleDaemonSet: Unable to unmarshal")
	}

	switch msgHeader.Event {
	case "update": // nolint
		err = b.handleDaemonSetUpdate(daemonSet, bkCluster, db)
		if err != nil {
			blog.Errorf("handleDaemonSetUpdate err: %s", err.Error())
			return fmt.Errorf("handleDaemonSetUpdate err: %s", err.Error())
		}
	case "delete": // nolint
		err = b.handleDaemonSetDelete(daemonSet, bkCluster, db)
		if err != nil {
			blog.Errorf("handleDaemonSetDelete err: %s", err.Error())
			return fmt.Errorf("handleDaemonSetDelete err: %s", err.Error())
		}
	default:
		blog.Errorf("handleDaemonSet: Unknown event: %s", msgHeader.Event)
	}
	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleDaemonSetUpdate(
	daemonSet *appv1.DaemonSet, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	bkDaemonSets, err := b.Syncer.GetBkWorkloads(bkCluster.BizID, "daemonSet", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{daemonSet.Name},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
			{
				Field:    "namespace",
				Operator: "in",
				Value:    []string{daemonSet.Namespace},
			},
		},
	}, true, db)

	if err != nil {
		return err
	}

	if len(*bkDaemonSets) == 0 {
		err := b.handleDaemonSetCreate(daemonSet, bkCluster, db)
		if err != nil {
			blog.Errorf(fmt.Sprintf("handleDaemonSetCreate err: %s", err.Error()))
			return err
		}
	}

	if len(*bkDaemonSets) == 1 {
		bd := (*bkDaemonSets)[0]
		bkDaemonSet := bkcmdbkube.DaemonSet{}
		err := common.InterfaceToStruct(bd, &bkDaemonSet)
		if err != nil {
			blog.Errorf("convert bk daemonSet failed, err: %s", err.Error())
			return err
		}

		daemonSetToUpdate := make(map[int64]*client.UpdateBcsWorkloadRequestData, 0)
		needToUpdate, updateData := b.Syncer.CompareDaemonSet(&bkDaemonSet, &storage.DaemonSet{Data: daemonSet})
		if needToUpdate {
			daemonSetToUpdate[bkDaemonSet.ID] = updateData
			blog.Infof("daemonSetToUpdate: %s+%s+%s", bkCluster.Uid, daemonSet.Namespace, daemonSet.Name)
			b.Syncer.UpdateBkWorkloads(bkCluster, "daemonSet", &daemonSetToUpdate, db)
		}
	}

	if len(*bkDaemonSets) > 1 {
		blog.Errorf("handleDaemonSetUpdate: More than one daemonSet found")
		return fmt.Errorf("handleDaemonSetUpdate: More than one daemonSet found")
	}

	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleDaemonSetDelete(
	daemonSet *appv1.DaemonSet, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	bkDaemonSets, err := b.Syncer.GetBkWorkloads(bkCluster.BizID, "daemonSet", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{daemonSet.Name},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
			{
				Field:    "namespace",
				Operator: "in",
				Value:    []string{daemonSet.Namespace},
			},
		},
	}, true, db)

	if err != nil {
		return err
	}

	if len(*bkDaemonSets) > 1 {
		return fmt.Errorf("len(bkDaemonSets) = %d", len(*bkDaemonSets))
	}

	if len(*bkDaemonSets) == 0 {
		return fmt.Errorf("daemonSet %s not found", daemonSet.Name)
	}

	bd := (*bkDaemonSets)[0]
	bkDaemonSet := bkcmdbkube.DaemonSet{}
	err = common.InterfaceToStruct(bd, &bkDaemonSet)
	if err != nil {
		blog.Errorf("convert bk daemonSet failed, err: %s", err.Error())
		return err
	}

	blog.Infof("daemonSetToDelete: %s+%s+%s", bkCluster.Uid, bkDaemonSet.Namespace, bkDaemonSet.Name)

	// err = b.Syncer.DeleteBkWorkloads(b.BkCluster.BizID, "daemonSet", &[]int64{bkDaemonSet.ID})
	err = retry.Do(
		func() error {
			return b.Syncer.DeleteBkWorkloads(bkCluster, "daemonSet", &[]int64{bkDaemonSet.ID}, db)
		},
		retry.Delay(time.Second*1),
		retry.Attempts(2),
		retry.DelayType(retry.FixedDelay),
	)

	return err
}

func (b *BcsBkcmdbSynchronizerHandler) handleDaemonSetCreate(
	daemonSet *appv1.DaemonSet, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	bkNamespaces, err := b.Syncer.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{daemonSet.Namespace},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)
	if err != nil {
		return err
	}

	if len(*bkNamespaces) != 1 {
		return fmt.Errorf("len(bkNamespaces) = %d", len(*bkNamespaces))
	}

	bkNamespace := (*bkNamespaces)[0]

	daemonSetToAdd := make(map[int64][]client.CreateBcsWorkloadRequestData, 0)
	toAddData := b.Syncer.GenerateBkDaemonSet(&bkNamespace, &storage.DaemonSet{Data: daemonSet})
	daemonSetToAdd[bkNamespace.BizID] = []client.CreateBcsWorkloadRequestData{*toAddData}
	blog.Infof("daemonSetToAdd: %s+%s+%s", bkCluster.Uid, daemonSet.Namespace, daemonSet.Name)

	b.Syncer.CreateBkWorkloads(bkCluster, "daemonSet", daemonSetToAdd, db)
	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleGameDeployment(msg amqp.Delivery,
	bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	blog.Infof("handleGameDeployment Message: %v", msg.Headers)
	msgHeader, err := getMsgHeader(&msg.Headers)
	if err != nil {
		blog.Errorf("handleGameDeployment unable to get headers, err: %s", err.Error())
		return fmt.Errorf("handleGameDeployment unable to get headers, err: %s", err.Error())
	}

	blog.Infof("Headers: %s", msgHeader.ClusterId)
	gameDeployment := &gdv1alpha1.GameDeployment{}
	err = json.Unmarshal(msg.Body, gameDeployment)
	if err != nil {
		blog.Errorf("handleGameDeployment: Unable to unmarshal")
		return fmt.Errorf("handleGameDeployment: Unable to unmarshal")
	}

	switch msgHeader.Event {
	case "update": // nolint
		err = b.handleGameDeploymentUpdate(gameDeployment, bkCluster, db)
		if err != nil {
			blog.Errorf("handleGameDeploymentUpdate err: %s", err.Error())
			return fmt.Errorf("handleGameDeploymentUpdate err: %s", err.Error())
		}
	case "delete": // nolint
		err = b.handleGameDeploymentDelete(gameDeployment, bkCluster, db)
		if err != nil {
			blog.Errorf("handleGameDeploymentDelete err: %s", err.Error())
			return fmt.Errorf("handleGameDeploymentDelete err: %s", err.Error())
		}
	default:
		blog.Errorf("handleGameDeployment: Unknown event: %s", msgHeader.Event)
	}
	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleGameDeploymentUpdate(
	gameDeployment *gdv1alpha1.GameDeployment, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	bkGameDeployments, err := b.Syncer.GetBkWorkloads(bkCluster.BizID, "gameDeployment", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{gameDeployment.Name},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
			{
				Field:    "namespace",
				Operator: "in",
				Value:    []string{gameDeployment.Namespace},
			},
		},
	}, true, db)

	if err != nil {
		return err
	}

	if len(*bkGameDeployments) == 0 {
		err := b.handleGameDeploymentCreate(gameDeployment, bkCluster, db)
		if err != nil {
			blog.Errorf(fmt.Sprintf("handleGameDeploymentCreate err: %s", err.Error()))
			return err
		}
	}

	if len(*bkGameDeployments) == 1 {
		bgd := (*bkGameDeployments)[0]
		bkGameDeployment := bkcmdbkube.GameDeployment{}
		err := common.InterfaceToStruct(bgd, &bkGameDeployment)
		if err != nil {
			blog.Errorf("convert bk gameDeployment failed, err: %s", err.Error())
			return err
		}

		gameDeploymentToUpdate := make(map[int64]*client.UpdateBcsWorkloadRequestData, 0)
		needToUpdate, updateData := b.Syncer.CompareGameDeployment(
			&bkGameDeployment, &storage.GameDeployment{Data: gameDeployment})
		if needToUpdate {
			gameDeploymentToUpdate[bkGameDeployment.ID] = updateData
			blog.Infof("gameDeploymentToUpdate: %s+%s+%s",
				bkCluster.Uid, gameDeployment.Namespace, gameDeployment.Name)
			b.Syncer.UpdateBkWorkloads(bkCluster, "gameDeployment", &gameDeploymentToUpdate, db)
		}
	}

	if len(*bkGameDeployments) > 1 {
		blog.Errorf("handleDaemonSetUpdate: More than one daemonSet found")
		return fmt.Errorf("handleDaemonSetUpdate: More than one daemonSet found")
	}

	return nil
}

// handle GameDeployment Delete
func (b *BcsBkcmdbSynchronizerHandler) handleGameDeploymentDelete(
	gameDeployment *gdv1alpha1.GameDeployment, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	bkGameDeployments, err :=
		b.Syncer.GetBkWorkloads(bkCluster.BizID, "gameDeployment", &client.PropertyFilter{
			Condition: "AND",
			Rules: []client.Rule{
				{
					Field:    "name",
					Operator: "in",
					Value:    []string{gameDeployment.Name},
				},
				{
					Field:    "cluster_uid",
					Operator: "in",
					Value:    []string{bkCluster.Uid},
				},
				{
					Field:    "namespace",
					Operator: "in",
					Value:    []string{gameDeployment.Namespace},
				},
			},
		}, true, db)

	if err != nil {
		return err
	}

	if len(*bkGameDeployments) > 1 {
		return fmt.Errorf("len(bkGameDeployments) = %d", len(*bkGameDeployments))
	}

	if len(*bkGameDeployments) == 0 {
		return fmt.Errorf("gameDeployment %s not found", gameDeployment.Name)
	}

	bgd := (*bkGameDeployments)[0]
	bkGameDeployment := bkcmdbkube.GameDeployment{}
	err = common.InterfaceToStruct(bgd, &bkGameDeployment)
	if err != nil {
		blog.Errorf("convert bk gameDeployment failed, err: %s", err.Error())
		return err
	}

	blog.Infof("gameDeploymentToDelete: %s+%s+%s",
		bkCluster.Uid, bkGameDeployment.Namespace, bkGameDeployment.Name)

	// err = b.Syncer.DeleteBkWorkloads(b.BkCluster.BizID, "gameDeployment", &[]int64{bkGameDeployment.ID})
	err = retry.Do(
		func() error {
			return b.Syncer.DeleteBkWorkloads(bkCluster, "gameDeployment", &[]int64{bkGameDeployment.ID}, db)
		},
		retry.Delay(time.Second*1),
		retry.Attempts(2),
		retry.DelayType(retry.FixedDelay),
	)

	return err
}

// handle GameDeployment Create
func (b *BcsBkcmdbSynchronizerHandler) handleGameDeploymentCreate(
	gameDeployment *gdv1alpha1.GameDeployment, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	bkNamespaces, err := b.Syncer.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{gameDeployment.Namespace},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)
	if err != nil {
		return err
	}

	if len(*bkNamespaces) != 1 {
		return fmt.Errorf("len(bkNamespaces) = %d", len(*bkNamespaces))
	}

	bkNamespace := (*bkNamespaces)[0]

	gameDeploymentToAdd := make(map[int64][]client.CreateBcsWorkloadRequestData, 0)
	toAddData := b.Syncer.GenerateBkGameDeployment(&bkNamespace, &storage.GameDeployment{Data: gameDeployment})
	gameDeploymentToAdd[bkNamespace.BizID] = []client.CreateBcsWorkloadRequestData{*toAddData}
	blog.Infof("gameDeploymentToAdd: %s+%s+%s", bkCluster.Uid, gameDeployment.Namespace, gameDeployment.Name)

	b.Syncer.CreateBkWorkloads(bkCluster, "gameDeployment", gameDeploymentToAdd, db)
	return nil
}

// handle GameStateful Set
func (b *BcsBkcmdbSynchronizerHandler) handleGameStatefulSet(
	msg amqp.Delivery, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	blog.Infof("handleGameStatefulSet Message: %v", msg.Headers)
	msgHeader, err := getMsgHeader(&msg.Headers)
	if err != nil {
		blog.Errorf("handleGameStatefulSet unable to get headers, err: %s", err.Error())
		return fmt.Errorf("handleGameStatefulSet unable to get headers, err: %s", err.Error())
	}

	blog.Infof("Headers: %s", msgHeader.ClusterId)
	gameStatefulSet := &gsv1alpha1.GameStatefulSet{}
	err = json.Unmarshal(msg.Body, gameStatefulSet)
	if err != nil {
		blog.Errorf("handleGameStatefulSet: Unable to unmarshal")
		return fmt.Errorf("handleGameStatefulSet: Unable to unmarshal")
	}

	switch msgHeader.Event {
	case "update": // nolint
		err = b.handleGameStatefulSetUpdate(gameStatefulSet, bkCluster, db)
		if err != nil {
			blog.Errorf("handleGameStatefulSetUpdate err: %s", err.Error())
			return fmt.Errorf("handleGameStatefulSetUpdate err: %s", err.Error())
		}
	case "delete": // nolint
		err = b.handleGameStatefulSetDelete(gameStatefulSet, bkCluster, db)
		if err != nil {
			blog.Errorf("handleGameStatefulSetDelete err: %s", err.Error())
			return fmt.Errorf("handleGameStatefulSetDelete err: %s", err.Error())
		}
	default:
		blog.Errorf("handleGameStatefulSet: Unknown event: %s", msgHeader.Event)
	}
	return nil
}

// handle GameStateful Set Update
func (b *BcsBkcmdbSynchronizerHandler) handleGameStatefulSetUpdate(
	gameStatefulSet *gsv1alpha1.GameStatefulSet, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	// GetBkWorkloads get bkworkloads
	bkGameStatefulSets, err :=
		b.Syncer.GetBkWorkloads(bkCluster.BizID, "gameStatefulSet", &client.PropertyFilter{
			Condition: "AND",
			Rules: []client.Rule{
				{
					Field:    "name",
					Operator: "in",
					Value:    []string{gameStatefulSet.Name},
				},
				{
					Field:    "cluster_uid",
					Operator: "in",
					Value:    []string{bkCluster.Uid},
				},
				{
					Field:    "namespace",
					Operator: "in",
					Value:    []string{gameStatefulSet.Namespace},
				},
			},
		}, true, db)

	if err != nil {
		return err
	}

	if len(*bkGameStatefulSets) == 0 {
		err := b.handleGameStatefulSetCreate(gameStatefulSet, bkCluster, db)
		if err != nil {
			blog.Errorf(fmt.Sprintf("handleGameStatefulSetCreate err: %s", err.Error()))
			return err
		}
	}

	if len(*bkGameStatefulSets) == 1 {
		bgs := (*bkGameStatefulSets)[0]
		bkGameStatefulSet := bkcmdbkube.GameStatefulSet{}
		err := common.InterfaceToStruct(bgs, &bkGameStatefulSet)
		if err != nil {
			blog.Errorf("convert bk gameStatefulSet failed, err: %s", err.Error())
			return err
		}

		gameStatefulSetToUpdate := make(map[int64]*client.UpdateBcsWorkloadRequestData, 0)
		needToUpdate, updateData := b.Syncer.CompareGameStatefulSet(&bkGameStatefulSet,
			&storage.GameStatefulSet{Data: gameStatefulSet})
		if needToUpdate {
			gameStatefulSetToUpdate[bkGameStatefulSet.ID] = updateData
			blog.Infof("gameStatefulSetToUpdate: %s+%s+%s",
				bkCluster.Uid, gameStatefulSet.Namespace, gameStatefulSet.Name)
			b.Syncer.UpdateBkWorkloads(bkCluster, "gameStatefulSet", &gameStatefulSetToUpdate, db)
		}
	}

	if len(*bkGameStatefulSets) > 1 {
		blog.Errorf("handleGameStatefulSetUpdate: More than one gameStatefulSet found")
		return fmt.Errorf("handleGameStatefulSetUpdate: More than one gameStatefulSet found")
	}

	return nil
}

// handle GameStatefulSet Delete
func (b *BcsBkcmdbSynchronizerHandler) handleGameStatefulSetDelete(
	gameStatefulSet *gsv1alpha1.GameStatefulSet, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	bkGameStatefulSets, err :=
		b.Syncer.GetBkWorkloads(bkCluster.BizID, "gameStatefulSet", &client.PropertyFilter{
			Condition: "AND",
			Rules: []client.Rule{
				{
					Field:    "name",
					Operator: "in",
					Value:    []string{gameStatefulSet.Name},
				},
				{
					Field:    "cluster_uid",
					Operator: "in",
					Value:    []string{bkCluster.Uid},
				},
				{
					Field:    "namespace",
					Operator: "in",
					Value:    []string{gameStatefulSet.Namespace},
				},
			},
		}, true, db)

	if err != nil {
		return err
	}

	if len(*bkGameStatefulSets) > 1 {
		return fmt.Errorf("len(bkGameStatefulSets) = %d", len(*bkGameStatefulSets))
	}

	if len(*bkGameStatefulSets) == 0 {
		return fmt.Errorf("gameStatefulSet %s not found", gameStatefulSet.Name)
	}

	bgs := (*bkGameStatefulSets)[0]
	bkGameStatefulSet := bkcmdbkube.GameStatefulSet{}
	err = common.InterfaceToStruct(bgs, &bkGameStatefulSet)
	if err != nil {
		blog.Errorf("convert bk gameStatefulSet failed, err: %s", err.Error())
		return err
	}

	// err = b.Syncer.DeleteBkWorkloads(b.BkCluster.BizID, "gameStatefulSet", &[]int64{bkGameStatefulSet.ID})
	err = retry.Do(
		func() error {
			return b.Syncer.DeleteBkWorkloads(bkCluster, "gameStatefulSet", &[]int64{bkGameStatefulSet.ID}, db)
		},
		retry.Delay(time.Second*1),
		retry.Attempts(2),
		retry.DelayType(retry.FixedDelay),
	)

	return err
}

// handle GameStatefulSet Create
func (b *BcsBkcmdbSynchronizerHandler) handleGameStatefulSetCreate(
	gameStatefulSet *gsv1alpha1.GameStatefulSet, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	bkNamespaces, err := b.Syncer.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{gameStatefulSet.Namespace},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)
	if err != nil {
		return err
	}

	if len(*bkNamespaces) != 1 {
		return fmt.Errorf("len(bkNamespaces) = %d", len(*bkNamespaces))
	}

	bkNamespace := (*bkNamespaces)[0]

	gameStatefulSetToAdd := make(map[int64][]client.CreateBcsWorkloadRequestData, 0)
	toAddData := b.Syncer.GenerateBkGameStatefulSet(&bkNamespace, &storage.GameStatefulSet{Data: gameStatefulSet})
	gameStatefulSetToAdd[bkNamespace.BizID] = []client.CreateBcsWorkloadRequestData{*toAddData}
	blog.Infof("gameStatefulSetToAdd: %s+%s+%s", bkCluster.Uid, gameStatefulSet.Namespace, gameStatefulSet.Name)

	b.Syncer.CreateBkWorkloads(bkCluster, "gameStatefulSet", gameStatefulSetToAdd, db)
	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleNamespace(
	msg amqp.Delivery, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	blog.Infof("handleNamespace Message: %v", msg.Headers)
	msgHeader, err := getMsgHeader(&msg.Headers)
	if err != nil {
		blog.Errorf("handleNamespace unable to get headers, err: %s", err.Error())
		return fmt.Errorf("handleNamespace unable to get headers, err: %s", err.Error())
	}

	blog.Infof("Headers: %s", msgHeader.ClusterId)
	namespace := &corev1.Namespace{}
	err = json.Unmarshal(msg.Body, namespace)
	if err != nil {
		blog.Errorf("handleNamespace: Unable to unmarshal")
		return fmt.Errorf("handleNamespace: Unable to unmarshal")
	}

	switch msgHeader.Event {
	case "update":
		err = b.handleNamespaceUpdate(namespace, bkCluster, db)
		if err != nil {
			blog.Errorf("handleNamespaceUpdate err: %s", err.Error())
			return fmt.Errorf("handleNamespaceUpdate err: %s", err.Error())
		}
	case "delete":
		err = b.handleNamespaceDelete(namespace, bkCluster, db)
		if err != nil {
			blog.Errorf("handleNamespaceDelete err: %s", err.Error())
			return fmt.Errorf("handleNamespaceDelete err: %s", err.Error())
		}
	default:
		blog.Errorf("handleNamespace: Unknown event: %s", msgHeader.Event)
	}
	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleNamespaceUpdate(
	namespace *corev1.Namespace, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	bkNamespaces, err := b.Syncer.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{namespace.Name},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)

	if err != nil {
		return err
	}

	if len(*bkNamespaces) == 0 {
		err := b.handleNamespaceCreate(namespace, bkCluster, db)
		if err != nil {
			blog.Errorf(fmt.Sprintf("handleNamespaceCreate err: %s", err.Error()))
			return err
		}
	}

	if len(*bkNamespaces) == 1 {
		bkNamespace := (*bkNamespaces)[0]
		nsToUpdate := make(map[int64]*client.UpdateBcsNamespaceRequestData, 0)
		needToUpdate, updateData := b.Syncer.CompareNamespace(&bkNamespace, &storage.Namespace{Data: namespace})
		if needToUpdate {
			nsToUpdate[bkNamespace.ID] = updateData
			blog.Infof("nsToUpdate: %s+%s", bkCluster.Uid, bkNamespace.Name)
			b.Syncer.UpdateBkNamespaces(bkCluster, &nsToUpdate, db)
		}
	}

	if len(*bkNamespaces) > 1 {
		blog.Errorf("handleNamespace: More than one namespace found")
		return fmt.Errorf("handleNamespace: More than one namespace found")
	}

	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleNamespaceDelete(
	namespace *corev1.Namespace, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	bkNamespaces, err := b.Syncer.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{namespace.Name},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)

	if err != nil {
		return err
	}

	if len(*bkNamespaces) > 1 {
		return fmt.Errorf("len(bkNamespaces) = %d", len(*bkNamespaces))
	}

	if len(*bkNamespaces) == 0 {
		return fmt.Errorf("namespace %s not found", namespace.Name)
	}

	bkNamespace := (*bkNamespaces)[0]
	blog.Infof("nsToDelete: %s+%s", bkCluster.Uid, bkNamespace.Name)

	// err = b.Syncer.DeleteBkNamespaces(b.BkCluster.BizID, &[]int64{bkNamespace.ID})
	err = retry.Do(
		func() error {
			return b.Syncer.DeleteBkNamespaces(bkCluster, &[]int64{bkNamespace.ID}, db)
		},
		retry.Delay(time.Second*1),
		retry.Attempts(2),
		retry.DelayType(retry.FixedDelay),
	)

	return err
}

// handle Namespace Create
func (b *BcsBkcmdbSynchronizerHandler) handleNamespaceCreate(
	namespace *corev1.Namespace, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	pmCli, err := b.Syncer.GetProjectManagerGrpcGwClient()
	if err != nil {
		blog.Errorf("get project manager grpc gw client failed, err: %s", err.Error())
		return nil
	}

	bizid := bkCluster.BizID
	if projectCode, ok := namespace.Annotations["io.tencent.bcs.projectcode"]; ok {
		gpr := pmp.GetProjectRequest{
			ProjectIDOrCode: projectCode,
		}

		if project, err := pmCli.Cli.GetProject(pmCli.Ctx, &gpr); err == nil {
			if project.Data.BusinessID != "" {
				bizid, err = strconv.ParseInt(project.Data.BusinessID, 10, 64)
				if err != nil {
					blog.Errorf("parse string err: %v", err)
				}
			}
		} else {
			blog.Errorf("get project error : %v", err)
		}
	}

	if bizid != bkCluster.BizID {
		bizid = int64(71)
	}

	nsToAdd := make(map[int64][]bkcmdbkube.Namespace, 0)
	nsToAdd[bizid] = []bkcmdbkube.Namespace{b.Syncer.GenerateBkNsData(bkCluster, &storage.Namespace{Data: namespace})}
	blog.Infof("nsToAdd: %s+%s", bkCluster.Uid, namespace.Name)
	b.Syncer.CreateBkNamespaces(bkCluster, nsToAdd, db)
	return nil
}

// Event handle
func (b *BcsBkcmdbSynchronizerHandler) handleEvent(msg amqp.Delivery, bkCluster *bkcmdbkube.Cluster) error { // nolint
	blog.Infof("handleEvent Message: %v", msg.Headers)
	msgHeader, err := getMsgHeader(&msg.Headers)
	if err != nil {
		blog.Errorf("handleEvent unable to get headers, err: %s", err.Error())
		return fmt.Errorf("handleEvent unable to get headers, err: %s", err.Error())
	}
	blog.Infof("Headers: %s", msgHeader.ClusterId)
	if resourceKind, resourceKindOk := msg.Headers["resourceKind"]; resourceKindOk {
		if resourceKind == "Pod" {
			if eventType, eventTypeOk := msg.Headers["type"]; eventTypeOk {
				switch eventType.(string) { // nolint
				case "BackOff":
					pod := corev1.Pod{}
					pod.Name = msgHeader.ResourceName
					pod.Namespace = msgHeader.Namespace
					err = b.handlePodDelete(&pod, bkCluster)
					if err != nil {
						blog.Errorf("handlePodDelete err: %s", err.Error())
						return fmt.Errorf("handlePodDelete err: %s", err.Error())
					}
				}
			}
		}
	}
	return nil
}

// Node handle
func (b *BcsBkcmdbSynchronizerHandler) handleNode(msg amqp.Delivery, bkCluster *bkcmdbkube.Cluster) error { // nolint
	blog.Infof("handleNode Message: %v", msg.Headers)
	msgHeader, err := getMsgHeader(&msg.Headers)
	if err != nil {
		blog.Errorf("handleNode unable to get headers, err: %s", err.Error())
		return fmt.Errorf("handleNode unable to get headers, err: %s", err.Error())
	}

	blog.Infof("Headers: %s", msgHeader.ClusterId)
	node := &corev1.Node{}
	err = json.Unmarshal(msg.Body, node)
	if err != nil {
		blog.Errorf("handleNode: Unable to unmarshal")
		return fmt.Errorf("handleNode: Unable to unmarshal")
	}

	switch msgHeader.Event {
	case "update": // nolint
		err = b.handleNodeUpdate(node, bkCluster)
		if err != nil {
			blog.Errorf("handleNodeUpdate err: %s", err.Error())
			return fmt.Errorf("handleNodeUpdate err: %s", err.Error())
		}
	case "delete": // nolint
		err = b.handleNodeDelete(node, bkCluster)
		if err != nil {
			blog.Errorf("handleNodeDelete err: %s", err.Error())
			return fmt.Errorf("handleNodeDelete err: %s", err.Error())
		}
	default:
		blog.Errorf("handleNode: Unknown event: %s", msgHeader.Event)
	}
	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleNodes(
	nodeMsg *msgBuffer, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	// blog.Infof("handleNode Message: %v", msg.Headers)
	// msgHeader, err := getMsgHeader(&msg.Headers)
	// if err != nil {
	//	blog.Errorf("handleNode unable to get headers, err: %s", err.Error())
	//	return fmt.Errorf("handleNode unable to get headers, err: %s", err.Error())
	// }
	//
	// blog.Infof("Headers: %s", msgHeader.ClusterId)
	// node := &corev1.Node{}
	// err = json.Unmarshal(msg.Body, node)
	// if err != nil {
	//	blog.Errorf("handleNode: Unable to unmarshal")
	//	return fmt.Errorf("handleNode: Unable to unmarshal")
	// }

	blog.Infof("nodeMsg: %d", len(nodeMsg.M))
	if time.Since(nodeMsg.T) < 10*time.Second {
		// blog.Infof("podMsg.T: %s, %s", podMsg.T, time.Now().Sub(podMsg.T))
		if len(nodeMsg.M) < 100 {
			return nil
		}
	}

	nodesUpdate := make(map[string]*corev1.Node)
	nodesDelete := make(map[string]*corev1.Node)

	for _, msg := range nodeMsg.M {
		blog.Infof("handleNode Message: %v", msg.Headers)
		msgHeader, err := getMsgHeader(&msg.Headers)
		if err != nil {
			blog.Errorf("handleNode unable to get headers, err: %s", err.Error())
			return fmt.Errorf("handleNode unable to get headers, err: %s", err.Error())
		}
		blog.Infof("Headers: %s", msgHeader.ClusterId)

		node := &corev1.Node{}
		err = json.Unmarshal(msg.Body, node)
		if err != nil {
			blog.Errorf("handleNode: Unable to unmarshal")
			return fmt.Errorf("handleNode: Unable to unmarshal")
		}
		switch msgHeader.Event {
		case "update":
			nodesUpdate[node.Name] = node
		case "delete":
			nodesDelete[node.Name] = node
			blog.Infof("nodesDelete: %s+%s", msg.Headers, node.Name)
		default:
			blog.Errorf("handleNode: Unknown event: %s", msgHeader.Event)
		}
	}

	err := b.handleNodesUpdate(nodesUpdate, bkCluster, db)
	if err != nil {
		blog.Errorf("handleNodesUpdate err: %s", err.Error())
		// return fmt.Errorf("handleNodesUpdate err: %s", err.Error())
	}

	err = b.handleNodesDelete(nodesDelete, bkCluster, db)
	if err != nil {
		blog.Errorf("handleNodesDelete err: %s", err.Error())
		// return fmt.Errorf("handlePodsDelete err: %s", err.Error())
	}

	nodeMsg.M = make([]amqp.Delivery, 0)
	nodeMsg.T = time.Now()

	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleNodesUpdate(
	nodesUpdate map[string]*corev1.Node, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	if len(nodesUpdate) == 0 {
		return nil
	}

	var nodeNames []string
	for _, v := range nodesUpdate {
		nodeNames = append(nodeNames, v.Name)
	}

	bkNodes, err := b.Syncer.GetBkNodes(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    nodeNames,
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)

	if err != nil {
		return err
	}

	bkNodesMap := make(map[string]bkcmdbkube.Node)

	for _, bkNode := range *bkNodes {
		bkNodesMap[*bkNode.Name] = bkNode
	}

	nodesDelete := make(map[string]*corev1.Node)
	nodesCreate := make(map[string]*corev1.Node)

	for k, k8sNode := range nodesUpdate {
		if bkNode, exist := bkNodesMap[k]; exist {
			// if k8sNode.Status.Phase != corev1.NodeRunning {
			//	nodesDelete[k8sNode.Name] = k8sNode
			//	blog.Infof("nodeToDelete: %s+%s", bkCluster.Uid, k8sNode.Name)
			//	continue
			// }
			nodeToUpdate := make(map[int64]*client.UpdateBcsNodeRequestData, 0)
			needToUpdate, updateData := b.Syncer.CompareNode(&bkNode, &storage.K8sNode{Data: k8sNode})
			if needToUpdate {
				nodeToUpdate[bkNode.ID] = updateData
				b.Syncer.UpdateBkNodes(bkCluster, &nodeToUpdate, db)
				blog.Infof("nodeToUpdate: %s+%s", bkCluster.Uid, *bkNode.Name)
			}

		} else {
			nodesCreate[k8sNode.Name] = k8sNode
		}
	}

	err = b.handleNodesDelete(nodesDelete, bkCluster, db)
	if err != nil {
		blog.Errorf("handleNodesDelete err: %s", err.Error())
		// return fmt.Errorf("handleNodesDelete err: %s", err.Error())
	}

	b.handleNodesCreate(nodesCreate, bkCluster, db)

	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleNodesDelete(
	nodesDelete map[string]*corev1.Node, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	if len(nodesDelete) == 0 {
		return nil
	}

	var nodeNames []string
	for _, v := range nodesDelete {
		nodeNames = append(nodeNames, v.Name)
	}

	bkNodes, err := b.Syncer.GetBkNodes(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    nodeNames,
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)

	if err != nil {
		return err
	}

	if len(*bkNodes) == 0 {
		return fmt.Errorf("node %s not found", nodeNames)
	}

	bkNodeIDs := make([]int64, 0)

	for _, bkNode := range *bkNodes {
		bkNodeIDs = append(bkNodeIDs, bkNode.ID)
		blog.Infof("nodeToDelete: %s+%s", bkCluster.Uid, *bkNode.Name)
	}

	// b.Syncer.DeleteBkNodes(b.BkCluster.BizID, &[]int64{bkNode.ID})
	err = retry.Do(
		func() error {
			return b.Syncer.DeleteBkNodes(bkCluster, &bkNodeIDs, db)
		},
		retry.Delay(time.Second*2),
		retry.Attempts(3),
		retry.DelayType(retry.FixedDelay),
	)

	return err
}

func (b *BcsBkcmdbSynchronizerHandler) handleNodeUpdate(node *corev1.Node, bkCluster *bkcmdbkube.Cluster) error { // nolint
	bkNodes, err := b.Syncer.GetBkNodes(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{node.Name},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, false, nil)

	if err != nil {
		return err
	}

	if len(*bkNodes) == 0 {
		err := b.handleNodeCreate(node, bkCluster)
		if err != nil {
			blog.Errorf(fmt.Sprintf("handleNodeCreate err: %s", err.Error()))
			return err
		}
	}

	if len(*bkNodes) == 1 {
		bkNode := (*bkNodes)[0]
		nodeToUpdate := make(map[int64]*client.UpdateBcsNodeRequestData, 0)
		needToUpdate, updateData := b.Syncer.CompareNode(&bkNode, &storage.K8sNode{Data: node})
		if needToUpdate {
			nodeToUpdate[bkNode.ID] = updateData
			b.Syncer.UpdateBkNodes(bkCluster, &nodeToUpdate, nil)
			blog.Infof("nodeToUpdate: %s+%s", bkCluster.Uid, bkNode.Name)
		}
	}

	if len(*bkNodes) > 1 {
		blog.Errorf("handleNode: More than one node found")
		return fmt.Errorf("handleNode: More than one node found")
	}

	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleNodeDelete(node *corev1.Node, bkCluster *bkcmdbkube.Cluster) error { // nolint
	bkNodes, err := b.Syncer.GetBkNodes(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "name",
				Operator: "in",
				Value:    []string{node.Name},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, false, nil)

	if err != nil {
		return err
	}

	if len(*bkNodes) > 1 {
		return fmt.Errorf("len(bkNodes) = %d", len(*bkNodes))
	}

	if len(*bkNodes) == 0 {
		return fmt.Errorf("node %s not found", node.Name)
	}

	bkNode := (*bkNodes)[0]

	// b.Syncer.DeleteBkNodes(b.BkCluster.BizID, &[]int64{bkNode.ID})
	blog.Infof("nodeToDelete: %s+%s", bkCluster.Uid, bkNode.Name)
	err = retry.Do(
		func() error {
			return b.Syncer.DeleteBkNodes(bkCluster, &[]int64{bkNode.ID}, nil)
		},
		retry.Delay(time.Second*1),
		retry.Attempts(2),
		retry.DelayType(retry.FixedDelay),
	)

	return err
}

func (b *BcsBkcmdbSynchronizerHandler) handleNodeCreate(node *corev1.Node, bkCluster *bkcmdbkube.Cluster) error { // nolint
	nodeToAdd := make([]client.CreateBcsNodeRequestData, 0)
	nodeData, err := b.Syncer.GenerateBkNodeData(bkCluster, &storage.K8sNode{Data: node})
	if err == nil {
		nodeToAdd = append(nodeToAdd, nodeData)
		b.Syncer.CreateBkNodes(bkCluster, &nodeToAdd, nil)
		blog.Infof("nodeToAdd: %s+%s", bkCluster.Uid, node.Name)
	}

	return nil
}

func (b *BcsBkcmdbSynchronizerHandler) handleNodesCreate(
	nodesCreate map[string]*corev1.Node, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) {
	nodeToAdd := make([]client.CreateBcsNodeRequestData, 0)

	for _, k8sNode := range nodesCreate {
		nodeData, err := b.Syncer.GenerateBkNodeData(bkCluster, &storage.K8sNode{Data: k8sNode})
		if err == nil {
			nodeToAdd = append(nodeToAdd, nodeData)
			blog.Infof("nodeToAdd: %s+%s", bkCluster.Uid, k8sNode.Name)
		}
	}

	b.Syncer.CreateBkNodes(bkCluster, &nodeToAdd, db)
}

func getMsgHeader(header *amqp.Table) (*MsgHeader, error) {
	var msgHeader MsgHeader
	if err := mapstructure.Decode(header, &msgHeader); err != nil {
		blog.Errorf("Unable to decode the message header, err: %s", err.Error())
		return nil, err
	}

	return &msgHeader, nil
}

// PublishMsg is a function that publishes a message to the RabbitMQ exchange.
func (b *BcsBkcmdbSynchronizerHandler) PublishMsg(msg amqp.Delivery, rep int32) error {
	if rep == 0 {
		rep = 2
	}
	// Set the exchange name with the source exchange name from the configuration.
	exchangeName := fmt.Sprintf("%s.headers", b.Syncer.BkcmdbSynchronizerOption.RabbitMQ.SourceExchange)

	// Check if the message has been republished before.
	if republish, ok := msg.Headers["republish"]; !ok {
		// If not, set the republish header to 1.
		msg.Headers["republish"] = 1
	} else {
		// If it has been republished before, check if the republish count is less than 10.
		if republish.(int32) > rep {
			// If it has been republished more than 10 times, return an error.
			return errors.New("no need to publish")
		}
		// If it is, increment the republish count.
		msg.Headers["republish"] = republish.(int32) + 1
	}

	// Publish the message to the exchange with the specified routing key.
	err := b.Chn.PublishWithContext(
		context.Background(),
		exchangeName,
		msg.RoutingKey,
		false,
		false,
		amqp.Publishing{
			Headers:      msg.Headers,
			DeliveryMode: msg.DeliveryMode,
			Body:         msg.Body,
		},
	)

	// If there is an error publishing the message, log the error.
	if err != nil {
		blog.Errorf("Error publishing message: %s", err.Error())
	}

	// Return the error if there is one, or nil if the message was published successfully.
	return err
}
