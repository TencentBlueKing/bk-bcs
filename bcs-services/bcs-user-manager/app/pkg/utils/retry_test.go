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

package utils

import (
	"errors"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	// test timeout
	a := 1
	err := RetryWithTimeout(func() error {
		time.Sleep(time.Second)
		a = 2
		return nil
	}, RetryAttempts(3), RetryTimeout(10*time.Microsecond))
	if err == nil {
		t.Fatal("should be error")
	}

	// test normal
	a = 1
	err = RetryWithTimeout(func() error {
		a = 3
		return nil
	}, RetryAttempts(3), RetryTimeout(10*time.Second))
	if err != nil {
		t.Fatalf("got error, %s", err.Error())
	}
	if a != 3 {
		t.Fatalf("retry error")
	}

	// test error
	err = RetryWithTimeout(func() error {
		t.Log("exec")
		return errors.New("123")
	}, RetryAttempts(3), RetryTimeout(time.Second))
	if err == nil {
		t.Fatal("should be error")
	}
}
