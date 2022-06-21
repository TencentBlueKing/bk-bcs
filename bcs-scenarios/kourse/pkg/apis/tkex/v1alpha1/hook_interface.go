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

package v1alpha1

// ----------------------------------------- GameDeployment ----------------------------------------

// GetPreDeleteHook returns predelete hook spec
func (g *GameDeployment) GetPreDeleteHook() *HookStep {
	return g.Spec.PreDeleteUpdateStrategy.Hook
}

// GetPreDeleteHookConditions returns predelete hook conditions
func (g *GameDeploymentStatus) GetPreDeleteHookConditions() []PreDeleteHookCondition {
	return g.PreDeleteHookConditions
}

// SetPreDeleteHookConditions sets predelete hook conditions
func (g *GameDeploymentStatus) SetPreDeleteHookConditions(newConditions []PreDeleteHookCondition) {
	g.PreDeleteHookConditions = newConditions
}

// GetPreInplaceHook returns preinplace hook spec
func (g *GameDeployment) GetPreInplaceHook() *HookStep {
	return g.Spec.PreInplaceUpdateStrategy.Hook
}

// GetPreInplaceHookConditions returns preinplace hook conditions
func (g *GameDeploymentStatus) GetPreInplaceHookConditions() []PreInplaceHookCondition {
	return g.PreInplaceHookConditions
}

// SetPreInplaceHookConditions sets preinplace hook conditions
func (g *GameDeploymentStatus) SetPreInplaceHookConditions(newConditions []PreInplaceHookCondition) {
	g.PreInplaceHookConditions = newConditions
}

// GetPostInplaceHook returns post inplace hook spec
func (g *GameDeployment) GetPostInplaceHook() *HookStep {
	return g.Spec.PostInplaceUpdateStrategy.Hook
}

// GetPostInplaceHookConditions returns post inplace hook conditions
func (g *GameDeploymentStatus) GetPostInplaceHookConditions() []PostInplaceHookCondition {
	return g.PostInplaceHookConditions
}

// SetPostInplaceHookConditions set post inplace hook conditions
func (g *GameDeploymentStatus) SetPostInplaceHookConditions(newConditions []PostInplaceHookCondition) {
	g.PostInplaceHookConditions = newConditions
}

// ----------------------------------------- GameStatefulSet ----------------------------------------

// GetPreDeleteHook returns predelete hook spec
func (g *GameStatefulSet) GetPreDeleteHook() *HookStep {
	return g.Spec.PreDeleteUpdateStrategy.Hook
}

// GetPreDeleteHookConditions returns predelete hook conditions
func (g *GameStatefulSetStatus) GetPreDeleteHookConditions() []PreDeleteHookCondition {
	return g.PreDeleteHookConditions
}

// SetPreDeleteHookConditions sets predelete hook conditions
func (g *GameStatefulSetStatus) SetPreDeleteHookConditions(newConditions []PreDeleteHookCondition) {
	g.PreDeleteHookConditions = newConditions
}

// GetPreInplaceHook returns preinplace hook spec
func (g *GameStatefulSet) GetPreInplaceHook() *HookStep {
	return g.Spec.PreInplaceUpdateStrategy.Hook
}

// GetPreInplaceHookConditions returns preinplace hook conditions
func (g *GameStatefulSetStatus) GetPreInplaceHookConditions() []PreInplaceHookCondition {
	return g.PreInplaceHookConditions
}

// SetPreInplaceHookConditions sets preinplace hook conditions
func (g *GameStatefulSetStatus) SetPreInplaceHookConditions(newConditions []PreInplaceHookCondition) {
	g.PreInplaceHookConditions = newConditions
}

// GetPostInplaceHook returns post inplace hook spec
func (g *GameStatefulSet) GetPostInplaceHook() *HookStep {
	return g.Spec.PostInplaceUpdateStrategy.Hook
}

// GetPostInplaceHookConditions returns post inplace hook conditions
func (g *GameStatefulSetStatus) GetPostInplaceHookConditions() []PostInplaceHookCondition {
	return g.PostInplaceHookConditions
}

// SetPostInplaceHookConditions set post inplace hook conditions
func (g *GameStatefulSetStatus) SetPostInplaceHookConditions(newConditions []PostInplaceHookCondition) {
	g.PostInplaceHookConditions = newConditions
}
