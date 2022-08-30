/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/envx"
)

var (
	// mongo
	mongoHosts        = envx.GetEnv("MONGO_ADDRESS", "")
	mongoConnTimeout  = envx.GetEnv("MONGO_CONNECT_TIMEOUT", "3")
	mongoAuthDatabase = envx.GetEnv("MONGO_AUTH_DATABASE", "admin")
	mongoDatabase     = envx.GetEnv("MONGO_DATABASE", "")
	mongoUsername     = envx.GetEnv("MONGO_USERNAME", "")
	mongoPassword     = envx.GetEnv("MONGO_PASSWORD", "")
	mongoMaxPoolSize  = envx.GetEnv("MONGO_MAX_POOL_SIZE", "")
	mongoMinPoolSize  = envx.GetEnv("MONGO_MIN_POOL_SIZE", "")
	mongoEncrypted    = envx.GetEnv("MONGO_ENCRYPTED", "true")

	// mysql
	mysqlHost     = envx.GetEnv("MYSQL_HOST", "")
	mysqlPort     = envx.GetEnv("MYSQL_PORT", "3306")
	mysqlUsername = envx.GetEnv("MYSQL_USERNAME", "")
	mysqlPassword = envx.GetEnv("MYSQL_PASSWORD", "")
	mysqlDatabase = envx.GetEnv("MYSQL_DATABASE", "")
)

func main() {
	// init log
	blog.InitLogs(conf.LogConfig{
		Verbosity: 3,
		ToStdErr:  true,
	})

	// init mongo
	password := mongoPassword
	if password != "" && mustBool(mongoEncrypted) {
		realPwd, err := encrypt.DesDecryptFromBase([]byte(password))
		if err != nil {
			blog.Fatalf("decrypt password failed, err %s", err.Error())
		}
		password = string(realPwd)
	}

	mongoOptions := &mongo.Options{
		Hosts:                 strings.Split(mongoHosts, ","),
		ConnectTimeoutSeconds: mustInt(mongoConnTimeout),
		AuthDatabase:          mongoAuthDatabase,
		Database:              mongoDatabase,
		Username:              mongoUsername,
		Password:              password,
		MaxPoolSize:           uint64(mustInt(mongoMaxPoolSize)),
		MinPoolSize:           uint64(mustInt(mongoMinPoolSize)),
	}
	mongoDB, err := mongo.NewDB(mongoOptions)
	if err != nil {
		blog.Fatalf("init mongo db failed, err %s", err.Error())
	}
	if err = mongoDB.Ping(); err != nil {
		blog.Fatalf("ping mongo db failed, err %s", err.Error())
	}
	model := store.New(mongoDB)
	blog.Info("init mongo db successfully")

	// init mysql
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlUsername, mysqlPassword, mysqlHost, mustInt(mysqlPort), mysqlDatabase)
	mysqlDB, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		blog.Fatalf("init mysql db failed, err %s", err.Error())
	}

	// migration
	migrateRepo(model, mysqlDB)
}

func migrateRepo(model store.HelmManagerModel, mysqlDB *gorm.DB) {
	repos := getSaasHelmRepo(mysqlDB)
	existRepos := getHelmRepo(model)
	for _, repo := range repos {
		exist := false
		for _, v := range existRepos {
			if v.ProjectID == repo.Name && v.Name == repo.Name {
				exist = true
				break
			}
		}
		blog.Infof("syncing project %s chart repo, exist: %v", repo.Name, exist)

		if !exist {
			if len(repo.CredentialString) == 0 {
				blog.Infof("skip sync project %s chart repo, because credentials is null", repo.Name)
				continue
			}
			cred := &credential{}
			err := json.Unmarshal([]byte(repo.CredentialString), cred)
			if err != nil {
				blog.Errorf("sync project %s chart repo failed, credentials %s, err %s", repo.Name,
					repo.CredentialString, err.Error())
				continue
			}
			now := time.Now().Unix()
			err = model.CreateRepository(context.TODO(), &entity.Repository{
				ProjectID:  repo.Name,
				Name:       repo.Name,
				Type:       "HELM",
				RepoURL:    repo.URL,
				Username:   cred.Username,
				Password:   cred.Password,
				CreateBy:   "admin",
				UpdateBy:   "admin",
				CreateTime: now,
				UpdateTime: now,
			})
			if err != nil {
				blog.Errorf("create project %s repository failed, err %s", repo.Name, err.Error())
				continue
			}
			blog.Infof("create project %s repository successful", repo.Name)
		}
	}
}

func getHelmRepo(model store.HelmManagerModel) []*entity.Repository {
	options := &utils.ListOption{
		Sort: map[string]int{},
		Page: 0,
		Size: 0,
	}
	cond := make(operator.M)
	cond.Update(entity.FieldKeyType, "HELM")
	_, repos, err := model.ListRepository(context.TODO(), operator.NewLeafCondition(operator.Eq, cond), options)
	if err != nil {
		blog.Fatalf("ListRepository failed, err %s", err.Error())
	}
	blog.Infof("get %d repo from helm manager", len(repos))
	return repos
}

func getSaasHelmRepo(db *gorm.DB) []repo {
	var repos []repo
	err := db.Raw("SELECT r.id, r.url, r.name, r.project_id, ra.credentials FROM helm_repository AS r "+
		"LEFT JOIN helm_repo_auth AS ra ON ra.repo_id = r.id WHERE r.provider = ?", "bkrepo").
		Scan(&repos).Error
	if err != nil {
		blog.Fatalf("get saas helm repo failed, err %s", err.Error())
	}
	blog.Infof("get %d repo from saas helm repo", len(repos))
	return repos
}

type repo struct {
	ID               int    `json:"id,omitempty"`
	URL              string `json:"url,omitempty" gorm:"column:url"`
	Name             string `json:"name,omitempty"`
	ProjectID        string `json:"project_id,omitempty"`
	CredentialString string `json:"credentials,omitempty" gorm:"column:credentials"`
}

type credential struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func mustInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func mustBool(s string) bool {
	v, _ := strconv.ParseBool(s)
	return v
}
