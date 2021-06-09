/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package extendedresource

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/logs"

	// import go sqlite library
	_ "github.com/mattn/go-sqlite3"
	"go4.org/lock"
)

const (
	// DataFileName file name of extended resource db data
	DataFileName = "bcs-executor-extendedresource.db"
	// LockerFileName file name of extended resource lock for multiple executor
	LockerFileName = "bcs-executor-extendedresource.lock"
)

// Driver data driver for extended resources
type Driver struct {
	dataFilePath string
	lockerPath   string
	closer       io.Closer
}

// NewDriver create extended resources data driver
func NewDriver(dir string) *Driver {
	return &Driver{
		dataFilePath: filepath.Join(dir, DataFileName),
		lockerPath:   filepath.Join(dir, LockerFileName),
	}
}

// Lock should not lock multiple times
func (d *Driver) Lock() error {
	ticker := time.NewTicker(1 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		closer, lockErr := lock.Lock(d.lockerPath)
		if lockErr == nil {
			d.closer = closer
			return nil
		}
		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return fmt.Errorf("try lock timeout")
		}
	}
}

// Unlock do unlock
func (d *Driver) Unlock() error {
	return d.closer.Close()
}

// AddRecord add record of extended resources
func (d *Driver) AddRecord(resourceType, taskKey string, devices []string) error {
	db, err := sql.Open("sqlite3", d.dataFilePath)
	if err != nil {
		logs.Errorf("open db %s failed, err %s", d.dataFilePath, err.Error())
		return err
	}
	defer db.Close()

	// ensure data table
	if err = d.ensureTable(db); err != nil {
		return err
	}

	// do insert transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if cErr := tx.Commit(); err != nil {
			logs.Errorf("db transaction commit failed, err %s", cErr.Error())
		}
	}()
	stmt, err := tx.Prepare("insert into ExtendedResource(ResourceType, TaskKey, DeviceID) values(?, ?, ?)")
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		logs.Errorf("prepare sql stmt failed, err %s", err.Error())
		return err
	}

	for _, d := range devices {
		_, exeErr := stmt.Exec(resourceType, taskKey, d)
		if exeErr != nil {
			logs.Errorf("sql stmt exec failed, err %s", exeErr.Error())
			return exeErr
		}
	}
	return nil
}

// ListRecordByResourceType list all records of certain resource type
func (d *Driver) ListRecordByResourceType(resourceType string) (map[string]string, error) {
	db, err := sql.Open("sqlite3", d.dataFilePath)
	if err != nil {
		logs.Errorf("open db %s failed, err %s", d.dataFilePath, err.Error())
		return nil, err
	}
	defer db.Close()

	// ensure data table
	if err = d.ensureTable(db); err != nil {
		return nil, err
	}

	rows, err := db.Query("select DeviceID, TaskKey from ExtendedResource where ResourceType = ?", resourceType)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		logs.Errorf("query rows failed, err %s", err.Error())
		return nil, err
	}
	retMap := make(map[string]string)
	for rows.Next() {
		var deviceID string
		var taskKey string
		err = rows.Scan(&deviceID, &taskKey)
		if err != nil {
			logs.Errorf("scan db row failed, err %s", err.Error())
		}
		retMap[deviceID] = taskKey
	}
	err = rows.Err()
	if err != nil {
		logs.Errorf("query rows failed, err %s", err.Error())
		return nil, err
	}
	return retMap, nil
}

// DelRecord delete record
func (d *Driver) DelRecord(resourceType, taskKey string) error {
	db, err := sql.Open("sqlite3", d.dataFilePath)
	if err != nil {
		logs.Infof("open db %s failed, err %s", d.dataFilePath, err.Error())
		return err
	}
	defer db.Close()

	// ensure data table
	if err = d.ensureTable(db); err != nil {
		return err
	}

	// do delete transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if cErr := tx.Commit(); err != nil {
			logs.Errorf("db transaction commit failed, err %s", cErr.Error())
		}
	}()
	stmt, err := tx.Prepare("delete from ExtendedResource where ResourceType = ? and TaskKey = ?")
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		logs.Errorf("prepare sql stmt failed, err %s", err.Error())
		return err
	}
	_, err = stmt.Exec(resourceType, taskKey)
	if err != nil {
		logs.Errorf("sql stmt exec failed, err %s", err.Error())
		return err
	}
	return nil
}

func (d *Driver) ensureTable(database *sql.DB) error {
	createTableSqlStmt := `
	create table if not exists ExtendedResource (
	ResourceType varchar(128) not null,
	TaskKey varchar(1024) not null,
	DeviceID varchar(64) not null
	);
	create unique index if not exists ExtendedResourceIndex on ExtendedResource (ResourceType, TaskKey, DeviceID);
	`
	_, err := database.Exec(createTableSqlStmt)
	if err != nil {
		logs.Infof("ensure ExtendedResource table failed, err %s", err.Error())
		return err
	}
	return nil
}
