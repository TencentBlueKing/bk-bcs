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

package locallock

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	lockKey = "test"
)

func TestNewLocalLock(t *testing.T) {
	require.NotNil(t, NewLocalLock())
}

func TestKeyedMutex_Lock(t *testing.T) {
	l := NewLocalLock()
	require.Nil(t, l.Lock(context.Background(), lockKey))
	ch := make(chan error)
	go func() {
		ch <- l.Lock(context.Background(), lockKey)
	}()
	// check the second lock will be stuck
	select {
	case <-time.After(2 * time.Second):
		break
	case <-ch:
		require.NotNil(t, nil)
		break
	}
}

func TestKeyedMutex_UnLock(t *testing.T) {
	l := NewLocalLock()
	require.Nil(t, l.UnLock(context.Background(), lockKey))

	require.Nil(t, l.Lock(context.Background(), lockKey))
	require.Nil(t, l.UnLock(context.Background(), lockKey))
	ch := make(chan error)
	go func() {
		ch <- l.Lock(context.Background(), lockKey)
	}()
	// check the second lock will not be stuck
	select {
	case <-time.After(2 * time.Second):
		require.NotNil(t, nil)
		break
	case <-ch:
		break
	}
}
