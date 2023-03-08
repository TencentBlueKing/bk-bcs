/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package dao

import (
	"bytes"
	"fmt"

	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
)

// isResExist judge the existence of resources matched by table name and where expression.
func isResExist(kit *kit.Kit, orm orm.Interface, sd *sharding.One, tableName table.Name,
	whereExpr string) (bool, error) {

	buff := bytes.NewBuffer([]byte{})
	buff.WriteString(fmt.Sprintf("SELECT EXISTS(SELECT * FROM %s %s)", tableName, whereExpr))

	var result int8
	if err := orm.Do(sd.DB()).Get(kit.Ctx, &result, buff.String()); err != nil {
		return false, err
	}

	return result == 1, nil
}
