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

package dao

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/sharding"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// IDGenInterface supplies all the method to generate a resource's
// unique identity id.
type IDGenInterface interface {
	// Batch return a list of resource's unique id as required.
	Batch(ctx *kit.Kit, resource table.Name, step int) ([]uint32, error)
	// One return one unique id for this resource.
	One(ctx *kit.Kit, resource table.Name) (uint32, error)
}

var _ IDGenInterface = new(idGenerator)

// NewIDGenerator create a id generator instance.
func NewIDGenerator(sd *sharding.Sharding) IDGenInterface {
	return &idGenerator{sd: sd}
}

type idGenerator struct {
	sd   *sharding.Sharding
	genQ *gen.Query
}

type generator struct {
	// MaxID record the last *already been used* resource id
	MaxID uint32 `db:"max_id"`
}

// Batch is to generate distribute unique resource id list.
// returned with a number of unique ids as required.
func (ig *idGenerator) Batch(ctx *kit.Kit, resource table.Name, step int) ([]uint32, error) {
	if err := resource.Validate(); err != nil {
		return nil, err
	}

	if step <= 0 {
		return nil, fmt.Errorf("gen %s unique id, but got invalid step", resource)
	}

	m := ig.genQ.IDGenerator
	genObj := new(generator)

	updateTx := func(tx *gen.Query) error {
		q := tx.IDGenerator.WithContext(ctx.Ctx)

		if _, err := q.Where(m.Resource.Eq(string(resource))).UpdateSimple(m.MaxID.Add(uint32(step))); err != nil {
			return err
		}

		obj, err := q.Where(m.Resource.Eq(string(resource))).Select(m.MaxID).Take()
		if err != nil {
			return err
		}

		genObj.MaxID = obj.MaxID
		return nil
	}

	if err := ig.genQ.Transaction(updateTx); err != nil {
		return nil, err
	}

	// validate the max id is valid or not.
	if genObj.MaxID < uint32(step) {
		return nil, fmt.Errorf("gen %s unique id, but got unexpected invalid max_id", resource)
	}

	// generate the id list that can be used.
	scope := genObj.MaxID - uint32(step)
	list := make([]uint32, step)
	for id := 1; id <= step; id++ {
		list[id-1] = scope + uint32(id)
	}
	return list, nil
}

// One generate one unique resource id.
func (ig *idGenerator) One(ctx *kit.Kit, resource table.Name) (uint32, error) {
	list, err := ig.Batch(ctx, resource, 1)
	if err != nil {
		return 0, err
	}

	if len(list) != 1 {
		return 0, errors.New("gen resource unique id, but got mismatched number of it ")
	}

	return list[0], nil
}
