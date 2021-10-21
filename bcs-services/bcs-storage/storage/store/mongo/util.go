/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mongo

import (
	"encoding/json"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/types"
)

func decodeM(data operator.M, v interface{}) error {
	bytes, _ := json.Marshal(data)
	return json.Unmarshal(bytes, v)
}

func decodeMongoEvent(e *drivers.WatchEvent) (*store.Event, error) {
	rawObj := types.RawObject{}
	switch e.Type {
	case drivers.EventAdd:
		if err := decodeM(e.Data, &rawObj); err != nil {
			return nil, err
		}
		return &store.Event{
			Type: store.EventAdd,
			Obj:  &rawObj,
		}, nil
	case drivers.EventUpdate:
		if err := decodeM(e.Data, &rawObj); err != nil {
			return nil, err
		}
		return &store.Event{
			Type: store.EventUpdate,
			Obj:  &rawObj,
		}, nil

	case drivers.EventDelete:
		if err := decodeM(e.Data, &rawObj); err != nil {
			return nil, err
		}
		return &store.Event{
			Type: store.EventDelete,
			Obj:  &rawObj,
		}, nil

	case drivers.EventError:
		return &store.Event{
			Type: store.EventError,
		}, nil

	case drivers.EventClose:
		return &store.Event{
			Type: store.EventClose,
		}, nil
	default:
		return &store.Event{
			Type: store.EventError,
		}, nil
	}
}
