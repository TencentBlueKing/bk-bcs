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
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/container"
)

//BcsTaskInfo task info relate to Mesos TaskGroupInfo & ContainerInfo
//we not only can find TaskInfo by containerID, but also can find ContainerInfo by TaskID
type BcsTaskInfo struct {
	lock          sync.RWMutex                           //lock for four maps
	ContainerRef  map[string]string                      //ContainerName to TaskID
	TaskRef       map[string]string                      //TaskID to ContainerName
	TaskInfo      map[string]*mesos.TaskInfo             //mesos task info, key is taskID
	ContainerInfo map[string]*container.BcsContainerInfo //BcsContainerInfo, key is containerName
}

//SetTaskInfo Set TaskInfo
func (bcsTask *BcsTaskInfo) SetTaskInfo(task *mesos.TaskInfo) error {
	bcsTask.lock.Lock()
	defer bcsTask.lock.Unlock()
	taskID := task.GetTaskId().GetValue()
	if _, exist := bcsTask.TaskInfo[taskID]; exist {
		return fmt.Errorf("Duplicate Task with ID %s", taskID)
	}
	bcsTask.TaskInfo[taskID] = task
	return nil
}

//SetContainerWithTaskID setting container with taskID.
func (bcsTask *BcsTaskInfo) SetContainerWithTaskID(taskID string, container *container.BcsContainerInfo) error {
	bcsTask.lock.Lock()
	defer bcsTask.lock.Unlock()
	if _, exist := bcsTask.TaskInfo[taskID]; !exist {
		return fmt.Errorf("No task %s in local", taskID)
	}
	//set container
	bcsTask.ContainerInfo[container.Name] = container
	//set reflection
	bcsTask.ContainerRef[container.Name] = taskID
	bcsTask.TaskRef[taskID] = container.Name
	return nil
}

//GetTask get TaskInfo by taskId
func (bcsTask *BcsTaskInfo) GetTask(taskID string) *mesos.TaskInfo {
	bcsTask.lock.Lock()
	defer bcsTask.lock.Unlock()
	return bcsTask.TaskInfo[taskID]
}

//GetTaskGroup return task group
func (bcsTask *BcsTaskInfo) GetTaskGroup() *mesos.TaskGroupInfo {
	bcsTask.lock.Lock()
	defer bcsTask.lock.Unlock()
	if len(bcsTask.TaskInfo) == 0 {
		return nil
	}
	group := &mesos.TaskGroupInfo{}
	for _, task := range bcsTask.TaskInfo {
		group.Tasks = append(group.Tasks, task)
	}
	return group
}

//GetContainer get BcsContainer by containerId
func (bcsTask *BcsTaskInfo) GetContainer(containerID string) *container.BcsContainerInfo {
	bcsTask.lock.Lock()
	defer bcsTask.lock.Unlock()
	return bcsTask.ContainerInfo[containerID]
}

//GetTaskByContainerID get container
func (bcsTask *BcsTaskInfo) GetTaskByContainerID(containerID string) *mesos.TaskInfo {
	bcsTask.lock.Lock()
	defer bcsTask.lock.Unlock()
	taskID, exist := bcsTask.ContainerRef[containerID]
	if !exist {
		return nil
	}
	if task, ok := bcsTask.TaskInfo[taskID]; ok {
		return task
	}
	return nil
}

//GetContainerByTaskID get container
func (bcsTask *BcsTaskInfo) GetContainerByTaskID(taskID string) *container.BcsContainerInfo {
	bcsTask.lock.Lock()
	defer bcsTask.lock.Unlock()
	containerID, exist := bcsTask.TaskRef[taskID]
	if !exist {
		return nil
	}
	if container, ok := bcsTask.ContainerInfo[containerID]; ok {
		return container
	}
	return nil
}

//CleanTask clean task & container by taskID
func (bcsTask *BcsTaskInfo) CleanTask(taskID string) (*mesos.TaskInfo, *container.BcsContainerInfo) {
	bcsTask.lock.Lock()
	defer bcsTask.lock.Unlock()
	task, taskExist := bcsTask.TaskInfo[taskID]
	if !taskExist {
		return nil, nil
	}
	delete(bcsTask.TaskInfo, taskID)
	containerID, idExist := bcsTask.TaskRef[taskID]
	if !idExist {
		return task, nil
	}
	delete(bcsTask.TaskRef, taskID)
	delete(bcsTask.ContainerRef, containerID)
	container, conExist := bcsTask.ContainerInfo[containerID]
	if !conExist {
		return task, nil
	}
	delete(bcsTask.ContainerInfo, containerID)
	return task, container
}

//CleanContainer clean task & container by containerID
func (bcsTask *BcsTaskInfo) CleanContainer(containerID string) (*mesos.TaskInfo, *container.BcsContainerInfo) {
	bcsTask.lock.Lock()
	defer bcsTask.lock.Unlock()
	container, conExist := bcsTask.ContainerInfo[containerID]
	if !conExist {
		return nil, container
	}
	delete(bcsTask.ContainerInfo, containerID)
	taskID, idExist := bcsTask.ContainerRef[containerID]
	if !idExist {
		return nil, container
	}
	delete(bcsTask.TaskRef, taskID)
	delete(bcsTask.ContainerRef, containerID)
	task, taskExist := bcsTask.TaskInfo[taskID]
	if !taskExist {
		return nil, container
	}
	delete(bcsTask.TaskInfo, taskID)
	return task, container
}

//Clean clean all data
func (bcsTask *BcsTaskInfo) Clean() {
	bcsTask.lock.Lock()
	defer bcsTask.lock.Unlock()
	for key := range bcsTask.ContainerInfo {
		delete(bcsTask.ContainerInfo, key)
		delete(bcsTask.ContainerRef, key)
	}
	for key := range bcsTask.TaskInfo {
		delete(bcsTask.TaskInfo, key)
		delete(bcsTask.TaskRef, key)
	}
}

//GetAllContainerID return all containerID
func (bcsTask *BcsTaskInfo) GetAllContainerID() (list []string) {
	bcsTask.lock.Lock()
	defer bcsTask.lock.Unlock()
	for key := range bcsTask.ContainerInfo {
		list = append(list, key)
	}
	return
}
