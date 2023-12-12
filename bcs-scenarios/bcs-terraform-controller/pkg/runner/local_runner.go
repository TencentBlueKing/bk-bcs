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

package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/pkg/secret"
	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/option"
)

// LocalTerraform 本地terraform
type LocalTerraform interface {
	// Init local runner
	Init(req *NewTerraformRequest) (*NewTerraformReply, error)
	// GetPlanOutFile 获取plan输出文件
	GetPlanOutFile() string
	// SetPlanOutFile 设置plan输出文件
	SetPlanOutFile(file string)
	// GetLog 获取tf命令的执行日志
	GetLog(action string) string
	// GetInitLog 获取tf命令init的执行日志
	GetInitLog() string
	// GetPlanLog 获取tf命令plan的执行日志
	GetPlanLog() string
	// GetApplyLog 获取tf命令apply的执行日志
	GetApplyLog() string
	// GetDestroyLog 获取tf命令destroy的执行日志
	GetDestroyLog() string

	// ExecShowPlanFileRaw 获取原生日志
	ExecShowPlanFileRaw(ctx context.Context) (string, error)
	// ExecInit 初始化
	ExecInit(ctx context.Context, req *InitRequest) (*InitReply, error)
	// ExecPlan 计划
	ExecPlan(ctx context.Context, req *PlanRequest) (*PlanReply, error)
	// ExecApply 执行
	ExecApply(ctx context.Context, req *ApplyRequest) (*ApplyReply, error)
	// ExecDestroy 销毁
	ExecDestroy(ctx context.Context, req *DestroyRequest) (*DestroyReply, error)
}

// NewLocalTerraform new terraformLocalRunner
func NewLocalTerraform(cli client.Client, project string) (LocalTerraform, error) {
	sec, err := newSecretClient(cli, project)
	if err != nil {
		return nil, err
	}

	return &terraformLocalRunner{
		done:     make(chan os.Signal),
		logStore: make(map[string][]byte),
		execPath: option.TerraformBinPath,
		secret:   sec,
	}, nil
}

// terraformLocalRunner run terraform in operator local
type terraformLocalRunner struct {
	// instanceID 实例id
	instanceID string
	// done end
	done chan os.Signal
	// logStore 日志存储(init/plan/apply/destroy)
	logStore map[string][]byte
	// workDir 工作路径
	workDir string
	// planOutFile plan输出的文件
	planOutFile string

	// execPath pass form config, where is tfexec located
	execPath string
	// exec tf命令的对象
	exec *tfexec.Terraform
	// terraform target, 被执行的对象
	terraform *tfv1.Terraform
	// stdOut 收集tf日志(标准)
	stdOut *bytes.Buffer
	// stdErr 收集tf日志(错误)
	stdErr *bytes.Buffer

	// GlobalSecretOpt vaultPlugin opt
	secret secret.SecretManagerWithVersion
}

// Init return new terraform instance
func (t *terraformLocalRunner) Init(req *NewTerraformRequest) (*NewTerraformReply, error) {
	t.instanceID = req.InstanceID
	blog.Infof("creating new terraform, instance-id: %s, workingDir: %s, execPath: %s", t.instanceID,
		req.WorkingDir, req.ExecPath)

	exec, err := tfexec.NewTerraform(req.WorkingDir, t.execPath)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create new tf exec, ins-id: %s, work dir: %s, exec path: %s",
			t.instanceID, req.WorkingDir, req.ExecPath)
	}
	// hold only 1 instance
	t.exec = exec
	// cache the Terraform resource when initializing
	t.terraform = req.Terraform
	// set std
	t.stdOut = bytes.NewBuffer([]byte{})
	t.stdErr = bytes.NewBuffer([]byte{})
	// set work dir
	t.workDir = req.WorkingDir
	// 设置plan out file
	t.planOutFile = path.Join(t.workDir, fmt.Sprintf("main-%s.tfplan", t.instanceID))
	// set log
	t.initLogger()

	return &NewTerraformReply{Id: t.instanceID}, nil
}

// ExecInit do terraform init
func (t *terraformLocalRunner) ExecInit(ctx context.Context, req *InitRequest) (*InitReply, error) {
	defer t.outTfExecLog(InitAction)
	blog.Infof("terraform initializing, instance-id: %s", t.instanceID)
	if req.TfInstance != t.instanceID {
		return nil, errors.Errorf("no TF instance found, instance-id: %s", t.instanceID)
	}
	// provider aksk render
	if err := t.GenerateSecretForTF(ctx, t.workDir); err != nil {
		return nil, errors.Wrapf(err, "generate provider tf err")
	}

	initOpts := []tfexec.InitOption{
		tfexec.Upgrade(req.Upgrade),
		tfexec.ForceCopy(req.ForceCopy),
		option.GetConsulScheme(),
		option.GetConsulAddress(),
		option.GetConsulPath(t.terraform.Namespace, t.terraform.Name),
	}
	if err := t.exec.Init(ctx, initOpts...); err != nil {
		return nil, err
	}
	blog.Infof("terraform init success, instance-id: %s", t.instanceID)

	return &InitReply{Message: "ok"}, nil
}

// ExecPlan do terraform plan
func (t *terraformLocalRunner) ExecPlan(ctx context.Context, req *PlanRequest) (*PlanReply, error) {
	defer t.outTfExecLog(PlanAction)
	blog.Infof("creating a plan, instance-id: %s", t.instanceID)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-t.done:
			cancel()
		case <-ctx.Done():
		}
	}()
	if req.TfInstance != t.instanceID {
		return nil, errors.Errorf("no TF instance found, instance-id: %s", t.instanceID)
	}
	req.Out = t.planOutFile

	var planOpt []tfexec.PlanOption
	if req.Out != "" {
		planOpt = append(planOpt, tfexec.Out(req.Out))
	} else {
		// if backend is disabled completely, there will be no plan output file (req.Out = "")
		blog.Info("backend seems to be disabled completely, so there will be no plan output file")
	}

	if req.Refresh == false {
		planOpt = append(planOpt, tfexec.Refresh(req.Refresh))
	}

	if req.Destroy {
		planOpt = append(planOpt, tfexec.Destroy(req.Destroy))
	}

	if len(req.Targets) != 0 {
		for _, target := range req.Targets {
			planOpt = append(planOpt, tfexec.Target(target))
		}
	} else { // 如果为空，则尝试从对象中拿Targets
		for _, target := range t.terraform.Spec.Targets {
			planOpt = append(planOpt, tfexec.Target(target))
		}
	}

	drifted, err := t.exec.Plan(ctx, planOpt...)
	if err != nil {
		return nil, err
	}

	planCreated := false
	if req.Out != "" {
		planCreated = true

		//plan, err := t.exec.ShowPlanFile(ctx, req.Out)
		//if err != nil {
		//	return nil, err
		//}
		////blog.Infof("plan output: %s", utils.ToJsonString(plan))
		//
		//// This is the case when the plan is empty.
		//if plan.PlannedValues.Outputs == nil &&
		//	plan.PlannedValues.RootModule.Resources == nil &&
		//	plan.ResourceChanges == nil &&
		//	plan.PriorState == nil &&
		//	plan.OutputChanges == nil {
		//	planCreated = false
		//}

	}
	blog.Infof("terraform plan end up, instance-id: %s", t.instanceID)

	return &PlanReply{Message: "ok", Drifted: drifted, PlanCreated: planCreated}, nil
}

// ExecApply do terraform apply
func (t *terraformLocalRunner) ExecApply(ctx context.Context, req *ApplyRequest) (*ApplyReply, error) {
	defer t.outTfExecLog(ApplyAction)
	if len(req.DirOrPlan) == 0 {
		req.DirOrPlan = t.planOutFile
	}
	blog.Infof("running apply, instance-id: %s", t.instanceID)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-t.done:
			cancel()
		case <-ctx.Done():
		}
	}()
	if req.TfInstance != t.instanceID {
		return nil, errors.Errorf("no TF instance found, instance-id: %s", t.instanceID)
	}

	var applyOpt []tfexec.ApplyOption
	if req.DirOrPlan != "" {
		applyOpt = []tfexec.ApplyOption{tfexec.DirOrPlan(req.DirOrPlan)}
	}
	if req.RefreshBeforeApply {
		applyOpt = []tfexec.ApplyOption{tfexec.Refresh(true)}
	}
	if req.Parallelism > 0 {
		applyOpt = append(applyOpt, tfexec.Parallelism(int(req.Parallelism)))
	}
	for _, target := range req.Targets {
		applyOpt = append(applyOpt, tfexec.Target(target))
	}

	if err := t.exec.Apply(ctx, applyOpt...); err != nil {
		return nil, err
	}
	blog.Infof("run apply success, instance-id: %s", t.instanceID)

	return &ApplyReply{Message: "ok"}, nil
}

// ExecDestroy terraform destroy命令可以用来销毁并回收所有Terraform管理的基础设施资源。
// Terraform管理的资源会被销毁，在执行销毁动作前会通过交互式界面征求用户的确认。
//
//该命令可以接收所有apply命令的参数，除了不可以指定plan文件。
//
//如果-auto-approve参数被设置为true，那么将不会征求用户确认直接销毁。
//
//如果用-target参数指定了某项资源，那么不但会销毁该资源，同时也会销毁一切依赖于该资源的资源。我们会在随后介绍plan命令时详细介绍。
//
//terraform destroy将执行的所有操作都可以随时通过执行terraform plan -destroy命令来预览。
func (t *terraformLocalRunner) ExecDestroy(ctx context.Context, req *DestroyRequest) (*DestroyReply, error) {
	defer t.outTfExecLog(DestroyAction)
	blog.Infof("running destroy, instance-id: %s", t.instanceID)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-t.done:
			cancel()
		case <-ctx.Done():
		}
	}()
	if req.TfInstance != t.instanceID {
		return nil, errors.Errorf("no TF instance found, instance-id: %s", t.instanceID)
	}

	var destroyOpt []tfexec.DestroyOption
	for _, target := range req.Targets {
		destroyOpt = append(destroyOpt, tfexec.Target(target))
	}

	if err := t.exec.Destroy(ctx, destroyOpt...); err != nil {
		return nil, err
	}
	blog.Infof("run destroy success, instance-id: %s", t.instanceID)

	return &DestroyReply{Message: "ok"}, nil
}

// ExecShowPlanFileRaw 获取原生日志
func (t *terraformLocalRunner) ExecShowPlanFileRaw(ctx context.Context) (string, error) {
	defer func() {
		t.stdOut.Reset()
		t.stdErr.Reset()
	}()

	raw, err := t.exec.ShowPlanFileRaw(ctx, t.planOutFile)
	if err != nil {
		return "", errors.Wrapf(err, "show tf plan failed, tf: %s/%s, trace-id: %s",
			t.terraform.Namespace, t.terraform.Name, t.instanceID)
	}

	return raw, nil
}

// SetPlanOutFile 设置plan输出文件(用于指定apply执行)
func (t *terraformLocalRunner) SetPlanOutFile(file string) {
	t.planOutFile = file
}

// GetLog 获取tf命令的执行日志
func (t *terraformLocalRunner) GetLog(action string) string {
	logs, ok := t.logStore[action]
	if !ok {
		return ""
	}
	return string(logs)
}

// GetInitLog 获取tf命令init的执行日志
func (t *terraformLocalRunner) GetInitLog() string {
	return t.GetLog(InitAction)
}

// GetPlanLog 获取tf命令plan的执行日志
func (t *terraformLocalRunner) GetPlanLog() string {
	return t.GetLog(PlanAction)
}

// GetApplyLog 获取tf命令apply的执行日志
func (t *terraformLocalRunner) GetApplyLog() string {
	return t.GetLog(ApplyAction)
}

// GetDestroyLog 获取tf命令destroy的执行日志
func (t *terraformLocalRunner) GetDestroyLog() string {
	return t.GetLog(DestroyAction)
}

// GetPlanOutFile 获取plan输出文件
func (t *terraformLocalRunner) GetPlanOutFile() string {
	return t.planOutFile
}

// initLogger 全部采用blog，取消原来的log方式
func (t *terraformLocalRunner) initLogger() {
	// disable test logging
	if os.Getenv("DISABLE_TF_LOGS") == "1" {
		return
	}
	if os.Getenv("ENABLE_SENSITIVE_TF_LOGS") == "1" {
		t.exec.SetLogger(&localPrintfer{})
	}
	t.exec.SetStdout(t.stdOut)
	t.exec.SetStderr(t.stdErr)
}

// outTfExecLog 获取tf exec日志
func (t *terraformLocalRunner) outTfExecLog(action string) {
	result := make([]byte, 0)
	if t.stdOut.Len() != 0 {
		result = append(result, t.stdOut.Bytes()...)
		blog.Infof("tf client '%s' execute log:\n%s", action, t.stdOut.String())
		t.stdOut.Reset()
	}
	if t.stdErr.Len() != 0 {
		result = append(result, t.stdErr.Bytes()...)
		blog.Errorf("tf client '%s' execute log(error):\n%s", action, t.stdErr.String())
		t.stdErr.Reset()
	}
	t.logStore[action] = result
}

// newSecretClient 通过project的annotations获取相关的secret信息
func newSecretClient(cli client.Client, project string) (secret.SecretManagerWithVersion, error) {
	gcli := store.NewStore(option.GlobalGitopsOpt)
	if err := gcli.Init(); err != nil {
		return nil, errors.Wrapf(err, "init git client failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	argoProj, err := gcli.GetProject(ctx, project)
	if err != nil {
		return nil, errors.Wrapf(err, "get project failed, project: %s", project)
	}
	if argoProj == nil {
		return nil, errors.Errorf("project '%s' is nil", project)
	}
	annot := argoProj.GetAnnotations()
	if annot == nil {
		return nil, errors.Errorf("project '%s' annotation is empty", project)
	}

	val, ok := annot[common.SecretKey]
	if !ok {
		return nil, errors.Errorf("project '%s' not have secret key", project)
	}
	blog.Infof("query secret key '%s' project: %s", val, project)

	splitVal := strings.Split(val, ":")
	if len(splitVal) != 2 {
		return nil, errors.Errorf("project annotations format err '%s'", val)
	}
	ksecret := &v1.Secret{}
	nn := types.NamespacedName{
		Namespace: splitVal[0],
		Name:      splitVal[1],
	}
	if err := cli.Get(ctx, nn, ksecret); err != nil {
		return nil, errors.Wrapf(err, "get secret err")
	}

	opt := &options.Options{
		Secret: options.SecretOptions{
			CA:        option.GetVaultCaPath(),
			Type:      string(ksecret.Data["AVP_TYPE"]),
			Endpoints: string(ksecret.Data["VAULT_ADDR"]),
			Token:     string(ksecret.Data["VAULT_TOKEN"]),
		},
	}
	blog.Infof("get project[%s] secret[%s] success, addr[%s], token[%s], ca[%s]", project, nn.String(),
		opt.Secret.Endpoints, opt.Secret.Token, opt.Secret.CA)
	smwv := secret.NewSecretManager(opt)

	return smwv, nil
}
