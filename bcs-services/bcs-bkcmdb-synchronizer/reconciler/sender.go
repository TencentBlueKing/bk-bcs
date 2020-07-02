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

package reconciler

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmdb "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/cmdbv3"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/common"
)

// EventType type for reconciler event
type EventType int

const (
	// EventAdd add event
	EventAdd = iota
	// EventManyAdd add many event
	EventManyAdd
	// EventUpdate update event
	EventUpdate
	// EventDel delete event
	EventDel
)

// PodEvent struct for pod event
type PodEvent struct {
	Type EventType
	Pod  *common.Pod
}

// Sender sender for pod event
type Sender struct {
	clusterInfo common.Cluster
	index       int64
	cmdbClient  cmdb.ClientInterface
	queue       chan PodEvent
}

// NewSender create new sender with event queue
func NewSender(clusterInfo common.Cluster, index int64, queueLength int64, cmdbClient cmdb.ClientInterface) *Sender {
	queue := make(chan PodEvent, queueLength)
	return &Sender{
		clusterInfo: clusterInfo,
		index:       index,
		cmdbClient:  cmdbClient,
		queue:       queue,
	}
}

// Push push into queue
func (s *Sender) Push(pod PodEvent) {
	s.queue <- pod
}

func (s *Sender) logPre() string {
	return fmt.Sprintf("[%s-sender-%d]", s.clusterInfo.ClusterID, s.index)
}

// Run run the sender
func (s *Sender) Run(ctx context.Context) {
	for {
		select {
		case e := <-s.queue:
			switch e.Type {
			case EventAdd:
				result, err := s.cmdbClient.CreatePod(s.clusterInfo.BizID, &cmdb.CreatePod{
					Pod: e.Pod.ToMapInterface(),
				})
				if err != nil || !result.Result {
					blog.Warnf("%s create pod failed, %+v, %+v", s.logPre(), result, err)
				}
			case EventUpdate:
				result, err := s.cmdbClient.UpdatePod(s.clusterInfo.BizID, &cmdb.UpdatePod{
					UpdateOption: cmdb.UpdateOption{
						Condition: map[string]interface{}{
							"bk_pod_name":      e.Pod.PodName,
							"bk_pod_namespace": e.Pod.PodNamespace,
							"bk_pod_cluster":   e.Pod.PodCluster,
						},
						Data: e.Pod.ToMapInterface(),
					},
				})
				if err != nil || !result.Result {
					blog.Warnf("%s update pod failed, %+v, %+v", s.logPre(), result, err)
				}
			case EventDel:
				result, err := s.cmdbClient.DeletePod(s.clusterInfo.BizID, &cmdb.DeletePod{
					DeleteOption: cmdb.DeleteOption{
						Condition: map[string]interface{}{
							"bk_pod_name":      e.Pod.PodName,
							"bk_pod_namespace": e.Pod.PodNamespace,
							"bk_pod_cluster":   e.Pod.PodCluster,
						},
					},
				})
				if err != nil || !result.Result {
					blog.Warnf("%s delete pod failed, %+v, %+v", s.logPre(), result, err)
				}
			}
		case <-ctx.Done():
			blog.Infof("%s context done", s.logPre())
			return
		}
	}
}
