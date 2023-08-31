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

package nodetemplate

import (
	"context"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/nodetemplate"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

const (
	nodeTemplate = "nt"

	sopsCMTemplate = "CM"
)

// getAllNodeTemplates list all nodeTemplates
func getAllNodeTemplates(ctx context.Context, model store.ClusterManagerModel, projectID string) ([]proto.NodeTemplate, error) {
	condM := make(operator.M)
	condM[nodetemplate.ProjectIDKey] = projectID
	cond := operator.NewLeafCondition(operator.Eq, condM)

	return model.ListNodeTemplate(ctx, cond, &options.ListOption{})
}

func isSopsCMTemplateVars(str string) bool {
	if len(strings.Split(str, ".")) == 3 && strings.HasPrefix(str, sopsCMTemplate) {
		return true
	}

	return false
}
