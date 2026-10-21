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
迁移数据，从 bcs cc 服务模块(mysql 存储)，迁移到 bcs project模块(mongo 存储)
允许重复执行
*/

// Package main project
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	timeutil "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/time"
)

const (
	mysqlTableName = "projects"
)

var (
	mysqlHost       string
	mysqlPort       uint
	mysqlUser       string
	mysqlPwd        string
	mysqlDBName     string
	mongoAddr       string
	mongoReplicaset string
	mongoUser       string
	mongoPwd        string
	mongoDBName     string
	initProject     bool
	migrateCC       bool

	ccdb  *gorm.DB
	model store.ProjectModel
)

// BCSCCProjectData ...
type BCSCCProjectData struct {
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	Name        string `json:"name" gorm:"size:64;unique"`
	EnglishName string `json:"english_name" gorm:"size:64;unique;index"`
	Creator     string `json:"creator" gorm:"size:32"`
	Updator     string `json:"updator" gorm:"size:32"`
	Description string `json:"desc" sql:"type:text"`
	ProjectType uint   `json:"project_type"`
	IsOfflined  bool   `json:"is_offlined" gorm:"default:false"`
	ProjectID   string `json:"project_id" gorm:"size:32;unique;index"`
	UseBK       bool   `json:"use_bk" gorm:"default:true"`
	CCAppID     uint   `json:"cc_app_id"`
	Kind        uint   `json:"kind"`
	DeployType  string `json:"deploy_type"`
	BGID        uint   `json:"bg_id"`
	BGName      string `json:"bg_name"`
	DeptID      uint   `json:"dept_id"`
	DeptName    string `json:"dept_name"`
	CenterID    uint   `json:"center_id"`
	CenterName  string `json:"center_name"`
	DataID      uint   `json:"data_id"`
	IsSecrecy   bool   `json:"is_secrecy" gorm:"default:false"`
}

func main() {
	parseFlags()
	// 获取数据
	fmt.Printf("[%s] migrate start ...\n", time.Now().Format(time.RFC3339))

	if err := initDB(); err != nil {
		fmt.Printf("init db failed, err: %s\n", err.Error())
		return
	}

	if migrateCC {
		migrateCCData()
	}

	if initProject {
		fmt.Println("start init built-in project...")
		if err := initBuiltInProject(); err != nil {
			fmt.Printf("check and upsert init project failed, err: %s\n", err.Error())
			return
		}
		fmt.Println("init built-in project success!")
	}
}

func parseFlags() {
	// mysql
	flag.StringVar(&mysqlHost, "mysql_host", "", "mysql host")
	flag.UintVar(&mysqlPort, "mysql_port", 0, "mysql port")
	flag.StringVar(&mysqlUser, "mysql_user", "", "access mysql username")
	flag.StringVar(&mysqlPwd, "mysql_pwd", "", "access mysql password")
	flag.StringVar(&mysqlDBName, "mysql_db_name", "", "access mysql db name")

	// mongo
	flag.StringVar(&mongoAddr, "mongo_addr", "", "mongo address")
	flag.StringVar(&mongoReplicaset, "mongo_replicaset", "", "mongo replicaset")
	flag.StringVar(&mongoUser, "mongo_user", "", "access mongo username")
	flag.StringVar(&mongoPwd, "mongo_pwd", "", "access mongo password")
	flag.StringVar(&mongoDBName, "mongo_db_name", "", "access mongo db name")

	// init built-in project
	flag.BoolVar(&initProject, "init_project", false, "whether to init the built-in project")
	// migrate cc data
	flag.BoolVar(&migrateCC, "migrate_cc", false, "whether to migrate cc data")

	flag.Parse()
}

func initDB() error {
	// mongo
	store.InitMongo(&config.MongoConfig{
		Address:        mongoAddr,
		Replicaset:     mongoReplicaset,
		ConnectTimeout: 5,
		Database:       mongoDBName,
		Username:       mongoUser,
		Password:       mongoPwd,
		MaxPoolSize:    10,
		MinPoolSize:    1,
		Encrypted:      false,
	})
	model = store.New(store.GetMongo())

	// mysql
	if !migrateCC {
		return nil
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", mysqlUser, mysqlPwd, mysqlHost,
		mysqlPort, mysqlDBName)
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("access mysql error, %s", err.Error())
	}
	db.DB().SetConnMaxLifetime(10 * time.Second)
	db.DB().SetMaxIdleConns(20)
	db.DB().SetMaxOpenConns(20)
	ccdb = db
	return nil
}

// fetchBCSCCData xxx
// bcs cc 中查询数据
func fetchBCSCCData() []BCSCCProjectData {
	// 读取数据
	var p []BCSCCProjectData
	ccdb.Table(mysqlTableName).Select("*").Scan(&p)
	return p
}

// migrateCCData 迁移cc数据
func migrateCCData() {
	ccProjects := fetchBCSCCData()
	var totalCount, insertCount, updateCount int
	fmt.Printf("total projects length in cc: %d\n", len(ccProjects))
	projects, _, err := model.ListProjects(context.Background(), operator.EmptyCondition, &page.Pagination{All: true})
	if err != nil {
		fmt.Printf("list projects in bcs db failed, err: %s\n", err.Error())
		return
	}
	projectsMap := map[string]pm.Project{}
	fmt.Printf("total projects length in bcs: %d\n", len(projects))
	for _, project := range projects {
		project.CreateTime = timeutil.TransStrToUTCStr(time.RFC3339Nano, project.CreateTime)
		project.UpdateTime = timeutil.TransStrToUTCStr(time.RFC3339Nano, project.UpdateTime)
		projectsMap[project.ProjectID] = project
	}
	for _, ccProject := range ccProjects {
		totalCount++
		if totalCount%1000 == 0 {
			fmt.Printf("[%s] checked projects num: %d\n", time.Now().Format(time.RFC3339), totalCount)
		}
		project, exists := projectsMap[ccProject.ProjectID]
		if !exists {
			if err := insertProject(ccProject); err != nil {
				fmt.Printf("insert project %s failed, err: %s\n", ccProject.ProjectID, err.Error())
				return
			}
			insertCount++
			fmt.Printf("insert project %s success, count %d\n", ccProject.ProjectID, insertCount)
			continue
		}
		if checkUpdate(ccProject, project) {
			if err := updateProject(ccProject, project); err != nil {
				fmt.Printf("update project %s failed, err: %s\n", ccProject.ProjectID, err.Error())
				return
			}
			updateCount++
			fmt.Printf("update project %s success, count %d\n", ccProject.ProjectID, updateCount)
		}
	}
	fmt.Printf("[%s] migrate success! inserted %d projects, updated %d projects\n",
		time.Now().Format(time.RFC3339), insertCount, updateCount)
}

// upsertInitProject 初始化集群配置，项目ID / Code 固定
func initBuiltInProject() error {
	projectID := stringx.GetEnv("INIT_PROJECT_ID", "")
	if projectID == "" {
		return errors.New("init projectID can not be empty")
	}

	var (
		tenantID, businessID, tenantProjectCode, projectCode string
	)

	// check tenant switch
	enableTenant := stringx.GetEnv("ENABLE_MULTI_TENANT", "false")
	if enableTenant == "true" {
		tenantID = constant.SystemTenantId
		businessID = stringx.GetEnv("INIT_PROJECT_BUSINESS_ID", "1")
		tenantProjectCode = stringx.GetEnv("INIT_PROJECT_CODE", "blueking")
		projectCode = fmt.Sprintf("%s-%s", tenantID, tenantProjectCode)
	} else {
		tenantID = constant.DefaultTenantId
		businessID = stringx.GetEnv("INIT_PROJECT_BUSINESS_ID", "2")
		tenantProjectCode = stringx.GetEnv("INIT_PROJECT_CODE", "blueking")
		projectCode = stringx.GetEnv("INIT_PROJECT_CODE", "blueking")
	}

	p := &pm.Project{
		ProjectID:         projectID,
		Name:              "蓝鲸",
		ProjectCode:       projectCode,
		TenantID:          tenantID,
		TenantProjectCode: tenantProjectCode,
		Creator:           stringx.GetEnv("INIT_PROJECT_USER", "admin"),
		Updater:           stringx.GetEnv("INIT_PROJECT_USER", "admin"),
		Managers:          stringx.GetEnv("INIT_PROJECT_USER", "admin"),
		ProjectType:       0,
		UseBKRes:          false,
		Description:       "BCS built-in project",
		IsOffline:         false,
		Kind:              "k8s",
		BusinessID:        businessID,
		DeployType:        2,
		BGID:              "0",
		DeptID:            "0",
		CenterID:          "0",
		IsSecret:          false,
	}
	_, err := model.GetProject(context.Background(), p.ProjectID)
	if err == nil {
		return nil
	}
	if err != nil && err != drivers.ErrTableRecordNotFound {
		return err
	}
	return model.CreateProject(context.Background(), p)
}

func insertProject(p BCSCCProjectData) error {
	project := &pm.Project{
		ProjectID:   p.ProjectID,
		Name:        p.Name,
		ProjectCode: p.EnglishName,
		Creator:     p.Creator,
		Updater:     p.Updator,
		Managers:    constructManagers(p.Creator, p.Updator),
		ProjectType: uint32(p.ProjectType),
		UseBKRes:    p.UseBK,
		Description: p.Description,
		IsOffline:   p.IsOfflined,
		Kind:        getStrKind(p.Kind),
		BusinessID:  getBusinessID(p.CCAppID),
		DeployType:  getDeployType(p.DeployType),
		BGID:        strconv.Itoa(int(p.BGID)),
		BGName:      p.BGName,
		DeptID:      strconv.Itoa(int(p.DeptID)),
		DeptName:    p.DeptName,
		CenterID:    strconv.Itoa(int(p.CenterID)),
		CenterName:  p.CenterName,
		IsSecret:    p.IsSecrecy,
	}
	return model.CreateProject(context.Background(), project)
}

func updateProject(c BCSCCProjectData, p pm.Project) error {
	p.Updater = c.Updator
	p.Managers = constructManagers(c.Creator, c.Updator)
	p.Kind = getStrKind(c.Kind)
	p.BusinessID = getBusinessID(c.CCAppID)
	p.DeployType = getDeployType(c.DeployType)
	return model.UpdateProject(context.Background(), &p)
}

func checkUpdate(c BCSCCProjectData, p pm.Project) bool {
	// kind == 0， 未开启， 不需要迁移
	if c.Kind != 1 && c.Kind != 2 {
		return false
	}

	// 类型不一致， 需要修改
	if getStrKind(c.Kind) != p.Kind {
		return true
	}

	// cc == 0， 未开启， 不需要迁移
	if c.CCAppID == 0 {
		return false
	}

	// 其他不相等，需要迁移
	if getBusinessID(c.CCAppID) != p.BusinessID {
		return true
	}

	return false
}

// getStrKind 获取字符串类型 kind，1 => k8s 2 => mesos
func getStrKind(kind uint) string {
	if kind == 1 {
		return "k8s"
	} else if kind == 2 {
		return "mesos"
	}
	return ""
}

// getBusinessID 获取字符串类型 0 => ""
func getBusinessID(ccAppID uint) string {
	if ccAppID == 0 {
		return ""
	}
	return strconv.Itoa(int(ccAppID))
}

func stringInSlice(str string, list []string) bool {
	for _, item := range list {
		if item == str {
			return true
		}
	}
	return false
}

// constructManagers 组装manager
func constructManagers(creator string, updater string) string {
	managers := []string{creator}
	if updater != "" {
		if !stringInSlice(updater, managers) {
			managers = append(managers, updater)
		}
	}
	return strings.Join(managers, ";")
}

// getDeployType 获取 int 型 deployType
func getDeployType(deployType string) uint32 {
	if deployType == "null" {
		return 1
	}
	return 2
}
