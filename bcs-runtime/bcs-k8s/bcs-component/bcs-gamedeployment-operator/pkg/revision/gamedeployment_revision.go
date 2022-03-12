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

package revision

import (
	"encoding/json"

	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	gdcore "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/core"
	gdutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/util"

	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubernetes/pkg/controller/history"
)

var (
	patchCodec = scheme.Codecs.LegacyCodec(gdv1alpha1.SchemeGroupVersion)
)

// Interface is a interface to new and apply ControllerRevision.
type Interface interface {
	NewRevision(deploy *gdv1alpha1.GameDeployment, revision int64, collisionCount *int32) (*apps.ControllerRevision, error)
	ApplyRevision(deploy *gdv1alpha1.GameDeployment, revision *apps.ControllerRevision) (*gdv1alpha1.GameDeployment, error)
}

// NewRevisionControl create a normal revision control.
func NewRevisionControl() Interface {
	return &realControl{}
}

type realControl struct {
}

func (c *realControl) NewRevision(deploy *gdv1alpha1.GameDeployment, revision int64, collisionCount *int32) (*apps.ControllerRevision, error) {
	coreControl := gdcore.New(deploy)
	patch, err := c.getPatch(deploy, coreControl)
	if err != nil {
		return nil, err
	}
	cr, err := history.NewControllerRevision(deploy,
		gdutil.ControllerKind,
		deploy.Spec.Template.Labels,
		runtime.RawExtension{Raw: patch},
		revision,
		collisionCount)
	if err != nil {
		return nil, err
	}
	if cr.ObjectMeta.Annotations == nil {
		cr.ObjectMeta.Annotations = make(map[string]string)
	}
	for key, value := range deploy.Annotations {
		cr.ObjectMeta.Annotations[key] = value
	}
	cr.Namespace = deploy.Namespace
	return cr, nil
}

// getPatch returns a strategic merge patch that can be applied to restore a GameDeployment to a
// previous version. If the returned error is nil the patch is valid. The current state that we save is just the
// PodSpecTemplate. We can modify this later to encompass more state (or less) and remain compatible with previously
// recorded patches.
func (c *realControl) getPatch(deploy *gdv1alpha1.GameDeployment, coreControl gdcore.Control) ([]byte, error) {
	str, err := runtime.Encode(patchCodec, deploy)
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	_ = json.Unmarshal([]byte(str), &raw)
	objCopy := make(map[string]interface{})
	specCopy := make(map[string]interface{})
	spec := raw["spec"].(map[string]interface{})
	template := spec["template"].(map[string]interface{})

	coreControl.SetRevisionTemplate(specCopy, template)
	objCopy["spec"] = specCopy
	patch, err := json.Marshal(objCopy)
	return patch, err
}

func (c *realControl) ApplyRevision(deploy *gdv1alpha1.GameDeployment, revision *apps.ControllerRevision) (*gdv1alpha1.GameDeployment, error) {
	clone := deploy.DeepCopy()
	patched, err := strategicpatch.StrategicMergePatch([]byte(runtime.EncodeOrDie(patchCodec, clone)), revision.Data.Raw, clone)
	if err != nil {
		return nil, err
	}
	coreControl := gdcore.New(clone)
	return coreControl.ApplyRevisionPatch(patched)
}
