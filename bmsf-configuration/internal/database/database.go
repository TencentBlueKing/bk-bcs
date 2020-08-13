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
)

const (
	// BSCPSHARDINGDB is bk-bscp sharding database name.
	BSCPSHARDINGDB = "bscpdb"

	// BSCPCHARSET is bk-bscp database default charset.
	BSCPCHARSET = "utf8mb4"

	// BSCPCONFIGSSIZELIMIT is bk-bscp configs size limit.
	BSCPCONFIGSSIZELIMIT = 2 * 1024 * 1024

	// BSCPTPLSIZELIMIT is bk-bscp configs template size limit.
	BSCPTPLSIZELIMIT = BSCPCONFIGSSIZELIMIT

	// BSCPCFGCONTENTSIZELIMIT is bk-bscp configs content size limit(for template or normal configs).
	BSCPCFGCONTENTSSIZELIMIT = 2 * BSCPCONFIGSSIZELIMIT

	// BSCPTPLRULESSIZELIMIT is bk-bscp configs template rules size limit.
	BSCPTPLRULESSIZELIMIT = 1024 * 1024

	// BSCPCHANGESSIZELIMIT is bk-bscp configs changes size limit.
	BSCPCHANGESSIZELIMIT = BSCPCONFIGSSIZELIMIT

	// BSCPCFGLINKLENLIMIT is bk-bscp configs link length limit.
	BSCPCFGLINKLENLIMIT = 4096

	// BSCPNORMALSTRLENLIMIT is bk-bscp normal string length limit.
	BSCPNORMALSTRLENLIMIT = 32

	// BSCPLONGSTRLENLIMIT is bk-bscp long string length limit.
	BSCPLONGSTRLENLIMIT = 64

	// BSCPIDLENLIMIT is bk-bscp normal ID length limit.
	BSCPIDLENLIMIT = 64

	// BSCPAUTHLENLIMIT is bk-bscp auth info length limit.
	BSCPAUTHLENLIMIT = 64

	// BSCPNAMELENLIMIT is bk-bscp normal name length limit.
	BSCPNAMELENLIMIT = 64

	// BSCPQUERYLIMIT is bk-bscp batch query count limit.
	BSCPQUERYLIMIT = 100

	// BSCPQUERYLIMITLB is bk-bscp batch query count limit for little batch.
	BSCPQUERYLIMITLB = 10

	// BSCPBATCHLIMIT is bk-bscp batch mode common limit.
	BSCPBATCHLIMIT = 100

	// BSCPSTRATEGYCONTENTSIZELIMIT is bk-bscp strategy content size limit.
	BSCPSTRATEGYCONTENTSIZELIMIT = 10 * 1024

	// BSCPLABELSSIZELIMIT is bk-bscp app instance labels size limit.
	BSCPLABELSSIZELIMIT = 1024

	// BSCPCFGSETFPATHLENLIMIT is bk-bscp configset fpath length limit.
	BSCPCFGSETFPATHLENLIMIT = 128

	// BSCPERRMSGLENLIMIT is bk-bscp error message length limit.
	BSCPERRMSGLENLIMIT = 256

	// BSCPITGTPLSIZELIMIT is bk-bscp integration template file size limit.
	BSCPITGTPLSIZELIMIT = 4 * 1024 * 1024

	// BSCPTEMPLATEBINDINGPARAMSSIZELIMIT is bk-bscp template binding params size limit.
	BSCPTEMPLATEBINDINGPARAMSSIZELIMIT = 100 * 1024

	// BSCPTEMPLATEBINDINGNUMLIMIT is bk-bscp template binding rules number limit of one binding
	BSCPTEMPLATEBINDINGNUMLIMIT = 100

	// BSCPVARIABLEKEYLENGTHLIMIT is bk-bscp variable key length limit
	BSCPVARIABLEKEYLENGTHLIMIT = 256

	// BSCPVARIABLEVALUESIZELIMIT is bk-bscp variable value size limit
	BSCPVARIABLEVALUESIZELIMIT = 1024

	// BSCPFILEENCODINGLENGTHLIMIT is bk-bscp template file encoding name length limit
	BSCPFILEENCODINGLENGTHLIMIT = 16

	// BSCPFILEUSERLENGTHLIMIT is bk-bscp file user name length limit
	BSCPFILEUSERLENGTHLIMIT = 64

	// BSCPFILEGROUPLENGTHLIMIT is bk-bscp file group name length limit
	BSCPFILEGROUPLENGTHLIMIT = 64

	// BSCPCLUSTERLABELSLENLIMIT is bk-bscp cluster labels length limit.
	BSCPCLUSTERLABELSLENLIMIT = 128

	// BSCPEFFECTRELOADERRLENLIMIT is bk-bscp effect reload op error message length limit.
	BSCPEFFECTRELOADERRLENLIMIT = 128
)

// Table is bscp database table definition.
type Table interface {
	// TableName returns the name of table.
	TableName() string
}

// Business is definition for t_business.
type Business struct {
	ID           uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Bid          string     `gorm:"column:Fbid;type:varchar(64);not null;unique_index"`
	Name         string     `gorm:"column:Fname;type:varchar(64);not null;unique_index"`
	Depid        string     `gorm:"column:Fdepid;type:varchar(64);not null"`
	Creator      string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	Memo         string     `gorm:"column:Fmemo;type:varchar(64);not null"`
	Auth         string     `gorm:"column:Fauth;type:varchar(64);not null default ''"`
	State        int32      `gorm:"column:Fstate"`
	CreatedAt    time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt    *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_business.
func (b *Business) TableName() string {
	return "t_business"
}

// App is definition for t_application.
type App struct {
	ID           uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Appid        string     `gorm:"column:Fappid;type:varchar(64);not null;unique_index"`
	Bid          string     `gorm:"column:Fbid;type:varchar(64);not null;unique_index:idx_bidname"`
	Name         string     `gorm:"column:Fname;type:varchar(64);not null;unique_index:idx_bidname"`
	DeployType   int32      `gorm:"column:Fdeploy_type"`
	Creator      string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	Memo         string     `gorm:"column:Fmemo;type:varchar(64);not null"`
	State        int32      `gorm:"column:Fstate;index:idx_state"`
	CreatedAt    time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt    *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_app.
func (a *App) TableName() string {
	return "t_application"
}

// Cluster is definition for t_cluster.
type Cluster struct {
	ID           uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Clusterid    string     `gorm:"column:Fclusterid;type:varchar(64);not null;unique_index"`
	Bid          string     `gorm:"column:Fbid;type:varchar(64);not null;index:idx_bid"`
	Appid        string     `gorm:"column:Fappid;type:varchar(64);not null;unique_index:idx_appidnamelabels"`
	Name         string     `gorm:"column:Fname;type:varchar(64);not null;unique_index:idx_appidnamelabels"`
	Labels       string     `gorm:"column:Flabels;type:varchar(128);not null;unique_index:idx_appidnamelabels"`
	RClusterid   string     `gorm:"column:Frclusterid;type:varchar(64);not null;index:idx_rclusterid"`
	Creator      string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	Memo         string     `gorm:"column:Fmemo;type:varchar(64);not null"`
	State        int32      `gorm:"column:Fstate;index:idx_state"`
	CreatedAt    time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt    *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_cluster.
func (c *Cluster) TableName() string {
	return "t_cluster"
}

// Zone is definition for t_zone.
type Zone struct {
	ID           uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Zoneid       string     `gorm:"column:Fzoneid;type:varchar(64);not null;unique_index"`
	Bid          string     `gorm:"column:Fbid;type:varchar(64);not null;index:idx_bid"`
	Appid        string     `gorm:"column:Fappid;type:varchar(64);not null;unique_index:idx_appidname"`
	Name         string     `gorm:"column:Fname;type:varchar(64);not null;unique_index:idx_appidname"`
	Clusterid    string     `gorm:"column:Fclusterid;type:varchar(64);not null;index:idx_clusterid"`
	Creator      string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	Memo         string     `gorm:"column:Fmemo;type:varchar(64);not null"`
	State        int32      `gorm:"column:Fstate;index:idx_state"`
	CreatedAt    time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt    *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_zone.
func (z *Zone) TableName() string {
	return "t_zone"
}

// ConfigSet is definition for t_configset.
type ConfigSet struct {
	ID           uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Cfgsetid     string     `gorm:"column:Fcfgsetid;type:varchar(64);not null;unique_index"`
	Bid          string     `gorm:"column:Fbid;type:varchar(64);not null;index:idx_bid"`
	Appid        string     `gorm:"column:Fappid;type:varchar(64);not null;unique_index:idx_appidnamepath"`
	Name         string     `gorm:"column:Fname;type:varchar(64);not null;unique_index:idx_appidnamepath"`
	Fpath        string     `gorm:"column:Ffpath;type:varchar(128);not null;unique_index:idx_appidnamepath"`
	Creator      string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	Memo         string     `gorm:"column:Fmemo;type:varchar(64);not null"`
	State        int32      `gorm:"column:Fstate;index:idx_state"`
	CreatedAt    time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt    *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_configset.
func (c *ConfigSet) TableName() string {
	return "t_configset"
}

// ConfigSetLock is definition for t_configset_lock.
type ConfigSetLock struct {
	ID         uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Cfgsetid   string     `gorm:"column:Fcfgsetid;type:varchar(64);not null;unique_index"`
	Operator   string     `gorm:"column:Foperator;type:varchar(64);not null"`
	LockTime   time.Time  `gorm:"column:Flock_time"`
	UnlockTime time.Time  `gorm:"column:Funlock_time"`
	Memo       string     `gorm:"column:Fmemo;type:varchar(64);not null"`
	State      int32      `gorm:"column:Fstate"`
	CreatedAt  time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt  time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt  *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_configset_lock.
func (c *ConfigSetLock) TableName() string {
	return "t_configset_lock"
}

// Configs is definition for t_configs.
type Configs struct {
	ID           uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Bid          string     `gorm:"column:Fbid;type:varchar(64);not null;unique_index:idx_unionids"`
	Cfgsetid     string     `gorm:"column:Fcfgsetid;type:varchar(64);not null;unique_index:idx_unionids"`
	Commitid     string     `gorm:"column:Fcommitid;type:varchar(64);not null;unique_index:idx_unionids"`
	Appid        string     `gorm:"column:Fappid;type:varchar(64);not null;unique_index:idx_unionids"`
	Clusterid    string     `gorm:"column:Fclusterid;type:varchar(64);not null;unique_index:idx_unionids"`
	Zoneid       string     `gorm:"column:Fzoneid;type:varchar(64);not null;unique_index:idx_unionids"`
	Index        string     `gorm:"column:Findex;type:varchar(64);not null;unique_index:idx_unionids"`
	Cid          string     `gorm:"column:Fcid;type:varchar(64);not null"`
	CfgLink      string     `gorm:"column:Fcfglink;type:varchar(4096);not null"`
	Content      []byte     `gorm:"column:Fcontent;type:LongBlob;not null"`
	Creator      string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	Memo         string     `gorm:"column:Fmemo;type:varchar(64);not null"`
	State        int32      `gorm:"column:Fstate"`
	CreatedAt    time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt    *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_configs.
func (c *Configs) TableName() string {
	return "t_configs"
}

// Commit is definition for t_commit.
type Commit struct {
	ID            uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Commitid      string     `gorm:"column:Fcommitid;type:varchar(64);not null;unique_index"`
	Bid           string     `gorm:"column:Fbid;type:varchar(64);not null;index:idx_bid"`
	Appid         string     `gorm:"column:Fappid;type:varchar(64);not null;index:idx_appid"`
	Cfgsetid      string     `gorm:"column:Fcfgsetid;type:varchar(64);not null;unique_index:uidx_multi;index:idx_configsetid"`
	Templateid    string     `gorm:"column:Ftemplateid;type:varchar(64);not null"`
	Template      string     `gorm:"column:Ftemplate;type:longtext;not null"`
	TemplateRule  string     `gorm:"column:Ftemplate_rule;type:longtext;not null"`
	Op            int32      `gorm:"column:Fop"`
	Operator      string     `gorm:"column:Foperator;type:varchar(64);not null;index:idx_operator"`
	PrevConfigs   []byte     `gorm:"column:Fprev_configs;type:LongBlob;not null"`
	Configs       []byte     `gorm:"column:Fconfigs;type:LongBlob;not null"`
	Changes       string     `gorm:"column:Fchanges;type:longtext;not null"`
	Releaseid     string     `gorm:"column:Freleaseid;type:varchar(64);not null;index:idx_releaseid"`
	MultiCommitid string     `gorm:"column:Fmulti_commitid;type:varchar(64);not null;unique_index:uidx_multi;index:idx_mcommitid"`
	Memo          string     `gorm:"column:Fmemo;type:varchar(64);not null"`
	State         int32      `gorm:"column:Fstate;index:idx_state"`
	CreatedAt     time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt     time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt     *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_commit.
func (c *Commit) TableName() string {
	return "t_commit"
}

// MultiCommit is definition for t_multi_commit.
type MultiCommit struct {
	ID             uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	MultiCommitid  string     `gorm:"column:Fmulti_commitid;type:varchar(64);not null;unique_index"`
	Bid            string     `gorm:"column:Fbid;type:varchar(64);not null;index:idx_bid"`
	Appid          string     `gorm:"column:Fappid;type:varchar(64);not null;index:idx_appid"`
	Operator       string     `gorm:"column:Foperator;type:varchar(64);not null;index:idx_operator"`
	MultiReleaseid string     `gorm:"column:Fmulti_releaseid;type:varchar(64);not null;index:idx_releaseid"`
	Memo           string     `gorm:"column:Fmemo;type:varchar(64);not null"`
	State          int32      `gorm:"column:Fstate;index:idx_state"`
	CreatedAt      time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt      time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt      *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_multi_commit.
func (mc *MultiCommit) TableName() string {
	return "t_multi_commit"
}

// AppInstance is definition for t_app_instance.
type AppInstance struct {
	ID        uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Bid       string     `gorm:"column:Fbid;type:varchar(64);not null;unique_index:idx_unionids"`
	Appid     string     `gorm:"column:Fappid;type:varchar(64);not null;unique_index:idx_unionids"`
	Clusterid string     `gorm:"column:Fclusterid;type:varchar(64);not null;unique_index:idx_unionids"`
	Zoneid    string     `gorm:"column:Fzoneid;type:varchar(64);not null;unique_index:idx_unionids"`
	Dc        string     `gorm:"column:Fdc;type:varchar(64);not null;unique_index:idx_unionids"`
	IP        string     `gorm:"column:Fip;type:varchar(32);not null;unique_index:idx_unionids"`
	Labels    string     `gorm:"column:Flabels;type:longtext;not null"`
	State     int32      `gorm:"column:Fstate"`
	CreatedAt time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_app_instance.
func (a *AppInstance) TableName() string {
	return "t_app_instance"
}

// AppInstanceRelease is definition for t_app_instance_release.
type AppInstanceRelease struct {
	ID         uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Instanceid uint64     `gorm:"column:Finstanceid;type:bigint(20);not null;unique_index:idx_unionids"`
	Bid        string     `gorm:"column:Fbid;type:varchar(64);not null"`
	Appid      string     `gorm:"column:Fappid;type:varchar(64);not null"`
	Clusterid  string     `gorm:"column:Fclusterid;type:varchar(64);not null"`
	Zoneid     string     `gorm:"column:Fzoneid;type:varchar(64);not null"`
	Dc         string     `gorm:"column:Fdc;type:varchar(64);not null"`
	IP         string     `gorm:"column:Fip;type:varchar(32);not null"`
	Labels     string     `gorm:"column:Flabels;type:longtext;not null"`
	Cfgsetid   string     `gorm:"column:Fcfgsetid;type:varchar(64);not null;unique_index:idx_unionids;index:idx_effected"`
	Releaseid  string     `gorm:"column:Freleaseid;type:varchar(64);not null;unique_index:idx_unionids;index:idx_effected"`
	EffectTime *time.Time `gorm:"column:Feffect_time;default null"`
	EffectCode int32      `gorm:"column:Feffect_code;not null default 0"`
	EffectMsg  string     `gorm:"column:Feffect_msg;type:varchar(128);not null default ''"`
	ReloadTime *time.Time `gorm:"column:Freload_time;default null"`
	ReloadCode int32      `gorm:"column:Freload_code;not null default 0"`
	ReloadMsg  string     `gorm:"column:Freload_msg;type:varchar(128);not null default ''"`
	CreatedAt  time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt  time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt  *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_app_instance_release.
func (a *AppInstanceRelease) TableName() string {
	return "t_app_instance_release"
}

// Release is definition for t_release.
type Release struct {
	ID             uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Releaseid      string     `gorm:"column:Freleaseid;type:varchar(64);not null;unique_index"`
	Name           string     `gorm:"column:Fname;type:varchar(64);not null;index:idx_name"`
	Bid            string     `gorm:"column:Fbid;type:varchar(64);not null;index:idx_bid"`
	Appid          string     `gorm:"column:Fappid;type:varchar(64);not null;index:idx_appid"`
	Cfgsetid       string     `gorm:"column:Fcfgsetid;type:varchar(64);not null;index:idx_configsetid"`
	CfgsetName     string     `gorm:"column:Fcfgsetname;type:varchar(64);not null"`
	CfgsetFpath    string     `gorm:"column:Fcfgsetfpath;type:varchar(128);not null"`
	Strategyid     string     `gorm:"column:Fstrategyid;type:varchar(64);not null"`
	Strategies     string     `gorm:"column:Fstrategies;type:longtext;not null"`
	Creator        string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	Commitid       string     `gorm:"column:Fcommitid;type:varchar(64);not null"`
	MultiReleaseid string     `gorm:"column:Fmulti_releaseid;type:varchar(64);not null;index:idx_mreleaseid"`
	LastModifyBy   string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	Memo           string     `gorm:"column:Fmemo;type:varchar(64);not null"`
	State          int32      `gorm:"column:Fstate"`
	CreatedAt      time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt      time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt      *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_release.
func (r *Release) TableName() string {
	return "t_release"
}

// MultiRelease is definition for t_multi_release.
type MultiRelease struct {
	ID             uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	MultiReleaseid string     `gorm:"column:Fmulti_releaseid;type:varchar(64);not null;unique_index"`
	Name           string     `gorm:"column:Fname;type:varchar(64);not null;index:idx_name"`
	Bid            string     `gorm:"column:Fbid;type:varchar(64);not null;index:idx_bid"`
	Appid          string     `gorm:"column:Fappid;type:varchar(64);not null;index:idx_appid"`
	Strategyid     string     `gorm:"column:Fstrategyid;type:varchar(64);not null"`
	Strategies     string     `gorm:"column:Fstrategies;type:longtext;not null"`
	Creator        string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	MultiCommitid  string     `gorm:"column:Fmulti_commitid;type:varchar(64);not null"`
	LastModifyBy   string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	Memo           string     `gorm:"column:Fmemo;type:varchar(64);not null"`
	State          int32      `gorm:"column:Fstate"`
	CreatedAt      time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt      time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt      *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_multi_release.
func (mr *MultiRelease) TableName() string {
	return "t_multi_release"
}

// Strategy is definition for t_strategy.
type Strategy struct {
	ID           uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Strategyid   string     `gorm:"column:Fstrategyid;type:varchar(64);not null;unique_index"`
	Appid        string     `gorm:"column:Fappid;type:varchar(64);not null;unique_index:uidx_strategy"`
	Name         string     `gorm:"column:Fname;type:varchar(64);not null;unique_index:uidx_strategy"`
	Content      string     `gorm:"column:Fcontent;type:longtext;not null"`
	Creator      string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	Memo         string     `gorm:"column:Fmemo;type:varchar(64);not null"`
	State        int32      `gorm:"column:Fstate;index:idx_state"`
	CreatedAt    time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt    *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_strategy.
func (s *Strategy) TableName() string {
	return "t_strategy"
}

// ProcAttr is definition for t_proc_attr.
type ProcAttr struct {
	ID           uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Cloudid      string     `gorm:"column:Fcloudid;type:varchar(64);not null;unique_index:uidx_attr"`
	IP           string     `gorm:"column:Fip;type:varchar(32);not null;unique_index:uidx_attr"`
	Bid          string     `gorm:"column:Fbid;type:varchar(64);not null;unique_index:uidx_attr"`
	Appid        string     `gorm:"column:Fappid;type:varchar(64);not null;unique_index:uidx_attr"`
	Clusterid    string     `gorm:"column:Fclusterid;type:varchar(64);not null"`
	Zoneid       string     `gorm:"column:Fzoneid;type:varchar(64);not null"`
	Dc           string     `gorm:"column:Fdc;type:varchar(64);not null"`
	Labels       string     `gorm:"column:Flabels;type:longtext;not null"`
	Path         string     `gorm:"column:Fpath;type:varchar(128);not null;unique_index:uidx_attr"`
	Creator      string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	Memo         string     `gorm:"column:Fmemo;type:varchar(64);not null"`
	State        int32      `gorm:"column:Fstate;index:idx_state"`
	CreatedAt    time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt    *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_proc_attr.
func (p *ProcAttr) TableName() string {
	return "t_proc_attr"
}

// Sharding is definition fot t_sharding.
type Sharding struct {
	ID        uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Key       string     `gorm:"column:Fkey;type:varchar(64);not null;unique_index"`
	DBid      string     `gorm:"column:Fdbid;type:varchar(64);not null"`
	DBName    string     `gorm:"column:Fdbname;type:varchar(64);not null"`
	Memo      string     `gorm:"column:Fmemo;type:varchar(64);not null"`
	State     int32      `gorm:"column:Fstate"`
	CreatedAt time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_sharding.
func (s *Sharding) TableName() string {
	return "t_sharding"
}

// ShardingDB is definition for t_sharding_db.
type ShardingDB struct {
	ID        uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	DBid      string     `gorm:"column:Fdbid;type:varchar(64);not null;unique_index"`
	Host      string     `gorm:"column:Fhost;type:varchar(64);not null"`
	Port      int32      `gorm:"column:Fport"`
	User      string     `gorm:"column:Fuser;type:varchar(32);not null"`
	Password  string     `gorm:"column:Fpassword;type:varchar(32);not null"`
	Memo      string     `gorm:"column:Fmemo;type:varchar(64);not null"`
	State     int32      `gorm:"column:Fstate"`
	CreatedAt time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_sharding_db.
func (s *ShardingDB) TableName() string {
	return "t_sharding_db"
}

// Audit is definition for t_audit.
type Audit struct {
	ID         uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	SourceType int32      `gorm:"column:Fsource_type;index:idx_sourcetype"`
	OpType     int32      `gorm:"column:Fop_type;index:idx_optype"`
	Bid        string     `gorm:"column:Fbid;type:varchar(64);not null;index:idx_bid"`
	Sourceid   string     `gorm:"column:Fsourceid;type:varchar(64);not null;index:idx_sourceid"`
	Operator   string     `gorm:"column:Foperator;type:varchar(64);not null;index:idx_operator"`
	Memo       string     `gorm:"column:Fmemo;type:varchar(64);not null"`
	State      int32      `gorm:"column:Fstate"`
	CreatedAt  time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt  time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt  *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_audit.
func (e *Audit) TableName() string {
	return "t_audit"
}

// ConfigTemplateSet definition for t_template_set
type ConfigTemplateSet struct {
	ID           uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Setid        string     `gorm:"column:Fsetid;type:varchar(64);not null;unique_index"`
	Bid          string     `gorm:"column:Fbid;type:varchar(64);not null;unique_index:idx_bid_name"`
	Name         string     `gorm:"column:Fname;type:varchar(64);not null;unique_index:idx_bid_name"`
	Memo         string     `gorm:"column:Fmemo;type:varchar(128)"`
	Fpath        string     `gorm:"column:Ffpath;type:varchar(128);not null"`
	Creator      string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	State        int32      `gorm:"column:Fstate"`
	CreatedAt    time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt    *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_template_set
func (c *ConfigTemplateSet) TableName() string {
	return "t_template_set"
}

// ConfigTemplate definition for t_template
type ConfigTemplate struct {
	ID           uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Templateid   string     `gorm:"column:Ftemplateid;type:varchar(64);not null;unique_index"`
	Bid          string     `gorm:"column:Fbid;type:varchar(64);not null;index:idx_bid"`
	Setid        string     `gorm:"column:Fsetid;type:varchar(64);not null;unique_index:idx_set_name"`
	Name         string     `gorm:"column:Fname;type:varchar(64);not null;unique_index:idx_set_name"`
	Memo         string     `gorm:"column:Fmemo;type:varchar(128)"`
	Fpath        string     `gorm:"column:Ffpath;type:varchar(128);not null"`
	User         string     `gorm:"column:Fuser;type:varchar(64)"`
	Group        string     `gorm:"column:Fgroup;type:varchar(64)"`
	Permission   int32      `gorm:"column:Fpermission;not null"`
	FileEncoding string     `gorm:"column:Ffile_encoding;type:varchar(16)"`
	EngineType   int32      `gorm:"column:Fengine_type;not null"`
	Creator      string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	State        int32      `gorm:"column:Fstate"`
	CreatedAt    time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt    *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_template
func (c *ConfigTemplate) TableName() string {
	return "t_template"
}

// ConfigTemplateVersion table name of t_template_version
type ConfigTemplateVersion struct {
	ID           uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Versionid    string     `gorm:"column:Fversionid;type:varchar(64);not null;unique_index"`
	Bid          string     `gorm:"column:Fbid;type:varchar(64);not null;index:idx_bid"`
	Templateid   string     `gorm:"column:Ftemplateid;type:varchar(64);not null;unique_index:idx_tpl_version"`
	VersionName  string     `gorm:"column:Fversion_name;type:varchar(64);not null;unique_index:idx_tpl_version"`
	Memo         string     `gorm:"column:Fmemo;type:varchar(128)"`
	Content      string     `gorm:"column:Fcontent;type:longtext"`
	Creator      string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	State        int32      `gorm:"column:Fstate"`
	CreatedAt    time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt    *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name of t_template_version
func (c *ConfigTemplateVersion) TableName() string {
	return "t_template_version"
}

// ConfigTemplateBinding definition for t_template_binding
type ConfigTemplateBinding struct {
	ID            uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Bid           string     `gorm:"column:Fbid;type:varchar(64);not null;unique_index:idx_unionids"`
	Templateid    string     `gorm:"column:Ftemplateid;type:varchar(64);not null;unique_index:idx_unionids"`
	Appid         string     `gorm:"column:Fappid;type:varchar(64);not null;unique_index:idx_unionids"`
	Versionid     string     `gorm:"column:Fversionid;type:varchar(64);not null"`
	Cfgsetid      string     `gorm:"column:Fcfgsetid;type:varchar(64);not null;index:idx_cfgset"`
	Commitid      string     `gorm:"column:Fcommitid;type:varchar(64)"`
	BindingParams string     `gorm:"column:Fbinding_params;type:longtext;not null"`
	Creator       string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy  string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	State         int32      `gorm:"column:Fstate"`
	CreatedAt     time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt     time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt     *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name for t_template_binding
func (c *ConfigTemplateBinding) TableName() string {
	return "t_template_binding"
}

// VariableGlobal definition for t_variable_global
type VariableGlobal struct {
	ID           uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Vid          string     `gorm:"column:Fvid;type:varchar(64);not null;unique_index"`
	Bid          string     `gorm:"column:Fbid;type:varchar(64);not null;unique_index:idx_unionids"`
	Key          string     `gorm:"column:Fkey;type:varchar(64);not null;unique_index:idx_unionids"`
	Value        string     `gorm:"column:Fvalue;type:varchar(1024);not null"`
	Memo         string     `gorm:"column:Fmemo;type:varchar(128)"`
	Creator      string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	State        int32      `gorm:"column:Fstate"`
	CreatedAt    time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt    time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt    *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name for t_variable_global
func (v *VariableGlobal) TableName() string {
	return "t_variable_global"
}

// VariableCluster definition for t_variable_cluster
type VariableCluster struct {
	ID            uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Vid           string     `gorm:"column:Fvid;type:varchar(64);not null;unique_index"`
	Bid           string     `gorm:"column:Fbid;type:varchar(64);not null;unique_index:idx_unionids"`
	Cluster       string     `gorm:"column:Fcluster;type:varchar(64);not null;unique_index:idx_unionids"`
	ClusterLabels string     `gorm:"column:Fcluster_labels;type:varchar(128);not null;unique_index:idx_unionids"`
	Key           string     `gorm:"column:Fkey;type:varchar(64);not null;unique_index:idx_unionids"`
	Value         string     `gorm:"column:Fvalue;type:varchar(1024);not null"`
	Memo          string     `gorm:"column:Fmemo;type:varchar(128)"`
	Creator       string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy  string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	State         int32      `gorm:"column:Fstate"`
	CreatedAt     time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt     time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt     *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name for t_variable_cluster
func (v *VariableCluster) TableName() string {
	return "t_variable_cluster"
}

// VariableZone definition for t_variable_zone
type VariableZone struct {
	ID            uint64     `gorm:"column:Fid;primary_key;AUTO_INCREMENT"`
	Vid           string     `gorm:"column:Fvid;type:varchar(64);not null;unique_index"`
	Bid           string     `gorm:"column:Fbid;type:varchar(64);not null;unique_index:idx_unionids"`
	Cluster       string     `gorm:"column:Fcluster;type:varchar(64);not null;unique_index:idx_unionids"`
	ClusterLabels string     `gorm:"column:Fcluster_labels;type:varchar(128);not null;unique_index:idx_unionids"`
	Zone          string     `gorm:"column:Fzone;type:varchar(64);not null;unique_index:idx_unionids"`
	Key           string     `gorm:"column:Fkey;type:varchar(64);not null;unique_index:idx_unionids"`
	Value         string     `gorm:"column:Fvalue;type:varchar(2048);not null"`
	Memo          string     `gorm:"column:Fmemo;type:varchar(128)"`
	Creator       string     `gorm:"column:Fcreator;type:varchar(64);not null"`
	LastModifyBy  string     `gorm:"column:Flast_modify_by;type:varchar(64);not null"`
	State         int32      `gorm:"column:Fstate"`
	CreatedAt     time.Time  `gorm:"column:Fcreate_time;index:idx_ctime"`
	UpdatedAt     time.Time  `gorm:"column:Fupdate_time;index:idx_utime"`
	DeletedAt     *time.Time `gorm:"column:Fdelete_time;index:idx_dtime"`
}

// TableName returns table name for t_variable_zone
func (v *VariableZone) TableName() string {
	return "t_variable_zone"
}
