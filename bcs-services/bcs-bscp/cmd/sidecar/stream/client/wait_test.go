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

package client

import (
	"context"
	"testing"
	"time"
)

func TestBlocker_WaitMS(t *testing.T) {
	bl := initBlocker()

	if !bl.TryBlock() {
		t.Errorf("should block success")
		return
	}

	unblockTimeMS := int64(1000)
	go func() {
		time.Sleep(time.Duration(unblockTimeMS) * time.Millisecond)
		bl.Unblock()
		t.Log("unblocked for now.")
	}()

	if err := bl.WaitMS(unblockTimeMS / 2); err == nil {
		t.Errorf("expected wait error, but not go error")
		return
	}

	t.Logf("test block timout success")

	if err := bl.WaitMS(0); err != nil {
		t.Errorf("unexpected wait error, err: %v", err)
		return
	}

	t.Logf("test block without timout success")
}

func TestBlocker_WaitWithContext(t *testing.T) {
	bl := initBlocker()

	if !bl.TryBlock() {
		t.Errorf("should block success")
		return
	}

	unblockTimeMS := int64(1000)
	go func() {
		time.Sleep(time.Duration(unblockTimeMS) * time.Millisecond)
		bl.Unblock()
		t.Log("unblocked for now.")
	}()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(unblockTimeMS/2)*time.Millisecond)
	if err := bl.WaitWithContext(ctx); err == nil {
		cancel()
		t.Errorf("expected wait error, but not go error")
		return
	}

	t.Logf("test block with context timout success")

	ctx, cancel = context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Millisecond)
		cancel()
	}()

	err := bl.WaitWithContext(ctx)
	if err == nil {
		t.Errorf("expected wait error because of context canceled, but not go error")
		return
	}

	t.Logf("expected block err: %v", err)
	t.Log("text block with context canceled success.")

	if err := bl.WaitWithContext(context.Background()); err != nil {
		t.Errorf("unexpected wait error, err: %v", err)
		return
	}

	t.Log("test block without context timeout success")

}
