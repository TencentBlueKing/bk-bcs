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
 */

package mongo

import (
	"context"
	"errors"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	datamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

// ModelOperationData model for bcs user operation data
type ModelOperationData struct {
	// Tables map[tableIndex]Public
	Tables map[string]*Public
}

// NewModelOperationData new bcs user operation data model
func NewModelOperationData(db drivers.DB, bkbaseConf *types.BkbaseConfig) *ModelOperationData {
	return &ModelOperationData{}
}

// GetUserOperationDataList get operation data for bcs user
func (pt *ModelOperationData) GetUserOperationDataList(ctx context.Context,
	request *datamanager.GetUserOperationDataListRequest) ([]*structpb.Struct, int64, error) {
	return nil, 0, errors.New("Not implemented by mongo store")
}
