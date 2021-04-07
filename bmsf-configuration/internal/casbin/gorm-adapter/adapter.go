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

package gormadapter

import (
	"errors"
	"runtime"
	"strings"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"bk-bscp/internal/database"
)

// Adapter represents the gorm adapter for policy storage.
type Adapter struct {
	driverName     string
	dataSourceName string
	databaseName   string
	tableName      string
	db             *gorm.DB
}

func finalizer(a *Adapter) error {
	sqlDB, err := a.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// NewAdapter creates a new Adapter object.
func NewAdapter(driverName string, dataSourceName, databaseName, tableName string) (*Adapter, error) {
	a := &Adapter{
		driverName:     driverName,
		dataSourceName: dataSourceName,
		databaseName:   databaseName,
		tableName:      tableName,
	}
	if err := a.open(); err != nil {
		return nil, err
	}
	runtime.SetFinalizer(a, finalizer)
	return a, nil
}

func (a *Adapter) open() error {
	db, err := openDBConnection(a.driverName, a.dataSourceName)
	if err != nil {
		return err
	}

	a.db = db.Scopes(func(db *gorm.DB) *gorm.DB {
		return db.Table(a.tableName)
	}).Session(&gorm.Session{})

	return nil
}

func openDBConnection(driverName, dataSourceName string) (*gorm.DB, error) {
	var err error
	var db *gorm.DB

	if driverName == "postgres" {
		db, err = gorm.Open(postgres.Open(dataSourceName), &gorm.Config{})
	} else if driverName == "mysql" {
		db, err = gorm.Open(mysql.Open(dataSourceName), &gorm.Config{})
	} else {
		return nil, errors.New("database dialect is not supported")
	}
	if err != nil {
		return nil, err
	}
	return db, err
}

func (a *Adapter) getTableInstance() *database.LocalAuth {
	return &database.LocalAuth{}
}

func loadPolicyLine(line database.LocalAuth, model model.Model) {
	p := []string{line.PType, line.V0, line.V1, line.V2, line.V3, line.V4, line.V5}

	var lineText string
	if line.V5 != "" {
		lineText = strings.Join(p, ", ")
	} else if line.V4 != "" {
		lineText = strings.Join(p[:6], ", ")
	} else if line.V3 != "" {
		lineText = strings.Join(p[:5], ", ")
	} else if line.V2 != "" {
		lineText = strings.Join(p[:4], ", ")
	} else if line.V1 != "" {
		lineText = strings.Join(p[:3], ", ")
	} else if line.V0 != "" {
		lineText = strings.Join(p[:2], ", ")
	}
	persist.LoadPolicyLine(lineText, model)
}

// LoadPolicy loads all policy rules from the storage.
func (a *Adapter) LoadPolicy(model model.Model) error {
	var lines []database.LocalAuth
	if err := a.db.Order("Fid").Find(&lines).Error; err != nil {
		return err
	}

	for _, line := range lines {
		loadPolicyLine(line, model)
	}
	return nil
}

func (a *Adapter) savePolicyLine(ptype string, rule []string) database.LocalAuth {
	line := a.getTableInstance()

	line.PType = ptype
	if len(rule) > 0 {
		line.V0 = rule[0]
	}
	if len(rule) > 1 {
		line.V1 = rule[1]
	}
	if len(rule) > 2 {
		line.V2 = rule[2]
	}
	if len(rule) > 3 {
		line.V3 = rule[3]
	}
	if len(rule) > 4 {
		line.V4 = rule[4]
	}
	if len(rule) > 5 {
		line.V5 = rule[5]
	}

	return *line
}

// SavePolicy saves all policy rules to the storage.
func (a *Adapter) SavePolicy(model model.Model) error {
	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			line := a.savePolicyLine(ptype, rule)
			err := a.db.Create(&line).Error
			if err != nil {
				return err
			}
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			line := a.savePolicyLine(ptype, rule)
			err := a.db.Create(&line).Error
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// AddPolicy adds a policy rule to the storage.
func (a *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	line := a.savePolicyLine(ptype, rule)
	err := a.db.Create(&line).Error
	return err
}

// RemovePolicy removes a policy rule from the storage.
func (a *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	line := a.savePolicyLine(ptype, rule)
	err := a.rawDelete(a.db, line)
	return err
}

// AddPolicies adds multiple policy rules to the storage.
func (a *Adapter) AddPolicies(sec string, ptype string, rules [][]string) error {
	return a.db.Transaction(func(tx *gorm.DB) error {
		for _, rule := range rules {
			line := a.savePolicyLine(ptype, rule)
			if err := tx.Create(&line).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// RemovePolicies removes multiple policy rules from the storage.
func (a *Adapter) RemovePolicies(sec string, ptype string, rules [][]string) error {
	return a.db.Transaction(func(tx *gorm.DB) error {
		for _, rule := range rules {
			line := a.savePolicyLine(ptype, rule)
			if err := a.rawDelete(tx, line); err != nil {
				return err
			}
		}
		return nil
	})
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	line := a.getTableInstance()

	line.PType = ptype
	if fieldIndex <= 0 && 0 < fieldIndex+len(fieldValues) {
		line.V0 = fieldValues[0-fieldIndex]
	}
	if fieldIndex <= 1 && 1 < fieldIndex+len(fieldValues) {
		line.V1 = fieldValues[1-fieldIndex]
	}
	if fieldIndex <= 2 && 2 < fieldIndex+len(fieldValues) {
		line.V2 = fieldValues[2-fieldIndex]
	}
	if fieldIndex <= 3 && 3 < fieldIndex+len(fieldValues) {
		line.V3 = fieldValues[3-fieldIndex]
	}
	if fieldIndex <= 4 && 4 < fieldIndex+len(fieldValues) {
		line.V4 = fieldValues[4-fieldIndex]
	}
	if fieldIndex <= 5 && 5 < fieldIndex+len(fieldValues) {
		line.V5 = fieldValues[5-fieldIndex]
	}
	err := a.rawDelete(a.db, *line)
	return err
}

func (a *Adapter) rawDelete(db *gorm.DB, line database.LocalAuth) error {
	queryArgs := []interface{}{line.PType}

	queryStr := "Fp_type = ?"
	if line.V0 != "" {
		queryStr += " and Fv0 = ?"
		queryArgs = append(queryArgs, line.V0)
	}
	if line.V1 != "" {
		queryStr += " and Fv1 = ?"
		queryArgs = append(queryArgs, line.V1)
	}
	if line.V2 != "" {
		queryStr += " and Fv2 = ?"
		queryArgs = append(queryArgs, line.V2)
	}
	if line.V3 != "" {
		queryStr += " and Fv3 = ?"
		queryArgs = append(queryArgs, line.V3)
	}
	if line.V4 != "" {
		queryStr += " and Fv4 = ?"
		queryArgs = append(queryArgs, line.V4)
	}
	if line.V5 != "" {
		queryStr += " and Fv5 = ?"
		queryArgs = append(queryArgs, line.V5)
	}
	args := append([]interface{}{queryStr}, queryArgs...)
	err := db.Delete(a.getTableInstance(), args...).Error
	return err
}
