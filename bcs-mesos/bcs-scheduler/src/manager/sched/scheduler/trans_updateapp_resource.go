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
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"time"
	/*"encoding/json"
	"fmt"
	"net/http"*/)

// The goroutine function for update application quota transaction
// You can create a transaction for update application quota, then call this function to do it
// This function will come to end as soon as the transaction is done, fail or timeout(as defined by transaction.LifePeriod)
func (s *Scheduler) RunUpdateApplicationResource(transaction *Transaction) {

	runAs := transaction.RunAs
	appID := transaction.AppID

	blog.Infof("transaction %s update resource for application(%s.%s) run begin", transaction.ID, runAs, appID)
	for {
		blog.Infof("transaction %s update resource for application(%s.%s) run check", transaction.ID, runAs, appID)

		if transaction.CreateTime+transaction.DelayTime > time.Now().Unix() {
			blog.V(3).Infof("transaction %s update resource for application(%s.%s) delaytime(%d), cannot do at now",
				transaction.ID, runAs, appID, transaction.DelayTime)
			time.Sleep(3 * time.Second)
			continue
		}

		opData := transaction.OpData.(*TransAPIUpdateOpdata)
		index := 0
		for index < opData.Instances {
			taskGroup := opData.Taskgroups[index]
			updated := false
			times := 0
			for times < 3 {
				times++
				blog.Infof("transaction %s try(%d times) update resource for taskgroup(%d: %s)",
					transaction.ID, times, index, taskGroup.ID)
				ret := s.doUpdateTaskgroupResource(transaction, taskGroup.ID)
				if ret < 0 {
					transaction.Status = types.OPERATION_STATUS_FAIL
					blog.Warnf("transaction %s try(%d times) update resource for taskgroup(%d: %s) fail",
						transaction.ID, times, index, taskGroup.ID)
					goto run_end
				}
				if ret == 0 {
					updated = true
					break
				}
				// wait to try again
				time.Sleep(5 * time.Second)
			}

			if updated == false {
				blog.Infof("transaction %s try(%d times) update resource for taskgroup(%d: %s) fail",
					transaction.ID, times, index, taskGroup.ID)
				transaction.Status = types.OPERATION_STATUS_FAIL
				goto run_end
			}

			blog.Infof("transaction %s try(%d times) update resource for taskgroup(%d: %s) succ",
				transaction.ID, times, index, taskGroup.ID)
			index++
		}

		transaction.Status = types.OPERATION_STATUS_FINISH
		goto run_end
	}

run_end:
	s.FinishTransaction(transaction)
	blog.Infof("transaction %s update resource for application(%s.%s) run end, result(%s)",
		transaction.ID, runAs, appID, transaction.Status)

}

func (s *Scheduler) doUpdateTaskgroupResource(trans *Transaction, taskGroupID string) int32 {

	runAs := trans.RunAs
	appID := trans.AppID

	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)

	blog.Info("transaction %s get taskgroup(%s) to do update resource", trans.ID, taskGroupID)
	taskGroup, err := s.store.FetchTaskGroup(taskGroupID)
	if taskGroup == nil {
		blog.Error("transaction %s fetch taskgroup(%s) err(%s)", trans.ID, taskGroupID, err.Error())
		return -1
	}

	opData := trans.OpData.(*TransAPIUpdateOpdata)
	version := opData.Version
	baseCPU := taskGroup.CurrResource.Cpus
	baseMem := taskGroup.CurrResource.Mem
	baseDisk := taskGroup.CurrResource.Disk
	targetCPU := version.AllCpus()
	targetMem := version.AllMems()
	targetDisk := version.AllDisk()
	blog.Infof("taskgroup(%s) update resource(cpu: %f->%f  | mem: %f->%f | disk: %f->%f) on host(%s)",
		taskGroup.ID, baseCPU, targetCPU, baseMem, targetMem, baseDisk, targetDisk, taskGroup.HostName)

	//var offerIdx int64 = 0
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
					trans.ID, taskGroup.ID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
				return 1
			}

			deltaCPU := targetCPU - baseCPU
			deltaMem := targetMem - baseMem
			deltaDisk := targetDisk - baseDisk
			if s.IsOfferResourceFitUpdate(&types.Resource{Cpus: deltaCPU, Mem: deltaMem, Disk: deltaDisk}, curOffer) {

				// send a msg to executor for update resource
				blog.Infof("transaction %s send msg to update resource for taskgroup(%s)", trans.ID, taskGroup.ID)
				s.notifyUpdateTaskgroupResource(taskGroup, version)

				totalDeltaCPU := targetCPU - taskGroup.LaunchResource.Cpus
				totalDeltaMem := targetMem - taskGroup.LaunchResource.Mem
				totalDeltaDisk := targetDisk - taskGroup.LaunchResource.Disk
				blog.Infof("transaction %s update delta resource for taskgroup(%s): %f | %f | %f",
					trans.ID, taskGroup.ID, totalDeltaCPU, totalDeltaMem, totalDeltaDisk)
				s.UpdateAgentSchedInfo(taskGroup.HostName, taskGroup.ID, &types.Resource{Cpus: totalDeltaCPU, Mem: totalDeltaMem, Disk: totalDeltaDisk})

				// update taskgroup
				taskGroup.CurrResource = version.AllResource()
				s.store.SaveTaskGroup(taskGroup)

				s.DeclineResource(offer.Id.Value)
				return 0
			}

			blog.Warnf("transaction %s taskgroup(%s) has not enough resource on host(%s) to do update now",
				trans.ID, taskGroup.ID, taskGroup.HostName, err.Error())
			s.DeclineResource(offer.Id.Value)
			return 1
		}
	}

	// retry later
	return 1
}

func (s *Scheduler) notifyUpdateTaskgroupResource(taskGroup *types.TaskGroup, version *types.Version) (*types.BcsMessage, error) {

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
					Cpu:    &container.DataClass.Resources.Cpus,
					Mem:    &container.DataClass.Resources.Mem,
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
