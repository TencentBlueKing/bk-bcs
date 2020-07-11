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

package alarm

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/bcs-health/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/statistic"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/master/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/alarm/bsalarm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/alarm/storagealarm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/alarm/utils"
)

func NewAlarmProxy(c config.Config) (utils.AlarmFactory, error) {
	proxy := &AlarmProxy{
		conf: c,
	}

	if c.EnableBsAlarm {
		bsAlarm, err := bsalarm.NewBlueShieldAlarm(c)
		if err != nil {
			return nil, fmt.Errorf("new blueshield alarm failed, err: %v", err)
		}
		proxy.bsAlarm = bsAlarm
	} else {
		blog.Info("blue shield alarm no enable, will not be init.")
	}

	if c.EnableStorageAlarm {
		storageAlarm, err := storagealarm.NewStorageAlarm(c)
		if err != nil {
			return nil, fmt.Errorf("new storage alarm failed, err: %v", err)
		}
		proxy.storageAlarm = storageAlarm
	} else {
		blog.Infof("storage alarm no enable, will not be init")
	}

	return proxy, nil
}

type AlarmProxy struct {
	bsAlarm      utils.AlarmFactory
	storageAlarm utils.AlarmFactory
	conf         config.Config
}

func (ap *AlarmProxy) SendAlarm(op *utils.AlarmOptions, source string) (err error) {
	statistic.IncAccess()

	if len(op.AppAlarmLevel) == 0 {
		op.AppAlarmLevel = "important"
		blog.Warnf("alarm uuid[%s] level is empty, use default *important* as its value.", op.UUID)
	}

	if len(op.Affiliation) == 0 {
		op.Affiliation = types.Both
		blog.Warnf("alarm uuid[%s] affiliation is empty, use default *both* as its value.", op.UUID)
	}

	if op.Labels == nil {
		op.Labels = make(map[string]string)
	}

	op.Labels[utils.DataIDLabelKey] = ap.conf.KafkaConf.DataID

	if ap.conf.EnableBsAlarm && ap.bsAlarm != nil {
		if err = ap.bsAlarm.SendAlarm(op, source); err != nil {
			blog.Errorf("send alarm uuid[%s] to blueshield failed, err: %v", op.UUID, err)
			return err
		}
	}

	if ap.conf.EnableStorageAlarm {
		if err = ap.storageAlarm.SendAlarm(op, source); err != nil {
			blog.Errorf("send alarm uuid[%s] to storage failed, err: %v", op.UUID, err)
			return err
		}
	}

	if ap.conf.EnableLogAlarm {
		var alarmData []byte
		_ = codec.EncJson(op, &alarmData)
		blog.Info("bcs-health get alarm(source:%s): %s", source, string(alarmData))
	}

	return nil
}
