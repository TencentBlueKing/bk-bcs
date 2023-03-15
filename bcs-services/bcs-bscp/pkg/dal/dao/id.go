/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/runtime/filter"
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
	sd *sharding.Sharding
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

	txn, err := ig.sd.Admin().DB().BeginTx(ctx.Ctx, new(sql.TxOptions))
	if err != nil {
		return nil, fmt.Errorf("gen %s unique id, but begin txn failed, err: %v", resource, err)
	}

	var sqlSentenceUp []string
	sqlSentenceUp = append(sqlSentenceUp, "UPDATE ", string(table.IDGeneratorTable), " SET max_id = max_id + ", strconv.Itoa(step),
		", updated_at = NOW()  WHERE resource = '", string(resource), "'")
	updateExpr := filter.SqlJoint(sqlSentenceUp)

	_, err = txn.ExecContext(ctx.Ctx, updateExpr)
	if err != nil {
		return nil, fmt.Errorf("gen %s unique id, but update max_id failed, err: %v", resource, err)
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT max_id FROM ", string(table.IDGeneratorTable), " WHERE resource = '", string(resource), "'")
	queryExpr := filter.SqlJoint(sqlSentence)

	rows, err := txn.QueryContext(ctx.Ctx, queryExpr)
	if err != nil {
		return nil, fmt.Errorf("gen %s unique id, but query max id failed, err: %v", resource, err)
	}
	defer rows.Close()

	gen := new(generator)
	for rows.Next() {
		if err := rows.Scan(&gen.MaxID); err != nil {
			return nil, fmt.Errorf("gen %s unique id, but scan max id failed, err: %v", resource, err)
		}
		// only one raw is queried, "resource" is a unique index key.
		break
	}

	if err := txn.Commit(); err != nil {
		return nil, fmt.Errorf("gen %s unique id, but commit failed, err: %v", resource, err)
	}

	// validate the max id is valid or not.
	if gen.MaxID < uint32(step) {
		return nil, fmt.Errorf("gen %s unique id, but got unexpected invalid max_id", resource)
	}

	// generate the id list that can be used.
	scope := gen.MaxID - uint32(step)
	list := make([]uint32, step)
	for id := 1; id <= int(step); id++ {
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
