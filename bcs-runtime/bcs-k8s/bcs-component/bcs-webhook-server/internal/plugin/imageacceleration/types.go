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

package imageacceleration

import (
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginmanager"
)

const (
	// pluginName defines the name of image acceleration plugin
	pluginName = "imageacceleration"
	// configMapName defines the name of configmap that can enable image acceleration
	configMapName = "bcs-image-acceleration"
	// secretImagePullItem defines the image pull data of secret
	secretImagePullItem = ".dockerconfigjson" // nolint NOCC:gas/crypto(设计如此)

	// configMapKeyEnabled key of configmap, namespace will enable image acceleration if value is "true"
	configMapKeyEnabled = "enabled"
	// configMapKeyMapping key of configmap, users can use it to define custom registry mapping
	configMapKeyMapping = "mapping"

	// containerImageKey the image key path of pod
	containerImageKey = "/spec/containers/%d/image"

	// patchOpReplace the type of patch op that replace the key value
	patchOpReplace = "replace"
)

func init() {
	h := &Handler{}
	pluginmanager.Register(pluginName, h)
	mh := &Handler{}
	pluginmanager.RegisterMesos(pluginName, mh)
}
