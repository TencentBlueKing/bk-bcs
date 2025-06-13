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

package operation

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// mockAction mock action
type mockAction struct {
	name           string
	validateError  bool
	prepareError   bool
	executeError   bool
	executeTimeout bool
	gotError       error
}

var _ Operation = &mockAction{}

// Action xxx
func (r *mockAction) Action() string {
	return "mock"
}

// Name xxx
func (r *mockAction) Name() string {
	return r.name
}

// Validate xxx
func (r *mockAction) Validate() error {
	if r.validateError {
		return fmt.Errorf("validate error")
	}
	return nil
}

// Prepare xxx
func (r *mockAction) Prepare(ctx context.Context) error {
	if r.prepareError {
		return fmt.Errorf("prepare error")
	}
	return nil
}

// Execute xxx
func (r *mockAction) Execute(ctx context.Context) error {
	if r.executeError {
		return fmt.Errorf("execute error")
	}
	if r.executeTimeout {
		for { // nolint
			select {
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return nil
}

// Done xxx
func (r *mockAction) Done(err error) {
	r.gotError = err
}

func TestDispatch(t *testing.T) {
	blog.InitLogs(conf.LogConfig{Verbosity: 4, ToStdErr: true})
	defaultTimeout := 5 * time.Millisecond

	// test normal operate
	mock := &mockAction{name: "normal"}
	done, _ := GlobalOperator.Dispatch(mock, defaultTimeout)
	<-done
	if mock.gotError != nil {
		t.Error("normal shouldn't be error")
	}

	// test validate error
	mock = &mockAction{name: "test-validate", validateError: true}
	done, _ = GlobalOperator.Dispatch(mock, defaultTimeout)
	<-done
	if mock.gotError == nil {
		t.Error("validate should be error")
	}

	// test prepare error
	mock = &mockAction{name: "test-prepare", prepareError: true}
	done, _ = GlobalOperator.Dispatch(mock, defaultTimeout)
	<-done
	if mock.gotError == nil {
		t.Error("prepare should be error")
	}

	// test execute error
	mock = &mockAction{name: "test-execute", executeError: true}
	done, _ = GlobalOperator.Dispatch(mock, defaultTimeout)
	<-done
	if mock.gotError == nil {
		t.Error("execute should be error")
	}

	// test exec timeout
	mock = &mockAction{name: "test-exec-timeout", executeTimeout: true}
	done, _ = GlobalOperator.Dispatch(mock, defaultTimeout)
	<-done
	if mock.gotError == nil {
		t.Error("execute should be timeout")
	}

	// test terminate all operations
	mock = &mockAction{name: "test-terminate-1", executeTimeout: true}
	_, _ = GlobalOperator.Dispatch(mock, time.Second)
	mock2 := &mockAction{name: "test-terminate-2", executeTimeout: true}
	_, _ = GlobalOperator.Dispatch(mock2, time.Second)
	time.Sleep(time.Millisecond)
	GlobalOperator.TerminateOperation()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	GlobalOperator.WaitTerminate(ctx, time.Millisecond)
	if mock.gotError == nil {
		t.Error("execute should be terminated")
	}

	// test terminate
	mock = &mockAction{name: "test-terminate"}
	_, err := GlobalOperator.Dispatch(mock, defaultTimeout)
	if err == nil {
		t.Error("dispatch should be error")
	}
}
