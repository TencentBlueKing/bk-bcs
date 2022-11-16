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

// Package sharding NOTES
package sharding

import (
	"errors"
	"fmt"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/uuid"

	"github.com/jmoiron/sqlx"
)

// InitSharding initialize a sharding management instance.
func InitSharding(sd *cc.Sharding) (*Sharding, error) {

	db, err := connect(sd.AdminDatabase)
	if err != nil {
		return nil, err
	}

	s := &Sharding{
		one: &One{
			shardingUid: uuid.UUID(),
			db:          db,
		},
	}

	return s, nil
}

// Sharding is used to manage all the mysql instances
// which works for all the biz and admin resources.
type Sharding struct {
	// we support only one db just for now.
	// TODO: support sharding management later.
	one *One
	// db pool, like a connection pool
}

// MustSharding get a db instance with biz id.
// It does not check the biz's value, caller should to
// guarantee that biz is > 0; Otherwise, it will panic.
func (s *Sharding) MustSharding(biz uint32) *sqlx.DB {
	return s.one.db
}

// ShardingOne get a db instance with biz id.
func (s *Sharding) ShardingOne(biz uint32) *One {
	if biz <= 0 {
		return &One{hitErr: fmt.Errorf("invalid sharding one, because biz: %d is invalid", biz)}
	}

	return s.one
}

// Admin get the admin db instance
func (s *Sharding) Admin() *One {
	return s.one
}

// Audit get the audit db instance
func (s *Sharding) Audit() *One {
	return s.one
}

// Event get the event db instance
func (s *Sharding) Event() *One {
	return s.one
}

// Healthz check mysql healthz.
func (s *Sharding) Healthz() error {
	if err := s.one.db.Ping(); err != nil {
		return errors.New("mysql ping failed, err: " + err.Error())
	}

	return nil
}
