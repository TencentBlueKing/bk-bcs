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

package tfhandler

import (
	"bytes"
	"context"
	"path"
	"time"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/internal/logctx"
	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/terraformextensions/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/tfhandler/tfparser"
)

// TerraformExec 执行实际的 Terraform 操作
type TerraformExec struct {
	repoPath   string
	workerPath string
	tf         *tfv1.Terraform

	stdoutBuffer bytes.Buffer
	stderrBuffer bytes.Buffer
	exec         *tfexec.Terraform

	parser tfparser.Interface
}

func newTerraformExec(tf *tfv1.Terraform, repoPath string) (*TerraformExec, error) {
	result := &TerraformExec{
		tf:       tf,
		repoPath: repoPath,
	}
	execPath := tf.Spec.Repository.Path
	result.workerPath = path.Join(repoPath, execPath)
	result.parser = tfparser.NewTerraformParser(tf.Spec.Project, result.workerPath)
	exec, err := tfexec.NewTerraform(result.workerPath, option.TerraformBinPath)
	if err != nil {
		return nil, errors.Wrapf(err, "create terraform client failed")
	}
	exec.SetStdout(&result.stdoutBuffer)
	exec.SetStderr(&result.stderrBuffer)
	result.exec = exec
	return result, nil
}

// ExecInit 执行 Init 动作，重写 Secret/检测 Consul Backend/注入 Consul 信息
func (t *TerraformExec) ExecInit(ctx context.Context) error {
	defer t.ResetStdBuffer()

	startTime := time.Now()
	if err := t.parser.RewriteSecret(ctx); err != nil {
		return errors.Wrapf(err, "rewrite secret failed")
	}
	if err := t.parser.CheckBackendConsul(); err != nil {
		return err
	}
	if err := t.exec.Init(ctx, []tfexec.InitOption{
		tfexec.Upgrade(false),
		tfexec.ForceCopy(true),
		option.GetConsulScheme(),
		option.GetConsulAddress(),
		option.GetConsulPath(t.tf.Namespace, t.tf.Name, string(t.tf.UID)),
	}...); err != nil {
		logctx.Errorf(ctx, "terraform init failed: stdout[%s], stderr[%s], cost[%v]",
			t.GetStdoutBuffer().String(), t.GetStdoutBuffer().String(), time.Now().Sub(startTime))
		return errors.Wrapf(err, "terraform init for '%s' failed", t.workerPath)
	}
	logctx.Infof(ctx, "terraform init success, cost[%v]", time.Now().Sub(startTime))
	return nil
}

// ExecPlan 进行实际的 Plan 动作
func (t *TerraformExec) ExecPlan(ctx context.Context) (string, bool, error) {
	defer t.ResetStdBuffer()

	startTime := time.Now()
	hasChanged, err := t.exec.Plan(ctx)
	if err != nil {
		logctx.Errorf(ctx, "terraform plan failed: stdout[%s], stderr[%s], cost[%v]",
			t.GetStdoutBuffer().String(), t.GetStdoutBuffer().String(), time.Now().Sub(startTime))
		return "", false, errors.Wrapf(err, "terraform plan for '%s' failed", t.workerPath)
	}
	logctx.Infof(ctx, "terraform plan success, changed[%v], cost[%v]", hasChanged, time.Now().Sub(startTime))
	if hasChanged {
		return t.GetStdoutBuffer().String(), true, nil
	}
	return "", hasChanged, nil
}

// ExecApply 进行实际的 Apply 动作
func (t *TerraformExec) ExecApply(ctx context.Context, destroy bool) (string, error) {
	defer t.ResetStdBuffer()

	startTime := time.Now()
	applyOps := make([]tfexec.ApplyOption, 0)
	if destroy == true {
		applyOps = append(applyOps, tfexec.Destroy(destroy))
	}
	if err := t.exec.Apply(ctx, applyOps...); err != nil {
		logctx.Errorf(ctx, "terraform apply failed: stdout[%s], stderr[%s], cost[%v]",
			t.GetStdoutBuffer().String(), t.GetStdoutBuffer().String(), time.Now().Sub(startTime))
		return "", errors.Wrapf(err, "terraform apply for '%s' failed", t.workerPath)
	}
	logctx.Infof(ctx, "terraform apply success, cost[%v]", time.Now().Sub(startTime))
	return t.GetStdoutBuffer().String(), nil
}

// ResetStdBuffer 重置标准输出
func (t *TerraformExec) ResetStdBuffer() {
	(&t.stdoutBuffer).Reset()
	(&t.stderrBuffer).Reset()
}

// GetStdoutBuffer 获取标准输出
func (t *TerraformExec) GetStdoutBuffer() *bytes.Buffer {
	return &t.stdoutBuffer
}

// GetStderrBuffer 获取标准错误输出
func (t *TerraformExec) GetStderrBuffer() *bytes.Buffer {
	return &t.stderrBuffer
}
