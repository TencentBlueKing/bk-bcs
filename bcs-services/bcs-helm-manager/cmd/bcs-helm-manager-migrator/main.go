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
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	microCfg "github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/encoder/yaml"
	"github.com/micro/go-micro/v2/config/reader"
	microJson "github.com/micro/go-micro/v2/config/reader/json"
	microFile "github.com/micro/go-micro/v2/config/source/file"
	microFlg "github.com/micro/go-micro/v2/config/source/flag"
	microRgt "github.com/micro/go-micro/v2/registry"
	microEtcd "github.com/micro/go-micro/v2/registry/etcd"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	helmrelease "helm.sh/helm/v3/pkg/release"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/envx"
)

var (
	// C define config
	C *options.HelmManagerOptions

	// mysql
	mysqlHost     = envx.GetEnv("MYSQL_HOST", "")
	mysqlPort     = envx.GetEnv("MYSQL_PORT", "3306")
	mysqlUsername = envx.GetEnv("MYSQL_USERNAME", "")
	mysqlPassword = envx.GetEnv("MYSQL_PASSWORD", "")
	mysqlDatabase = envx.GetEnv("MYSQL_DATABASE", "")
)

func parseFlags() {
	// config file path
	flag.String("conf", "", "config file path")
	flag.Parse()
}

func main() {
	parseFlags()
	loadConfig()
	if err := initRegistry(); err != nil {
		blog.Fatalf("init registry error, %s", err.Error())
	}

	// init log
	blog.InitLogs(conf.LogConfig{
		Verbosity: 3,
		ToStdErr:  true,
	})

	// init mongo
	password := C.Mongo.Password
	if password != "" && C.Mongo.Encrypted {
		realPwd, err := encrypt.DesDecryptFromBase([]byte(C.Mongo.Password))
		if err != nil {
			blog.Fatalf("decrypt password failed, err %s", err.Error())
		}
		password = string(realPwd)
	}

	mongoOptions := &mongo.Options{
		Hosts:                 strings.Split(C.Mongo.Address, ","),
		ConnectTimeoutSeconds: int(C.Mongo.ConnectTimeout),
		AuthDatabase:          C.Mongo.AuthDatabase,
		Database:              C.Mongo.Database,
		Username:              C.Mongo.Username,
		Password:              password,
		MaxPoolSize:           uint64(C.Mongo.MaxPoolSize),
		MinPoolSize:           uint64(C.Mongo.MinPoolSize),
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
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		mysqlUsername, mysqlPassword, mysqlHost, mustInt(mysqlPort), mysqlDatabase)
	mysqlDB, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		blog.Fatalf("init mysql db failed, err %s", err.Error())
	}

	// migration
	migrateRepo(model, mysqlDB)
	migrateReleases(model, mysqlDB)
}

func migrateRepo(model store.HelmManagerModel, mysqlDB *gorm.DB) {
	repos := getSaasHelmRepo(mysqlDB)
	existRepos := getHelmRepo(model)
	blog.Infof("get %d repos from saas, %d repos from helmmanager", len(repos), len(existRepos))
	syncRepos := 0
	for _, repo := range repos {
		exist := false
		for _, v := range existRepos {
			if v.ProjectID == repo.Name && v.Name == repo.Name {
				exist = true
				break
			}
		}
		blog.Infof("syncing project %s chart repo, exist: %v", repo.Name, exist)

		// sync project repo
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
			syncRepos++
			blog.Infof("create project %s repository successful", repo.Name)
		}

		// sync public repo
		if len(common.GetPublicRepoURL(C.Repo.URL, C.Repo.PublicRepoProject, C.Repo.PublicRepoName)) != 0 {
			createOrUpdatePublicRepo(model, repo.Name)
		}
	}
	blog.Infof("%d repos are synced", syncRepos)
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
		"LEFT JOIN helm_repo_auth AS ra ON ra.repo_id = r.id WHERE r.name != ?", common.PublicRepoName).
		Scan(&repos).Error
	if err != nil {
		blog.Fatalf("get saas helm repo failed, err %s", err.Error())
	}
	blog.Infof("get %d repo from saas helm repo", len(repos))
	return repos
}

func createOrUpdatePublicRepo(model store.HelmManagerModel, projectID string) {
	_, err := model.GetRepository(context.TODO(), projectID, common.PublicRepoName)
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("create project %s public repo failed, err %s", projectID, err.Error())
		return
	}
	if err == nil {
		blog.Errorf("project %s has public repo, skip", projectID)
		return
	}
	now := time.Now().Unix()
	err = model.CreateRepository(context.TODO(), &entity.Repository{
		ProjectID:   projectID,
		Name:        common.PublicRepoName,
		DisplayName: common.PublicRepoDisplayName,
		Public:      true,
		Type:        "HELM",
		RepoURL:     common.GetPublicRepoURL(C.Repo.URL, C.Repo.PublicRepoProject, C.Repo.PublicRepoName),
		CreateBy:    "admin",
		UpdateBy:    "admin",
		CreateTime:  now,
		UpdateTime:  now,
	})
	if err != nil {
		blog.Errorf("create project %s public repo failed, err %s", projectID, err.Error())
	}
}

func migrateReleases(model store.HelmManagerModel, mysqlDB *gorm.DB) {
	releases := getSaasHelmReleases(mysqlDB)
	blog.Infof("get %d releases from saas, syncing", len(releases))
	syncReleases := 0
	existReleases := 0
	for _, v := range releases {
		if len(v.Name) == 0 || len(v.ClusterID) == 0 || len(v.Namespace) == 0 {
			continue
		}
		rl, err := v.toEntity()
		if err != nil {
			blog.Errorf("create releases %s in cluster %s namespace %s, err %s",
				v.Name, v.ClusterID, v.Namespace, err.Error())
			continue
		}
		_, err = model.GetRelease(context.TODO(), v.ClusterID, v.Namespace, v.Name)
		if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Errorf("create releases %s in cluster %s namespace %s, err %s",
				v.Name, v.ClusterID, v.Namespace, err.Error())
			continue
		}
		if err == nil {
			existReleases++
			up := entity.M{
				entity.FieldKeyChartVersion: rl.ChartVersion,
				entity.FieldKeyRevision:     rl.Revision,
				entity.FieldKeyValueFile:    rl.ValueFile,
				entity.FieldKeyValues:       rl.Values,
				entity.FieldKeyArgs:         rl.Args,
				entity.FieldKeyUpdateBy:     rl.UpdateBy,
				entity.FieldKeyUpdateTime:   rl.UpdateTime,
				entity.FieldKeyStatus:       rl.Status,
				entity.FieldKeyMessage:      rl.Message,
			}
			err = model.UpdateRelease(context.TODO(), v.ClusterID, v.Namespace, v.Name, up)
			if err != nil {
				blog.Errorf("update releases %s in cluster %s namespace %s, err %s",
					v.Name, v.ClusterID, v.Namespace, err.Error())
			}
			continue
		}
		err = model.CreateRelease(context.TODO(), rl)
		if err != nil {
			blog.Errorf("create releases %s in cluster %s namespace %s, err %s",
				v.Name, v.ClusterID, v.Namespace, err.Error())
			continue
		}
		syncReleases++
	}
	blog.Infof("%d releases are synced, %d releases are exist", syncReleases, existReleases)
}

func getSaasHelmReleases(db *gorm.DB) []release {
	var releases []release
	err := db.Raw("SELECT app.name,app.namespace,app.project_id,app.cluster_id,r.name AS repo," +
		"hc.name AS chart_name,app.version,rl.revision,rl.valuefile_name,rl.valuefile,app.cmd_flags," +
		"app.creator,app.updator,app.created,app.updated,app.transitioning_result as status," +
		"app.transitioning_message AS message from bcs_k8s_app as app LEFT JOIN helm_chart as hc ON " +
		"hc.id=app.chart_id LEFT JOIN helm_repository AS r ON r.id=hc.repository_id LEFT JOIN " +
		"helm_chart_release AS rl ON rl.id=app.release_id").
		Scan(&releases).Error
	if err != nil {
		blog.Fatalf("get saas helm releases failed, err %s", err.Error())
	}
	return releases
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

type release struct {
	Name         string    `json:"name"`
	Namespace    string    `json:"namespace"`
	ProjectID    string    `json:"projectID" gorm:"column:project_id"`
	ClusterID    string    `json:"clusterID" gorm:"column:cluster_id"`
	Repo         string    `json:"repo"`
	ChartName    string    `json:"chartName" gorm:"column:chart_name"`
	ChartVersion string    `json:"chartVersion" gorm:"column:version"`
	Revision     int       `json:"revision"`
	ValueFile    string    `json:"valueFile" gorm:"column:valuefile_name"`
	Values       string    `json:"values" gorm:"column:valuefile"`
	ArgString    string    `json:"args" gorm:"column:cmd_flags"`
	CreateBy     string    `json:"createBy" gorm:"column:creator"`
	UpdateBy     string    `json:"updateBy" gorm:"column:updator"`
	CreateTime   time.Time `json:"createTime" gorm:"column:created"`
	UpdateTime   time.Time `json:"updateTime" gorm:"column:updated"`
	Status       int       `json:"status"`
	Message      string    `json:"message"`
}

func (r *release) toEntity() (*entity.Release, error) {
	args := make([]string, 0)
	flags := make([]map[string]interface{}, 0)
	err := json.Unmarshal([]byte(r.ArgString), &flags)
	if err != nil {
		return nil, fmt.Errorf("get %s cmd_flags %s error, err %s", r.Name, r.ArgString, err.Error())
	}
	for _, flag := range flags {
		for k, v := range flag {
			args = append(args, fmt.Sprintf("%s=%v", k, v))
		}
	}

	status := helmrelease.StatusDeployed
	if r.Status != 1 {
		status = helmrelease.StatusFailed
	}
	return &entity.Release{
		Name:         r.Name,
		Namespace:    r.Namespace,
		ProjectCode:  r.getProjectCode(),
		ClusterID:    r.ClusterID,
		Repo:         r.Repo,
		ChartName:    r.ChartName,
		ChartVersion: r.ChartVersion,
		Revision:     r.Revision,
		ValueFile:    r.ValueFile,
		Values:       []string{r.Values},
		Args:         args,
		CreateBy:     r.CreateBy,
		UpdateBy:     r.UpdateBy,
		CreateTime:   r.CreateTime.Unix(),
		UpdateTime:   r.UpdateTime.Unix(),
		Status:       status.String(),
		Message:      r.Message,
	}, nil
}

func (r *release) getProjectCode() string {
	p, err := project.GetProjectByCode(r.ProjectID)
	if err != nil {
		blog.Errorf("get project for %s error, %s", r.ProjectID, err.Error())
		return ""
	}
	return p.ProjectCode
}

func mustInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func loadConfig() {
	opt := &options.HelmManagerOptions{}
	config, err := microCfg.NewConfig(microCfg.WithReader(microJson.NewReader(
		reader.WithEncoder(yaml.NewEncoder()),
	)))
	if err != nil {
		blog.Fatalf("create config failed, %s", err.Error())
	}

	if err = config.Load(
		microFlg.NewSource(
			microFlg.IncludeUnset(true),
		),
	); err != nil {
		blog.Fatalf("load config from flag failed, %s", err.Error())
	}

	if len(config.Get("conf").String("")) > 0 {
		err = config.Load(microFile.NewSource(microFile.WithPath(config.Get("conf").String(""))))
		if err != nil {
			blog.Fatalf("load config from file failed, err %s", err.Error())
		}
	}

	if err = config.Scan(opt); err != nil {
		blog.Fatalf("scan config failed, %s", err.Error())
	}
	C = opt
}

func initRegistry() error {
	var (
		tlsConfig *tls.Config
		err       error
	)
	if len(C.TLS.ClientCert) != 0 && len(C.TLS.ClientKey) != 0 && len(C.TLS.ClientCa) != 0 {
		tlsConfig, err = ssl.ClientTslConfVerity(C.TLS.ClientCa, C.TLS.ClientCert,
			C.TLS.ClientKey, static.ClientCertPwd)
		if err != nil {
			blog.Fatalf("load helm manager client tls config failed, err %s", err.Error())
		}
		blog.Info("load helm manager client tls config successfully")
	}

	etcdEndpoints := common.SplitAddrString(C.Etcd.EtcdEndpoints)
	etcdSecure := false

	var etcdTLS *tls.Config
	if len(C.Etcd.EtcdCa) != 0 && len(C.Etcd.EtcdCert) != 0 && len(C.Etcd.EtcdKey) != 0 {
		etcdSecure = true
		etcdTLS, err = ssl.ClientTslConfVerity(C.Etcd.EtcdCa, C.Etcd.EtcdCert, C.Etcd.EtcdKey, "")
		if err != nil {
			return err
		}
	}

	blog.Infof("get etcd endpoints for registry: %v, with secure %t", etcdEndpoints, etcdSecure)

	reg := microEtcd.NewRegistry(
		microRgt.Addrs(etcdEndpoints...),
		microRgt.Secure(etcdSecure),
		microRgt.TLSConfig(etcdTLS),
	)
	if err := reg.Init(); err != nil {
		return err
	}
	return project.NewClient(tlsConfig, reg)
}
