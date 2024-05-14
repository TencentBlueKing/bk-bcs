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

// Package uuid NOTES
package uuid

import (
	"bytes"
	"sync"

	"github.com/google/uuid"
)

var uuidLock sync.Mutex
var lastUUID uuid.UUID

// UUID generates a UUID string which is version 4 or version 1.
func UUID() string {
	uuidLock.Lock()
	defer uuidLock.Unlock()

	// version 1
	result := uuid.Must(uuid.NewUUID())

	// The UUID package is naive and can generate identical UUIDs if the
	// time interval is quick enough.
	// The UUID uses 100 ns increments, so it's short enough to actively
	// wait for a new value.
	for bytes.Equal(lastUUID[:], result[:]) {
		result = uuid.Must(uuid.NewUUID())
	}

	lastUUID = result
	return result.String()
}
