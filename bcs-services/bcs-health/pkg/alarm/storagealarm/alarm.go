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

package storagealarm

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/master/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/alarm/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/register"
)

const (
	storageAlarm = "%s/bcsstorage/v1/alarms"
)

func NewStorageAlarm(c config.Config) (*StorageAlarm, error) {
	client := httpclient.NewHttpClient()
	if c.ClientCertFile != "" && c.ClientKeyFile != "" {
		if err := client.SetTlsVerity(c.CAFile, c.ClientCertFile, c.ClientKeyFile, static.ClientCertPwd); err != nil {
			return nil, err
		}
	}
	a := &StorageAlarm{
		silence: c.Silence,
		client:  client,
	}
	return a, nil
}

type StorageAlarm struct {
	silence bool
	client  *httpclient.HttpClient
}

type StorageResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (sa StorageAlarm) SendAlarm(op *utils.AlarmOptions, source string) error {
	if sa.silence {
		blog.Infof("run with silence mode, pretend to send alarm success. alarm: %+v", op)
		return nil
	}

	if err := sa.Validate(op); err != nil {
		return fmt.Errorf("validate alarm failed. err: %v", err)
	}

	// send the event.
	return sa.send(sa.Convert(op))
}

func (sa StorageAlarm) send(event *AlarmEvent) error {
	storageAddress, err := register.GetStorageServer()
	if err != nil {
		blog.Errorf("send storage alarm: failed to get storage server: %v. alarm: %+v", err, event)
		return err
	}

	storageAlarmData := types.BcsStorageAlarmIf{
		ClusterId:    event.Spec.ClusterID,
		Namespace:    event.Spec.Namespace,
		Module:       event.Spec.ModuleName,
		Source:       event.Spec.IP,
		ReceivedTime: event.Event.AtTime,
		Type:         "alarm",
		Data:         event,
	}

	var data []byte
	_ = codec.EncJson(storageAlarmData, &data)
	resp, err := sa.client.RequestEx(fmt.Sprintf(storageAlarm, storageAddress), http.MethodPost, nil, data)
	if err != nil {
		blog.Errorf("send storage alarm: failed to request to storage: %v", err)
		return err
	}

	var storageResp StorageResp
	if err = codec.DecJson(resp.Reply, &storageResp); err != nil {
		blog.Errorf("send storage alarm: failed to decode resp: %v", err)
		return err
	}

	if storageResp.Code != 0 {
		blog.Errorf("send storage alarm: failed to insert storage data: %s", storageResp.Message)
		return fmt.Errorf("failed to insert storage data: %s", storageResp.Message)
	}

	return nil
}

func (sa StorageAlarm) Validate(op *utils.AlarmOptions) error {
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

func (sa StorageAlarm) Convert(op *utils.AlarmOptions) *AlarmEvent {
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
