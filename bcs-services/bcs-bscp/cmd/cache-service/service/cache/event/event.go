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

// Package event NOTES
package event

import (
	"sync"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/bedis"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/dao"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
)

// nolint: unused
var stm *stream
var once = sync.Once{}

// Run the event watch task.
func Run(set dao.Set, state serviced.State, bds bedis.Client) error {
	s := &stream{
		ds:    daoSet{event: set.Event()},
		state: state,
	}

	s.cum = &consumer{
		bds: bds,
		op:  set,
	}

	s.lw = &loopWatch{
		ds:       s.ds,
		state:    s.state,
		consumer: s.cum.consume,
		mc:       initMetric(),
	}

	if err := s.lw.run(); err != nil {
		return err
	}

	s.pg = &purger{
		ds:    s.ds,
		state: s.state,
	}
	go s.pg.purge()

	once.Do(func() {
		stm = s
	})

	return nil
}

// daoSet defines all the DAO related operations which is needed by
// event handling.
type daoSet struct {
	event dao.Event
}

type stream struct {
	ds    daoSet
	state serviced.State
	lw    *loopWatch
	pg    *purger
	cum   *consumer
}
