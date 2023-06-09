/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

// Package dao NOTES
package dao

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/kit"
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
	StrategySet() StrategySet
	CRInstance() CRInstance
	Strategy() Strategy
	Hook() Hook
	HookRelease() HookRelease
	TemplateSpace() TemplateSpace
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
	ConfigHook() ConfigHook
}

// NewDaoSet create the DAO set instance.
func NewDaoSet(opt cc.Sharding, credentialSetting cc.Credential) (Set, error) {

	sd, err := sharding.InitSharding(&opt)
	if err != nil {
		return nil, fmt.Errorf("init sharding failed, err: %v", err)
	}

	adminDB, err := gorm.Open(mysql.Open(sharding.URI(opt.AdminDatabase)), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, err
	}

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

	s := &set{
		orm:               ormInst,
		db:                adminDB,
		genQ:              genQ,
		sd:                sd,
		credentialSetting: credentialSetting,
		idGen:             idDao,
		auditDao:          auditDao,
		event:             &eventDao{orm: ormInst, sd: sd, idGen: idDao},
		lock:              &lockDao{orm: ormInst, idGen: idDao},
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
		orm:      s.orm,
		sd:       s.sd,
		idGen:    s.idGen,
		auditDao: s.auditDao,
		event:    s.event,
	}
}

// Commit returns the commits' DAO
func (s *set) Commit() Commit {
	return &commitDao{
		orm:      s.orm,
		sd:       s.sd,
		idGen:    s.idGen,
		auditDao: s.auditDao,
	}
}

// ConfigItem returns the config item's DAO
func (s *set) ConfigItem() ConfigItem {
	return &configItemDao{
		orm:      s.orm,
		sd:       s.sd,
		idGen:    s.idGen,
		auditDao: s.auditDao,
		lock:     s.lock,
	}
}

// Content returns the content's DAO
func (s *set) Content() Content {
	return &contentDao{
		orm:      s.orm,
		sd:       s.sd,
		idGen:    s.idGen,
		auditDao: s.auditDao,
	}
}

// Release returns the release's DAO
func (s *set) Release() Release {
	return &releaseDao{
		orm:      s.orm,
		genQ:     s.genQ,
		sd:       s.sd,
		idGen:    s.idGen,
		auditDao: s.auditDao,
	}
}

// ReleasedCI returns the released config item's DAO
func (s *set) ReleasedCI() ReleasedCI {
	return &releasedCIDao{
		orm:      s.orm,
		sd:       s.sd,
		idGen:    s.idGen,
		auditDao: s.auditDao,
	}
}

// CRInstance returns the current released instance's DAO
func (s *set) CRInstance() CRInstance {
	return &crInstanceDao{
		orm:      s.orm,
		sd:       s.sd,
		idGen:    s.idGen,
		auditDao: s.auditDao,
		event:    s.event,
		lock:     s.lock,
	}
}

// StrategySet returns the strategy set's DAO
func (s *set) StrategySet() StrategySet {
	return &strategySetDao{
		orm:      s.orm,
		sd:       s.sd,
		idGen:    s.idGen,
		auditDao: s.auditDao,
		lock:     s.lock,
	}
}

// Strategy returns the strategy's DAO
func (s *set) Strategy() Strategy {
	return &strategyDao{
		orm:      s.orm,
		sd:       s.sd,
		idGen:    s.idGen,
		auditDao: s.auditDao,
		event:    s.event,
		lock:     s.lock,
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

// HookRelease returns the hookRelease's DAO
func (s *set) HookRelease() HookRelease {
	return &hookReleaseDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// TemplateSpace returns the templateSpace's DAO
func (s *set) TemplateSpace() TemplateSpace {
	return &templateSpaceDao{
		idGen:    s.idGen,
		auditDao: s.auditDao,
		genQ:     s.genQ,
	}
}

// Group returns the group's DAO
func (s *set) Group() Group {
	return &groupDao{
		orm:      s.orm,
		sd:       s.sd,
		idGen:    s.idGen,
		auditDao: s.auditDao,
		lock:     s.lock,
	}
}

// GroupAppBind returns the group app bind's DAO
func (s *set) GroupAppBind() GroupAppBind {
	return &groupAppDao{
		orm:      s.orm,
		sd:       s.sd,
		idGen:    s.idGen,
		auditDao: s.auditDao,
		lock:     s.lock,
	}
}

// ReleasedGroup returns the currnet release's DAO
func (s *set) ReleasedGroup() ReleasedGroup {
	return &releasedGroupDao{
		orm:      s.orm,
		sd:       s.sd,
		idGen:    s.idGen,
		auditDao: s.auditDao,
		lock:     s.lock,
	}
}

// Publish returns the publish operation related DAO
func (s *set) Publish() Publish {
	return &pubDao{
		orm:      s.orm,
		sd:       s.sd,
		idGen:    s.idGen,
		auditDao: s.auditDao,
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
		orm: s.orm,
		sd:  s.sd,
	}
}

// Event returns the event operation related DAO
func (s *set) Event() Event {
	return &eventDao{
		orm:   s.orm,
		sd:    s.sd,
		idGen: s.idGen,
	}
}

// Healthz check mysql healthz.
func (s *set) Healthz() error {
	return s.sd.Healthz()
}

// Credential returns the Credential's DAO
func (s *set) Credential() Credential {
	return &credentialDao{
		orm:               s.orm,
		sd:                s.sd,
		credentialSetting: s.credentialSetting,
		idGen:             s.idGen,
		auditDao:          s.auditDao,
		event:             s.event,
	}
}

// CredentialScope returns the Credential scope's DAO
func (s *set) CredentialScope() CredentialScope {
	return &credentialScopeDao{
		orm:      s.orm,
		sd:       s.sd,
		idGen:    s.idGen,
		auditDao: s.auditDao,
	}
}

// ConfigHook returns the configHook's DAO
func (s *set) ConfigHook() ConfigHook {
	return &configHookDao{
		genQ:     s.genQ,
		idGen:    s.idGen,
		auditDao: s.auditDao,
	}
}
