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

// transaction for update application

package scheduler

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

// RunUpdateApplicationResource The goroutine function for update application quota transaction
// You can create a transaction for update application quota, then call this function to do it
// This function will come to end as soon as the transaction is done, fail or timeout(as defined by transaction.LifePeriod)
func (s *Scheduler) RunUpdateApplicationResource(transaction *types.Transaction) bool {

	runAs := transaction.Namespace
	appID := transaction.ObjectName
	blog.Infof("transaction %s update resource for application(%s.%s) run check",
		transaction.TransactionID, runAs, appID)

	if transaction.CreateTime.Add(transaction.DelayTime).After(time.Now()) {
		blog.V(3).Infof("transaction %s update resource for application(%s.%s) delaytime(%d), cannot do at now",
			transaction.TransactionID, runAs, appID, transaction.DelayTime)
		time.Sleep(3 * time.Second)
		return true
	}

	opData := transaction.CurOp.OpUpdateData
	index := 0
	var successTaskgroupIDs []string
	var failedTaskgroupIDs []string
	for index < opData.Instances {
		taskGroup := opData.Taskgroups[index]
		updated := false
		times := 0
		for times < 3 {
			times++
			blog.Infof("transaction %s try(%d times) update resource for taskgroup(%d: %s)",
				transaction.TransactionID, times, index, taskGroup.ID)
			ret := s.doUpdateTaskgroupResource(transaction, taskGroup.ID)
			if ret < 0 {
				transaction.Status = types.OPERATION_STATUS_FAIL
				blog.Warnf("transaction %s try(%d times) update resource for taskgroup(%d: %s) fail",
					transaction.TransactionID, times, index, taskGroup.ID)
				transaction.Message = fmt.Sprintf(
					"transaction %s try(%d times) update resource for taskgroup(%d: %s) fail",
					transaction.TransactionID, times, index, taskGroup.ID)
				return false
			}
			if ret == 0 {
				updated = true
				break
			}
			// wait to try again
			time.Sleep(2 * time.Second)
		}

		if updated == false {
			// continue to update next taskgroup even if current update operation failed
			blog.Warnf("transaction %s try(%d times) update resource for taskgroup(%d: %s) fail",
				transaction.TransactionID, times, index, taskGroup.ID)
			failedTaskgroupIDs = append(failedTaskgroupIDs, taskGroup.ID)
		} else {
			blog.Infof("transaction %s try(%d times) update resource for taskgroup(%d: %s) succ",
				transaction.TransactionID, times, index, taskGroup.ID)
			successTaskgroupIDs = append(successTaskgroupIDs, taskGroup.Name)
		}
		index++
	}

	if len(failedTaskgroupIDs) != 0 {
		transaction.Status = types.OPERATION_STATUS_FAIL
		transaction.Message = fmt.Sprintf("Warning: update resource %v failed, %v successfully",
			failedTaskgroupIDs, successTaskgroupIDs)
		return false
	}
	transaction.Status = types.OPERATION_STATUS_FINISH
	blog.Infof("update resource transaction %s finish, application(%s.%s) %d success, %d taskgroup failed",
		transaction.TransactionID, runAs, appID, len(successTaskgroupIDs), len(failedTaskgroupIDs))
	return false
}

func (s *Scheduler) doUpdateTaskgroupResource(trans *types.Transaction, taskGroupID string) int32 {

	runAs := trans.Namespace
	appID := trans.ObjectName

	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)

	blog.Info("transaction %s get taskgroup(%s) to do update resource", trans.TransactionID, taskGroupID)
	taskGroup, err := s.store.FetchTaskGroup(taskGroupID)
	if taskGroup == nil {
		blog.Error("transaction %s fetch taskgroup(%s) err(%s)", trans.TransactionID, taskGroupID, err.Error())
		return -1
	}

	if taskGroup.Status != types.TASKGROUP_STATUS_RUNNING {
		blog.Warnf("transaction %s get taskgroup(%s) status %s is not running, cannot update resource",
			trans.TransactionID, taskGroupID, taskGroup.Status)
		return 1
	}

	opData := trans.CurOp.OpUpdateData
	version := opData.Version
	baseCPU := taskGroup.CurrResource.Cpus
	baseMem := taskGroup.CurrResource.Mem
	baseDisk := taskGroup.CurrResource.Disk
	targetCPU := version.AllCpus()
	targetMem := version.AllMems()
	targetDisk := version.AllDisk()
	deltaCPU := targetCPU - baseCPU
	deltaMem := targetMem - baseMem
	deltaDisk := targetDisk - baseDisk

	lackLimit := false
	for _, container := range version.Container {
		if container.LimitResoures == nil {
			lackLimit = true
			break
		}
		if container.LimitResoures.Cpus == 0 || container.LimitResoures.Mem == 0 {
			lackLimit = true
			break
		}
	}

	if (deltaCPU < 0 || deltaMem < 0) || (deltaCPU == 0 && deltaMem == 0) || lackLimit {
		blog.Infof("taskgroup(%s) no need to update resource, stay(cpu: %f | mem: %f | disk: %f on host(%s)",
			taskGroup.ID, baseCPU, baseMem, baseDisk, taskGroup.HostName)
		return 0
	}
	blog.Infof("taskgroup(%s) update resource(cpu: %f->%f  | mem: %f->%f | disk: %f->%f) on host(%s)",
		taskGroup.ID, baseCPU, targetCPU, baseMem, targetMem, baseDisk, targetDisk, taskGroup.HostName)

	offerOut := s.GetFirstOffer()
	for offerOut != nil {
		offerIdx := offerOut.Id
		offer := offerOut.Offer

		curOffer := offerOut
		offerOut = s.GetNextOffer(offerOut)
		if taskGroup.HostName == offer.GetHostname() {

			// to update offer
			if s.UseOffer(curOffer) == false {
				blog.Warnf("transaction %s update resource for taskgroup(%s) use offer(%d) %s||%s fail",
					trans.TransactionID, taskGroup.ID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
				return 1
			}

			if s.IsOfferResourceFitUpdate(&types.Resource{Cpus: deltaCPU, Mem: deltaMem, Disk: deltaDisk}, curOffer) {

				// send a msg to executor for update resource
				blog.Infof("transaction %s send msg to update resource for taskgroup(%s)",
					trans.TransactionID, taskGroup.ID)
				s.notifyUpdateTaskgroupResource(taskGroup, version)

				totalDeltaCPU := targetCPU - taskGroup.LaunchResource.Cpus
				totalDeltaMem := targetMem - taskGroup.LaunchResource.Mem
				totalDeltaDisk := targetDisk - taskGroup.LaunchResource.Disk
				blog.Infof("transaction %s update delta resource for taskgroup(%s): %f | %f | %f",
					trans.TransactionID, taskGroup.ID, totalDeltaCPU, totalDeltaMem, totalDeltaDisk)
				s.UpdateAgentSchedInfo(taskGroup.HostName, taskGroup.ID,
					&types.Resource{
						Cpus: totalDeltaCPU,
						Mem:  totalDeltaMem,
						Disk: totalDeltaDisk})

				// update taskgroup
				taskGroup.CurrResource = version.AllResource()
				for index, task := range taskGroup.Taskgroup {
					if task.DataClass != nil {
						task.DataClass.Resources = version.Container[index].Resources
						task.DataClass.LimitResources = version.Container[index].LimitResoures
					}
				}
				s.store.SaveTaskGroup(taskGroup)
				s.DeclineResource(offer.Id.Value)
				return 0
			}

			blog.Warnf("transaction %s taskgroup(%s) has not enough resource on host(%s) to do update now",
				trans.TransactionID, taskGroup.ID, taskGroup.HostName)
			s.DeclineResource(offer.Id.Value)
			return 1
		}
	}

	// retry later
	return 1
}

func (s *Scheduler) notifyUpdateTaskgroupResource(
	taskGroup *types.TaskGroup, version *types.Version) (*types.BcsMessage, error) {

	msg := &types.Msg_UpdateTaskResources{}

	for _, task := range taskGroup.Taskgroup {

		exist := false
		for _, container := range version.Container {
			if container.Docker.Image == task.Image {
				exist = true
				blog.Warnf("task(%s: %s) update resource: %f | %f ",
					task.ID, task.Image, container.DataClass.Resources.Cpus, container.DataClass.Resources.Mem)
				msg.Resources = append(msg.Resources, &types.TaskResources{
					TaskId: &task.ID,
					ReqCpu: &container.DataClass.Resources.Cpus,
					ReqMem: &container.DataClass.Resources.Mem,
					Cpu:    &container.DataClass.LimitResources.Cpus,
					Mem:    &container.DataClass.LimitResources.Mem,
				})
			}
		}

		if exist == false {
			blog.Warnf("task(%s: %s) not find in version for updating resource", task.ID, task.Image)
		}
	}

	bcsMsg := &types.BcsMessage{
		Type:                types.Msg_UPDATE_TASK.Enum(),
		UpdateTaskResources: msg,
	}

	return s.SendBcsMessage(taskGroup, bcsMsg)
}
