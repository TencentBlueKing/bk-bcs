/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package generator

import (
	"io"
	"testing"

	"github.com/helmfile/helmfile/pkg/helmexec"
	"go.uber.org/zap"
)

type helmVersionOnlyRunner struct{}

func (helmVersionOnlyRunner) Execute(_ string, args []string, _ map[string]string, _ bool) ([]byte, error) {
	if len(args) == 2 && args[0] == "version" && args[1] == "--short" {
		return []byte("v3.16.3"), nil
	}
	return nil, nil
}

func (helmVersionOnlyRunner) ExecuteStdIn(_ string, _ []string, _ map[string]string, _ io.Reader) ([]byte, error) {
	return nil, nil
}

func TestNewHelmExec_CreatesHelmExecutor(t *testing.T) {
	execer, err := newHelmExec("helm", zap.NewNop().Sugar(), "default", helmVersionOnlyRunner{})
	if err != nil {
		t.Fatalf("newHelmExec() error = %v", err)
	}
	if execer == nil {
		t.Fatal("newHelmExec() returned nil executor")
	}
	if !execer.IsHelm3() {
		t.Fatalf("expected Helm 3 executor, got version=%+v", execer.GetVersion())
	}
}

var _ helmexec.Runner = helmVersionOnlyRunner{}
