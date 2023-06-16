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

package logrule

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"k8s.io/klog/v2"

	bklog "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_log"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

const (
	defaultTimeout = time.Minute * 10
	month          = time.Hour * 24 * 30
)

// AutoCleanLogRule 自动删除已生成新规则的 bcslogconfigs，当新规则创建时间 > 30天并且新规则已产生日志数据，则执行自动清除操作。
func AutoCleanLogRule() {
	klog.Info("AutoCleanLogRule running")
	go func() {
		// run every hours
		for range time.Tick(time.Hour) {
			klog.Info("start to check logrules")
			runAutoCleanLogRule()
		}
	}()
}

func runAutoCleanLogRule() {
	store := storage.GlobalStorage
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	cond := operator.NewLeafCondition(operator.Ne, operator.M{
		entity.FieldKeyFromRuleID: "",
	})
	count, rul, err := store.ListLogRules(ctx, cond, &utils.ListOption{})
	if err != nil {
		klog.Errorf("list log rules error: %s", err.Error())
		return
	}
	klog.Infof("get %d log rules", count)
	for _, v := range rul {
		if !isBcsLogConfigID(v.FromRule) {
			continue
		}
		// create at 30 days ago
		if v.CreatedAt.After(time.Now().Add(-month)) {
			continue
		}
		// check rule has log
		if v.FileIndexSetID == 0 || v.STDIndexSetID == 0 {
			continue
		}
		var fileLog, stdLog bool
		fileLog, err = bklog.HasLog(ctx, v.FileIndexSetID)
		if err != nil {
			klog.Errorf("check log rule %s has file log in indexSet %d error: %s",
				v.Name, v.FileIndexSetID, err.Error())
			continue
		}
		stdLog, err = bklog.HasLog(ctx, v.STDIndexSetID)
		if err != nil {
			klog.Errorf("check log rule %s has std log in indexSet %d error: %s",
				v.Name, v.STDIndexSetID, err.Error())
			continue
		}
		if !fileLog && !stdLog {
			continue
		}
		// delete bcslogconfig
		ns, name := getBcsLogConfigNamespaces(v.FromRule)
		err = k8sclient.DeleteBcsLogConfig(ctx, v.ClusterID, ns, name)
		if err != nil {
			klog.Errorf("delete bcslogconfig error: %s", err.Error())
			continue
		}
		klog.Infof("deleted bcslogconfig %s", v.FromRule)
	}
}
