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
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commonTypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	sched "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/scheduler"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/task"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"sort"
	"time"
)

//ScaleApplication is used to scale application instances.
func (b *backend) ScaleApplication(runAs, appID string, instances uint64, kind commonTypes.BcsDataType, isFromAPI bool) error {
	blog.V(3).Infof("scale application(%s.%s) to instances:%d", runAs, appID, instances)

	b.store.LockApplication(runAs + "." + appID)
	defer b.store.UnLockApplication(runAs + "." + appID)

	app, err := b.store.FetchApplication(runAs, appID)
	if err != nil {
		blog.Error("fetch application(%s.%s) to scale err:%s", runAs, appID, err.Error())
		return err
	}

	// added  20181011, app.Kind is a new field and the current running app kind maybe empty.
	// if app.Kind is empty and request kind is PROCESS, the scale will not be allowed.
	// if app.Kind is not empty and it is different from request kind, also the scale will not be allowed.
	// isFromAPI means the caller is from API, we should check the scale kind, or if false it means caller is from inner
	// functions, which should not be checked.
	currentKind := app.Kind
	if currentKind == "" {
		currentKind = commonTypes.BcsDataType_APP
	}
	if isFromAPI && currentKind != kind {
		blog.Warnf("try to scale %s through kind %s failed", currentKind, kind)
		return fmt.Errorf("target is not %s, can not be scale", kind)
	}

	if app.Status != types.APP_STATUS_RUNNING && app.Status != types.APP_STATUS_ABNORMAL {
		blog.Warn("application(%s.%s) status(%s), cannot scale now ", runAs, appID, app.Status)
		return fmt.Errorf("Operation Not Allowed, the status of the app is %s", app.Status)
	}

	versions, err := b.store.ListVersions(runAs, appID)
	if err != nil {
		blog.Error("scale application(%s.%s) fail, list version err:%s", runAs, appID, err.Error())
		return err
	}

	sort.Strings(versions)
	newestVersion := versions[len(versions)-1]
	version, err := b.store.FetchVersion(runAs, appID, newestVersion)
	if err != nil {
		blog.Error("scale application(%s.%s) fail, fetch version(%s) err:%s", runAs, appID, newestVersion, err.Error())
		return err
	}

	requestIpNum := task.GetVersionRequestIpCount(version)
	if requestIpNum > 0 && requestIpNum < int(instances) {
		blog.Error("scale application(%s.%s) fail, label netsvc.requestip not enough", runAs, appID)
		return fmt.Errorf("application(%s.%s) cannot scale for label netsvc.requestip not enough", runAs, appID)
	}

	blog.Info("get newest version(%s) for application(%s.%s) to do scale", newestVersion, runAs, appID)
	version.Instances = int32(instances)
	err = b.store.SaveVersion(version)
	if err != nil {
		return err
	}

	app.DefineInstances = instances

	var isDown bool

	if app.Instances > instances {
		isDown = true
	}

	scaleTrans := sched.CreateTransaction()
	scaleTrans.RunAs = runAs
	scaleTrans.AppID = appID
	scaleTrans.OpType = types.OPERATION_SCALE
	scaleTrans.Status = types.OPERATION_STATUS_INIT

	var scaleOpdata sched.TransAPIScaleOpdata
	scaleOpdata.Version = version
	scaleOpdata.NeedResource = version.AllResource()
	scaleOpdata.Instances = instances
	scaleOpdata.IsDown = isDown

	scaleTrans.OpData = &scaleOpdata

	go b.sched.RunScaleApplication(scaleTrans)

	app.LastStatus = app.Status
	app.Status = types.APP_STATUS_OPERATING
	app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
	app.UpdateTime = time.Now().Unix()
	app.Message = "application in scaling"
	if err := b.store.SaveApplication(app); err != nil {
		blog.Error("scale applicaiton(%s.%s) fail, save application to db err:%s", app.RunAs, app.ID, err.Error())
		return err
	}

	return nil
}
