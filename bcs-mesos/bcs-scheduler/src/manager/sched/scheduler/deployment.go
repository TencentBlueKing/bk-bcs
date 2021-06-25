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
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
)

func (s *Scheduler) startCheckDeployments() {
	time.Sleep(60 * time.Second)
	blog.Info("check deployments begin")
	runAses, err := s.store.ListDeploymentRunAs()
	if err != nil {
		blog.Error("check deployment: list namespaces err:%s", err.Error())
		return
	}
	for _, runAs := range runAses {
		blog.Info("check deployments under namespace(%s)", runAs)
		IDs, err := s.store.ListDeploymentNodes(runAs)
		if err != nil {
			blog.Error("check deployments: list deployment under namespace(%s) err:%s, pass it", runAs, err.Error())
			continue
		}
		for _, ID := range IDs {
			blog.Infof("to check deployment(%s.%s)", runAs, ID)
			go s.DeploymentCheck(runAs, ID, true)
		}
	}

	return
}

// DeploymentCheck Check deployment status and maintain updating progress when it is in updating
// If the updating is finish or canceled, the function will come to end
func (s *Scheduler) DeploymentCheck(ns string, name string, recover bool) {

	blog.Info("deployment(%s.%s) check begin", ns, name)
	if recover == true {
		if s.Role != SchedulerRoleMaster {
			blog.Warn("deployment(%s.%s) check exit, because scheduler is not master now", ns, name)
			return
		}
		finish := s.deploymentCheckTick(ns, name, true)
		if finish == true {
			blog.Info("deployment(%s.%s) check finish", ns, name)
			return
		}
		time.Sleep(1 * time.Second)
	}

	for {
		if s.Role != SchedulerRoleMaster {
			blog.Warn("deployment(%s.%s) check exit, because scheduler is not master now", ns, name)
			return
		}
		finish := s.deploymentCheckTick(ns, name, false)
		if finish == true {
			blog.Info("deployment(%s.%s) check finish", ns, name)
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func (s *Scheduler) deploymentCheckTick(ns string, name string, recover bool) bool {
	blog.V(3).Infof("check rollingupdate for deployment(%s.%s)", ns, name)
	s.store.LockDeployment(fmt.Sprintf("%s.%s", ns, name))
	defer s.store.UnLockDeployment(fmt.Sprintf("%s.%s", ns, name))

	deployment, err := s.store.FetchDeployment(ns, name)
	if err != nil && err != store.ErrNoFound {
		blog.Warn("deployment(%s.%s) rolling update, fetch deployment err:%s",
			ns, name, err.Error())
		return false
	}
	if deployment == nil {
		blog.Warn("deployment(%s.%s) rolling update, deployment not exist, finish", ns, name)
		return true
	}

	if deployment.Status != types.DEPLOYMENT_STATUS_ROLLINGUPDATE &&
		deployment.Status != types.DEPLOYMENT_STATUS_ROLLINGUPDATE_PAUSED &&
		deployment.Status != types.DEPLOYMENT_STATUS_ROLLINGUPDATE_SUSPEND {
		blog.Infof("deployment(%s.%s) rolling update, deployment status(%s), finish", ns, name, deployment.Status)
		return true
	}

	if deployment.Status == types.DEPLOYMENT_STATUS_ROLLINGUPDATE_SUSPEND {
		blog.V(3).Infof("deployment(%s.%s) rolling update: deployment status(%s)", ns, name, deployment.Status)
		return false
	}

	//restart current rolling
	if recover == true {
		deployment.IsInRolling = false
	}
	//change check time
	deployment.CheckTime = time.Now().Unix()
	s.store.SaveDeployment(deployment)

	if deployment.IsInRolling == false {
		now := time.Now().Unix()
		if deployment.LastRollingTime > now {
			deployment.LastRollingTime = now
			s.store.SaveDeployment(deployment)
		}
		if now-deployment.LastRollingTime > int64(deployment.Strategy.RollingUpdate.UpgradeDuration) {
			if deployment.Status == types.DEPLOYMENT_STATUS_ROLLINGUPDATE_PAUSED {
				return false
			}
			return s.deploymentBeginRolling(deployment)
		}
	} else {
		return s.deploymentCheckRolling(deployment)
	}

	return false
}

func (s *Scheduler) deploymentBeginRolling(deployment *types.Deployment) bool {

	ns := deployment.ObjectMeta.NameSpace
	name := deployment.ObjectMeta.Name
	app, err := s.store.FetchApplication(ns, deployment.Application.ApplicationName)
	if err != nil && err != store.ErrNoFound {
		blog.Warn("deployment(%s.%s) rolling update: fetch application(%s.%s) err:%s",
			ns, name, ns, deployment.Application.ApplicationName, err.Error())
		return false
	}
	if app == nil {
		blog.Warn("deployment(%s.%s) rolling update, application(%s.%s) not exist, finish",
			ns, name, ns, deployment.Application.ApplicationName)
		return true
	}
	if app.Instances >= uint64(deployment.Strategy.RollingUpdate.MaxUnavailable) {
		deployment.Application.CurrentTargetInstances = int(
			app.Instances - uint64(deployment.Strategy.RollingUpdate.MaxUnavailable))
	} else {
		deployment.Application.CurrentTargetInstances = 0
	}
	deployment.Application.CurrentRollingInstances = int(app.Instances) - deployment.Application.CurrentTargetInstances

	appExt, err := s.store.FetchApplication(ns, deployment.ApplicationExt.ApplicationName)
	if err != nil && err != store.ErrNoFound {
		blog.Warn("deployment(%s.%s) rolling update: fetch ext application(%s.%s) err:%s",
			ns, name, ns, deployment.ApplicationExt.ApplicationName, err.Error())
		return false
	}
	if appExt == nil {
		blog.Warn("deployment(%s.%s) rolling update, ext application(%s.%s) not exist, finish",
			ns, name, ns, deployment.ApplicationExt.ApplicationName)
		return true
	}
	deployment.ApplicationExt.CurrentTargetInstances = int(
		appExt.Instances + uint64(deployment.Strategy.RollingUpdate.MaxSurge))

	if deployment.ApplicationExt.CurrentTargetInstances > int(appExt.DefineInstances) {
		deployment.ApplicationExt.CurrentTargetInstances = int(appExt.DefineInstances)
	}
	deployment.ApplicationExt.CurrentRollingInstances =
		deployment.ApplicationExt.CurrentTargetInstances - int(appExt.Instances)

	blog.Infof("====deployment(%s.%s) rolling update begin one step: application(%s: %d->%d),"+
		" applicationExt(%s: %d->%d)",
		ns, name, app.ID, app.Instances, deployment.Application.CurrentTargetInstances,
		appExt.ID, appExt.Instances, deployment.ApplicationExt.CurrentTargetInstances)

	deployment.IsInRolling = true
	deployment.LastRollingTime = time.Now().Unix()
	deployment.Message = ""

	if deployment.Strategy.RollingUpdate.RollingOrder == commtypes.CreateFirstOrder {
		// begin launch
		deployment.CurrRollingOp = types.DEPLOYMENT_OPERATION_START
		s.innerScaleApplication(appExt.RunAs, appExt.ID, uint64(deployment.ApplicationExt.CurrentTargetInstances))
		s.store.SaveDeployment(deployment)
	} else {
		// begin delete
		deployment.CurrRollingOp = types.DEPLOYMENT_OPERATION_DELETE
		s.innerScaleApplication(app.RunAs, app.ID, uint64(deployment.Application.CurrentTargetInstances))
		s.store.SaveDeployment(deployment)
	}

	return false
}

func (s *Scheduler) deploymentEndRolling(deployment *types.Deployment) {

	ns := deployment.ObjectMeta.NameSpace
	name := deployment.ObjectMeta.Name

	deployment.LastRollingTime = time.Now().Unix()
	deployment.IsInRolling = false
	deployment.CurrRollingOp = types.DEPLOYMENT_OPERATION_NIL
	deployment.Message = ""

	blog.Info("====deployment(%s.%s) rolling update finish one step", ns, name)

	if deployment.Strategy.RollingUpdate.RollingManually {
		deployment.Status = types.DEPLOYMENT_STATUS_ROLLINGUPDATE_PAUSED
		deployment.Message = "rollingupdate paused"
		blog.Info(
			"deployment(%s.%s) rolling update is paused by system after one step finish for RollingManually(ture)",
			ns, name)
	}
}

func (s *Scheduler) checkCreateFirstRollingStart(deployment *types.Deployment) bool {
	ns := deployment.ObjectMeta.NameSpace
	name := deployment.ObjectMeta.Name
	allDeleted := false
	allStarted := false
	appExt, err := s.store.FetchApplication(ns, deployment.ApplicationExt.ApplicationName)
	if err != nil {
		blog.Warn("deployment(%s.%s) rolling update: fetch ext application(%s) err:%s",
			ns, name, deployment.ApplicationExt.ApplicationName, err.Error())
		return false
	}
	if s.isRollingStartFinished(
		appExt, deployment.ApplicationExt.CurrentRollingInstances, deployment.ApplicationExt.CurrentTargetInstances) {
		blog.Info("deployment(%s.%s) rolling update, applicationExt(%s) instances change to %d/%d",
			ns, name, appExt.ID, appExt.Instances, appExt.DefineInstances)
		if appExt.Instances >= appExt.DefineInstances {
			allStarted = true
		}

		app, err := s.store.FetchApplication(ns, deployment.Application.ApplicationName)
		if err != nil {
			blog.Warn("deployment(%s.%s) rolling update, fetch application(%s) err:%s",
				ns, name, deployment.Application.ApplicationName, err.Error())
			return false
		}
		// do delete
		if app.Instances > uint64(deployment.Application.CurrentTargetInstances) {
			deployment.CurrRollingOp = types.DEPLOYMENT_OPERATION_DELETE
			deployment.LastRollingTime = time.Now().Unix()
			s.innerScaleApplication(app.RunAs, app.ID, uint64(deployment.Application.CurrentTargetInstances))
		} else {
			blog.Info("deployment(%s.%s) rolling update, application(%s) instances change to %d",
				ns, name, app.ID, app.Instances)
			s.deploymentEndRolling(deployment)
			if app.Instances <= 0 {
				allDeleted = true
			}
		}
		s.store.SaveDeployment(deployment)
	}

	if allDeleted == true && allStarted == true {
		return s.finishRollingUpdate(deployment)
	}

	return false
}

func (s *Scheduler) checkCreateFirstRollingDelete(deployment *types.Deployment) bool {
	ns := deployment.ObjectMeta.NameSpace
	name := deployment.ObjectMeta.Name
	allDeleted := false
	allStarted := false

	app, err := s.store.FetchApplication(ns, deployment.Application.ApplicationName)
	if err != nil {
		blog.Warn("deployment(%s.%s) rolling update: fetch application(%s) err:%s",
			ns, name, deployment.Application.ApplicationName, err.Error())
		return false
	}

	if app.Instances <= uint64(deployment.Application.CurrentTargetInstances) {
		blog.Info("deployment(%s.%s) rolling update: application(%s) instances change to %d",
			ns, name, app.ID, app.Instances)
		s.deploymentEndRolling(deployment)
		if app.Instances <= 0 {
			allDeleted = true
		}
		s.store.SaveDeployment(deployment)
	}

	if allDeleted == true {
		appExt, err := s.store.FetchApplication(ns, deployment.ApplicationExt.ApplicationName)
		if err != nil {
			blog.Warn("deployment(%s.%s) rolling update: fetch ext application(%s) err:%s",
				ns, name, deployment.ApplicationExt.ApplicationName, err.Error())
			return false
		}
		if appExt.Instances >= uint64(appExt.DefineInstances) {
			allStarted = true
		}
	}

	if allDeleted == true && allStarted == true {
		return s.finishRollingUpdate(deployment)
	}
	return false
}

func (s *Scheduler) checkCreateFirstRolling(deployment *types.Deployment) bool {

	if deployment.CurrRollingOp == types.DEPLOYMENT_OPERATION_START {
		return s.checkCreateFirstRollingStart(deployment)
	}

	return s.checkCreateFirstRollingDelete(deployment)
}

func (s *Scheduler) checkDeleteFirstRollingDelete(deployment *types.Deployment) bool {
	ns := deployment.ObjectMeta.NameSpace
	name := deployment.ObjectMeta.Name
	allDeleted := false
	allStarted := false
	app, err := s.store.FetchApplication(ns, deployment.Application.ApplicationName)
	if err != nil {
		blog.Warn("deployment(%s.%s) rolling update: fetch application(%s) err:%s",
			ns, name, deployment.Application.ApplicationName, err.Error())
		return false
	}
	if app.Instances <= uint64(deployment.Application.CurrentTargetInstances) {
		blog.Info("deployment(%s.%s) rolling update: application(%s) instances change to %d",
			ns, name, app.ID, app.Instances)
		if app.Instances <= 0 {
			allDeleted = true
		}
		appExt, err := s.store.FetchApplication(ns, deployment.ApplicationExt.ApplicationName)
		if err != nil {
			blog.Warn("deployment(%s.%s) rolling update: fetch ext application(%s) err:%s",
				ns, name, deployment.ApplicationExt.ApplicationName, err.Error())
			return false
		}
		// do start
		if appExt.Instances < uint64(deployment.ApplicationExt.CurrentTargetInstances) {
			deployment.CurrRollingOp = types.DEPLOYMENT_OPERATION_START
			deployment.LastRollingTime = time.Now().Unix()
			s.innerScaleApplication(appExt.RunAs, appExt.ID, uint64(deployment.ApplicationExt.CurrentTargetInstances))
		} else {
			blog.Info("deployment(%s.%s) rolling update: applicationExt(%s) instances change to %d/%d",
				ns, name, appExt.ID, appExt.Instances, appExt.DefineInstances)
			s.deploymentEndRolling(deployment)
			if appExt.Instances >= appExt.DefineInstances {
				allStarted = true
			}
		}
		s.store.SaveDeployment(deployment)
	}

	if allDeleted == true && allStarted == true {
		return s.finishRollingUpdate(deployment)
	}
	return false

}

func (s *Scheduler) checkDeleteFirstRollingStart(deployment *types.Deployment) bool {
	ns := deployment.ObjectMeta.NameSpace
	name := deployment.ObjectMeta.Name
	allDeleted := false
	allStarted := false
	appExt, err := s.store.FetchApplication(ns, deployment.ApplicationExt.ApplicationName)
	if err != nil {
		blog.Error("deployment(%s.%s) rolling update: fetch ext application(%s) err:%s",
			ns, name, deployment.ApplicationExt.ApplicationName)
		return false
	}
	if s.isRollingStartFinished(appExt, deployment.ApplicationExt.CurrentRollingInstances,
		deployment.ApplicationExt.CurrentTargetInstances) {
		blog.Info("deployment(%s.%s) rolling update: applicationExt(%s) instances change to %d/%d",
			ns, name, appExt.ID, appExt.Instances, appExt.DefineInstances)
		s.deploymentEndRolling(deployment)
		if appExt.Instances >= uint64(appExt.DefineInstances) {
			allStarted = true
		}
		s.store.SaveDeployment(deployment)
	}
	if allStarted == true {
		app, err := s.store.FetchApplication(ns, deployment.Application.ApplicationName)
		if err != nil {
			blog.Error("deployment(%s.%s) rolling update: fetch application(%s) err:%s",
				ns, name, deployment.Application.ApplicationName, err.Error())
			return false
		}
		if app.Instances <= 0 {
			allDeleted = true
		}
	}

	if allDeleted == true && allStarted == true {
		return s.finishRollingUpdate(deployment)
	}
	return false
}

func (s *Scheduler) checkDeleteFirstRolling(deployment *types.Deployment) bool {

	if deployment.CurrRollingOp == types.DEPLOYMENT_OPERATION_DELETE {
		return s.checkDeleteFirstRollingDelete(deployment)
	}

	return s.checkDeleteFirstRollingStart(deployment)
}

func (s *Scheduler) finishRollingUpdate(deployment *types.Deployment) bool {

	ns := deployment.ObjectMeta.NameSpace
	name := deployment.ObjectMeta.Name
	blog.Info("deployment(%s.%s) rolling update finish, call delete bind application(%s)",
		ns, name, deployment.Application.ApplicationName)
	s.InnerDeleteApplication(ns, deployment.Application.ApplicationName, false)
	blog.Info("deployment(%s.%s) rolling update finish, call delete bind application(%s) return",
		ns, name, deployment.Application.ApplicationName)

	s.store.LockApplication(ns + "." + deployment.ApplicationExt.ApplicationName)
	defer s.store.UnLockApplication(ns + "." + deployment.ApplicationExt.ApplicationName)
	app, err := s.store.FetchApplication(ns, deployment.ApplicationExt.ApplicationName)
	if err != nil && err != store.ErrNoFound {
		blog.Warn("deployment(%s.%s) rolling update finish,  get application(%s.%s) err %s",
			ns, deployment.ApplicationExt.ApplicationName, err.Error())
		return false
	}
	if app != nil {
		app.LastStatus = app.Status
		app.Status = types.APP_STATUS_RUNNING
		app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
		if err := s.store.SaveApplication(app); err != nil {
			blog.Error("deployment(%s.%s) rolling update finish, save application err:%s",
				ns, deployment.ApplicationExt.ApplicationName, err.Error())
			return false
		}
	}
	deployment.Status = types.DEPLOYMENT_STATUS_RUNNING
	deployment.Application = new(types.DeploymentReferApplication)
	deployment.Application.ApplicationName = deployment.ApplicationExt.ApplicationName
	deployment.ApplicationExt = nil
	deployment.LastRollingTime = 0
	deployment.CurrRollingOp = types.DEPLOYMENT_OPERATION_NIL
	deployment.IsInRolling = false
	deployment.Message = ""
	if err := s.store.SaveDeployment(deployment); err != nil {
		blog.Error("deployment(%s.%s) rolling update finish, save to db err:%s", ns, name, err.Error())
		return false
	}

	return true
}

func (s *Scheduler) deploymentCheckRolling(deployment *types.Deployment) bool {

	ns := deployment.ObjectMeta.NameSpace
	name := deployment.ObjectMeta.Name

	now := time.Now().Unix()
	var lifeperiod int64
	if deployment.CurrRollingOp == types.DEPLOYMENT_OPERATION_START {
		lifeperiod = TRANSACTION_DEPLOYMENT_ROLLING_UP_LIFEPERIOD
	} else {
		lifeperiod = TRANSACTION_DEPLOYMENT_ROLLING_DOWN_LIFEPERIOD
	}

	if deployment.LastRollingTime+lifeperiod+60 < now {
		blog.Warnf("====deployment(%s.%s) rolling update: %s timeout, suspend", ns, name, deployment.CurrRollingOp)
		if deployment.CurrRollingOp == types.DEPLOYMENT_OPERATION_START {
			deployment.Message = "create taskgroup timeout, rollingupdate suspend"
		} else {
			deployment.Message = "delete taskgroup timeout, rollingupdate suspend"
		}

		deployment.IsInRolling = false
		deployment.Status = types.DEPLOYMENT_STATUS_ROLLINGUPDATE_SUSPEND
		deployment.CurrRollingOp = types.DEPLOYMENT_OPERATION_NIL
		s.store.SaveDeployment(deployment)
		return false
	}

	if deployment.Strategy.RollingUpdate.RollingOrder == commtypes.CreateFirstOrder {
		return s.checkCreateFirstRolling(deployment)
	}

	return s.checkDeleteFirstRolling(deployment)
}

func (s *Scheduler) isRollingStartFinished(app *types.Application, rollingNum int, targetNum int) bool {

	if app.Instances < uint64(targetNum) {
		blog.Info("application(%s:%s) instances(%d) < targetNum(%d), rolling not finish",
			app.RunAs, app.ID, app.Instances, targetNum)
		return false
	}

	//add pod status check
	taskGroups, err := s.store.ListTaskGroups(app.RunAs, app.ID)
	if err != nil {
		blog.Error("list taskgroup(%s:%s) for rolling start check error:%s",
			app.RunAs, app.ID, err.Error())
		return false
	}

	hasHealthCheck := false
	for _, taskGroup := range taskGroups {

		Idx := taskGroup.InstanceID
		if Idx < uint64(targetNum-rollingNum) {
			continue
		}
		if Idx >= uint64(targetNum) {
			continue
		}

		if taskGroup.Status != types.TASKGROUP_STATUS_RUNNING {
			blog.Info("taskgroup(%s) status(%s), rolling update not finish",
				taskGroup.ID, taskGroup.Status)
			return false
		}

		for _, task := range taskGroup.Taskgroup {
			hasLocalCheck := false
			for _, healthStatus := range task.HealthCheckStatus {
				switch healthStatus.Type {
				case commtypes.BcsHealthCheckType_COMMAND:
					hasLocalCheck = true
					hasHealthCheck = true
				case commtypes.BcsHealthCheckType_TCP:
					hasLocalCheck = true
					hasHealthCheck = true
				case commtypes.BcsHealthCheckType_HTTP:
					hasLocalCheck = true
					hasHealthCheck = true
				}
			}
			if hasLocalCheck == false {
				continue
			}
			if task.IsChecked == false {
				blog.Info("task(%s) is running but not do healthcheck, rolling update not finish", task.ID)
				return false
			}
			if task.Healthy == false {
				blog.Info("task(%s) is running but healthcheck not ok, rolling update not finish", task.ID)
				return false
			}
		}
	}

	if hasHealthCheck {
		blog.Infof("application(%s:%s) taskgroup(%d) HealthCheck is ok, rolling update finish",
			app.RunAs, app.ID, app.Instances)
	} else {
		blog.Infof("application(%s:%s) taskgroup(%d) don't have HealthCheck, rolling update finish",
			app.RunAs, app.ID, app.Instances)
	}

	return true
}

func (s *Scheduler) innerScaleApplication(runAs, appID string, instances uint64) error {

	blog.Info("inner scale application(%s.%s) to instances:%d", runAs, appID, instances)
	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)
	app, err := s.store.FetchApplication(runAs, appID)
	if err != nil {
		blog.Error("fetch application(%s.%s) to do inner scale err:%s",
			runAs, appID, err.Error())
		return err
	}

	isDown := false
	if app.Instances == instances {
		return nil
	} else if app.Instances > instances {
		isDown = true
	}

	versions, err := s.store.ListVersions(runAs, appID)
	if err != nil {
		blog.Error("fail to list version(%s.%s), err:%s", runAs, appID, err.Error())
		return err
	}
	sort.Strings(versions)
	newestVersion := versions[len(versions)-1]
	version, err := s.store.FetchVersion(runAs, appID, newestVersion)
	if err != nil {
		blog.Error("fail fetch version(%s) for application(%s.%s), err:%s",
			newestVersion, runAs, appID, err.Error())
		return err
	}

	scaleTrans := &types.Transaction{
		TransactionID: types.GenerateTransactionID(string(commtypes.BcsDataType_APP)),
		ObjectKind:    string(commtypes.BcsDataType_APP),
		ObjectName:    appID,
		Namespace:     runAs,
		CreateTime:    time.Now(),
		CheckInterval: time.Second,
		CurOp: &types.TransactionOperartion{
			OpType: types.TransactionOpTypeInnerScale,
			OpScaleData: &types.TransAPIScaleOpdata{
				Version:      version,
				Instances:    instances,
				IsDown:       isDown,
				IsInner:      true,
				NeedResource: version.AllResource(),
			},
		},
		Status: types.OPERATION_STATUS_INIT,
	}

	if err := s.store.SaveTransaction(scaleTrans); err != nil {
		blog.Errorf("save transaction(%s,%s) into db failed, err %s", runAs, appID, err.Error())
		return err
	}

	s.PushEventQueue(scaleTrans)

	return nil
}

//InnerDeleteApplication delete application, specially for deployment
func (s *Scheduler) InnerDeleteApplication(runAs, appID string, enforce bool) error {
	blog.Info("inner delete application(%s.%s), enforce(%t)", runAs, appID, enforce)

	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)

	app, err := s.store.FetchApplication(runAs, appID)
	if err != nil {
		blog.Warn("inner delete application:  get application(%s.%s) err %s", runAs, appID, err.Error())
		return err
	}
	if app == nil {
		blog.Warn("inner delete application:  get application(%s.%s) return nil", runAs, appID)
		return errors.New("application not found")
	}

	if app.Status == types.APP_STATUS_OPERATING {
		blog.Warn("inner delete application:  application (%s.%s) is in status(%s) now ", runAs, appID, app.Status)
	}

	taskGroups, err := s.store.ListTaskGroups(runAs, appID)
	if err != nil {
		return err
	}

	for _, taskGroup := range taskGroups {
		blog.Info("inner delete application: kill taskgroup(%s)", taskGroup.ID)
		resp, err := s.KillTaskGroup(taskGroup)
		if err != nil {
			blog.Error("inner delete application: kill taskgroup(%s) failed: %s", taskGroup.ID, err.Error())
			return err
		}
		if resp == nil {
			blog.Error("inner delete application: kill taskGroup(%s) resp nil", taskGroup.ID)
			err = fmt.Errorf("kill taskGroup(%s) resp is nil", taskGroup.ID)
			return err
		} else if resp.StatusCode != http.StatusAccepted {
			blog.Error("inner delete application: kill taskGroup(%s) resp code %d", taskGroup.ID, resp.StatusCode)
			err = fmt.Errorf("kill taskGroup(%s) status code %d received", taskGroup.ID, resp.StatusCode)
			return err
		}
	}

	deleteTrans := &types.Transaction{
		ObjectKind:    string(commtypes.BcsDataType_APP),
		ObjectName:    appID,
		Namespace:     runAs,
		TransactionID: types.GenerateTransactionID(string(commtypes.BcsDataType_APP)),
		CreateTime:    time.Now(),
		CheckInterval: 3 * time.Second,
		CurOp: &types.TransactionOperartion{
			OpType: types.TransactionOpTypeDelete,
			OpDeleteData: &types.TransAPIDeleteOpdata{
				Enforce: enforce,
			},
		},
		Status: types.OPERATION_STATUS_INIT,
	}
	if err := s.store.SaveTransaction(deleteTrans); err != nil {
		blog.Errorf("save transaction(%s,%s) into db failed, err %s", runAs, appID, err.Error())
		return err
	}
	blog.Infof("transaction %s delete application(%s.%s) run begin",
		deleteTrans.TransactionID, deleteTrans.Namespace, deleteTrans.ObjectName)

	s.PushEventQueue(deleteTrans)

	app.LastStatus = app.Status
	app.Status = types.APP_STATUS_OPERATING
	app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
	app.UpdateTime = time.Now().Unix()
	app.Message = "application in deleting"
	if err := s.store.SaveApplication(app); err != nil {
		blog.Error("inner delete application: save application(%s.%s) status(%s) into db failed! err:%s",
			runAs, appID, app.Status, err.Error())
		return err
	}

	return nil
}
