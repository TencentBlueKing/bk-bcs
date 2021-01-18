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
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

// EventUpdateDescription udpate description in mongo event
type EventUpdateDescription struct {
	UpdatedFields bson.M   `bson:"updatedFields"`
	RemovedFields []string `bson:"removedFields"`
}

// EventNs ns fields in mongo event
type EventNs struct {
	DBName  string `bson:"db"`
	ColName string `bson:"col"`
}

// Event event in change stream
type Event struct {
	ID           bson.M                  `bson:"_id"`
	OpType       string                  `bson:"operationType"`
	FullDocument bson.M                  `bson:"fullDocument"`
	Ns           *EventNs                `bson:"ns"`
	DocumentKey  bson.M                  `bson:"documentKey"`
	UpdateDesc   *EventUpdateDescription `bson:"updateDescription"`
	ClusterTime  primitive.Timestamp     `bson:"clusterTime"`
	TxnNumber    int64                   `bson:"txnNumber"`
}

func convertMongoEvent(data Event) (*drivers.WatchEvent, error) {
	if data.Ns == nil {
		return nil, fmt.Errorf("miss ns in change stream event %+v", data)
	}
	var newEvent *drivers.WatchEvent
	switch data.OpType {
	case operationTypeInsert:
		newEvent = &drivers.WatchEvent{
			Type:           drivers.EventAdd,
			DBName:         data.Ns.DBName,
			CollectionName: data.Ns.ColName,
			Data:           operator.M(data.FullDocument),
			ClusterTime:    time.Unix(int64(data.ClusterTime.T), 0),
			TxnNumber:      data.TxnNumber,
		}
	case operationTypeUpdate:
		newEvent = &drivers.WatchEvent{
			Type:           drivers.EventUpdate,
			DBName:         data.Ns.DBName,
			CollectionName: data.Ns.ColName,
			Data:           operator.M(data.FullDocument),
			ClusterTime:    time.Unix(int64(data.ClusterTime.T), 0),
			TxnNumber:      data.TxnNumber,
		}
		if data.UpdateDesc != nil {
			newEvent.UpdatedFields = data.UpdateDesc.UpdatedFields
			newEvent.RemovedFields = data.UpdateDesc.RemovedFields
		}

	case operationTypeReplace:
		newEvent = &drivers.WatchEvent{
			Type:           drivers.EventUpdate,
			DBName:         data.Ns.DBName,
			CollectionName: data.Ns.ColName,
			Data:           operator.M(data.FullDocument),
			ClusterTime:    time.Unix(int64(data.ClusterTime.T), 0),
			TxnNumber:      data.TxnNumber,
		}

	case operationTypeDelete:
		newEvent = &drivers.WatchEvent{
			Type:           drivers.EventDelete,
			DBName:         data.Ns.DBName,
			CollectionName: data.Ns.ColName,
			Data:           operator.M(data.FullDocument),
			ClusterTime:    time.Unix(int64(data.ClusterTime.T), 0),
			TxnNumber:      data.TxnNumber,
		}

	default:
		blog.Errorf("unsupport event type %s", data.OpType)
		return nil, fmt.Errorf("unsupport event type %s", data.OpType)
	}
	return newEvent, nil
}
