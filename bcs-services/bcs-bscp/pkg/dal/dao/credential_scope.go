package dao

import (
	"fmt"
	"strconv"

	"github.com/jmoiron/sqlx"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// Credential supplies all the Credential related operations.
type CredentialScope interface {
	Create(kit *kit.Kit, credential *table.CredentialScope) (uint32, error)

	Get(kit *kit.Kit, credentialId, bizId uint32) (*types.ListCredentialScopeDetails, error)

	Delete(kit *kit.Kit, strategy *table.CredentialScope) error
}

var _ CredentialScope = new(credentialScopeDao)

type credentialScopeDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create
func (dao *credentialScopeDao) Create(kit *kit.Kit, c *table.CredentialScope) (uint32, error) {

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

	err = dao.sd.ShardingOne(c.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			if err := dao.orm.Txn(txn).Insert(kit.Ctx, sql, c); err != nil {
				return err
			}

			//
			au := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
			if err = dao.auditDao.Decorator(kit, c.Attachment.BizID,
				enumor.CredentialScope).AuditCreate(c, au); err != nil {
				return fmt.Errorf("audit create credential scope failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		logs.Errorf("create credential scope, but do auto txn failed, err: %v, rid: %s", err, kit.Rid)
		return 0, fmt.Errorf("create credential scope, but auto run txn failed, err: %v", err)
	}

	return id, nil
}

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

func (dao *credentialScopeDao) Delete(kit *kit.Kit, g *table.CredentialScope) error {
	if g == nil {
		return errf.New(errf.InvalidParameter, "credential scope is nil")
	}

	if err := g.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	ab := dao.auditDao.Decorator(kit, g.Attachment.BizID, enumor.CredentialScope).PrepareDelete(g.ID)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", table.CredentialScopeTable.Name(), " WHERE id = ", strconv.Itoa(int(g.ID)),
		" AND biz_id = ", strconv.Itoa(int(g.Attachment.BizID)))
	expr := filter.SqlJoint(sqlSentence)

	err := dao.sd.ShardingOne(g.Attachment.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {

		err := dao.orm.Txn(txn).Delete(kit.Ctx, expr)
		if err != nil {
			return err
		}

		auditOpt := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
		if err := ab.Do(auditOpt); err != nil {
			return fmt.Errorf("audit delete credential scope failed, err: %v", err)
		}

		return nil
	})

	if err != nil {
		logs.Errorf("delete credential scope: %d failed, err: %v, rid: %v", g.ID, err, kit.Rid)
		return fmt.Errorf("delete credential scope, but run txn failed, err: %v", err)
	}
	return nil
}
