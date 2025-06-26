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

package tenant

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
)

var (
	model store.ProjectModel
)

func initDB() error {
	// mongo
	store.InitMongo(&config.MongoConfig{
		Address:        config.GlobalConf.Mongo.Address,
		Replicaset:     config.GlobalConf.Mongo.Replicaset,
		ConnectTimeout: config.GlobalConf.Mongo.ConnectTimeout,
		Database:       config.GlobalConf.Mongo.Database,
		Username:       config.GlobalConf.Mongo.Username,
		Password:       config.GlobalConf.Mongo.Password,
		MaxPoolSize:    config.GlobalConf.Mongo.MaxPoolSize,
		MinPoolSize:    config.GlobalConf.Mongo.MinPoolSize,
		Encrypted:      config.GlobalConf.Mongo.Encrypted,
	})
	model = store.New(store.GetMongo())
	return nil
}

func InitProject() error {
	err := initDB()
	if err != nil {
		return err
	}
	// 使用 model，在 porject 表中，把每个记录中，如果没有tenantId的记录，加上 tenantId，值为default
	ctx := context.Background()

	// 查询所有记录
	projects, _, err := model.ListProjects(ctx, nil, &page.Pagination{All: true})
	if err != nil {
		return fmt.Errorf("failed to list projects: %v", err)
	}

	// 更新每条记录
	for _, p := range projects {
		// 如果 TenantID 字段为空字符串，说明这个字段不存在
		if p.TenantID == "" {
			p.TenantID = "default"
			err = model.UpdateProject(ctx, &p)
			if err != nil {
				return fmt.Errorf("failed to update project %s: %v", p.ProjectID, err)
			}
		}
	}

	return nil
}
