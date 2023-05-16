package dao

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// Credential supplies all the Credential related operations.
type Credential interface {
	// Get get credential
	Get(kit *kit.Kit, bizID, id uint32) (*table.Credential, error)
	// GetByCredentialString get credential by credential string
	GetByCredentialString(kit *kit.Kit, bizID uint32, credential string) (*table.Credential, error)
	// Create one credential instance.
	Create(kit *kit.Kit, credential *table.Credential) (uint32, error)
	// List get credentials
	List(kit *kit.Kit, opts *types.ListCredentialsOption) (*types.ListCredentialDetails, error)
	// Delete delete credential
	Delete(kit *kit.Kit, strategy *table.Credential) error
	// Update update credential
	Update(kit *kit.Kit, credential *table.Credential) error
	// UpdateRevisionWithTx update credential revision with transaction
	UpdateRevisionWithTx(kit *kit.Kit, tx *sharding.Tx, bizID, id uint32) error
}

var _ Credential = new(credentialDao)

type credentialDao struct {
	orm               orm.Interface
	sd                *sharding.Sharding
	credentialSetting cc.Credential
	idGen             IDGenInterface
	auditDao          AuditDao
	event             Event
}

func (dao *credentialDao) Get(kit *kit.Kit, bizID, id uint32) (*table.Credential, error) {
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "bizID is empty")
	}
	if id == 0 {
		return nil, errf.New(errf.InvalidParameter, "credential id is empty")
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.CredentialColumns.NamedExpr(), " FROM ",
		table.CredentialTable.Name(), " WHERE id = '", strconv.Itoa(int(id)), "' AND biz_id = ", strconv.Itoa(int(bizID)))
	sql := filter.SqlJoint(sqlSentence)

	one := new(table.Credential)
	if err := dao.orm.Do(dao.sd.MustSharding(bizID)).Get(kit.Ctx, one, sql); err != nil {
		return nil, fmt.Errorf("get credential failed, err: %v", err)
	}
	return one, nil
}

// Get Credential by encoded credential string.
func (dao *credentialDao) GetByCredentialString(kit *kit.Kit, bizID uint32, str string) (*table.Credential, error) {
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "bizID is empty")
	}
	if str == "" {
		return nil, errf.New(errf.InvalidParameter, "credential string is empty")
	}

	// encode credential string
	encryptionAlgorithm := dao.credentialSetting.EncryptionAlgorithm
	masterKey := dao.credentialSetting.MasterKey
	encrypted, err := tools.EncryptCredential(str, masterKey, encryptionAlgorithm)
	if err != nil {
		return nil, errf.ErrCredentialInvalid
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.CredentialColumns.NamedExpr(), " FROM ",
		table.CredentialTable.Name(), " WHERE enc_credential = '", encrypted, "' AND biz_id = ", strconv.Itoa(int(bizID)))
	sql := filter.SqlJoint(sqlSentence)

	one := new(table.Credential)
	if err := dao.orm.Do(dao.sd.MustSharding(bizID)).Get(kit.Ctx, one, sql); err != nil {
		return nil, fmt.Errorf("get credential failed, err: %v", err)
	}

	return one, nil
}

// Create create credential
func (dao *credentialDao) Create(kit *kit.Kit, c *table.Credential) (uint32, error) {

	if c == nil {
		return 0, errf.New(errf.InvalidParameter, "credential is nil")
	}

	if err := c.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	// generate a credential id and update to credential.
	id, err := dao.idGen.One(kit, table.CredentialTable)
	if err != nil {
		return 0, err
	}

	c.ID = id
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.CredentialTable.Name(), " (", table.CredentialColumns.ColumnExpr(), ")  VALUES(", table.CredentialColumns.ColonNameExpr(), ")")

	sql := filter.SqlJoint(sqlSentence)
	err = dao.sd.ShardingOne(c.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			if err := dao.orm.Txn(txn).Insert(kit.Ctx, sql, c); err != nil {
				return err
			}

			au := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
			if err = dao.auditDao.Decorator(kit, c.Attachment.BizID,
				enumor.Credential).AuditCreate(c, au); err != nil {
				return fmt.Errorf("audit create credential failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		logs.Errorf("create credential, but do auto txn failed, err: %v, rid: %s", err, kit.Rid)
		return 0, fmt.Errorf("create credential, but auto run txn failed, err: %v", err)
	}

	return id, nil
}

// List get credentials
func (dao *credentialDao) List(kit *kit.Kit, opts *types.ListCredentialsOption) (*types.ListCredentialDetails, error) {
	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list credential options null")
	}
	if err := opts.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}
	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id", "biz_id"},
		CrownedOption: &filter.CrownedOption{
			CrownedOp: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "biz_id",
					Op:    filter.Equal.Factory(),
					Value: opts.BizID,
				},
			},
		},
	}
	whereExpr, args, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}
	var sqlSentenceCount []string
	sqlSentenceCount = append(sqlSentenceCount, "SELECT COUNT(*) FROM ", table.CredentialTable.Name(), whereExpr)
	countSql := filter.SqlJoint(sqlSentenceCount)
	count, err := dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, countSql, args...)
	if err != nil {
		return nil, err
	}

	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.CredentialColumns.NamedExpr(), " FROM ", table.CredentialTable.Name(), whereExpr, pageExpr)
	sql := filter.SqlJoint(sqlSentence)

	list := make([]*table.Credential, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListCredentialDetails{Count: count, Details: list}, nil
}

// Delete delete credential
func (dao *credentialDao) Delete(kit *kit.Kit, g *table.Credential) error {
	if g == nil {
		return errf.New(errf.InvalidParameter, "credential is nil")
	}

	if err := g.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	ab := dao.auditDao.Decorator(kit, g.Attachment.BizID, enumor.Credential).PrepareDelete(g.ID)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", table.CredentialTable.Name(), " WHERE id = ", strconv.Itoa(int(g.ID)),
		" AND biz_id = ", strconv.Itoa(int(g.Attachment.BizID)))
	expr := filter.SqlJoint(sqlSentence)

	err := dao.sd.ShardingOne(g.Attachment.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {

		err := dao.orm.Txn(txn).Delete(kit.Ctx, expr)
		if err != nil {
			return err
		}

		auditOpt := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
		if err := ab.Do(auditOpt); err != nil {
			return fmt.Errorf("audit delete credential failed, err: %v", err)
		}

		return nil
	})

	if err != nil {
		logs.Errorf("delete credential: %d failed, err: %v, rid: %v", g.ID, err, kit.Rid)
		return fmt.Errorf("delete credential, but run txn failed, err: %v", err)
	}
	return nil
}

// Update update credential
// Note: only update name, description, enable
func (dao *credentialDao) Update(kit *kit.Kit, g *table.Credential) error {
	if g == nil {
		return errf.New(errf.InvalidParameter, "credential is nil")
	}

	if err := g.ValidateUpdate(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	opts := orm.NewFieldOptions().AddBlankedFields("enable", "memo").
		AddIgnoredFields("id", "biz_id", "enc_algorithm", "enc_credential", "credential_type")
	expr, toUpdate, err := orm.RearrangeSQLDataWithOption(g, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql expr failed, err: %v", err)
	}

	ab := dao.auditDao.Decorator(kit, g.Attachment.BizID, enumor.Credential).PrepareUpdate(g)
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", table.CredentialTable.Name(), " SET ", expr, " WHERE id = ", strconv.Itoa(int(g.ID)),
		" AND biz_id = ", strconv.Itoa(int(g.Attachment.BizID)))
	sql := filter.SqlJoint(sqlSentence)
	err = dao.sd.ShardingOne(g.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			var effected int64
			effected, err = dao.orm.Txn(txn).Update(kit.Ctx, sql, toUpdate)
			if err != nil {
				logs.Errorf("update credential: %d failed, err: %v, rid: %v", g.ID, err, kit.Rid)
				return err
			}

			if effected == 0 {
				logs.Errorf("update one credential: %d, but record not found, rid: %v", g.ID, kit.Rid)
				return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
			}

			if effected > 1 {
				logs.Errorf("update one credential: %d, but got updated credential count: %d, rid: %v", g.ID,
					effected, kit.Rid)
				return fmt.Errorf("matched credential count %d is not as excepted", effected)
			}

			// do audit
			if err := ab.Do(&AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}); err != nil {
				return fmt.Errorf("do credential update audit failed, err: %v", err)
			}
			return nil
		})

	if err != nil {
		return err
	}

	return nil
}

// UpdateRevisionWithTx update credential revision with transaction
func (dao *credentialDao) UpdateRevisionWithTx(kit *kit.Kit, tx *sharding.Tx, bizID uint32, id uint32) error {
	if bizID == 0 || id == 0 {
		return errf.New(errf.InvalidParameter, "credential bizID or id is zero")
	}

	var sqlSentence []string
	now := time.Now().Format(constant.TimeStdFormat)
	sqlSentence = append(sqlSentence, "UPDATE ", table.CredentialTable.Name(),
		" SET updated_at = :updated_at, reviser = :reviser"+
			" WHERE id = ", strconv.Itoa(int(id)), " AND biz_id = ", strconv.Itoa(int(bizID)))
	sql := filter.SqlJoint(sqlSentence)

	toUpdate := map[string]interface{}{
		"updated_at": now,
		"reviser":    kit.User,
	}

	_, err := dao.orm.Txn(tx.Tx()).Update(kit.Ctx, sql, toUpdate)
	if err != nil {
		logs.Errorf("update credential %d revision failed, err: %v, rid: %v", id, err, kit.Rid)
		return err
	}

	return nil
}
