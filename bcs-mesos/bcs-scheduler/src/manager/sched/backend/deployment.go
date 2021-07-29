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

package backend

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	comm "github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/utils"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
)

// GetDeployment get deployment by namespace and name
func (b *backend) GetDeployment(ns string, name string) (*types.Deployment, error) {
	return b.store.FetchDeployment(ns, name)
}

// CreateDeployment create deployment
func (b *backend) CreateDeployment(deploymentDef *types.DeploymentDef) (int, error) {
	ns := deploymentDef.ObjectMeta.NameSpace
	name := deploymentDef.ObjectMeta.Name
	b.store.LockDeployment(fmt.Sprintf("%s.%s", ns, name))
	defer b.store.UnLockDeployment(fmt.Sprintf("%s.%s", ns, name))

	blog.Info("request create deployment(%s.%s) begin", ns, name)
	currDeployment, err := b.store.FetchDeployment(ns, name)
	if err != nil && err != store.ErrNoFound {
		blog.Error("request create deployment(%s.%s), fetch deployment err:%s",
			ns, name, err.Error())
		return comm.BcsErrMesosSchedCommon, err
	}
	if currDeployment != nil {
		err := fmt.Errorf("deployment(%s.%s) already exist", ns, name)
		blog.Errorf("request create error: deployment(%s.%s) already exist", ns, name)
		return comm.BcsErrMesosSchedResourceExist, err
	}
	deployment := types.Deployment{
		ObjectMeta:    deploymentDef.ObjectMeta,
		Selector:      deploymentDef.Selector,
		Strategy:      deploymentDef.Strategy,
		Status:        types.DEPLOYMENT_STATUS_DEPLOYING,
		RawJson:       deploymentDef.RawJson,
		RawJsonBackup: nil,
	}
	if err = b.store.SaveDeployment(&deployment); err != nil {
		blog.Error("request create deployment, save deployment(%s.%s) err:%s", ns, name, err.Error())
		errin := fmt.Errorf("save deployment(%s.%s) err: %s", ns, name, err.Error())
		return comm.BcsErrCommCreateZkNodeFail, errin
	}

	//check current matching applications
	matchCount := 0
	matchApp := ""
	if deployment.Selector != nil {
		blog.Info("request create deployment(%s.%s): to match current applications by selector", ns, name)
		apps, err := b.ListApplications(ns)
		if err != nil {
			blog.Info("request to create deployment(%s.%s), list application under namespace(%s) err:%s",
				ns, name, ns, err.Error())
			errin := fmt.Errorf("list application under namespace(%s) err: %s", ns, err.Error())
			return comm.BcsErrCommListZkNodeFail, errin
		}
		for _, app := range apps {
			version, _ := b.GetVersion(app.RunAs, app.ID)
			if version == nil {
				blog.Warn("cannot get version for application(%s.%s), ignore it", app.RunAs, app.ID)
				continue
			}
			for kd, vd := range deployment.Selector {
				for ka, va := range version.Labels {
					if kd == ka && vd == va {
						blog.Info("to create deployment(%s.%s), matched application(%s.%s) by label(%s:%s)",
							ns, name, app.RunAs, app.ID, kd, vd)
						matchApp = app.ID
						matchCount++
					}
				}
			}
		}
	} else {
		blog.Info("request create deployment(%s.%s): selector is nil, will create new application", ns, name)
	}

	if matchCount == 0 {
		version := deploymentDef.Version
		if version == nil || version.RunAs != ns {
			blog.Error("request create deployment(%s.%s): version empty or namespace error", ns, name)
			return comm.BcsErrCommRequestDataErr, errors.New("version empty or namespace error")
		}

		blog.Info("request create deployment(%s.%s), to create and bind application(%s)", ns, name, version.ID)
		errcode, err := b.createDeploymentApplication(version)
		if err != nil {
			err1 := b.store.DeleteDeployment(ns, name)
			if err1 != nil {
				blog.Errorf("delete deployment %s:%s error %s", ns, name, err1.Error())
			}
			blog.Infof("delete deployment %s:%s success", ns, name)
			return errcode, err
		}

		deployment.Application = new(types.DeploymentReferApplication)
		deployment.Application.ApplicationName = version.ID
		deployment.Status = types.DEPLOYMENT_STATUS_RUNNING
		if err := b.store.SaveDeployment(&deployment); err != nil {
			blog.Error("request create deployment: save(%s.%s), err:%s", ns, name, err.Error())
			return comm.BcsErrCommCreateZkNodeFail, err
		}
	} else if matchCount == 1 {
		if deploymentDef.Version == nil {
			blog.Info("request create deployment(%s.%s), just bind current exist application(%s)", ns, name, matchApp)
			deployment.Application = new(types.DeploymentReferApplication)
			deployment.Application.ApplicationName = matchApp
			deployment.Status = types.DEPLOYMENT_STATUS_RUNNING
			if err := b.store.SaveDeployment(&deployment); err != nil {
				blog.Error("request create deployment: save(%s.%s), err:%s", ns, name, err.Error())
				return comm.BcsErrCommCreateZkNodeFail, err
			}
		} else {
			version := deploymentDef.Version
			if version.RunAs != ns {
				blog.Error("request create deployment(%s.%s): version namespace error", ns, name)
				return comm.BcsErrCommRequestDataErr, errors.New("version namespace error")
			}

			if version.ID == matchApp {
				version.ID = version.ID + "-v" + strconv.Itoa(int(time.Now().Unix()))
			}

			blog.Info("request create deployment(%s.%s), call delete current matched application(%s)",
				ns, name, matchApp)
			err := b.sched.InnerDeleteApplication(ns, matchApp, false)
			if err != nil {
				blog.Error("request create deployment(%s.%s), delete application(%s) err:%s",
					ns, name, matchApp, err.Error())
				return comm.BcsErrMesosSchedCommon, fmt.Errorf("delete current matched application(%s) error", matchApp)
			}

			blog.Info("request create deployment(%s.%s), to create and bind new application(%s)", ns, name, version.ID)
			errcode, err := b.createDeploymentApplication(version)
			if err != nil {
				err1 := b.store.DeleteDeployment(ns, name)
				if err1 != nil {
					blog.Errorf("delete deployment %s:%s error %s", ns, name, err1.Error())
				}
				return errcode, err
			}

			deployment.Application = new(types.DeploymentReferApplication)
			deployment.Application.ApplicationName = version.ID
			deployment.Status = types.DEPLOYMENT_STATUS_RUNNING
			if err = b.store.SaveDeployment(&deployment); err != nil {
				blog.Error("request create deployment: save(%s.%s), err:%s", ns, name, err.Error())
				return comm.BcsErrCommCreateZkNodeFail, err
			}
		}
	} else {
		err1 := b.store.DeleteDeployment(ns, name)
		if err1 != nil {
			blog.Errorf("delete deployment %s:%s error %s", ns, name, err1.Error())
		}
		blog.Error("request create deployment(%s.%s): too many application exist matching deployment", ns, name)
		return comm.BcsErrCommRequestDataErr, errors.New("too many application exist matching deployment")
	}

	blog.Info("request create deployment(%s.%s) end", ns, name)
	return comm.BcsSuccess, nil
}

func (b *backend) createDeploymentApplication(version *types.Version) (int, error) {

	if version == nil {
		err := errors.New("version data empty, cannot create application")
		return comm.BcsErrCommRequestDataErr, err
	}
	blog.Info("do create deployment application(%s.%s)", version.RunAs, version.ID)
	if version.Instances <= 0 {
		blog.Error("deployment application(%s.%s) instances(%d) err",
			version.RunAs, version.ID, version.Instances)
		err := errors.New("application instances error")
		return comm.BcsErrCommRequestDataErr, err
	}

	versionErr := b.CheckVersion(version)
	if versionErr != nil {
		blog.Error("deployment application(%s.%s) version error: %s",
			version.RunAs, version.ID, versionErr.Error())
		return comm.BcsErrCommRequestDataErr, versionErr
	}

	err := version.CheckAndDefaultResource()
	if err != nil {
		blog.Error("deployment application(%s.%s) version error: %s", version.RunAs, version.ID, err.Error())
		return comm.BcsErrCommRequestDataErr, err
	}
	if version.CheckConstraints() == false {
		blog.Error("deployment application(%s.%s) constraints error", version.RunAs, version.ID)
		err := errors.New("version constraints error")
		return comm.BcsErrCommRequestDataErr, err
	}
	app, err := b.store.FetchApplication(version.RunAs, version.ID)
	if err != nil && err != store.ErrNoFound {
		blog.Error("create deployment application, fetch application(%s.%s) ret:%s", version.RunAs, version.ID, err.Error())
		return comm.BcsErrCommGetZkNodeFail, err
	}

	if app != nil {
		err = errors.New("application duplicated when create application for deployment")
		blog.Warn("create deployment application(%s.%s): already exist", version.RunAs, version.ID)
		return comm.BcsErrCommRequestDataErr, err
	}
	application := types.Application{
		Kind:             version.Kind,
		ID:               version.ID,
		Name:             version.ID,
		DefineInstances:  uint64(version.Instances),
		Instances:        0,
		RunningInstances: 0,
		RunAs:            version.RunAs,
		ClusterId:        b.ClusterId(),
		Status:           types.APP_STATUS_STAGING,
		Created:          time.Now().Unix(),
		UpdateTime:       time.Now().Unix(),
		ObjectMeta:       version.ObjectMeta,
	}
	if err := b.SaveApplication(&application); err != nil {
		blog.Error("create deployment application: fail to SaveApplication(%s.%s), err:%s",
			application.RunAs, application.ID, err.Error())
		return comm.BcsErrCommCreateZkNodeFail, err
	}
	if err := b.SaveVersion(version.RunAs, version.ID, version); err != nil {
		blog.Error("create deployment application: fail to SaveVersion(%s.%s), err:%s",
			version.RunAs, version.ID, err.Error())
		return comm.BcsErrCommCreateZkNodeFail, err
	}
	if err := b.LaunchApplication(version); err != nil {
		blog.Error("create deployment application: application(%s.%s) launch error: %s",
			version.RunAs, version.ID, err.Error())
		return comm.BcsErrMesosSchedCommon, err
	}

	return comm.BcsSuccess, nil
}

// UpdateDeployment update deployment
func (b *backend) UpdateDeployment(deployment *types.DeploymentDef) (int, error) {
	ns := deployment.ObjectMeta.NameSpace
	name := deployment.ObjectMeta.Name
	blog.Info("request update deployment(%s.%s) begin", ns, name)

	if deployment.Strategy.RollingUpdate == nil {
		blog.Error("update deployment(%s.%s): rollingupdate strategy not set", ns, name)
		return comm.BcsErrCommRequestDataErr, errors.New("update strategy not set")
	}
	if deployment.Strategy.RollingUpdate.RollingOrder != commtypes.CreateFirstOrder &&
		deployment.Strategy.RollingUpdate.RollingOrder != commtypes.DeleteFirstOrder {
		blog.Error("update deployment(%s.%s): RollingOrder(%s) err",
			ns, name, deployment.Strategy.RollingUpdate.RollingOrder)
		return comm.BcsErrCommRequestDataErr, errors.New("update strategy rolling order error")
	}
	if deployment.Strategy.RollingUpdate.MaxUnavailable <= 0 {
		deployment.Strategy.RollingUpdate.MaxUnavailable = 1
	}
	if deployment.Strategy.RollingUpdate.MaxSurge <= 0 {
		deployment.Strategy.RollingUpdate.MaxSurge = 1
	}

	b.store.LockDeployment(fmt.Sprintf("%s.%s", ns, name))
	defer b.store.UnLockDeployment(fmt.Sprintf("%s.%s", ns, name))

	currDeployment, err := b.store.FetchDeployment(ns, name)
	if err != nil && err != store.ErrNoFound {
		blog.Error("update deployment(%s.%s) fetch deployment err: %s", ns, name, err.Error())
		return comm.BcsErrCommGetZkNodeFail, err
	}

	if currDeployment == nil {
		err = errors.New("deployment not exist")
		blog.Warn("update deployment(%s.%s): data not exist", ns, name)
		return comm.BcsErrMesosSchedNotFound, err
	}

	if currDeployment.Status != types.DEPLOYMENT_STATUS_RUNNING {
		err = errors.New("deployment is not running, cannot update")
		blog.Warn("update deployment(%s.%s): status(%s) is not running", ns, name, currDeployment.Status)
		return comm.BcsErrMesosSchedCommon, err
	}

	if currDeployment.Application == nil {
		err = errors.New("deployment has not application, cannot do update, you can delete and recreate it")
		blog.Warn("update deployment(%s.%s): no bind application", ns, name)
		return comm.BcsErrMesosSchedNotFound, err
	}

	// lock current application
	b.store.LockApplication(ns + "." + currDeployment.Application.ApplicationName)
	defer b.store.UnLockApplication(ns + "." + currDeployment.Application.ApplicationName)

	app, err := b.store.FetchApplication(ns, currDeployment.Application.ApplicationName)
	if err != nil && err != store.ErrNoFound {
		blog.Info("update deployment(%s.%s), fetch application(%s) err:%s",
			ns, name, currDeployment.Application.ApplicationName, err.Error())
		return comm.BcsErrCommGetZkNodeFail, err
	}

	if app == nil {
		err = errors.New("deployment has not application, cannot update, you can delete and recreate it")
		blog.Warn("update deployment(%s.%s): no bind application", ns, name)
		return comm.BcsErrMesosSchedNotFound, err
	}
	if app.Status != types.APP_STATUS_RUNNING && app.Status != types.APP_STATUS_ABNORMAL {
		err = errors.New("deployment bind application is not running, cannot update, please try later")
		blog.Warn("update deployment(%s.%s): application under status(%s) can not update", ns, name, app.Status)
		return comm.BcsErrMesosSchedCommon, err
	}

	for k, v := range currDeployment.Selector {
		if deployment.Selector[k] != v {
			err = errors.New("deployment's selector cannot be updated'")
			blog.Warn("update deployment(%s.%s): selector(%s-%s) changed", ns, name, k, v)
			return comm.BcsErrCommRequestDataErr, err
		}
	}
	for k, v := range deployment.Selector {
		if currDeployment.Selector[k] != v {
			err = errors.New("deployment's selector cannot be updated'")
			blog.Warn("update deployment(%s.%s): selector(%s-%s) changed", ns, name, k, v)
			return comm.BcsErrCommRequestDataErr, err
		}
	}

	//create extension application but not launch
	version := deployment.Version
	version.ID = version.ID + "-v" + strconv.Itoa(int(time.Now().Unix()))
	blog.Info("update deployment(%s.%s): create application(%s.%s)",
		ns, name, version.RunAs, version.ID)
	if version.Instances <= 0 {
		blog.Error("update deployment, application(%s.%s) instances(%d) err",
			version.RunAs, version.ID, version.Instances)
		err := errors.New("instances error")
		return comm.BcsErrCommRequestDataErr, err
	}
	if version.RunAs != ns {
		blog.Error("update deployment, application(%s.%s) namespace err", version.RunAs, version.ID)
		err := errors.New("namespace error")
		return comm.BcsErrCommRequestDataErr, err
	}

	versionErr := b.CheckVersion(version)
	if versionErr != nil {
		blog.Error("update deployment, application(%s.%s) version error: %s",
			version.RunAs, version.ID, versionErr.Error())
		return comm.BcsErrCommRequestDataErr, versionErr
	}

	err = version.CheckAndDefaultResource()
	if err != nil {
		blog.Error("update deployment, application(%s.%s) version error: %s", version.RunAs, version.ID, err.Error())
		return comm.BcsErrCommRequestDataErr, err
	}
	if version.CheckConstraints() == false {
		err := errors.New("constraints error")
		return comm.BcsErrCommRequestDataErr, err
	}
	// lock extension application
	b.store.LockApplication(ns + "." + version.ID)
	defer b.store.UnLockApplication(ns + "." + version.ID)

	app, err = b.store.FetchApplication(version.RunAs, version.ID)
	if err != nil && err != store.ErrNoFound {
		blog.Error("update deployment, fetch application(%s.%s) err:%s", version.RunAs, version.ID, err.Error())
		return comm.BcsErrCommGetZkNodeFail, err
	}

	if app != nil {
		err = errors.New("application already exist")
		blog.Error("update deployment, application(%s.%s) already exist", version.RunAs, version.ID)
		return comm.BcsErrMesosSchedCommon, err
	}
	application := types.Application{
		Kind:             version.Kind,
		ID:               version.ID,
		Name:             version.ID,
		DefineInstances:  uint64(version.Instances),
		Instances:        0,
		RunningInstances: 0,
		RunAs:            version.RunAs,
		ClusterId:        b.ClusterId(),
		Status:           types.APP_STATUS_STAGING,
		Created:          time.Now().Unix(),
		UpdateTime:       time.Now().Unix(),
		ObjectMeta:       version.ObjectMeta,
	}

	if err := b.SaveApplication(&application); err != nil {
		blog.Error("update deployment, save application(%s.%s), err:%s",
			application.RunAs, application.ID, err.Error())
		return comm.BcsErrCommCreateZkNodeFail, err
	}

	if err := b.SaveVersion(version.RunAs, version.ID, version); err != nil {
		blog.Error("update deployment, save application(%s.%s), err:%s",
			version.RunAs, version.ID, err.Error())
		return comm.BcsErrCommCreateZkNodeFail, err
	}

	blog.Info("request update deployment(%s.%s), create and bind application(%s)", ns, name, application.ID)
	currDeployment.ApplicationExt = new(types.DeploymentReferApplication)
	currDeployment.ApplicationExt.ApplicationName = application.ID
	currDeployment.Strategy = deployment.Strategy
	currDeployment.Status = types.DEPLOYMENT_STATUS_ROLLINGUPDATE
	currDeployment.LastRollingTime = 0
	currDeployment.CurrRollingOp = types.DEPLOYMENT_OPERATION_NIL
	currDeployment.IsInRolling = false

	//store deployment definition
	currDeployment.RawJsonBackup = currDeployment.RawJson
	currDeployment.RawJson = deployment.RawJson

	if err := b.store.SaveDeployment(currDeployment); err != nil {
		blog.Error("update deployment(%s.%s), save deployment err: %s", ns, name, err.Error())
		return comm.BcsErrCommCreateZkNodeFail, err
	}
	//set applications status to RollingUpdate
	app, err = b.store.FetchApplication(ns, currDeployment.Application.ApplicationName)
	if err != nil {
		blog.Error("update deployment, fetch application(%s.%s) err:%s",
			ns, currDeployment.Application.ApplicationName, err.Error())
		return comm.BcsErrCommGetZkNodeFail, err
	}

	if app != nil {
		app.LastStatus = app.Status
		app.Status = types.APP_STATUS_ROLLINGUPDATE
		app.SubStatus = types.APP_SUBSTATUS_ROLLINGUPDATE_DOWN
		err = b.store.SaveApplication(app)
		if err != nil {
			blog.Errorf("save app %s error %s", app.ID, err.Error())
		}
	}

	appExt, err := b.store.FetchApplication(ns, currDeployment.ApplicationExt.ApplicationName)
	if err != nil {
		blog.Error("update deployment, fetch ext application(%s.%s) err:%s",
			ns, currDeployment.ApplicationExt.ApplicationName, err.Error())
		return comm.BcsErrCommGetZkNodeFail, err
	}

	if appExt != nil {
		appExt.LastStatus = appExt.Status
		appExt.Status = types.APP_STATUS_ROLLINGUPDATE
		appExt.SubStatus = types.APP_SUBSTATUS_ROLLINGUPDATE_UP

		err = b.store.SaveApplication(appExt)
		if err != nil {
			blog.Errorf("save app %s error %s", app.ID, err.Error())
		}
	}

	go b.sched.DeploymentCheck(ns, name, false)

	blog.Info("request update deployment(%s.%s) end", ns, name)
	return comm.BcsSuccess, nil
}

// CancelUpdateDeployment cancel deployment update process
func (b *backend) CancelUpdateDeployment(ns string, name string) error {
	blog.Info("request cancelupdate deployment(%s.%s) begin", ns, name)
	b.store.LockDeployment(fmt.Sprintf("%s.%s", ns, name))
	defer b.store.UnLockDeployment(fmt.Sprintf("%s.%s", ns, name))

	deployment, err := b.store.FetchDeployment(ns, name)
	if err != nil {
		blog.Error("request cancelupdate deployment(%s.%s) fetch deployment err: %s", ns, name, err.Error())
		return err
	}

	blog.Info("cancelupdate deployment(%s.%s): current status(%s)", ns, name, deployment.Status)
	if deployment.Status != types.DEPLOYMENT_STATUS_ROLLINGUPDATE && deployment.Status != types.DEPLOYMENT_STATUS_ROLLINGUPDATE_PAUSED && deployment.Status != types.DEPLOYMENT_STATUS_ROLLINGUPDATE_SUSPEND {
		err := errors.New("deployment is in not rollingupdate")
		blog.Warn("request cancelupdate deployment(%s.%s): status(%s) err", ns, name, deployment.Status)
		return err
	}

	times := 0
	for {
		if deployment.IsInRolling && deployment.CurrRollingOp == types.DEPLOYMENT_OPERATION_DELETE {
			times++
			if times > 8 {
				blog.Error("request cancelupdate deployment(%s.%s): in deleting taskgroups", ns, name)
				return errors.New("deployment is deleting taskgroups, cannot cancelupdate now, try later")
			}
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}

	if deployment.ApplicationExt != nil {
		blog.Info("cancelupdate deployment(%s.%s), call delete ext application(%s)",
			ns, name, deployment.ApplicationExt.ApplicationName)
		err = b.sched.InnerDeleteApplication(ns, deployment.ApplicationExt.ApplicationName, false)
		if err != nil {
			blog.Errorf("delete app (%s:%s) error %s", ns, deployment.ApplicationExt.ApplicationName, err.Error())
		}
		blog.Info("cancelupdate deployment(%s.%s), call delete ext application(%s) return",
			ns, name, deployment.ApplicationExt.ApplicationName)
		deployment.ApplicationExt = nil
	}

	if deployment.Application == nil {
		blog.Warn("cancelupdate deployment(%s.%s), no bind application", ns, name)
		return errors.New("no application to do recover")
	}

	blog.Info("cancelupdate deployment(%s.%s), call recover application(%s)",
		ns, name, deployment.Application.ApplicationName)
	app, err := b.store.FetchApplication(ns, deployment.Application.ApplicationName)
	if err != nil {
		blog.Error("cancelupdate deployment(%s.%s): fetch application(%s), err:%s",
			ns, name, deployment.Application.ApplicationName, err.Error())
		return err
	}

	version, _ := b.GetVersion(app.RunAs, app.ID)
	if version == nil {
		blog.Error("cancelupdate deployment(%s.%s): fail to get version for application(%s.%s)",
			ns, name, app.RunAs, app.ID)
		err = errors.New("cannot get version to do recover")
		return err
	}
	err = b.RecoverApplication(version)
	if err != nil {
		blog.Error("cancelupdate deployment(%s.%s), recover application err: %s",
			ns, name, err.Error())
		return err
	}
	blog.Info("cancelupdate deployment(%s.%s), call recover application(%s) return",
		ns, name, deployment.Application.ApplicationName)
	deployment.Application.CurrentTargetInstances = 0

	// rollback deployment definition
	deployment.RawJson = deployment.RawJsonBackup
	deployment.RawJsonBackup = nil

	deployment.Status = types.DEPLOYMENT_STATUS_RUNNING
	deployment.ApplicationExt = nil
	deployment.LastRollingTime = 0
	deployment.CurrRollingOp = types.DEPLOYMENT_OPERATION_NIL
	deployment.IsInRolling = false
	if err := b.store.SaveDeployment(deployment); err != nil {
		blog.Info("cancelupdate deployment(%s.%s), save deployment err:%s", ns, name, err.Error())
		return err
	}

	blog.Info("request cancelupdate deployment(%s.%s) end", ns, name)
	return nil
}

// DeleteDeployment do delete deployment
func (b *backend) DeleteDeployment(ns string, name string, enforce bool) (int, error) {
	blog.Info("request delete deployment(%s.%s) begin", ns, name)
	b.store.LockDeployment(fmt.Sprintf("%s.%s", ns, name))
	defer b.store.UnLockDeployment(fmt.Sprintf("%s.%s", ns, name))

	deployment, err := b.store.FetchDeployment(ns, name)
	if err != nil && err != store.ErrNoFound {
		blog.Error("delete deployment(%s.%s) fetch deployment err: %s", ns, name, err.Error())
		return comm.BcsErrCommGetZkNodeFail, err
	}
	if deployment == nil {
		err := errors.New("deployment not exist")
		blog.Warn("delete deployment(%s.%s), deployment not exist", ns, name)
		return comm.BcsErrMesosSchedNotFound, err
	}

	blog.Info("delete deployment(%s.%s): current status(%s)", ns, name, deployment.Status)

	if deployment.Application != nil {
		blog.Info("delete deployment(%s.%s), call delete bind application(%s)",
			ns, name, deployment.Application.ApplicationName)
		err = b.sched.InnerDeleteApplication(ns, deployment.Application.ApplicationName, enforce)
		if err != nil {
			blog.Errorf("delete application(%s.%s) error %s", ns, deployment.Application.ApplicationName, err.Error())
		}
		blog.Info("delete deployment(%s.%s), call delete bind application(%s) return",
			ns, name, deployment.Application.ApplicationName)
	} else {
		blog.Warn("delete deployment(%s.%s), no bind application", ns, name)
	}

	if deployment.ApplicationExt != nil {
		blog.Info("delete deployment(%s.%s), call delete binded extension application(%s)",
			ns, name, deployment.ApplicationExt.ApplicationName)
		err = b.sched.InnerDeleteApplication(ns, deployment.ApplicationExt.ApplicationName, enforce)
		if err != nil {
			blog.Errorf("delete application(%s.%s) error %s", ns, deployment.ApplicationExt.ApplicationName, err.Error())
		}
		blog.Info("delete deployment(%s.%s), call delete binded extension application(%s) return",
			ns, name, deployment.ApplicationExt.ApplicationName)
	}

	deployment.Status = types.DEPLOYMENT_STATUS_DELETING
	deployment.Message = "waiting applications to be deleted"
	if err := b.store.SaveDeployment(deployment); err != nil {
		blog.Info("save deployment(%s.%s) to db err:%s", ns, name, err.Error())
		return comm.BcsErrCommCreateZkNodeFail, err
	}

	go b.CheckDeleteDeployment(ns, name)

	blog.Info("request delete deployment(%s.%s) end", ns, name)
	return comm.BcsSuccess, nil
}

// CheckDeleteDeployment check deployment deletion
func (b *backend) CheckDeleteDeployment(ns string, name string) {
	blog.Infof("check delete deployment(%s.%s)", ns, name)

	b.store.LockDeployment(fmt.Sprintf("%s.%s", ns, name))
	defer b.store.UnLockDeployment(fmt.Sprintf("%s.%s", ns, name))

	deployment, err := b.store.FetchDeployment(ns, name)
	if err != nil {
		blog.Warn("check delete deployment(%s.%s): get deployment err(%s)", ns, name, err.Error())
		return
	}

	if deployment.Status != types.DEPLOYMENT_STATUS_DELETING {
		blog.Warn("check delete deployment(%s.%s): status(%s) not in deleting", ns, name, deployment.Status)
		return
	}

	if deployment.Application != nil {
		_, err := b.store.FetchApplication(ns, deployment.Application.ApplicationName)
		if err == store.ErrNoFound {
			deployment.Application = nil
		} else {
			blog.Infof("check delete deployment(%s.%s), application(%s) not deleted",
				ns, name, deployment.Application.ApplicationName)
			deployment.Message = "application still not deleted: " + deployment.Application.ApplicationName
		}
	}

	if deployment.ApplicationExt != nil {
		_, err := b.store.FetchApplication(ns, deployment.ApplicationExt.ApplicationName)
		if err == store.ErrNoFound {
			deployment.ApplicationExt = nil
		} else {
			blog.Infof("check delete deployment(%s.%s), application(%s) not deleted",
				ns, name, deployment.ApplicationExt.ApplicationName)
			deployment.Message = "application still not deleted: " + deployment.ApplicationExt.ApplicationName
		}
	}

	// real delete
	if deployment.Application == nil && deployment.ApplicationExt == nil {
		blog.Info("real delete deployment(%s.%s)", ns, name)
		if err := b.store.DeleteDeployment(ns, name); err != nil {
			blog.Warn("delete deployment(%s.%s) from db err:%s", ns, name, err.Error())
		}
		return

	}
	time.Sleep(3 * time.Second)
	go b.CheckDeleteDeployment(ns, name)
	return
}

// PauseUpdateDeployment pause deployment update process
func (b *backend) PauseUpdateDeployment(ns string, name string) error {
	blog.Info("request pauseupdate deployment(%s.%s) begin", ns, name)
	b.store.LockDeployment(fmt.Sprintf("%s.%s", ns, name))
	defer b.store.UnLockDeployment(fmt.Sprintf("%s.%s", ns, name))

	deployment, err := b.store.FetchDeployment(ns, name)
	if err != nil && err != store.ErrNoFound {
		blog.Error("pauseupdate deployment(%s.%s) fetch deployment err: %s", ns, name, err.Error())
		return err
	}
	if deployment == nil {
		err := errors.New("deployment not exist")
		blog.Warn("pauseupdate deployment(%s.%s): data not exist", ns, name)
		return err
	}

	blog.Info("pauseupdate deployment(%s.%s): current status(%s)", ns, name, deployment.Status)
	if deployment.Status != types.DEPLOYMENT_STATUS_ROLLINGUPDATE {
		err := errors.New("deployment is not in updating")
		blog.Warn("pauseupdate deployment(%s.%s): status(%s) err", ns, name, deployment.Status)
		return err
	}

	deployment.Status = types.DEPLOYMENT_STATUS_ROLLINGUPDATE_PAUSED
	if err := b.store.SaveDeployment(deployment); err != nil {
		blog.Error("pauseupdate deployment(%s.%s), save deployment err:%s", ns, name, err.Error())
		return err
	}

	blog.Info("request pauseupdate deployment(%s.%s) end", ns, name)
	return nil
}

// ResumeUpdateDeployment resume deployment update process
func (b *backend) ResumeUpdateDeployment(ns string, name string) error {
	blog.Info("request resumeupdate deployment(%s.%s) begin", ns, name)
	b.store.LockDeployment(fmt.Sprintf("%s.%s", ns, name))
	defer b.store.UnLockDeployment(fmt.Sprintf("%s.%s", ns, name))

	deployment, err := b.store.FetchDeployment(ns, name)
	if err != nil && err != store.ErrNoFound {
		blog.Error("resumeupdate deployment(%s.%s) fetch deployment err: %s", ns, name, err.Error())
		return err
	}
	if deployment == nil {
		err := errors.New("deployment not exist")
		blog.Warn("resumeupdate deployment(%s.%s): data not exist", ns, name)
		return err
	}

	blog.Info("resumeupdate deployment(%s.%s): current status(%s)", ns, name, deployment.Status)
	if deployment.Status != types.DEPLOYMENT_STATUS_ROLLINGUPDATE_PAUSED && deployment.Status != types.DEPLOYMENT_STATUS_ROLLINGUPDATE_SUSPEND {
		err := errors.New("deployment is not update paused or suspend")
		blog.Warn("resumeupdate deployment(%s.%s): status(%s) err", ns, name, deployment.Status)
		return err
	}

	deployment.Status = types.DEPLOYMENT_STATUS_ROLLINGUPDATE
	if err := b.store.SaveDeployment(deployment); err != nil {
		blog.Error("resumeupdate deployment(%s.%s), save deployment err:%s", ns, name, err.Error())
		return err
	}

	blog.Info("request resumeupdate deployment(%s.%s) end", ns, name)

	return nil
}

// ScaleDeployment scale deployment to certain instances
func (b *backend) ScaleDeployment(runAs, name string, instances uint64) error {
	blog.Info("request scale deployment(%s.%s) to instances %d", runAs, name, instances)
	b.store.LockDeployment(fmt.Sprintf("%s.%s", runAs, name))
	defer b.store.UnLockDeployment(fmt.Sprintf("%s.%s", runAs, name))

	deployment, err := b.store.FetchDeployment(runAs, name)
	if err != nil && err != store.ErrNoFound {
		blog.Error("scale deployment(%s.%s) fetch deployment err: %s", runAs, name, err.Error())
		return err
	}
	if deployment == nil {
		err := errors.New("deployment not exist")
		blog.Warn("scale deployment(%s.%s): data not exist", runAs, name)
		return err
	}

	blog.Info("scale deployment(%s.%s): current status(%s)", runAs, name, deployment.Status)

	if deployment.Status != types.DEPLOYMENT_STATUS_RUNNING {
		blog.Warn("scale deployment(%s %s) status (%s), cannot scale now ", runAs, name, deployment.Status)
		return fmt.Errorf("operation forbidden, the deployment status is %s", deployment.Status)
	}

	if deployment.Application == nil {
		blog.Warn("scale deployment(%s %s) application is nil", runAs, name)
		return fmt.Errorf("deployment has no application, can't scale")
	}

	return b.ScaleApplication(runAs, deployment.Application.ApplicationName, instances, "", false)
}

// UpdateDeploymentResource update deployment resource only
func (b *backend) UpdateDeploymentResource(deployment *types.DeploymentDef) (int, error) {
	ns := deployment.ObjectMeta.NameSpace
	name := deployment.ObjectMeta.Name
	blog.Infof("request update deployment (%s.%s) resource begin", ns, name)

	b.store.LockDeployment(fmt.Sprintf("%s.%s", ns, name))
	defer b.store.UnLockDeployment(fmt.Sprintf("%s.%s", ns, name))

	// when current deployment is not found or status is not running, it cannot be updated
	currDeployment, err := b.store.FetchDeployment(ns, name)
	if err != nil && err != store.ErrNoFound {
		blog.Error("update deployment(%s.%s) resource, fetch deployment err: %s", ns, name, err.Error())
		return comm.BcsErrCommGetZkNodeFail, err
	}
	if currDeployment == nil {
		err = errors.New("deployment not exist")
		blog.Warn("update deployment(%s.%s) resource: data not exist", ns, name)
		return comm.BcsErrMesosSchedNotFound, err
	}
	if currDeployment.Status != types.DEPLOYMENT_STATUS_RUNNING {
		err = errors.New("deployment is not running, cannot update resource")
		blog.Warn("update deployment(%s.%s) resource: status(%s) is not running", ns, name, currDeployment.Status)
		return comm.BcsErrMesosSchedCommon, err
	}

	// check deployment differences
	if deployment.ObjectMeta.Name != currDeployment.ObjectMeta.Name ||
		deployment.ObjectMeta.NameSpace != currDeployment.ObjectMeta.NameSpace ||
		!reflect.DeepEqual(deployment.Selector, currDeployment.Selector) ||
		!reflect.DeepEqual(deployment.Strategy, currDeployment.Strategy) {
		err = errors.New("cannot change deployment meta and strategy when update resource")
		blog.Warnf("update deployment(%s.%s) resource: meta data changed", ns, name, currDeployment.Status)
		return comm.BcsErrMesosSchedCommon, err
	}

	// lock current application
	b.store.LockApplication(ns + "." + currDeployment.Application.ApplicationName)
	defer b.store.UnLockApplication(ns + "." + currDeployment.Application.ApplicationName)

	app, err := b.store.FetchApplication(ns, currDeployment.Application.ApplicationName)
	if err != nil && err != store.ErrNoFound {
		blog.Warnf("update deployment(%s.%s) resource, fetch application(%s) err:%s",
			ns, name, currDeployment.Application.ApplicationName, err.Error())
		return comm.BcsErrCommGetZkNodeFail, err
	}
	if app.Status == types.APP_STATUS_OPERATING || app.Status == types.APP_STATUS_ROLLINGUPDATE {
		blog.Warnf("application(%s.%s) of deployment(%s.%s) cannot do update under status(%s)",
			ns, currDeployment.Application.ApplicationName, ns, name, app.Status)
		return comm.BcsErrMesosSchedCommon, fmt.Errorf(
			"application(%s.%s) of deployment(%s.%s) cannot do update under status(%s)",
			ns, currDeployment.Application.ApplicationName, ns, name, app.Status)
	}
	currVersion, _ := b.GetVersion(ns, currDeployment.Application.ApplicationName)
	if currVersion == nil {
		blog.Warnf("update deployment(%s.%s) resource failed, err get current version", ns, name)
		return comm.BcsErrCommGetZkNodeFail, err
	}

	err = utils.IsOnlyResourceIncreased(currVersion, deployment.Version)
	if err != nil {
		blog.Warnf("check update resource failed, err %s", err.Error())
		return comm.BcsErrMesosSchedCommon, fmt.Errorf("check update resource failed, err %s", err.Error())
	}

	// launch update transaction
	updateTrans := &types.Transaction{
		TransactionID: types.GenerateTransactionID(string(commtypes.BcsDataType_DEPLOYMENT)),
		ObjectKind:    string(commtypes.BcsDataType_DEPLOYMENT),
		ObjectName:    name,
		Namespace:     ns,
		CreateTime:    time.Now(),
		CheckInterval: 3 * time.Second,
		CurOp: &types.TransactionOperartion{
			OpType: types.TransactionOpTypeDepUpdateResource,
			OpDepUpdateData: &types.TransDeploymentUpdateOpData{
				Version:          deployment.Version,
				IsUpdateResource: true,
			},
		},
		Status: types.OPERATION_STATUS_INIT,
	}

	// set deployment status
	currDeployment.Status = types.DEPLOYMENT_STATUS_UPDATERESOURCE
	currDeployment.RawJsonBackup = currDeployment.RawJson
	currDeployment.RawJson = deployment.RawJson
	if err = b.store.SaveDeployment(currDeployment); err != nil {
		blog.Errorf("update deployment(%s.%s) to status %s failed, err %s",
			ns, name, types.DEPLOYMENT_STATUS_UPDATERESOURCE, err.Error())
		return comm.BcsErrMesosSchedCommon, fmt.Errorf("update deployment(%s.%s) to status %s failed, err %s",
			ns, name, types.DEPLOYMENT_STATUS_UPDATERESOURCE, err.Error())
	}

	// save version
	if err = b.store.SaveVersion(deployment.Version); err != nil {
		return comm.BcsErrMesosSchedCommon, fmt.Errorf(
			"save version(%s.%s) failed when UpdateDeploymentResource, err %s", ns, name, err.Error())
	}

	if err := b.store.SaveTransaction(updateTrans); err != nil {
		blog.Errorf("save transaction(%s,%s) into db failed, err %s", ns, name, err.Error())
		return comm.BcsErrMesosSchedCommon, fmt.Errorf(
			"save transaction(%s,%s) into db failed, err %s", ns, name, err.Error())
	}
	b.sched.PushEventQueue(updateTrans)
	blog.Infof("request update resource of deployment(%s.%s) end", ns, name)
	return comm.BcsSuccess, nil
}
