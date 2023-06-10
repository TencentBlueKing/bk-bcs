package dao

import (
	"fmt"
	"strconv"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// CredentialScope supplies all the credential scope related operations.
type CredentialScope interface {
	// CreateWithTx create credential scope with transaction
	CreateWithTx(kit *kit.Kit, tx *sharding.Tx, credential *table.CredentialScope) (uint32, error)
	// Get get credential scopes
	Get(kit *kit.Kit, credentialId, bizId uint32) (*types.ListCredentialScopeDetails, error)
	// DeleteWithTx delete credential scope with transaction
	DeleteWithTx(kit *kit.Kit, tx *sharding.Tx, bizID, id uint32) error
	// UpdateWithTx update credential scope with transaction
	UpdateWithTx(kit *kit.Kit, tx *sharding.Tx, credentialScope *table.CredentialScope) error
	// // UpdateCredentialScopes update credential scopes
	// UpdateCredentialScopes(kit *kit.Kit, option *types.UpdateCredentialScopesOption) error
}

var _ CredentialScope = new(credentialScopeDao)

type credentialScopeDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao

	orm orm.Interface
	sd  *sharding.Sharding
}

// CreateWithTx create credential scope with transaction
func (dao *credentialScopeDao) CreateWithTx(kit *kit.Kit, tx *sharding.Tx, c *table.CredentialScope) (uint32, error) {

	if c == nil {
		return 0, errf.New(errf.InvalidParameter, "credential scope is nil")
	}

	if err := c.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	// generate a credential id and update to credential.
	id, err := dao.idGen.One(kit, table.CredentialScopeTable)
	if err != nil {
		return 0, err
	}

	c.ID = id
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.CredentialScopeTable.Name(), " (", table.CredentialScopeColumns.ColumnExpr(), ")  VALUES(", table.CredentialScopeColumns.ColonNameExpr(), ")")

	sql := filter.SqlJoint(sqlSentence)

	if err := dao.orm.Txn(tx.Tx()).Insert(kit.Ctx, sql, c); err != nil {
		return 0, err
	}

	//
	au := &AuditOption{Txn: tx.Tx(), ResShardingUid: tx.ShardingUid()}
	if err = dao.auditDao.Decorator(kit, c.Attachment.BizID,
		enumor.CredentialScope).AuditCreate(c, au); err != nil {
		return 0, fmt.Errorf("audit create credential scope failed, err: %v", err)
	}

	return id, nil
}

// Get get credential scope
func (dao *credentialScopeDao) Get(kit *kit.Kit, credentialId, bizId uint32) (*types.ListCredentialScopeDetails, error) {
	if credentialId == 0 {
		return nil, errf.New(errf.InvalidParameter, "credential scope credential id null")
	}
	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id", "biz_id", "credential_id"},
		CrownedOption: &filter.CrownedOption{
			CrownedOp: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "biz_id",
					Op:    filter.Equal.Factory(),
					Value: bizId,
				},
				&filter.AtomRule{
					Field: "credential_id",
					Op:    filter.Equal.Factory(),
					Value: credentialId,
				},
			},
		},
	}
	ft := &filter.Expression{
		Op:    filter.Or,
		Rules: []filter.RuleFactory{},
	}
	whereExpr, args, err := ft.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}
	var sqlSentenceCount []string
	sqlSentenceCount = append(sqlSentenceCount, "SELECT COUNT(*) FROM ", table.CredentialScopeTable.Name(), whereExpr)
	countSql := filter.SqlJoint(sqlSentenceCount)
	count, err := dao.orm.Do(dao.sd.ShardingOne(bizId).DB()).Count(kit.Ctx, countSql, args...)
	if err != nil {
		return nil, err
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.CredentialScopeColumns.NamedExpr(), " FROM ", table.CredentialScopeTable.Name(), whereExpr)
	sql := filter.SqlJoint(sqlSentence)
	list := make([]*table.CredentialScope, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(bizId).DB()).Select(kit.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListCredentialScopeDetails{Count: count, Details: list}, nil
}

// DeleteWithTx delete credential scope with transaction
func (dao *credentialScopeDao) DeleteWithTx(kit *kit.Kit, tx *sharding.Tx, bizID, id uint32) error {
	if bizID == 0 {
		return errf.New(errf.InvalidParameter, "biz id is zero")
	}

	if id == 0 {
		return errf.New(errf.InvalidParameter, "credential scope id is zero")
	}

	ab := dao.auditDao.Decorator(kit, bizID, enumor.CredentialScope).PrepareDelete(id)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", table.CredentialScopeTable.Name(), " WHERE id = ", strconv.Itoa(int(id)),
		" AND biz_id = ", strconv.Itoa(int(bizID)))
	expr := filter.SqlJoint(sqlSentence)

	err := dao.orm.Txn(tx.Tx()).Delete(kit.Ctx, expr)
	if err != nil {
		logs.Errorf("delete credential scope: %d failed, err: %v, rid: %v", id, err, kit.Rid)
		return err
	}

	auditOpt := &AuditOption{Txn: tx.Tx(), ResShardingUid: tx.ShardingUid()}
	if err := ab.Do(auditOpt); err != nil {
		return fmt.Errorf("audit delete credential scope failed, err: %v", err)
	}
	return nil
}

// UpdateWithTx update credential scope with transaction
func (dao *credentialScopeDao) UpdateWithTx(kit *kit.Kit, tx *sharding.Tx, c *table.CredentialScope) error {

	if c == nil {
		return errf.New(errf.InvalidParameter, "credential scope is nil")
	}

	if err := c.ValidateUpdate(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	opts := orm.NewFieldOptions().AddIgnoredFields(
		"id", "biz_id")
	expr, toUpdate, err := orm.RearrangeSQLDataWithOption(c, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql expr failed, err: %v", err)
	}

	ab := dao.auditDao.Decorator(kit, c.Attachment.BizID, enumor.CredentialScope).PrepareUpdate(c)
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", table.CredentialScopeTable.Name(), " SET ", expr, " WHERE id = ", strconv.Itoa(int(c.ID)),
		" AND biz_id = ", strconv.Itoa(int(c.Attachment.BizID)))
	sql := filter.SqlJoint(sqlSentence)

	var effected int64
	effected, err = dao.orm.Txn(tx.Tx()).Update(kit.Ctx, sql, toUpdate)
	if err != nil {
		logs.Errorf("update credential scope: %d failed, err: %v, rid: %v", c.ID, err, kit.Rid)
		return err
	}

	if effected == 0 {
		logs.Errorf("update one credential scope: %d, but record not found, rid: %v", c.ID, kit.Rid)
		return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
	}

	if effected > 1 {
		logs.Errorf("update one credential scope: %d, but got updated credential count: %d, rid: %v", c.ID,
			effected, kit.Rid)
		return fmt.Errorf("matched credential scope count %d is not as excepted", effected)
	}

	// do audit
	if err := ab.Do(&AuditOption{Txn: tx.Tx(), ResShardingUid: tx.ShardingUid()}); err != nil {
		return fmt.Errorf("do credential scope update audit failed, err: %v", err)
	}

	return nil
}
