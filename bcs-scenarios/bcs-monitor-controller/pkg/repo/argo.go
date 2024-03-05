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

// Package repo xxx
package repo

import (
	"context"

	"github.com/argoproj/argo-cd/v2/util/db"
	settings_util "github.com/argoproj/argo-cd/v2/util/settings"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NewArgoDB create the DB of argocd
func NewArgoDB(ctx context.Context, adminNamespace string) (db.ArgoDB, *settings_util.SettingsManager, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "get in-cluster config failed")
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "create in-cluster client failed")
	}
	settingsMgr := settings_util.NewSettingsManager(ctx, kubeClient, adminNamespace)
	dbInstance := db.NewDB(adminNamespace, settingsMgr, kubeClient)
	return dbInstance, settingsMgr, nil
}
