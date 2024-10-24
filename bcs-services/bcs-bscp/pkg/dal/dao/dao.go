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

// Package dao NOTES
package dao

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/orm"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/sharding"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// Set defines all the DAO to be operated.
type Set interface {
	GenQuery() *gen.Query
	ID() IDGenInterface
	App() App
	Commit() Commit
	ConfigItem() ConfigItem
	Content() Content
	Release() Release
	ReleasedCI() ReleasedCI
	Hook() Hook
	HookRevision() HookRevision
	ReleasedHook() ReleasedHook
	TemplateSpace() TemplateSpace
	Template() Template
	TemplateRevision() TemplateRevision
	TemplateSet() TemplateSet
	AppTemplateBinding() AppTemplateBinding
	ReleasedAppTemplate() ReleasedAppTemplate
	AppTemplateVariable() AppTemplateVariable
	ReleasedAppTemplateVariable() ReleasedAppTemplateVariable
	TemplateBindingRelation() TemplateBindingRelation
	TemplateVariable() TemplateVariable
	Validator() Validator
	Group() Group
	GroupAppBind() GroupAppBind
	ReleasedGroup() ReleasedGroup
	Publish() Publish
	IAM() IAM
	Event() Event
	BeginTx(kit *kit.Kit, bizID uint32) (*sharding.Tx, error)
	Healthz() error
	Credential() Credential
	CredentialScope() CredentialScope
	Kv() Kv
	ReleasedKv() ReleasedKv
	Client() Client
	ClientEvent() ClientEvent
	ClientQuery() ClientQuery
}

// NewDaoSet create the DAO set instance.
func NewDaoSet(opt cc.Sharding, credentialSetting cc.Credential, gormSetting cc.Gorm) (Set, error) {

	// opt cc.Database
	sd, err := sharding.InitSharding(&opt)
	if err != nil {
		return nil, fmt.Errorf("init sharding failed, err: %v", err)
	}

	adminDB, err := gorm.Open(mysql.Open(sharding.URI(opt.AdminDatabase)),
		&gorm.Config{Logger: logger.Default.LogMode(gormSetting.GetLogLevel())})
	if err != nil {
		return nil, err
	}

	db, err := adminDB.DB()
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(int(opt.AdminDatabase.MaxOpenConn))
	db.SetMaxIdleConns(int(opt.AdminDatabase.MaxIdleConn))
	db.SetConnMaxLifetime(time.Duration(opt.AdminDatabase.MaxIdleTimeoutMin) * time.Minute)

	if e := adminDB.Use(tracing.NewPlugin(tracing.WithoutMetrics())); e != nil {
		return nil, err
	}

	// 会定期执行 SHOW STATUS; 拿状态数据
	// metricsCollector := []prometheus.MetricsCollector{
	// 	&prometheus.MySQL{VariableNames: []string{"Threads_running"}},
	// }

	if e := adminDB.Use(prometheus.New(prometheus.Config{})); e != nil {
		return nil, err
	}

	// auditor 分库, 注意需要在分表前面
	// auditorDB := sharding.MustShardingAuditor(adminDB)

	// biz 分表 mysql.Dialector -> sharding.ShardingDialector
	// 不支持 sqlparser.QualifiedRef, 暂时去掉, 参考 issue https://github.com/go-gorm/sharding/pull/32
	// if err := sharding.InitBizSharding(adminDB); err != nil {
	// 	return nil, err
	// }

	// 初始化 Gen 配置
	genQ := gen.Use(adminDB)

	ormInst := orm.Do(opt)
	idDao := &idGenerator{sd: sd, genQ: genQ}
	auditDao, err := NewAuditDao(adminDB, ormInst, sd, idDao)
	if err != nil {
		return nil, fmt.Errorf("new audit dao failed, err: %v", err)
	}
	eventDao := &eventDao{genQ: genQ, idGen: idDao, auditDao: auditDao}
	lockDao := &lockDao{genQ: genQ, idGen: idDao}

	s := &set{
		orm:               ormInst,
		db:                adminDB,
		genQ:              genQ,
		sd:                sd,
		credentialSetting: credentialSetting,
		idGen:             idDao,
		auditDao:          auditDao,
		event:             eventDao,
		lock:              lockDao,
	}

	return s, nil
}

type set struct {
	orm               orm.Interface
	genQ              *gen.Query
	db                *gorm.DB
	sd                *sharding.Sharding
	credentialSetting cc.Credential
	idGen             IDGenInterface
	auditDao          AuditDao
	event             Event
	lock              LockDao
}

// GenQuery returns the gen Query object
func (s *set) GenQuery() *gen.Query {
	return s.genQ
}

// ID returns the resource id generator DAO
func (s *set) ID() IDGenInterface {
	return s.idGen
}

// App returns the application's DAO
func (s *set) App() App {
	return &appDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
		event:    s.event,
	}
}

// Commit returns the commits' DAO
func (s *set) Commit() Commit {
	return &commitDao{
		genQ:     s.genQ,
		idGen:    s.idGen,
		auditDao: s.auditDao,
	}
}

// ConfigItem returns the config item's DAO
func (s *set) ConfigItem() ConfigItem {
	return &configItemDao{
		genQ:     s.genQ,
		idGen:    s.idGen,
		auditDao: s.auditDao,
		lock:     s.lock,
	}
}

// Content returns the content's DAO
func (s *set) Content() Content {
	return &contentDao{
		genQ:     s.genQ,
		idGen:    s.idGen,
		auditDao: s.auditDao,
	}
}

// Release returns the release's DAO
func (s *set) Release() Release {
	return &releaseDao{
		genQ:     s.genQ,
		sd:       s.sd,
		idGen:    s.idGen,
		auditDao: s.auditDao,
	}
}

// ReleasedCI returns the released config item's DAO
func (s *set) ReleasedCI() ReleasedCI {
	return &releasedCIDao{
		genQ:     s.genQ,
		idGen:    s.idGen,
		auditDao: s.auditDao,
	}
}

// Hook returns the hook's DAO
func (s *set) Hook() Hook {
	return &hookDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// HookRevision returns the hookRevision's DAO
func (s *set) HookRevision() HookRevision {
	return &hookRevisionDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// ReleasedHook returns the released hook's DAO
func (s *set) ReleasedHook() ReleasedHook {
	return &releasedHookDao{
		genQ:     s.genQ,
		idGen:    s.idGen,
		auditDao: s.auditDao,
	}
}

// TemplateSpace returns the template space's DAO
func (s *set) TemplateSpace() TemplateSpace {
	return &templateSpaceDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// Template returns the template's DAO
func (s *set) Template() Template {
	return &templateDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// TemplateRevision returns the template release's DAO
func (s *set) TemplateRevision() TemplateRevision {
	return &templateRevisionDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// TemplateSet returns the template set's DAO
func (s *set) TemplateSet() TemplateSet {
	return &templateSetDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// AppTemplateBinding returns the app template binding's DAO
func (s *set) AppTemplateBinding() AppTemplateBinding {
	return &appTemplateBindingDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// ReleasedAppTemplate returns the released app template's DAO
func (s *set) ReleasedAppTemplate() ReleasedAppTemplate {
	return &releasedAppTemplateDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// AppTemplateVariable returns the app template variable's DAO
func (s *set) AppTemplateVariable() AppTemplateVariable {
	return &appTemplateVariableDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// ReleasedAppTemplateVariable returns the released app template variable's DAO
func (s *set) ReleasedAppTemplateVariable() ReleasedAppTemplateVariable {
	return &releasedAppTemplateVariableDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// TemplateBindingRelation returns the template binding relation's DAO
func (s *set) TemplateBindingRelation() TemplateBindingRelation {
	return &templateBindingRelationDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// TemplateVariable returns the template variable's DAO
func (s *set) TemplateVariable() TemplateVariable {
	return &templateVariableDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// Validator returns the template binding relation's DAO
func (s *set) Validator() Validator {
	return &validatorDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// Group returns the group's DAO
func (s *set) Group() Group {
	return &groupDao{
		genQ:     s.genQ,
		idGen:    s.idGen,
		auditDao: s.auditDao,
		lock:     s.lock,
	}
}

// GroupAppBind returns the group app bind's DAO
func (s *set) GroupAppBind() GroupAppBind {
	return &groupAppDao{
		genQ:     s.genQ,
		idGen:    s.idGen,
		auditDao: s.auditDao,
		lock:     s.lock,
	}
}

// ReleasedGroup returns the currnet release's DAO
func (s *set) ReleasedGroup() ReleasedGroup {
	return &releasedGroupDao{
		genQ:     s.genQ,
		idGen:    s.idGen,
		auditDao: s.auditDao,
		lock:     s.lock,
	}
}

// Publish returns the publish operation related DAO
func (s *set) Publish() Publish {
	return &pubDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
		event:    s.event,
	}
}

// BeginTx return sharding one db instance's transaction.
func (s *set) BeginTx(kit *kit.Kit, bizID uint32) (*sharding.Tx, error) {
	if bizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id not set")
	}

	tx, err := s.sd.ShardingOne(bizID).BeginTx(kit)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// IAM returns the iam operation related DAO
func (s *set) IAM() IAM {
	return &iamDao{
		orm:  s.orm,
		sd:   s.sd,
		genQ: s.genQ,
	}
}

// Event returns the event operation related DAO
func (s *set) Event() Event {
	return &eventDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// Healthz check mysql healthz.
func (s *set) Healthz() error {
	return s.sd.Healthz()
}

// Credential returns the Credential's DAO
func (s *set) Credential() Credential {
	return &credentialDao{
		credentialSetting: &s.credentialSetting,
		idGen:             s.idGen,
		auditDao:          s.auditDao,
		genQ:              s.genQ,
		event:             s.event,
	}
}

// CredentialScope returns the Credential scope's DAO
func (s *set) CredentialScope() CredentialScope {
	return &credentialScopeDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// Kv returns the kv DAO
func (s *set) Kv() Kv {
	return &kvDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// ReleasedKv returns the ReleasedKv scope's DAO
func (s *set) ReleasedKv() ReleasedKv {
	return &releasedKvDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// Client returns the Client scope's DAO
func (s *set) Client() Client {
	return &clientDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// ClientEvent returns the ClientEvent scope's DAO
func (s *set) ClientEvent() ClientEvent {
	return &clientEventDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// ClientQuery returns the ClientQuery scope's DAO
func (s *set) ClientQuery() ClientQuery {
	return &clientQueryDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}
