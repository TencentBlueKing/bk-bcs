/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package database

import (
	"time"

	"gorm.io/gorm"
)

var (
	// RECORDNOTFOUND returns a "record not found error".
	// Occurs only when attempting to query the database with a struct,
	// querying with a slice won't return this error.
	RECORDNOTFOUND = gorm.ErrRecordNotFound
)

const (
	// BSCPDB is bk-bscp system main database name.
	BSCPDB = "bscpdb"

	// BSCPDEFAULTSHARDINGDBID is bk-bscp default sharding database id.
	BSCPDEFAULTSHARDINGDBID = "default"

	// BSCPDEFAULTSHARDINGDB is bk-bscp default sharding database name.
	BSCPDEFAULTSHARDINGDB = "bscp_default"

	// BSCPCHARSET is bk-bscp database default charset.
	BSCPCHARSET = "utf8mb4"

	// BSCPINNODBENGINETYPE is bk-bscp database InnoDB engine type.
	BSCPINNODBENGINETYPE = "InnoDB"

	// BSCPDEFAULTENGINETYPE is bk-bscp database default engine type.
	BSCPDEFAULTENGINETYPE = BSCPINNODBENGINETYPE

	// BSCPDEFAULTVARGROUP is bk-bscp default variable group.
	BSCPDEFAULTVARGROUP = "default"

	// BSCPEMPTY is bk-bscp database empty limit.
	BSCPEMPTY = 0

	// BSCPNOTEMPTY is bk-bscp database not empty limit.
	BSCPNOTEMPTY = 1

	// BSCPNORMALSTRLENLIMIT is bk-bscp normal string length limit.
	BSCPNORMALSTRLENLIMIT = 32

	// BSCPLONGSTRLENLIMIT is bk-bscp long string length limit.
	BSCPLONGSTRLENLIMIT = 64

	// BSCPIDLENLIMIT is bk-bscp normal ID length limit.
	BSCPIDLENLIMIT = 64

	// BSCPCREATEBATCHLIMIT is bk-bscp batch mode create num limit.
	BSCPCREATEBATCHLIMIT = 500

	// BSCPCONTENTIDLENLIMIT is bk-bscp content id length limit.
	BSCPCONTENTIDLENLIMIT = 64

	// BSCPNAMELENLIMIT is bk-bscp normal name length limit.
	BSCPNAMELENLIMIT = 64

	// BSCPVARVALUESIZELIMIT is bk-bscp variable value size limit.
	BSCPVARVALUESIZELIMIT = 2048

	// BSCPQUERYLIMIT is bk-bscp batch query count limit.
	BSCPQUERYLIMIT = 500

	// BSCPQUERYLIMITLB is bk-bscp batch query count limit for little batch.
	BSCPQUERYLIMITLB = 100

	// BSCPQUERYLIMITMB is bk-bscp batch query count limit for much batch.
	BSCPQUERYLIMITMB = 1000

	// BSCPQUERYNEWESTLIMIT is bk-bscp batch query count limit for newest release.
	BSCPQUERYNEWESTLIMIT = 500

	// BSCPSTRATEGYCONTENTSIZELIMIT is bk-bscp strategy content size limit.
	BSCPSTRATEGYCONTENTSIZELIMIT = 1024 * 1024

	// BSCPLABELSSIZELIMIT is bk-bscp app instance labels size limit.
	BSCPLABELSSIZELIMIT = 8192

	// BSCPCFGFPATHLENLIMIT is bk-bscp config fpath length limit.
	BSCPCFGFPATHLENLIMIT = 256

	// BSCPCFGCACHEPATHLENLIMIT is bk-bscp config cache path length limit.
	BSCPCFGCACHEPATHLENLIMIT = 256

	// BSCPERRMSGLENLIMIT is bk-bscp error message length limit.
	BSCPERRMSGLENLIMIT = 256

	// BSCPEFFECTRELOADERRLENLIMIT is bk-bscp effect reload op error message length limit.
	BSCPEFFECTRELOADERRLENLIMIT = 256

	// BSCPTEMPLATEVARSLENLIMIT is bk-bscp template variables length limit.
	BSCPTEMPLATEVARSLENLIMIT = 1024 * 1024
)

// NOTE: must sync "bk-bscp/scripts/sql/*.sql" when the database updated for new package version to install and upgrade.

// Table is bscp database table definition.
type Table interface {
	// TableName returns the name of table.
	TableName() string

	// DBEngineType returns the db engine type of the table.
	DBEngineType() string
}

// LocalAuth is definition for t_local_auth.
type LocalAuth struct {
	ID    uint   `gorm:"column:Fid;primaryKey;autoIncrement"`
	PType string `gorm:"column:Fp_type;type:varchar(64);not null;uniqueIndex:uidx_policy"`
	V0    string `gorm:"column:Fv0;type:varchar(64);not null;uniqueIndex:uidx_policy"`
	V1    string `gorm:"column:Fv1;type:varchar(64);not null;uniqueIndex:uidx_policy"`
	V2    string `gorm:"column:Fv2;type:varchar(64);not null;uniqueIndex:uidx_policy"`
	V3    string `gorm:"column:Fv3;type:varchar(64);not null;uniqueIndex:uidx_policy"`
	V4    string `gorm:"column:Fv4;type:varchar(64);not null;uniqueIndex:uidx_policy"`
	V5    string `gorm:"column:Fv5;type:varchar(64);not null;uniqueIndex:uidx_policy"`
}

// TableName returns table name of t_local_auth.
func (l *LocalAuth) TableName() string {
	return "t_local_auth"
}

// DBEngineType returns the db engine type of the table.
func (l *LocalAuth) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// System is definition for t_system
type System struct {
	ID             uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	CurrentVersion string    `gorm:"column:Fcurrent_version;type:varchar(64);not null;index:idx_cversion"`
	Kind           string    `gorm:"column:Fkind;type:varchar(64);not null;uniqueIndex:uidx_kind"`
	Operator       string    `gorm:"column:Foperator;type:varchar(64);index:idx_operator"`
	CreatedAt      time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt      time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_system.
func (s *System) TableName() string {
	return "t_system"
}

// DBEngineType returns the db engine type of the table.
func (s *System) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// App is definition for t_application.
type App struct {
	ID           uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	AppID        string    `gorm:"column:Fapp_id;type:varchar(64);not null;uniqueIndex"`
	BizID        string    `gorm:"column:Fbiz_id;type:varchar(64);not null;uniqueIndex:uidx_bizidname"`
	Name         string    `gorm:"column:Fname;type:varchar(64);not null;uniqueIndex:uidx_bizidname"`
	DeployType   int32     `gorm:"column:Fdeploy_type"`
	Creator      string    `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string    `gorm:"column:Flast_modify_by;type:varchar(64);not null default 0"`
	Memo         string    `gorm:"column:Fmemo;type:varchar(128);not null default 0"`
	State        int32     `gorm:"column:Fstate;index:idx_state"`
	CreatedAt    time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_app.
func (a *App) TableName() string {
	return "t_application"
}

// DBEngineType returns the db engine type of the table.
func (a *App) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// TemplateBind is definition for t_template_bind.
type TemplateBind struct {
	ID           uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	BizID        string    `gorm:"column:Fbiz_id;type:varchar(64);not null;index:idx_bizid"`
	TemplateID   string    `gorm:"column:Ftemplate_id;type:varchar(64);not null;index:idx_tplid;uniqueIndex:uidx_bind"`
	AppID        string    `gorm:"column:Fapp_id;type:varchar(64);not null;index:idx_appid;uniqueIndex:uidx_bind"`
	CfgID        string    `gorm:"column:Fcfg_id;type:varchar(64);not null;uniqueIndex"`
	State        int32     `gorm:"column:Fstate;index:idx_state"`
	Creator      string    `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string    `gorm:"column:Flast_modify_by;type:varchar(64);not null default 0"`
	CreatedAt    time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_template_bind.
func (p *TemplateBind) TableName() string {
	return "t_template_bind"
}

// DBEngineType returns the db engine type of the table.
func (p *TemplateBind) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// ConfigTemplate definition for t_template.
type ConfigTemplate struct {
	ID            uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	TemplateID    string    `gorm:"column:Ftemplate_id;type:varchar(64);not null;uniqueIndex"`
	BizID         string    `gorm:"column:Fbiz_id;type:varchar(64);not null;index:idx_bizid;uniqueIndex:uidx_bizidname"`
	Name          string    `gorm:"column:Fname;type:varchar(64);not null;uniqueIndex:uidx_bizidname"`
	CfgName       string    `gorm:"column:Fcfg_name;type:varchar(64);not null"`
	CfgFpath      string    `gorm:"column:Fcfg_fpath;type:varchar(256);not null"`
	CfgType       int32     `gorm:"column:Fcfg_type;not null default 0"`
	User          string    `gorm:"column:Fuser;type:varchar(64);not null"`
	UserGroup     string    `gorm:"column:Fuser_group;type:varchar(64);not null"`
	FilePrivilege string    `gorm:"column:Ffile_privilege;type:varchar(64);not null"`
	FileFormat    string    `gorm:"column:Ffile_format;type:varchar(64);not null"`
	FileMode      int32     `gorm:"column:Ffile_mode"`
	EngineType    int32     `gorm:"column:Fengine_type;not null default 0"`
	Memo          string    `gorm:"column:Fmemo;type:varchar(128);not null default 0"`
	Creator       string    `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy  string    `gorm:"column:Flast_modify_by;type:varchar(64);not null default 0"`
	State         int32     `gorm:"column:Fstate;index:idx_state"`
	CreatedAt     time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt     time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_template.
func (c *ConfigTemplate) TableName() string {
	return "t_template"
}

// DBEngineType returns the db engine type of the table.
func (c *ConfigTemplate) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// ConfigTemplateVersion table name of t_template_version.
type ConfigTemplateVersion struct {
	ID           uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	VersionID    string    `gorm:"column:Fversion_id;type:varchar(64);not null;uniqueIndex"`
	BizID        string    `gorm:"column:Fbiz_id;type:varchar(64);not null;index:idx_bizid"`
	TemplateID   string    `gorm:"column:Ftemplate_id;type:varchar(64);not null;uniqueIndex:uidx_version"`
	VersionTag   string    `gorm:"column:Fversion_tag;type:varchar(64);not null;uniqueIndex:uidx_version"`
	ContentID    string    `gorm:"column:Fcontent_id;type:varchar(64);not null"`
	ContentSize  uint64    `gorm:"column:Fcontent_size"`
	Memo         string    `gorm:"column:Fmemo;type:varchar(128);not null default 0"`
	Creator      string    `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string    `gorm:"column:Flast_modify_by;type:varchar(64);not null default 0"`
	State        int32     `gorm:"column:Fstate;index:idx_state"`
	CreatedAt    time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_template_version.
func (c *ConfigTemplateVersion) TableName() string {
	return "t_template_version"
}

// DBEngineType returns the db engine type of the table.
func (c *ConfigTemplateVersion) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// VariableGroup table name of t_variable_group.
type VariableGroup struct {
	ID           uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	VarGroupID   string    `gorm:"column:Fvar_group_id;type:varchar(64);not null;uniqueIndex"`
	BizID        string    `gorm:"column:Fbiz_id;type:varchar(64);not null;uniqueIndex:uidx_bizidname"`
	Name         string    `gorm:"column:Fname;type:varchar(64);not null;uniqueIndex:uidx_bizidname"`
	Memo         string    `gorm:"column:Fmemo;type:varchar(128);not null default 0"`
	Creator      string    `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string    `gorm:"column:Flast_modify_by;type:varchar(64);not null default 0"`
	State        int32     `gorm:"column:Fstate;index:idx_state"`
	CreatedAt    time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_variable_group.
func (vg *VariableGroup) TableName() string {
	return "t_variable_group"
}

// DBEngineType returns the db engine type of the table.
func (vg *VariableGroup) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// Variable table name of t_variable.
type Variable struct {
	ID           uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	VarID        string    `gorm:"column:Fvar_id;type:varchar(64);not null;uniqueIndex"`
	BizID        string    `gorm:"column:Fbiz_id;type:varchar(64);not null;uniqueIndex:uidx_unionids"`
	VarGroupID   string    `gorm:"column:Fvar_group_id;type:varchar(64);not null;uniqueIndex:uidx_unionids"`
	Name         string    `gorm:"column:Fname;type:varchar(64);not null;uniqueIndex:uidx_unionids"`
	Value        string    `gorm:"column:Fvalue;type:longtext;not null"`
	Memo         string    `gorm:"column:Fmemo;type:varchar(128);not null default 0"`
	Creator      string    `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string    `gorm:"column:Flast_modify_by;type:varchar(64);not null default 0"`
	State        int32     `gorm:"column:Fstate;index:idx_state"`
	CreatedAt    time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_variable.
func (v *Variable) TableName() string {
	return "t_variable"
}

// DBEngineType returns the db engine type of the table.
func (v *Variable) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// Config is definition for t_config.
type Config struct {
	ID            uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	CfgID         string    `gorm:"column:Fcfg_id;type:varchar(64);not null;uniqueIndex"`
	BizID         string    `gorm:"column:Fbiz_id;type:varchar(64);not null;index:idx_bizid"`
	AppID         string    `gorm:"column:Fapp_id;type:varchar(64);not null;uniqueIndex:uidx_appidnamepath"`
	Name          string    `gorm:"column:Fname;type:varchar(64);not null;uniqueIndex:uidx_appidnamepath"`
	Fpath         string    `gorm:"column:Ffpath;type:varchar(256);not null;uniqueIndex:uidx_appidnamepath"`
	Type          int32     `gorm:"column:Ftype;not null default 0"`
	User          string    `gorm:"column:Fuser;type:varchar(64);not null"`
	UserGroup     string    `gorm:"column:Fuser_group;type:varchar(64);not null"`
	FilePrivilege string    `gorm:"column:Ffile_privilege;type:varchar(64);not null"`
	FileFormat    string    `gorm:"column:Ffile_format;type:varchar(64);not null"`
	FileMode      int32     `gorm:"column:Ffile_mode"`
	Creator       string    `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy  string    `gorm:"column:Flast_modify_by;type:varchar(64);not null default 0"`
	Memo          string    `gorm:"column:Fmemo;type:varchar(128);not null default 0"`
	State         int32     `gorm:"column:Fstate;index:idx_state"`
	CreatedAt     time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt     time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_config.
func (c *Config) TableName() string {
	return "t_config"
}

// DBEngineType returns the db engine type of the table.
func (c *Config) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// Commit is definition for t_commit.
type Commit struct {
	ID            uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	CommitID      string    `gorm:"column:Fcommit_id;type:varchar(64);not null;uniqueIndex"`
	BizID         string    `gorm:"column:Fbiz_id;type:varchar(64);not null;index:idx_bizid"`
	AppID         string    `gorm:"column:Fapp_id;type:varchar(64);not null;index:idx_appid"`
	CfgID         string    `gorm:"column:Fcfg_id;type:varchar(64);not null;index:idx_cfgid;uniqueIndex:uidx_multi"`
	CommitMode    int32     `gorm:"column:Fcommit_mode"`
	Operator      string    `gorm:"column:Foperator;type:varchar(64);not null;index:idx_operator"`
	ReleaseID     string    `gorm:"column:Frelease_id;type:varchar(64);not null;index:idx_releaseid"`
	MultiCommitID string    `gorm:"column:Fmulti_commit_id;type:varchar(64);not null;index:idx_mcommitid;uniqueIndex:uidx_multi"`
	Memo          string    `gorm:"column:Fmemo;type:varchar(128);not null default 0"`
	State         int32     `gorm:"column:Fstate;index:idx_state"`
	CreatedAt     time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt     time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_commit.
func (c *Commit) TableName() string {
	return "t_commit"
}

// DBEngineType returns the db engine type of the table.
func (c *Commit) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// Content is definition for t_content.
type Content struct {
	ID           uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	BizID        string    `gorm:"column:Fbiz_id;type:varchar(64);not null;index:idx_unionids"`
	AppID        string    `gorm:"column:Fapp_id;type:varchar(64);not null;index:idx_unionids"`
	CfgID        string    `gorm:"column:Fcfg_id;type:varchar(64);not null;index:idx_unionids"`
	CommitID     string    `gorm:"column:Fcommit_id;type:varchar(64);not null;index:idx_unionids"`
	Index        string    `gorm:"column:Findex;type:longtext;not null default 0"`
	ContentID    string    `gorm:"column:Fcontent_id;type:varchar(64);not null"`
	ContentSize  uint64    `gorm:"column:Fcontent_size"`
	Creator      string    `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string    `gorm:"column:Flast_modify_by;type:varchar(64);not null default 0"`
	Memo         string    `gorm:"column:Fmemo;type:varchar(128);not null default 0"`
	State        int32     `gorm:"column:Fstate;index:idx_state"`
	CreatedAt    time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_content.
func (c *Content) TableName() string {
	return "t_content"
}

// DBEngineType returns the db engine type of the table.
func (c *Content) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// MultiCommit is definition for t_multi_commit.
type MultiCommit struct {
	ID             uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	MultiCommitID  string    `gorm:"column:Fmulti_commit_id;type:varchar(64);not null;uniqueIndex"`
	BizID          string    `gorm:"column:Fbiz_id;type:varchar(64);not null;index:idx_bid"`
	AppID          string    `gorm:"column:Fapp_id;type:varchar(64);not null;index:idx_appid"`
	Operator       string    `gorm:"column:Foperator;type:varchar(64);not null;index:idx_operator"`
	MultiReleaseID string    `gorm:"column:Fmulti_release_id;type:varchar(64);not null;index:idx_releaseid"`
	Memo           string    `gorm:"column:Fmemo;type:varchar(128);not null default 0"`
	State          int32     `gorm:"column:Fstate;index:idx_state"`
	CreatedAt      time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt      time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_multi_commit.
func (mc *MultiCommit) TableName() string {
	return "t_multi_commit"
}

// DBEngineType returns the db engine type of the table.
func (mc *MultiCommit) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// AppInstance is definition for t_app_instance.
type AppInstance struct {
	ID        uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	BizID     string    `gorm:"column:Fbiz_id;type:varchar(64);not null;uniqueIndex:uidx_unionids"`
	AppID     string    `gorm:"column:Fapp_id;type:varchar(64);not null;uniqueIndex:uidx_unionids"`
	CloudID   string    `gorm:"column:Fcloud_id;type:varchar(64);not null;uniqueIndex:uidx_unionids"`
	IP        string    `gorm:"column:Fip;type:varchar(32);not null;uniqueIndex:uidx_unionids"`
	Path      string    `gorm:"column:Fpath;type:varchar(256);not null;uniqueIndex:uidx_unionids"`
	Labels    string    `gorm:"column:Flabels;type:longtext;not null default 0"`
	State     int32     `gorm:"column:Fstate;index:idx_state"`
	CreatedAt time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_app_instance.
func (a *AppInstance) TableName() string {
	return "t_app_instance"
}

// DBEngineType returns the db engine type of the table.
func (a *AppInstance) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// AppInstanceRelease is definition for t_app_instance_release.
type AppInstanceRelease struct {
	ID         uint64     `gorm:"column:Fid;primaryKey;autoIncrement"`
	InstanceID uint64     `gorm:"column:Finstance_id;type:bigint(20);not null;uniqueIndex:uidx_unionids"`
	BizID      string     `gorm:"column:Fbiz_id;type:varchar(64);not null"`
	AppID      string     `gorm:"column:Fapp_id;type:varchar(64);not null"`
	CfgID      string     `gorm:"column:Fcfg_id;type:varchar(64);not null;uniqueIndex:uidx_unionids;index:idx_effected"`
	ReleaseID  string     `gorm:"column:Frelease_id;type:varchar(64);not null;uniqueIndex:uidx_unionids;index:idx_effected"`
	EffectTime *time.Time `gorm:"column:Feffect_time;default null"`
	EffectCode int32      `gorm:"column:Feffect_code;not null default 0"`
	EffectMsg  string     `gorm:"column:Feffect_msg;type:varchar(128);not null default 0"`
	ReloadTime *time.Time `gorm:"column:Freload_time;default null"`
	ReloadCode int32      `gorm:"column:Freload_code;not null default 0"`
	ReloadMsg  string     `gorm:"column:Freload_msg;type:varchar(128);not null default 0"`
	CreatedAt  time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt  time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_app_instance_release.
func (a *AppInstanceRelease) TableName() string {
	return "t_app_instance_release"
}

// DBEngineType returns the db engine type of the table.
func (a *AppInstanceRelease) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// Release is definition for t_release.
type Release struct {
	ID             uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	ReleaseID      string    `gorm:"column:Frelease_id;type:varchar(64);not null;uniqueIndex"`
	Name           string    `gorm:"column:Fname;type:varchar(64);not null;index:idx_name"`
	BizID          string    `gorm:"column:Fbiz_id;type:varchar(64);not null;index:idx_bid"`
	AppID          string    `gorm:"column:Fapp_id;type:varchar(64);not null;index:idx_appid"`
	CfgID          string    `gorm:"column:Fcfg_id;type:varchar(64);not null;index:idx_cfgid"`
	CfgName        string    `gorm:"column:Fcfg_name;type:varchar(64);not null"`
	CfgFpath       string    `gorm:"column:Fcfg_fpath;type:varchar(256);not null"`
	CfgType        int32     `gorm:"column:Fcfg_type;not null default 0"`
	User           string    `gorm:"column:Fuser;type:varchar(64);not null"`
	UserGroup      string    `gorm:"column:Fuser_group;type:varchar(64);not null"`
	FilePrivilege  string    `gorm:"column:Ffile_privilege;type:varchar(64);not null"`
	FileFormat     string    `gorm:"column:Ffile_format;type:varchar(64);not null"`
	FileMode       int32     `gorm:"column:Ffile_mode"`
	StrategyID     string    `gorm:"column:Fstrategy_id;type:varchar(64);not null"`
	Strategies     string    `gorm:"column:Fstrategies;type:longtext;not null default 0"`
	Creator        string    `gorm:"column:Fcreator;type:varchar(64);not null;index:idx_creator"`
	CommitID       string    `gorm:"column:Fcommit_id;type:varchar(64);not null"`
	MultiReleaseID string    `gorm:"column:Fmulti_release_id;type:varchar(64);not null default 0;index:idx_mreleaseid"`
	LastModifyBy   string    `gorm:"column:Flast_modify_by;type:varchar(64);not null default 0"`
	Memo           string    `gorm:"column:Fmemo;type:varchar(128);not null default 0"`
	State          int32     `gorm:"column:Fstate;index:idx_state"`
	CreatedAt      time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt      time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_release.
func (r *Release) TableName() string {
	return "t_release"
}

// DBEngineType returns the db engine type of the table.
func (r *Release) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// MultiRelease is definition for t_multi_release.
type MultiRelease struct {
	ID             uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	MultiReleaseID string    `gorm:"column:Fmulti_release_id;type:varchar(64);not null;uniqueIndex"`
	Name           string    `gorm:"column:Fname;type:varchar(64);not null;index:idx_name"`
	BizID          string    `gorm:"column:Fbiz_id;type:varchar(64);not null;index:idx_bid"`
	AppID          string    `gorm:"column:Fapp_id;type:varchar(64);not null;index:idx_appid"`
	StrategyID     string    `gorm:"column:Fstrategy_id;type:varchar(64);not null"`
	Strategies     string    `gorm:"column:Fstrategies;type:longtext;not null default 0"`
	Creator        string    `gorm:"column:Fcreator;type:varchar(64);not null;index:idx_creator"`
	MultiCommitID  string    `gorm:"column:Fmulti_commit_id;type:varchar(64);not null"`
	LastModifyBy   string    `gorm:"column:Flast_modify_by;type:varchar(64);not null default 0"`
	Memo           string    `gorm:"column:Fmemo;type:varchar(128);not null default 0"`
	State          int32     `gorm:"column:Fstate;index:idx_state"`
	CreatedAt      time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt      time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_multi_release.
func (mr *MultiRelease) TableName() string {
	return "t_multi_release"
}

// DBEngineType returns the db engine type of the table.
func (mr *MultiRelease) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// Strategy is definition for t_strategy.
type Strategy struct {
	ID           uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	StrategyID   string    `gorm:"column:Fstrategy_id;type:varchar(64);not null;uniqueIndex"`
	BizID        string    `gorm:"column:Fbiz_id;type:varchar(64);not null;index:idx_bid"`
	AppID        string    `gorm:"column:Fapp_id;type:varchar(64);not null;uniqueIndex:uidx_strategy"`
	Name         string    `gorm:"column:Fname;type:varchar(64);not null;uniqueIndex:uidx_strategy"`
	Content      string    `gorm:"column:Fcontent;type:longtext;not null"`
	Creator      string    `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string    `gorm:"column:Flast_modify_by;type:varchar(64);not null default 0"`
	Memo         string    `gorm:"column:Fmemo;type:varchar(128);not null default 0"`
	State        int32     `gorm:"column:Fstate;index:idx_state"`
	CreatedAt    time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_strategy.
func (s *Strategy) TableName() string {
	return "t_strategy"
}

// DBEngineType returns the db engine type of the table.
func (s *Strategy) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// ProcAttr is definition for t_proc_attr.
type ProcAttr struct {
	ID           uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	CloudID      string    `gorm:"column:Fcloud_id;type:varchar(64);not null;uniqueIndex:uidx_attr"`
	IP           string    `gorm:"column:Fip;type:varchar(32);not null;uniqueIndex:uidx_attr"`
	BizID        string    `gorm:"column:Fbiz_id;type:varchar(64);not null;uniqueIndex:uidx_attr;index:idx_bizapp"`
	AppID        string    `gorm:"column:Fapp_id;type:varchar(64);not null;uniqueIndex:uidx_attr;index:idx_bizapp"`
	Path         string    `gorm:"column:Fpath;type:varchar(256);not null;uniqueIndex:uidx_attr"`
	Labels       string    `gorm:"column:Flabels;type:longtext;not null default 0"`
	Creator      string    `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string    `gorm:"column:Flast_modify_by;type:varchar(64);not null default 0"`
	Memo         string    `gorm:"column:Fmemo;type:varchar(128);not null default 0"`
	State        int32     `gorm:"column:Fstate;index:idx_state"`
	CreatedAt    time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_proc_attr.
func (p *ProcAttr) TableName() string {
	return "t_proc_attr"
}

// DBEngineType returns the db engine type of the table.
func (p *ProcAttr) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// Sharding is definition fot t_sharding.
type Sharding struct {
	ID        uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	Key       string    `gorm:"column:Fkey;type:varchar(64);not null;uniqueIndex"`
	DBID      string    `gorm:"column:Fdb_id;type:varchar(64);not null"`
	DBName    string    `gorm:"column:Fdb_name;type:varchar(64);not null"`
	Memo      string    `gorm:"column:Fmemo;type:varchar(128);not null default 0"`
	State     int32     `gorm:"column:Fstate"`
	CreatedAt time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_sharding.
func (s *Sharding) TableName() string {
	return "t_sharding"
}

// DBEngineType returns the db engine type of the table.
func (s *Sharding) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// ShardingDB is definition for t_sharding_db.
type ShardingDB struct {
	ID        uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	DBID      string    `gorm:"column:Fdb_id;type:varchar(64);not null;uniqueIndex"`
	Host      string    `gorm:"column:Fhost;type:varchar(64);not null"`
	Port      int32     `gorm:"column:Fport"`
	User      string    `gorm:"column:Fuser;type:varchar(32);not null"`
	Password  string    `gorm:"column:Fpassword;type:varchar(32);not null"`
	Memo      string    `gorm:"column:Fmemo;type:varchar(128);not null default 0"`
	State     int32     `gorm:"column:Fstate"`
	CreatedAt time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_sharding_db.
func (s *ShardingDB) TableName() string {
	return "t_sharding_db"
}

// DBEngineType returns the db engine type of the table.
func (s *ShardingDB) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}

// Audit is definition for t_audit.
type Audit struct {
	ID         uint64    `gorm:"column:Fid;primaryKey;autoIncrement"`
	SourceType int32     `gorm:"column:Fsource_type;index:idx_sourcetype"`
	OpType     int32     `gorm:"column:Fop_type;index:idx_optype"`
	BizID      string    `gorm:"column:Fbiz_id;type:varchar(64);not null;index:idx_bid"`
	SourceID   string    `gorm:"column:Fsource_id;type:varchar(64);not null;index:idx_sourceid"`
	Operator   string    `gorm:"column:Foperator;type:varchar(64);not null;index:idx_operator"`
	Memo       string    `gorm:"column:Fmemo;type:varchar(128);not null default 0"`
	State      int32     `gorm:"column:Fstate"`
	CreatedAt  time.Time `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt  time.Time `gorm:"column:Fupdate_time;index:idx_utime"`
}

// TableName returns table name of t_audit.
func (a *Audit) TableName() string {
	return "t_audit"
}

// DBEngineType returns the db engine type of the table.
func (a *Audit) DBEngineType() string {
	return BSCPDEFAULTENGINETYPE
}
