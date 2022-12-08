/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

/*
迁移数据，从 bcs cc 服务模块(mysql 存储)，迁移到 bcs project模块(mongo 存储)
允许重复执行
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
)

const (
	mysqlTableName = "projects"
	mongoTableName = "bcsproject_project"
	timeLayout     = "2006-01-02T15:04:05Z"
)

var (
	mysqlHost   string
	mysqlPort   uint
	mysqlUser   string
	mysqlPwd    string
	mysqlDBName string
	mongoAddr   string
	mongoUser   string
	mongoPwd    string
	mongoDBName string

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
	fmt.Println("migrate start ...")

	if err := initDB(); err != nil {
		fmt.Printf("init db failed, err: %s\n", err.Error())
		return
	}

	ccProjects, err := fetchBCSCCData()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var insertCount, updateCount int
	fmt.Printf("total projects length in cc: %d\n", len(ccProjects))
	for _, ccProject := range ccProjects {
		project, err := model.GetProject(context.Background(), ccProject.ProjectID)
		if err != nil && err != drivers.ErrTableRecordNotFound {
			fmt.Printf("get project %s failed, err: %s\n", ccProject.ProjectID, err.Error())
			return
		}
		if err == drivers.ErrTableRecordNotFound {
			if err := insertProject(ccProject); err != nil {
				fmt.Printf("insert project %s failed, err: %s\n", ccProject.ProjectID, err.Error())
				return
			}
			insertCount++
			fmt.Printf("insert project %s success, count %d\n", ccProject.ProjectID, insertCount)
			continue
		}
		if checkUpdate(&ccProject, project) {
			if err := updateProject(ccProject, project); err != nil {
				fmt.Printf("update project %s failed, err: %s\n", ccProject.ProjectID, err.Error())
			}
			updateCount++
			fmt.Printf("update project %s success, count %d\n", ccProject.ProjectID, updateCount)
		}
	}
	fmt.Println("migrate success!")
	fmt.Printf("inserted %d projects, updated %d projects\n", insertCount, updateCount)
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
	flag.StringVar(&mongoUser, "mongo_user", "", "access mongo username")
	flag.StringVar(&mongoPwd, "mongo_pwd", "", "access mongo password")
	flag.StringVar(&mongoDBName, "mongo_db_name", "", "access mongo db name")

	flag.Parse()
}

func initDB() error {
	// mongo
	store.InitMongo(&config.MongoConfig{
		Address:        mongoAddr,
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
func fetchBCSCCData() ([]BCSCCProjectData, error) {
	// 读取数据
	var p []BCSCCProjectData
	ccdb.Table(mysqlTableName).Select("*").Scan(&p)
	return p, nil
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
		BusinessID:  strconv.Itoa(int(p.CCAppID)),
		DeployType:  getDeployType(p.DeployType),
		BGID:        strconv.Itoa(int(p.BGID)),
		BGName:      p.BGName,
		DeptID:      strconv.Itoa(int(p.DeptID)),
		DeptName:    p.DeptName,
		CenterID:    strconv.Itoa(int(p.CenterID)),
		CenterName:  p.CenterName,
		IsSecret:    p.IsSecrecy,
		CreateTime:  p.CreatedAt.Format(timeLayout),
		UpdateTime:  p.CreatedAt.Format(timeLayout),
	}
	return model.CreateProject(context.Background(), project)
}

func updateProject(c BCSCCProjectData, p *pm.Project) error {
	p.Name = c.Name
	p.Creator = c.Creator
	p.Updater = c.Updator
	p.Managers = constructManagers(c.Creator, c.Updator)
	p.ProjectType = uint32(c.ProjectType)
	p.UseBKRes = c.UseBK
	p.Description = c.Description
	p.IsOffline = c.IsOfflined
	p.Kind = getStrKind(c.Kind)
	p.BusinessID = strconv.Itoa(int(c.CCAppID))
	p.DeployType = getDeployType(c.DeployType)
	p.BGID = strconv.Itoa(int(c.BGID))
	p.BGName = c.BGName
	p.DeptID = strconv.Itoa(int(c.DeptID))
	p.DeptName = c.DeptName
	p.CenterID = strconv.Itoa(int(c.CenterID))
	p.CenterName = c.CenterName
	p.IsSecret = c.IsSecrecy
	p.CreateTime = c.CreatedAt.Format(timeLayout)
	p.UpdateTime = c.CreatedAt.Format(timeLayout)
	return model.UpdateProject(context.Background(), p)
}

func checkUpdate(c *BCSCCProjectData, p *pm.Project) bool {
	return getStrKind(c.Kind) != p.Kind || strconv.Itoa(int(c.CCAppID)) != p.BusinessID
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
