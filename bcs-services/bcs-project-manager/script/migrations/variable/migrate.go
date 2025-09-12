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

/*
迁移变量数据到 bcs-project-manager
允许重复执行
*/

// Package main variable
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	vvm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variablevalue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
)

const (
	saasVariableTableName          = "variable_variable"
	saasClusterVariableTableName   = "variable_clustervariable"
	saasNamespaceVariableTableName = "variable_namespacevariable"
	paasCcNamespaceTableName       = "kubernetes_namespaces"
	timeLayout                     = "2006-01-02T15:04:05Z"
)

var (
	// saas db config
	saasDBHost string
	saasDBPort uint
	saasDBUser string
	saasDBPwd  string
	saasDBName string

	// bcs-cc db config
	ccDBHost string
	ccDBPort uint
	ccDBUser string
	ccDBPwd  string
	ccDBName string

	// bcs mongodb config
	mongoAddr       string
	mongoReplicaset string
	mongoUser       string
	mongoPwd        string
	mongoAuthDBName string
	mongoDBName     string

	// db instance
	ccDB   *gorm.DB
	saasDB *gorm.DB
	model  store.ProjectModel
)

func parseFlags() {
	// mysql for saas
	flag.StringVar(&saasDBHost, "saas_db_host", "", "mysql host")
	flag.UintVar(&saasDBPort, "saas_db_port", 0, "mysql port")
	flag.StringVar(&saasDBUser, "saas_db_user", "", "access mysql username")
	flag.StringVar(&saasDBPwd, "saas_db_pwd", "", "access mysql password")
	flag.StringVar(&saasDBName, "saas_db_name", "", "access mysql db name")

	// mysql for paas-cc
	flag.StringVar(&ccDBHost, "cc_db_host", "", "mysql host")
	flag.UintVar(&ccDBPort, "cc_db_port", 0, "mysql port")
	flag.StringVar(&ccDBUser, "cc_db_user", "", "access mysql username")
	flag.StringVar(&ccDBPwd, "cc_db_pwd", "", "access mysql password")
	flag.StringVar(&ccDBName, "cc_db_name", "", "access mysql db name")

	// mongo for bcs-project-manager
	flag.StringVar(&mongoAddr, "mongo_addr", "", "mongo address")
	flag.StringVar(&mongoReplicaset, "mongo_replicaset", "", "mongo replicaset")
	flag.StringVar(&mongoUser, "mongo_user", "", "access mongo username")
	flag.StringVar(&mongoPwd, "mongo_pwd", "", "access mongo password")
	flag.StringVar(&mongoAuthDBName, "mongo_auth_db_name", "", "access mongo db name")
	flag.StringVar(&mongoDBName, "mongo_db_name", "", "access mongo db name")

	flag.Parse()
}

func main() {
	parseFlags()
	if err := initDB(); err != nil {
		fmt.Printf("init db failed, err: %s\n", err.Error())
		return
	}
	fmt.Println("migrate start ...")
	if err := migrateVariableDefinition(); err != nil {
		fmt.Printf("migrate variable definitions failed, err: %s\n", err.Error())
		return
	}
	fmt.Println("migrate variable definitions success!")
	if err := migrateClusterVariables(); err != nil {
		fmt.Printf("migrate cluster variables failed, err: %s\n", err.Error())
		return
	}
	if err := migrateNamespaceVariables(); err != nil {
		fmt.Printf("migrate namespace variables failed, err: %s\n", err.Error())
		return
	}
	fmt.Println("migrate variable values success!")
	fmt.Println("migrate success!")
}

// SaasVariableDefinition ...
type SaasVariableDefinition struct {
	ID          uint       `json:"id"`
	Creator     string     `json:"creator" gorm:"size:64"`
	Updator     string     `json:"updator" gorm:"size:64"`
	Created     *time.Time `json:"created"`
	Updated     *time.Time `json:"updated"`
	IsDeleted   bool       `json:"is_deleted"`
	DeletedTime *time.Time `json:"deleted_time"`
	ProjectID   string     `json:"project_id" gorm:"size:32;unique;index"`
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Default     string     `json:"default"`
	Desc        string     `json:"desc"`
	Category    string     `json:"category"`
	Scope       string     `json:"scope"`
}

// TableName overrides the table name
func (SaasVariableDefinition) TableName() string {
	return saasVariableTableName
}

// SaasClusterVariable ...
type SaasClusterVariable struct {
	ID          uint       `json:"id"`
	Creator     string     `json:"creator" gorm:"size:64"`
	Updator     string     `json:"updator" gorm:"size:64"`
	Created     *time.Time `json:"created"`
	Updated     *time.Time `json:"updated"`
	IsDeleted   bool       `json:"is_deleted"`
	DeletedTime *time.Time `json:"deleted_time"`
	VarID       uint       `json:"var_id"`
	ClusterID   string     `json:"cluster_id"`
	Data        string     `json:"data" gorm:"type:longtext"`
}

// SaasNamespaceVariable ...
type SaasNamespaceVariable struct {
	ID          uint       `json:"id"`
	Creator     string     `json:"creator" gorm:"size:64"`
	Updator     string     `json:"updator" gorm:"size:64"`
	Created     *time.Time `json:"created"`
	Updated     *time.Time `json:"updated"`
	IsDeleted   bool       `json:"is_deleted"`
	DeletedTime *time.Time `json:"deleted_time"`
	VarID       uint       `json:"var_id"`
	NsID        uint       `json:"ns_id"`
	Data        string     `json:"data" gorm:"type:longtext"`
}

// PaasccNamespace ...
type PaasccNamespace struct {
	ID             uint       `json:"id"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at"`
	Name           string     `json:"name"`
	Creator        string     `json:"creator"`
	Description    string     `json:"description"`
	ProjectID      string     `json:"project_id"`
	ClusterID      string     `json:"cluster_id"`
	EnvType        string     `json:"env_type"`
	Status         string     `json:"status"`
	HasImageSecret bool       `json:"has_image_secret"`
}

// TableName overrides the table name
func (PaasccNamespace) TableName() string {
	return paasCcNamespaceTableName
}

// SaasVariableValue ...
type SaasVariableValue struct {
	Value string `json:"value"`
}

func initDB() error {
	var err error
	// mysql for saas
	saasDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		saasDBUser, saasDBPwd, saasDBHost, saasDBPort, saasDBName)
	saasDB, err = gorm.Open("mysql", saasDSN)
	if err != nil {
		return fmt.Errorf("access mysql error, %s", err.Error())
	}
	saasDB.DB().SetConnMaxLifetime(10 * time.Second)
	saasDB.DB().SetMaxIdleConns(20)
	saasDB.DB().SetMaxOpenConns(20)

	// mysql for paas-cc
	ccDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		ccDBUser, ccDBPwd, ccDBHost, ccDBPort, ccDBName)
	ccDB, err = gorm.Open("mysql", ccDSN)
	if err != nil {
		return fmt.Errorf("access mysql error, %s", err.Error())
	}
	ccDB.DB().SetConnMaxLifetime(10 * time.Second)
	ccDB.DB().SetMaxIdleConns(20)
	ccDB.DB().SetMaxOpenConns(20)

	// mongo
	store.InitMongo(&config.MongoConfig{
		Address:        mongoAddr,
		ConnectTimeout: 5,
		AuthDatabase: mongoAuthDBName,
		Database:       mongoDBName,
		Username:       mongoUser,
		Password:       mongoPwd,
		MaxPoolSize:    10,
		MinPoolSize:    1,
		Encrypted:      false,
	})
	model = store.New(store.GetMongo())
	return nil
}

// 迁移变量定义数据
func migrateVariableDefinition() error {
	variables, err := fetchSaasVariableDefinitions()
	if err != nil {
		fmt.Printf("fetch BCS CC variable definitions failed, err: %s\n", err.Error())
		return err
	}
	// 组装数据
	for _, variable := range variables {
		project, err := model.GetProject(context.Background(), variable.ProjectID)
		if err != nil {
			fmt.Printf("get project %s failed, err: %s\n", variable.ProjectID, err.Error())
			continue
		}
		// 跳过已删除的变量
		if variable.IsDeleted {
			continue
		}
		vd := &vdm.VariableDefinition{
			Key:         variable.Key,
			Name:        variable.Name,
			Description: variable.Desc,
			ProjectCode: project.ProjectCode,
			Scope:       variable.Scope,
			Category:    variable.Category,
			Creator:     variable.Creator,
			Updater:     variable.Updator,
			IsDeleted:   variable.IsDeleted,
		}
		temp := &SaasVariableValue{}
		if err := json.Unmarshal([]byte(variable.Default), temp); err != nil {
			fmt.Printf("unmarshal default value for variable key [%s] in project [%s] failed, err: %s\n",
				variable.Key, project.ProjectCode, err.Error())
			// 跳过解析失败的变量
			continue
		}
		vd.Default = temp.Value
		if variable.Created != nil {
			vd.CreateTime = variable.Created.Format(timeLayout)
		}
		if variable.Updated != nil {
			vd.UpdateTime = variable.Updated.Format(timeLayout)
		}
		if err := doUpsertVariableDefinition(model, vd); err != nil {
			fmt.Printf("upsert variable definition key [%s] in project [%s] failed, err: %s\n",
				vd.Key, project.Name, err.Error())
			continue
		}
	}
	return nil
}

func migrateClusterVariables() error {
	// migrate cluster variables
	clusterVariables, err := fetchSaasClusterVariables()
	if err != nil {
		fmt.Printf("fetch BCS CC cluster variables failed, err: %s\n", err.Error())
		return err
	}
	fmt.Printf("found %d cluster variables need to migrate\n", len(clusterVariables))

	// 组装数据
	for _, variable := range clusterVariables {
		orgninalDef := &SaasVariableDefinition{}
		if result := saasDB.Table(saasVariableTableName).First(&orgninalDef, variable.VarID); result.Error != nil {
			fmt.Printf("get variable definition [%d] failed, err: %s\n", variable.VarID, result.Error.Error())
			continue
		}
		// 跳过已删除的变量
		if orgninalDef.IsDeleted {
			continue
		}
		project, err := model.GetProject(context.Background(), orgninalDef.ProjectID)
		if err != nil {
			fmt.Printf("get project [%s] failed, err: %s\n", orgninalDef.ProjectID, err.Error())
			continue
		}
		// exclude mesos variables
		if project.Kind == "mesos" {
			continue
		}
		newDef, err := model.GetVariableDefinitionByKey(context.Background(), project.ProjectCode, orgninalDef.Key)
		if err != nil {
			fmt.Printf("get new variable definition by projectCode [%s] key [%s] failed, err: %s\n",
				project.ProjectCode, orgninalDef.Key, err.Error())
			continue
		}
		value := &vvm.VariableValue{}
		value.VariableID = newDef.ID
		value.Scope = vdm.VariableScopeCluster
		value.ClusterID = variable.ClusterID
		temp := &SaasVariableValue{}
		if err := json.Unmarshal([]byte(variable.Data), temp); err != nil {
			fmt.Printf("unmarshal cluster value for variable [%s] in cluster [%s] failed, err: %s\n",
				value.VariableID, value.ClusterID, err.Error())
			continue
		}
		value.Value = temp.Value
		if variable.Created != nil {
			value.CreateTime = variable.Created.Format(timeLayout)
		}
		if variable.Updated != nil {
			value.UpdateTime = variable.Updated.Format(timeLayout)
		}
		if variable.Updator != "" {
			value.Updater = variable.Updator
		}
		if err := model.UpsertVariableValue(context.Background(), value); err != nil {
			fmt.Printf("upsert cluster value for variable [%s] in cluster [%s] failed, err: %s\n",
				value.VariableID, value.ClusterID, err.Error())
			continue
		}
	}
	return nil
}

func migrateNamespaceVariables() error {
	// migrate namespace variables
	namespaceVariables, err := fetchSaasNamespaceVariables()
	if err != nil {
		fmt.Printf("fetch BCS CC cluster variables failed, err: %s\n", err.Error())
		return err
	}
	fmt.Printf("found %d namespace variables need to migrate\n", len(namespaceVariables))

	// 组装数据
	for _, variable := range namespaceVariables {
		orgninalDef := &SaasVariableDefinition{}
		if result := saasDB.Table(saasVariableTableName).First(&orgninalDef, variable.VarID); result.Error != nil {
			fmt.Printf("get variable definition [%d] failed, err: %s\n", variable.VarID, result.Error.Error())
			continue
		}
		// 跳过已删除的变量
		if orgninalDef.IsDeleted {
			continue
		}
		project, err := model.GetProject(context.Background(), orgninalDef.ProjectID)
		if err != nil {
			fmt.Printf("get project [%s] failed, err: %s\n", orgninalDef.ProjectID, err.Error())
			continue
		}
		// exclude mesos variables
		if project.Kind == "mesos" {
			continue
		}
		newDef, err := model.GetVariableDefinitionByKey(context.Background(), project.ProjectCode, orgninalDef.Key)
		if err != nil {
			fmt.Printf("get new variable definition by projectCode [%s] key [%s] failed, err: %s\n",
				project.ProjectCode, orgninalDef.Key, err.Error())
			continue
		}
		namespace := &PaasccNamespace{}
		if result := ccDB.First(namespace, variable.NsID); result.Error != nil {
			fmt.Printf("get ns for variable value [%d] by ns_id [%d] failed, err: %s\n",
				variable.ID, variable.NsID, result.Error.Error())
			continue
		}
		value := &vvm.VariableValue{}
		value.VariableID = newDef.ID
		value.Scope = vdm.VariableScopeNamespace
		value.ClusterID = namespace.ClusterID
		value.Namespace = namespace.Name
		temp := &SaasVariableValue{}
		if err := json.Unmarshal([]byte(variable.Data), temp); err != nil {
			fmt.Printf("unmarshal cluster value for variable [%s] in cluster [%s] failed, err: %s\n",
				value.VariableID, value.ClusterID, err.Error())
			continue
		}
		value.Value = temp.Value
		if variable.Created != nil {
			value.CreateTime = variable.Created.Format(timeLayout)
		}
		if variable.Updated != nil {
			value.UpdateTime = variable.Updated.Format(timeLayout)
		}
		if err := model.UpsertVariableValue(context.Background(), value); err != nil {
			fmt.Printf("upsert cluster value for variable [%s] in cluster [%s] failed, err: %s\n",
				value.VariableID, value.ClusterID, err.Error())
			continue
		}
	}
	return nil
}

// fetchSaasVariableDefinitions 拉取 saas 中所有的变量定义
func fetchSaasVariableDefinitions() ([]SaasVariableDefinition, error) {
	var variables []SaasVariableDefinition
	if result := saasDB.Table(saasVariableTableName).Select("*").Scan(&variables); result.Error != nil {
		return variables, result.Error
	}
	return variables, nil
}

// fetchSaasClusterVariables 拉取 saas 中所有的集群变量
func fetchSaasClusterVariables() ([]SaasClusterVariable, error) {
	var variables []SaasClusterVariable
	if result := saasDB.Table(saasClusterVariableTableName).Select("*").Scan(&variables); result.Error != nil {
		return variables, result.Error
	}
	return variables, nil
}

// fetchSaasNamespaceVariables 拉取 saas 中所有的命名空间变量
func fetchSaasNamespaceVariables() ([]SaasNamespaceVariable, error) {
	var variables []SaasNamespaceVariable
	if result := saasDB.Table(saasNamespaceVariableTableName).Select("*").Scan(&variables); result.Error != nil {
		return variables, result.Error
	}
	return variables, nil
}

// doUpsertVariableDefinition do upsert variable definition
func doUpsertVariableDefinition(model store.ProjectModel, new *vdm.VariableDefinition) error {
	old, err := model.GetVariableDefinitionByKey(context.Background(), new.ProjectCode, new.Key)
	if err != nil {
		if err == drivers.ErrTableRecordNotFound {
			return tryGenerateIDAndCreateVD(model, new)
		}
		return err
	}
	new.ID = old.ID
	return model.UpsertVariableDefinition(context.Background(), new)
}

// tryGenerateIDAndCreateVD try to generate a new id and create variable definition
func tryGenerateIDAndCreateVD(model store.ProjectModel, definition *vdm.VariableDefinition) error {
	var count = 3
	var err error
	for i := 0; i < count; i++ {
		definition.ID = stringx.RandomString("variable-", 8)
		err = model.CreateVariableDefinition(context.Background(), definition)
		if err == nil {
			return nil
		}
		if err != drivers.ErrTableRecordDuplicateKey {
			return err
		}
	}
	return err
}
