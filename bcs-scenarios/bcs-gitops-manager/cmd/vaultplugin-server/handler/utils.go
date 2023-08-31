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

package handler

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
)

func getGitopsAppSecretkey(
	ctx context.Context,
	st store.Store,
	application string,
) (secretname string, projectname string, err error) {
	app, err := st.GetApplication(ctx, application)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get gitops application")
	}
	project, err := st.GetProject(ctx, app.Spec.Project)
	if err != nil {
		return "", project.Name, errors.Wrap(err, "failed to get gitops project")
	}
	secretKey, ok := project.GetAnnotations()[common.SecretKey]
	if !ok {
		return "", project.Name, fmt.Errorf("get secret key error, secretkey is empty")
	}
	return secretKey, project.Name, nil
}
