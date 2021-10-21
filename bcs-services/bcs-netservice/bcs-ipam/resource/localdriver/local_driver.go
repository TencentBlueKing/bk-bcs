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

package localdriver

import (
	"database/sql"
	"fmt"
	types "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/bcs-ipam/resource"

	"go4.org/lock"
	//import sqlite3
	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultSQLiteDatabase = "/data/bcs/bcs-cni/bin/bcs-ipam.db"
	defaultBcsIPAMLocker  = "/data/bcs/bcs-cni/bin/ipam.lock"
)

//NewDriver create SQLite3 standard IPDriver
func NewDriver() (resource.IPDriver, error) {
	db, err := sql.Open("sqlite3", defaultSQLiteDatabase)
	if err != nil {
		return nil, err
	}
	driver := &SQLiteDriver{
		database: db,
	}
	return driver, nil
}

//SQLiteDriver driver for sqlite3
type SQLiteDriver struct {
	database *sql.DB
}

//GetIPAddr get available ip resource for contaienr
func (driver *SQLiteDriver) GetIPAddr(host string, containerID, requestIP string) (*types.IPInfo, error) {
	//safe lock for multiprocess
	closer, lockErr := lock.Lock(defaultBcsIPAMLocker)
	if lockErr != nil {
		return nil, fmt.Errorf("get bcs-ipam lock err: %v", lockErr)
	}
	defer closer.Close()
	var sql string
	if len(requestIP) == 0 {
		sql = "select Host,Net,Mask,Gateway from Resource where Status = 'available'"
	} else {
		sql = "select Host,Net,Mask,Gateway from Resource where Status = 'reserved' and Host = '" + requestIP + "'"
	}
	rows, err := driver.database.Query(sql)
	if err != nil {
		return nil, fmt.Errorf("Get available ip resource Err: %s", err.Error())
	}
	if rows.Next() {
		info := new(types.IPInfo)
		readErr := rows.Scan(&info.IPAddr, &info.Pool, &info.Mask, &info.Gateway)
		if readErr != nil {
			return nil, fmt.Errorf("Read available ip failed, %s", readErr.Error())
		}
		rows.Close()
		//active ip address selected
		stmt, _ := driver.database.Prepare("update Resource set Status = ?, Container = ? where Host = ?")
		defer stmt.Close()
		_, exErr := stmt.Exec("active", containerID, info.IPAddr)
		if exErr != nil {
			return nil, fmt.Errorf("Lock %s failed, %s", info.IPAddr, exErr.Error())
		}
		return info, nil
	}
	return nil, fmt.Errorf("No available ip resource")
}

//ReleaseIPAddr release ip address for container
func (driver *SQLiteDriver) ReleaseIPAddr(host string, containerID string, ipInfo *types.IPInfo) error {
	//safe lock for multiprocess
	closer, lockErr := lock.Lock(defaultBcsIPAMLocker)
	if lockErr != nil {
		return fmt.Errorf("get bcs-ipam lock err: %v", lockErr)
	}
	defer closer.Close()
	status := "available"
	if len(ipInfo.IPAddr) != 0 {
		status = "reserved"
	}
	//release ip address
	stmt, _ := driver.database.Prepare("update Resource set Status = ?, Container = ? where Container = ? and Status = ?")
	defer stmt.Close()
	res, exErr := stmt.Exec(status, "none container", containerID, "active")
	if exErr != nil {
		return fmt.Errorf("Release %s failed, %s", containerID, exErr.Error())
	}
	num, _ := res.RowsAffected()
	if num < 1 {
		return fmt.Errorf("Release %s failed, IP status Not Affected", containerID)
	}
	return nil
}

//GetHostInfo Get host info from driver
func (driver *SQLiteDriver) GetHostInfo(host string) (*types.HostInfo, error) {
	//safe lock for multiprocess
	closer, lockErr := lock.Lock(defaultBcsIPAMLocker)
	if lockErr != nil {
		return nil, fmt.Errorf("get bcs-ipam lock err: %v", lockErr)
	}
	defer closer.Close()
	rows, err := driver.database.Query("select Host,Net,Mask,Gateway,Container from Resource where Status = 'active'")
	if err != nil {
		return nil, fmt.Errorf("Get active ip resource Err: %s", err.Error())
	}
	hostInfo := &types.HostInfo{
		IPAddr:     host,
		Containers: make(map[string]*types.IPInst),
	}
	for rows.Next() {
		info := &types.IPInst{
			Status:     "active",
			LastStatus: "available",
			Host:       host,
		}
		readErr := rows.Scan(&info.IPAddr, &info.Pool, &info.Mask, &info.Gateway, &info.Container)
		if readErr != nil {
			return nil, fmt.Errorf("Read active ip info err, %s", readErr.Error())
		}
		hostInfo.Containers[info.Container] = info
	}
	rows.Close()
	return hostInfo, nil
}
