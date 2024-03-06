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

package bcs

import (
	"sync"
	"time"

	"k8s.io/klog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/bcs/clustermanager"
	metricsinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/metrics"
)

var (
	scaleTypeUp   = "scaleUp"
	scaleTypeDown = "scaleDown"
	dateString    = "2006-01-02T15:04:05.999999999Z07:00"
	taskChecker   *TaskChecker
)

// TaskChecker checks task status
type TaskChecker struct {
	scaleUpTaskList   []string
	scaleDownTaskList []string
	lock              *sync.Mutex
	lastUpdateTime    time.Time
	client            clustermanager.NodePoolClientInterface
}

// NewTaskChecker news a task checker
func NewTaskChecker(client clustermanager.NodePoolClientInterface) *TaskChecker {
	return &TaskChecker{
		scaleUpTaskList:   make([]string, 0),
		scaleDownTaskList: make([]string, 0),
		lock:              &sync.Mutex{},
		client:            client,
	}
}

// RecordScaleUpTask append task id in scale up task list
func (t *TaskChecker) RecordScaleUpTask(id string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.scaleUpTaskList = append(t.scaleUpTaskList, id)
}

// RecordScaleDownTask append task id in scale down list
func (t *TaskChecker) RecordScaleDownTask(id string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.scaleDownTaskList = append(t.scaleDownTaskList, id)
}

func (t *TaskChecker) checkScaleTaskStatus() {

	now := time.Now()
	if t.lastUpdateTime.Add(2 * time.Minute).After(now) {
		klog.V(5).Infof("Refresh TaskCache latest updateTime %s, now %s, return",
			t.lastUpdateTime.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"))
		return
	}

	t.lock.Lock()
	defer t.lock.Unlock()

	if len(t.scaleUpTaskList) == 0 && len(t.scaleDownTaskList) == 0 {
		t.lastUpdateTime = time.Now()
		return
	}

	newScaleUpTaskList := t.performCheck(t.scaleUpTaskList, scaleTypeUp)
	t.scaleUpTaskList = newScaleUpTaskList

	newScaleDownTaskList := t.performCheck(t.scaleDownTaskList, scaleTypeDown)
	t.scaleDownTaskList = newScaleDownTaskList

	t.lastUpdateTime = time.Now()
}

func (t *TaskChecker) performCheck(taskList []string, scaleType string) []string {
	newTaskList := []string{}
	for _, taskID := range taskList {
		if taskID == "" {
			continue
		}
		task, err := t.client.GetTask(taskID)
		if err != nil {
			klog.Warningf("failed to get task %s: %v", taskID, err)
			newTaskList = append(newTaskList, taskID)
			continue
		}
		if task == nil {
			klog.Warningf("failed to get task %s: task is nil", taskID)
			newTaskList = append(newTaskList, taskID)
			continue
		}
		if task.TaskID == "" {
			klog.V(4).Infof("empty task return, origin: %s", taskID)
			newTaskList = append(newTaskList, taskID)
			continue
		}
		metricsinternal.UpdateScaleTask(task.TaskID, task.NodeGroupID, scaleType, task.Status)
		klog.V(4).Infof("type: %s, task: %s, nodeGroup: %s, status: %s",
			scaleType, task.TaskID, task.NodeGroupID, task.Status)
		// delete task after 10 min
		if task.End == "" {
			newTaskList = append(newTaskList, taskID)
			continue
		}
		endTime, err := time.Parse(dateString, task.End)
		if err != nil {
			klog.Warningf("failed to parse time from %s of task %s: %v", task.End, task.TaskID, err)
			newTaskList = append(newTaskList, taskID)
			continue
		}
		if endTime.Add(10 * time.Minute).Before(time.Now()) {
			klog.V(4).Infof("task %s has been finished for 10 min, ignore it", task.TaskID)
			continue
		}
		newTaskList = append(newTaskList, taskID)
	}
	return newTaskList
}
