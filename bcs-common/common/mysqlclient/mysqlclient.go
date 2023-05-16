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
 *
 */

// Package mysqlclient xxx
package mysqlclient

import (
	"database/sql"
	"strconv"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
)

// MySql instance for mysql
type MySql struct {
	host       string
	port       int
	usr        string
	pwd        string
	databese   string
	driverName string
	db         *sql.DB
}

// NewMySql xxx
func NewMySql() (*MySql, error) {
	mysql := new(MySql)
	mysql.db = nil

	return mysql, nil
}

// Open xxx
func (m *MySql) Open(host, usr, pwd, database string, port, maxOpenConns, maxIdleConns int) error {
	m.host = host
	m.port = port
	m.usr = usr
	m.pwd = pwd
	m.databese = database

	// driver: usr:pwd@tcp(host:port)/database
	m.driverName = m.usr + ":" + m.pwd + "@tcp(" + m.host + ":" + strconv.Itoa(m.port) + ")/" + m.databese

	db, err := sql.Open("mysql", m.driverName)
	if err != nil {
		return err
	}

	m.db = db

	if maxOpenConns < 0 {
		maxOpenConns = 0
	}

	if maxIdleConns < 0 {
		maxIdleConns = 0
	}

	m.db.SetMaxOpenConns(maxOpenConns)
	m.db.SetMaxIdleConns(maxIdleConns)

	return m.db.Ping()
}

// Close xxx
func (m *MySql) Close() {
	if m.db != nil {
		m.db.Close()
	}
}

// Query xxx
func (m *MySql) Query(sql string) (map[string]map[string]string, error) {
	rows, err := m.db.Query(sql)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	cols, _ := rows.Columns()
	valCols := make([][]byte, len(cols))
	scans := make([]interface{}, len(cols))
	for i := range valCols {
		scans[i] = &valCols[i]
	}

	results := make(map[string]map[string]string)
	i := 0
	for rows.Next() {
		if err := rows.Scan(scans...); err != nil {
			return nil, err
		}

		row := make(map[string]string)
		for k, v := range valCols {
			key := cols[k]
			row[key] = string(v)
		}

		results[strconv.Itoa(i)] = row
		i++
	}

	return results, nil
}

// Insert xxx
func (m *MySql) Insert(sql string, data []interface{}) (int64, error) {
	stmt, err := m.db.Prepare(sql)
	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	res, err := stmt.Exec(data...)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Update xxx
func (m *MySql) Update(sql string, data []interface{}) (int64, error) {
	stmt, err := m.db.Prepare(sql)
	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	res, err := stmt.Exec(data...)
	if err != nil {
		return 0, err
	}
	num, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return num, nil
}

// Remove xxx
func (m *MySql) Remove(sql string, data []interface{}) (int64, error) {
	stmt, err := m.db.Prepare(sql)
	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	res, err := stmt.Exec(data...)
	if err != nil {
		return 0, err
	}
	num, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return num, nil
}
