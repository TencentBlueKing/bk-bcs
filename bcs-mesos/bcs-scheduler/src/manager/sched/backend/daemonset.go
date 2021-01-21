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
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/util"
)

//launch daemonset
func (b *backend) LaunchDaemonset(def *types.BcsDaemonsetDef) error {
	by, _ := json.Marshal(def.Version)
	blog.Infof("launch daemonset(%s.%s) version(%s)", def.NameSpace, def.Name, string(by))
	//lock
	util.Lock.Lock(types.BcsDaemonset{}, def.NameSpace+"."+def.Name)
	defer util.Lock.UnLock(types.BcsDaemonset{}, def.NameSpace+"."+def.Name)

	//check daemonset whether exists
	daemonset, err := b.store.FetchDaemonset(def.NameSpace, def.Name)
	if err != nil && err != store.ErrNoFound {
		blog.Errorf("Launch Daemonset(%s.%s), but FetchDaemonset failed: %s", def.NameSpace, def.Name, err.Error())
		return err
	}
	if daemonset != nil {
		err = fmt.Errorf("Daemonset(%s.%s) have exists", def.NameSpace, def.Name)
		blog.Warnf(err.Error())
		return err
	}
	//save version
	err = b.store.SaveVersion(def.Version)
	if err != nil {
		blog.Errorf("Launch Daemonset(%s.%s), but SaveVersion failed: %s", def.NameSpace, def.Name, err.Error())
		return err
	}
	blog.Infof("Launch Daemonset(%s.%s), and SaveVersion success", def.NameSpace, def.Name)
	//launch new daemonset
	daemonset = &types.BcsDaemonset{
		ObjectMeta: commtypes.ObjectMeta{
			Name:      def.Name,
			NameSpace: def.NameSpace,
		},
		Status:  types.Daemonset_Status_Starting,
		Created: time.Now().Unix(),
		Pods:    make(map[string]struct{}, 0),
	}
	err = b.store.SaveDaemonset(daemonset)
	if err != nil {
		blog.Errorf("Launch Daemonset(%s.%s), but SaveDaemonset failed: %s", def.NameSpace, def.Name, err.Error())
		return err
	}
	blog.Infof("Launch Daemonset(%s) success", daemonset.GetUuid())
	return nil
}

//delete daemonset
func (b *backend) DeleteDaemonset(namespace, name string, force bool) error {
	//lock
	util.Lock.Lock(types.BcsDaemonset{}, namespace+"."+name)
	defer util.Lock.UnLock(types.BcsDaemonset{}, namespace+"."+name)

	//check daemonset whether exists
	daemonset, err := b.store.FetchDaemonset(namespace, name)
	if err != nil && err != store.ErrNoFound {
		blog.Errorf("delete Daemonset(%s.%s), but FetchDaemonset failed: %s", namespace, name, err.Error())
		return err
	}
	if err == store.ErrNoFound {
		return nil
	}
	for podId := range daemonset.Pods {
		pod, err := b.store.FetchTaskGroup(podId)
		if err != nil {
			blog.Errorf("delete daemonset(%s:%s), but FetchTaskGroup(%s) failed:",
				daemonset.NameSpace, daemonset.Name, podId, err.Error())
			continue
		}
		if pod.Status == types.TASKGROUP_STATUS_FINISH || pod.Status == types.TASKGROUP_STATUS_FAIL {
			continue
		}
		//kill taskgroup in mesos cluster
		b.sched.KillTaskGroup(pod)
	}

	daemonset.ForceDeleting = force
	daemonset.LastStatus = daemonset.Status
	daemonset.Status = types.Daemonset_Status_Deleting
	err = b.store.SaveDaemonset(daemonset)
	if err != nil {
		blog.Errorf("Delete Daemonset(%s.%s), but SaveDaemonset failed: %s", namespace, name, err.Error())
		return err
	}
	blog.Infof("daemonset(%s) status from(%s)->to(%s)", daemonset.GetUuid(), daemonset.LastStatus, daemonset.Status)
	return nil
}
