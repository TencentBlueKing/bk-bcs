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

package bsalarm

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/master/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/alarm/utils"
)

func NewBlueShieldAlarm(c config.Config) (*BSAlarm, error) {
	kafka, err := NewKafkaClient(c.KafkaConf.DataID, c.KafkaConf.PluginPath, c.KafkaConf.ConfigFile, c.LogDir)
	if err != nil {
		return nil, fmt.Errorf("new kafka client failed, err: %v", err)
	}

	a := &BSAlarm{
		silence: c.Silence,
		client:  kafka,
	}
	return a, nil
}

type BSAlarm struct {
	silence bool
	client  *KafkaClient
}

func (b BSAlarm) SendAlarm(op *utils.AlarmOptions, source string) error {
	if b.silence {
		// for test only.
		blog.Infof("run with silence mode, pretend to send alarm success. alarm: %+v", op)
		return nil
	}

	if err := b.Validate(op); err != nil {
		return fmt.Errorf("validate alarm failed. err: %v", err)
	}

	// send the event.
	event := b.Convert(op)
	if err := b.client.SendEvent(event); err != nil {
		return fmt.Errorf("send event failed, uuid: %s, err: %v", event.Event.UUID, err)
	}
	return nil
}

func (b BSAlarm) Validate(op *utils.AlarmOptions) error {
	var errs []string
	if len(op.Module) == 0 {
		errs = append(errs, "module name is empty")
	}

	if len(op.AlarmName) == 0 {
		errs = append(errs, "alarm name is empty")
	}

	if len(op.EventMessage) == 0 {
		errs = append(errs, "alarm message is empty")
	}

	if len(op.ModuleIP) == 0 {
		errs = append(errs, "module ip is empty")
	}

	if len(op.UUID) == 0 {
		errs = append(errs, "uuid is empty")
	}

	if len(errs) != 0 {
		return errors.New(strings.Join(errs, ";"))
	}

	return nil
}

func (b BSAlarm) Convert(op *utils.AlarmOptions) *AlarmEvent {
	return &AlarmEvent{
		Meta: Meta{
			APIVersion: defaultVersion,
		},
		Spec: ModuleSpec{
			ModuleName:    op.Module,
			ClusterID:     op.ClusterID,
			Namespace:     op.Namespace,
			IP:            op.ModuleIP,
			ModuleVersion: op.ModuleVersion,
		},
		Event: Event{
			Reason:        op.AlarmName,
			Messages:      op.EventMessage,
			UUID:          op.UUID,
			AtTime:        op.AtTime,
			Affiliation:   op.Affiliation,
			AppAlarmLevel: op.AppAlarmLevel,
			ResourceType:  op.ResourceType,
			ResourceName:  op.ResourceName,
		},
		Extensions: EventExtension{
			Labels:  op.Labels,
			Context: "",
		},
	}
}
