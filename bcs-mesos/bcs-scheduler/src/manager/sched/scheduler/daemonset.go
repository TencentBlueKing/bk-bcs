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

package scheduler

import (
	"net/http"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/offer"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/task"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/util"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/mesosproto/mesos"
)

//buildDaemonsetPod check builded taskgroup
//if some offers don't deploy daemonset, then build new taskgroup in the offer
func (s *Scheduler) buildDaemonsetPod(daemon *types.BcsDaemonset)error{
	blog.Infof("start check daemonset pod list, and build taskgroup on new offer")
	util.Lock.Lock(types.BcsDaemonset{}, daemon.GetUuid())
	defer util.Lock.UnLock(types.BcsDaemonset{}, daemon.GetUuid())

	//get current all mesos offers
	offers := s.GetAllOffers()
	for _,inoffer := range offers {
		//get offer hostip
		hostIp,ok := offer.GetOfferIp(inoffer.Offer)
		//if offer don't have InnerIP attribute, then continue
		if !ok {
			blog.Warnf("offer(%s:%s) don't have attribute InnerIP",
				inoffer.Offer.GetId().GetValue(), inoffer.Offer.GetHostname())
			continue
		}

		exist,err := s.checkPodWhetherBuildOffer(daemon, hostIp)
		//if exist==true or error, then continue
		if err!=nil || exist {
			continue
		}

		//the offer don't contain the damonset instance, then build new taskgroup on new pod

	}

}

//check daemonset whether build taskgroup in offer(hostIp)
//if build taskgroup return true, then return false
func (s *Scheduler) checkPodWhetherBuildOffer(daemon *types.BcsDaemonset, hostIp string)(bool,error){
	//if daemonset don't have any pod, return false
	if len(daemon.Pods)==0 {
		return false, nil
	}

	for _,podId :=range daemon.Pods {
		pod,err := s.store.FetchTaskGroup(podId.Name)
		if err!=nil {
			blog.Errorf("check daemonset(%s:%s) whether build offer, fetch taskgroup(%s) failed:",
				daemon.NameSpace, daemon.Name, podId.Name, err.Error())
			return false, err
		}
		//if pod.AgentIp == hostIp, show the offer have builded the daemonset taskgroup
		if hostIp==pod.GetAgentIp() {
			return true, nil
		}
	}

	return false, nil
}

func (s *Scheduler) doLaunchDaemonset(daemon *types.BcsDaemonset, outOffer *offer.Offer){
	//get offer innerip
	offerIp,_ := offer.GetOfferIp(outOffer.Offer)
	offer := outOffer.Offer
	cpus, mem, disk := s.OfferedResources(offer)
	blog.Info("launch daemonset(%s.%s) with offer:%s||%s, cpu:%f, mem:%f, disk:%f",
		daemon.NameSpace, daemon.Name, offerIp, *(offer.Id.Value), cpus, mem, disk)

	var taskGroupInfos []*mesos.TaskGroupInfo
	version,err := s.store.GetVersion(daemon.NameSpace, daemon.Name)
	if err!=nil {
		blog.Errorf("launch daemonset(%s:%s) with offer(%s), but get version failed:%s",
			daemon.NameSpace, daemon.Name, offerIp, err.Error())
		return
	}
	resources := task.BuildResources(version.AllResource())
	var taskgroupName string
	//check offer resource whether fit launch daemonset
	if s.IsOfferResourceFitLaunch(version.AllResource(), outOffer) && s.IsOfferExtendedResourcesFitLaunch(version.GetExtendedResources(), outOffer) {
		instanceId := uint64(len(daemon.Pods))
		//create taskgroup base on version
		taskgroup, err := task.CreateTaskGroup(version, "", instanceId, s.GetClusterId(), "", s.store)
		if err!=nil {
			blog.Errorf("launch daemonset(%s) create taskgroup err: %s", daemon.GetUuid(), err.Error())
			return
		}
		//save taskgroup
		err = s.store.SaveTaskGroup(taskgroup)
		if err != nil {
			blog.Errorf("launch daemonset(%s) save taskgroup(%s) err: %s", daemon.GetUuid(), taskgroup.ID, err.Error())
			return
		}
		taskgroupName = taskgroup.Name
		//create mesos taskgroup base inner taskgroup
		taskGroupInfo := task.CreateTaskGroupInfo(offer, version, resources, taskgroup)
		if taskGroupInfo == nil {
			blog.Warn("transaction %s: build taskgroupinfo fail", trans.ID)
			s.DeleteTaskGroup(app, taskGroup, "create taskgroupinfo fail")
			return
		}

		if err := s.store.SaveTaskGroup(taskGroup); err != nil {
			blog.Error("transaction %s: save taskgroup error %s", trans.ID, err.Error())
			s.DeleteTaskGroup(app, taskGroup, "save taskgroup fail")
			//trans.Status = types.OPERATION_STATUS_FAIL
			s.DeclineResource(offer.Id.Value)
			return
		}
		opData.LaunchedNum++
		taskGroupInfos = append(taskGroupInfos, taskGroupInfo)

		//lock agentsetting
		util.Lock.Lock(commtypes.BcsClusterAgentSetting{}, taskGroup.GetAgentIp())
		//update agentsettings taskgroup index info
		agentsetting, _ := s.store.FetchAgentSetting(taskGroup.GetAgentIp())
		if agentsetting != nil {
			agentsetting.Pods = append(agentsetting.Pods, taskGroup.ID)
			err := s.store.SaveAgentSetting(agentsetting)
			if err != nil {
				blog.Errorf("save agentsetting %s pods error %s", agentsetting.InnerIP, err.Error())
			}
		} else {
			blog.Errorf("fetch agentsetting %s Not Found", taskGroup.GetAgentIp())
		}
		util.Lock.UnLock(commtypes.BcsClusterAgentSetting{}, taskGroup.GetAgentIp())
	}

	if len(taskGroupInfos) <= 0 {
		blog.Error("transaction %s: have no any taskgroup to launch for this offer", trans.ID)
		s.DeclineResource(offer.Id.Value)
		return
	}

	resp, err := s.LaunchTaskGroups(offer, taskGroupInfos, version)
	if err != nil {
		blog.Error("transaction %s: launch taskgroups err:%s", trans.ID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return
	}
	if resp != nil && resp.StatusCode != http.StatusAccepted {
		blog.Error("transaction %s: resp status err code : %d", trans.ID, resp.StatusCode)
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return
	}

	//launch taskgroup success, and metrics
	reportScheduleTaskgroupMetrics(app.RunAs, app.Name, taskgroupName, LaunchTaskgroupType, started)

	if opData.LaunchedNum >= int(version.Instances) {
		blog.Info("transaction %s finish", trans.ID)
		app.LastStatus = app.Status
		app.Status = types.APP_STATUS_DEPLOYING
		app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
		app.UpdateTime = time.Now().Unix()
		app.Message = "application in deploying"
		trans.Status = types.OPERATION_STATUS_FINISH
	}

	err = s.store.SaveApplication(app)
	if err != nil {
		blog.Error("transaction %s finish, set application(%s.%s) to APP_STATUS_DEPLOYING err:%s",
			trans.ID, app.RunAs, app.ID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		return
	}
	blog.Info("transaction %s finish, set application(%s.%s) to APP_STATUS_DEPLOYING", trans.ID, app.RunAs, app.ID)
	blog.Info("do transaction %s end", trans.ID)
	return
}