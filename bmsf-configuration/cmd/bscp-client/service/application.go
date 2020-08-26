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

	"google.golang.org/grpc"

	"bk-bscp/internal/protocol/accessserver"
	"bk-bscp/internal/protocol/common"
	pkgcommon "bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

//CreateAppOption all option for create application, all details
// information for yaml
type CreateAppOption struct {
	//Kind to indicate data type
	Kind string
	//APIVersion version information for compatibility
	APIVersion string
	//Name app name
	Name string
	//Business app belongs to
	Business string
	//Type deployType, 0 is ontainer, 1 is GSE
	Type int32
	//Creator of operation
	Creator string
	//Memo
	Memo string
}

//Valid check option is valid at least information
func (option *CreateAppOption) Valid() bool {
	if len(option.Name) == 0 || len(option.Business) == 0 {
		return false
	}
	if len(option.Creator) == 0 {
		return false
	}
	return true
}

//CreateApp create new application
//return:
//	appID: when creating successfully, system will response ID for application
//	error: any error if happened
func (operator *AccessOperator) CreateApp(cxt context.Context, option *CreateAppOption) (string, error) {
	if option == nil {
		return "", fmt.Errorf("Lost create new application info")
	}
	if option.Valid() {
		logger.V(3).Infof("CreateApp: lost app required information")
		return "", fmt.Errorf("lost app required information")
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	//get business information first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return "", err
	}
	if business == nil {
		logger.V(3).Infof("CreateApp: No relative Business %s", operator.Business)
		return "", fmt.Errorf("No relative Business %s", operator.Business)
	}
	//ready to create application
	request := &accessserver.CreateAppReq{
		Seq:        pkgcommon.Sequence(),
		Bid:        business.Bid,
		Name:       option.Name,
		DeployType: option.Type,
		Creator:    option.Creator,
		Memo:       option.Memo,
	}
	response, err := operator.Client.CreateApp(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("CreateApp: post new application [%s] failed, %s", option.Name, err.Error())
		return "", err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof(
			"CreateApp: post new application [%s] successfully, but response failed: %s",
			option.Name, response.ErrMsg,
		)
		return "", fmt.Errorf("%s", response.ErrMsg)
	}
	if len(response.Appid) == 0 {
		logger.V(3).Infof("CreateApp: BSCP system error, No AppID response")
		return "", fmt.Errorf("Lost AppID from configuraiotn platform")
	}
	return response.Appid, nil
}

//GetApp get specified Application information
//Args:
//	name: app name
//return:
//	app: specified application, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetApp(cxt context.Context, name string) (*common.App, error) {
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return nil, err
	}
	if business == nil {
		logger.V(3).Infof("GetApp: No relative Business %s", operator.Business)
		return nil, fmt.Errorf("No relative Business %s", operator.Business)
	}
	request := &accessserver.QueryAppReq{
		Seq:  pkgcommon.Sequence(),
		Bid:  business.Bid,
		Name: name,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryApp(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetApp %s failed, %s", name, err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetApp %s: resource Not Found.", name)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetApp %s successfully, but response Err, %s", name, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.App, nil
}

//GetAppByID get specified Application information
//Args:
//  bID: business ID
//	name: app name
//return:
//	app: specified application, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetAppByID(cxt context.Context, bID, name string) (*common.App, error) {
	request := &accessserver.QueryAppReq{
		Seq:  pkgcommon.Sequence(),
		Bid:  bID,
		Name: name,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryApp(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetApp %s failed, %s", name, err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetApp %s: resource Not Found.", name)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetApp %s successfully, but response Err, %s", name, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.App, nil
}

//GetAppByAppID get specified Application information
//Args:
//  bID: business ID
//	name: app name
//return:
//	app: specified application, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetAppByAppID(cxt context.Context, bID, appID string) (*common.App, error) {
	request := &accessserver.QueryAppReq{
		Seq:   pkgcommon.Sequence(),
		Bid:   bID,
		Appid: appID,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryApp(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetApp %s failed, %s", appID, err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetApp %s: resource Not Found.", appID)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetApp %s successfully, but response Err, %s", appID, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.App, nil
}

//ListApps list all application information under specified business
//return:
//	businesses: all App, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) ListApps(cxt context.Context) ([]*common.App, error) {
	//list business first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		logger.V(3).Infof("ListApp: get relative Business %s failed, %s", operator.Business, err.Error())
		return nil, err
	}
	if business == nil {
		logger.V(3).Infof("ListApp: No such relative Business %s", operator.Business)
		return nil, fmt.Errorf("No relative Business %s", operator.Business)
	}
	request := &accessserver.QueryAppListReq{
		Seq:   pkgcommon.Sequence(),
		Bid:   business.Bid,
		Index: operator.index,
		Limit: operator.limit,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryAppList(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ListApps failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ListApps all successfully, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Apps, nil
}

//UpdateAppOption all option for update application, all details
// information for yaml
type UpdateAppOption struct {
	//Name app name
	Name string
	//Type deployType, 0 is ontainer, 1 is GSE
	Type int32
	//Type deployType, 0 is Affectived, 1 is deleted
	Status int32
	//Creator of operation
	Memo string
}

//UpdateApp update specified application information
//Args:
//	appID: specified AppID
//	option: updated information[now only name, DeployType]
//return:
//	error: error info if that happened
func (operator *AccessOperator) UpdateApp(cxt context.Context, appID string, option *UpdateAppOption) error {
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		return fmt.Errorf("no business resource")
	}
	app, err := operator.GetAppByAppID(cxt, business.Bid, appID)
	if err != nil {
		return err
	}
	if app == nil {
		return fmt.Errorf("no application was found through the id you entered")
	}
	request := &accessserver.UpdateAppReq{
		Seq:        pkgcommon.Sequence(),
		Bid:        business.Bid,
		Appid:      appID,
		Operator:   operator.User,
		Name:       app.Name,
		Memo:       app.Memo,
		DeployType: app.DeployType,
		State:      app.State,
	}
	if option.Name != "" {
		request.Name = option.Name
	}
	if option.Memo != "" {
		request.Memo = option.Memo
	}
	if option.Type != -1 {
		request.DeployType = option.Type
	}
	//if option.Status != -1 {
	//	request.State = option.Status
	//}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.UpdateApp(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("UpdateApp %s failed, %s", appID, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("UpdateApp %s success, but response Err, %s", appID, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}

//UpdateZoneOption all option for update zone, all details
// information for yaml
type UpdateZoneOption struct {
	//Name app name
	Name string
	//Status deployType, 0 is Affectived, 1 is deleted
	Status int32
	//Creator of operation
	Memo string
}

func (operator *AccessOperator) UpdateZone(cxt context.Context, zoneId string, option *UpdateZoneOption) error {
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		return fmt.Errorf("no business resource")
	}
	zone, err := operator.innerGetZone(cxt, business.Bid, "", zoneId, "")
	if err != nil {
		return err
	}
	if zone == nil {
		return fmt.Errorf("no zone was found through the id you entered")
	}
	request := &accessserver.UpdateZoneReq{
		Seq:      pkgcommon.Sequence(),
		Bid:      business.Bid,
		Zoneid:   zoneId,
		Operator: operator.User,
		Name:     zone.Name,
		Memo:     zone.Memo,
		State:    zone.State,
	}
	if option.Name != "" {
		request.Name = option.Name
	}
	if option.Memo != "" {
		request.Memo = option.Memo
	}
	//if option.Status != -1 {
	//	request.State = option.Status
	//}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.UpdateZone(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("UpdateZone %s failed, %s", zoneId, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("UpdateZone %s success, but response Err, %s", zoneId, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}

//UpdateClusterOption all option for update cluster, all details
// information for yaml
type UpdateClusterOption struct {
	//Name app name
	Name string
	//Status deployType, 0 is Affectived, 1 is deleted
	Status int32
	//Memo
	Memo string
}

func (operator *AccessOperator) UpdateCluster(cxt context.Context, clusterId string, option *UpdateClusterOption) error {
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		return fmt.Errorf("no business resource")
	}
	cluster, err := operator.GetClusterAllByID(cxt, business.Bid, "", clusterId)
	if err != nil {
		return err
	}
	if cluster == nil {
		return fmt.Errorf("no cluster was found through the id you entered")
	}
	request := &accessserver.UpdateClusterReq{
		Seq:       pkgcommon.Sequence(),
		Bid:       business.Bid,
		Clusterid: clusterId,
		Operator:  operator.User,
		Name:      cluster.Name,
		Memo:      cluster.Memo,
		State:     cluster.State,
	}
	if option.Name != "" {
		request.Name = option.Name
	}
	if option.Memo != "" {
		request.Memo = option.Memo
	}
	//if option.Status != -1 {
	//	request.State = option.Status
	//}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.UpdateCluster(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("UpdateCluster %s failed, %s", clusterId, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("UpdateCluster %s success, but response Err, %s", clusterId, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}

// UpdateShardingOption all option for update sharding, all details
// information for yaml
type ShardingOption struct {
	DBID   string
	Key    string
	DbName string
	Memo   string
}

// CreateSharding to create sharding
func (operator *AccessOperator) CreateSharding(cxt context.Context, option *ShardingOption) error {
	if option == nil {
		return fmt.Errorf("create resources content not exist")
	}

	// create param
	request := &accessserver.CreateShardingReq{
		Seq:    pkgcommon.Sequence(),
		Dbid:   option.DBID,
		Key:    option.Key,
		Dbname: option.DbName,
		Memo:   option.Memo,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.CreateSharding(cxt, request, grpcOptions...)
	// check result
	if err != nil {
		logger.V(3).Infof("CreateSharding %s failed, %s", request, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("CreateSharding %s success, but response Err, %s", request, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return err
}

// UpdateSharding to update sharding by key
func (operator *AccessOperator) UpdateSharding(cxt context.Context, key string, option *ShardingOption) error {
	if option == nil {
		return fmt.Errorf("update resources content not exist")
	}
	sharding, err := operator.GetSharding(cxt, key)
	if err != nil {
		return err
	}
	if sharding == nil {
		return fmt.Errorf("sharding no exist by you enter key")
	}

	// create param
	request := &accessserver.UpdateShardingReq{
		Seq:    pkgcommon.Sequence(),
		Dbid:   sharding.Dbid,
		Key:    key,
		Dbname: sharding.Dbname,
		Memo:   sharding.Memo,
	}
	if option.DBID != "" {
		request.Dbid = option.DBID
	}
	if option.DbName != "" {
		request.Dbname = option.DbName
	}
	if option.Memo != "" {
		request.Memo = option.Memo
	}

	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.UpdateSharding(cxt, request, grpcOptions...)
	// check result
	if err != nil {
		logger.V(3).Infof("UpdateSharding %s failed, %s", key, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("UpdateSharding %s success, but response Err, %s", key, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return err
}

// UpdateShardingDBOption all option for update shardingdb, all details
// information for yaml
type ShardingDBOption struct {
	Host     string
	Port     int32
	User     string
	Password string
	Memo     string
}

// UpdateShardingDB to update shardingdb
func (operator *AccessOperator) UpdateShardingDB(cxt context.Context, dbid string, option *ShardingDBOption) error {
	if option == nil {
		return fmt.Errorf("update resources content not exist")
	}
	shardingDB, err := operator.GetShardingDB(cxt, dbid)
	if err != nil {
		return err
	}
	if shardingDB == nil {
		return fmt.Errorf("shardingDB no exist by you enter dbid")
	}
	// update param
	request := &accessserver.UpdateShardingDBReq{
		Seq:      pkgcommon.Sequence(),
		Dbid:     dbid,
		Host:     shardingDB.Host,
		Port:     shardingDB.Port,
		User:     shardingDB.User,
		Password: shardingDB.Password,
		Memo:     shardingDB.Memo,
	}

	// check update content
	if option.Host != "" {
		request.Host = option.Host
	}
	if option.Port != 0 {
		request.Port = option.Port
	}
	if option.User != "" {
		request.User = option.User
	}
	if option.Password != "" {
		request.Password = option.Password
	}
	if option.Memo != "" {
		request.Memo = option.Memo
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.UpdateShardingDB(cxt, request, grpcOptions...)
	// check result
	if err != nil {
		logger.V(3).Infof("UpdateShardingDB %s failed, %s", dbid, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("UpdateShardingDB %s success, but response Err, %s", dbid, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}

// CreateShardingDB to create shardingdb
func (operator *AccessOperator) CreateShardingDB(cxt context.Context, dbid string, option *ShardingDBOption) error {
	if option == nil {
		return fmt.Errorf("create resources content not exist")
	}
	// create param
	request := &accessserver.CreateShardingDBReq{
		Seq:      pkgcommon.Sequence(),
		Dbid:     dbid,
		Host:     option.Host,
		Port:     option.Port,
		User:     option.User,
		Password: option.Password,
		Memo:     option.Memo,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.CreateShardingDB(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("CreateShardingDB %s failed, %s", dbid, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("CreateShardingDB %s success, but response Err, %s", dbid, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}

//DeleteApp delete specified application
//return:
//	err: if any error happened
func (operator *AccessOperator) DeleteApp(cxt context.Context, name string) error {
	//do not support right now
	return nil
}

//CreateLogicCluster create new application
//return:
//	appID: when creating successfully, system will response ID for application
//	error: any error if happened
func (operator *AccessOperator) CreateLogicCluster(cxt context.Context, option *accessserver.CreateClusterReq) (string, error) {
	if option == nil {
		return "", fmt.Errorf("Lost cluster info")
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	//ready to create application
	option.Seq = pkgcommon.Sequence()
	response, err := operator.Client.CreateCluster(cxt, option, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("CreateCluster: post new cluster [%s] failed, %s", option.Name, err.Error())
		return "", err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof(
			"CreateCluster: post new Cluster [%s] successfully, but response failed: %s",
			option.Name, response.ErrMsg,
		)
		return "", fmt.Errorf("%s", response.ErrMsg)
	}
	if len(response.Clusterid) == 0 {
		logger.V(3).Infof("CreateCluster: BSCP system error, No ClusterID response")
		return "", fmt.Errorf("Lost ClusterID from configuraiotn platform")
	}
	return response.Clusterid, nil
}

//getBusinessAndApp fast way to get Business & App by their names
func getBusinessAndApp(operator *AccessOperator, businessName, appName string) (*common.Business, *common.App, error) {
	//check business first
	business, err := operator.GetBusiness(context.TODO(), businessName)
	if err != nil {
		logger.V(3).Infof("Query relative business %s failed, %s", businessName, err.Error())
		return nil, nil, err
	}
	if business == nil {
		return nil, nil, fmt.Errorf("No relative Business %s Resource", businessName)
	}
	app, err := operator.GetAppByID(context.TODO(), business.Bid, appName)
	if err != nil {
		logger.V(3).Infof("Query relative App %s failed, %s", appName, err.Error())
		return nil, nil, err
	}
	if app == nil {
		return nil, nil, fmt.Errorf("No relative Application %s Resource", appName)
	}
	return business, app, nil
}

//GetLogicCluster get specified logic cluster information
//Args:
//	appName: app name
//	clusterName: cluster name
//return:
//	cluster: specified logic cluster, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetLogicCluster(cxt context.Context, appName, clusterName string) (*common.Cluster, error) {
	//check business first
	business, app, err := getBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return nil, err
	}
	request := &accessserver.QueryClusterReq{
		Seq:   pkgcommon.Sequence(),
		Bid:   business.Bid,
		Appid: app.Appid,
		Name:  clusterName,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryCluster(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetLogicCluster failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetLogicCluster, no relative Cluster %s", clusterName)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetLogicCluster successfully, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Cluster, nil
}

//GetClusterAllByID get specified logic cluster information
//Args:
//	bID: businessID
//	clusterID: cluster ID
//return:
//	cluster: specified logic cluster, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetClusterAllByID(cxt context.Context, bID, appID, clusterID string) (*common.Cluster, error) {
	request := &accessserver.QueryClusterReq{
		Seq:       pkgcommon.Sequence(),
		Bid:       bID,
		Appid:     appID,
		Clusterid: clusterID,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryCluster(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetLogicCluster failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("innerGetClusterByID, no relative Cluster %s", clusterID)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetLogicCluster successfully, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Cluster, nil
}

func (operator *AccessOperator) innerGetClusterByID(cxt context.Context, bID, appID, clusterName string) (*common.Cluster, error) {
	request := &accessserver.QueryClusterReq{
		Seq:   pkgcommon.Sequence(),
		Bid:   bID,
		Appid: appID,
		Name:  clusterName,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryCluster(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetLogicCluster failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("innerGetClusterByID, no relative Cluster %s", clusterName)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetLogicCluster successfully, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Cluster, nil
}

//ListLogicClusterByApp list all logic cluster information under specified business
//return:
//	clusters: all clusters, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) ListLogicClusterByApp(cxt context.Context, appName string) ([]*common.Cluster, error) {
	//check business first
	business, app, err := getBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return nil, err
	}
	request := &accessserver.QueryClusterListReq{
		Seq:   pkgcommon.Sequence(),
		Bid:   business.Bid,
		Appid: app.Appid,
		Index: operator.index,
		Limit: operator.limit,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryClusterList(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ListLogicClusterByApp failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ListLogicClusterByApp all successfully, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Clusters, nil
}

//CreateZone create new zone
//return:
//	zoneID: when creating successfully, system will response ID
//	error: any error if happened
func (operator *AccessOperator) CreateZone(cxt context.Context, option *accessserver.CreateZoneReq) (string, error) {
	if option == nil {
		return "", fmt.Errorf("Lost zone info")
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	//ready to create application
	option.Seq = pkgcommon.Sequence()
	response, err := operator.Client.CreateZone(cxt, option, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("CreateZone: post new zone [%s] failed, %s", option.Name, err.Error())
		return "", err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof(
			"CreateZone: post new Zone [%s] successfully, but response failed: %s",
			option.Name, response.ErrMsg,
		)
		return "", fmt.Errorf("%s", response.ErrMsg)
	}
	if len(response.Zoneid) == 0 {
		logger.V(3).Infof("CreateZone: BSCP system error, No ZoneID response")
		return "", fmt.Errorf("Lost ZoneID from configuraiotn platform")
	}
	return response.Zoneid, nil
}

//GetZone get specified zone information
//Args:
//	appName: app name
//	cluster: cluster name
//return:
//	cluster: specified logic cluster, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetZoneByName(cxt context.Context, appName, zoneName string) (*common.Zone, error) {
	//do not supported
	business, app, err := getBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return nil, err
	}
	return operator.innerGetZone(cxt, business.Bid, app.Appid, "", zoneName)
}

func (operator *AccessOperator) innerGetZone(cxt context.Context, bID, appID, zoneID, zoneName string) (*common.Zone, error) {
	request := &accessserver.QueryZoneReq{
		Seq:    pkgcommon.Sequence(),
		Bid:    bID,
		Appid:  appID,
		Zoneid: zoneID,
		Name:   zoneName,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryZone(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetZone failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetZone, no relative Zone %s", zoneID)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetZone successfully, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Zone, nil
}

//GetZoneAllByID get specified zone information
//Args:
//	appName: app name
//	cluster: cluster name
//return:
//	cluster: specified logic cluster, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetZoneAllByID(cxt context.Context, bID, appID, zoneID string) (*common.Zone, error) {
	request := &accessserver.QueryZoneReq{
		Seq:    pkgcommon.Sequence(),
		Bid:    bID,
		Appid:  appID,
		Zoneid: zoneID,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryZone(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetZone failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetZone, no relative Zone %s", zoneID)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetZone successfully, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Zone, nil
}

//ListZones list all zone information under specified application
//return:
//	zone: all zones, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) ListZones(cxt context.Context, appName, clusterName string) ([]*common.Zone, error) {
	business, app, err := getBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return nil, err
	}
	if business == nil {
		logger.V(3).Infof("ListZone: no relative business %s", operator.Business)
		return nil, fmt.Errorf("No relative business %s", operator.Business)
	}
	if app == nil {
		logger.V(3).Infof("ListZone: no relative Application %s", appName)
		return nil, fmt.Errorf("No relative Application %s", appName)
	}
	cluster, err := operator.innerGetClusterByID(cxt, business.Bid, app.Appid, clusterName)
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		logger.V(3).Infof("ListZone: no relative cluster %s", clusterName)
		return nil, fmt.Errorf("No relative cluster %s", clusterName)
	}
	request := &accessserver.QueryZoneListReq{
		Seq:       pkgcommon.Sequence(),
		Bid:       business.Bid,
		Clusterid: cluster.Clusterid,
		Index:     operator.index,
		Limit:     operator.limit,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryZoneList(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ListZones failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ListZones all successfully, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Zones, nil
}
