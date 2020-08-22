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

package service

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"bk-bscp/internal/protocol/accessserver"
	"bk-bscp/internal/protocol/common"
	pkgcommon "bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

//CreateBusinessOption all option for create business, all details
// information for yaml
type CreateBusinessOption struct {
	//Kind to indicate data type
	Kind string
	//APIVersion version information for compatibility
	APIVersion string
	//Spec business details
	Spec *common.Business
	//DBName for business storage
	DBName string
	//DB shardingDB details
	DB *common.ShardingDB
}

//LoadConfig loading yaml configuration for creating new business
func (option *CreateBusinessOption) LoadConfig(cfg string) error {
	vip := viper.New()
	vip.SetConfigFile(cfg)
	if err := vip.ReadInConfig(); err != nil {
		logger.V(3).Infof("reading createBusiness yaml file failed, %s", err.Error())
		return err
	}
	//construct business detail information
	if !vip.IsSet("spec.name") {
		logger.V(3).Infof("Loading CreateBusinessOption err, lost spec.name in yaml")
		return fmt.Errorf("spec.name is required")
	}
	if !vip.IsSet("spec.creator") {
		logger.V(3).Infof("Loading CreateBusinessOption err, lost spec.creator in yaml")
		return fmt.Errorf("spec.creator is required")
	}
	if !vip.IsSet("spec.auth") {
		logger.V(3).Infof("Loading CreateBusinessOption err, lost spec.auth in yaml")
		return fmt.Errorf("spec.auth is required")
	}
	if !vip.IsSet("spec.deptID") {
		logger.V(3).Infof("Loading CreateBusinessOption err, lost spec.deptID in yaml")
		return fmt.Errorf("spec.deptID is required")
	}
	if !vip.IsSet("spec.memo") {
		vip.SetDefault("spec.memo", vip.GetString("spec.name"))
	}
	option.Spec = &common.Business{
		Name:    vip.GetString("spec.name"),
		Depid:   vip.GetString("spec.deptID"),
		Creator: vip.GetString("spec.creator"),
		Memo:    vip.GetString("spec.memo"),
		Auth:    vip.GetString("spec.auth"),
	}
	if !vip.IsSet("db.dbID") {
		logger.V(3).Infof("Loading CreateBusinessOption err, lost db.dbID in yaml")
		return fmt.Errorf("db.dbID is required")
	}
	if !vip.IsSet("db.dbName") {
		logger.V(3).Infof("Loading CreateBusinessOption err, lost db.dbName in yaml")
		return fmt.Errorf("db.dbName is required")
	}
	option.DBName = vip.GetString("db.dbName")
	if !vip.IsSet("db.host") {
		vip.SetDefault("db.host", "")
	}
	if !vip.IsSet("db.port") {
		vip.SetDefault("db.port", 0)
	}
	if !vip.IsSet("db.user") {
		vip.SetDefault("db.user", "")
	}
	if !vip.IsSet("db.password") {
		vip.SetDefault("db.password", "")
	}
	if !vip.IsSet("db.memo") {
		vip.SetDefault("db.memo", "")
	}
	option.DB = &common.ShardingDB{
		Dbid:     vip.GetString("db.dbID"),
		Host:     vip.GetString("db.host"),
		Port:     vip.GetInt32("db.port"),
		User:     vip.GetString("db.user"),
		Password: vip.GetString("db.password"),
		Memo:     vip.GetString("db.memo"),
	}
	return nil
}

//IsNewShardingDB check if we need creating new ShardingDB
func (option *CreateBusinessOption) IsNewShardingDB() bool {
	if len(option.DB.Host) != 0 && option.DB.Port != 0 {
		return true
	}
	return false
}

//Valid check option is valid at least information
func (option *CreateBusinessOption) Valid() bool {
	if len(option.Spec.Name) == 0 || len(option.Spec.Creator) == 0 {
		return false
	}
	return true
}

//IsShardingValid check if ShardingDB information is valid
func (option *CreateBusinessOption) IsShardingValid() bool {
	if len(option.DB.Dbid) == 0 {
		return false
	}
	if option.IsNewShardingDB() {
		if len(option.DB.User) == 0 || len(option.DB.Password) == 0 {
			return false
		}
	}
	return true
}

//CreateBusiness create new business with new database sharding or
//create new business with specified database sharding
//return:
//	businessID: when creating successfully, system will response ID for business
//	error: any error if happened
func (operator *AccessOperator) CreateBusiness(cxt context.Context, option *CreateBusinessOption) (string, error) {
	if option == nil {
		return "", fmt.Errorf("Lost create new business info")
	}
	if !option.Valid() {
		logger.V(3).Infof("CreateBusiness: lost business name or creator in request")
		return "", fmt.Errorf("business name or creator is required")
	}
	if !option.IsShardingValid() {
		logger.V(3).Infof("CreateBusiness: shardingDB info is invalid, lost user/passwd or database index")
		return "", fmt.Errorf("shardingDB information is required")
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	//todo(DeveloperJim): Dbid is lost, need to confirm
	//that is gennerating from bscp system
	if option.IsNewShardingDB() {
		logger.V(3).Infof(
			"CreateBusiness: %s create new shardingDB for new business %s. ShardingDB: %s:%d",
			option.Spec.Creator, option.Spec.Name, option.DB.Host, option.DB.Port)
		//ready to create new shardingDB
		newSharding := &accessserver.CreateShardingDBReq{
			//todo(DeveloperJim): DbName is lost, need to confirm
			//that bscp needs to store in system for user interactive convenience
			Seq:      pkgcommon.Sequence(),
			Dbid:     option.DB.Dbid,
			Host:     option.DB.Host,
			Port:     option.DB.Port,
			User:     option.DB.User,
			Password: option.DB.Password,
		}
		shardingRes, err := operator.Client.CreateShardingDB(cxt, newSharding, grpcOptions...)
		if err != nil {
			logger.V(3).Infof("CreateBusiness: creating new shardingDB failed, %s", err.Error())
			return "", err
		}
		if shardingRes.ErrCode != common.ErrCode_E_OK {
			logger.V(3).Infof(
				"CreateBusiness: post %s for createShardingDB %s/%d successfully, but response err: %s",
				newSharding.User, newSharding.Host, newSharding.Port, shardingRes.ErrMsg)
			return "", fmt.Errorf("%s", shardingRes.ErrMsg)
		}
	}
	//ready to create Business with newly creating ShardingDB information
	request := &accessserver.CreateBusinessReq{
		Seq:     pkgcommon.Sequence(),
		Name:    option.Spec.Name,
		Depid:   option.Spec.Depid,
		Dbid:    option.DB.Dbid,
		Dbname:  option.DBName,
		Creator: option.Spec.Creator,
		Auth:    option.Spec.Auth,
		Memo:    option.Spec.Memo,
	}
	response, err := operator.Client.CreateBusiness(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("CreateBusiness: post new business [%s] failed, %s", option.Spec.Name, err.Error())
		return "", err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof(
			"CreateBusiness: post new business [%s] successfully, but response failed: %s",
			option.Spec.Name, response.ErrMsg,
		)
		return "", fmt.Errorf("%s", response.ErrMsg)
	}
	if len(response.Bid) == 0 {
		logger.V(3).Infof("CreateBusiness: BSCP system error, No BusinessID response")
		return "", fmt.Errorf("Lost BusinessID from configuraiotn platform")
	}
	return response.Bid, nil
}

//GetBusiness get specified Business information
//return:
//	business: specified business, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetBusiness(cxt context.Context, name string) (*common.Business, error) {
	request := &accessserver.QueryBusinessReq{
		Seq:  pkgcommon.Sequence(),
		Name: name,
	}
	grpcOptions := []grpc.CallOption{
		//grpc.WaitForReady(true),
	}

	response, err := operator.Client.QueryBusiness(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetBusiness %s failed, %s", name, err.Error())
		return nil, err
	}
	//data not found
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetBusiness %s: resource Not Found.", name)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetBusiness %s success, but response Err, %s", name, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Business, nil
}

//GetBusinessByID get specified Business information
//return:
//	business: specified business, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetBusinessByID(cxt context.Context, ID string) (*common.Business, error) {
	request := &accessserver.QueryBusinessReq{
		Seq: pkgcommon.Sequence(),
		Bid: ID,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryBusiness(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetBusiness %s failed, %s", ID, err.Error())
		return nil, err
	}
	//data not found
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetBusiness %s: resource Not Found.", ID)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetBusiness %s success, but response Err, %s", ID, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Business, nil
}

//ListBusiness list all Business information
//return:
//	businesses: all business, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) ListBusiness(cxt context.Context) ([]*common.Business, error) {
	request := &accessserver.QueryBusinessListReq{
		Seq:   pkgcommon.Sequence(),
		Index: operator.index,
		Limit: operator.limit,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryBusinessList(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ListBusiness failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ListBusiness all successfully, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Businesses, nil
}

//GetShardingDB get specified ShardingDB information
//return:
//	shardingDB: specified DB sharding, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetShardingDB(cxt context.Context, DBID string) (*common.ShardingDB, error) {
	request := &accessserver.QueryShardingDBReq{
		Seq:  pkgcommon.Sequence(),
		Dbid: DBID,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryShardingDB(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetShardingDB %s failed, %s", DBID, err.Error())
		return nil, err
	}
	//data not found
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetShardingDB %s: resource Not Found.", DBID)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetShardingDB %s success, but response Err, %s", DBID, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.ShardingDB, nil
}

//GetSharding get specified Sharding information
//return:
//	sharding: specified sharding, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetSharding(cxt context.Context, key string) (*common.Sharding, error) {
	request := &accessserver.QueryShardingReq{
		Seq: pkgcommon.Sequence(),
		Key: key,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QuerySharding(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetSharding %s failed, %s", key, err.Error())
		return nil, err
	}
	//data not found
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetSharding %s: resource Not Found.", key)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetSharding %s success, but response Err, %s", key, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Sharding, nil
}

//ListShardingDB list all ShardingDB from BSCP, at least one shardingDB info for platform
//return:
//	shardingDB: all sharding DB sharding, nil if no sharding info
//	error: error info if that happened
func (operator *AccessOperator) ListShardingDB(cxt context.Context) ([]*common.ShardingDB, error) {
	//query all shardingDB information
	request := &accessserver.QueryShardingDBListReq{
		Seq: pkgcommon.Sequence(),
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryShardingDBList(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ListAllShardingDB failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ListShardingDB all success, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	if len(response.ShardingDBs) == 0 {
		//if system initialize successfully, we get one shardingDB
		//from platform at least
		logger.V(3).Infof("ListShardingDB get 0 sharding in BSCP")
		return nil, fmt.Errorf("Lost all shardingDB in system")
	}
	return response.ShardingDBs, nil
}
