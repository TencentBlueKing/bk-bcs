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
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-exec/tfexec"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/api/v1"
)

const loggerName = "runner.terraform"

// doto add interface

// TerraformLocalRunner run terraform in operator local
type TerraformLocalRunner struct {
	Cli client.Client
	// pass form config, where is tfexec located
	ExecPath string
	// Scheme     *runtime.Scheme
	Done       chan os.Signal
	tf         *tfexec.Terraform
	terraform  *tfv1.Terraform
	InstanceID string
}

// NewTerraform return new terraform instance
func (r *TerraformLocalRunner) NewTerraform(ctx context.Context, req *NewTerraformRequest) (*NewTerraformReply, error) {
	r.InstanceID = req.InstanceID
	log := ctrl.LoggerFrom(ctx, "instance-id", r.InstanceID).WithName(loggerName)
	log.Info("creating new terraform", "workingDir", req.WorkingDir, "execPath", req.ExecPath)
	tf, err := tfexec.NewTerraform(req.WorkingDir, r.ExecPath)
	if err != nil {
		log.Error(err, "unable to create new terraform", "workingDir", req.WorkingDir, "execPath", req.ExecPath)
		return nil, err
	}

	// hold only 1 instance
	r.tf = tf

	// cache the Terraform resource when initializing
	r.terraform = &req.Terraform

	// init default logger
	// r.initLogger(log)

	return &NewTerraformReply{Id: r.InstanceID}, nil
}

// Init do terraform init
func (r *TerraformLocalRunner) Init(ctx context.Context, req *InitRequest) (*InitReply, error) {
	log := ctrl.LoggerFrom(ctx, "instance-id", r.InstanceID).WithName(loggerName)
	log.Info("initializing")
	if req.TfInstance != r.InstanceID {
		err := fmt.Errorf("no TF instance found")
		log.Error(err, "no terraform")
		return nil, err
	}

	terraform := r.terraform

	log.Info("mapping the Spec.BackendConfigsFrom")
	backendConfigsOpts := []tfexec.InitOption{}
	for _, bf := range terraform.Spec.BackendConfigsFrom {
		objectKey := types.NamespacedName{
			Namespace: terraform.Namespace,
			Name:      bf.Name,
		}
		if bf.Kind == "Secret" {
			var s corev1.Secret
			err := r.Cli.Get(ctx, objectKey, &s)
			if err != nil && bf.Optional == false {
				log.Error(err, "unable to get object key", "objectKey", objectKey, "secret", s.ObjectMeta.Name)
				return nil, err
			}
			// if VarsKeys is null, use all
			if bf.Keys == nil {
				for key, val := range s.Data {
					backendConfigsOpts = append(backendConfigsOpts, tfexec.BackendConfig(key+"="+string(val)))
				}
			} else {
				for _, key := range bf.Keys {
					backendConfigsOpts = append(backendConfigsOpts, tfexec.BackendConfig(key+"="+string(s.Data[key])))
				}
			}
		} else if bf.Kind == "ConfigMap" {
			var cm corev1.ConfigMap
			err := r.Cli.Get(ctx, objectKey, &cm)
			if err != nil && bf.Optional == false {
				log.Error(err, "unable to get object key", "objectKey", objectKey, "configmap", cm.ObjectMeta.Name)
				return nil, err
			}

			// if Keys is null, use all
			if bf.Keys == nil {
				for key, val := range cm.Data {
					backendConfigsOpts = append(backendConfigsOpts, tfexec.BackendConfig(key+"="+val))
				}
				for key, val := range cm.BinaryData {
					backendConfigsOpts = append(backendConfigsOpts, tfexec.BackendConfig(key+"="+string(val)))
				}
			} else {
				for _, key := range bf.Keys {
					if val, ok := cm.Data[key]; ok {
						backendConfigsOpts = append(backendConfigsOpts, tfexec.BackendConfig(key+"="+val))
					}
					if val, ok := cm.BinaryData[key]; ok {
						backendConfigsOpts = append(backendConfigsOpts, tfexec.BackendConfig(key+"="+string(val)))
					}
				}
			}
		}
	}

	initOpts := []tfexec.InitOption{tfexec.Upgrade(req.Upgrade)}
	initOpts = append(initOpts, backendConfigsOpts...)
	if err := r.tf.Init(ctx, initOpts...); err != nil {
		return nil, err
	}

	return &InitReply{Message: "ok"}, nil
}

// Plan do terraform plan
func (r *TerraformLocalRunner) Plan(ctx context.Context, req *PlanRequest) (*PlanReply, error) {
	log := ctrl.LoggerFrom(ctx, "instance-id", r.InstanceID).WithName(loggerName)
	log.Info("creating a plan")
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-r.Done:
			cancel()
		case <-ctx.Done():
		}
	}()

	if req.TfInstance != r.InstanceID {
		err := fmt.Errorf("no TF instance found")
		log.Error(err, "no terraform")
		return nil, err
	}

	var planOpt []tfexec.PlanOption
	if req.Out != "" {
		planOpt = append(planOpt, tfexec.Out(req.Out))
	} else {
		// if backend is disabled completely, there will be no plan output file (req.Out = "")
		log.Info("backend seems to be disabled completely, so there will be no plan output file")
	}

	if req.Refresh == false {
		planOpt = append(planOpt, tfexec.Refresh(req.Refresh))
	}

	if req.Destroy {
		planOpt = append(planOpt, tfexec.Destroy(req.Destroy))
	}

	for _, target := range req.Targets {
		planOpt = append(planOpt, tfexec.Target(target))
	}

	drifted, err := r.tf.Plan(ctx, planOpt...)
	if err != nil {
		return nil, err
	}

	planCreated := false
	if req.Out != "" {
		planCreated = true

		plan, err := r.tf.ShowPlanFile(ctx, req.Out)
		if err != nil {
			return nil, err
		}

		// This is the case when the plan is empty.
		if plan.PlannedValues.Outputs == nil &&
			plan.PlannedValues.RootModule.Resources == nil &&
			plan.ResourceChanges == nil &&
			plan.PriorState == nil &&
			plan.OutputChanges == nil {
			planCreated = false
		}

	}

	return &PlanReply{Message: "ok", Drifted: drifted, PlanCreated: planCreated}, nil
}

// Apply do terraform apply
func (r *TerraformLocalRunner) Apply(ctx context.Context, req *ApplyRequest) (*ApplyReply, error) {
	log := ctrl.LoggerFrom(ctx, "instance-id", r.InstanceID).WithName(loggerName)
	log.Info("running apply")
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-r.Done:
			cancel()
		case <-ctx.Done():
		}
	}()

	if req.TfInstance != r.InstanceID {
		err := fmt.Errorf("no TF instance found")
		log.Error(err, "no terraform")
		return nil, err
	}

	var applyOpt []tfexec.ApplyOption
	if req.DirOrPlan != "" {
		applyOpt = []tfexec.ApplyOption{tfexec.DirOrPlan(req.DirOrPlan)}
	}

	if req.RefreshBeforeApply {
		applyOpt = []tfexec.ApplyOption{tfexec.Refresh(true)}
	}

	for _, target := range req.Targets {
		applyOpt = append(applyOpt, tfexec.Target(target))
	}

	if req.Parallelism > 0 {
		applyOpt = append(applyOpt, tfexec.Parallelism(int(req.Parallelism)))
	}

	if err := r.tf.Apply(ctx, applyOpt...); err != nil {
		return nil, err
	}

	return &ApplyReply{Message: "ok"}, nil
}

// Destroy do terraform destroy
func (r *TerraformLocalRunner) Destroy(ctx context.Context, req *DestroyRequest) (*DestroyReply, error) {
	log := ctrl.LoggerFrom(ctx, "instance-id", r.InstanceID).WithName(loggerName)
	log.Info("running destroy")
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-r.Done:
			cancel()
		case <-ctx.Done():
		}
	}()

	if req.TfInstance != r.InstanceID {
		err := fmt.Errorf("no TF instance found")
		log.Error(err, "no terraform")
		return nil, err
	}

	var destroyOpt []tfexec.DestroyOption
	for _, target := range req.Targets {
		destroyOpt = append(destroyOpt, tfexec.Target(target))
	}

	if err := r.tf.Destroy(ctx, destroyOpt...); err != nil {
		return nil, err
	}

	return &DestroyReply{Message: "ok"}, nil
}
