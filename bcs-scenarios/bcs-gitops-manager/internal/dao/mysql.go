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

package dao

import (
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	_ "github.com/go-sql-driver/mysql" // nolint
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao/mysqlrate"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
)

type driver struct {
	db         *gorm.DB
	rateClient mysqlrate.RateInterface
}

var (
	globalDB *driver
)

// GlobalDB global db
func GlobalDB() Interface {
	return globalDB
}

// NewDriver creates the MySQL instance
func NewDriver(dbCfg *common.DBConfig) (Interface, error) {
	connArgs := fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbCfg.Username, dbCfg.Password, dbCfg.Addr, dbCfg.Database)
	var err error
	globalDB, err = newDriver(connArgs)
	if err != nil {
		return nil, err
	}
	if dbCfg.LimitQPS == 0 {
		dbCfg.LimitQPS = 200
	}
	globalDB.rateClient = mysqlrate.NewRateLimit(globalDB.db, dbCfg.LimitQPS)
	return globalDB, nil
}

func newDriver(connArgs string) (*driver, error) {
	db, err := gorm.Open(mysql.Open(connArgs), &gorm.Config{})
	if err != nil {
		blog.Errorf("Connect to MySQL '%s' failed, err: %s", connArgs, err.Error())
		return nil, err
	}
	return &driver{
		db: db,
	}, nil
}

// Init will auto create the tables if not exist
func (d *driver) Init() error {
	if err := d.autoCreateTable(); err != nil {
		return errors.Wrapf(err, "db driver init failed")
	}
	return nil
}

func (d *driver) autoCreateTable() error {
	if err := d.createTable(tableActivityUser, &ActivityUser{}); err != nil {
		return errors.Wrapf(err, "create table '%s' failed", tableActivityUser)
	}
	if err := d.createTable(tableHistoryManifest, &ApplicationHistoryManifest{}); err != nil {
		return errors.Wrapf(err, "create table '%s' failed", tableHistoryManifest)
	}
	if err := d.createTable(tableUserPermission, &UserPermission{}); err != nil {
		return errors.Wrapf(err, "create table '%s' failed", tableUserPermission)
	}
	if err := d.createTable(tableUserAudit, &UserAudit{}); err != nil {
		return errors.Wrapf(err, "create table '%s' failed", tableUserAudit)
	}
	if err := d.createTable(tableAppSetClusterScope, &AppSetClusterScope{}); err != nil {
		return errors.Wrapf(err, "create table '%s' failed", tableUserAudit)
	}
	return nil
}

func (d *driver) createTable(tableName string, obj interface{}) error {
	if d.db.Migrator().HasTable(tableName) {
		blog.Infof("[DB] table '%s' existed.", tableName)
	} else {
		if err := d.db.Table(tableName).Set("gorm:table_options", "ENGINE=InnoDB AUTO_INCREMENT = 0").
			AutoMigrate(obj); err != nil {
			return errors.Wrapf(err, "create table '%s' failed", tableName)
		}
		blog.Infof("[DB] create table '%s' success.", tableName)
	}
	return nil
}

// SaveApplicationHistoryManifest create application history manifest object
func (d *driver) SaveApplicationHistoryManifest(hm *ApplicationHistoryManifest) error {
	return d.db.Table(tableHistoryManifest).Save(hm).Error
}

// GetApplicationHistoryManifest get application history manifest
func (d *driver) GetApplicationHistoryManifest(appName, appUID string,
	historyID int64) (*ApplicationHistoryManifest, error) {
	rows, err := d.rateClient.Table(tableHistoryManifest).Where("name = ?", appName).
		Where("applicationUID = ?", appUID).
		Where("historyID = ?", historyID).Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "get application history manifest failed")
	}
	defer rows.Close() // nolint

	result := make([]ApplicationHistoryManifest, 0)
	for rows.Next() {
		appHM := new(ApplicationHistoryManifest)
		if err = d.db.ScanRows(rows, appHM); err != nil {
			return nil, errors.Wrapf(err, "scan syncinfo rows failed")
		}
		result = append(result, *appHM)
	}
	if len(result) == 0 {
		return nil, nil
	}
	return &result[0], nil
}

// CheckAppHistoryManifestExist check manifest whether exist
func (d *driver) CheckAppHistoryManifestExist(appName, appUID string, historyID int64) (bool, error) {
	rows, err := d.rateClient.Table(tableHistoryManifest).Select("id").
		Where("name = ?", appName).
		Where("applicationUID = ?", appUID).
		Where("historyID = ?", historyID).Rows()
	if err != nil {
		return false, errors.Wrapf(err, "get application history manifest failed")
	}
	defer rows.Close() // nolint

	for rows.Next() {
		return true, nil
	}
	return false, nil
}

// ActivityUserItem defines the activity user
type ActivityUserItem struct {
	Project string
	User    string
}

// SaveActivityUser save the activity user
func (d *driver) SaveActivityUser(user *ActivityUser) error {
	return d.rateClient.Table(tableActivityUser).Save(user).Error
}

// GetActivityUser get activity user
func (d *driver) GetActivityUser(project, user string) (*ActivityUser, error) {
	rows, err := d.rateClient.Table(tableActivityUser).Where("project = ?", project).
		Where("userName = ?", user).Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "get activity user failed")
	}
	defer rows.Close() // nolint

	result := make([]*ActivityUser, 0)
	for rows.Next() {
		activityUser := new(ActivityUser)
		if err = d.db.ScanRows(rows, activityUser); err != nil {
			return nil, errors.Wrapf(err, "scan activity user rows failed")
		}
		result = append(result, activityUser)
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result[0], nil
}

// UpdateActivityUser update the activity user
func (d *driver) UpdateActivityUser(user *ActivityUser) error {
	if err := d.rateClient.Table(tableActivityUser).Where("id = ?", user.ID).UpdateColumns(map[string]interface{}{
		"operateNum":       user.OperateNum,
		"lastActivityTime": time.Now(),
	}).Error; err != nil {
		return errors.Wrapf(err, "update activity_user '%d' failed", user.ID)
	}
	return nil
}

// UpdateActivityUserWithName update user activity with user name
func (d *driver) UpdateActivityUserWithName(item *ActivityUserItem) {
	activityUser, err := d.GetActivityUser(item.Project, item.User)
	if err != nil {
		blog.Errorf("[analysis] get activity user '%s/%s' failed: %s", item.Project, item.User, err.Error())
		return
	}
	if activityUser == nil {
		activityUser = &ActivityUser{
			Project:          item.Project,
			UserName:         item.User,
			OperateNum:       1,
			LastActivityTime: time.Now(),
		}
		if err = d.SaveActivityUser(activityUser); err != nil {
			blog.Errorf("[analysis] save activity user failed: %s", err.Error())
			return
		}
		return
	}
	activityUser.OperateNum++
	if err = d.UpdateActivityUser(activityUser); err != nil {
		blog.Errorf("[analysis] update activity user failed: %s", err.Error())
		return
	}
}

// UpdateResourcePermissions update the resource's permission with users
func (d *driver) UpdateResourcePermissions(project, rsType, rsName, rsAction string, users []string) error {
	err := d.db.Transaction(func(tx *gorm.DB) error {
		if err := d.db.Table(tableUserPermission).Where("project = ?", project).
			Where("resourceType = ?", rsType).Where("resourceName = ?", rsName).
			Where("resourceAction = ?", rsAction).
			Not(map[string]interface{}{"user": users}).Delete(&UserPermission{}).Error; err != nil {
			return errors.Wrapf(err, "delete user permissions failed")
		}
		for _, user := range users {
			up := &UserPermission{
				Project:        project,
				User:           user,
				ResourceType:   rsType,
				ResourceName:   rsName,
				ResourceAction: rsAction,
			}
			if err := d.db.Table(tableUserPermission).Save(up).Error; err != nil {
				if strings.Contains(err.Error(), "Duplicate") {
					continue
				}
				return errors.Wrapf(err, "save user permissions failed")
			}
		}
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "update user permissions failed")
	}
	return nil
}

// CreateUserPermission create user permission
func (d *driver) CreateUserPermission(permission *UserPermission) error {
	if err := d.db.Table(tableUserPermission).Save(permission).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return nil
		}
		return errors.Wrapf(err, "save user permission failed")
	}
	return nil
}

// DeleteUserPermission delete user permission
func (d *driver) DeleteUserPermission(permission *UserPermission) error {
	if err := d.db.Table(tableUserPermission).Where("project = ?", permission.Project).
		Where("user = ?", permission.User).
		Where("resourceType = ?", permission.ResourceType).
		Where("resourceName = ?", permission.ResourceName).
		Where("resourceAction = ?", permission.ResourceAction).
		Delete(&UserPermission{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return errors.Wrapf(err, "delete permission failed")
	}
	return nil
}

// ListUserPermissions list user permissions by resource type
func (d *driver) ListUserPermissions(user, project, resourceType string) ([]*UserPermission, error) {
	rows, err := d.db.Table(tableUserPermission).Where("project = ?", project).
		Where("user = ?", user).
		Where("resourceType = ?", resourceType).Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "query user permissions failed")
	}
	defer rows.Close() // nolint

	result := make([]*UserPermission, 0)
	for rows.Next() {
		permission := new(UserPermission)
		if err = d.db.ScanRows(rows, permission); err != nil {
			return nil, errors.Wrapf(err, "scan user permission failed")
		}
		result = append(result, permission)
	}
	return result, nil
}

// ListResourceUsers list resource's auth user
func (d *driver) ListResourceUsers(project, resourceType string, resourceNames []string) ([]*UserPermission, error) {
	rows, err := d.db.Table(tableUserPermission).Where("project = ?", project).
		Where("resourceName IN (?)", resourceNames).
		Where("resourceType = ?", resourceType).Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "query user permissions failed")
	}
	defer rows.Close() // nolint

	result := make([]*UserPermission, 0)
	for rows.Next() {
		permission := new(UserPermission)
		if err = d.db.ScanRows(rows, permission); err != nil {
			return nil, errors.Wrapf(err, "scan user permission failed")
		}
		result = append(result, permission)
	}
	return result, nil
}

// GetUserPermission get user permission with resource
func (d *driver) GetUserPermission(permission *UserPermission) (*UserPermission, error) {
	rows, err := d.db.Table(tableUserPermission).Where("project = ?").
		Where("user = ?", permission.User).
		Where("resourceType = ?", permission.ResourceType).
		Where("resourceName = ?", permission.ResourceName).
		Where("resourceAction = ?", permission.ResourceAction).Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "get user permissions failed")
	}
	defer rows.Close() // nolint

	result := make([]*UserPermission, 0)
	for rows.Next() {
		obj := new(UserPermission)
		if err = d.db.ScanRows(rows, obj); err != nil {
			return nil, errors.Wrapf(err, "scan user permission failed")
		}
		result = append(result, obj)
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result[0], nil
}

// SaveAuditMessage save audit message
func (d *driver) SaveAuditMessage(audit *UserAudit) error {
	if err := d.db.Table(tableUserAudit).Save(audit).Error; err != nil {
		return errors.Wrapf(err, "save user audit failed")
	}
	return nil
}

// QueryUserAudits query the user audits
func (d *driver) QueryUserAudits(query *UserAuditQuery) ([]*UserAudit, error) {
	dbQuery := d.db.Table(tableUserAudit).Where("project IN (?)", query.Projects)
	if len(query.Users) != 0 {
		dbQuery.Where("user IN (?)", query.Users)
	}
	if len(query.Actions) != 0 {
		dbQuery.Where("action IN (?)", query.Actions)
	}
	if len(query.ResourceTypes) != 0 {
		dbQuery.Where("resourceType IN (?)", query.ResourceTypes)
	}
	if len(query.ResourceNames) != 0 {
		dbQuery.Where("resourceName IN (?)", query.ResourceNames)
	}
	if len(query.RequestIDs) != 0 {
		dbQuery.Where("requestID IN (?)", query.RequestIDs)
	}
	if query.RequestURI != "" {
		dbQuery.Where("requestURI LIKE ?", "%"+query.RequestURI+"%")
	}
	if query.RequestType != "" {
		dbQuery.Where("requestType = ?", query.RequestType)
	}
	if query.RequestMethod != "" {
		dbQuery.Where("requestMethod = ?", query.RequestMethod)
	}
	if query.StartTime != "" {
		dbQuery.Where("startTime >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		dbQuery.Where("endTime <>>= ?", query.EndTime)
	}
	audits := make([]*UserAudit, 0)
	if err := dbQuery.Order("id desc").Offset(query.Offset).Limit(query.Limit).Find(&audits).Error; err != nil {
		return nil, errors.Wrapf(err, "query user audits failed")
	}
	return audits, nil
}

// UpdateAppSetClusterScope update appset cluster scope
func (d *driver) UpdateAppSetClusterScope(appSet, clusters string) error {
	scope, err := d.GetAppSetClusterScope(appSet)
	if err != nil {
		return errors.Wrapf(err, "get appset cluster scope failed")
	}
	if scope == nil {
		if err = d.db.Table(tableAppSetClusterScope).Save(&AppSetClusterScope{
			AppSetName: appSet,
			Clusters:   clusters,
			UpdateTime: time.Now(),
		}).Error; err != nil {
			return errors.Wrapf(err, "save appset cluster scope failed")
		}
		return nil
	}
	if err = d.rateClient.Table(tableAppSetClusterScope).Where("id = ?", scope.ID).UpdateColumns(
		map[string]interface{}{
			"clusters":   clusters,
			"updateTime": time.Now(),
		}).Error; err != nil {
		return errors.Wrapf(err, "update appset's cluster scope failed")
	}
	return nil
}

// GetAppSetClusterScope get appSet's cluster scope with appset name
func (d *driver) GetAppSetClusterScope(appSet string) (*AppSetClusterScope, error) {
	rows, err := d.db.Table(tableAppSetClusterScope).Where("appSetName = ?", appSet).Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "get appset's cluster scope failed")
	}
	defer rows.Close() // nolint

	result := make([]*AppSetClusterScope, 0)
	for rows.Next() {
		obj := new(AppSetClusterScope)
		if err = d.db.ScanRows(rows, obj); err != nil {
			return nil, errors.Wrapf(err, "scan user permission failed")
		}
		result = append(result, obj)
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result[0], nil
}

// DeleteAppSetClusterScope delete appSet's cluster scope with appset name
func (d *driver) DeleteAppSetClusterScope(appSet string) error {
	if err := d.db.Table(tableAppSetClusterScope).Where("appSetName = ?", appSet).
		Delete(&AppSetClusterScope{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return errors.Wrapf(err, "delete appset's cluster scope failed")
	}
	return nil
}
