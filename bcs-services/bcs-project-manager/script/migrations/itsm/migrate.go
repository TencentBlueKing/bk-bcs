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

// Package itsm 在 ITSM 注册服务，包括：创建命名空间、更新命名空间、删除命名空间, 允许重复执行
package itsm

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/itsm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	cm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/config"
)

var (
	model store.ProjectModel

	// WorkflowTemplates itsm templates
	//go:embed templates
	WorkflowTemplates embed.FS
)

// InitServices 初始化BCS相关流程服务
func InitServices() error {

	if err := initDB(); err != nil {
		fmt.Printf("init db failed, err: %s\n", err.Error())
		return err
	}

	if err := InitNamespaceITSMServices(); err != nil {
		fmt.Printf("init namespace itsm services failed, err: %s\n", err.Error())
		return err
	}

	return nil
}

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

// InitNamespaceITSMServices 初始化命名空间相关流程服务
func InitNamespaceITSMServices() error {
	// 2. create itsm catalog
	catalogID, err := createITSMCatalog()
	if err != nil {
		return err
	}
	// 3. import namespace services
	if err := importCreateNamespaceService(catalogID); err != nil {
		return err
	}
	if err := importUpdateNamespaceService(catalogID); err != nil {
		return err
	}
	if err := importDeleteNamespaceService(catalogID); err != nil {
		return err
	}
	return nil
}

func createITSMCatalog() (uint32, error) {
	catalogs, err := itsm.ListCatalogs()
	if err != nil {
		return 0, err
	}

	var rootID uint32
	var parentID uint32
	for _, rootCatalog := range catalogs {
		if rootCatalog.Key == "root" {
			rootID = rootCatalog.ID
			for _, parentCatalog := range rootCatalog.Children {
				if parentCatalog.Name == "蓝鲸容器管理平台" {
					parentID = parentCatalog.ID
					for _, catalog := range parentCatalog.Children {
						if catalog.Name == "命名空间" {
							return catalog.ID, nil
						}
					}
				}
			}
		}
	}
	if rootID == 0 {
		return 0, fmt.Errorf("root catalog not found")
	}
	if parentID == 0 {
		parentID, err = itsm.CreateCatalog(itsm.CreateCatalogReq{
			ProjectKey: "0",
			ParentID:   rootID,
			Name:       "蓝鲸容器管理平台",
			Desc:       "蓝鲸容器管理平台相关流程",
		})
		if err != nil {
			return 0, err
		}
	}
	// create namespace catalog
	catalogID, err := itsm.CreateCatalog(itsm.CreateCatalogReq{
		ProjectKey: "0",
		ParentID:   parentID,
		Name:       "命名空间",
		Desc:       "共享集群命名空间操作审批",
	})
	if err != nil {
		return 0, err
	}
	return catalogID, nil
}

func importCreateNamespaceService(catalogID uint32) error {
	// check whether the service has been imported before
	// if not, import it, else update it.
	serviceID, err := getServiceIDByName(catalogID, cm.ConfigCreateNamespaceItsmServiceName)
	if err != nil {
		return err
	}
	// 自定义模板分隔符为 [[ ]]，例如 [[ .Name ]]，避免和 ITSM 模板变量格式冲突
	tmpl, err := template.New("create_shared_namespace.json.tpl").Delims("[[", "]]").
		ParseFS(WorkflowTemplates, "templates/create_shared_namespace.json.tpl")
	if err != nil {
		return err
	}
	stringBuffer := &strings.Builder{}
	if err = tmpl.Execute(stringBuffer, map[string]string{
		"BCSGateway": config.GlobalConf.BcsGateway.Host,
		"BCSToken":   config.GlobalConf.BcsGateway.Token,
		"Approvers":  config.GlobalConf.ITSM.Approvers,
	}); err != nil {
		return err
	}
	mp := map[string]interface{}{}
	if err = json.Unmarshal([]byte(stringBuffer.String()), &mp); err != nil {
		return err
	}
	importReq := itsm.ImportServiceReq{
		Key:             "request",
		Name:            cm.ConfigCreateNamespaceItsmServiceName,
		Desc:            cm.ConfigCreateNamespaceItsmServiceName,
		CatelogID:       catalogID,
		Owners:          "admin",
		CanTicketAgency: false,
		IsValid:         true,
		DisplayType:     "OPEN",
		DisplayRole:     "",
		Source:          "custom",
		ProjectKey:      "0",
		Workflow:        mp,
	}
	if serviceID == 0 {
		logging.Info("service(name: %s) not found in itsm, creating", cm.ConfigCreateNamespaceItsmServiceName)
		serviceID, err = itsm.ImportService(importReq)
		if err != nil {
			return err
		}
	} else {
		logging.Info("service(name: %s, id: %d) found in itsm, updating",
			cm.ConfigCreateNamespaceItsmServiceName, serviceID)
		err = itsm.UpdateService(itsm.UpdateServiceReq{
			ID:               serviceID,
			ImportServiceReq: importReq,
		})
		if err != nil {
			return err
		}
	}
	return model.SetConfig(context.Background(), cm.ConfigKeyCreateNamespaceItsmServiceID,
		strconv.Itoa(serviceID))
}

func importUpdateNamespaceService(catalogID uint32) error {
	// check whether the service has been imported before
	// if not, import it, else update it.
	serviceID, err := getServiceIDByName(catalogID, cm.ConfigUpdateNamespaceItsmServiceName)
	if err != nil {
		return err
	}
	// 自定义模板分隔符为 [[ ]]，例如 [[ .Name ]]，避免和 ITSM 模板变量格式冲突
	tmpl, err := template.New("update_shared_namespace.json.tpl").Delims("[[", "]]").
		ParseFS(WorkflowTemplates, "templates/update_shared_namespace.json.tpl")
	if err != nil {
		return err
	}
	stringBuffer := &strings.Builder{}
	if err = tmpl.Execute(stringBuffer, map[string]string{
		"BCSGateway": config.GlobalConf.BcsGateway.Host,
		"BCSToken":   config.GlobalConf.BcsGateway.Token,
		"Approvers":  config.GlobalConf.ITSM.Approvers,
	}); err != nil {
		return err
	}
	mp := map[string]interface{}{}
	if err = json.Unmarshal([]byte(stringBuffer.String()), &mp); err != nil {
		return err
	}
	importReq := itsm.ImportServiceReq{
		Key:             "request",
		Name:            cm.ConfigUpdateNamespaceItsmServiceName,
		Desc:            cm.ConfigUpdateNamespaceItsmServiceName,
		CatelogID:       catalogID,
		Owners:          "admin",
		CanTicketAgency: false,
		IsValid:         true,
		DisplayType:     "OPEN",
		DisplayRole:     "",
		Source:          "custom",
		ProjectKey:      "0",
		Workflow:        mp,
	}
	if serviceID == 0 {
		logging.Info("service(name: %s) not found in itsm, creating", cm.ConfigUpdateNamespaceItsmServiceName)
		serviceID, err = itsm.ImportService(importReq)
		if err != nil {
			return err
		}
	} else {
		logging.Info("service(name: %s, id: %d) found in itsm, updating",
			cm.ConfigUpdateNamespaceItsmServiceName, serviceID)
		err = itsm.UpdateService(itsm.UpdateServiceReq{
			ID:               serviceID,
			ImportServiceReq: importReq,
		})
		if err != nil {
			return err
		}
	}
	return model.SetConfig(context.Background(), cm.ConfigKeyUpdateNamespaceItsmServiceID,
		strconv.Itoa(serviceID))
}

func importDeleteNamespaceService(catalogID uint32) error {
	// check whether the service has been imported before
	// if not, import it, else update it.
	serviceID, err := getServiceIDByName(catalogID, cm.ConfigDeleteNamespaceItsmServiceName)
	if err != nil {
		return err
	}
	// 自定义模板分隔符为 [[ ]]，例如 [[ .Name ]]，避免和 ITSM 模板变量格式冲突
	tmpl, err := template.New("delete_shared_namespace.json.tpl").Delims("[[", "]]").
		ParseFS(WorkflowTemplates, "templates/delete_shared_namespace.json.tpl")
	if err != nil {
		return err
	}
	stringBuffer := &strings.Builder{}
	if err = tmpl.Execute(stringBuffer, map[string]string{
		"BCSGateway": config.GlobalConf.BcsGateway.Host,
		"BCSToken":   config.GlobalConf.BcsGateway.Token,
		"Approvers":  config.GlobalConf.ITSM.Approvers,
	}); err != nil {
		return err
	}
	mp := map[string]interface{}{}
	if err = json.Unmarshal([]byte(stringBuffer.String()), &mp); err != nil {
		return err
	}
	importReq := itsm.ImportServiceReq{
		Key:             "request",
		Name:            cm.ConfigDeleteNamespaceItsmServiceName,
		Desc:            cm.ConfigDeleteNamespaceItsmServiceName,
		CatelogID:       catalogID,
		Owners:          "admin",
		CanTicketAgency: false,
		IsValid:         true,
		DisplayType:     "OPEN",
		DisplayRole:     "",
		Source:          "custom",
		ProjectKey:      "0",
		Workflow:        mp,
	}

	if serviceID == 0 {
		logging.Info("service(name: %s) not found in itsm, creating", cm.ConfigDeleteNamespaceItsmServiceName)
		serviceID, err = itsm.ImportService(importReq)
		if err != nil {
			return err
		}
	} else {
		logging.Info("service(name: %s, id: %d) found in itsm, updating",
			cm.ConfigDeleteNamespaceItsmServiceName, serviceID)
		err = itsm.UpdateService(itsm.UpdateServiceReq{
			ID:               serviceID,
			ImportServiceReq: importReq,
		})
		if err != nil {
			return err
		}
	}
	return model.SetConfig(context.Background(), cm.ConfigKeyDeleteNamespaceItsmServiceID,
		strconv.Itoa(serviceID))
}

func getServiceIDByName(catalogID uint32, name string) (int, error) {
	services, err := itsm.ListServices(catalogID)
	if err != nil {
		return 0, err
	}
	for _, service := range services {
		if service.Name == name {
			return service.ID, nil
		}
	}
	return 0, nil
}
